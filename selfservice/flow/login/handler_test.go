package login_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/gobuffalo/httptest"
	"github.com/justinas/nosurf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/kratos/selfservice/strategy/password"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

func init() {
	internal.RegisterFakes()
}

func TestEnsureSessionRedirect(t *testing.T) {
	_, reg := internal.NewRegistryDefault(t)

	router := x.NewRouterPublic()
	reg.LoginHandler().RegisterPublicRoutes(router)
	reg.LoginStrategies().RegisterPublicRoutes(router)
	ts := httptest.NewServer(router)
	defer ts.Close()

	redirTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("already authenticated"))
	}))
	defer redirTS.Close()

	viper.Set(configuration.ViperKeyURLsDefaultReturnTo, redirTS.URL)
	viper.Set(configuration.ViperKeyURLsSelfPublic, ts.URL)

	_, _ = reg.SchemaPersister().RegisterDefaultSchema("file://./stub/login.schema.json")

	for k, tc := range [][]string{
		{"GET", login.BrowserLoginPath},

		{"POST", password.LoginPath},

		// it is ok that these contain the parameters as raw strings as we are only interested in checking if the middleware is working
		{"POST", oidc.AuthPath},
		{"GET", oidc.AuthPath},
		{"GET", oidc.CallbackPath},
	} {
		t.Run(fmt.Sprintf("case=%d/method=%s/path=%s", k, tc[0], tc[1]), func(t *testing.T) {
			body, _ := session.MockMakeAuthenticatedRequest(t, reg, router.Router, x.NewTestHTTPRequest(t, tc[0], ts.URL+tc[1], nil))
			assert.EqualValues(t, "already authenticated", string(body))
		})
	}
}

func TestLoginHandler(t *testing.T) {
	_, reg := internal.NewRegistryDefault(t)

	kratos := func() *httptest.Server {
		router := x.NewRouterPublic()
		reg.LoginHandler().RegisterPublicRoutes(router)
		reg.LoginStrategies().RegisterPublicRoutes(router)
		return httptest.NewServer(nosurf.New(router))
	}()
	defer kratos.Close()

	redirTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer redirTS.Close()

	loginTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res, err := http.Get(kratos.URL + login.BrowserLoginRequestsPath + "?request=" + r.URL.Query().Get("request"))
		require.NoError(t, err)
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		_, _ = w.Write(body)
	}))
	defer loginTS.Close()

	errTS := errorx.NewErrorTestServer(t, reg)
	defer errTS.Close()

	viper.Set(configuration.ViperKeyURLsLogin, loginTS.URL)
	viper.Set(configuration.ViperKeyURLsSelfPublic, kratos.URL)
	viper.Set(configuration.ViperKeyURLsError, errTS.URL)

	t.Run("", func(t *testing.T) {
		res, err := kratos.Client().Get(kratos.URL + login.BrowserLoginPath)
		require.NoError(t, err)

		defer res.Body.Close()
		require.Equal(t, http.StatusOK, res.StatusCode)

		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)

		assert.Equal(t, "password", gjson.GetBytes(body, "methods.password.method").String(), "%s", body)
		assert.NotEmpty(t, gjson.GetBytes(body, "methods.password.config.fields.csrf_token.value").String(), "%s", body)
		assert.NotEmpty(t, gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Empty(t, gjson.GetBytes(body, "headers").Value(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "methods.password.config.action").String(), gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "methods.password.config.action").String(), kratos.URL, "%s", body)
	})
}
