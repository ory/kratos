INSERT INTO identity_credential_types (id, name)
SELECT '7bddcf6c-f50e-4a18-9b0f-429114c33419', 'saml'
    WHERE NOT EXISTS ( SELECT * FROM identity_credential_types WHERE name = 'saml');