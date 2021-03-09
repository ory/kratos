DROP TABLE "selfservice_login_flow_methods";
ALTER TABLE "selfservice_login_flows" ADD COLUMN "ui" jsonb;
ALTER TABLE "selfservice_login_flows" DROP COLUMN "messages";