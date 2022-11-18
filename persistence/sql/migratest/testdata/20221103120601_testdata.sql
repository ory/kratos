INSERT INTO
    selfservice_verification_flows (
        id,
        nid,
        request_url,
        issued_at,
        expires_at,
        csrf_token,
        created_at,
        updated_at,
        type,
        ui,
        submit_count
    )
VALUES
    (
        '81f74e5d-1fa5-4e1b-a9bf-e9511926047c',
        '884f556e-eb3a-4b9f-bee3-11345642c6c0',
        'http://kratos:4433/self-service/browser/flows/verification/email',
        '2022-11-03 08:23:19',
        '2022-11-03 08:23:19',
        '8xoIMa1+UkDqTt+tIHmIEHztQkk0AWk2PJhWWYDmB6dSE+RtJinnxtwH5lNNCnYyQuCF2ugy7rWjCgiwYPJNOw==',
        '2022-11-03 08:23:19',
        '2022-11-03 08:23:19',
        'api',
        '{}',
        0
    );

INSERT INTO
    identity_verification_codes (
        id,
        code_hmac,
        used_at,
        identity_verifiable_address_id,
        expires_at,
        issued_at,
        selfservice_verification_flow_id,
        created_at,
        updated_at,
        nid
    )
VALUES
    (
        '5ab4e72b-332a-48ab-b1e9-3b1ef1ba5b60',
        '7eb71370d8497734ec78dfe613bf0f08967e206d2b5c2fc1243be823cfcd57a7',
        null,
        '45e867e9-2745-4f16-8dd4-84334a252b61',
        '2022-11-03 08:28:18',
        '2022-11-03 08:28:18',
        '81f74e5d-1fa5-4e1b-a9bf-e9511926047c',
        '2022-11-03 08:28:18',
        '2022-11-03 08:28:18',
        '884f556e-eb3a-4b9f-bee3-11345642c6c0'
    );