package login_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
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
	admin := x.NewRouterAdmin()
	reg.LoginHandler().RegisterPublicRoutes(router, admin)
	reg.LoginStrategies().RegisterPublicRoutes(router)
	ts := httptest.NewServer(router)
	defer ts.Close()

	redirTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("already authenticated"))
	}))
	defer redirTS.Close()

	viper.Set(configuration.ViperKeyURLsDefaultReturnTo, redirTS.URL)
	viper.Set(configuration.ViperKeyURLsSelfPublic, ts.URL)
	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/login.schema.json")

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

	public, admin := func() (*httptest.Server, *httptest.Server) {
		public := x.NewRouterPublic()
		admin := x.NewRouterAdmin()
		reg.LoginHandler().RegisterPublicRoutes(public, admin)
		reg.LoginStrategies().RegisterPublicRoutes(public)
		return httptest.NewServer(nosurf.New(public)), httptest.NewServer(admin)
	}()
	defer public.Close()

	redirTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer redirTS.Close()

	easyGet := func(t *testing.T, c *http.Client, url string) []byte {
		res, err := c.Get(url)
		require.NoError(t, err)
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		return body
	}

	newLoginTS := func(t *testing.T, upstream string, c *http.Client) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if c == nil {
				c = http.DefaultClient
			}
			_, _ = w.Write(easyGet(t, c, upstream+login.BrowserLoginRequestsPath+"?request="+r.URL.Query().Get("request")))
		}))
	}

	assertRequestPayload := func(t *testing.T, body []byte) {
		assert.Equal(t, "password", gjson.GetBytes(body, "methods.password.method").String(), "%s", body)
		assert.NotEmpty(t, gjson.GetBytes(body, "methods.password.config.fields.#(name==csrf_token).value").String(), "%s", body)
		assert.NotEmpty(t, gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Empty(t, gjson.GetBytes(body, "headers").Value(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "methods.password.config.action").String(), gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "methods.password.config.action").String(), public.URL, "%s", body)
	}

	errTS := errorx.NewErrorTestServer(t, reg)
	defer errTS.Close()

	viper.Set(configuration.ViperKeyURLsSelfPublic, public.URL)
	viper.Set(configuration.ViperKeyURLsError, errTS.URL)

	t.Run("daemon=admin", func(t *testing.T) {
		loginTS := newLoginTS(t, admin.URL, nil)
		defer loginTS.Close()

		viper.Set(configuration.ViperKeyURLsLogin, loginTS.URL)
		assertRequestPayload(t, easyGet(t, public.Client(), public.URL+login.BrowserLoginPath))
	})

	t.Run("daemon=public", func(t *testing.T) {
		j, err := cookiejar.New(nil)
		require.NoError(t, err)
		hc := &http.Client{Jar: j}

		t.Run("case=with_csrf", func(t *testing.T) {
			loginTS := newLoginTS(t, public.URL, hc)
			defer loginTS.Close()
			viper.Set(configuration.ViperKeyURLsLogin, loginTS.URL)

			assertRequestPayload(t, easyGet(t, hc, public.URL+login.BrowserLoginPath))
		})

		t.Run("case=without_csrf", func(t *testing.T) {
			loginTS := newLoginTS(t, public.URL,
				// using a different client because it doesn't have access to the cookie jar
				new(http.Client))
			defer loginTS.Close()
			viper.Set(configuration.ViperKeyURLsLogin, loginTS.URL)

			body := easyGet(t, hc, public.URL+login.BrowserLoginPath)
			assert.Contains(t, gjson.GetBytes(body, "error").String(), "csrf_token", "%s", body)
		})
	})
}
