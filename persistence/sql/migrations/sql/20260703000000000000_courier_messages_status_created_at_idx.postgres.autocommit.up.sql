CREATE INDEX CONCURRENTLY IF NOT EXISTS courier_messages_status_created_at_idx ON courier_messages (status ASC, created_at ASC);
