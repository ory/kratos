INSERT INTO selfservice_login_flows (id, nid, request_url, issued_at, expires_at, active_method, csrf_token, created_at, updated_at, forced, type, ui, requested_aal, internal_context, oauth2_login_challenge)
VALUES ('349c945a-60f8-436a-a301-7a42c92604f9', '884f556e-eb3a-4b9f-bee3-11345642c6c0',
        'http://kratos:4433/self-service/browser/flows/login?login_challenge=3caddfd599034bce83ffcae36f42dff7', '2013-10-07 08:23:19', '2013-10-07 08:23:19', '',
        'fpeVSZ9ZH7YvUkhXsOVEIssxbfauh5lcoQSYxTcN0XkMneg1L42h+HtvisjlNjBF4ElcD2jApCHoJYq2u9sVWg==',
        '2013-10-07 08:23:19', '2013-10-07 08:23:19', false, 'browser', '{}', 'aal2', '{"foo":"bar"}', '3caddfd5-9903-4bce-83ff-cae36f42dff7');

INSERT INTO selfservice_registration_flows (id, nid, request_url, issued_at, expires_at, active_method, csrf_token, created_at, updated_at, type, ui, internal_context, oauth2_login_challenge)
VALUES ('ef18b06e-4700-4021-9949-ef783cd86be8', '884f556e-eb3a-4b9f-bee3-11345642c6c0',
        'http://kratos:4433/self-service/browser/flows/registration?login_challenge=', '2013-10-07 08:23:19', '2013-10-07 08:23:19',
        'password', 'vYYuhWXBfXKzBC+BlnbDmXfBKsUWY6SU/v04gHF9GYzPjFP51RXDPOc57R7Dpbf+XLkbPNAkmem33Crz/avdrw==',
        '2013-10-07 08:23:19', '2013-10-07 08:23:19', 'browser', '{}', '{"foo":"bar"}', '3caddfd5-9903-4bce-83ff-cae36f42dff7');
