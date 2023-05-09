UPDATE sessions SET token = LEFT(token, 32) WHERE TRUE;
UPDATE sessions SET logout_token = LEFT(logout_token, 32) WHERE TRUE;
