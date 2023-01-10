// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package password

import (
	_ "embed"
)

//go:embed .schema/login.schema.json
var loginSchema []byte

//go:embed .schema/registration.schema.json
var registrationSchema []byte

//go:embed .schema/settings.schema.json
var settingsSchema []byte
