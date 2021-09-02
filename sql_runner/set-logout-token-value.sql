/* 20210618103120000001 */
UPDATE sessions SET logout_token = token;

/* 20210618103120000002 */
ALTER TABLE `sessions` MODIFY `logout_token` VARCHAR (32);
