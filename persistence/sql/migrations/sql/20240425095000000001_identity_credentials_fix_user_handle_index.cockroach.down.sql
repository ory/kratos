CREATE INVERTED INDEX identity_credentials_user_handle_idx
    ON identity_credentials (config)
    WHERE config ->> 'user_handle' IS NOT NULL;