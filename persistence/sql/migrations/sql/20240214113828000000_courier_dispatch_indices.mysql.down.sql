CREATE INDEX courier_message_dispatches_id_message_id_nid_idx (id ASC, message_id ASC, nid ASC);

DROP INDEX courier_message_dispatches_message_id_idx ON courier_message_dispatches;
DROP INDEX courier_message_dispatches_message_nid_idx ON courier_message_dispatches;
