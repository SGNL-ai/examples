# AWS Lambda Authorizer Example
This AWS Lambda authorizer example demonstrates how to call the SGNL Access Service from an AWS API Gateway implementation.

# Prerequisites
1. You have an SGNL environment up and running.
   
2. You have an example backend API you want to provide authorization for. You can follow [these steps](https://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-mock-integration-console.html) enable a mock integration using the AWS API gateway.
 
3. You have configured connectors to ingest data and have defined the appropriate relationship mappings.
 
4. You have created policy snippets and policies for evaluation.
 
5. You have created an integration to test.
 
6. You have a test AWS API Gateway instance setup. You can follow this [guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/getting-started.html) to get started.
 
7. You will need Git to clone this examples repository. You can follow the steps to install Git [here](https://github.com/git-guides/install-git).

8. You have a Go development environment. You can install Go by following this [link](https://go.dev/doc/install).

9. Finally, you will need an API testing tool such as [Postman](https://www.postman.com/). 


See our [Help Guides](https://help.sgnl.ai) for steps on configuring data sources and policies.


## Steps For Running The Example


1. Clone the example repo using Git from [here](https://github.com/SGNL-ai/examples.git).


2. Switch to the aws/lambda/authorizer directory.


3. Tidy up dependencies. Run the following command.
   
   ```go mod tidy``` 

4. Compile the authorizer. Run the following command.
   
   ```GOOS=linux GOARCH=amd64 go build ./sgnlAuthorizer.go```

5. Zip the executable, and prepare to upload it to the AWS Lambda function. Run the following command.
   
   ```zip sgnlAuthorizer.zip sgnlAuthorizer```

6. Create the SGNL Lambda authorizer function. You can follow these [steps](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-use-lambda-authorizer.html#api-gateway-lambda-authorizer-lambda-function-create) to create an authorizer. Take care to call the function "sgnlAuthorizer". Ensure you are using the Go 1.x runtime, your handler name is set to "sgnlAuthorizer" and you have selected "x86_64" for the chip architecture. You can find these settings in the "Runtime Settings" section.

7. Once the function is created, click on the "Code" tab, and select the "Upload from" drop down. Choose your zip file with the compiled authorizer and upload it.

8. Click on the "Test" tab and create a new event and use the "API Gateway Authorizer" template.

9. Replace the event JSON with the contents of the example event in the test_event.json file. This file is located under the test directory.

10. Click on the "Configuration" tab and create a new environment variable named "token". Once created, set the value to the SGNL integration bearer token. Be sure to include the Bearer prefix (i.e. your token value should be "Bearer {SGNL Integration Token}").

11. In the same configuration section, create a new environment variable named "sgnl_url". Set the value to the [SGNL Access Service URL for Access Evaluation](https://developer.sgnl.ai/#sgnl-public-api).

12. Click on the "Test" tab and click on the "Test" button. You should receive a successful execution result with the authorization decision from the SGNL access service. It will look like this:

```json{
  "principalId": "adela.cervantsz@sgnl.ai",
  "policyDocument": {
    "Version": "2012-10-17",
    "Statement": [
      {
        "Action": [
          "execute-api:Invoke"
        ],
        "Effect": "Deny",
        "Resource": [
          "arn:aws:execute-api:us-east-1:123456789012:abcdef123/test/All/*"
        ]
      }
    ]
  }
}
```

## Activating The Authorizer In AWS API Gateway

1. Log in to the AWS console and search for your API gateway.
2. Click on the API you want to protect.
3. Click on the authorizers menu item.
4. Create a new authorizer and name it "sgnlAuthorizer". Ensure you select "Lambda" as the type.
5. Select the "sgnlAuthorizer" Lambda function from the drop down.
6. Ensure you select "Request" for the Lambda Event Payload type.
7. This authorizer was tested by simply sending a principal in the query string. Select "Query String" from the Identity Source drop down.
8. Set the query string parameter to "principal".
9. Disable authorization caching.
10. Create the authorizer.

## Activating The Authorizer For The Resource
1. Click the "Resources" menu item.
2. Select your resource and HTTP method (e.g., POST).
3. Click on the "Method Request" link.
4. Click the pencil icon to edit the setting for "Authorization".
5. Select the "sgnlAuthorizer" from the drop down list box.
6. Click on the "check" icon to save the setting.

## Test With Postman
1. Create a new request and set the HTTP verb (e.g., POST) to the verb implemented by your API.
2. Set the URL to your API endpoint and URI (e.g., https://{AppId}.execute-api.us-east-1.amazonaws.com/mock/api).
3. Set JSON body to:
   ```json
   {
    "principal": {
        "id": "{Replace with your principal id.}"
    }
   }
4. Send the request.

## Lambda Authorizer Trigger Notes
1. If the authorizer is not triggering, you will need to define a trigger in the Lambda Function configuration.
2. Search for your SGNL Lambda Authorizer.
3. Click on the "Configuration" tab.
4. Click on "Triggers".
5. Add a new trigger for your API. 
6. Select the "API Gateway" as a source.
7. Select "Use existing API".
8. Select your API from the search drop down list.
9. Select your deployment stage.
10. Select "Open" for your security mechanism.
11. Save the new trigger configuration.
12. Call your API once more with Postman.

# Congratulations
You have run an SGNL authorization query through the AWS API Gateway and Lambda Authorizer. By doing this, you are taking advantage of the power SGNL has to extend the AWS API Gateway with enterprise capabilities such as just-in-time access management, centralized policy management, audit, and reporting.



