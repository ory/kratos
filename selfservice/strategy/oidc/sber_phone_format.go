// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import "strings"

// normalizeRussianMobileE164 приводит номер к виду +7 и ровно 10 цифр после семёрки
// (например +79991234567). Пустая строка, если после очистки цифр распознать формат нельзя.
func normalizeRussianMobileE164(phone string) string {
	s := strings.TrimSpace(phone)
	if s == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			b.WriteByte(c)
		}
	}
	d := b.String()

	switch len(d) {
	case 11:
		switch d[0] {
		case '8':
			return "+7" + d[1:]
		case '7':
			return "+" + d
		}
	case 10:
		return "+7" + d
	}

	return ""
}
