INSERT INTO identity_credential_types (id, name)
SELECT '8d0ca544-9bf6-45d3-bd75-0bbb3aeba3c7', 'passkey'
WHERE NOT EXISTS ( SELECT * FROM identity_credential_types WHERE name = 'passkey');