local claims = std.extVar('claims');
local identity = std.extVar('identity');

if std.length(claims.sub) == 0 then
  error 'claim sub not set'
else
  {
    identity: {
      traits: {
        subject: claims.sub,
        [if "website" in claims then "website" else null]: claims.website,
        // Keep existing groups if the claims don't provide them.
        [if "groups" in claims.raw_claims then "groups" else null]:
          claims.raw_claims.groups,
      },
      metadata_public: {
        [if "picture" in claims then "picture" else null]: claims.picture,
        // Preserve existing metadata that isn't in the claims.
        [if "existing_field" in identity.metadata_public then "existing_field" else null]:
          identity.metadata_public.existing_field,
      },
      metadata_admin: {
        [if "phone_number" in claims then "phone_number" else null]: claims.phone_number,
      },
    },
  }
