CREATE INDEX IF NOT EXISTS courier_message_dispatches_id_message_id_nid_idx ON courier_message_dispatches (id ASC, message_id ASC, nid ASC);

DROP INDEX courier_message_dispatches_message_id_idx;
DROP INDEX courier_message_dispatches_nid_idx;
