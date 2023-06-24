-- Drop columns in a later migration.
-- ALTER TABLE selfservice_login_flows DROP COLUMN oauth2_login_challenge;
-- ALTER TABLE selfservice_registration_flows DROP COLUMN oauth2_login_challenge;

ALTER TABLE selfservice_login_flows ADD COLUMN oauth2_login_challenge_data TEXT NULL;
ALTER TABLE selfservice_registration_flows ADD COLUMN oauth2_login_challenge_data TEXT NULL;
