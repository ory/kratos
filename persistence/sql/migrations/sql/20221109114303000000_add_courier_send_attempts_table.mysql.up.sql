CREATE TABLE courier_message_dispatches (
  dispatch_id UUID PRIMARY KEY,
  message_id UUID NOT NULL,
  status TEXT NOT NULL,
  error TEXT,
  nid UUID NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  CONSTRAINT "courier_message_dispatches_message_id_fk" FOREIGN KEY ("message_id") REFERENCES "courier_messages" ("id") ON DELETE cascade,
  CONSTRAINT "courier_message_dispatches_nid_fk" FOREIGN KEY ("nid") REFERENCES "networks" ("id") ON DELETE cascade
);