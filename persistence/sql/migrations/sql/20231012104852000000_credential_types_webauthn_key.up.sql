INSERT INTO
  identity_credential_types (id, name)
SELECT
  '92a3a4d1-f045-4fb2-b6c4-9a0ce104682f',
  'webauthn_key'
WHERE
  NOT EXISTS (
    SELECT
      *
    FROM
      identity_credential_types
    WHERE
      name = 'webauthn_key'
  );