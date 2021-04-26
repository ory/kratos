package hook

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNoopAuthStrategy(t *testing.T) {
	req := http.Request{Header: map[string][]string{}}
	auth := noopAuthStrategy{}

	auth.apply(&req)

	assert.Empty(t, req.Header, "Empty auth strategy shall not modify any request headers")
}

func TestBasicAuthStrategy(t *testing.T) {
	req := http.Request{Header: map[string][]string{}}
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
	req := http.Request{Header: map[string][]string{}}
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
	req := http.Request{Header: map[string][]string{}}
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

func TestJsonNetSupport(t *testing.T) {
	id, _ := uuid.NewV1()
	td := struct {
		ID  uuid.UUID `json:"id"`
		Foo string    `json:"foo"`
	}{id, "Bar"}
	headers := http.Header{}
	headers.Add("Some-Header", "Some-Value")
	headers.Add("Cookie", "c1=v1")
	headers.Add("Cookie", "c2=v2")
	data := &templateContext{
		Flow:           &td,
		RequestHeaders: headers,
		RequestMethod:  "POST",
		RequestUrl:     "https://test.kratos.ory.sh/some-test-path",
		Session:        &td,
	}

	b, err := createBody("test_body.jsonnet", data)
	assert.NoError(t, err)

	buf := new(strings.Builder)
	io.Copy(buf, b)

	expected := fmt.Sprintf(`
		{
			"flow_id": "%s",
			"session_id": "%s",
			"headers": {
				"Cookie": ["%s", "%s"],
				"Some-Header": ["%s"]
			},
			"method": "%s",
			"url": "%s"
		}`,
		td.ID,
		td.Foo,
		data.RequestHeaders.Values("Cookie")[0],
		data.RequestHeaders.Values("Cookie")[1],
		data.RequestHeaders.Get("Some-Header"),
		data.RequestMethod,
		data.RequestUrl)

	assert.JSONEq(t, expected, buf.String())
}

func TestWebHookConfig(t *testing.T) {
	for _, tc := range []struct {
		strategy     string
		method       string
		url          string
		body         string
		rawConfig    string
		authStrategy AuthStrategy
	}{
		{
			strategy: "empty",
			method:   "POST",
			url:      "https://test.kratos.ory.sh/my_hook1",
			body:     "/path/to/my/jsonnet1.file",
			rawConfig: `{
				"url": "https://test.kratos.ory.sh/my_hook1",
				"method": "POST",
				"body": "/path/to/my/jsonnet1.file"
			}`,
			authStrategy: &noopAuthStrategy{},
		},
		{
			strategy: "basic-auth",
			method:   "GET",
			url:      "https://test.kratos.ory.sh/my_hook2",
			body:     "/path/to/my/jsonnet2.file",
			rawConfig: `{
				"url": "https://test.kratos.ory.sh/my_hook2",
				"method": "GET",
				"body": "/path/to/my/jsonnet2.file",
				"auth": {
					"type": "basic-auth",
					"config": {
						"user": "test-api-user",
						"password": "secret"
					}
				}
			}`,
			authStrategy: &basicAuthStrategy{},
		},
		{
			strategy: "api-key/header",
			method:   "DELETE",
			url:      "https://test.kratos.ory.sh/my_hook3",
			body:     "/path/to/my/jsonnet3.file",
			rawConfig: `{
				"url": "https://test.kratos.ory.sh/my_hook3",
				"method": "DELETE",
				"body": "/path/to/my/jsonnet3.file",
				"auth": {
					"type": "api-key",
					"config": {
						"in": "header",
						"name": "my-api-key",
						"value": "secret"
					}
				}
			}`,
			authStrategy: &apiKeyStrategy{},
		},
		{
			strategy: "api-key/cookie",
			method:   "POST",
			url:      "https://test.kratos.ory.sh/my_hook4",
			body:     "/path/to/my/jsonnet4.file",
			rawConfig: `{
				"url": "https://test.kratos.ory.sh/my_hook4",
				"method": "POST",
				"body": "/path/to/my/jsonnet4.file",
				"auth": {
					"type": "api-key",
					"config": {
						"in": "cookie",
						"name": "my-api-key",
						"value": "secret"
					}
				}
			}`,
			authStrategy: &apiKeyStrategy{},
		},
	} {
		t.Run("auth-strategy="+tc.strategy, func(t *testing.T) {
			conf, err := newWebHookConfig([]byte(tc.rawConfig))
			assert.Nil(t, err)

			assert.Equal(t, tc.url, conf.url)
			assert.Equal(t, tc.method, conf.method)
			assert.Equal(t, tc.body, conf.templatePath)
			assert.NotNil(t, conf.auth)
			assert.IsTypef(t, tc.authStrategy, conf.auth, "Auth should be of the expected type")
		})
	}
}
