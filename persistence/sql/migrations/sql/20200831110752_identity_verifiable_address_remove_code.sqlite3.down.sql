ALTER TABLE "identity_verifiable_addresses" ADD COLUMN "code" TEXT;
UPDATE identity_verifiable_addresses SET code = substr(hex(randomblob(32)), 0, 32) WHERE code IS NULL;