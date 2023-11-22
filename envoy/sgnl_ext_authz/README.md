# Envoy External Authorization Example
This Envoy external authorization service is a gRPC server which accepts an authorization request from the Envoy proxy, authorizes the request with the SGNL Access Service, and sends an authorization response back to the Envoy proxy.

# Prerequisites
1. You have an SGNL environment up and running.
   
2. You have an example backend API (Integrated with Envoy) you want to provide authorization for.
 
3. You have configured SGNL connectors to ingest data and have defined the appropriate relationship mappings.
 
4. You have created SGNL policy snippets and policies for evaluation.
 
5. You have a test Envoy proxy. See [installing the envoy proxy](https://www.envoyproxy.io/docs/envoy/latest/start/install) for more details.
 
6. You will need Git to clone this examples repository. You can follow the steps to install Git [here](https://github.com/git-guides/install-git).

7. You have a Go development environment. You can install Go by following this [link](https://go.dev/doc/install).

8. Finally, you will need an API testing tool such as [Postman](https://www.postman.com/). 


See our [Help Guides](https://help.sgnl.ai) for steps on configuring data sources and policies.


## Steps For Running The Example


1. Clone the example repo using Git from [here](https://github.com/SGNL-ai/examples.git).


2. Switch to the /envoy/sgnl_ext_authz directory.


3. Tidy up dependencies. Run the following command.
   
   ```go mod tidy``` 

4. Compile the Envoy authorization service. Run the following command.
   
   ```go build ./sgnl_grpc_server.go```

5. Update the authorization service configuration. Change directory to the config directory. Edit the config.json file and replace the following configuration items: 
   
   ```"sgnl_token": "Bearer {token value}```

   ```"sgnl_url": "{sgnl access service url}"```

	 ```"service_port": {port}```

	 ```"log_level": {Info, Error, Or Debug}```


6. Start the authorization service.
   
   ```./sgnl_grpc_server```

7. The log level is set to Info by default (you may also set it to Error or Debug). You should see two log lines that look like:

  ```json
    {"level":"info","msg":"Current log level is Info","time":"2023-05-03T15:04:00-05:00"}
    {"level":"info","msg":"Service port: :8223","time":"2023-05-03T15:04:00-05:00"}
  ```

8. Change to the ```/envoy_config``` directory and the ```sgnl_ext_authz.yaml``` Envoy configuration file.

9.  Ensure the ext-authz cluster configuration section has a valid IP address and port for the authorization service. Save configuration changes if any.

10. Run Envoy with the updated configuration file.

    ```envoy -c ./sgnl_ext_authz.yaml -l info```

# Test With Postman
1. Create a new request and set the HTTP verb (e.g., POST) to the verb implemented by your backend API.
2. Set the URL to your Envoy proxy endpoint and URI (e.g., https://host/api/endpoint).
3. Set a principal header with the principal value you want to send to the authorization service:
   ```Header Name: principal Value: user@domain.com```
4. Send the request.

# Congratulations
You have run an SGNL authorization query through the [Envoy external authorization service](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/security/ext_authz_filter#arch-overview-ext-authz). By doing this, you are taking advantage of the power SGNL has to extend the Envoy proxy with enterprise capabilities such as just-in-time access management, centralized policy management, audit, and reporting.



