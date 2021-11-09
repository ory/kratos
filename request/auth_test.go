package request

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthStrategy(t *testing.T) {
	for _, tc := range map[string]struct {
		name     string
		config   string
		expected AuthStrategy
	}{
		"noop": {
			name:     "",
			config:   "",
			expected: &noopAuthStrategy{},
		},
		"basic_auth": {
			name: "basic_auth",
			config: `{
				"user": "test-api-user",
				"password": "secret"
			}`,
			expected: &basicAuthStrategy{},
		},
		"api-key/header": {
			name: "api_key",
			config: `{
				"in": "header",
				"name": "my-api-key",
				"value": "secret"
			}`,
			expected: &apiKeyStrategy{},
		},
		"api-key/cookie": {
			name: "api_key",
			config: `{
				"in": "cookie",
				"name": "my-api-key",
				"value": "secret"
			}`,
			expected: &apiKeyStrategy{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			strategy, err := authStrategy(tc.name, json.RawMessage(tc.config))
			require.NoError(t, err)

			assert.IsTypef(t, tc.expected, strategy, "auth strategy should be of the expected type")
		})
	}
}
