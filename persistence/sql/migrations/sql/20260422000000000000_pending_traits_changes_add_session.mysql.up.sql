ALTER TABLE identity_pending_traits_changes
    ADD COLUMN session_id CHAR(36) NULL,
    ADD CONSTRAINT identity_pending_traits_changes_sessions_id_fk
        FOREIGN KEY (session_id) REFERENCES sessions (id) ON DELETE SET NULL,
    ADD COLUMN origin_settings_flow_id CHAR(36) NULL,
    ADD CONSTRAINT identity_pending_traits_changes_settings_flow_id_fk
        FOREIGN KEY (origin_settings_flow_id) REFERENCES selfservice_settings_flows (id) ON DELETE CASCADE;

CREATE INDEX identity_pending_traits_changes_nid_session_id_idx
    ON identity_pending_traits_changes (nid, session_id);

CREATE INDEX identity_pending_traits_changes_nid_origin_settings_flow_id_idx
    ON identity_pending_traits_changes (nid, origin_settings_flow_id);
