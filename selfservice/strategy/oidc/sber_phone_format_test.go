// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeRussianMobileE164(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in, want string
	}{
		{"+7 (999) 123-45-67", "+79991234567"},
		{"89991234567", "+79991234567"},
		{"79991234567", "+79991234567"},
		{"9991234567", "+79991234567"},
		{"  +79991234567  ", "+79991234567"},
		{"", ""},
		{"   ", ""},
		{"abc", ""},
		{"+44 20 7946 0958", ""},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, normalizeRussianMobileE164(tc.in))
		})
	}
}
