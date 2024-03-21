// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import "strings"

func EscapeLikePattern(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, "\\", "\\\\"), "%", "\\%"), "_", "\\_")
}
