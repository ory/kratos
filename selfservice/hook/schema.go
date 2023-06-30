package hook

import _ "embed"

//go:embed .schema/verification.schema.json
var verificationMethodSchema []byte
