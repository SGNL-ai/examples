#r "Newtonsoft.Json"
using System.Net;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Primitives;
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;
using System;
using System.IO;
using System.Text;


public static async Task<IActionResult> Run(HttpRequest req, ILogger log)
{
    
    log.LogInformation("SGNL custom authentication extension HTTP trigger. Calling SGNL API...");
    
    // Get request body.

    string requestBody = await new StreamReader(req.Body).ReadToEndAsync();
    
    // Deserialize request body into JSON.

    dynamic data = JsonConvert.DeserializeObject(requestBody);

    // Extract the relevant query fields to construct SGNL query from the request body.
    
    string principalId = data?.data.authenticationContext.user.userPrincipalName;
    string ipAddress = data?.data.authenticationContext.client.ip;
    string assetId = data?.data.authenticationContext.clientServicePrincipal.appDisplayName;

    // Read the correlation ID from the Azure AD  request. This ID is sent back as part of the custom claims.

    string correlationId = data?.data.authenticationContext.correlationId;

    

     // Instantiate new query class with action and protected assetid. Note: You can instantiate mutliple query objects to be serialized into JSON.

    var query = new Query{
         action = "access",
         assetId = assetId,
      };

    // Instantiate new AzurePrincipal class with the principal.
    var azurePrincipal = new AzurePrincipal{
        id = principalId
       };

    // Initialize query array, and add queries to be send to the SGNL access service.
    var queryArray = new Query[] {query};

    // Build final request object to be serialized into JSON before sending to the SGNL access service.
    var sgnlRequest = new SGNLRequest{
         principal = azurePrincipal,
         ipAddress = ipAddress,
         queries = queryArray
         };
        
    // Serialize sgnlRequest into JSON.
       
    string postData = JsonConvert.SerializeObject(sgnlRequest);
       
    // Uncomment the line below to print the postData to the Azure function log.
    // log.LogInformation(postData);

    // Create a request for the SGNL Access API 
    WebRequest request = WebRequest.Create("https://access.sgnlapis.cloud/access/v2/evaluations");

    request.Method = "POST";
    byte[] byteArray = Encoding.UTF8.GetBytes(postData);
    
    // Set the ContentType property of the WebRequest.
    request.ContentType = "application/json";
    request.Headers["Authorization"] = "Bearer {SGNL protected system bearer token}";
    
    // Get the request stream.
    Stream dataStream = request.GetRequestStream();
    dataStream.Write(byteArray, 0, byteArray.Length);
    dataStream.Close();
    
    // Get the response.
    WebResponse response = request.GetResponse();
    dataStream = response.GetResponseStream();
    StreamReader reader = new StreamReader(dataStream);
    
    // Read the content
    string responseFromServer = reader.ReadToEnd();
    dynamic responseData = JsonConvert.DeserializeObject(responseFromServer);
     
    // Get the authorization decision from decision array.

    string decision = responseData?.decisions[0].decision;
      
    // Uncomment line below to print out the SGNL API decision to the Azure function log.
    
    // log.LogInformation(decision);

    // Close reader, data stream, and response.
    reader.Close();
    dataStream.Close();
    response.Close();

    // Determine whether to stop the login flow or continue based on authorization decision.

    if (decision == "Allow") {
        // Allow the login flow to continue.
        // Optionally setup claims to return to Azure AD and application.
        ResponseContent r = new ResponseContent();
        r.data.actions[0].claims.CorrelationId = correlationId;
        r.data.actions[0].claims.ApiVersion = "1.0.0";
        r.data.actions[0].claims.DateOfBirth = "01/01/2000";
        r.data.actions[0].claims.CustomRoles.Add("Customer Service");
        r.data.actions[0].claims.CustomRoles.Add("Support Agent");

        return new OkObjectResult(r);
    }
    else {
        // return forbidden with custom message
        log.LogInformation("Access denied, returning ObjectResult with { StatusCode = 403}");
        return new ObjectResult("Unauthorized to access resource, by SGNL.") { StatusCode = 403};
    }

}

    
    // Class for token response content.

    public class ResponseContent{
        [JsonProperty("data")]
        public Data data { get; set; }
        public ResponseContent()
        {
            data = new Data();
        }
    }

    
    // Class for the token response data.

    public class Data{
        [JsonProperty("@odata.type")]
        public string odatatype { get; set; }
        public List<Action> actions { get; set; }
        public Data()
        {
            odatatype = "microsoft.graph.onTokenIssuanceStartResponseData";
            actions = new List<Action>();
            actions.Add(new Action());
        }
    }

    
    // Action class.

    public class Action{
        [JsonProperty("@odata.type")]
        public string odatatype { get; set; }
        public Claims claims { get; set; }
        public Action()
        {
            odatatype = "microsoft.graph.tokenIssuanceStart.provideClaimsForToken";
            claims = new Claims();
        }
    }

    // Custom claims class.

    public class Claims{
        [JsonProperty(NullValueHandling = NullValueHandling.Ignore)]
        public string CorrelationId { get; set; }
        [JsonProperty(NullValueHandling = NullValueHandling.Ignore)]
        public string DateOfBirth { get; set; }
        public string ApiVersion { get; set; }
        public List<string> CustomRoles { get; set; }
        public Claims()
        {
        CustomRoles = new List<string>();
        }
    }

    // Class definition for the query sent to the SGNL Access Service.

    public class Query {
        public string action {get;set;} = "";
        public string assetId {get;set;} = "";
    }

    // Class definition for the principal sent to the SGNL access service api.

    public class AzurePrincipal {
        public string id {get;set;}="";
    }


    // Class definition for the entire request JSON sent to the SGNL access service api.

     public class SGNLRequest
    {
        public  AzurePrincipal principal { get; set; } = new AzurePrincipal{id="someone@domain.com"}; 
        public string   ipAddress   {get;set;} = "0.0.0.0";  

        public Query[] queries {get; set; } = new Query[0];
        }
