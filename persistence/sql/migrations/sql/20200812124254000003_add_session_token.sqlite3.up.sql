INSERT INTO "_sessions_tmp" (id, issued_at, expires_at, authenticated_at, identity_id, created_at, updated_at, token) SELECT id, issued_at, expires_at, authenticated_at, identity_id, created_at, updated_at, token FROM "sessions";