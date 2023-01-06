CREATE INDEX courier_messages_nid_created_at_id_idx ON courier_messages (nid, created_at, id);

CREATE INDEX courier_messages_recipient_idx ON courier_messages (recipient);
