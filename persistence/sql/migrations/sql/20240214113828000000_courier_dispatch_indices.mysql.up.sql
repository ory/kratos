-- Remove unused index
DROP INDEX courier_message_dispatches_id_message_id_nid_idx ON courier_message_dispatches;

-- For pop eager load
CREATE INDEX courier_message_dispatches_message_id_idx ON courier_message_dispatches (message_id, created_at DESC);

-- For delete by nid
CREATE INDEX courier_message_dispatches_nid_idx ON courier_message_dispatches (nid);
