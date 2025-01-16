-- Step 1: Create a temporary table without the nid column and foreign key constraint
CREATE TABLE session_token_exchanges_temp
(
  id             TEXT        NOT NULL,
  flow_id        TEXT        NOT NULL,
  session_id     TEXT,
  init_code      VARCHAR(64) NOT NULL,
  return_to_code VARCHAR(64) NOT NULL,
  created_at     TIMESTAMP   NOT NULL,
  updated_at     TIMESTAMP   NOT NULL,
  PRIMARY KEY (id)
);

-- Step 2: Copy data from the original table to the temporary table (excluding the nid column)
INSERT INTO session_token_exchanges_temp (id, flow_id, session_id, init_code, return_to_code, created_at, updated_at)
SELECT id, flow_id, session_id, init_code, return_to_code, created_at, updated_at
FROM session_token_exchanges;

-- Step 3: Drop the original table
DROP TABLE session_token_exchanges;

-- Step 4: Rename the temporary table to the original table name
ALTER TABLE session_token_exchanges_temp RENAME TO session_token_exchanges;

-- Step 5: Recreate indexes as needed (excluding nid)
CREATE INDEX session_token_exchanges_nid_code_idx ON session_token_exchanges (init_code);
CREATE INDEX session_token_exchanges_nid_flow_id_idx ON session_token_exchanges (flow_id);
