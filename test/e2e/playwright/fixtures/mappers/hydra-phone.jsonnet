local claims = std.extVar('claims');

if std.length(claims.sub) == 0 then
  error 'claim sub not set'
else
  {
    identity: {
      traits: {
        phone: claims.sub,
      },
    },
  }
