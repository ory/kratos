// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cliclient

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/kratos/x"
)

func TestNormalizePhoneIdentifier(t *testing.T) {
	for _, tc := range []struct {
		name       string
		input      string
		expected   string
		shouldSkip bool
	}{
		{
			name:     "German phone with spaces",
			input:    "+49 176 671 11 638",
			expected: "+4917667111638",
		},
		{
			name:     "German phone with dashes",
			input:    "+49-176-671-11-638",
			expected: "+4917667111638",
		},
		{
			name:       "Already normalized E.164",
			input:      "+4917667111638",
			expected:   "+4917667111638",
			shouldSkip: true,
		},
		{
			name:     "US phone with parens",
			input:    "+1 (415) 555-2671",
			expected: "+14155552671",
		},
		{
			name:     "Malaysian phone with extra zero",
			input:    "+600136866982",
			expected: "+60136866982",
		},
		{
			name:       "Not a valid phone number",
			input:      "+invalid",
			expected:   "+invalid",
			shouldSkip: true,
		},
		{
			name:       "Email address (not a phone)",
			input:      "+foo@bar.com",
			expected:   "+foo@bar.com",
			shouldSkip: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			normalized := x.NormalizePhoneIdentifier(tc.input)
			assert.Equal(t, tc.expected, normalized)
			if tc.shouldSkip {
				assert.Equal(t, tc.input, normalized, "should be unchanged (skip)")
			} else {
				assert.NotEqual(t, tc.input, normalized, "should be changed (update)")
			}
		})
	}
}
