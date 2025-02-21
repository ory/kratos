function(ctx) {
  to: ctx.To,
  code: if std.objectHas(ctx, "VerificationCode") then ctx.VerificationCode else ctx.Code,
}
