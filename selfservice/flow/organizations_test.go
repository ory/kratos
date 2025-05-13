// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package flow

import (
	"context"
	"net/url"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestParseOrganizationFromURLQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    url.Values
		expected uuid.NullUUID
	}{
		{
			name:     "valid organization ID",
			query:    url.Values{"organization": {"123e4567-e89b-12d3-a456-426614174000"}},
			expected: uuid.NullUUID{UUID: uuid.FromStringOrNil("123e4567-e89b-12d3-a456-426614174000"), Valid: true},
		},
		{
			name:     "invalid organization ID",
			query:    url.Values{"organization": {"invalid-uuid"}},
			expected: uuid.NullUUID{Valid: false},
		},
		{
			name:     "missing organization ID",
			query:    url.Values{},
			expected: uuid.NullUUID{Valid: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result := ParseOrganizationFromURLQuery(ctx, tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}
