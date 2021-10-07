package oidc

import (
	_ "embed"
)

//go:embed .schema/login.schema.json
var loginSchema []byte

//go:embed .schema/registration.schema.json
var registrationSchema []byte
