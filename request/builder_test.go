// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package request

import (
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/x"
	"github.com/ory/x/jsonnetsecure"
	"github.com/ory/x/logrusx"
)

type testRequestBody struct {
	To   string
	From string
	Body string
}

//go:embed stub/test_body.jsonnet
var testJSONNetTemplate []byte

func TestBuildRequest(t *testing.T) {
	for _, tc := range []struct {
		name            string
		method          string
		url             string
		authStrategy    string
		expectedHeader  http.Header
		bodyTemplateURI string
		body            *testRequestBody
		expectedBody    string
		rawConfig       string
	}{
		{
			name:            "POST request without auth",
			method:          "POST",
			url:             "https://test.kratos.ory.sh/my_endpoint1",
			authStrategy:    "", // noop strategy
			bodyTemplateURI: "file://./stub/test_body.jsonnet",
			body: &testRequestBody{
				To:   "+15056445993",
				From: "+12288534869",
				Body: "test-sms-body",
			},
			expectedBody: "{\n   \"Body\": \"test-sms-body\",\n   \"From\": \"+12288534869\",\n   \"To\": \"+15056445993\"\n}\n",
			rawConfig: `{
				"url": "https://test.kratos.ory.sh/my_endpoint1",
				"method": "POST",
				"body": "file://./stub/test_body.jsonnet"
			}`,
		},
		{
			name:            "POST request with legacy template path",
			method:          "POST",
			url:             "https://test.kratos.ory.sh/my_endpoint1",
			bodyTemplateURI: "./stub/test_body.jsonnet",
			body: &testRequestBody{
				To:   "+15056445993",
				From: "+12288534869",
				Body: "test-sms-body",
			},
			expectedBody: "{\n   \"Body\": \"test-sms-body\",\n   \"From\": \"+12288534869\",\n   \"To\": \"+15056445993\"\n}\n",
			rawConfig: `{
				"url": "https://test.kratos.ory.sh/my_endpoint1",
				"method": "POST",
				"body": "./stub/test_body.jsonnet"
			}`,
		},
		{
			name:            "POST request with base64 encoded template path",
			method:          "POST",
			url:             "https://test.kratos.ory.sh/my_endpoint1",
			bodyTemplateURI: "base64://" + base64.StdEncoding.EncodeToString(testJSONNetTemplate),
			body: &testRequestBody{
				To:   "+15056445993",
				From: "+12288534869",
				Body: "test-sms-body",
			},
			expectedBody: "{\n   \"Body\": \"test-sms-body\",\n   \"From\": \"+12288534869\",\n   \"To\": \"+15056445993\"\n}\n",
			rawConfig: fmt.Sprintf(
				`{
				"url": "https://test.kratos.ory.sh/my_endpoint1",
				"method": "POST",
				"body": "base64://%s"
			}`, base64.StdEncoding.EncodeToString(testJSONNetTemplate),
			),
		},
		{
			name:            "POST request with custom header",
			method:          "POST",
			url:             "https://test.kratos.ory.sh/my_endpoint2",
			authStrategy:    "",
			expectedHeader:  map[string][]string{"Custom-Header": {"test"}},
			bodyTemplateURI: "file://./stub/test_body.jsonnet",
			body: &testRequestBody{
				To:   "+12127110378",
				From: "+15822228108",
				Body: "test-sms-body",
			},
			expectedBody: "{\n   \"Body\": \"test-sms-body\",\n   \"From\": \"+15822228108\",\n   \"To\": \"+12127110378\"\n}\n",
			rawConfig: `{
				"url": "https://test.kratos.ory.sh/my_endpoint2",
				"method": "POST",
				"headers": {
					"Custom-Header": "test"
				},
				"body": "file://./stub/test_body.jsonnet"
			}`,
		},
		{
			name:            "GET request with body",
			method:          "GET",
			url:             "https://test.kratos.ory.sh/my_endpoint3",
			authStrategy:    "basic_auth",
			bodyTemplateURI: "file://./stub/test_body.jsonnet",
			body: &testRequestBody{
				To:   "+14134242223",
				From: "+13104661805",
				Body: "test-sms-body",
			},
			expectedBody: "{\n   \"Body\": \"test-sms-body\",\n   \"From\": \"+13104661805\",\n   \"To\": \"+14134242223\"\n}\n",
			rawConfig: `{
				"url": "https://test.kratos.ory.sh/my_endpoint3",
				"method": "GET",
				"auth": {
					"type": "basic_auth",
					"config": {
						"user": "test-api-user",
						"password": "secret"
					}
				},
				"body": "file://./stub/test_body.jsonnet"
			}`,
		},
		{
			name:         "GET request without body",
			method:       "GET",
			url:          "https://test.kratos.ory.sh/my_endpoint4",
			authStrategy: "basic_auth",
			rawConfig: `{
				"url": "https://test.kratos.ory.sh/my_endpoint4",
				"method": "GET",
				"auth": {
					"type": "basic_auth",
					"config": {
						"user": "test-api-user",
						"password": "secret"
					}
				}
			}`,
		},
		{
			name:            "DELETE request with body",
			method:          "DELETE",
			url:             "https://test.kratos.ory.sh/my_endpoint5",
			authStrategy:    "api_key",
			bodyTemplateURI: "file://./stub/test_body.jsonnet",
			body: &testRequestBody{
				To:   "+12235499085",
				From: "+14253787846",
				Body: "test-sms-body",
			},
			expectedBody: "{\n   \"Body\": \"test-sms-body\",\n   \"From\": \"+14253787846\",\n   \"To\": \"+12235499085\"\n}\n",
			rawConfig: `{
				"url": "https://test.kratos.ory.sh/my_endpoint5",
				"method": "DELETE",
				"body": "file://./stub/test_body.jsonnet",
				"auth": {
					"type": "api_key",
					"config": {
						"in": "header",
						"name": "my-api-key",
						"value": "secret"
					}
				}
			}`,
		},
		{
			name:            "POST request with urlencoded body",
			method:          "POST",
			url:             "https://test.kratos.ory.sh/my_endpoint6",
			bodyTemplateURI: "file://./stub/test_body.jsonnet",
			authStrategy:    "api_key",
			expectedHeader:  map[string][]string{"Content-Type": {ContentTypeForm}},
			body: &testRequestBody{
				To:   "+14134242223",
				From: "+13104661805",
				Body: "test-sms-body",
			},
			expectedBody: "Body=test-sms-body&From=%2B13104661805&To=%2B14134242223",
			rawConfig: `{
				"url": "https://test.kratos.ory.sh/my_endpoint6",
				"method": "POST",
				"body": "file://./stub/test_body.jsonnet",
				"headers": {
					"Content-Type": "application/x-www-form-urlencoded"
				},
				"auth": {
					"type": "api_key",
					"config": {
						"in": "cookie",
						"name": "my-api-key",
						"value": "secret"
					}
				}
			}`,
		},
		{
			name:            "POST request with default body type",
			method:          "POST",
			url:             "https://test.kratos.ory.sh/my_endpoint7",
			bodyTemplateURI: "file://./stub/test_body.jsonnet",
			authStrategy:    "basic_auth",
			expectedHeader:  map[string][]string{"Content-Type": {ContentTypeJSON}},
			body: &testRequestBody{
				To:   "+14134242223",
				From: "+13104661805",
				Body: "test-sms-body",
			},
			expectedBody: "{\n   \"Body\": \"test-sms-body\",\n   \"From\": \"+13104661805\",\n   \"To\": \"+14134242223\"\n}\n",
			rawConfig: `{
				"url": "https://test.kratos.ory.sh/my_endpoint7",
				"method": "POST",
				"body": "file://./stub/test_body.jsonnet",
				"auth": {
					"type": "basic_auth",
					"config": {
						"user": "test-api-user",
						"password": "secret"
					}
				}
			}`,
		},
	} {
		t.Run(
			"request-type="+tc.name, func(t *testing.T) {
				rb, err := NewBuilder(json.RawMessage(tc.rawConfig), newTestDependencyProvider(t))
				require.NoError(t, err)

				assert.Equal(t, tc.bodyTemplateURI, rb.Config.TemplateURI)
				assert.Equal(t, tc.authStrategy, rb.Config.Auth.Type)

				req, err := rb.BuildRequest(context.Background(), tc.body)
				require.NoError(t, err)

				assert.Equal(t, tc.url, req.URL.String())
				assert.Equal(t, tc.method, req.Method)

				if tc.body != nil {
					requestBody, err := req.BodyBytes()
					require.NoError(t, err)

					assert.Equal(t, tc.expectedBody, string(requestBody))
				}

				if tc.expectedHeader != nil {
					mustContainHeader(t, tc.expectedHeader, req.Header)
				}
			},
		)
	}

	t.Run(
		"cancel request", func(t *testing.T) {
			rb, err := NewBuilder(json.RawMessage(
				`{
	"url": "https://test.kratos.ory.sh/my_endpoint6",
	"method": "POST",
	"body": "file://./stub/cancel_body.jsonnet"
}`,
			), newTestDependencyProvider(t))
			require.NoError(t, err)

			_, err = rb.BuildRequest(context.Background(), json.RawMessage(`{}`))
			require.ErrorIs(t, err, ErrCancel)
		},
	)
}

type testDependencyProvider struct {
	x.SimpleLoggerWithClient
	*jsonnetsecure.TestProvider
}

func newTestDependencyProvider(t *testing.T) *testDependencyProvider {
	return &testDependencyProvider{
		SimpleLoggerWithClient: x.SimpleLoggerWithClient{
			L: logrusx.New("kratos", "test"),
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
