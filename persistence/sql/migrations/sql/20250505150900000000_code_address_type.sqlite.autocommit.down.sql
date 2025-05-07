ALTER TABLE identity_login_codes ADD COLUMN address_type_old CHAR(36);
UPDATE identity_login_codes SET address_type_old = address_type;
ALTER TABLE identity_login_codes DROP COLUMN address_type;
ALTER TABLE identity_login_codes RENAME COLUMN address_type_old TO address_type;

ALTER TABLE identity_registration_codes ADD COLUMN address_type_old CHAR(36);
UPDATE identity_registration_codes SET address_type_old = address_type;
ALTER TABLE identity_registration_codes DROP COLUMN address_type;
ALTER TABLE identity_registration_codes RENAME COLUMN address_type_old TO address_type;
