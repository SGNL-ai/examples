package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/gogo/googleapis/google/rpc"
)

func newCheckRequest(headers map[string]string, method, path string) *auth.CheckRequest {
	return &auth.CheckRequest{
		Attributes: &auth.AttributeContext{
			Request: &auth.AttributeContext_Request{
				Http: &auth.AttributeContext_HttpRequest{
					Headers: headers,
					Method:  method,
					Path:    path,
				},
			},
		},
	}
}

func TestCheck(t *testing.T) {
	const token = "Bearer test-token"

	tests := map[string]struct {
		headers            map[string]string
		serverPrincipalHdr string
		serverAction       string
		serverAssetHdr     string
		sgnlStatus         int
		sgnlBody           string
		sgnlUnreachable    bool
		wantCode           rpc.Code
		wantDenied         envoy_type.StatusCode
		wantSGNLCalls      int
		wantPrincipalID    string // Defaults to the value of the "principal" header.
		wantAction         string // Defaults to GET.
		wantAssetID        string // Defaults to /anything.
	}{
		"custom_principal_header": {
			headers:            map[string]string{"x-user-email": "alice@example.com"},
			serverPrincipalHdr: "x-user-email",
			sgnlStatus:         http.StatusOK,
			sgnlBody:           `{"decisions":[{"decision":"Allow"}]}`,
			wantCode:           rpc.OK,
			wantSGNLCalls:      1,
			wantPrincipalID:    "alice@example.com",
		},
		"custom_principal_header_ignores_default_header": {
			headers:            map[string]string{"principal": "mallory@example.com"},
			serverPrincipalHdr: "x-user-email",
			wantCode:           rpc.UNAUTHENTICATED,
			wantDenied:         envoy_type.StatusCode_Unauthorized,
		},
		"action_and_asset_header_overrides": {
			headers:        map[string]string{"principal": "alice@example.com", "asset": "orders-api"},
			serverAction:   "read",
			serverAssetHdr: "asset",
			sgnlStatus:     http.StatusOK,
			sgnlBody:       `{"decisions":[{"decision":"Allow"}]}`,
			wantCode:       rpc.OK,
			wantSGNLCalls:  1,
			wantAction:     "read",
			wantAssetID:    "orders-api",
		},
		"asset_header_configured_but_missing": {
			headers:        map[string]string{"principal": "alice@example.com"},
			serverAssetHdr: "asset",
			wantCode:       rpc.PERMISSION_DENIED,
			wantDenied:     envoy_type.StatusCode_Forbidden,
		},
		"allow": {
			headers:       map[string]string{"principal": "alice@example.com"},
			sgnlStatus:    http.StatusOK,
			sgnlBody:      `{"decisions":[{"action":"GET","assetId":"/anything","decision":"Allow"}]}`,
			wantCode:      rpc.OK,
			wantSGNLCalls: 1,
		},
		"deny": {
			headers:       map[string]string{"principal": "alice@example.com"},
			sgnlStatus:    http.StatusOK,
			sgnlBody:      `{"decisions":[{"action":"GET","assetId":"/anything","decision":"Deny"}]}`,
			wantCode:      rpc.PERMISSION_DENIED,
			wantDenied:    envoy_type.StatusCode_Forbidden,
			wantSGNLCalls: 1,
		},
		"sgnl_rejects_token": {
			headers:       map[string]string{"principal": "alice@example.com"},
			sgnlStatus:    http.StatusUnauthorized,
			sgnlBody:      `{"error":"unauthorized"}`,
			wantCode:      rpc.PERMISSION_DENIED,
			wantDenied:    envoy_type.StatusCode_Forbidden,
			wantSGNLCalls: 1,
		},
		"sgnl_returns_no_decisions": {
			headers:       map[string]string{"principal": "alice@example.com"},
			sgnlStatus:    http.StatusOK,
			sgnlBody:      `{"decisions":[]}`,
			wantCode:      rpc.PERMISSION_DENIED,
			wantDenied:    envoy_type.StatusCode_Forbidden,
			wantSGNLCalls: 1,
		},
		"sgnl_returns_invalid_json": {
			headers:       map[string]string{"principal": "alice@example.com"},
			sgnlStatus:    http.StatusOK,
			sgnlBody:      `not json`,
			wantCode:      rpc.PERMISSION_DENIED,
			wantDenied:    envoy_type.StatusCode_Forbidden,
			wantSGNLCalls: 1,
		},
		"sgnl_unreachable": {
			headers:         map[string]string{"principal": "alice@example.com"},
			sgnlUnreachable: true,
			wantCode:        rpc.PERMISSION_DENIED,
			wantDenied:      envoy_type.StatusCode_Forbidden,
		},
		"principal_with_json_special_characters": {
			headers:       map[string]string{"principal": `a"b\c@example.com`},
			sgnlStatus:    http.StatusOK,
			sgnlBody:      `{"decisions":[{"decision":"Allow"}]}`,
			wantCode:      rpc.OK,
			wantSGNLCalls: 1,
		},
		"missing_principal": {
			headers:    map[string]string{},
			wantCode:   rpc.UNAUTHENTICATED,
			wantDenied: envoy_type.StatusCode_Unauthorized,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				mu         sync.Mutex
				calls      int
				gotAuth    string
				gotRequest map[string]any
			)

			sgnl := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)

				mu.Lock()
				calls++
				gotAuth = r.Header.Get("Authorization")
				_ = json.Unmarshal(body, &gotRequest)
				mu.Unlock()

				w.WriteHeader(tc.sgnlStatus)
				_, _ = w.Write([]byte(tc.sgnlBody))
			}))

			if tc.sgnlUnreachable {
				sgnl.Close()
			} else {
				defer sgnl.Close()
			}

			server := &sgnlServer{
				url:             sgnl.URL + "/access/v2/evaluations",
				token:           token,
				client:          &http.Client{Timeout: 2 * time.Second},
				principalHeader: tc.serverPrincipalHdr,
				action:          tc.serverAction,
				assetHeader:     tc.serverAssetHdr,
			}

			resp, err := server.Check(context.Background(), newCheckRequest(tc.headers, "GET", "/anything?x=1"))
			if err != nil {
				t.Fatalf("Check() returned unexpected error: %v", err)
			}

			if got := rpc.Code(resp.GetStatus().GetCode()); got != tc.wantCode {
				t.Errorf("Check() status code = %v, want %v", got, tc.wantCode)
			}

			if tc.wantCode == rpc.OK {
				if resp.GetOkResponse() == nil {
					t.Error("Check() expected an OkResponse")
				}
			} else if got := resp.GetDeniedResponse().GetStatus().GetCode(); got != tc.wantDenied {
				t.Errorf("Check() denied HTTP status = %v, want %v", got, tc.wantDenied)
			}

			mu.Lock()
			defer mu.Unlock()

			if calls != tc.wantSGNLCalls {
				t.Fatalf("SGNL calls = %d, want %d", calls, tc.wantSGNLCalls)
			}

			if tc.wantSGNLCalls == 0 {
				return
			}

			if gotAuth != token {
				t.Errorf("Authorization header = %q, want %q", gotAuth, token)
			}

			wantPrincipalID := tc.wantPrincipalID
			if wantPrincipalID == "" {
				wantPrincipalID = tc.headers["principal"]
			}

			principal, _ := gotRequest["principal"].(map[string]any)
			if got := principal["id"]; got != wantPrincipalID {
				t.Errorf("principal.id = %v, want %v", got, wantPrincipalID)
			}

			queries, _ := gotRequest["queries"].([]any)
			if len(queries) != 1 {
				t.Fatalf("queries = %v, want exactly one", gotRequest["queries"])
			}

			wantAssetID, wantAction := tc.wantAssetID, tc.wantAction
			if wantAssetID == "" {
				wantAssetID = "/anything"
			}

			if wantAction == "" {
				wantAction = "GET"
			}

			query, _ := queries[0].(map[string]any)
			if got := query["assetId"]; got != wantAssetID {
				t.Errorf("queries[0].assetId = %v, want %v", got, wantAssetID)
			}

			if got := query["action"]; got != wantAction {
				t.Errorf("queries[0].action = %v, want %v", got, wantAction)
			}

			if _, found := query["assetID"]; found {
				t.Error("queries[0] contains the mis-cased assetID key")
			}
		})
	}
}
