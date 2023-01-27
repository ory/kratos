// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	_ "embed"
)

//go:embed .schema/link.schema.json
var linkSchema []byte
