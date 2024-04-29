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

// IfThenElse returns t if cond is true, otherwise f
func IfThenElse[T any](cond bool, t, f T) T {
	if cond {
		return t
	}
	return f
}
