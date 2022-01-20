package hook

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/sirupsen/logrus/hooks/test"

	"github.com/ory/x/logrusx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/selfservice/flow/login"

	"github.com/stretchr/testify/assert"
)

func TestNoopAuthStrategy(t *testing.T) {
	req := retryablehttp.Request{Request: &http.Request{Header: map[string][]string{}}}
	auth := NoopAuthStrategy{}

	auth.apply(&req)

	assert.Empty(t, req.Header, "Empty auth strategy shall not modify any request headers")
}

func TestBasicAuthStrategy(t *testing.T) {
	req := retryablehttp.Request{Request: &http.Request{Header: map[string][]string{}}}
	auth := BasicAuthStrategy{
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
	auth := ApiKeyStrategy{
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
	auth := ApiKeyStrategy{
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

//go:embed stub/test_body.jsonnet
var testBodyJSONNet []byte

func TestJsonNetSupport(t *testing.T) {
	f := &login.Flow{ID: x.NewUUID()}
	i := identity.NewIdentity("")
	l := logrusx.New("kratos", "test")

	for _, tc := range []struct {
		desc, template string
		data           *templateContext
	}{
		{
			desc:     "simple file URI",
			template: "file://./stub/test_body.jsonnet",
			data: &templateContext{
				Flow: f,
				RequestHeaders: http.Header{
					"Cookie":      []string{"c1=v1", "c2=v2"},
					"Some-Header": []string{"Some-Value"},
				},
				RequestMethod: "POST",
				RequestUrl:    "https://test.kratos.ory.sh/some-test-path",
				Identity:      i,
			},
		},
		{
			desc:     "legacy filepath without scheme",
			template: "./stub/test_body.jsonnet",
			data: &templateContext{
				Flow: f,
				RequestHeaders: http.Header{
					"Cookie":      []string{"c1=v1", "c2=v2"},
					"Some-Header": []string{"Some-Value"},
				},
				RequestMethod: "POST",
				RequestUrl:    "https://test.kratos.ory.sh/some-test-path",
				Identity:      i,
			},
		},
		{
			desc:     "base64 encoded template URI",
			template: "base64://" + base64.StdEncoding.EncodeToString(testBodyJSONNet),
			data: &templateContext{
				Flow: f,
				RequestHeaders: http.Header{
					"Cookie":           []string{"foo=bar"},
					"My-Custom-Header": []string{"Cumstom-Value"},
				},
				RequestMethod: "PUT",
				RequestUrl:    "https://test.kratos.ory.sh/other-test-path",
				Identity:      i,
			},
		},
	} {
		t.Run("case="+tc.desc, func(t *testing.T) {
			b, err := createBody(l, tc.template, tc.data)
			require.NoError(t, err)
			body, err := io.ReadAll(b)
			require.NoError(t, err)

			expected, err := json.Marshal(map[string]interface{}{
				"flow_id":     tc.data.Flow.GetID(),
				"identity_id": tc.data.Identity.ID,
				"headers":     tc.data.RequestHeaders,
				"method":      tc.data.RequestMethod,
				"url":         tc.data.RequestUrl,
			})
			require.NoError(t, err)

			assert.JSONEq(t, string(expected), string(body))
		})
	}

	t.Run("case=warns about legacy usage", func(t *testing.T) {
		hook := test.Hook{}
		l := logrusx.New("kratos", "test", logrusx.WithHook(&hook))

		_, _ = createBody(l, "./foo", nil)

		require.Len(t, hook.Entries, 1)
		assert.Contains(t, hook.LastEntry().Message, "support for filepaths without a 'file://' scheme will be dropped")
	})

	t.Run("case=return non nil body reader on empty templateURI", func(t *testing.T) {
		body, err := createBody(l, "", nil)
		assert.NotNil(t, body)
		assert.Nil(t, err)
	})
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
			authStrategy: &NoopAuthStrategy{},
		},
		{
			strategy: "basic_auth",
			method:   "GET",
			url:      "https://test.kratos.ory.sh/my_hook2",
			body:     "/path/to/my/jsonnet2.file",
			rawConfig: `{
				"url": "https://test.kratos.ory.sh/my_hook2",
				"method": "GET",
				"body": "/path/to/my/jsonnet2.file",
				"auth": {
					"type": "basic_auth",
					"config": {
						"user": "test-api-user",
						"password": "secret"
					}
				}
			}`,
			authStrategy: &BasicAuthStrategy{},
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
					"type": "api_key",
					"config": {
						"in": "header",
						"name": "my-api-key",
						"value": "secret"
					}
				}
			}`,
			authStrategy: &ApiKeyStrategy{},
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
					"type": "api_key",
					"config": {
						"in": "cookie",
						"name": "my-api-key",
						"value": "secret"
					}
				}
			}`,
			authStrategy: &ApiKeyStrategy{},
		},
	} {
		t.Run("auth-strategy="+tc.strategy, func(t *testing.T) {
			conf, err := newWebHookConfig([]byte(tc.rawConfig))
			assert.Nil(t, err)

			assert.Equal(t, tc.url, conf.URL)
			assert.Equal(t, tc.method, conf.Method)
			assert.Equal(t, tc.body, conf.TemplateURI)
			assert.NotNil(t, conf.Auth)
			assert.IsTypef(t, tc.authStrategy, conf.Auth, "Auth should be of the expected type")
		})
	}
}
