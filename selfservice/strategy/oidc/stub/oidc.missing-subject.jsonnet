local claims = std.extVar('claims');

{
  identity: {
    traits: {
      website: claims.website,
    },
  },
}
