// Copyright © 2022 Ory Corp

package code

import (
	_ "embed"
)

//go:embed .schema/recovery.schema.json
var recoveryMethodSchema []byte
