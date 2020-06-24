INSERT INTO identity_recovery_addresses (id, via, value, identity_id, created_at, updated_at)
VALUES ('b8293f1c-010f-45d9-b809-f3fc5365ba80', 'email', 'foobar@ory.sh', 'a251ebc2-880c-4f76-a8f3-38e6940eab0e', '2013-10-07 08:23:19', '2013-10-07 08:23:19');

INSERT INTO selfservice_recovery_requests (id, request_url, issued_at, expires_at, messages, active_method, csrf_token, state, recovered_identity_id, created_at, updated_at)
VALUES ('13178936-095a-466b-abe0-36d977d3dc18', 'http://kratos:4433/self-service/browser/flows/registration', '2013-10-07 08:23:19', '2013-10-07 08:23:19', '[]', 'link', 'vYYuhWXBfXKzBC+BlnbDmXfBKsUWY6SU/v04gHF9GYzPjFP51RXDPOc57R7Dpbf+XLkbPNAkmem33Crz/avdrw==', 'choose_method', 'a251ebc2-880c-4f76-a8f3-38e6940eab0e', '2013-10-07 08:23:19', '2013-10-07 08:23:19');

INSERT INTO selfservice_recovery_request_methods (id, method, selfservice_recovery_request_id, config, created_at, updated_at)
VALUES ('921462a6-8af6-4cda-97b4-dc0930ed271b', 'link', '13178936-095a-466b-abe0-36d977d3dc18', '{"action":"http://127.0.0.1:4455/.ory/kratos/public/self-service/browser/flows/settings/strategies/profile?request=21c5f714-3089-49d2-b387-f244d4dd9e00","method":"POST","fields":[{"name":"csrf_token","type":"hidden","required":true,"value":"yDwSg0quCmc4kBl7lBqYwGh4W8awrc+TpeWiigZs3iemRCwqeDhGdrW3sIv8T7u742pN+Kryx/NrdRpEXcT9qA=="},{"name":"traits.email","type":"text","value":"foo","errors":[{"message":"validation failed"},{"message":"foo is not valid email"},{"message":"foo is not valid email"}]}]}', '2013-10-07 08:23:19', '2013-10-07 08:23:19');

INSERT INTO identity_recovery_tokens (id, token, used, used_at, identity_recovery_address_id, selfservice_recovery_request_id, created_at, updated_at)
VALUES ('5529d454-2946-404e-b681-d950f8657fd0', 'd40c167d-a7f2-41a6-86b2-a5483001a010', false, null, 'b8293f1c-010f-45d9-b809-f3fc5365ba80', '13178936-095a-466b-abe0-36d977d3dc18', '2013-10-07 08:23:19', '2013-10-07 08:23:19');
