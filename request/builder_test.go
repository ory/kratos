// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package request

import (
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/x"
	"github.com/ory/x/jsonnetsecure"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/otelx"
)

type testRequestBody struct {
	To   string `json:"to"`
	From string `json:"from"`
	Body string `json:"body"`
}

//go:embed stub/test_body.jsonnet
var testJSONNetTemplate []byte

func TestBuildRequest(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name             string
		method           string
		url              string
		authStrategy     AuthStrategy
		expectedHeader   http.Header
		bodyTemplateURI  string
		body             *testRequestBody
		expectedJSONBody string
		expectedRawBody  string
		config           Config
	}{
		{
			name:            "POST request without auth",
			method:          "POST",
			url:             "https://test.kratos.ory.sh/my_endpoint1",
			authStrategy:    NewNoopAuthStrategy(),
			bodyTemplateURI: "file://./stub/test_body.jsonnet",
			body: &testRequestBody{
				To:   "+15056445993",
				From: "+12288534869",
				Body: "test-sms-body",
			},
			expectedJSONBody: `{
				"body": "test-sms-body",
				"from": "+12288534869",
				"to": "+15056445993"
			}`,
			config: Config{
				URL:         "https://test.kratos.ory.sh/my_endpoint1",
				Method:      "POST",
				TemplateURI: "file://./stub/test_body.jsonnet",
			},
		},
		{
			name:            "POST request with legacy template path",
			method:          "POST",
			url:             "https://test.kratos.ory.sh/my_endpoint1",
			authStrategy:    NewNoopAuthStrategy(),
			bodyTemplateURI: "./stub/test_body.jsonnet",
			body: &testRequestBody{
				To:   "+15056445993",
				From: "+12288534869",
				Body: "test-sms-body",
			},
			expectedJSONBody: `{
				"body": "test-sms-body",
				"from": "+12288534869",
				"to": "+15056445993"
			}`,
			config: Config{
				URL:         "https://test.kratos.ory.sh/my_endpoint1",
				Method:      "POST",
				TemplateURI: "./stub/test_body.jsonnet",
			},
		},
		{
			name:            "POST request with base64 encoded template path",
			method:          "POST",
			url:             "https://test.kratos.ory.sh/my_endpoint1",
			authStrategy:    NewNoopAuthStrategy(),
			bodyTemplateURI: "base64://" + base64.StdEncoding.EncodeToString(testJSONNetTemplate),
			body: &testRequestBody{
				To:   "+15056445993",
				From: "+12288534869",
				Body: "test-sms-body",
			},
			expectedJSONBody: `{
				"body": "test-sms-body",
				"from": "+12288534869",
				"to": "+15056445993"
			}`,
			config: Config{
				URL:         "https://test.kratos.ory.sh/my_endpoint1",
				Method:      "POST",
				TemplateURI: "base64://" + base64.StdEncoding.EncodeToString(testJSONNetTemplate),
			},
		},
		{
			name:            "POST request with custom header",
			method:          "POST",
			url:             "https://test.kratos.ory.sh/my_endpoint2",
			authStrategy:    NewNoopAuthStrategy(),
			expectedHeader:  map[string][]string{"Custom-Header": {"test"}},
			bodyTemplateURI: "file://./stub/test_body.jsonnet",
			body: &testRequestBody{
				To:   "+12127110378",
				From: "+15822228108",
				Body: "test-sms-body",
			},
			expectedJSONBody: `{
				"body": "test-sms-body",
				"from": "+15822228108",
				"to": "+12127110378"
			}`,
			config: Config{
				URL:    "https://test.kratos.ory.sh/my_endpoint2",
				Method: "POST",
				Headers: map[string]string{
					"Custom-Header": "test",
				},
				TemplateURI: "file://./stub/test_body.jsonnet",
			},
		},
		{
			name:            "GET request with body",
			method:          "GET",
			url:             "https://test.kratos.ory.sh/my_endpoint3",
			authStrategy:    NewBasicAuthStrategy("test-api-user", "secret"),
			bodyTemplateURI: "file://./stub/test_body.jsonnet",
			body: &testRequestBody{
				To:   "+14134242223",
				From: "+13104661805",
				Body: "test-sms-body",
			},
			expectedJSONBody: `{
				"body": "test-sms-body",
				"from": "+13104661805",
				"to": "+14134242223"
			}`,
			config: Config{
				URL:    "https://test.kratos.ory.sh/my_endpoint3",
				Method: "GET",
				Auth: AuthConfig{
					Type: "basic_auth",
					Config: map[string]any{
						"user":     "test-api-user",
						"password": "secret",
					},
				},
				TemplateURI: "file://./stub/test_body.jsonnet",
			},
		},
		{
			name:         "GET request without body",
			method:       "GET",
			url:          "https://test.kratos.ory.sh/my_endpoint4",
			authStrategy: NewBasicAuthStrategy("test-api-user", "secret"),
			config: Config{
				URL:    "https://test.kratos.ory.sh/my_endpoint4",
				Method: "GET",
				Auth: AuthConfig{
					Type: "basic_auth",
					Config: map[string]any{
						"user":     "test-api-user",
						"password": "secret",
					},
				},
			},
		},
		{
			name:            "DELETE request with body",
			method:          "DELETE",
			url:             "https://test.kratos.ory.sh/my_endpoint5",
			authStrategy:    NewAPIKeyStrategy("header", "my-api-key", "secret"),
			bodyTemplateURI: "file://./stub/test_body.jsonnet",
			body: &testRequestBody{
				To:   "+12235499085",
				From: "+14253787846",
				Body: "test-sms-body",
			},
			expectedJSONBody: `{
				"body": "test-sms-body",
				"from": "+14253787846",
				"to": "+12235499085"
			}`,
			config: Config{
				URL:         "https://test.kratos.ory.sh/my_endpoint5",
				Method:      "DELETE",
				TemplateURI: "file://./stub/test_body.jsonnet",
				Auth: AuthConfig{
					Type: "api_key",
					Config: map[string]any{
						"in":    "header",
						"name":  "my-api-key",
						"value": "secret",
					},
				},
			},
		},
		{
			name:            "POST request with urlencoded body",
			method:          "POST",
			url:             "https://test.kratos.ory.sh/my_endpoint6",
			bodyTemplateURI: "file://./stub/test_body.jsonnet",
			authStrategy:    NewAPIKeyStrategy("cookie", "my-api-key", "secret"),
			expectedHeader:  map[string][]string{"Content-Type": {ContentTypeForm}},
			body: &testRequestBody{
				To:   "+14134242223",
				From: "+13104661805",
				Body: "test-sms-body",
			},
			expectedRawBody: "body=test-sms-body&from=%2B13104661805&to=%2B14134242223",
			config: Config{
				URL:         "https://test.kratos.ory.sh/my_endpoint6",
				Method:      "POST",
				TemplateURI: "file://./stub/test_body.jsonnet",
				Headers: map[string]string{
					"Content-Type": ContentTypeForm,
				},
				Auth: AuthConfig{
					Type: "api_key",
					Config: map[string]any{
						"in":    "cookie",
						"name":  "my-api-key",
						"value": "secret",
					},
				},
			},
		},
		{
			name:            "POST request with default body type",
			method:          "POST",
			url:             "https://test.kratos.ory.sh/my_endpoint7",
			bodyTemplateURI: "file://./stub/test_body.jsonnet",
			authStrategy:    NewBasicAuthStrategy("test-api-user", "secret"),
			expectedHeader:  map[string][]string{"Content-Type": {ContentTypeJSON}},
			body: &testRequestBody{
				To:   "+14134242223",
				From: "+13104661805",
				Body: "test-sms-body",
			},
			expectedJSONBody: `{
				"body": "test-sms-body",
				"from": "+13104661805",
				"to": "+14134242223"
			}`,
			config: Config{
				URL:         "https://test.kratos.ory.sh/my_endpoint7",
				Method:      "POST",
				TemplateURI: "file://./stub/test_body.jsonnet",
				Auth: AuthConfig{
					Type: "basic_auth",
					Config: map[string]any{
						"user":     "test-api-user",
						"password": "secret",
					},
				},
			},
		},
	} {
		t.Run("request-type="+tc.name, func(t *testing.T) {
			t.Parallel()

			rb, err := NewBuilder(context.Background(), &tc.config, newTestDependencyProvider(t))
			require.NoError(t, err)

			assert.Equal(t, tc.bodyTemplateURI, rb.Config.TemplateURI)
			assert.Equal(t, tc.authStrategy, rb.Config.auth)

			req, err := rb.BuildRequest(context.Background(), tc.body)
			require.NoError(t, err)

			assert.Equal(t, tc.url, req.URL.String())
			assert.Equal(t, tc.method, req.Method)

			if tc.expectedJSONBody != "" {
				requestBody, err := req.BodyBytes()
				require.NoError(t, err)

				assert.JSONEq(t, tc.expectedJSONBody, string(requestBody))
			} else if tc.expectedRawBody != "" {
				requestBody, err := req.BodyBytes()
				require.NoError(t, err)

				assert.Equal(t, tc.expectedRawBody, string(requestBody))
			}

			if tc.expectedHeader != nil {
				mustContainHeader(t, tc.expectedHeader, req.Header)
			}
		})
	}

	t.Run("cancel request", func(t *testing.T) {
		rb, err := NewBuilder(context.Background(), &Config{
			URL:         "https://test.kratos.ory.sh/my_endpoint6",
			Method:      "POST",
			TemplateURI: "file://./stub/cancel_body.jsonnet",
		}, newTestDependencyProvider(t))
		require.NoError(t, err)

		_, err = rb.BuildRequest(context.Background(), json.RawMessage(`{}`))
		require.ErrorIs(t, err, ErrCancel)
	})
}

type testDependencyProvider struct {
	x.SimpleLoggerWithClient
	*jsonnetsecure.TestProvider
}

func newTestDependencyProvider(t *testing.T) *testDependencyProvider {
	return &testDependencyProvider{
		SimpleLoggerWithClient: x.SimpleLoggerWithClient{
			L: logrusx.New("kratos", "test"),
			T: otelx.NewNoop(nil, nil),
		},
		TestProvider: jsonnetsecure.NewTestProvider(t),
	}
}

func mustContainHeader(t *testing.T, expected http.Header, actual http.Header) {
	t.Helper()
	for k := range expected {
		require.Contains(t, actual, k)
		assert.Equal(t, expected[k], actual[k])
	}
}
