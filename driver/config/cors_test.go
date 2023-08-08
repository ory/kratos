// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"net/http"
	"testing"

	"github.com/rs/cors"
	"github.com/stretchr/testify/assert"
)

func TestAllowOrigin(t *testing.T) {
	for _, tc := range []struct {
		name                string
		allowedOrigins      []string
		expect, expectOther bool
	}{
		{
			name:           "empty",
			allowedOrigins: []string{},
			expect:         true,
			expectOther:    true,
		},
		{
			name:           "wildcard",
			allowedOrigins: []string{"https://example.com", "*"},
			expect:         true,
			expectOther:    true,
		},
		{
			name:           "exact",
			allowedOrigins: []string{"https://www.ory.sh"},
			expect:         true,
		},
		{
			name:           "wildcard in the beginning",
			allowedOrigins: []string{"*.ory.sh"},
			expect:         true,
		},
		{
			name:           "wildcard in the middle",
			allowedOrigins: []string{"https://*.ory.sh"},
			expect:         true,
		},
		{
			name:           "wildcard in the end",
			allowedOrigins: []string{"https://www.ory.*"},
			expect:         true,
		},
		{
			name:           "second wildcard is ignored",
			allowedOrigins: []string{"https://*.ory.*"},
			expect:         false,
		},
		{
			name:           "multiple exact origins",
			allowedOrigins: []string{"https://example.com", "https://www.ory.sh"},
			expect:         true,
		},
		{
			name:           "multiple wildcard origins",
			allowedOrigins: []string{"https://*.example.com", "https://*.ory.sh"},
			expect:         true,
		},
		{
			name:           "wildcard and exact origins 1",
			allowedOrigins: []string{"https://*.example.com", "https://www.ory.sh"},
			expect:         true,
		},
		{
			name:           "wildcard and exact origins 2",
			allowedOrigins: []string{"https://example.com", "https://*.ory.sh"},
			expect:         true,
		},
		{
			name:           "unrelated exact origins",
			allowedOrigins: []string{"https://example.com", "https://example.org"},
			expect:         false,
		},
		{
			name:           "unrelated wildcard origins",
			allowedOrigins: []string{"https://*.example.com", "https://*.example.org"},
			expect:         false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, allowOrigin(tc.allowedOrigins, "https://www.ory.sh"))

			assert.Equal(t, tc.expectOther, allowOrigin(tc.allowedOrigins, "https://google.com"))

			// check for consistency with rs/cors
			assert.Equal(t, tc.expect, cors.New(cors.Options{AllowedOrigins: tc.allowedOrigins}).
				OriginAllowed(&http.Request{Header: http.Header{"Origin": []string{"https://www.ory.sh"}}}))

			assert.Equal(t, tc.expectOther, cors.New(cors.Options{AllowedOrigins: tc.allowedOrigins}).
				OriginAllowed(&http.Request{Header: http.Header{"Origin": []string{"https://google.com"}}}))
		})
	}
}
