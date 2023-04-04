// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

// Coalesce returns the first non-empty string value
func Coalesce[T ~string](str ...T) T {
	for _, s := range str {
		if len(s) > 0 {
			return s
		}
	}
	return ""
}
