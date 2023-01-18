CREATE INDEX courier_messages_nid_created_at_id_idx ON courier_messages (nid, created_at DESC);

CREATE INDEX courier_messages_nid_status_created_at_id_idx ON courier_messages (nid, status, created_at DESC);

CREATE INDEX courier_messages_nid_recipient_created_at_id_idx ON courier_messages (nid, recipient, created_at DESC);
