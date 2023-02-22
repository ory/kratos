INSERT INTO identity_credential_types (id, name) SELECT 'ff5a1823-8b47-4255-860f-4b70ed122740', 'saml' WHERE NOT EXISTS ( SELECT * FROM identity_credential_types WHERE name = 'saml');
