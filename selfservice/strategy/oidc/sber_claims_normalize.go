// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"strings"
	"time"
	"unicode"
)

func normalizeSberBirthdateISO8601(v string) string {
	s := strings.TrimSpace(v)
	if s == "" {
		return ""
	}

	layouts := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"02.01.2006",
		"02-01-2006",
		"02/01/2006",
		"2006/01/02",
	}
	for _, layout := range layouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t.UTC().Format(time.RFC3339)
		}
	}

	return ""
}

func normalizeRussianMobilePlus79(phone string) (normalized string) {
	raw := strings.TrimSpace(phone)
	if raw == "" {
		return ""
	}

	cleaned := strings.NewReplacer("(", "", ")", "", " ", "").Replace(raw)
	if strings.Count(cleaned, "+") > 1 {
		return ""
	}

	return cleaned
}

func normalizeNameTitle(s string) string {
	raw := strings.TrimSpace(s)
	if raw == "" {
		return ""
	}

	lower := []rune(strings.ToLower(raw))
	lower[0] = unicode.ToUpper(lower[0])
	return string(lower)
}

func normalizeEmailLower(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
