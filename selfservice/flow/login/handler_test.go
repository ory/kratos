package login_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"
	"time"

	"github.com/gobuffalo/httptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

type (
	withTokenGenerator interface {
		WithTokenGenerator(g form.CSRFGenerator)
	}
)

func init() {
	internal.RegisterFakes()
}

func TestHandlerSettingReauth(t *testing.T) {
	_, reg := internal.NewRegistryDefault(t)
	for _, strategy := range reg.LoginStrategies() {
		// We need to know the csrf token
		strategy.(withTokenGenerator).WithTokenGenerator(x.FakeCSRFTokenGenerator)
	}

	router := x.NewRouterPublic()
	admin := x.NewRouterAdmin()
	reg.LoginHandler().RegisterPublicRoutes(router)
	reg.LoginHandler().RegisterAdminRoutes(admin)
	reg.LoginHandler().WithTokenGenerator(x.FakeCSRFTokenGenerator)
	reg.LoginStrategies().RegisterPublicRoutes(router)
	ts := httptest.NewServer(router)
	defer ts.Close()

	loginTS := httptest.NewServer(login.TestRequestHandler(t, reg))

	viper.Set(configuration.ViperKeyURLsSelfPublic, ts.URL)
	viper.Set(configuration.ViperKeyURLsLogin, loginTS.URL)
	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/login.schema.json")

	t.Run("does not set reauth flag on unauthenticated request", func(t *testing.T) {
		c := ts.Client()
		res, err := c.Get(ts.URL + login.BrowserLoginPath)
		require.NoError(t, err)
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)

		assert.Equal(t, false, gjson.GetBytes(body, "is_reauthentication").Bool(), "%s", body)
	})

	t.Run("does set reauth flag on authenticated request", func(t *testing.T) {
		rid := x.NewUUID()
		req := x.NewTestHTTPRequest(t, "GET", ts.URL+login.BrowserLoginPath, nil)
		loginReq := login.NewLoginRequest(time.Minute, x.FakeCSRFToken, req)
		loginReq.ID = rid
		for _, s := range reg.LoginStrategies() {
			require.NoError(t, s.PopulateLoginMethod(req, loginReq))
		}
		require.NoError(t, reg.LoginRequestPersister().CreateLoginRequest(context.TODO(), loginReq), "%+v", loginReq)

		req.URL.RawQuery = url.Values{
			"request": {rid.String()},
		}.Encode()

		body, _ := session.MockMakeAuthenticatedRequest(t, reg, router.Router, req)

		assert.Equal(t, true, gjson.GetBytes(body, "is_reauthentication").Bool(), "%s", body)
	})
}

func TestLoginHandler(t *testing.T) {
	_, reg := internal.NewRegistryDefault(t)

	public, admin := func() (*httptest.Server, *httptest.Server) {
		public := x.NewRouterPublic()
		admin := x.NewRouterAdmin()
		reg.LoginHandler().RegisterPublicRoutes(public)
		reg.LoginHandler().RegisterAdminRoutes(admin)
		reg.LoginStrategies().RegisterPublicRoutes(public)
		return httptest.NewServer(x.NewTestCSRFHandler(public)), httptest.NewServer(admin)
	}()
	defer public.Close()

	redirTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer redirTS.Close()

	newLoginTS := func(t *testing.T, upstream string, c *http.Client) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if c == nil {
				c = http.DefaultClient
			}
			_, err := w.Write(x.EasyGetBody(t, c, upstream+login.BrowserLoginRequestsPath+"?request="+r.URL.Query().Get("request")))
			require.NoError(t, err)
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
		assertRequestPayload(t, x.EasyGetBody(t, public.Client(), public.URL+login.BrowserLoginPath))
	})

	t.Run("daemon=public", func(t *testing.T) {
		j, err := cookiejar.New(nil)
		require.NoError(t, err)
		hc := &http.Client{Jar: j}

		t.Run("case=with_csrf", func(t *testing.T) {
			loginTS := newLoginTS(t, public.URL, hc)
			defer loginTS.Close()
			viper.Set(configuration.ViperKeyURLsLogin, loginTS.URL)

			assertRequestPayload(t, x.EasyGetBody(t, hc, public.URL+login.BrowserLoginPath))
		})

		t.Run("case=without_csrf", func(t *testing.T) {
			loginTS := newLoginTS(t, public.URL,
				// using a different client because it doesn't have access to the cookie jar
				new(http.Client))
			defer loginTS.Close()
			viper.Set(configuration.ViperKeyURLsLogin, loginTS.URL)

			body := x.EasyGetBody(t, hc, public.URL+login.BrowserLoginPath)
			assert.Contains(t, gjson.GetBytes(body, "error").String(), "csrf_token", "%s", body)
		})
	})
}
