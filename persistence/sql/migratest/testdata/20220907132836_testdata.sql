INSERT INTO sessions (id, nid, issued_at, expires_at, authenticated_at, created_at, updated_at, token, identity_id,
                      active, logout_token, aal, authentication_methods)
VALUES ('7458af86-c1d8-401c-978a-8da89133f98b', '884f556e-eb3a-4b9f-bee3-11345642c6c0', '2013-10-07 08:23:19',
        '2013-10-07 08:23:19', '2013-10-07 08:23:19', '2013-10-07 08:23:19', '2013-10-07 08:23:19',
        'eVwBt7UAAAAVwBt7XAMw', '5ff66179-c240-4703-b0d8-494592cefff5', true, '123eVwBt7UAAAeVwBt7XAMw', 'aal2',
        '[{"method":"password"},{"method":"totp"}]');

INSERT INTO session_devices (id, nid, session_id, ip_address, user_agent, location, created_at, updated_at)
VALUES ('884f556e-eb3a-4b9f-bee3-11763642c6c0', '884f556e-eb3a-4b9f-bee3-11345642c6c0',
        '7458af86-c1d8-401c-978a-8da89133f98b', '54.155.246.232',
        'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36',
        'Munich, Germany', '2022-08-07 08:23:19', '2022-08-09 08:35:19');
