-- +migrate Up
INSERT INTO identity (id, traits_schema_url)
VALUES ('data-1', 'foo');

INSERT INTO identity_credentials (identity_id, method)
VALUES (1, 'password');

INSERT INTO identity_credentials_identifiers (identity_credentials_id, identifier)
VALUES (1, 'data-1@example.org');

INSERT INTO identity_credentials_identifiers (identity_credentials_id, identifier)
VALUES (1, 'data-1@example.com');

INSERT INTO self_service_request (id, expires_at, issued_at, request_url, request_headers, active, methods, kind)
VALUES (1, NOW(), NOW(), 'https://www.ory.sh/', '{}', 'password' , '{}', 'login');

-- +migrate Down
