ALTER TABLE `selfservice_verification_flows` ADD COLUMN `state` VARCHAR (255) NOT NULL DEFAULT 'show_form';
UPDATE selfservice_verification_flows SET state='passed_challenge' WHERE success IS TRUE;
CREATE TABLE `selfservice_verification_flow_methods` (
`id` char(36) NOT NULL,
PRIMARY KEY(`id`),
`method` VARCHAR (32) NOT NULL,
`selfservice_verification_flow_id` char(36) NOT NULL,
`config` JSON NOT NULL,
`created_at` DATETIME NOT NULL,
`updated_at` DATETIME NOT NULL
) ENGINE=InnoDB;
ALTER TABLE `selfservice_verification_flows` ADD COLUMN `active_method` VARCHAR (32);
INSERT INTO selfservice_verification_flow_methods (id, method, selfservice_verification_flow_id, config) SELECT id, 'link', id, form FROM selfservice_verification_flows;
ALTER TABLE `selfservice_verification_flows` DROP COLUMN `form`;
ALTER TABLE `selfservice_verification_flows` DROP COLUMN `via`;
ALTER TABLE `selfservice_verification_flows` DROP COLUMN `success`;