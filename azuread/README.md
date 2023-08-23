# Azure AD / Entra ID App Function HTTP Trigger Example
This example uses Custom Authentication Extension in Entra Id (formerly Azure AD) to make a request for an access decision during sign-in to Applications/Services configured in Microsoft Entra. The sample can be run in Azure Functions, is written in C#, and can be modified to support additional configuration and/or metadata that may be configured for your applications.


# Prerequisites

1. You have a SGNL environment up and running.
 
2. You have configured connectors to ingest data and have defined the appropriate relationship mappings.
 
3. You have created policy snippets and policy for evaluation.
 
4. You have created an Entra ID Application to test. You can test with this [example](https://jwt.ms/) provided by Microsoft.
 
5. You will need Git to clone the example repository. You can follow the steps to install Git [here](https://github.com/git-guides/install-git).


See our [Help Guides](https://support.sgnl.ai) for steps on configuring data sources and policies.


## Steps For Configuring the custom authentication extension.


1. Clone the SGNL examples repo using Git from https://github.com/SGNL-ai/examples.git.
2. Login to the Azure portal.
3. Follow the steps found [here](https://learn.microsoft.com/en-us/azure/active-directory/develop/custom-extension-get-started?tabs=azure-portal%2Chttp) to configure a custom extension.
4. Replace the default C# script in the example in step 3, with the run.csx C# code from the SGNL examples repository.


## Steps For Testing The Example

1. Once your configuration is complete, you can build your login URL for initiating the Azure AD IDP authentication flow. See the example URL below:
   
    `` https://login.microsoftonline.com/{tenant-id}/oauth2/v2.0/authorize?client_id={App_Client_ID}&response_type=id_token&redirect_uri=https://jwt.ms&scope=openid&state=12345&nonce=12345``


## Congratulations
You have now run an SGNL authorization query from an Azure AD custom authentication extension. By doing this, you are taking advantage of the power SGNL has to extend Azure AD / Entra ID with enterprise capabilities such as continous access management, centralized policy management, audit, and reporting.



