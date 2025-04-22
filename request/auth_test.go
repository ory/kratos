// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package request

import (
	"net/http"
	"testing"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoopAuthStrategy(t *testing.T) {
	req := retryablehttp.Request{Request: &http.Request{Header: map[string][]string{}}}
	auth := noopAuthStrategy{}

	auth.apply(&req)

	assert.Empty(t, req.Header, "Empty auth strategy shall not modify any request headers")
}

func TestBasicAuthStrategy(t *testing.T) {
	req := retryablehttp.Request{Request: &http.Request{Header: map[string][]string{}}}
	auth := basicAuthStrategy{
		user:     "test-user",
		password: "test-pass",
	}

	auth.apply(&req)

	assert.Len(t, req.Header, 1)

	user, pass, _ := req.BasicAuth()
	assert.Equal(t, "test-user", user)
	assert.Equal(t, "test-pass", pass)
}

func TestApiKeyInHeaderStrategy(t *testing.T) {
	req := retryablehttp.Request{Request: &http.Request{Header: map[string][]string{}}}
	auth := apiKeyStrategy{
		in:    "header",
		name:  "my-api-key-name",
		value: "my-api-key-value",
	}

	auth.apply(&req)

	require.Len(t, req.Header, 1)

	actualValue := req.Header.Get("my-api-key-name")
	assert.Equal(t, "my-api-key-value", actualValue)
}

func TestApiKeyInCookieStrategy(t *testing.T) {
	req := retryablehttp.Request{Request: &http.Request{Header: map[string][]string{}}}
	auth := apiKeyStrategy{
		in:    "cookie",
		name:  "my-api-key-name",
		value: "my-api-key-value",
	}

	auth.apply(&req)

	cookies := req.Cookies()
	assert.Len(t, cookies, 1)

	assert.Equal(t, "my-api-key-name", cookies[0].Name)
	assert.Equal(t, "my-api-key-value", cookies[0].Value)
}

func TestAuthStrategy(t *testing.T) {
	t.Parallel()

	for _, tc := range map[string]struct {
		name     string
		config   map[string]any
		expected AuthStrategy
	}{
		"noop": {
			name:     "",
			config:   map[string]any{},
			expected: &noopAuthStrategy{},
		},
		"basic_auth": {
			name: "basic_auth",
			config: map[string]any{
				"user":     "test-api-user",
				"password": "secret",
			},
			expected: &basicAuthStrategy{},
		},
		"api-key/header": {
			name: "api_key",
			config: map[string]any{
				"in":    "header",
				"name":  "my-api-key",
				"value": "secret",
			},
			expected: &apiKeyStrategy{},
		},
		"api-key/cookie": {
			name: "api_key",
			config: map[string]any{
				"in":    "cookie",
				"name":  "my-api-key",
				"value": "secret",
			},
			expected: &apiKeyStrategy{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			strategy, err := authStrategy(tc.name, tc.config)
			require.NoError(t, err)

			assert.IsTypef(t, tc.expected, strategy, "auth strategy should be of the expected type")
		})
	}
}
