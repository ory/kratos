// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package registration

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPrepareOrganizations(t *testing.T) {
	tests := []struct {
		name           string
		flow           *Flow
		query          url.Values
		expectedFilter []StrategyFilter
	}{
		{
			name:           "valid organization ID in flow",
			flow:           &Flow{OrganizationID: uuid.NullUUID{UUID: uuid.FromStringOrNil("123e4567-e89b-12d3-a456-426614174000"), Valid: true}},
			query:          url.Values{},
			expectedFilter: organizationFilter,
		},
		{
			name:           "valid organization ID in query",
			flow:           &Flow{},
			query:          url.Values{"organization": {"123e4567-e89b-12d3-a456-426614174000"}},
			expectedFilter: organizationFilter,
		},
		{
			name:           "invalid organization ID in query",
			flow:           &Flow{},
			query:          url.Values{"organization": {"invalid-uuid"}},
			expectedFilter: []StrategyFilter{},
		},
		{
			name:           "missing organization ID",
			flow:           &Flow{},
			query:          url.Values{},
			expectedFilter: []StrategyFilter{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{URL: &url.URL{RawQuery: tt.query.Encode()}}
			result := PrepareOrganizations(req.WithContext(context.Background()), tt.flow)
			assert.Equal(t, tt.expectedFilter, result)
		})
	}
}
