package totp

import (
	_ "embed"
)

//go:embed .schema/settings.schema.json
var settingsSchema []byte

//go:embed .schema/login.schema.json
var loginSchema []byte
