UPDATE sessions SET authentication_methods='[{"method":"v0.6_legacy_session"}]' WHERE JSON_LENGTH(authentication_methods)=0 AND aal='aal1';
