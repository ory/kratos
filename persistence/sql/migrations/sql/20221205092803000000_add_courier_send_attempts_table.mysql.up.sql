CREATE TABLE courier_message_dispatches (
  id CHAR(36) PRIMARY KEY,
  message_id CHAR(36) NOT NULL,
  status VARCHAR(7) NOT NULL,
  error JSON,
  nid CHAR(36) NOT NULL,
  created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT courier_message_dispatches_message_id_fk FOREIGN KEY (message_id) REFERENCES courier_messages (id) ON DELETE cascade,
  CONSTRAINT courier_message_dispatches_nid_fk FOREIGN KEY (nid) REFERENCES networks (id) ON DELETE cascade
);

CREATE INDEX courier_message_dispatches_id_message_id_nid_idx ON courier_message_dispatches (id, message_id, nid);