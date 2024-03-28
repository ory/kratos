// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hooktest

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/x/configx"
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

// SetConfig adds the webhook to the list of hooks for the given key and restores
// the original configuration after the test.
func (s *Server) SetConfig(t *testing.T, conf *configx.Provider, key string) {
	var newValue []config.SelfServiceHook
	original := conf.Get(key)
	if originalHooks, ok := original.([]config.SelfServiceHook); ok {
		newValue = slices.Clone(originalHooks)
	}
	require.NoError(t, conf.Set(key, append(newValue, s.HookConfig())))
	t.Cleanup(func() {
		_ = conf.Set(key, original)
	})
}
