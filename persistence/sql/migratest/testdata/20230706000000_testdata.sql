INSERT INTO identities (id, nid, schema_id, traits, created_at, updated_at, metadata_public, metadata_admin,
                        available_aal)
VALUES ('0149ce5f-76a8-4efe-b2e3-431b8c6cceb6', '884f556e-eb3a-4b9f-bee3-11345642c6c0', 'default',
        '{"email":"bazbar@ory.sh"}', '2013-10-07 08:23:19', '2013-10-07 08:23:19', '{"foo":"bar"}', '{"baz":"bar"}',
        'aal1');

INSERT INTO identities (id, nid, schema_id, traits, created_at, updated_at, metadata_public, metadata_admin,
                        available_aal)
VALUES ('0149ce5f-76a8-4efe-b2e3-431b8c6cceb7', '884f556e-eb3a-4b9f-bee3-11345642c6c0', 'default',
        '{"email":"bazbarbar@ory.sh"}', '2013-10-07 08:23:19', '2013-10-07 08:23:19', '{"foo":"bar"}', '{"baz":"bar"}',
        NULL);
