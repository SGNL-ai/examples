# SGNL PAM Module Example
This C example demonstrates how to implement a Linux PAM module for providing just in time access management for Linux administrators and users. This PAM module may be configured for any Linux service that requires just in time authorization capabilities.

## Caution
This is a security module that provides just in time access management for Linux services such as sudo and ssh. Please ensure you have an administrative root account to the development Linux server you are using to build and test the module. This will enable you to enable and disable the module on a Linux service such as sudo/ssh and not remain in a locked state if an unexpected module configuraiton error occurs.

# Prerequisites
1. You have a SGNL environment up and running. Please send us a message for requesting your environment.
 
2. You have configured connectors to ingest data and have defined the appropriate relationship mappings.
 
3. You have created policy snippets and policy for evaluation. You will need to create policy for the Linux service you are providing JITAM for such as sudo and ssh. Please consult with the SGNL team for guidance.
 
4. You have created an integration to test.
 
5. You have a C Linux development environment setup. You will need to install gcc and other development packages in order to comiple and build the PAM module. You can run the following commands respective to your operating system version (RHEL,Ubuntu).

    ### Red Hat
    ``` sudo yum install gcc ``` 

    ``` sudo yum install pam-devel```

    ``` sudo yum install json-c-devel```

    ``` sudo yum install libcurl-devel```

    ### Debian Based Linux (Ubuntu)
    ``` sudo apt update ```

    ``` sudo apt install gcc ```
    
    ``` sudo apt install libpam-dev```

    ``` sudo apt install libcurl4-openssl-dev```

    ``` sudo apt install libjson-c-dev ```
     

6. You will need Git to clone the example repository. You can follow the steps to install Git [here](https://github.com/git-guides/install-git).


See our [Help Guides](https://support.sgnl.ai) for steps on configuring data sources and policies.


## Config File
This example PAM module loads configuration from ```/etc/sgnl/``` sgnl_pam.json. Ensure you replace your SGNL access service URL and token. Please contact support with any questions.


## Steps For Running The PAM module


1. Ensure all the pre-requisites are met before attempting to build and install the PAM module.

2. Clone the example repo using Git from https://github.com/SGNL-ai/examples.git to your development Linux server.

3. Change to the /pam/build directory.

4. Run the build script ```./build_pam.sh```.

5. Update the SGNL PAM module configuration. Switch to the ```/etc/sgnl``` directory. Edit the sgnl_pam.json file and replace token value (vault integration under development) with your SGNL integration authentication token.

6. Enable the PAM module for a Linux service such as sudo. For sudo, modify the /etc/pam.d/sudo file and add the following line:

    ```account  required   sgnl_pam.so```

**Notes:** After you save your sudo PAM configuration, the operating system will now call the SGNL PAM module which calls the SGNL access service API to authorize the user based on the human readable policy. To disable the PAM module, you simply edit the sudo file and comment out the line that refers to the sgnl_pam.so module.

## Congratulations
You have now run an SGNL authorization query via a Linux PAM module. By doing this, you are taking advantage of the power SGNL has to extend privileged access with enterprise capabilities such as just-in-time access management, centralized policy management, audit, and reporting.




