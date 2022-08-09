CREATE TABLE "identity_recovery_codes"
(
    "id" UUID NOT NULL PRIMARY KEY,
    "code" VARCHAR (64) NOT NULL, -- HMACed value of the actual code
    "used" bool NOT NULL DEFAULT 'false',
    "used_at" timestamp,
    "identity_recovery_address_id" UUID,
    "expires_at" timestamp NOT NULL DEFAULT '2000-01-01 00:00:00',
    "issued_at" timestamp NOT NULL DEFAULT '2000-01-01 00:00:00',
    "selfservice_recovery_flow_id" UUID,
    "created_at" timestamp NOT NULL,
    "updated_at" timestamp NOT NULL,
    "nid" UUID NOT NULL,
    "identity_id" UUID NOT NULL,
    CONSTRAINT "identity_recovery_codes_identity_recovery_addresses_id_fk" 
        FOREIGN KEY ("identity_recovery_address_id")
        REFERENCES "identity_recovery_addresses" ("id")
        ON DELETE cascade,
    CONSTRAINT "identity_recovery_codes_selfservice_recovery_flows_id_fk" 
        FOREIGN KEY ("selfservice_recovery_flow_id") 
        REFERENCES "selfservice_recovery_flows" ("id")
        ON DELETE cascade,
    CONSTRAINT "identity_recovery_tokens_identity_id_fk" 
        FOREIGN KEY ("identity_id") 
        REFERENCES "identities" ("id")
        ON UPDATE RESTRICT ON DELETE CASCADE
);

CREATE INDEX "identity_recovery_codes_nid_idx" ON "identity_recovery_codes" (id, nid);
