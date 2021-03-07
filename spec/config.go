package spec

import _ "embed"

//go:embed .schema/config.json
var ConfigValidationSchema []byte
