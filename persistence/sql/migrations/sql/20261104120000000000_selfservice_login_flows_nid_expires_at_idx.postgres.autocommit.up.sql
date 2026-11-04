CREATE INDEX CONCURRENTLY IF NOT EXISTS selfservice_login_flows_nid_expires_at_idx ON selfservice_login_flows (nid ASC, expires_at ASC);
