-- +migrate Up
INSERT INTO identity (id, traits_schema_url)
VALUES ('data-2', 'foo');

INSERT INTO identity_credential (identity_pk, method, config)
VALUES (2, 'password', '{"foo":"bar"}');

INSERT INTO identity_credential_identifier (identity_credential_pk, identifier)
VALUES (2, 'data-2@example.org');

INSERT INTO identity_credential_identifier (identity_credential_pk, identifier)
VALUES (2, 'data-2@example.com');

INSERT INTO self_service_request (id, expires_at, issued_at, request_url, request_headers, active, methods, kind)
VALUES (2, NOW(), NOW(), 'https://www.ory.sh/', '{}', 'password', '{}', 'login');

INSERT INTO session (sid, expires_at, issued_at, authenticated_at, identity_pk)
VALUES ('sid-2', NOW(), NOW(), NOW(), 2);

INSERT INTO self_service_error (id, errors, seen_at, was_seen)
VALUES ('2222-bc99-9c0b-4ef8-bb6d-6bb9-bd38-0a11', '[
  "foo",
  {
    "name": "bar"
  }
]', NULL, false);

INSERT INTO self_service_error (id, errors, seen_at, was_seen)
VALUES ('2223-bc99-9c0b-4ef8-bb6d-6bb9-bd38-0a11', '[
  "foo",
  {
    "name": "bar"
  }
]', NOW(), true);

-- +migrate Down
