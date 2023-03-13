ALTER TABLE sessions MODIFY COLUMN token varchar(39) NULL;
ALTER TABLE sessions MODIFY COLUMN logout_token varchar(39) NULL;
