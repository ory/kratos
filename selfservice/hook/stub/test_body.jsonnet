function(ctx) std.prune({
  flow_id: ctx.flow.id,
  identity_id: if std.objectHas(ctx, "identity") then ctx.identity.id,
  headers: ctx.request_headers,
  url: ctx.request_url,
  method: ctx.request_method,
  cookies: ctx.request_cookies,
  transient_payload: if std.objectHas(ctx.flow, "transient_payload") then ctx.flow.transient_payload,
})
