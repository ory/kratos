function(ctx) {
  code: ctx.verification_code,
  [if "transient_payload" in ctx then "transient_payload"]: ctx.transient_payload
}
