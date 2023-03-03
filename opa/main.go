package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/open-policy-agent/opa/rego"
)

func main() {

	// Set Go context for Rego
	ctx := context.Background()

	// Construct a Rego object that can be prepared or evaluated.
	r := rego.New(
		rego.Query("data.sgnl.authz.allow"),
		rego.Load([]string{os.Args[1]}, nil))

	// Create a prepared query that can be evaluated.
	query, err := r.PrepareForEval(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Input map to evaluate against policy. There is no input since we are delegating authorization decision to SGNL.

	input := map[string]interface{}{
		"values": map[string]interface{}{},
	}

	// Evaulate OPA policy through SGNL API and get the results.

	results, err := query.Eval(ctx, rego.EvalInput(input))

	// The Rego results set contains a helper function to determine if there is a true/false in expression evaluation.

	if err != nil {
		log.Fatal(err)
	}

	if !results.Allowed() {
		fmt.Println("SGNL:OPA Authorization decision is Deny.")
	} else {
		fmt.Println("SGNL:OPA Authorization decision is Allow.")
	}
}
