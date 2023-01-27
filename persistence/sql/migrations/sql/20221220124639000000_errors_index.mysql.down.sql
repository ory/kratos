CREATE INDEX selfservice_errors_nid_idx ON selfservice_errors (id, nid);

-- needed for foreign key constraint, was there before implicitly
CREATE INDEX selfservice_errors_nid_only_idx ON selfservice_errors (nid);

DROP INDEX selfservice_errors_errors_nid_id_idx ON selfservice_errors;
