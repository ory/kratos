CREATE INDEX CONCURRENTLY IF NOT EXISTS selfservice_registration_flows_nid_expires_at_idx ON selfservice_registration_flows (nid ASC, expires_at ASC);
