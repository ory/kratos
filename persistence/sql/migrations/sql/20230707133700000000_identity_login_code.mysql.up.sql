CREATE TABLE identity_login_codes
(
    id CHAR(36) NOT NULL PRIMARY KEY,
    code VARCHAR(64) NOT NULL, -- HMACed value of the actual code
    address VARCHAR(255) NOT NULL,
    address_type CHAR(36) NOT NULL,
    used_at timestamp NULL DEFAULT NULL,
    expires_at timestamp NOT NULL DEFAULT '2000-01-01 00:00:00',
    issued_at timestamp NOT NULL DEFAULT '2000-01-01 00:00:00',
    selfservice_login_flow_id CHAR(36),
    identity_id CHAR(36) NOT NULL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    nid CHAR(36) NOT NULL,
    CONSTRAINT identity_login_codes_selfservice_login_flows_id_fk
        FOREIGN KEY (selfservice_login_flow_id)
        REFERENCES selfservice_login_flows (id)
        ON DELETE cascade,
    CONSTRAINT identity_login_codes_networks_id_fk
        FOREIGN KEY (nid)
        REFERENCES networks (id)
        ON UPDATE RESTRICT ON DELETE CASCADE
);

CREATE INDEX identity_login_codes_nid_flow_id_idx ON identity_login_codes (nid, selfservice_login_flow_id);
CREATE INDEX identity_login_codes_id_nid_idx ON identity_login_codes (id, nid);


ALTER TABLE selfservice_login_flows ADD submit_count int NOT NULL DEFAULT 0;
ALTER TABLE selfservice_login_flows ADD skip_csrf_check boolean NOT NULL DEFAULT FALSE;
