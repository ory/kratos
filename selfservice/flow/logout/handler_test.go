package logout_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/gobuffalo/httptest"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/nosurf"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/flow/logout"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

func TestLogoutHandler(t *testing.T) {
	_, reg := internal.NewMemoryRegistry(t)
	handler := reg.LogoutHandler()

	router := x.NewRouterPublic()
	handler.RegisterPublicRoutes(router)
	reg.WithCSRFHandler(x.NewCSRFHandler(router, reg.Writer(), logrus.New(), "/", "", false))
	ts := httptest.NewServer(reg.CSRFHandler())
	defer ts.Close()

	var sess session.Session
	sess.SID = uuid.New().String()
	sess.Identity = new(identity.Identity)
	require.NoError(t, reg.SessionManager().Create(context.Background(), &sess))

	router.GET("/set", session.MockSetSession(t, reg))

	router.GET("/csrf", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		_, _ = w.Write([]byte(nosurf.Token(r)))
	})

	redirTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer redirTS.Close()

	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/registration.schema.json")
	viper.Set(configuration.ViperKeySelfServiceLogoutRedirectURL, redirTS.URL)
	viper.Set(configuration.ViperKeyURLsSelfPublic, ts.URL)

	client := session.MockCookieClient(t)

	t.Run("case=set initial session", func(t *testing.T) {
		session.MockHydrateCookieClient(t, client, ts.URL+"/set")
	})

	var token string
	t.Run("case=get csrf token", func(t *testing.T) {
		res, err := ts.Client().Get(ts.URL + "/csrf")
		require.NoError(t, err)
		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		token = string(body)
		require.NotEmpty(t, token)
	})

	t.Run("case=log out", func(t *testing.T) {
		res, err := client.Get(ts.URL + logout.BrowserLogoutPath)
		require.NoError(t, err)

		var found bool
		for _, c := range res.Cookies() {
			if c.Name == session.DefaultSessionCookieName {
				found = true
			}
		}
		require.False(t, found)
	})

	t.Run("case=csrf token should be reset", func(t *testing.T) {
		res, err := ts.Client().Get(ts.URL + "/csrf")
		require.NoError(t, err)
		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		require.NotEmpty(t, body)
		assert.NotEqual(t, token, string(body))
	})
}
