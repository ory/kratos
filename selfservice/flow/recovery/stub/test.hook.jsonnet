function(ctx) std.prune({
  userId: ctx.identity.id,
  traits: {
    email: ctx.identity.traits.email,
  },
})
