ALTER TABLE sessions MODIFY COLUMN token varchar(32) NULL;
ALTER TABLE sessions MODIFY COLUMN logout_token varchar(32) NULL;
