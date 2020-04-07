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
	"github.com/justinas/nosurf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

func init() {
	internal.RegisterFakes()
}

func TestHandlerSettingForced(t *testing.T) {
	_, reg := internal.NewRegistryDefault(t)
	reg.WithCSRFTokenGenerator(x.FakeCSRFTokenGenerator)

	router := x.NewRouterPublic()
	admin := x.NewRouterAdmin()
	reg.LoginHandler().RegisterPublicRoutes(router)
	reg.LoginHandler().RegisterAdminRoutes(admin)
	reg.LoginStrategies().RegisterPublicRoutes(router)
	ts := httptest.NewServer(router)
	defer ts.Close()

	loginTS := httptest.NewServer(login.TestRequestHandler(t, reg))

	viper.Set(configuration.ViperKeyURLsSelfPublic, ts.URL)
	viper.Set(configuration.ViperKeyURLsLogin, loginTS.URL)
	viper.Set(configuration.ViperKeyURLsDefaultReturnTo, "https://www.ory.sh")
	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/login.schema.json")

	// assert bool
	ab := func(body []byte, exp bool) {
		r := gjson.GetBytes(body, "forced")
		assert.True(t, r.Exists(), "%s", body)
		assert.Equal(t, exp, r.Bool(), "%s", body)
	}

	// make authenticated request
	mar := func(t *testing.T, extQuery url.Values) (*http.Response, []byte) {
		rid := x.NewUUID()
		req := x.NewTestHTTPRequest(t, "GET", ts.URL+login.BrowserLoginPath, nil)
		loginReq := login.NewLoginRequest(time.Minute, x.FakeCSRFToken, req)
		loginReq.ID = rid
		for _, s := range reg.LoginStrategies() {
			require.NoError(t, s.PopulateLoginMethod(req, loginReq))
		}
		require.NoError(t, reg.LoginRequestPersister().CreateLoginRequest(context.TODO(), loginReq), "%+v", loginReq)

		q := url.Values{"request": {rid.String()}}
		for key := range extQuery {
			q.Set(key, extQuery.Get(key))
		}
		req.URL.RawQuery = q.Encode()

		body, res := session.MockMakeAuthenticatedRequest(t, reg, router.Router, req)
		return res, body
	}

	// make unauthenticated request
	mur := func(t *testing.T, query url.Values) (*http.Response, []byte) {
		c := ts.Client()
		res, err := c.Get(ts.URL + login.BrowserLoginPath + "?" + query.Encode())
		require.NoError(t, err)
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		return res, body
	}

	t.Run("case=does not set forced flag on unauthenticated request", func(t *testing.T) {
		res, body := mur(t, url.Values{})
		ab(body, false)
		assert.Contains(t, res.Request.URL.String(), loginTS.URL)
	})

	t.Run("case=does not set forced flag on unauthenticated request with prompt=login", func(t *testing.T) {
		res, body := mur(t, url.Values{
			"prompt": {"login"},
		})
		ab(body, false)
		assert.Contains(t, res.Request.URL.String(), loginTS.URL)
	})

	t.Run("case=does not set forced flag on authenticated request without prompt=login", func(t *testing.T) {
		res, _ := mar(t, url.Values{})
		assert.Contains(t, res.Request.URL.String(), "https://www.ory.sh")
	})

	t.Run("case=does not set forced flag on authenticated request with prompt=false", func(t *testing.T) {
		res, _ := mar(t, url.Values{
			"prompt": {"false"},
		})
		assert.Contains(t, res.Request.URL.String(), "https://www.ory.sh")
	})

	t.Run("case=does set forced flag on authenticated request with prompt=login", func(t *testing.T) {
		res, body := mar(t, url.Values{
			"prompt": {"login"},
		})
		ab(body, true)
		assert.Contains(t, res.Request.URL.String(), loginTS.URL)
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
		return httptest.NewServer(x.NewTestCSRFHandler(public, reg)), httptest.NewServer(admin)
	}()
	defer public.Close()
	defer admin.Close()

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

	assertExpiredPayload := func(t *testing.T, res *http.Response, body []byte) {
		assert.EqualValues(t, http.StatusGone, res.StatusCode)
		assert.Equal(t, public.URL+login.BrowserLoginPath, gjson.GetBytes(body, "error.details.redirect_to").String(), "%s", body)
	}

	newExpiredRequest := func() *login.Request {
		return &login.Request{
			ID:         x.NewUUID(),
			ExpiresAt:  time.Now().Add(-time.Minute),
			IssuedAt:   time.Now().Add(-time.Minute * 2),
			RequestURL: public.URL + login.BrowserLoginPath,
			CSRFToken:  x.FakeCSRFToken,
		}
	}

	errTS := testhelpers.NewErrorTestServer(t, reg)
	defer errTS.Close()

	viper.Set(configuration.ViperKeyURLsSelfPublic, public.URL)
	viper.Set(configuration.ViperKeyURLsError, errTS.URL)

	t.Run("daemon=admin", func(t *testing.T) {
		loginTS := newLoginTS(t, admin.URL, nil)
		defer loginTS.Close()
		viper.Set(configuration.ViperKeyURLsLogin, loginTS.URL)

		t.Run("case=valid", func(t *testing.T) {
			assertRequestPayload(t, x.EasyGetBody(t, admin.Client(), public.URL+login.BrowserLoginPath))
		})

		t.Run("case=expired", func(t *testing.T) {
			lr := newExpiredRequest()
			require.NoError(t, reg.LoginRequestPersister().CreateLoginRequest(context.Background(), lr))
			res, body := x.EasyGet(t, admin.Client(), admin.URL+login.BrowserLoginRequestsPath+"?request="+lr.ID.String())
			assertExpiredPayload(t, res, body)
		})
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

		t.Run("case=expired", func(t *testing.T) {
			reg.WithCSRFTokenGenerator(x.FakeCSRFTokenGenerator)
			t.Cleanup(func() {
				reg.WithCSRFTokenGenerator(nosurf.Token)
			})

			loginTS := newLoginTS(t, public.URL, hc)
			defer loginTS.Close()

			lr := newExpiredRequest()
			require.NoError(t, reg.LoginRequestPersister().CreateLoginRequest(context.Background(), lr))
			res, body := x.EasyGet(t, admin.Client(), admin.URL+login.BrowserLoginRequestsPath+"?request="+lr.ID.String())
			assertExpiredPayload(t, res, body)
		})
	})
}
