/**
 * C code for an example SGNL PAM module.
 * Author: Aldo Pietropaolo
 * Last Update: 4-4-2023
 *
 * Src: sgnl_pam.c
 */


#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <security/pam_appl.h>
#include <security/pam_modules.h>
#include <syslog.h>
#include <security/_pam_macros.h>
#include <security/pam_modules.h>
#include <security/pam_ext.h>
/* libcurl (https://curl.se/libcurl/c/) */
#include <curl/curl.h>
/* json-c (https://github.com/json-c/json-c) */
#include <json-c/json.h>

/* Struct for the call_sgnl_api function. */
struct curl_post_st {
    char *payload;
    size_t size;
};

/* Structs for the SGNL config. */

struct json_object *sgnl_config;
struct json_object *sgnl_url;
struct json_object *sgnl_token;
struct json_object *idp_domain;

void read_config()
{
  FILE *fp;
  char buffer[1024];
  fp = fopen("/etc/sgnl/sgnl_pam.json","r");
  fread(buffer,1024,1,fp);
  fclose(fp);
 
  sgnl_config = json_tokener_parse(buffer);

  json_object_object_get_ex(sgnl_config, "url", &sgnl_url);
  json_object_object_get_ex(sgnl_config, "token", &sgnl_token); 
  json_object_object_get_ex(sgnl_config, "idp_domain", &idp_domain);
}


/* Callback for the  call_sgnl_api function.  */
size_t sgnl_curl_callback (void *contents, size_t size, size_t nmemb, void *userp) {
    size_t realsize = size * nmemb;                             
    struct curl_post_st *p = (struct curl_post_st *) userp;   

    /* Expand buffer using a temporary pointer to avoid memory leaks */
    char * temp = realloc(p->payload, p->size + realsize + 1);

    /* Check memory allocation */
    if (temp == NULL) {
      /* Oh no! */
      /* free buffer */
      free(p->payload);
      /* return */
      return 1;
    }

    /* assign payload */
    p->payload = temp;

    /* copy contents to buffer */
    memcpy(&(p->payload[p->size]), contents, realsize);

    /* set new buffer size */
    p->size += realsize;

    /* ensure null termination */
    p->payload[p->size] = 0;

    /* return size */
    return realsize;
}

/* Post request to the SGNL API and return the response. */
CURLcode call_sgnl_api(CURL *ch, const char *url, struct curl_post_st *post) {
   /*CURL result code.*/
    CURLcode rcode;                

    /* Initialize the POST payload */
    post->payload = (char *) calloc(1, sizeof(post->payload));

    /* Check the payload */
    if (post->payload == NULL) {
        /* log error */
        /* return error */
        return CURLE_FAILED_INIT;
    }

    /* Initialize POST struct size. */
    post->size = 0;

    /* Set the url to get */
    curl_easy_setopt(ch, CURLOPT_URL, url);

    /* Set the CURL calback function. */
    curl_easy_setopt(ch, CURLOPT_WRITEFUNCTION, sgnl_curl_callback);

    /* Pass the post struct pointer to the CURL lib. */
    curl_easy_setopt(ch, CURLOPT_WRITEDATA, (void *) post);

    /* Set the  sgnl pam user agent. */
    curl_easy_setopt(ch, CURLOPT_USERAGENT, "sgnlpam-agent/1.0");

    /* Set a timeout.*/
    curl_easy_setopt(ch, CURLOPT_TIMEOUT, 10);

    /*Make the call to the SGNL API. */
    rcode = curl_easy_perform(ch);

    /* return */
    return rcode;
}

/* Function to call the SGNL Access API and return decision to the PAM account function. */

int callSGNL (const char *username, const char *service){
         CURL *curl;
         CURLcode rcode;
         json_object *json;                                      
         json_object *qobj;
         json_object *pobj;
         json_object *dobj;
	 json_object *error;
	 json_object *errcode;
         struct curl_post_st curl_post;                        
         struct curl_post_st *cp = &curl_post;                 
         struct curl_slist *headers = NULL;                     
         const char *decision_str = NULL;
         char *allow = "Allow";
         char *deny = "Deny";

        /* Get information from config.*/
        const char *url = json_object_get_string(sgnl_url);
	const char *token = json_object_get_string(sgnl_token);

        /* Initialize the curl handle. */
        if ((curl = curl_easy_init()) == NULL) {
                /* return error */
                return 500;
        }

        /* set content type */
        headers = curl_slist_append(headers, "Accept: application/json");
        headers = curl_slist_append(headers, "Content-Type: application/json");
        headers = curl_slist_append(headers,token);

        /* Create json-c objects. */
        json = json_object_new_object();
        pobj = json_object_new_object();
        qobj = json_object_new_object();
        dobj = json_object_new_object();
	error = json_object_new_object();

        /* Build the principal object.*/
	json_object_object_add(pobj, "id", json_object_new_string(username));

	 /*Creating a json array*/
        json_object *qarray = json_object_new_array();

        /*Creating json strings*/
        json_object_object_add(qobj, "assetId", json_object_new_string("Linux File"));
        json_object_object_add(qobj, "action", json_object_new_string(service));

        /*Adding the above created json strings to the array*/
        json_object_array_add(qarray,qobj);

        /* build post data */
        json_object_object_add(json, "principal",pobj);
        json_object_object_add(json,"queries",qarray);


        /* set curl options */
        curl_easy_setopt(curl, CURLOPT_CUSTOMREQUEST, "POST");
        curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);
        curl_easy_setopt(curl, CURLOPT_POSTFIELDS, json_object_to_json_string(json));

        /* Call the SGNL Access Service API and get the return code */
        rcode = call_sgnl_api(curl, url, cp);

        /* Cleanup the curl handle.We don't want any memory issues.*/
        curl_easy_cleanup(curl);

        /* Free the allocated headers.*/
        curl_slist_free_all(headers);

        /* Free the json object.*/
        json_object_put(json);

         /* check return code */
           if (rcode != CURLE_OK || cp->size < 1) {
                /* Return error code.*/
                return 500;
            }

	    /* check payload */
           if (cp->payload != NULL) {

                /* Parse SGNL response and extract errors or decisions.*/
                dobj = json_tokener_parse(cp->payload);

	        /* Free the payload in memory. We don't want memory issues. */
                curl_free(cp->payload);

		/* Check for errors. */
		error = json_object_object_get(dobj, "error");
               
	        /* If there is an error returned from the SGNL API, return to PAM handle log.*/	
		if (json_object_get_string(error) != NULL){
                 errcode = json_object_object_get(error, "code"); 
		 printf("SGNL API error code: %i \n",json_object_get_int(errcode));
		 return json_object_get_int(errcode);
		} else {

		/* Get the decisions array.*/ 
                json_object *decisions = json_object_object_get(dobj, "decisions");

                /* Count items in array for iterator.*/
                int decision_count = json_object_array_length(decisions);
                
                /* Return error if array count is zero.*/

                if (decision_count == 0){
                        return 500;
                } else {

                /*Iterate through decisions*/
                for (int i = 0; i < decision_count; i++)
                         {
                        json_object *element = json_object_array_get_idx(decisions, i);
                        json_object *decision = json_object_object_get(element, "decision");
                        decision_str = json_object_get_string(decision);
                         }

                 /*Check if decision is allow or deny. If allow, respond with true, if deny false*/

                if(strcmp(decision_str,allow)==0) {
                                return 1;
                            } else if(strcmp(decision_str,deny)==0) {
                                return 0;
                                }
                /* Free payload memory. We don't want memory issues.*/ 
                free(cp->payload);

              }
             }

      }
	    /* If we can't determine a decision, always respond with deny. Deny by default */

           return 0;
}



/* Expected hook  by the PAM implementation*/
PAM_EXTERN int pam_sm_setcred( pam_handle_t *pamh, int flags, int argc, const char **argv ) {
	/*we are only implementing the account interface.*/
        pam_syslog(pamh, LOG_INFO, "SGNL PAM Module (pam_sm_setcr (pam_sm_setcred). Return PAM_SUCCESS.");
	return PAM_SUCCESS;
}

/* Expected hook by the PAM implementation. This is where the call to SGNL happens. */
PAM_EXTERN int pam_sm_acct_mgmt(pam_handle_t *pamh, int flags, int argc, const char **argv) {
        const char* username;
	const char* host = NULL;
        const char* service = NULL;
	int res = 0;

	/* Read SGNL configuration file. */
	read_config();

	(void) pam_get_item(pamh, PAM_SERVICE,(const void **)  &service);
        (void) pam_get_item(pamh, PAM_RHOST,  (const void **) &host);
        (void) pam_get_user(pamh, &username, "Username: ");
	pam_syslog(pamh, LOG_INFO, "SGNL PAM Module: username [%s]", username);
        pam_syslog(pamh, LOG_INFO, "SGNL PAM Module: Host: %s",host);
	pam_syslog(pamh, LOG_INFO, "SGNL PAM Module: Linux service: %s",service);

	/* Call SGNL for authorization */
	res = callSGNL(username,service);

      if(res != 1 || 0 ){
	/* Oh no! something bad happend. Log error and return an error to the PAM handle.*/      
        pam_syslog(pamh, LOG_INFO, "SGNL PAM Module: Received error from SGNL API %i",res);      
	pam_syslog(pamh, LOG_INFO, "SGNL PAM Module: Please check configuration in the sgnl_pam.json file in /etc/sgnl.");
	return PAM_PERM_DENIED;
        }
        else {	

	pam_syslog(pamh, LOG_INFO, "Called SGNL: %i", res);

	/* Send and allow or deny to the PAM handle. */

	if (res == 1){
	   return PAM_SUCCESS;

	} else if (res == 0){
	return PAM_PERM_DENIED;
	}
      }
        /* Deny by default if we did not receive a decision from SGNL. */

	return PAM_PERM_DENIED;
}

/* Expected hook by the PAM implementation,*/
PAM_EXTERN int pam_sm_authenticate( pam_handle_t *pamh, int flags,int argc, const char **argv ) {
	/*we are only implementing the account interface.*/
        pam_syslog(pamh, LOG_INFO, "SGNL PAM Module (pam_sm_setcr (pam_sm_authenticate). Return PAM_SUCCESS.");
	return PAM_SUCCESS;
}
