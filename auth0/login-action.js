/**
* Handler that will be called during the execution of a PostLogin flow.
*
* @param {Event} event - Details about the user and the context.
* @param {PostLoginAPI} api - Auth0 Interface.
*/
exports.onExecutePostLogin = async (event, api) => {
  // Require the axios library
  const axios = require('axios');

  // Create an instance of Axios HTTP client

  const sgnlAPI = axios.create({
    baseURL: event.secrets.sgnl_url,
    timeout: 1000,
    headers: {'Authorization': "Bearer " + event.secrets.Token}
  });

  if (isValidUrl(event.secrets.sgnl_url)) {
    
    // Call SGNL access service API
    
    await sgnlAPI.post( event.secrets.sgnl_url, {
      "principal": {
        "id": event.user.email
      },
      "queries": [{
        "assetId": event.request.query.redirect_uri,
        "action": event.request.method
      }]
    })
      .then(function(response) {
          // Uncomment line below to print the SGNL authorization decision to the console log.
          // console.log("SGNL policy decision: " + response.data.decisions[0].decision)

          // Deny access if SGNL policy results in a deny.
          if (response.data.decisions[0].decision == "Deny") {
          api.access.deny(`SGNL Authorization: Access to ${event.client.name} is not allowed.`);
          } else {
          // Success. Allow the login flow to continue.
          }
          // uncomment for debug purposes.
          // console.log(response);
        })
        .catch(function (error) {
          console.log(error);
          api.access.deny(`SGNL Authorization: Access to ${event.client.name} is not allowed.`);
        });
  } else {
      console.log(new Error('SGNL URL validation failed.'));
      api.access.deny(`SGNL Authorization: Access to ${event.client.name} is not allowed.`);
    }

  /**
 * Check for a valid SGNL URL.
 * @param {url} string
 * @return {boolean}
 */
  function isValidUrl(string) {
    try {
      const secretUrl = new URL(string);
      
      if (secretUrl.host === 'access.sgnlapis.cloud' && secretUrl.protocol === 'https:') {
          return true;
      } else {
        return false
      }
    } catch (err) {
      console.log(new Error('Error occured during URL validation.'));
      return false;
    }
  }
};
