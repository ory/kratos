// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"strings"

	"github.com/ory/x/randx"
)

func RandomEmail() string {
	return strings.ToLower(randx.MustString(16, randx.Alpha) + "@ory.sh")
}

func RandomPhone() string {
	return "+49151" + randx.MustString(8, randx.Numeric)
}
