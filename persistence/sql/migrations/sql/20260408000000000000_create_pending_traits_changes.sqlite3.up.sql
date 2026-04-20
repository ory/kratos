CREATE TABLE identity_pending_traits_changes (
    "id" TEXT NOT NULL PRIMARY KEY,
    "identity_id" char(36) NOT NULL,
    "nid" char(36) NOT NULL,
    "new_address_value" VARCHAR(255) NOT NULL,
    "new_address_via" VARCHAR(16) NOT NULL,
    "original_traits_hash" VARCHAR(64) NOT NULL,
    "proposed_traits" TEXT NOT NULL,
    "verification_flow_id" char(36) NOT NULL,
    "status" VARCHAR(16) NOT NULL DEFAULT 'pending',
    "created_at" DATETIME NOT NULL,
    "updated_at" DATETIME NOT NULL,
    CONSTRAINT identity_pending_traits_changes_identities_id_fk FOREIGN KEY (identity_id) REFERENCES identities (id) ON DELETE CASCADE,
    CONSTRAINT identity_pending_traits_changes_networks_id_fk FOREIGN KEY (nid) REFERENCES networks (id) ON UPDATE RESTRICT ON DELETE CASCADE,
    CONSTRAINT identity_pending_traits_changes_verification_flows_id_fk FOREIGN KEY (verification_flow_id) REFERENCES selfservice_verification_flows (id) ON DELETE CASCADE
);

CREATE INDEX identity_pending_traits_changes_nid_identity_id_status_idx ON identity_pending_traits_changes (nid, identity_id, status);
CREATE INDEX identity_pending_traits_changes_nid_verification_flow_id_idx ON identity_pending_traits_changes (nid, verification_flow_id);
CREATE UNIQUE INDEX identity_pending_traits_changes_nid_identity_pending_idx
  ON identity_pending_traits_changes (nid, identity_id)
  WHERE status = 'pending';
