		---------------------------------------------------------------------------------------------
		-- This plugin is an access plugin that calls the SGNL Access Service
    -- for making authorization decisions. The handler is based on the OpenResty handlers, please refer to the OpenResty
    -- documentation for more details.
		---------------------------------------------------------------------------------------------
		
    local kong_response = kong.response
    local http = require "resty.http"
    local jsonparser = require "cjson"
    local jwt = require 'resty.jwt'

		local plugin = {
		  PRIORITY = 1000, -- Plugin priority, which determines plugin execution order.
		  VERSION = "0.1", -- Plugin version in X.Y.Z format. Check hybrid-mode compatibility requirements.
		}
  		
	
		-- This function handles more initialization, but AFTER the worker process has been forked/created.
		function plugin:init_worker()

		  -- Say hello!
		  kong.log.info("Hello 'SGNL plugin init_worker' handler.")

		end --]]
		
		
		-- This function runs in the 'access_by_lua_block'
		function plugin:access(plugin_conf)
		      
    -- Get request headers for any request context needed.
    local headers = kong.request.get_headers()

    -- Get Curity Bearer Token
    local auth_header = headers.Authorization

    -- Uncomment for debug purposes.
    -- kong.log.info("Bearer token in request: "..auth_header)

    -- Get bearer token from authorization header
    local access_token_untrimmed = string.sub(auth_header, 8)
    local access_token = string.gsub(access_token_untrimmed, "%s+", "")
    
            
     if not plugin_conf.introspect_token then
        local decoded_sub = decode_token(access_token)
        decoded_sub = '"'.. decoded_sub..'"'
        
        --Authorization
        if not authorized(decoded_sub,plugin_conf) then
          return kong_response.exit(403, "SGNL Unauthorized")
        end
     end
      
     -- Introspect access token with Curity Identity Server
     if plugin_conf.introspect_token then
      local result = Introspect_access_token(access_token, plugin_conf)
      
     -- Uncomment for debug purposes
     --  kong.log.info("Introspection result: "..result.body)
        if result.status ~= 200 then
          return kong_response.exit(403, "SGNL Plugin: Error occured while introspecting token. HTTP status code from IDP: "..result.status) 
        else
        
        local introdata = jsonparser.decode(result.body)
        local sub = jsonparser.encode(introdata.sub)
        local active = jsonparser.encode(introdata.active)
        
        if active == "false" then
          kong.log.info("SGNL Plugin: Access token is NOT active.")
         return kong_response.exit(403, "Token introspection error. Access token in request is not valid.")
       end

        --Authorization
        if not authorize(sub,plugin_conf) then
          return kong_response.exit(403, "SGNL Unauthorized")
        end

       end
      end        
		end --]]
		
		
		-- This function runs in the 'header_filter_by_lua_block'.
		function plugin:header_filter(plugin_conf)
		
		  -- Send back a sample header with a message so the client knows the SGNL plugin is enabled.
		  kong.response.set_header(plugin_conf.response_header, "SGNL authorization plugin is enabled.")
		
		end --]]
		
		--
-- Introspect the Curity access token
--
 function Introspect_access_token(access_token, config)
 
  local httpc = http:new()
  local introspectCredentials = ngx.encode_base64(config.client_id .. ':' .. config.client_secret)
  local result, error = httpc:request_uri(config.introspection_endpoint, {
      method = 'POST',
      body = 'token=' .. access_token,
      headers = { 
          ['authorization'] = 'Basic ' .. introspectCredentials,
          ['content-type'] = 'application/x-www-form-urlencoded',
          ['accept'] = 'application/json'
      },
      ssl_verify = false
  })
  
  if result.status ~= 200 then
      kong.log.info("[SGNL Plugin:Introspect]Introspection failed with status: "..result.status)
      return { status = result.status }
  end

  
  return { status = result.status, body = result.body, sub = result.body.sub }
end	

function decode_token(access_token)
  local jwt_obj = jwt:load_jwt(access_token)
  if jwt_obj.valid then
    kong.log.info("SGNL Plugin: Extracted subject from valid JWT.")
    return jwt_obj.payload.sub
  end
end

function authorized(principal, conf)
   local cjson2 = jsonparser.new()
   local client = http.new()

    -- Ensure subject does not contain any extra string quotations   
    kong.log.info("Sub string: ".. principal)

    -- Build the JSON document for the request to the SGNL Access Service API.

    local json_text = '{"principal":{"id": '..principal..'},"queries":[{"assetId":"'..string.gsub(kong.request.get_path(),"/records/","")..'" ,"action":"'..kong.request.get_method()..'"}]}'

    -- Uncomment for debug purposes
    kong.log.info("SGNL Query: "..json_text)
 
    -- Set the HTTP client timeouts.
     client:set_timeouts(10000, 60000, 60000)
    
     kong.log.info("SGNL endpoint: "..conf.sgnl_endpoint)
     -- Configure HTTP client and make the request to the SGNL Access Service API.
 
     local res, err = client:request_uri(conf.sgnl_endpoint, {
       method = "POST",
       path = "/access/v2/evaluations",
       headers = {
         ["Authorization"] = "Bearer "..conf.sgnl_token,
         ["Content-Type"] = "application/json"
       },
       body = json_text
     })

        -- JSON decode resopnse body, encode decisions, and decode once more to build a JSON object.
        local jdata = jsonparser.decode(res.body)
        local decisions = jsonparser.encode(jdata.decisions)
        local jsondecisions = jsonparser.decode (decisions)
        
        -- Uncomment for debugging SGNL response:
        kong.log.info("SGNL response: "..res.body)
  
        -- Check the first element in the decisions array. If more than one decision, you can iterate through array and check all decisions.
        -- This example only returns one decision at index 1. Note: LUA array indexes start at 1.
        
        if jsondecisions[1].decision ~= 'Allow' then
            return false 
          else
            return true
        end
 end
  -- Return the plugin object.
	return plugin

