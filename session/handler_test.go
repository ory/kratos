package session_test

import (
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"

	"github.com/bxcodec/faker"
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

func TestHandler(t *testing.T) {
	t.Run("public", func(t *testing.T) {
		_, reg := internal.NewMemoryRegistry(t)
		r := x.NewRouterPublic()

		var sess Session
		require.NoError(t, faker.FakeData(&sess))
		require.NoError(t, reg.SessionManager().Create(&sess))

		r.GET("/set", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			require.NoError(t, reg.SessionManager().Create(&sess))
			require.NoError(t, reg.SessionManager().SaveToRequest(&sess, w, r))
		})

		h := NewHandler(reg, herodot.NewJSONWriter(nil))
		h.RegisterPublicRoutes(r)
		ts := httptest.NewServer(r)
		defer ts.Close()

		viper.Set(configuration.ViperKeyURLsSelfPublic, ts.URL)

		var err error
		client := ts.Client()
		client.Jar, err = cookiejar.New(&cookiejar.Options{})
		require.NoError(t, err)

		// No cookie yet -> 401
		res, err := ts.Client().Get(ts.URL + "/sessions/me")
		require.NoError(t, err)
		assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode)

		// Set cookie
		res, err = ts.Client().Get(ts.URL + "/set")
		require.NoError(t, err)
		assert.EqualValues(t, http.StatusOK, res.StatusCode)

		// Cookie set -> 200
		res, err = ts.Client().Get(ts.URL + "/sessions/me")
		require.NoError(t, err)
		assert.EqualValues(t, http.StatusOK, res.StatusCode)
	})
}
