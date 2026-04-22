local claims = std.extVar('claims');
local identity = std.extVar('identity');

// Keep the website if the user has set one.
local website = std.get(identity.traits, "website", std.get(claims, "website", ""));

if std.length(claims.sub) == 0 then
  error 'claim sub not set'
else
  {
    identity: {
      traits: {
        subject: claims.sub,
        [if website != "" then "website" else null]: website,
        [if "groups" in claims.raw_claims then "groups" else null]: claims.raw_claims.groups,
      },
      metadata_public: {
        [if "picture" in claims then "picture" else null]: claims.picture,
      },
      metadata_admin: {
        [if "phone_number" in claims then "phone_number" else null]: claims.phone_number,
      },
      verified_addresses: [
        { via: "email", value: claims.sub },
      ],
    },
  }
