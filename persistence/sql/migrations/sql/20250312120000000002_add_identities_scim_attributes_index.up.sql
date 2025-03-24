CREATE INDEX identities_scim_userName_idx ON identities ((scim ->> 'userName'))
    WHERE scim IS NOT NULL AND
          scim ->> 'userName' IS NOT NULL;