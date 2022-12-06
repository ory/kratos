local claims = std.extVar('claims');

if std.length(claims.sub) == 0 then
  error 'claim sub not set'
else
  {
    identity: {
      traits: {
        subject: claims.sub,
        [if "email" in claims then "email" else null]: claims.email,
      },
    },
  }
