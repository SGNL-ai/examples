# Azure APIM Example
This Azure APIM policy example demonstrates how to call the SGNL Access Service from an Azure APIM inbound policy.

This example shows you how to get the request context variables and JSON key/value pairs from the HTTP request body. You should adjust these variable names to get values necessary to make authorization decisions for your protected backend API. If you would like support for your tests or proof of concept, you can contact SGNL sales.


# Prerequisites
1. You have a SGNL environment up and running.
   
2. You have an example backend API you want to provide authorization for. You can follow [these steps](https://learn.microsoft.com/en-us/azure/api-management/import-api-from-oas?tabs=portal) to import an API using it's OpenAPI specification.
 
3. You have configured connectors to ingest data and have defined the appropriate relationship mappings.
 
4. You have created policy snippets and policy for evaluation.
 
5. You have created an integration to test.
 
6. You have a test Azure APIM instance setup. You can follow this [quick start](https://learn.microsoft.com/en-us/azure/api-management/get-started-create-service-instance) for creating an APIM instance.
 
7. You will need Git to clone this examples repository. You can follow the steps to install Git [here](https://github.com/git-guides/install-git).

8. Finally, you will need an API testing tool such as [Postman](https://www.postman.com/). 


See our [Help Guides](https://support.sgnl.ai) for steps on configuring data sources and policies.


## Steps For Running The Example


1. Clone the example repo using Git from https://github.com/SGNL-ai/examples.git.


2. Switch to the apim directory.


3. Copy the contents of the **sgnl-apim.policy** file and insert it into the inbound policy of the service operation you want to provide authorization for. 
 
4. Take note of the line below. The **GetValueOrDefault** function gets the Authorization header and value from the HTTP request context. Ensure you are sending a SGNL integration token in the Authorization header. APIM will forward this token to SGNL for validation.


   ```context.Request.Headers.GetValueOrDefault("Authorization")```

5. Update the following line with your SGNL client access service URL.

   ```<set-url>@("{{your SGNL client URL}}")</set-url>```

7. Update the POST body to send to the SGNL access service. This POST body should contain the principal, assets, and actions you want to check authorization on. The principal, assetId, and action JSON key/value pairs are required. In this example, the backend API path (e.g. /v1/path) is sent as an asset and the HTTP verb (e.g. POST) is sent as an action. The second authorization query sets the assetId with an account type and the action is "Create". The SGNL Access service will return the authorization decisions either "Allow" or "Deny" for each query set.

```
<set-body template="liquid">
            {
	                    "principal": {
		                "id":"{{context.Variables["principal"]}}"
	                    },
	                        "queries": [{
	                    	"assetId": "{{context.Variables["backendUri"]}}",
                             "action": "{{context.Variables["method"]}}"
	                            },
                                {
                               "assetId": "{{context.Variables["accountType"]}}",
                               "action": "Create"
                                }
                                ]
                         }
            }	        
            </set-body>   
```

8. Ensure you **send the appropriate JSON body for your service**. This policy example was used to front end a sample banking API. The request to fetch the associated account for the user principal from the sample banking service looked like this:


   ``` 
   {
   "principal":"aldo@sgnl.ai",
   "account": {
      "type": "Checking"
    }
   }
   ```

9. Now that you have updated your Azure APIM policy and request body for the SGNL Access Service, you are ready to make a request (e.g. via [Postman](https://www.postman.com/)) to your example backend API **through the Azure APIM gateway**. 

**Notes** This Azure APIM policy will make an external call to your SGNL access service with the appropriate access query. The SGNL Access Service returns either an "Allow" or "Deny". If the decision is "Allow" , the request will be allowed to your example backend API. If the decision is "Deny" the Azure APIM policy will return a 403 forbidden to your API client.


# Congratulations
You have now run an SGNL authorization query through Azure APIM. By doing this, you are taking advantage of the power SGNL has to extend Azure APIM with enterprise capabilities such as just in time access management, centralized policy management, audit, and reporting.



