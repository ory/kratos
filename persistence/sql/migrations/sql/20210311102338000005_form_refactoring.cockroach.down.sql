CREATE TABLE "selfservice_settings_flow_methods" (
"id" UUID NOT NULL,
PRIMARY KEY("id"),
"method" VARCHAR (32) NOT NULL,
"selfservice_settings_flow_id" UUID NOT NULL,
"config" json NOT NULL,
"created_at" timestamp NOT NULL,
"updated_at" timestamp NOT NULL,
CONSTRAINT "selfservice_settings_flow_methods_selfservice_settings_flow_methods_id_fk" FOREIGN KEY ("selfservice_settings_flow_id") REFERENCES "selfservice_settings_flow_methods" ("id") ON DELETE cascade
);