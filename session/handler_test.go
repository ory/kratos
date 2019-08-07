package session_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/herodot"

	"github.com/ory/hive/driver/configuration"
	"github.com/ory/hive/internal"
	. "github.com/ory/hive/session"
	"github.com/ory/hive/x"
)

func send(code int) httprouter.Handle {
	return func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
		w.WriteHeader(code)
	}
}

func TestHandler(t *testing.T) {
	t.Run("public", func(t *testing.T) {
		_, reg := internal.NewMemoryRegistry(t)
		r := x.NewRouterPublic()

		r.GET("/set", MockSessionCreateHandler(t, reg))

		h := NewHandler(reg, herodot.NewJSONWriter(nil))
		h.RegisterPublicRoutes(r)
		ts := httptest.NewServer(r)
		defer ts.Close()

		viper.Set(configuration.ViperKeyURLsSelfPublic, ts.URL)

		client := MockCookieClient(t)

		// No cookie yet -> 401
		res, err := client.Get(ts.URL + "/sessions/me")
		require.NoError(t, err)
		assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode)

		// Set cookie
		MockHydrateCookieClient(t, client, ts.URL+"/set")

		// Cookie set -> 200
		res, err = client.Get(ts.URL + "/sessions/me")
		require.NoError(t, err)
		assert.EqualValues(t, http.StatusOK, res.StatusCode)
	})
}

func TestIsNotAuthenticated(t *testing.T) {
	_, reg := internal.NewMemoryRegistry(t)
	r := x.NewRouterPublic()

	r.GET("/set", MockSessionCreateHandler(t, reg))
	r.GET("/public/with-callback", reg.SessionHandler().IsNotAuthenticated(send(http.StatusOK), send(http.StatusBadRequest)))
	r.GET("/public/without-callback", reg.SessionHandler().IsNotAuthenticated(send(http.StatusOK), nil))
	ts := httptest.NewServer(r)
	defer ts.Close()

	sessionClient := MockCookieClient(t)
	MockHydrateCookieClient(t, sessionClient, ts.URL+"/set")

	for k, tc := range []struct {
		c    *http.Client
		call string
		code int
	}{
		{
			c:    sessionClient,
			call: "/public/with-callback",
			code: http.StatusBadRequest,
		},
		{
			c:    http.DefaultClient,
			call: "/public/with-callback",
			code: http.StatusOK,
		},

		{
			c:    sessionClient,
			call: "/public/without-callback",
			code: http.StatusForbidden,
		},
		{
			c:    http.DefaultClient,
			call: "/public/without-callback",
			code: http.StatusOK,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			res, err := tc.c.Get(ts.URL + tc.call)
			require.NoError(t, err)

			assert.EqualValues(t, tc.code, res.StatusCode)
		})
	}
}
