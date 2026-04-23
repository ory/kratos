// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"strings"
	"unicode"
)

func hasLetters(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func isAllUpperText(s string) bool {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" || !hasLetters(trimmed) {
		return false
	}
	return trimmed == strings.ToUpper(trimmed)
}
