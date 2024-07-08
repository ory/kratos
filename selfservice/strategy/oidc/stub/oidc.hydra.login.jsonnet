local claims = std.extVar('claims');
local provider = std.extVar('provider');
local identity = std.extVar('identity');
local mp = if 'metadata_public' in identity then identity.metadata_public else {};

if std.length(claims.sub) == 0 then
  error 'claim sub not set'
else
  {
    identity: {
      metadata_public: mp {
        sso_groups+: {
          [if 'groups' in claims.raw_claims then provider]: claims.raw_claims.groups,
        },
      },
    },
  }
