CREATE TABLE "selfservice_login_flow_methods" (
"id" UUID NOT NULL,
PRIMARY KEY("id"),
"method" VARCHAR (32) NOT NULL,
"selfservice_login_flow_id" UUID NOT NULL,
"config" json NOT NULL,
"created_at" timestamp NOT NULL,
"updated_at" timestamp NOT NULL,
CONSTRAINT "selfservice_login_flow_methods_selfservice_login_flow_methods_id_fk" FOREIGN KEY ("selfservice_login_flow_id") REFERENCES "selfservice_login_flow_methods" ("id") ON DELETE cascade
);COMMIT TRANSACTION;BEGIN TRANSACTION;
ALTER TABLE "selfservice_login_flows" DROP COLUMN "ui";COMMIT TRANSACTION;BEGIN TRANSACTION;
ALTER TABLE "selfservice_login_flows" ADD COLUMN "messages" json;COMMIT TRANSACTION;BEGIN TRANSACTION;