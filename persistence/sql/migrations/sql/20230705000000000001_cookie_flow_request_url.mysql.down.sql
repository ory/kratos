UPDATE selfservice_login_flows SET request_url = substr(request_url, 1, 1024);
UPDATE selfservice_recovery_flows SET request_url = substr(request_url, 1, 1024);
UPDATE selfservice_registration_flows SET request_url = substr(request_url, 1, 1024);
UPDATE selfservice_settings_flows SET request_url = substr(request_url, 1, 1024);
UPDATE selfservice_verification_flows SET request_url = substr(request_url, 1, 1024);

ALTER TABLE selfservice_login_flows MODIFY request_url VARCHAR(1024);
ALTER TABLE selfservice_recovery_flows MODIFY request_url VARCHAR(1024);
ALTER TABLE selfservice_registration_flows MODIFY request_url VARCHAR(1024);
ALTER TABLE selfservice_settings_flows MODIFY request_url VARCHAR(1024);
ALTER TABLE selfservice_verification_flows MODIFY request_url VARCHAR(1024);
