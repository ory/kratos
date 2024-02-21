CREATE INDEX courier_message_dispatches_id_message_id_nid_idx ON courier_message_dispatches (id ASC, message_id ASC, nid ASC);

-- These can't be removed because of foreign key constraints which disallow index deletion in MySQL.

-- DROP INDEX courier_message_dispatches_message_id_idx ON courier_message_dispatches;
-- DROP INDEX courier_message_dispatches_nid_idx ON courier_message_dispatches;
