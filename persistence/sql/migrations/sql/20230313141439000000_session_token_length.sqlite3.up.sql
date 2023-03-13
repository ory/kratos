DROP INDEX sessions_token_uq_idx;
DROP INDEX sessions_logout_token_uq_idx;
DROP INDEX sessions_token_nid_idx;

ALTER TABLE sessions RENAME COLUMN token TO old_token;
ALTER TABLE sessions RENAME COLUMN logout_token TO old_logout_token;
ALTER TABLE sessions
  ADD COLUMN token varchar(39) NULL;
ALTER TABLE sessions
  ADD COLUMN logout_token varchar(39) NULL;

UPDATE sessions
SET token = old_token
WHERE true;

UPDATE sessions
SET logout_token = old_logout_token
WHERE true;

ALTER TABLE sessions
  DROP COLUMN old_token;
ALTER TABLE sessions
  DROP COLUMN old_logout_token;

CREATE UNIQUE INDEX sessions_token_uq_idx ON sessions (logout_token);
CREATE UNIQUE INDEX sessions_logout_token_uq_idx ON sessions (token);
CREATE INDEX sessions_token_nid_idx ON sessions (nid, token);
