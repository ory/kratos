INSERT INTO identity_verifiable_addresses (id, status, via, verified, value, verified_at, identity_id, created_at, updated_at) VALUES
('b2d59320-8564-4400-a39f-a22a497a23f1', 'pending', 'email', false, 'foobar+without-code@ory.sh', null, 'a251ebc2-880c-4f76-a8f3-38e6940eab0e', '2013-10-07 08:23:19', '2013-10-07 08:23:19');

INSERT INTO identity_verification_tokens (id, token, used, used_at, identity_verifiable_address_id, selfservice_verification_flow_id, created_at, updated_at, expires_at, issued_at)
VALUES ('ee56574d-2f0c-43f6-8d26-0062938ae330', '1001ba7ddd644cb68478e8947e4jfhc', false, null, 'b2d59320-8564-4400-a39f-a22a497a23f1', '5385c962-0295-4575-9b1b-d7eef13c0eda', '2013-10-07 08:23:19', '2013-10-07 08:23:19', '2013-10-07 08:23:19', '2013-10-07 08:23:19');

INSERT INTO identity_verification_tokens (id, token, used, used_at, identity_verifiable_address_id, selfservice_verification_flow_id, created_at, updated_at, expires_at, issued_at)
VALUES ('f81fd924-23bb-4cdf-8fa0-56253eff6cc9', '1001ba7ddd644cb68478e8947e4jfhd', false, null, 'b2d59320-8564-4400-a39f-a22a497a23f1', NULL, '2013-10-07 08:23:19', '2013-10-07 08:23:19', '2013-10-07 08:23:19', '2013-10-07 08:23:19');
