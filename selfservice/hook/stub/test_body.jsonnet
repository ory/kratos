function(ctx) {
  flow_id: ctx.flow.id,
  identity_id: if std.objectHas(ctx, "identity") then ctx.identity.id,
  headers: ctx.request_headers,
  url: ctx.request_url,
  method: ctx.request_method
}
