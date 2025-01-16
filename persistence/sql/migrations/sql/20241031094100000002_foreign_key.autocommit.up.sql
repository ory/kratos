ALTER TABLE session_token_exchanges ADD CONSTRAINT session_token_exchanges_nid_fk FOREIGN KEY (nid) REFERENCES networks (id) ON DELETE CASCADE;
