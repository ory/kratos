ALTER TABLE "selfservice_login_flows" ADD COLUMN "hydra_login_challenge" CHAR(36) NULL;
ALTER TABLE "selfservice_registration_flows" ADD COLUMN "hydra_login_challenge" CHAR(36) NULL;
