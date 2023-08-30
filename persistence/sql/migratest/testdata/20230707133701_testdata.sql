INSERT INTO selfservice_registration_flows (id, nid, request_url, issued_at, expires_at, active_method, csrf_token,
                                            created_at, updated_at, type, ui, internal_context, oauth2_login_challenge, state)
VALUES ('69c80296-36cd-4afc-921a-15369cac5bf0', '884f556e-eb3a-4b9f-bee3-11345642c6c0',
        'http://kratos:4433/self-service/browser/flows/registration?login_challenge/self-service/browser/flows/registration?login_challenge/self-service/browser/flows/registration?login_challenge/self-service/browser/flows/registration?login_challenge/self-service/browser/flows/registration?login_challenge/self-service/browser/flows/registration?login_challenge/self-service/browser/flows/registration?login_challenge/self-service/browser/flows/registration?login_challenge/self-service/browser/flows/registration?login_challenge/self-service/browser/flows/registration?login_challenge/self-service/browser/flows/registration?login_challenge/self-service/browser/flows/registration?login_challenge/self-service/browser/flows/registration?login_challenge/self-service/browser/flows/registration?login_challenge/self-service/browser/flows/registration?login_challenge/self-service/browser/flows/registration?login_challenge/self-service/browser/flows/registration?login_challenge/self-service/browser/flows/registration?login_challenge/self-service/browser/flows/registration?login_challenge/self-service/browser/flows/registration?login_challenge/self-service/browser/flows/registration?login_challenge/self-service/browser/flows/registration?login_challenge/self-service/browser/flows/registration?login_challenge/self-service/browser/flows/registration?login_challenge/self-service/browser/flows/registration?login_challenge/self-service/browser/flows/registration?login_challenge/self-service/browser/flows/registration?login_challenge=',
        '2013-10-07 08:23:19', '2013-10-07 08:23:19',
        'password', 'vYYuhWXBfXKzBC+BlnbDmXfBKsUWY6SU/v04gHF9GYzPjFP51RXDPOc57R7Dpbf+XLkbPNAkmem33Crz/avdrw==',
        '2013-10-07 08:23:19', '2013-10-07 08:23:19', 'browser', '{}', '{"foo":"bar"}',
        '3caddfd5-9903-4bce-83ff-cae36f42dff7', 'choose_method');

INSERT INTO identity_registration_codes (id, address, address_type, code, used_at, expires_at, issued_at, selfservice_registration_flow_id,
                                  created_at, updated_at, nid)
VALUES ('f1f66a69-ce02-4a12-9591-9e02dda30a0d',
'example@example.com',
'email',
'7eb71370d8497734ec78dfe613bf0f08967e206d2b5c2fc1243be823cfcd57a7',
null,
'2022-08-18 08:28:18',
'2022-08-18 07:28:18',
'69c80296-36cd-4afc-921a-15369cac5bf0',
'2022-08-18 07:28:18',
'2022-08-18 07:28:18',
'884f556e-eb3a-4b9f-bee3-11345642c6c0'
)
