// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEscapeLikePattern(t *testing.T) {
	for name, tc := range map[string]struct {
		input    string
		expected string
	}{
		"empty": {
			input:    "",
			expected: "",
		},
		"no escape": {
			input:    "foo",
			expected: "foo",
		},
		"escape": {
			input:    "foo%bar_baz\\",
			expected: "foo\\%bar\\_baz\\\\",
		},
	} {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.expected, EscapeLikePattern(tc.input))
		})
	}
}
