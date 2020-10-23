package logout_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/gobuffalo/httptest"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/logout"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"github.com/ory/nosurf"
	"github.com/ory/viper"
	"github.com/ory/x/logrusx"
)

func TestLogoutHandler(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	handler := reg.LogoutHandler()

	viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")
	viper.Set(configuration.ViperKeyPublicBaseURL, "http://example.com")

	router := x.NewRouterPublic()
	handler.RegisterPublicRoutes(router)
	reg.WithCSRFHandler(x.NewCSRFHandler(router, reg.Writer(), logrusx.New("", ""), "/", "", false))
	ts := httptest.NewServer(reg.CSRFHandler())
	defer ts.Close()

	var sess session.Session
	sess.ID = x.NewUUID()
	sess.Identity = new(identity.Identity)
	require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), sess.Identity))
	require.NoError(t, reg.SessionPersister().CreateSession(context.Background(), &sess))

	router.GET("/set", testhelpers.MockSetSession(t, reg, conf))

	router.GET("/csrf", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		_, _ = w.Write([]byte(nosurf.Token(r)))
	})

	redirTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer redirTS.Close()

	viper.Set(configuration.ViperKeySelfServiceLogoutBrowserDefaultReturnTo, redirTS.URL)
	viper.Set(configuration.ViperKeyPublicBaseURL, ts.URL)

	client := testhelpers.NewClientWithCookies(t)

	t.Run("case=set initial session", func(t *testing.T) {
		testhelpers.MockHydrateCookieClient(t, client, ts.URL+"/set")
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
		res, err := client.Get(ts.URL + logout.RouteBrowser)
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
