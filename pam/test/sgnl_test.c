/**
 * Test  C code using libcurl and json-c
 * to post and return a payload  from the SGNL Access Service API.
 *
 * json-c - https://github.com/json-c/json-c
 * libcurl - http://curl.haxx.se/libcurl/c
 *
 * Build:
 *
 * gcc curltest.c -lcurl -ljson-c -o curltest
 *
 * Run:
 *
 * ./sgnl_test
 * 
 */

/* standard includes */
#include <stdio.h>
#include <stdlib.h>
#include <stdarg.h>
#include <string.h>

/* json-c (https://github.com/json-c/json-c) */
#include <json-c/json.h>

/* libcurl (http://curl.haxx.se/libcurl/c) */
#include <curl/curl.h>

/* Struct for curl post. */
struct curl_post_st {
    char *payload;
    size_t size;
};

/* Callback for the curl post. */
size_t curl_post_callback (void *contents, size_t size, size_t nmemb, void *userp) {
    size_t realsize = size * nmemb;                             
    struct curl_post_st *p = (struct curl_post_st *) userp;   

    /* Use a temporary pointer and reallocate size to avoid memory leaks. */
    char * temp = realloc(p->payload, p->size + realsize + 1);

    /* Check memory  allocation. */
    if (temp == NULL) {
      /* Oh no! */
      printf( "ERROR: Failed to expand buffer in curl_post_callback");
      curl_free(p->payload);
      return 1;
    }

    /* Assign the payload from temp. */
    p->payload = temp;

    /* Copy the contents to the buffer. */
    memcpy(&(p->payload[p->size]), contents, realsize);

    /* Set the new buffer size. */
    p->size += realsize;

    /* Ensure null term.*/
    p->payload[p->size] = 0;

    /* Return payload size. */
    return realsize;
}

/* POST  and return response from the SGNL API. */
CURLcode curl_post_url(CURL *ch, const char *url, struct curl_post_st *post) {
    CURLcode rcode;                   
    post->payload = (char *) calloc(1, sizeof(post->payload));

    if (post->payload == NULL) {
        /* log error */
        printf( "ERROR: Failed to allocate payload in curl_post_url.");
        /* return error */
        return CURLE_FAILED_INIT;
    }

    /* Initialize  size. */
    post->size = 0;

    /* Set url for the POST. */
    curl_easy_setopt(ch, CURLOPT_URL, url);

    /* Set the calback function to curl_post_callback. */
    curl_easy_setopt(ch, CURLOPT_WRITEFUNCTION, curl_post_callback);

    /* Pass  the post struct pointer. */
    curl_easy_setopt(ch, CURLOPT_WRITEDATA, (void *) post);

    /* Set the default user agent. Used to identify the PAM agent. */
    curl_easy_setopt(ch, CURLOPT_USERAGENT, "sgnl-pam-agent/1.0");

    /* Set the timeout to 1 second. */
    curl_easy_setopt(ch, CURLOPT_TIMEOUT, 1);

    /* POST to the SGNL Access Service API. */
    rcode = curl_easy_perform(ch);

    /* return */
    return rcode;
}



int main(int argc, char *argv[]) {
    /* This is the libcurl handle.*/	
    CURL *ch;                                               
    /* Curl return code. */
    CURLcode rcode;                                         
    json_object *json;                                      
    json_object *qobj;
    json_object *pobj;
    json_object *dobj;
    int retval = 0;
    struct curl_post_st curl_post;                        
    struct curl_post_st *cf = &curl_post;                 
    /*Headers to send. */
    struct curl_slist *headers = NULL;                      
    struct json_object *tmp,*tmp2 ;

    /* Url for SGNL API. */
    char *url = "https://access.sgnlapis.cloud/access/v1/evaluations";

    if ( argc <= 3) 
    {
	printf("ERROR:Check parameters: [principalId] [assetId] [action]\n");
	exit(1);
    }

    /* Initialize the lib curl handle. */
    if ((ch = curl_easy_init()) == NULL) {
        printf("ERROR: Failed to create curl handle in main().");
        /* return error */
	exit(1);
    }

    /* Set headers to send to the SGNL API. */
    headers = curl_slist_append(headers, "Accept: application/json");
    headers = curl_slist_append(headers, "Content-Type: application/json");
    headers = curl_slist_append(headers, "Authorization:Bearer eyJkYXRhIjoibFdHbHVvV3gzRllmYVlZSXdLMmdTemhNSkRDbkFSb0pmSUJWVDZLMGZFRkxQbUJ5T3lXd3pSUjZPdGRQdHhwOHNlUjZyK2Q5MkVzT1dxQUFVMkJhTmc9PSIsImlkIjoiZjlmNWY2YzktMWJmOS00YzkwLTg0OWEtMjJiZTYxMGU0NzA1IiwiaW50ZWdyYXRpb25JZCI6IjY5ZTE5NThmLTQ5N2EtNDY5Ni1hYjBiLWVlNTY1ZTEwNmI1ZSIsInRlbmFudElkIjoiNzU1MDY0OTItYjZhYi00M2Q0LWFmZGItNWE1MTQ1MTI4YTVhIn0=");

    /* Create the json objects for the post. */
    json = json_object_new_object();
    pobj = json_object_new_object();
    qobj = json_object_new_object();
    dobj = json_object_new_object();

    /* Build the principal object.*/
    json_object_object_add(pobj, "id", json_object_new_string(argv[1]));

    /*Creating a json array*/
    json_object *qarray = json_object_new_array();
    
    /*Creating json strings*/
    json_object_object_add(qobj, "assetId", json_object_new_string(argv[2]));
    json_object_object_add(qobj, "action", json_object_new_string(argv[3]));

    /*Adding the above created json strings to the array*/
    json_object_array_add(qarray,qobj);

    /* Build the post data. */
    json_object_object_add(json, "principal",pobj);
    json_object_object_add(json,"queries",qarray );


    /* Set the curl options. */
    curl_easy_setopt(ch, CURLOPT_CUSTOMREQUEST, "POST");
    curl_easy_setopt(ch, CURLOPT_HTTPHEADER, headers);
    curl_easy_setopt(ch, CURLOPT_POSTFIELDS, json_object_to_json_string(json));

    /* Make post and get the return code. */
    rcode = curl_post_url(ch, url, cf);

    /* Cleanup the curl handle. */
    curl_easy_cleanup(ch);

    /* Free the headers. */
    curl_slist_free_all(headers);

    /* Free the json objects. */
    json_object_put(json);

    /* Check the return code. */
    if (rcode != CURLE_OK || cf->size < 1) {
        printf("Error: Did not receive a CURLE_OK in main()");
	exit(1);
    }
    /* Check the payload. */
    if (cf->payload != NULL) {
        printf("SGNL API Returned: \n%s\n", cf->payload);

        /*Parse the return. */
        dobj = json_tokener_parse(cf->payload);
	json_object *decisions = json_object_object_get(dobj, "decisions");
        int decision_count = json_object_array_length(decisions);

	/*Iterate through decisions.*/
	for (int i = 0; i < decision_count; i++)
	{
    		json_object *element = json_object_array_get_idx(decisions, i);
    		json_object *decision = json_object_object_get(element, "decision");
    		const char *decision_str = json_object_get_string(decision);
		printf("SGNL Decision: %s\n",decision_str);
	}
	free(cf->payload);
        /* exit */
        return 0;
}
}
