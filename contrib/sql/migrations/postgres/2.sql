-- +migrate Up
CREATE TABLE IF NOT EXISTS self_service_error
(
    pk       BIGSERIAL PRIMARY KEY,
    id       UUID                     NOT NULL UNIQUE,
    errors   jsonb                    NOT NULL,
    seen_at  TIMESTAMP WITH TIME ZONE NULL,
    was_seen BOOL                     NOT NULL DEFAULT false
);

-- +migrate Down
DROP TABLE self_service_error;
