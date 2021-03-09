ALTER TABLE "identity_verifiable_addresses" ADD COLUMN "code" VARCHAR (32);
UPDATE identity_verifiable_addresses SET code = substr(md5(random()::text), 0, 32) WHERE code IS NULL;COMMIT TRANSACTION;BEGIN TRANSACTION;
ALTER TABLE "identity_verifiable_addresses" ALTER COLUMN "code" TYPE VARCHAR (32), ALTER COLUMN "code" SET NOT NULL;
CREATE UNIQUE INDEX "identity_verifiable_addresses_code_uq_idx" ON "identity_verifiable_addresses" (code);
CREATE INDEX "identity_verifiable_addresses_code_idx" ON "identity_verifiable_addresses" (code);
ALTER TABLE "identity_verifiable_addresses" ADD COLUMN "expires_at" timestamp NOT NULL DEFAULT '2000-01-01 00:00:00';