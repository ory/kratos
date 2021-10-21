INSERT INTO selfservice_login_flows (id, nid, request_url, issued_at, expires_at, active_method, csrf_token, created_at,
                                     updated_at, forced, type, ui, requested_aal, internal_context)
VALUES ('38caf592-b042-4551-b92f-8d5223c2a4e2', '884f556e-eb3a-4b9f-bee3-11345642c6c0',
        'http://kratos:4433/self-service/browser/flows/login', '2013-10-07 08:23:19', '2013-10-07 08:23:19', '',
        'fpeVSZ9ZH7YvUkhXsOVEIssxbfauh5lcoQSYxTcN0XkMneg1L42h+HtvisjlNjBF4ElcD2jApCHoJYq2u9sVWg==',
        '2013-10-07 08:23:19', '2013-10-07 08:23:19', false, 'api', '{}', 'aal2', '{"foo":"bar"}');

INSERT INTO selfservice_registration_flows (id, nid, request_url, issued_at, expires_at, active_method, csrf_token,
                                            created_at, updated_at, type, ui, internal_context)
VALUES ('8f32efdc-f6fc-4c27-a3c2-579d109eff60', '884f556e-eb3a-4b9f-bee3-11345642c6c0',
        'http://kratos:4433/self-service/browser/flows/registration', '2013-10-07 08:23:19', '2013-10-07 08:23:19',
        'password', 'vYYuhWXBfXKzBC+BlnbDmXfBKsUWY6SU/v04gHF9GYzPjFP51RXDPOc57R7Dpbf+XLkbPNAkmem33Crz/avdrw==',
        '2013-10-07 08:23:19', '2013-10-07 08:23:19', 'api', '{}', '{"foo":"bar"}');
