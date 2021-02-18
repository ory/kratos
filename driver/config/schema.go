package config

import _ "embed"

//go:embed .schema/config.schema.json
var ValidationSchema []byte
