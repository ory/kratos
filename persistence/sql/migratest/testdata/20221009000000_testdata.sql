INSERT INTO
    sessions (
        id,
        nid,
        issued_at,
        expires_at,
        privileged_until,
        authenticated_at,
        created_at,
        updated_at,
        token,
        identity_id,
        active,
        logout_token,
        aal,
        authentication_methods
    )
VALUES
    (
        -- id
        '1ba253ce-0501-429d-ab2b-4ae17db13c6d',
        -- nid,
        '884f556e-eb3a-4b9f-bee3-11345642c6c0',
        -- issued_at
        '2013-10-07 08:23:19',
        -- expires_at
        '2013-10-07 08:23:19',
        -- privileged_until
        '2013-10-07 09:23:19',
        -- authenticated_at
        '2013-10-07 08:23:19',
        -- created_at
        '2013-10-07 08:23:19',
        -- updated_at
        '2013-10-07 08:23:19',
        -- token
        '1234ba7ddd644cb68478e8947e4jfhb',
        -- identity_id
        '5ff66179-c240-4703-b0d8-494592cefff5',
        -- active
        true,
        -- logout_token
        '1234ba7ddd644cb68478',
        -- aal,
        'aal1',
        -- authentication_methods
        '[{"method":"password"}]'
    );
