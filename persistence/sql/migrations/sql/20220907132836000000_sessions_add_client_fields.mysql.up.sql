CREATE TABLE `session_devices`
(
  `id`         char(36) NOT NULL,
  PRIMARY KEY (`id`),
  `ip_address` VARCHAR(50)  DEFAULT '',
  `user_agent` VARCHAR(255) DEFAULT '',
  `location`   VARCHAR(255) DEFAULT '',
  `session_id` char(36) NOT NULL,
  `nid`        char(36) NOT NULL,
  `created_at` DATETIME NOT NULL,
  `updated_at`  DATETIME NOT NULL,
  FOREIGN KEY (`session_id`) REFERENCES `sessions` (`id`) ON DELETE cascade,
  FOREIGN KEY (`nid`) REFERENCES `networks` (`id`) ON DELETE cascade,
  CONSTRAINT unique_session_device UNIQUE (nid, session_id, ip_address, user_agent)
) ENGINE = InnoDB;
CREATE INDEX `session_devices_session_id_nid_idx` ON `session_devices` (`session_id`, `nid`);
