// Copyright © 2022 Ory Corp

package lookup

import (
	_ "embed"
)

//go:embed .schema/login.schema.json
var loginSchema []byte

//go:embed .schema/settings.schema.json
var settingsSchema []byte
