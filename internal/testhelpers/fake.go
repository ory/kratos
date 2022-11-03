// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import "github.com/ory/x/randx"

func RandomEmail() string {
	return randx.MustString(16, randx.Alpha) + "@ory.sh"
}
