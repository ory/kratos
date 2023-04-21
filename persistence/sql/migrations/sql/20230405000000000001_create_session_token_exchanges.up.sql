CREATE TABLE session_token_exchanges (
    "id" UUID NOT NULL PRIMARY KEY,
    "nid" UUID NOT NULL,
    "flow_id" UUID NOT NULL,
    "session_id" UUID DEFAULT NULL,
    "init_code" VARCHAR(64) NOT NULL,
    "return_to_code" VARCHAR(64) NOT NULL,


    "created_at" timestamp NOT NULL,
    "updated_at" timestamp NOT NULL
);

-- Relevant query:
--   SELECT * from session_token_exchanges
--   WHERE nid = ? AND code = ? AND code <> '' AND session_id IS NOT NULL
CREATE INDEX session_token_exchanges_nid_code_idx ON session_token_exchanges (init_code, nid);

-- Relevant query:
--   UPDATE session_token_exchanges SET session_id = ? WHERE flow_id = ? AND nid = ?
CREATE INDEX session_token_exchanges_nid_flow_id_idx ON session_token_exchanges (flow_id, nid);
