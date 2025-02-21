function(ctx) {
  code: ctx.login_code,
  [if "transient_payload" in ctx then "transient_payload"]: ctx.transient_payload
}
