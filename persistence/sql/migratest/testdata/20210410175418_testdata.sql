INSERT INTO selfservice_login_flows (id, nid, request_url, issued_at, expires_at, active_method, csrf_token, created_at, updated_at, forced, type, ui)
VALUES ('d6aa1f23-88c9-4b9b-a850-392f48c7f9e8', '884f556e-eb3a-4b9f-bee3-11345642c6c0', 'http://kratos:4433/self-service/browser/flows/login', '2013-10-07 08:23:19', '2013-10-07 08:23:19', '', 'fpeVSZ9ZH7YvUkhXsOVEIssxbfauh5lcoQSYxTcN0XkMneg1L42h+HtvisjlNjBF4ElcD2jApCHoJYq2u9sVWg==', '2013-10-07 08:23:19', '2013-10-07 08:23:19', false, 'api', '{}');

INSERT INTO selfservice_registration_flows (id, nid, request_url, issued_at, expires_at, active_method, csrf_token, created_at, updated_at, type, ui)
VALUES ('f1b5ed18-113a-4a98-aae7-d4eba007199c', '884f556e-eb3a-4b9f-bee3-11345642c6c0', 'http://kratos:4433/self-service/browser/flows/registration', '2013-10-07 08:23:19', '2013-10-07 08:23:19', 'password', 'vYYuhWXBfXKzBC+BlnbDmXfBKsUWY6SU/v04gHF9GYzPjFP51RXDPOc57R7Dpbf+XLkbPNAkmem33Crz/avdrw==', '2013-10-07 08:23:19', '2013-10-07 08:23:19', 'api', '{}');

INSERT INTO selfservice_settings_flows (id, nid, request_url, issued_at, expires_at, state, identity_id, created_at, updated_at, active_method, ui)
VALUES ('19ede218-928c-4e02-ab49-b76e12b34f31', '884f556e-eb3a-4b9f-bee3-11345642c6c0', 'http://kratos:4433/self-service/browser/flows/settings', '2013-10-07 08:23:19', '2013-10-07 08:23:19', 'show_form', 'a251ebc2-880c-4f76-a8f3-38e6940eab0e', '2013-10-07 08:23:19', '2013-10-07 08:23:19', 'profile', '{}');

INSERT INTO selfservice_verification_flows (id, nid, request_url, issued_at, expires_at, csrf_token, created_at, updated_at, type, ui)
VALUES ('7be6c72c-c868-4b61-a1f0-1130603665d8', '884f556e-eb3a-4b9f-bee3-11345642c6c0', 'http://kratos:4433/self-service/browser/flows/verification/email', '2013-10-07 08:23:19', '2013-10-07 08:23:19', '8xoIMa1+UkDqTt+tIHmIEHztQkk0AWk2PJhWWYDmB6dSE+RtJinnxtwH5lNNCnYyQuCF2ugy7rWjCgiwYPJNOw==', '2013-10-07 08:23:19', '2013-10-07 08:23:19', 'api', '{}');

INSERT INTO selfservice_recovery_flows (id, nid, request_url, issued_at, expires_at, active_method, csrf_token, state, recovered_identity_id, created_at, updated_at, type, ui)
VALUES ('68fb4010-84a9-4d1e-9f92-2705978ee89e', '884f556e-eb3a-4b9f-bee3-11345642c6c0', 'http://kratos:4433/self-service/browser/flows/recovery', '2013-10-07 08:23:19', '2013-10-07 08:23:19', 'link', 'vYYuhWXBfXKzBC+BlnbDmXfBKsUWY6SU/v04gHF9GYzPjFP51RXDPOc57R7Dpbf+XLkbPNAkmem33Crz/avdrw==', 'choose_method', 'a251ebc2-880c-4f76-a8f3-38e6940eab0e', '2013-10-07 08:23:19', '2013-10-07 08:23:19', 'api', '{}');

INSERT INTO courier_messages (id, nid, status, type, recipient, body, subject, template_type, template_data, created_at, updated_at) VALUES
('b821adf0-a067-4b3c-9f90-cac496d02a92',  '884f556e-eb3a-4b9f-bee3-11345642c6c0', 1, 1, 'foo@bar.com', 'body', 'subject', 'recovery_invalid', 'binary_data', '2013-10-07 08:23:19', '2013-10-07 08:23:19');

INSERT INTO identity_verifiable_addresses (id, nid, status, via, verified, value, verified_at, identity_id, created_at, updated_at) VALUES
('b2d59320-8564-4400-a39f-a22a497a23f1',  '884f556e-eb3a-4b9f-bee3-11345642c6c0', 'pending', 'email', false, 'foobar+without-code@ory.sh', null, 'a251ebc2-880c-4f76-a8f3-38e6940eab0e', '2013-10-07 08:23:19', '2013-10-07 08:23:19');

INSERT INTO identities (id, nid, traits_schema_id, traits, created_at, updated_at) VALUES ('196d8c1e-4f04-40f0-94b3-5ec43996b28a',  '884f556e-eb3a-4b9f-bee3-11345642c6c0', 'default', '{"email":"foobar@ory.sh"}', '2013-10-07 08:23:19', '2013-10-07 08:23:19');
INSERT INTO identities (id, nid, traits_schema_id, traits, created_at, updated_at) VALUES ('ed253b2c-48ed-4c58-9b6f-1dc963c30a66',  '884f556e-eb3a-4b9f-bee3-11345642c6c0', 'default', '{"email":"bazbar@ory.sh"}', '2013-10-07 08:23:19', '2013-10-07 08:23:19');
