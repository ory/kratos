// Copyright © 2022 Ory Corp

package profile

import (
	_ "embed"
)

//go:embed .schema/settings.schema.json
var settingsSchema []byte
