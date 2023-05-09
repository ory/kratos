local claims =
{
  email_verified: false
} + std.extVar('claims');
{
  identity:
  {
    traits:
    {
      email: claims.email,
      [if "picture" in claims then "picture" else null]: claims.picture,
      name:
      {
        first: claims.name,
        last: claims.last_name,
      }
    },
  },
}
