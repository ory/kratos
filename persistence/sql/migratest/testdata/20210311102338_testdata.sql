INSERT INTO selfservice_login_flows (id, request_url, issued_at, expires_at, active_method, csrf_token, created_at, updated_at, forced, type, ui)
VALUES ('0bc96cc9-dda4-4700-9e42-35731f2af91e', 'http://kratos:4433/self-service/browser/flows/login', '2013-10-07 08:23:19', '2013-10-07 08:23:19', '', 'fpeVSZ9ZH7YvUkhXsOVEIssxbfauh5lcoQSYxTcN0XkMneg1L42h+HtvisjlNjBF4ElcD2jApCHoJYq2u9sVWg==', '2013-10-07 08:23:19', '2013-10-07 08:23:19', false, 'api', '{}');

INSERT INTO selfservice_registration_flows (id, request_url, issued_at, expires_at, active_method, csrf_token, created_at, updated_at, type, ui)
VALUES ('e2150cdc-23ac-4940-a240-6c79c27ab029', 'http://kratos:4433/self-service/browser/flows/registration', '2013-10-07 08:23:19', '2013-10-07 08:23:19', 'password', 'vYYuhWXBfXKzBC+BlnbDmXfBKsUWY6SU/v04gHF9GYzPjFP51RXDPOc57R7Dpbf+XLkbPNAkmem33Crz/avdrw==', '2013-10-07 08:23:19', '2013-10-07 08:23:19', 'api', '{}');

INSERT INTO selfservice_settings_flows (id, request_url, issued_at, expires_at, state, identity_id, created_at, updated_at, active_method, ui)
VALUES ('aeba85bd-1a8c-44bf-8fc3-3be83c01a3dc', 'http://kratos:4433/self-service/browser/flows/settings', '2013-10-07 08:23:19', '2013-10-07 08:23:19', 'show_form', 'a251ebc2-880c-4f76-a8f3-38e6940eab0e', '2013-10-07 08:23:19', '2013-10-07 08:23:19', 'profile', '{}');

INSERT INTO selfservice_verification_flows (id, request_url, issued_at, expires_at, csrf_token, created_at, updated_at, type, ui)
VALUES ('6aae3159-b880-4cfb-a863-03b114b1371b', 'http://kratos:4433/self-service/browser/flows/verification/email', '2013-10-07 08:23:19', '2013-10-07 08:23:19', '8xoIMa1+UkDqTt+tIHmIEHztQkk0AWk2PJhWWYDmB6dSE+RtJinnxtwH5lNNCnYyQuCF2ugy7rWjCgiwYPJNOw==', '2013-10-07 08:23:19', '2013-10-07 08:23:19', 'api', '{}');

INSERT INTO selfservice_recovery_flows (id, request_url, issued_at, expires_at, active_method, csrf_token, state, recovered_identity_id, created_at, updated_at, type, ui)
VALUES ('4963f305-e874-4a68-8424-a00bec679e7b', 'http://kratos:4433/self-service/browser/flows/recovery', '2013-10-07 08:23:19', '2013-10-07 08:23:19', 'link', 'vYYuhWXBfXKzBC+BlnbDmXfBKsUWY6SU/v04gHF9GYzPjFP51RXDPOc57R7Dpbf+XLkbPNAkmem33Crz/avdrw==', 'choose_method', 'a251ebc2-880c-4f76-a8f3-38e6940eab0e', '2013-10-07 08:23:19', '2013-10-07 08:23:19', 'api', '{}');
