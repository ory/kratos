INSERT INTO sessions (id, nid, issued_at, expires_at, authenticated_at, created_at, updated_at, token, identity_id,
                      active, logout_token, aal, authentication_methods)
VALUES ('7458af86-c1d8-401c-978a-8da89133f78b', '884f556e-eb3a-4b9f-bee3-11345642c6c0', '2013-10-07 08:23:19',
        '2080-10-07 08:23:19', '2013-10-07 08:23:19', '2013-10-07 08:23:19', '2013-10-07 08:23:19',
        'eVwBt7UAAAAVwBt7UWPw', '5ff66179-c240-4703-b0d8-494592cefff5', true, '123eVwBt7UAAAeVwBt7UWPw', 'aal2',
        '[{"method":"password"},{"method":"totp"}]');

INSERT INTO selfservice_login_flows (id, nid, request_url, issued_at, expires_at, active_method, csrf_token, created_at,
                                     updated_at, forced, type, ui, requested_aal)
VALUES ('1fb23c75-b809-42cc-8984-6ca2d0a1192f', '884f556e-eb3a-4b9f-bee3-11345642c6c0',
        'http://kratos:4433/self-service/browser/flows/login', '2013-10-07 08:23:19', '2013-10-07 08:23:19', '',
        'fpeVSZ9ZH7YvUkhXsOVEIssxbfauh5lcoQSYxTcN0XkMneg1L42h+HtvisjlNjBF4ElcD2jApCHoJYq2u9sVWg==',
        '2013-10-07 08:23:19', '2013-10-07 08:23:19', false, 'api', '{}', 'aal2');
