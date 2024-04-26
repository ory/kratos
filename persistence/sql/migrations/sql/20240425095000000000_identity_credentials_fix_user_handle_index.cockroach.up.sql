CREATE INDEX identity_credentials_config_user_handle_idx
    ON identity_credentials ((config ->> 'user_handle'))
    WHERE config ->> 'user_handle' IS NOT NULL
;
