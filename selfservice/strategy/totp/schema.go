package totp

import (
	_ "embed"
)

//go:embed .schema/settings.schema.json
var settingsSchema []byte
