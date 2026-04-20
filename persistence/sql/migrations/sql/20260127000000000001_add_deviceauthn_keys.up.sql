INSERT INTO identity_credential_types (id, name) SELECT '26433aa2-c33a-4870-85dc-230af354e0bf', 'deviceauthn' WHERE NOT EXISTS ( SELECT * FROM identity_credential_types WHERE name = 'deviceauthn');
