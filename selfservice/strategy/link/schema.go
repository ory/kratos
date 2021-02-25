package link

import (
	_ "embed"
)

//go:embed .schema/email.schema.json
var emailSchema []byte
