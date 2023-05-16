CREATE TABLE session_token_exchanges (
    id CHAR(36) NOT NULL PRIMARY KEY,
    nid CHAR(36) NOT NULL,
    flow_id CHAR(36) NOT NULL,
    session_id CHAR(36) DEFAULT NULL,
    init_code VARCHAR(64) NOT NULL,
    return_to_code VARCHAR(64) NOT NULL,

    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Relevant query:
--   SELECT * from session_token_exchanges
--   WHERE nid = ? AND code = ? AND code <> '' AND session_id IS NOT NULL
CREATE INDEX session_token_exchanges_nid_code_idx ON session_token_exchanges (init_code, nid);

-- Relevant query:
--   UPDATE session_token_exchanges SET session_id = ? WHERE flow_id = ? AND nid = ?
CREATE INDEX session_token_exchanges_nid_flow_id_idx ON session_token_exchanges (flow_id, nid);
