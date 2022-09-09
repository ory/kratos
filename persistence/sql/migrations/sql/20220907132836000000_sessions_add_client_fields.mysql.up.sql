CREATE TABLE `session_metadata`
(
  `id`         char(36) NOT NULL,
  PRIMARY KEY (`id`),
  `ip_address` VARCHAR(50)  DEFAULT '',
  `user_agent` VARCHAR(255) DEFAULT '',
  `location`   VARCHAR(255) DEFAULT '',
  `session_id` char(36) NOT NULL,
  `nid`        char(36) NOT NULL,
  `created_at` DATETIME NOT NULL,
  `last_seen`  DATETIME NOT NULL,
  FOREIGN KEY (`session_id`) REFERENCES `sessions` (`id`) ON DELETE cascade,
  FOREIGN KEY (`nid`) REFERENCES `networks` (`id`) ON DELETE cascade
) ENGINE = InnoDB;
