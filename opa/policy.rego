package sgnl.authz

default allow := false
            
allow {
      response := http.send({
			"method": "POST",
			"url": "https://access.sgnlapis.cloud/access/v1/evaluations",
			"tls_use_system_certs": true,
			"headers": {
			"Authorization": "Bearer eyJkYXRhIjoiU1NIL25nYWdUdTlTOG1LVk15elNpNVQ3NVBIOFpWSHQ0ODNyT1luczhXUm1XUm5kdW1TSkY4MFlYQU5CUjJaSzgxY2IyNWE3aGhvUHRXRms4M2dmTkE9PSIsImlkIjoiNDlkMjExNTctMDVlOC00ZDdiLTg0ZDctZmE1ZjBhNTA2Yjk2IiwiaW50ZWdyYXRpb25JZCI6IjZmNTBkN2Y5LTZmZTItNDg5Yi1hODgwLWIwZGZkM2JhNDk2NSIsInRlbmFudElkIjoiNzU1MDY0OTItYjZhYi00M2Q0LWFmZGItNWE1MTQ1MTI4YTVhIn0=",
			"Content-type": "application/json",
            },
            "body": {
	                    "principal": {
		                "id": "aldo@sgnl.ai.sandbox"
	                    },
	                        "queries": [{
	                    	"assetId": "fb4c0e53-0778-45cd-a10b-85b10ee61ac7",
                             "action": "Write"
	                            }]
                    },
            "force_cache": false,
            "force_json_decode" : true
			}
            )
        response.body.decisions[0].decision = "Allow"
		}