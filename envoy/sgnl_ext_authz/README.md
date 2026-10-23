# Envoy External Authorization Example
This Envoy external authorization service is a gRPC server which accepts an authorization request from the Envoy proxy, authorizes the request with the SGNL Access Service, and sends an authorization response back to the Envoy proxy.

SGNL does not host an Envoy `ext_authz` endpoint. This service is the adapter that sits between Envoy's `ext_authz` filter and the SGNL Access Service REST API (`POST /access/v2/evaluations`). You run it yourself, next to Envoy.

This is sample code, not a supported product. Review it against your own security requirements before using it beyond a test environment (see [Security Notes](#security-notes)).

## How a request is evaluated

| SGNL query field | Value taken from the Envoy request |
|------------------|------------------------------------|
| `principal.id`   | The `principal` request header. Set `principal_header` to read it from a different header instead. |
| `assetId`        | The request path, without the query string (e.g. `/anything`). Set `asset_id_header` to read it from a request header instead. |
| `action`         | The HTTP method (e.g. `GET`). Set `action` to send a fixed value (e.g. `read`) instead. |

These values must match what your SGNL policies check. SGNL compares the action and asset ID exactly as sent. For example, a policy that checks the action `read` and the asset ID `orders-api` will deny a request that is sent as `GET /orders-api`, because both the action and the leading `/` differ. Use `action` and `asset_id_header` to send values that match your policies.

| Result | Response to Envoy |
|--------|-------------------|
| SGNL returns `Allow` | OK. The request continues to the backend with an `x-custom-sgnl-authz: allow` header added. |
| SGNL returns `Deny` | `403 Forbidden` |
| Any error calling SGNL (unreachable, timeout, non-200 response, malformed or empty response) | `403 Forbidden`. The error is logged. The service fails closed. |
| No principal header (`principal` by default) | `401 Unauthorized` |

## Security Notes
1. **The principal header must come from a trusted source.** This sample trusts whatever value is in the principal request header (`principal` by default). If clients can set it, they can claim any identity. In a real deployment, have Envoy set (and overwrite) this header from an authenticated identity, for example a claim validated by the `jwt_authn` filter, before the `ext_authz` filter runs.
2. **Traffic between Envoy and this service is plaintext gRPC.** Run the service on the same host or pod as Envoy, or add TLS yourself. The call from this service to SGNL uses whatever scheme `sgnl_url` has; use `https`.
3. **The `Debug` log level logs all inbound request headers**, including any credentials the client sent. Do not use it in production.
4. **The SGNL token is stored in plaintext in `config.json`.** Protect the file, or adapt the service to read it from a secret store.

# Prerequisites
1. You have an SGNL environment up and running.

2. You have an example backend API (Integrated with Envoy) you want to provide authorization for.

3. You have configured SGNL connectors to ingest data and have defined the appropriate relationship mappings.

4. You have created SGNL policy snippets and policies for evaluation.

5. You have created an SGNL Protected System and generated a token for it. See the [Envoy Proxy Protected System guide](https://help.sgnl.ai/articles/protected-systems/protected-system-infra-envoy-proxy/).

6. You have a test Envoy proxy. See [installing the envoy proxy](https://www.envoyproxy.io/docs/envoy/latest/start/install) for more details. The sample Envoy configuration was validated and exercised end to end against Envoy 1.40.

7. You will need Git to clone this examples repository. You can follow the steps to install Git [here](https://github.com/git-guides/install-git).

8. You have a Go development environment. You can install Go by following this [link](https://go.dev/doc/install).

See our [Help Guides](https://help.sgnl.ai) for steps on configuring data sources and policies.


## Steps For Running The Example

1. Clone the example repo using Git from [here](https://github.com/SGNL-ai/examples.git).

2. Switch to the /envoy/sgnl_ext_authz directory.

3. Tidy up dependencies. Run the following command.

   ```go mod tidy```

4. Compile the Envoy authorization service. Run the following command.

   ```go build -o sgnl_grpc_server .```

5. Update the authorization service configuration. Edit `config/config.json` and replace the following configuration items:

   | Item | Description |
   |------|-------------|
   | `sgnl_token` | The Protected System token, e.g. `"Bearer {token value}"`. The `Bearer ` prefix is added if you leave it out. |
   | `sgnl_url` | The full URL of the SGNL Access Service evaluations endpoint, e.g. `"https://{access api host}/access/v2/evaluations"`. The path is required; the service posts to this URL as-is. |
   | `service_port` | The port this service listens on, e.g. `8223`. |
   | `log_level` | `Info`, `Error`, or `Debug`. |
   | `request_timeout` | Optional. How long to wait for SGNL, as a duration string. Defaults to `"3s"`. |
   | `principal_header` | Optional. The name of the request header that carries the principal ID, e.g. `"x-user-email"`. Defaults to `principal`. When set, the `principal` header is ignored. |
   | `action` | Optional. A fixed action to send to SGNL for every request, e.g. `"read"`. Defaults to the HTTP method. |
   | `asset_id_header` | Optional. The name of a request header that carries the asset ID, e.g. `"asset"`. Defaults to the request path. If set and the header is missing, the request is denied without calling SGNL. |

6. Start the authorization service.

   ```./sgnl_grpc_server```

7. The log level is set to Info by default (you may also set it to Error or Debug). You should see two log lines that look like:

  ```json
    {"level":"info","msg":"Current log level is Info","time":"2023-05-03T15:04:00-05:00"}
    {"level":"info","msg":"Service port: :8223","time":"2023-05-03T15:04:00-05:00"}
  ```

8. Change to the ```/envoy_config``` directory and the ```sgnl_ext_authz.yaml``` Envoy configuration file.

9. Ensure the `ext-authz` cluster configuration section has a valid IP address and port for the authorization service, matching `service_port`. Save configuration changes if any.

10. Run Envoy with the updated configuration file.

    ```envoy -c ./sgnl_ext_authz.yaml -l info```

# Test
1. Send a request through the Envoy listener (port 8080 in the sample configuration) with a `principal` header. The sample route matches paths starting with `/anything`.

   ```curl -i -H 'principal: user@domain.com' http://localhost:8080/anything```

   You can also use an API testing tool such as [Postman](https://www.postman.com/): set the URL to your Envoy proxy endpoint and URI, add the header `principal` with the principal value to evaluate, and send the request.

2. A `403` means SGNL denied the request or could not be reached. Check this service's logs for `SGNL ERROR` entries. A `401` means the principal header (`principal` by default) was missing. If SGNL allows the request but nothing is listening on the backend address in the Envoy configuration, Envoy returns a `503`, which still shows that authorization passed.

3. Run the unit tests (they use a fake SGNL Access Service and need no network access):

   ```go test ./...```

# Congratulations
You have run an SGNL authorization query through the [Envoy external authorization service](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/security/ext_authz_filter#arch-overview-ext-authz). By doing this, you are taking advantage of the power SGNL has to extend the Envoy proxy with enterprise capabilities such as just-in-time access management, centralized policy management, audit, and reporting.
