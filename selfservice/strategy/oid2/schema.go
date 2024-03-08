// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oid2

import (
	_ "embed"
)

//go:embed .schema/link.schema.json
var linkSchema []byte
