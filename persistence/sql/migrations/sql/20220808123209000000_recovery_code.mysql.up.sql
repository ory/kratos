CREATE TABLE identity_recovery_codes
(
    id CHAR(36) NOT NULL PRIMARY KEY,
    code VARCHAR (64) NOT NULL, -- HMACed value of the actual code
    used bool NOT NULL DEFAULT FALSE,
    used_at timestamp,
    identity_recovery_address_id CHAR(36),
    expires_at timestamp NOT NULL DEFAULT '2000-01-01 00:00:00',
    issued_at timestamp NOT NULL DEFAULT '2000-01-01 00:00:00',
    selfservice_recovery_flow_id CHAR(36),
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    nid CHAR(36) NOT NULL,
    identity_id CHAR(36) NOT NULL,
    CONSTRAINT identity_recovery_codes_identity_recovery_addresses_id_fk 
        FOREIGN KEY (identity_recovery_address_id)
        REFERENCES identity_recovery_addresses (id)
        ON DELETE cascade,
    CONSTRAINT identity_recovery_codes_selfservice_recovery_flows_id_fk 
        FOREIGN KEY (selfservice_recovery_flow_id) 
        REFERENCES selfservice_recovery_flows (id)
        ON DELETE cascade,
    CONSTRAINT identity_recovery_tokens_identity_id_fk 
        FOREIGN KEY (identity_id) 
        REFERENCES identities (id)
        ON UPDATE RESTRICT ON DELETE CASCADE
);

CREATE INDEX identity_recovery_codes_nid_idx ON identity_recovery_codes (id, nid);

ALTER TABLE selfservice_recovery_flows ADD submit_count int NOT NULL DEFAULT 0;
