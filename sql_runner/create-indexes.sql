/* 20210618103120000003 */
CREATE UNIQUE INDEX `sessions_logout_token_uq_idx` ON `sessions` (`logout_token`);

/* 20210618103120000004 */
CREATE INDEX `sessions_logout_token_idx` ON `sessions` (`logout_token`);
