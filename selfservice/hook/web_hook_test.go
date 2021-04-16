package hook

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"strings"
	"testing"
)

type testData struct {
	ID  uuid.UUID `json:"id"`
	Foo string    `json:"foo"`
}

func TestJsonNet(t *testing.T) {
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

func TestParseConfigWithBasicAuth(t *testing.T) {
	var rawJsonBasicAuth = `{
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
	json.Unmarshal([]byte(rawJsonBasicAuth), conf)

	assert.Equal(t, "https://test.kratos.ory.sh/after_registration_hook", conf.Url)
	assert.Equal(t, "POST", conf.Method)
	assert.NotNil(t, conf.Auth)
	assert.Equal(t, "basic-auth", conf.Auth.Type)
	assert.IsTypef(t, &basicAuthConfig{}, conf.Auth.AuthConfig, "Auth should be an Basic Auth!")

	req := http.Request{Header: map[string][]string{}}
	conf.Auth.AuthConfig.apply(&req)
	actualValue := req.Header.Get("Authorization")
	expectedValue := "Basic " + base64.RawStdEncoding.EncodeToString([]byte("test-api-user:secret"))
	assert.Equal(t, expectedValue, actualValue)
}
