DROP TABLE "selfservice_verification_flow_methods";
ALTER TABLE "selfservice_verification_flows" DROP COLUMN "active_method";
ALTER TABLE "selfservice_verification_flows" DROP COLUMN "state";
ALTER TABLE "selfservice_verification_flows" ADD COLUMN "via" VARCHAR (16) NOT NULL DEFAULT 'email';
ALTER TABLE "selfservice_verification_flows" ADD COLUMN "success" bool NOT NULL DEFAULT FALSE;