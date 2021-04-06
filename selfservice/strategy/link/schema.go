package link

import (
	_ "embed"
)

//go:embed .schema/email.schema.json
var emailSchema []byte

//go:embed .schema/recovery.schema.json
var methodSchema []byte
