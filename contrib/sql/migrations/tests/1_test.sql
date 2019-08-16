-- +migrate Up
INSERT INTO identity (id, traits_schema_url)
VALUES ('data-1', 'foo');

INSERT INTO identity_credential (identity_pk, method, config)
VALUES (1, 'password', '{"foo":"bar"}');

INSERT INTO identity_credential_identifier (identity_credential_pk, identifier)
VALUES (1, 'data-1@example.org');

INSERT INTO identity_credential_identifier (identity_credential_pk, identifier)
VALUES (1, 'data-1@example.com');

INSERT INTO self_service_request (id, expires_at, issued_at, request_url, request_headers, active, methods, kind)
VALUES (1, NOW(), NOW(), 'https://www.ory.sh/', '{}', 'password', '{}', 'login');

INSERT INTO session (sid, expires_at, issued_at, authenticated_at, identity_pk)
VALUES ('sid-1', NOW(), NOW(), NOW(), 1);

-- +migrate Down
