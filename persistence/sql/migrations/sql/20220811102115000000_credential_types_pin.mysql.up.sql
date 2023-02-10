INSERT INTO identity_credential_types (id, name) SELECT 'b2f010e4-0126-4db2-b24d-8ea390d1e25f', 'pin' WHERE NOT EXISTS ( SELECT * FROM identity_credential_types WHERE name = 'pin')
