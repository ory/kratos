CREATE INDEX selfservice_errors_errors_nid_id_idx ON selfservice_errors (nid, id);

-- This index is not needed anymore, because the primary ID index together with the new index cover all queries.
DROP INDEX selfservice_errors_nid_idx ON selfservice_errors;
