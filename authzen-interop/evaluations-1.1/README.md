# SGNL AuthZEN v1.1 Demo Implementation
This directory contains the source code of a Google Cloud Function that connects a subset of the proposed OpenID AuthZEN API v 1.1 to the SGNL Access API. It currently only supports the "/evaluations" endpoint of the AuthZEN API

## Running the Code
Simply copy the code from the `index.js` file into a Google Cloud Function that is configured to use Node.js. To run the function provide the bearer token required for the Access API as a bearer token in the authorization header of the HTTP request to this function.

