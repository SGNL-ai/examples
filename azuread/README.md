# Azure AD / Entra ID App Function HTTP Trigger Example
This example uses Custom Authentication Extension in Entra Id (formerly Azure AD) to make a request for an access decision during sign-in to Applications/Services configured in Microsoft Entra. The sample can be run in Azure Functions, is written in C#, and can be modified to support additional configuration and/or metadata that may be configured for your applications.


# Prerequisites

1. You have a SGNL environment up and running.
 
2. You have configured connectors to ingest data and have defined the appropriate relationship mappings.
 
3. You have created policy snippets and policy for evaluation.
 
4. You have created an Entra ID Application to test. You can test with this [example](https://jwt.ms/) provided by Microsoft.
 
5. You will need Git to clone the example repository. You can follow the steps to install Git [here](https://github.com/git-guides/install-git).


See our [Help Guides](https://support.sgnl.ai) for steps on configuring data sources and policies.


## Steps For Configuring the Custom Authentication Extension


1. Clone the SGNL examples repo using Git from https://github.com/SGNL-ai/examples.git.
2. Login to the Azure portal.
3. Follow the steps found [here](https://learn.microsoft.com/en-us/azure/active-directory/develop/custom-extension-get-started?tabs=azure-portal%2Chttp) to configure a custom extension.
4. Replace the default C# script in the example in step 3, with the run.csx C# code from the SGNL examples repository.


## Steps For Testing the Example

You can choose one of two options for testing a custom authentication extension. The first option is your own custom OAuth 2/OIDC appplication and the second option is a pre-integrated SaaS enterprise application such as [Smartsheets](https://www.smartsheet.com/).

**Option (1):** Once your custom application configuration is complete and you have completed the configuration of the custom claims provider (see step 4 in [this article](https://learn.microsoft.com/en-us/azure/active-directory/develop/custom-extension-get-started?tabs=azure-portal%2Chttp), you can build your login URL for initiating the Azure AD IDP authentication flow. See the example URL below:
   
    https://login.microsoftonline.com/{tenant-id}/oauth2/v2.0/authorize?client_id={App_Client_ID}&response_type=id_token&redirect_uri=https://jwt.ms&scope=openid&state=12345&nonce=12345 

**Option 2:** Follow [these steps](https://learn.microsoft.com/en-us/azure/active-directory/manage-apps/add-application-portal-setup-oidc-sso) to add an OAuth 2/OpenID connect application. This example uses Smartsheets but you may use any other OAuth 2/OpenID Connect compliant application in the Azure AD Gallery. After you add the application, ensure you configure the custom claims provider as in step 4 in [this article](https://learn.microsoft.com/en-us/azure/active-directory/develop/custom-extension-get-started?tabs=azure-portal%2Chttp). Be sure to navigate to your Enterprise Applications view, select the application, and select **Single Sign On** under the **Manage** menu item. You can then edit the attributes and claims configuration to include the custom claims provider implemented by the custom authentication extension. 


## Congratulations
You have now run an SGNL authorization query from an Azure AD custom authentication extension. By doing this, you are taking advantage of the power SGNL has to extend Azure AD / Entra ID with enterprise capabilities such as continous access management, centralized policy management, audit, and reporting.



