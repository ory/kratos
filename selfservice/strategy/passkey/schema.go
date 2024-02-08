// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package passkey

import _ "embed"

//go:embed .schema/registration.schema.json
var registrationSchema []byte

//go:embed .schema/login.schema.json
var loginSchema []byte

//go:embed .schema/settings.schema.json
var settingsSchema []byte
