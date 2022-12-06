local claims = std.extVar('claims');

if std.length(claims.sub) == 0 then
  error 'claim sub not set'
else
  {
    identity: {
      traits: {
        subject: claims.sub,
        [if "website" in claims then "website" else null]: claims.website,
        [if "groups" in claims.raw_claims then "groups" else null]: claims.raw_claims.groups,
      },
      metadata_public: {
        [if "picture" in claims then "picture" else null]: claims.picture,
      },
      metadata_admin: {
        [if "phone_number" in claims then "phone_number" else null]: claims.phone_number,
      }
    },
  }
