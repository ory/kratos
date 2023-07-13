INSERT INTO identity_credential_types (id, name) SELECT '14f3b7e2-8725-4068-be39-8a796485fd97', 'code' WHERE NOT EXISTS ( SELECT * FROM identity_credential_types WHERE name = 'code');
