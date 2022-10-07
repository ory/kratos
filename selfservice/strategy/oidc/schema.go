// Copyright © 2022 Ory Corp

package oidc

import (
	_ "embed"
)

//go:embed .schema/link.schema.json
var linkSchema []byte
