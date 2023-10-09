INSERT INTO selfservice_login_flows (id,nid, request_url, issued_at, expires_at, active_method, csrf_token, created_at,
                                     updated_at, forced, type, ui, internal_context, oauth2_login_challenge_data, state)
VALUES ('00b1517f-2467-4aaf-b0a5-82b4a27dcaf5',
        '884f556e-eb3a-4b9f-bee3-11345642c6c0',
        'http://kratos:4433/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login/self-service/browser/flows/login',
        '2013-10-07 08:23:19', '2013-10-07 08:23:19', '',
        'fpeVSZ9ZH7YvUkhXsOVEIssxbfauh5lcoQSYxTcN0XkMneg1L42h+HtvisjlNjBF4ElcD2jApCHoJYq2u9sVWg==',
        '2013-10-07 08:23:19', '2013-10-07 08:23:19', false, 'api', '{}', '{"foo":"bar"}', 'challenge data', 'choose_method');

INSERT INTO identities (id, nid, schema_id, traits, created_at, updated_at, metadata_public, metadata_admin,
                        available_aal)
VALUES ('28ff0031-190b-4253-bd15-14308dec013e', '884f556e-eb3a-4b9f-bee3-11345642c6c0', 'default',
        '{"email":"bazbarbarfoo@ory.sh"}', '2013-10-07 08:23:19', '2013-10-07 08:23:19', '{"foo":"bar"}', '{"baz":"bar"}',
        NULL);

INSERT INTO identity_login_codes (id, code, address, address_type, used_at, expires_at, issued_at, selfservice_login_flow_id, identity_id,
                                  created_at, updated_at, nid)
VALUES ('bd292366-af32-4ba6-bdf0-11d6d1a217f3',
'7eb71370d8497734ec78dfe613bf0f08967e206d2b5c2fc1243be823cfcd57a7',
'bazbarbarfoo@ory.com',
'email',
null,
'2022-08-18 08:28:18',
'2022-08-18 07:28:18',
'00b1517f-2467-4aaf-b0a5-82b4a27dcaf5',
'28ff0031-190b-4253-bd15-14308dec013e',
'2022-08-18 07:28:18',
'2022-08-18 07:28:18',
'884f556e-eb3a-4b9f-bee3-11345642c6c0'
)
