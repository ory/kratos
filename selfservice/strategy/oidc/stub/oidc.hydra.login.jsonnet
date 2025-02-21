local claims = std.extVar('claims');
local provider = std.extVar('provider');
local identity = std.prune(std.extVar('identity'));
local ma = if 'metadata_admin' in identity then identity.metadata_admin else {};
local mp = if 'metadata_public' in identity then identity.metadata_public else {};

if std.length(claims.sub) == 0 then
  error 'claim sub not set'
else
  {
    identity: {
      metadata_admin: ma {
        sso_groups+: {
          [if 'groups' in claims.raw_claims then provider]: claims.raw_claims.groups,
        },
      },
      metadata_public: mp {
        sso_groups+: {
          [if 'groups' in claims.raw_claims then provider]: claims.raw_claims.groups,
        },
      },
    },
  }
