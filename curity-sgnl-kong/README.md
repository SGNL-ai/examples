# The SGNL, Curity Identity Server, and Kong Demo

This is a containerized demo environment with a patient records API that is proxied by the [Kong](https://konghq.com/) API Gateway and protected by the Curity Identity Server and [SGNL](https://www.sgnl.ai). Kong acts as an enforcement point and will enforce coarse-grained authorization through the [Curity Phantom Token Plugin](https://github.com/curityio/kong-phantom-token-plugin) and fine-grained authorization using the [SGNL Kong Plugin](https://github.com/SGNL-ai/examples/tree/main/curity-sgnl-kong). When the Kong configuration for the SGNL Kong Plugin is set to introspect the access token, the SGNL Kong plugin extracts the Curity access token from the request, introspects it, extracts the subject, and authorizes it against the SGNL policy engine. If the configuration is to extract the subject from the Curity phantom token, then the SGNL Kong Plugin simply loads the pre-validated access token JWT and extracts the subject (e.g. Alice) for authorization.

## Use Case
This demonstration showcases a patient record use case. In healthcare organizations, It is common for patients to have access to their own patient records. For example, if Alice attempts to access Bob’s patient data, she should be denied. This demo implementation includes a sample patient records API. This API is then protected by the Curity Identity Server, SGNL, and Kong. Curity authenticates the user and generates appropriate access tokens. The Kong gateway’s plugins intercept the request for the record data and orchestrates the validation of the identity in the request and enforce fine-grained access through the SGNL Kong plugin.

## Documentation
The overall solution approach is documented and described in the [SGNL Blog](https://sgnl.ai/2023/10/authorization-for-curity-protected-apis/) article on the SGNL website. To learn more about Curity.io  visit the [phantom token approach](https://curity.io/resources/learn/phantom-token-pattern/) post on the Curity.io website.

## Prerequisites
To begin, you need an SGNL client. If you do not have an SGNL client, please request one [here](https://sgnl.ai/demo/index.html).
 
2. You have configured SGNL data ingest adapters to ingest data and have defined the appropriate entity, attribute, and relationship mappings.
 
3. You have created policy snippets and access policy for evaluation. You can refer to the SGNL policy snippets under the **sgnl_policy** directory. Contact your SGNL representative for assistance with policy setup.
 
4. You have created a SGNL Protected System to test.
   
5. You will need to [install docker](https://docs.docker.com/engine/install/) on your system.
 
See our [Help Guides](https://support.sgnl.ai) for steps on configuring data sources and policies.

## Quickstart

1. Pull down the git repo `git clone https://github.com/AldoSGNL/curity-sgnl-kong.git`
   
2. Configure the SGNL plugin by navigating under the kong-image/configuration directory and editing the kong.yml file.
   1. Update the **sgnl_token** value with the SGNL-protected system token you created as part of configuring your SGNL client. Reach out to the SGNL representative if you need assistance.
   2. Update the **client_id** and **client_secret** if you created your own OAuth client in the Curity Identity Server. If not, the defaults are ok.
   3. Update the value of **introspect_token**. If this value is **true**, the SGNL plugin will call the Curity Identity Server introspection endpoint to validate the access token. If the value is **false**, then the SGNL plugin will simply rely only on the Curity Phantom Plugin to validate the JWT and simply extract the subject from the pre-validated JWT.
   4. The introspection_endpoint and sgnl_endpoint do not need to be changed.
   
3. Build the environment `docker compose build`
   
4. Start the environment `docker compose up`
   
5. Add the following entry to your `/etc/hosts` file, so that you're able to correctly call the containers from your local machine:

```
127.0.0.1 sgnl-kong-tutorial-idsvr sgnl-kong-tutorial-kong
```

5. Update the Curity server configuration. When the environment has started, go to `https://localhost:6749/admin` and log in with the user admin and password defined in `docker-compose.yml`. Go through the basic wizard and make sure to enable SSL (`Use `the `Existing``` SSL key` and `select `default-admin-ssl-key` works, or choose your own). Upload a valid license and upload the example policy, `curity/curity-sgnl-kong-config.xml`. This policy can be merged but requires the wizard to be completed and committed first.
   
6. To upload the Curity configuration, click on the "Changes" menu item. Click on "Upload". Select the `curity/curity-sgnl-kong-config.xml` (file is in the git repository you cloned) configuration file. Ensure the "Merge" check box is checked. Click the "Upload" button.
   
7. Now you are ready to commit the Curity configuration changes. Click on "Changes" and click on "Commit". Specify a comment and click on the ok button.
   
8. With the system configured, a client can obtain a token using the `www` client. Make sure to request the `openid` and `records` scope. E.g., you can call the authorization endpoint with this request sent from a browser:

```
https://sgnl-kong-tutorial-idsvr:8443/oauth/v2/oauth-authorizeclient_id=www&scope=openid%20records&response_type=code&redirect_uri=http://localhost:8080/cb
```

There are no users pre-populated in the environment. As part of the authentication process, create a user. The default SGNL policy checks that the user is the owner of the record so authorization will fail if there is a mismatch. The owners (patient) of the records are detailed in `api/server/data/records.json`. Either create a user that matches or make changes to `records.json`.

Once you receive the authorization code, you can redeem it with a curl command:

```bash
curl -k -Ss -X POST \
https://sgnl-kong-tutorial-idsvr:8443/oauth/v2/oauth-token \
-H 'Authorization: Basic {{insert base64 secret}}' \
-H 'Content-Type: application/x-www-form-urlencoded' \
-d 'grant_type=authorization_code&redirect_uri=http://localhost:8080/cdb&code={{insert code}}'
```

7. Use the Access Token and perform a GET request to the API exposed by Kong.

```bash
curl -Ss -X GET \
http://sgnl-kong-tutorial-kong:8000/records/0 \
-H 'Authorization: Bearer {{access code}}'
```

## More Information

Please visit [curity.io](https://curity.io/) for more information about the Curity Identity Server.

Please visit [sgnl.ai](https://www.sgnl.ai) for more information and assistance with testing this integration example.
