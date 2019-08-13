-- +migrate Up
CREATE TYPE credentials_type AS ENUM ('oidc', 'password');
CREATE TYPE self_service_request_type AS ENUM ('login', 'registration');

CREATE TABLE IF NOT EXISTS identity
(
    pk                BIGSERIAL PRIMARY KEY,
    id                VARCHAR(255) NOT NULL UNIQUE,
    traits            jsonb        NOT NULL DEFAULT '{}'::jsonb,
    traits_schema_url text         NOT NULL
);

CREATE TABLE IF NOT EXISTS identity_credentials
(
    pk          BIGSERIAL PRIMARY KEY,
    identity_id BIGINT REFERENCES identity (pk),
    method      credentials_type NOT NULL,
    options     jsonb            NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE IF NOT EXISTS identity_credentials_identifiers
(
    pk                      BIGSERIAL PRIMARY KEY,
    identity_credentials_id BIGINT REFERENCES identity_credentials (pk),
    identifier              VARCHAR(255) NOT NULL UNIQUE,
    CHECK (length(identifier) > 0)
);

CREATE TABLE IF NOT EXISTS self_service_request
(
    pk              BIGSERIAL PRIMARY KEY,
    id              VARCHAR(36)               NOT NULL UNIQUE,
    expires_at      TIMESTAMP                 NOT NULL,
    issued_at       TIMESTAMP                 NOT NULL DEFAULT NOW(),
    request_url     text                      NOT NULL,
    request_headers jsonb                     NOT NULL DEFAULT '{}'::jsonb,
    active          credentials_type          NOT NULL,
    methods         jsonb                     NOT NULL DEFAULT '{}'::jsonb,
    kind            self_service_request_type NOT NULL
);

CREATE UNIQUE INDEX name ON self_service_request (id, kind);

-- +migrate Down
DROP TABLE identity_credentials_identifiers;
DROP TABLE identity_credentials;
DROP TABLE identity;
DROP TABLE self_service_request;
