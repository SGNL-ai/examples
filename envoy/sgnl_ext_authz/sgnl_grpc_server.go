package main

/*
  Sample SGNL Envoy gRPC external authorization service.
*/
import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	rpcstatus "google.golang.org/genproto/googleapis/rpc/status"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/spf13/viper"

	auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/gogo/googleapis/google/rpc"

	log "github.com/sirupsen/logrus"
)

// Structs for SGNL request and response

type sgnlRequest struct {
	Principal Principal `json:"principal"`
	Queries   []Queries `json:"queries"`
}

type sgnlResponse struct {
	Decisions          []Decisions `json:"decisions"`
	EvaluationDuration float32     `json:"evaluationDuration"`
	IssuedAt           string      `json:"issuedAt"`
	PrincipalID        string      `json:"principalId"`
}

// The SGNL Access Service matches JSON keys case-sensitively ("assetId", not "assetID").
type Principal struct {
	Id string `json:"id"`
}

type Queries struct {
	AssetID string `json:"assetId"`
	Action  string `json:"action"`
}

type Decisions struct {
	Action   string `json:"action"`
	AssetID  string `json:"assetId"`
	Decision string `json:"decision"`
}

const (
	defaultRequestTimeout  = 3 * time.Second
	defaultPrincipalHeader = "principal"
	maxResponseBytes       = 1 << 20
)

var (
	log_level     = ""
	sgnl_port_str = ""
)

type healthServer struct{}

func (s *healthServer) Check(ctx context.Context, in *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	log.Debug("SGNL ext_authz health Ok.")
	// Envoy is checking health of service, all good, return that we are ok.
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (s *healthServer) Watch(in *healthpb.HealthCheckRequest, srv healthpb.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "The watch function is not implemented.")

	//This function is needed in order to implement the health server functions. Only Check() is implemented in this example.
}

type sgnlServer struct {
	url    string // Full SGNL Access Service evaluations URL.
	token  string // Complete Authorization header value, including the "Bearer " prefix.
	client *http.Client

	// Optional overrides for what is sent to SGNL. By default the principal is the value of the
	// "principal" request header, the asset is the request path (without the query string), and
	// the action is the HTTP method.
	principalHeader string // Lowercase name of the request header that carries the principal ID.
	action          string // Fixed action sent for every request, e.g. "read".
	assetHeader     string // Lowercase name of a request header that carries the asset ID.
}

func denyResponse() *auth.CheckResponse {
	log.Debug("Returning rpc.PERMISSION_DENIED and HTTP Forbidden.")

	return &auth.CheckResponse{
		Status: &rpcstatus.Status{
			Code: int32(rpc.PERMISSION_DENIED),
		},
		HttpResponse: &auth.CheckResponse_DeniedResponse{
			DeniedResponse: &auth.DeniedHttpResponse{
				Status: &envoy_type.HttpStatus{
					Code: envoy_type.StatusCode_Forbidden,
				},
				Body: "SGNL Authorization Result: Denied",
			},
		},
	}
}

// evaluate asks the SGNL Access Service whether the principal may perform the action on the asset.
// Anything other than a well-formed decision is returned as an error so that the caller fails closed.
func (a *sgnlServer) evaluate(ctx context.Context, principal, assetID, action string) (bool, error) {
	body, err := json.Marshal(sgnlRequest{
		Principal: Principal{Id: principal},
		Queries:   []Queries{{AssetID: assetID, Action: action}},
	})
	if err != nil {
		return false, fmt.Errorf("could not marshal the SGNL request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.url, bytes.NewReader(body))
	if err != nil {
		return false, fmt.Errorf("could not initialize the SGNL request: %w", err)
	}

	req.Header.Set("Authorization", a.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("request to the SGNL access service failed: %w", err)
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return false, fmt.Errorf("could not read the SGNL response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("SGNL access service returned HTTP %d: %.512s", resp.StatusCode, respBody)
	}

	var jsonResponse sgnlResponse

	if err := json.Unmarshal(respBody, &jsonResponse); err != nil {
		return false, fmt.Errorf("could not unmarshal the SGNL response: %w", err)
	}

	if len(jsonResponse.Decisions) == 0 {
		return false, errors.New("SGNL response contained no decisions")
	}

	return jsonResponse.Decisions[0].Decision == "Allow", nil
}

func (a *sgnlServer) Check(ctx context.Context, req *auth.CheckRequest) (*auth.CheckResponse, error) {
	log.Println("SGNL Authorization called gRPC Check()")

	httpReq := req.GetAttributes().GetRequest().GetHttp()

	b, err := json.MarshalIndent(httpReq.GetHeaders(), "", "  ")
	if err == nil {
		log.Debug("Inbound Headers: ")
		log.Debug((string(b)))
	}

	ct, err := json.MarshalIndent(req.GetAttributes().GetContextExtensions(), "", "  ")
	if err == nil {
		log.Debug("Context:")
		log.Debug((string(ct)))
	}

	// Get principal from request. Envoy lowercases header names.

	principalHeader := a.principalHeader

	if principalHeader == "" {
		principalHeader = defaultPrincipalHeader
	}

	principal := httpReq.GetHeaders()[principalHeader]

	if len(principal) == 0 {
		// If no principal in header, then return unauthenticated.
		log.Debug("Principal not in header.")
		return &auth.CheckResponse{
			Status: &rpcstatus.Status{
				Code: int32(rpc.UNAUTHENTICATED),
			},
			HttpResponse: &auth.CheckResponse_DeniedResponse{
				DeniedResponse: &auth.DeniedHttpResponse{
					Status: &envoy_type.HttpStatus{
						Code: envoy_type.StatusCode_Unauthorized,
					},
					Body: "Principal header is malformed or was not provided. ",
				},
			},
		}, nil
	}

	// The asset defaults to the request path without the query string, and the action to the HTTP method.
	// Both can be overridden in config.json to match the policies in SGNL.

	assetID, _, _ := strings.Cut(httpReq.GetPath(), "?")

	if a.assetHeader != "" {
		assetID = httpReq.GetHeaders()[a.assetHeader]

		if len(assetID) == 0 {
			log.Errorf("SGNL ERROR: the %q header that carries the asset ID is missing.", a.assetHeader)
			return denyResponse(), nil
		}
	}

	action := httpReq.GetMethod()

	if a.action != "" {
		action = a.action
	}

	allowed, err := a.evaluate(ctx, principal, assetID, action)
	if err != nil {
		// Fail closed: an error talking to SGNL must never result in an allow.
		log.Error("SGNL ERROR: ", err)
	}

	// If allowed, return ok.
	if err == nil && allowed {
		log.Debug("Returning rpc.OK")
		return &auth.CheckResponse{
			Status: &rpcstatus.Status{
				Code: int32(rpc.OK),
			},
			HttpResponse: &auth.CheckResponse_OkResponse{
				OkResponse: &auth.OkHttpResponse{
					Headers: []*core.HeaderValueOption{
						{
							Header: &core.HeaderValue{
								Key:   "x-custom-sgnl-authz",
								Value: "allow",
							},
						},
					},
				},
			},
		}, nil
	}

	// If not allowed, then return PERMISSION_DENIED and 403 Forbidden.
	return denyResponse(), nil
}

func main() {

	// Set log format to JSON.
	log.SetFormatter(&log.JSONFormatter{})

	// Load service configuration
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.SetConfigName("config") // Register config file name (no extension)
	viper.SetConfigType("json")   // Look for specific type

	// Read config file.

	err := viper.ReadInConfig()

	if err != nil {
		log.Error("Error reading config file: ", err)
		os.Exit(1)
	} else {

		// Check config items.

		sgnlURL := viper.GetString("service.base_config.sgnl_url")

		if len(sgnlURL) == 0 {
			log.Error("Missing sgnl_url configuration. Check config file.")
			os.Exit(1)
		}

		parsedURL, err := url.Parse(sgnlURL)

		if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Host == "" {
			log.Error("sgnl_url must be an absolute URL such as https://<access-api-host>/access/v2/evaluations. Check config file.")
			os.Exit(1)
		}

		if !strings.HasSuffix(strings.TrimRight(parsedURL.Path, "/"), "/evaluations") {
			log.Warn("sgnl_url does not end in /evaluations. The SGNL Access Service evaluation endpoint is /access/v2/evaluations.")
		}

		token := strings.TrimSpace(viper.GetString("service.tokens.sgnl_token"))

		if len(token) == 0 {
			log.Error("Missing sgnl_token configuration. Check config file.")
			os.Exit(1)
		}

		if strings.Contains(token, "{{") {
			log.Error("sgnl_token still contains the {{insert your token here}} placeholder. Check config file.")
			os.Exit(1)
		}

		// The SGNL Access Service rejects an Authorization header that does not start with "Bearer ".
		if !strings.HasPrefix(token, "Bearer ") {
			token = "Bearer " + token
		}

		requestTimeout := defaultRequestTimeout

		if d := viper.GetDuration("service.base_config.request_timeout"); d > 0 {
			requestTimeout = d
		}

		if len(viper.GetString("service.base_config.service_port")) == 0 {
			log.Error("Missing service_port configuration. Check config file.")
			os.Exit(1)
		} else {
			sgnl_port_str = ":" + viper.GetString("service.base_config.service_port")
		}

		if len(viper.GetString("service.base_config.log_level")) == 0 {
			log.Error("Missing log_level configuration. Check config file.")
			os.Exit(1)
		} else {
			log_level = viper.GetString("service.base_config.log_level")
		}

		// Set and print log level.

		if log_level == "Debug" {
			log.SetLevel(log.DebugLevel)
			log.Debug("Current log level is " + log_level)
			log.Debug("Service port: ", sgnl_port_str)
		} else if log_level == "Error" {
			log.SetLevel(log.ErrorLevel)
			log.Error("Current log level is " + log_level)
		} else if log_level == "Info" {
			log.SetLevel(log.InfoLevel)
			log.Info("Current log level is " + log_level)
			log.Info("Service port: ", sgnl_port_str)
		}

		// Create the listener.
		grpcport := flag.String("grpcport", sgnl_port_str, "grpcport")
		listener, err := net.Listen("tcp", *grpcport)

		if err != nil {
			log.Error("Failed on port: ", sgnl_port_str)
			log.Fatalf("Failed to initialize a gRPC listener.: %v", err)
		}

		// Limit concurrent streams to 10.

		options := []grpc.ServerOption{grpc.MaxConcurrentStreams(10)}

		//Create gRPC server.

		server := grpc.NewServer(options...)

		// Register the SGNL authorization service.

		auth.RegisterAuthorizationServer(server, &sgnlServer{
			url:             sgnlURL,
			token:           token,
			client:          &http.Client{Timeout: requestTimeout},
			principalHeader: strings.ToLower(viper.GetString("service.base_config.principal_header")),
			action:          viper.GetString("service.base_config.action"),
			assetHeader:     strings.ToLower(viper.GetString("service.base_config.asset_id_header")),
		})

		// Register the health service.

		healthpb.RegisterHealthServer(server, &healthServer{})

		// Listen and serve.

		if err := server.Serve(listener); err != nil {
			log.Fatalf("gRPC server stopped: %v", err)
		}
	}
}
