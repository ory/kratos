local claims = std.extVar('claims');
local identity = std.extVar('identity');
local metadata_public = if 'metadata_public' in identity then identity.metadata_public else {};

if std.length(claims.sub) == 0 then
  error 'claim sub not set'
else
  {
    identity: {
      metadata_public: metadata_public + { [if "groups" in claims.raw_claims then "groups" else null]: claims.raw_claims.groups },
    },
  }
