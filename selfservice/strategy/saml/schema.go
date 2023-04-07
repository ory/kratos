package saml

import (
	_ "embed"
)

//go:embed .schema/link.schema.json
var linkSchema []byte
