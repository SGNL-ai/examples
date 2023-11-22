local typedefs = require "kong.db.schema.typedefs"

local schema = {
  name = "sgnlplugin",
  fields = {
    -- the 'fields' array is the top-level entry with fields defined by Kong
    { consumer = typedefs.no_consumer },  -- this plugin cannot be configured on a consumer (typical for auth plugins)
    { protocols = typedefs.protocols_http },
    { config = {
        -- The 'config' record is the custom part of the plugin schema
        type = "record",
        fields = {
          -- a standard defined field (typedef), with some customizations
          { request_header = typedefs.header_name {
              required = false,
              default = "SGNL-Hello-World" } },
          { response_header = typedefs.header_name {
              required = false,
              default = "SGNL-Bye-World" } },
          { sgnl_token = { type = "string", required = true, default = "none" } },
          { client_id = { type = "string", required = true, default = "none" } },
          { client_secret = { type = "string", required = true, default = "none" } },
          {introspect_token = { type = "boolean", required = true, default = false } },
          {introspection_endpoint = { type = "string", required = true, default = "none" } },
          {sgnl_endpoint = { type = "string", required = true, default = "https://access.sgnlapis.cloud" } }
        },
        entity_checks = {
          -- add some validation rules across fields
          -- the following is silly because it is always true, since they are both required
          { at_least_one_of = { "request_header", "response_header", "sgnl_token" }, },
          -- We specify that both header-names cannot be the same
          { distinct = { "request_header", "response_header"} },
        },
      },
    },
  },
}

return schema
