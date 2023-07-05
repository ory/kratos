UPDATE selfservice_login_flows SET request_url = substr(request_url, 1, 1024);
UPDATE selfservice_recovery_flows SET request_url = substr(request_url, 1, 1024);
UPDATE selfservice_registration_flows SET request_url = substr(request_url, 1, 1024);
UPDATE selfservice_settings_flows SET request_url = substr(request_url, 1, 1024);
UPDATE selfservice_verification_flows SET request_url = substr(request_url, 1, 1024);
