DROP INDEX identity_pending_traits_changes_nid_origin_settings_flow_id_idx ON identity_pending_traits_changes;
DROP INDEX identity_pending_traits_changes_nid_session_id_idx ON identity_pending_traits_changes;

ALTER TABLE identity_pending_traits_changes
    DROP FOREIGN KEY identity_pending_traits_changes_settings_flow_id_fk,
    DROP FOREIGN KEY identity_pending_traits_changes_sessions_id_fk,
    DROP COLUMN origin_settings_flow_id,
    DROP COLUMN session_id;
