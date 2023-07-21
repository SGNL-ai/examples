package main

/*
  Sample SGNL Envoy gRPC external authorization service.
*/
import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"net"
	"net/http"
	"os"

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

type Principal struct {
	Id string `json:"Id"`
}

type Queries struct {
	AssetID string `json:"assetID"`
	Action  string `json:"action"`
}

type Decisions struct {
	Action   string `json:"action"`
	AssetID  string `json:"assetId"`
	Decision string `json:"decision"`
}

var (
	sgnl_url             = ""
	service_bearer_token = ""
	log_level            = ""
	sgnl_port_str        = ""
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

type sgnlServer struct{}

func (a *sgnlServer) Check(ctx context.Context, req *auth.CheckRequest) (*auth.CheckResponse, error) {
	log.Println("SGNL Authorization called gRPC Check()")

	b, err := json.MarshalIndent(req.Attributes.Request.Http.Headers, "", "  ")
	if err == nil {
		log.Debug("Inbound Headers: ")
		log.Debug((string(b)))
	}

	ct, err := json.MarshalIndent(req.Attributes.ContextExtensions, "", "  ")
	if err == nil {
		log.Debug("Context:")
		log.Debug((string(ct)))
	}

	// Get principal from request.

	principalHeader := req.Attributes.Request.Http.Headers["principal"]

	// Get URI from request.

	path := req.Attributes.Request.Http.Path

	// Get HTTP method from request.

	method := req.Attributes.Request.Http.Method

	if len(principalHeader) > 0 {

		// Construct the SGNL access service query string.

		principalJson := `{"Id":"` + principalHeader + `"}`

		queriesJson := `[{"assetID":"` + path + `","action":"` + method + `"}]`

		// Initialize the query variable as an array of SGNL queries, based on the queries struct.
		var queries []Queries
		var principal Principal

		// Unmarshall the query string into the query array struct.

		json.Unmarshal([]byte(queriesJson), &queries)

		// Unmarshall the principal string into the principal struct.

		json.Unmarshal([]byte(principalJson), &principal)

		// Build final JSON request body using the SGNL request struct.

		sgnlReq1 := &sgnlRequest{
			Principal: principal,
			Queries:   queries}

		// Finally marshall the JSON into a final request variable containing the JSON request body.

		sgnlReq2, _ := json.Marshal(sgnlReq1)

		// Initialize http client

		client := &http.Client{}

		// Initialize request

		req, err := http.NewRequest("POST", sgnl_url, bytes.NewBuffer(sgnlReq2))

		if err != nil {
			log.Error("SGNL ERROR: Could not initialize request to the SGNL access service.")
		}

		// Add headers to request

		req.Header.Add("Authorization", service_bearer_token)
		req.Header.Add("Content-Type", "application/json")

		// Make the request to the SGNL Access Service API
		resp, err := client.Do(req)

		if err != nil {
			log.Error("SGNL ERROR: Error while making the request to SGNL access service, ", err)
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)

		var jsonResponse sgnlResponse

		// Unmarshal response into the jsonResponse struct.

		json.Unmarshal([]byte(body), &jsonResponse)

		if err != nil {
			log.Error("SGNL Error: could not unmarshal json response from the access service: %s\n", err)
		}

		// Check SGNL decision. If allowed, return ok. If not allowed, then return PERMISSION_DENIED and 403 Forbidden.

		if jsonResponse.Decisions[0].Decision == "Allow" {
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
		} else {
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
			}, nil

		}

	}

	// If not principal in header, then return unauthenticated.
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
		log.Error("Error reading config file.")
		os.Exit(1)
	} else {

		// Check config items.

		if len(viper.GetString("service.base_config.sgnl_url")) == 0 {
			log.Error("Missing sgnl_url configuration. Check config file.")
			os.Exit(1)
		} else {
			sgnl_url = viper.GetString("service.base_config.sgnl_url")
		}

		if len(viper.GetString("service.tokens.sgnl_token")) == 0 {
			log.Error("Missing sgnl_token configuration. Check config file.")
			os.Exit(1)
		} else {
			service_bearer_token = viper.GetString("service.tokens.sgnl_token")
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
			log.Fatalf("Failed to initialize a gRPC listener.: %v", err)
			log.Error("Faild on port: ", sgnl_port_str)
		}

		// Limit concurrent streams to 10.

		options := []grpc.ServerOption{grpc.MaxConcurrentStreams(10)}

		//Append to opts.
		options = append(options)

		//Create gRPC server.

		server := grpc.NewServer(options...)

		// Register the SGNL authorization service.

		auth.RegisterAuthorizationServer(server, &sgnlServer{})

		// Register the health service.

		healthpb.RegisterHealthServer(server, &healthServer{})

		// Listen and serve.

		server.Serve(listener)
	}
}
