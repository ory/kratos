CREATE INDEX CONCURRENTLY IF NOT EXISTS selfservice_verification_flows_nid_expires_at_idx ON selfservice_verification_flows (nid ASC, expires_at ASC);
