// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package idfirst

import (
	_ "embed"
)

//go:embed .schema/login.schema.json
var loginSchema []byte
