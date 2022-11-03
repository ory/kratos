// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	_ "embed"
)

//go:embed .schema/settings.schema.json
var settingsSchema []byte
