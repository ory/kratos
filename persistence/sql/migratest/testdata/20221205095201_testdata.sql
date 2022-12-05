INSERT INTO
    courier_messages (
        id,
        type,
        status,
        body,
        subject,
        recipient,
        created_at,
        updated_at,
        template_type,
        nid,
        send_count
    )
VALUES
    (
        'd9d4401c-08a1-434c-8ab5-4a7edefde351',
        1,
        2,
        'Hi, please verify your account by clicking the following link: <a href=http://127.0.0.1:4455/.ory/kratos/public/self-service/browser/flows/verification/email/confirm/u9ZcBr5HbRTR8f53Qj2Ng3KR8Mv1Zjdb>http://127.0.0.1:4455/.ory/kratos/public/self-service/browser/flows/verification/email/confirm/u9ZcBr5HbRTR8f53Qj2Ng3KR8Mv1Zjdb</a>',
        'Please verify your email address',
        'foobar@ory.sh',
        '2013-10-07 08:23:19',
        '2013-10-07 08:23:19',
        'verification_valid',
        '884f556e-eb3a-4b9f-bee3-11345642c6c0',
        4
    );

INSERT INTO
    courier_message_dispatches
VALUES
    (
        'ea0c9a03-44ce-43a8-8c5d-52f468991cb9',
        'd9d4401c-08a1-434c-8ab5-4a7edefde351',
        'success',
        '{}',
        '884f556e-eb3a-4b9f-bee3-11345642c6c0',
        '2013-10-07 08:23:19',
        '2013-10-07 08:23:19'
    )