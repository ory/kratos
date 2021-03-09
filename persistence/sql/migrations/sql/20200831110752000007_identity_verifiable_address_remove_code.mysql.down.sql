ALTER TABLE `identity_verifiable_addresses` ADD COLUMN `code` VARCHAR (32);
UPDATE identity_verifiable_addresses SET code = LEFT(MD5(RAND()), 32) WHERE code IS NULL;
ALTER TABLE `identity_verifiable_addresses` MODIFY `code` VARCHAR (32) NOT NULL;
CREATE UNIQUE INDEX `identity_verifiable_addresses_code_uq_idx` ON `identity_verifiable_addresses` (`code`);
CREATE INDEX `identity_verifiable_addresses_code_idx` ON `identity_verifiable_addresses` (`code`);
ALTER TABLE `identity_verifiable_addresses` ADD COLUMN `expires_at` DATETIME NOT NULL DEFAULT '2000-01-01 00:00:00';