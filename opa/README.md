# OPA Example
This Go example demonstrates how to call the SGNL Access Service from an OPA Rego policy. This example uses the OPA library for Go.


# Prerequisites
1. You have a SGNL environment up and running. Please send us a message [todo: insert external link for client request form.]
 
2. You have configured connectors to ingest data and have defined the appropriate relationship mappings.
 
3. You have created policy snippets and policy for evaluation.
 
4. You have created an integration to test.
 
5. You have a Golang development environment setup. You can set up go v 1.18 by following the instructions [here](https://go.dev/doc/install).
 
6. You will need Git to clone the example repository. You can follow the steps to install Git [here](https://github.com/git-guides/install-git).


See our [Help Guides](https://support.sgnl.ai) for steps on configuring data sources and policies.


## Rego File
You will find the example Rego policy in the policy.rego file. A few notes about the Rego policy. You may run the policy on an OPA instance running as a server or you can evaluate the policy using the Go library. This example uses the Go library. Please reach out to us if you are interested in seeing an example with OPA running as a server.


## Steps For Running The Example


1. Clone the example repo using Git from https://github.com/SGNL-ai/examples.git.


2. Switch to the opa directory.


3. Ensure the dependencies in the source file are downloaded. Run:
 
   ```go mod tidy```


   You should see the command downloading the OPA Go library.


4. Update the following placeholders in the policy.rego file with your values:


   ```{{SGNL Access Service URL}}``` : Replace with your access service URL.

   ```Bearer {{token}}``` : Replace this with the integration access token. Ensure you keep the "Bearer" prefix.

   ```{{principal id}}```: Replace this with the principal you would like to test authorization for. In most cases this is an email address. Keep in mind that this principal must have been previously ingested by configuring the appropriate data source.

   ```{{asset id}}```: Replace this with the asset id you would like to test. Keep in mind that in most instances the asset is also ingested as part of the data source configuration. You may alternatively create your own asset snippet to test against.

5. You are now ready to compile the example. From your command line run:


   ```$ go build ./main.go```


6. Now that the code is compiled, you are ready to run the example and evaluate the rego policy. From your command line run:

    ```$ ./main ./policy.rego```

**Notes:** This rego policy will make an external call to your SGNL access service with the appropriate access query. The SGNL Access Service returns either an "Allow" or "Deny". If the decision is "Allow" the rego policy will evaluate to true. If the decision is "Deny" the rego policy evaluates to false.


## Congratulations
You have now run an SGNL authorization query through OPA. By doing this, you are taking advantage of the power SGNL has to extend OPA with enterprise capabilities such as just in time access management, centralized policy management, audit, and reporting.



