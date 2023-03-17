function(ctx) {
    identity_id: if std.objectHas(ctx, "identity") then ctx.identity.id else null,
    email: if std.objectHas(ctx, "identity") then ctx.identity.traits.email else null,
    flow_id: ctx.flow.id,
    flow_type: ctx.flow.type,
    identity: if std.objectHas(ctx, "identity") then ctx.identity else null,
    ctx: ctx
}
