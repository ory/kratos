function(ctx) {
  flow_id: ctx.flow.id,
  session_id: if ctx["session"] != null then ctx.session.id,
  identity_id: if ctx["identity"] != null then ctx.identity.id,
  headers: ctx.request_headers,
  url: ctx.request_url,
  method: ctx.request_method
}
