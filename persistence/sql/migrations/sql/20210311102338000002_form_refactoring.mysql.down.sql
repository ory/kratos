CREATE TABLE `selfservice_verification_flow_methods` (
`id` char(36) NOT NULL,
PRIMARY KEY(`id`),
`method` VARCHAR (32) NOT NULL,
`selfservice_verification_flow_id` char(36) NOT NULL,
`config` JSON NOT NULL,
`created_at` DATETIME NOT NULL,
`updated_at` DATETIME NOT NULL,
FOREIGN KEY (`selfservice_verification_flow_id`) REFERENCES `selfservice_verification_flow_methods` (`id`) ON DELETE cascade
) ENGINE=InnoDB;