/* Example SGNL Lambda authorizer.
This authorizer may be used to protect API endpoint proxied by an AWS API Gateway.
*/

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Struct for the request to the SGNL Access Service.

type sgnlRequest struct {
	Principal Principal `json:"principal"`
	Queries   []Queries `json:"queries"`
}

// Struct for the SGNL response.

type sgnlResponse struct {
	Decisions          []Decisions `json:"decisions"`
	EvaluationDuration float32     `json:"evaluationDuration"`
	IssuedAt           string      `json:"issuedAt"`
	PrincipalID        string      `json:"principalId"`
}

// Principal struct for marshalling JSON into the SGNL request struct.

type Principal struct {
	Id string `json:"Id"`
}

// Queries struct for marshalling JSON into the SGNL request struct.

type Queries struct {
	AssetID string `json:"assetID"`
	Action  string `json:"action"`
}

// SGNL authorization decisions struct.

type Decisions struct {
	Action   string `json:"action"`
	AssetID  string `json:"assetId"`
	Decision string `json:"decision"`
}

// Lambda authorizer handler.

func sgnlAuthorizer(ctx context.Context, request events.APIGatewayCustomAuthorizerRequestTypeRequest) (events.APIGatewayCustomAuthorizerResponse, error) {

	// Print MethodArn
	log.Println("SGNL Authorizer: Method ARN: " + request.MethodArn)

	// Get bearer token from the configured "token" environment variable. In production it's advisable to get the toekn from a key vault.
	token := os.Getenv("token")

	// Get the SGNL Access Service URL from the Lambda environment variables.
	sgnlUrl := os.Getenv("sgnl_url")

	// Get the principal id from the request. In production implementations it's advisable to get the principal from an encrypted and signed IDP JWT.

	principalID := request.QueryStringParameters["principal"]

	// Parse MethodArn
	tmp := strings.Split(request.MethodArn, ":")
	apiGatewayArnTmp := strings.Split(tmp[5], "/")
	awsAccountID := tmp[4]
	method := request.HTTPMethod
	path := request.Path

	//Initialize response document

	resp := NewAuthorizerResponse(principalID, awsAccountID)
	resp.Region = tmp[3]
	resp.APIID = apiGatewayArnTmp[0]
	resp.Stage = apiGatewayArnTmp[1]

	// you can send a 401 Unauthorized response to the client by failing like so:

	if len(principalID) == 0 {
		log.Println("SGNL Authorizer: Error, empty principal id.")
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}

	// Construct the query string. You can build this string from any input source such as an HTTP(S) request context or API client.

	principalJson := `{"Id":"` + principalID + `"}`
	queriesJson := `[{"assetID":"` + path + `","action":"` + method + `"}]`
	log.Println("Principal ID:", principalID)
	log.Println("AssetID:", path)

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

	log.Println("SGNL Authorizer: SGNL Request:", string(sgnlReq2))
	// Initialize http client

	client := &http.Client{}

	// Initialize request

	req, err := http.NewRequest("POST", sgnlUrl, bytes.NewBuffer(sgnlReq2))

	if err != nil {
		fmt.Print("Error while initializing request.")
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}

	// Add headers to request

	req.Header.Add("Authorization", token)
	req.Header.Add("Content-Type", "application/json")

	// Make the request to the SGNL Access Service API
	response, err := client.Do(req)

	if err != nil {
		fmt.Print("Error when making SGNL Request.")
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)

	//Print the JSON body to ensure it all works.
	if err != nil {
		fmt.Print("Error while reading response body.")
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}

	// Check response from SGNL access service API
	var jsonResponse sgnlResponse

	json.Unmarshal([]byte(body), &jsonResponse)

	if err != nil {
		log.Println("could not unmarshal json: %s\n", err)
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}

	// Prepare policy document based on SGNL access service API.

	if jsonResponse.Decisions[0].Decision == "Allow" {
		log.Println("SGNL Authorizer Decision:", jsonResponse.Decisions[0].Decision)
		resp.AllowMethod(method, path)
		return resp.APIGatewayCustomAuthorizerResponse, nil

	} else {
		log.Println("SGNL Authorizer Decision:", jsonResponse.Decisions[0].Decision)
		resp.DenyAllMethods()
		return resp.APIGatewayCustomAuthorizerResponse, nil

	}
}

func main() {
	lambda.Start(sgnlAuthorizer)
}

type Effect int

const (
	Allow Effect = iota
	Deny
)

func (e Effect) String() string {
	switch e {
	case Allow:
		return "Allow"
	case Deny:
		return "Deny"
	}
	return ""
}

type AuthorizerResponse struct {
	events.APIGatewayCustomAuthorizerResponse

	// The region where the API is deployed.
	Region string

	// The AWS account id the policy will be generated for.
	AccountID string

	// The API Gateway API id.
	APIID string

	// The name of the stage used in the policy.
	Stage string
}

func NewAuthorizerResponse(principalID string, AccountID string) *AuthorizerResponse {
	return &AuthorizerResponse{
		APIGatewayCustomAuthorizerResponse: events.APIGatewayCustomAuthorizerResponse{
			PrincipalID: principalID,
			PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
				Version: "2012-10-17",
			},
		},
		// Specify region. This is taken from MethodARN.

		Region:    "",
		AccountID: AccountID,

		//Specify the APP ID for the API. This is taken from the MethodARN.

		APIID: "",

		// Specify the stage of deployment. This is taken from MethodArn.

		Stage: "",
	}
}

func (r *AuthorizerResponse) addMethod(effect Effect, verb string, resource string) {
	resourceArn := "arn:aws:execute-api:" +
		r.Region + ":" +
		r.AccountID + ":" +
		r.APIID + "/" +
		r.Stage + "/" +
		verb + "/" +
		strings.TrimLeft(resource, "/")

	s := events.IAMPolicyStatement{
		Effect:   effect.String(),
		Action:   []string{"execute-api:Invoke"},
		Resource: []string{resourceArn},
	}
	r.PolicyDocument.Statement = append(r.PolicyDocument.Statement, s)

}

func (r *AuthorizerResponse) AllowAllMethods() {
	r.addMethod(Allow, "All", "*")
}

func (r *AuthorizerResponse) DenyAllMethods() {
	r.addMethod(Deny, "All", "*")
}

func (r *AuthorizerResponse) AllowMethod(verb string, resource string) {
	r.addMethod(Allow, verb, resource)
}

func (r *AuthorizerResponse) DenyMethod(verb string, resource string) {
	r.addMethod(Deny, verb, resource)
}
