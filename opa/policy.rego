package sgnl.authz

default allow := false
            
allow {
      response := http.send({
			"method": "POST",
			"url": "{{SGNL Access Service URL}}",
			"tls_use_system_certs": true,
			"headers": {
			"Authorization": "Bearer {{token}}",
			"Content-type": "application/json",
            },
            "body": {
	                    "principal": {
		                "id": "{{principal id}}"
	                    },
	                        "queries": [{
	                    	"assetId": "{{asset id}}",
                             "action": "Write"
	                            }]
                    },
            "force_cache": false,
            "force_json_decode" : true
			}
            )
        response.body.decisions[0].decision = "Allow"
		}