ALTER TABLE "identity_verifiable_addresses" ADD COLUMN "code" VARCHAR (32);COMMIT TRANSACTION;BEGIN TRANSACTION;
ALTER TABLE "identity_verifiable_addresses" RENAME COLUMN "code" TO "_code_tmp";COMMIT TRANSACTION;BEGIN TRANSACTION;
ALTER TABLE "identity_verifiable_addresses" ADD COLUMN "code" VARCHAR (32);COMMIT TRANSACTION;BEGIN TRANSACTION;
UPDATE "identity_verifiable_addresses" SET "code" = "_code_tmp";COMMIT TRANSACTION;BEGIN TRANSACTION;
ALTER TABLE "identity_verifiable_addresses" ALTER COLUMN "code" SET NOT NULL;COMMIT TRANSACTION;BEGIN TRANSACTION;
ALTER TABLE "identity_verifiable_addresses" DROP COLUMN "_code_tmp";COMMIT TRANSACTION;BEGIN TRANSACTION;
CREATE UNIQUE INDEX "identity_verifiable_addresses_code_uq_idx" ON "identity_verifiable_addresses" (code);COMMIT TRANSACTION;BEGIN TRANSACTION;
CREATE INDEX "identity_verifiable_addresses_code_idx" ON "identity_verifiable_addresses" (code);COMMIT TRANSACTION;BEGIN TRANSACTION;