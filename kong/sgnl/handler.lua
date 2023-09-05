		---------------------------------------------------------------------------------------------
		-- This plugin is an access plugin that calls the SGNL Access Service 
    -- for making authorization decisions. The handler is based on the OpenResty handlers, please refer to the OpenResty
    -- documentation for more details.
		---------------------------------------------------------------------------------------------
		
    local kong_response = kong.response
    local http = require "resty.http"
    local jsonparser = require "cjson"

		local plugin = {
		  PRIORITY = 1000, -- Plugin priority, which determines plugin execution order.
		  VERSION = "0.1", -- Plugin version in X.Y.Z format. Check hybrid-mode compatibility requirements.
		}
		
		
		-- This function handles more initialization, but AFTER the worker process has been forked/created.
		function plugin:init_worker()
		
		  -- Say hello!
		  kong.log.info("Hello 'SGNL plugin init_worker' handler")
		
		end --]]
		
		
		-- This function runs in the 'access_by_lua_block'
		function plugin:access(plugin_conf)
		  local cjson2 = jsonparser.new()
      local client = http.new()
      
      -- Get request headers for any request context needed.
      local headers = kong.request.get_headers()
      
      -- Build the JSON document for the request to the SGNL Access Service API.
      local json_text = '{"principal":{"id": "'..headers.principal..'"},"queries":[{"assetId":"'..kong.request.get_path()..'" ,"action":"'..kong.request.get_method()..'"}]}'
      
     -- Set the HTTP client timeouts.
      client:set_timeouts(10000, 60000, 60000)
            
      -- Configure HTTP client and make the request to the SGNL Access Service API.
      local res, err = client:request_uri("https://access.sgnlapis.cloud", {
        method = "POST",
        path = "/access/v2/evaluations",
        headers = {
          ["Authorization"] = "Bearer {bearer token}",
          ["Content-Type"] = "application/json"
        },
        body = json_text
      })
      
      if not res then
        ngx.ERR(ngx.NOTICE, "Request to SGNL failed: ",err)
        return
      end
      
       -- JSON decode resopnse body, encode decisions, and decode once more to build a JSON object.
      local jdata = jsonparser.decode(res.body)
      local decisions = jsonparser.encode(jdata.decisions)
      local jsondecisions = jsonparser.decode (decisions)
      
      -- Check the first element in the decisions array. If more than one decision, you can iterate through array and check all decisions.
      -- This example only returns one decision at index 1. Note: LUA array indexes start at 1.
      if jsondecisions[1].decision ~= 'Allow' then
          return kong_response.exit(403, "SGNL Unauthorized")
		  end
      
   
		end --]]
		
		
		-- This function runs in the 'header_filter_by_lua_block'.
		function plugin:header_filter(plugin_conf)
		
		  -- Send back a sample header with a message so the client knows the SGNL plugin is enabled.
		  kong.response.set_header(plugin_conf.response_header, "SGNL authorization plugin is enabled.")
		
		end --]]
		
		
		-- Return the plugin object.
		return plugin