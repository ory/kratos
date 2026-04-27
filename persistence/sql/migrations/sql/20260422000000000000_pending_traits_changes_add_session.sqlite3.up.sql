ALTER TABLE identity_pending_traits_changes
    ADD COLUMN "session_id" char(36) NULL REFERENCES sessions (id) ON DELETE SET NULL;

ALTER TABLE identity_pending_traits_changes
    ADD COLUMN "origin_settings_flow_id" char(36) NULL REFERENCES selfservice_settings_flows (id) ON DELETE CASCADE;

CREATE INDEX identity_pending_traits_changes_nid_session_id_idx
    ON identity_pending_traits_changes (nid, session_id);

CREATE INDEX identity_pending_traits_changes_nid_origin_settings_flow_id_idx
    ON identity_pending_traits_changes (nid, origin_settings_flow_id);
