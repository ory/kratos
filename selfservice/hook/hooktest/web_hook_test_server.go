package hooktest

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/x/ioutilx"
)

var jsonnet = base64.StdEncoding.EncodeToString([]byte("function(ctx) ctx"))

type Server struct {
	*httptest.Server
	LastBody []byte
}

// NewServer returns a new webhook server for testing.
func NewServer() *Server {
	s := new(Server)
	httptestServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.LastBody = ioutilx.MustReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))

	s.Server = httptestServer

	return s
}

// HookConfig returns the hook configuration for calling the webhook server.
func (s *Server) HookConfig() config.SelfServiceHook {
	return config.SelfServiceHook{
		Name: "web_hook",
		Config: []byte(fmt.Sprintf(`
{
	"method": "POST",
	"url": "%s",
	"body": "base64://%s"
}`, s.URL, jsonnet)),
	}
}

func (s *Server) AssertTransientPayload(t *testing.T, expected string) {
	require.NotEmpty(t, s.LastBody)
	actual := gjson.GetBytes(s.LastBody, "flow.transient_payload").String()
	assert.JSONEq(t, expected, actual, "%+v", actual)
}
