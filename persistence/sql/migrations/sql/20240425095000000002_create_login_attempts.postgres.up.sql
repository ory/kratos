CREATE TABLE
  selfservice_login_attempts (
    id UUID PRIMARY KEY,
    identity_id UUID NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW (),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW (),
    CONSTRAINT fk_selfservice_login_attempts_identity_id FOREIGN KEY (identity_id) REFERENCES identities (id)
  );

CREATE INDEX idx_selfservice_login_attempts_identity_id ON selfservice_login_attempts (identity_id);

CREATE INDEX idx_selfservice_login_attempts_created_at ON selfservice_login_attempts (created_at);
