ALTER TABLE "selfservice_login_flows" ADD COLUMN "hydra_login_challenge" UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000';
ALTER TABLE "selfservice_registration_flows" ADD COLUMN "hydra_login_challenge" UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000';
