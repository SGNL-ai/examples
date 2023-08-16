/**
* Handler that will be called during the execution of a PostLogin flow.
*
* @param {Event} event - Details about the user and the context in which they are logging in.
* @param {PostLoginAPI} api - Interface whose methods can be used to change the behavior of the login.
*/
exports.onExecutePostLogin = async (event, api) => {
  const axios = require('axios');
  
  // Create an instance of Axios HTTP client

  const sgnl_api = axios.create({
  baseURL: event.secrets.sgnl_url,
  timeout: 1000,
  headers: {'Authorization': event.secrets.Token}
  });

  // Output requested resource to log for debug
  console.log("URI: " + event.request.query.redirect_uri)

  // Call SGNL access service API
  await sgnl_api.post( event.secrets.sgnl_url,{
	"principal": {
		"id": event.user.email
	},
	"queries": [{
		"assetId": event.request.query.redirect_uri,
        "action": event.request.method
	}]
})
  .then(function (response) {
    console.log("SGNL policy decision: " + response.data.decisions[0].decision)
   
    //Deny access if SGNL policy results in a deny.
    if (response.data.decisions[0].decision == "Deny") {
    api.access.deny(`SGNL Authorization: Access to ${event.client.name} is not allowed.`);
    api.access.deny
  } else {
    // Success
  }
   // uncomment for debug purposes
    console.log(response);
  })
  .catch(function (error) {
    console.log(error);
  });
};
