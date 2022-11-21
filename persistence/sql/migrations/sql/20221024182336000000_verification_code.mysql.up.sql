CREATE TABLE identity_verification_codes (
    id CHAR(36) NOT NULL PRIMARY KEY,
    code_hmac VARCHAR (64) NOT NULL,
    -- HMACed value of the actual code
    used_at timestamp NULL DEFAULT NULL,
    identity_verifiable_address_id CHAR(36),
    expires_at timestamp NOT NULL DEFAULT '2000-01-01 00:00:00',
    issued_at timestamp NOT NULL DEFAULT '2000-01-01 00:00:00',
    selfservice_verification_flow_id CHAR(36) NOT NULL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    nid CHAR(36) NOT NULL,
    CONSTRAINT identity_verification_codes_identity_verifiable_addresses_id_fk FOREIGN KEY (identity_verifiable_address_id) REFERENCES identity_verifiable_addresses (id) ON DELETE cascade,
    CONSTRAINT identity_verification_codes_selfservice_verification_flows_id_fk FOREIGN KEY (selfservice_verification_flow_id) REFERENCES selfservice_verification_flows (id) ON DELETE cascade,
    CONSTRAINT identity_verification_codes_networks_id_fk FOREIGN KEY (nid) REFERENCES networks (id) ON UPDATE RESTRICT ON DELETE CASCADE
);

ALTER TABLE
    selfservice_verification_flows
ADD
    COLUMN submit_count INT NOT NULL DEFAULT 0;

CREATE INDEX identity_verification_codes_nid_flow_id_idx ON identity_verification_codes (nid, selfservice_verification_flow_id);

CREATE INDEX identity_verification_codes_id_nid_idx ON identity_verification_codes (id, nid);