package link

import (
	_ "embed"
)

//go:embed .schema/recovery.schema.json
var recoveryMethodSchema []byte

//go:embed .schema/verification.schema.json
var verificationMethodSchema []byte
