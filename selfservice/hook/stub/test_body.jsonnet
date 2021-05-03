function(ctx) {
  flow_id: ctx.flow.id,
  session_id: ctx.session.id,
  identity_id: ctx.identity.id,
  headers: ctx.request_headers,
  url: ctx.request_url,
  method: ctx.request_method
}
