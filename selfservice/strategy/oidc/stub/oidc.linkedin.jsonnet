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
      picture: claims.picture,
      name:
      {
        first: claims.name,
        last: claims.last_name,
      }
    },
  },
}
