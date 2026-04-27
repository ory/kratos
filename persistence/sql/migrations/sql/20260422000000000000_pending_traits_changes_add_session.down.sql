DROP INDEX IF EXISTS identity_pending_traits_changes_nid_origin_settings_flow_id_idx;
DROP INDEX IF EXISTS identity_pending_traits_changes_nid_session_id_idx;

ALTER TABLE identity_pending_traits_changes
    DROP CONSTRAINT IF EXISTS identity_pending_traits_changes_settings_flow_id_fk;

ALTER TABLE identity_pending_traits_changes
    DROP CONSTRAINT IF EXISTS identity_pending_traits_changes_sessions_id_fk;

ALTER TABLE identity_pending_traits_changes DROP COLUMN IF EXISTS origin_settings_flow_id;
ALTER TABLE identity_pending_traits_changes DROP COLUMN IF EXISTS session_id;
