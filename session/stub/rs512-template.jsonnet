function(ctx) {
  claims: {
    foo: "bar",
    sub: "can not be overwritten",
    schema_id: ctx.session.identity.schema_id,
    amr: ctx.session.amr,
    second_claim: ctx.claims.exp,
  }
}
