CREATE TABLE identity_pending_traits_changes (
    id CHAR(36) NOT NULL PRIMARY KEY,
    identity_id CHAR(36) NOT NULL,
    nid CHAR(36) NOT NULL,
    new_address_value VARCHAR(255) NOT NULL,
    new_address_via VARCHAR(16) NOT NULL,
    original_traits_hash VARCHAR(64) NOT NULL,
    proposed_traits JSON NOT NULL,
    verification_flow_id CHAR(36) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    pending_identity_key CHAR(36) GENERATED ALWAYS AS (
        CASE WHEN status = 'pending' THEN identity_id ELSE NULL END
    ) VIRTUAL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT identity_pending_traits_changes_identities_id_fk FOREIGN KEY (identity_id) REFERENCES identities (id) ON DELETE CASCADE,
    CONSTRAINT identity_pending_traits_changes_networks_id_fk FOREIGN KEY (nid) REFERENCES networks (id) ON UPDATE RESTRICT ON DELETE CASCADE,
    CONSTRAINT identity_pending_traits_changes_verification_flows_id_fk FOREIGN KEY (verification_flow_id) REFERENCES selfservice_verification_flows (id) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE INDEX identity_pending_traits_changes_nid_identity_id_status_idx ON identity_pending_traits_changes (nid, identity_id, status);
CREATE INDEX identity_pending_traits_changes_nid_verification_flow_id_idx ON identity_pending_traits_changes (nid, verification_flow_id);
CREATE UNIQUE INDEX identity_pending_traits_changes_nid_identity_pending_idx
  ON identity_pending_traits_changes (nid, pending_identity_key);
