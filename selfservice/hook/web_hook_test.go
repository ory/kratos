package hook

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testData struct {
	ID  uuid.UUID `json:"id"`
	Foo string    `json:"foo"`
}

func TestJsonNetSupport(t *testing.T) {
	id, _ := uuid.NewV1()
	td := testData{id, "Bar"}
	h := NewWebHook(nil, json.RawMessage{})

	b, err := h.createBody("test_body.jsonnet", &td, &td)
	require.NoError(t, err)

	buf := new(strings.Builder)
	io.Copy(buf, b)

	expected := fmt.Sprintf("{\n   \"flow_id\": \"%s\",\n   \"session_id\": \"%s\"\n}\n", td.ID, td.Foo)

	require.Equal(t, expected, buf.String())
}

func TestParseConfigWithoutAuthPart(t *testing.T) {
	var rawConfig = `{
		  "url": "https://test.kratos.ory.sh/after_registration_hook",
		  "method": "POST"
		}
	`
	var conf = webHookConfig{}
	conf.Unmarshal([]byte(rawConfig))

	assert.Equal(t, "https://test.kratos.ory.sh/after_registration_hook", conf.Url)
	assert.Equal(t, "POST", conf.Method)
	assert.NotNil(t, conf.Auth)
	assert.Empty(t, conf.Auth.Type)
	assert.Empty(t, conf.Auth.RawConfig)

	assert.IsTypef(t, &emptyAuthConfig{}, conf.Auth.AuthConfig, "Auth should be a dummy implementation!")

	req := http.Request{Header: map[string][]string{}}
	conf.Auth.AuthConfig.apply(&req)

	assert.Empty(t, req.Header)
}

func TestParseConfigWithBasicAuth(t *testing.T) {
	var rawConfig = `{
		  "url": "https://test.kratos.ory.sh/after_registration_hook",
		  "method": "POST",
		  "auth": {
			"type": "basic-auth",
			"config": {
			  "user": "test-api-user",
			  "password": "secret"
			}
		  }
		}
	`
	var conf = &webHookConfig{}
	conf.Unmarshal([]byte(rawConfig))

	assert.Equal(t, "https://test.kratos.ory.sh/after_registration_hook", conf.Url)
	assert.Equal(t, "POST", conf.Method)
	assert.NotNil(t, conf.Auth)
	assert.Equal(t, "basic-auth", conf.Auth.Type)

	assert.IsTypef(t, &basicAuthConfig{}, conf.Auth.AuthConfig, "Auth should be Basic Auth!")

	req := http.Request{Header: map[string][]string{}}
	conf.Auth.AuthConfig.apply(&req)

	user, pass, _ := req.BasicAuth()

	assert.Equal(t, "test-api-user", user)
	assert.Equal(t, "secret", pass)
}

func TestParseConfigWithApiKeyInHeader(t *testing.T) {
	var rawConfig = `{
	"url": "https://test.kratos.ory.sh/after_registration_hook",
	"method": "POST",
	"auth": {
		"type": "api-key",
		"config": {
			"in": "header",
			"name": "my-api-key",
			"value": "secret"
		}
	}
}
	`
	var conf = &webHookConfig{}
	conf.Unmarshal([]byte(rawConfig))

	assert.Equal(t, "https://test.kratos.ory.sh/after_registration_hook", conf.Url)
	assert.Equal(t, "POST", conf.Method)
	assert.NotNil(t, conf.Auth)
	assert.Equal(t, "api-key", conf.Auth.Type)
	assert.IsTypef(t, &apiKeyConfig{}, conf.Auth.AuthConfig, "Auth should be an api key config!")

	req := http.Request{Header: map[string][]string{}}
	conf.Auth.AuthConfig.apply(&req)
	actualValue := req.Header.Get("my-api-key")
	assert.Equal(t, "secret", actualValue)
}
