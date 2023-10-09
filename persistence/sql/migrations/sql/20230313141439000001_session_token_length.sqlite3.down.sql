UPDATE sessions SET token = substr(token, 0, 32) WHERE TRUE;
UPDATE sessions SET logout_token = substr(logout_token, 0, 32) WHERE TRUE;
