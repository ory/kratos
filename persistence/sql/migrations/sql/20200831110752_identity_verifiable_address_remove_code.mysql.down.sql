ALTER TABLE `identity_verifiable_addresses` ADD COLUMN `code` VARCHAR (32);
ALTER TABLE `identity_verifiable_addresses` MODIFY `code` VARCHAR (32) NOT NULL;
CREATE UNIQUE INDEX `identity_verifiable_addresses_code_uq_idx` ON `identity_verifiable_addresses` (`code`);
CREATE INDEX `identity_verifiable_addresses_code_idx` ON `identity_verifiable_addresses` (`code`);