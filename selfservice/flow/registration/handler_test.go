package registration_test

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/justinas/nosurf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

func init() {
	internal.RegisterFakes()
}

func TestHandlerRedirectOnAuthenticated(t *testing.T) {
	_, reg := internal.NewRegistryDefault(t)

	router := x.NewRouterPublic()
	reg.RegistrationHandler().RegisterPublicRoutes(router)
	reg.RegistrationStrategies().RegisterPublicRoutes(router)
	ts := httptest.NewServer(router)
	defer ts.Close()

	redirTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("already authenticated"))
	}))
	defer redirTS.Close()

	viper.Set(configuration.ViperKeyURLsDefaultReturnTo, redirTS.URL)
	viper.Set(configuration.ViperKeyURLsSelfPublic, ts.URL)
	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/registration.schema.json")

	t.Run("does redirect to default on authenticated request", func(t *testing.T) {
		body, _ := session.MockMakeAuthenticatedRequest(t, reg, router.Router, x.NewTestHTTPRequest(t, "GET", ts.URL+registration.BrowserRegistrationPath, nil))
		assert.EqualValues(t, "already authenticated", string(body))
	})
}

func TestRegistrationHandler(t *testing.T) {
	_, reg := internal.NewRegistryDefault(t)

	public, admin := func() (*httptest.Server, *httptest.Server) {
		public := x.NewRouterPublic()
		admin := x.NewRouterAdmin()
		reg.RegistrationHandler().RegisterPublicRoutes(public)
		reg.RegistrationHandler().RegisterAdminRoutes(admin)
		reg.RegistrationStrategies().RegisterPublicRoutes(public)
		return httptest.NewServer(x.NewTestCSRFHandler(public, reg)), httptest.NewServer(admin)
	}()
	defer public.Close()
	defer admin.Close()

	redirTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer redirTS.Close()

	newRegistrationTS := func(t *testing.T, upstream string, c *http.Client) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if c == nil {
				c = http.DefaultClient
			}
			_, _ = w.Write(x.EasyGetBody(t, c, upstream+registration.BrowserRegistrationRequestsPath+"?request="+r.URL.Query().Get("request")))
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
		assert.Equal(t, public.URL+registration.BrowserRegistrationPath, gjson.GetBytes(body, "error.details.redirect_to").String(), "%s", body)
	}

	newExpiredRequest := func() *registration.Request {
		return &registration.Request{
			ID:         x.NewUUID(),
			ExpiresAt:  time.Now().Add(-time.Minute),
			IssuedAt:   time.Now().Add(-time.Minute * 2),
			RequestURL: public.URL + registration.BrowserRegistrationPath,
			CSRFToken:  x.FakeCSRFToken,
		}
	}

	errTS := testhelpers.NewErrorTestServer(t, reg)
	defer errTS.Close()

	viper.Set(configuration.ViperKeyURLsSelfPublic, public.URL)
	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/registration.schema.json")
	viper.Set(configuration.ViperKeyURLsError, errTS.URL)

	t.Run("daemon=admin", func(t *testing.T) {
		regTS := newRegistrationTS(t, admin.URL, nil)
		defer regTS.Close()
		viper.Set(configuration.ViperKeyURLsRegistration, regTS.URL)

		t.Run("case=valid", func(t *testing.T) {
			assertRequestPayload(t, x.EasyGetBody(t, public.Client(), public.URL+registration.BrowserRegistrationPath))
		})

		t.Run("case=expired", func(t *testing.T) {
			rr := newExpiredRequest()
			require.NoError(t, reg.RegistrationRequestPersister().CreateRegistrationRequest(context.Background(), rr))
			res, body := x.EasyGet(t, admin.Client(), admin.URL+registration.BrowserRegistrationRequestsPath+"?request="+rr.ID.String())
			assertExpiredPayload(t, res, body)
		})
	})

	t.Run("daemon=public", func(t *testing.T) {
		t.Run("case=with_csrf", func(t *testing.T) {
			j, err := cookiejar.New(nil)
			require.NoError(t, err)
			hc := &http.Client{Jar: j}

			regTS := newRegistrationTS(t, public.URL, hc)
			defer regTS.Close()
			viper.Set(configuration.ViperKeyURLsRegistration, regTS.URL)

			body := x.EasyGetBody(t, hc, public.URL+registration.BrowserRegistrationPath)
			assertRequestPayload(t, body)
		})

		t.Run("case=without_csrf", func(t *testing.T) {
			regTS := newRegistrationTS(t, public.URL,
				// using a different client because it doesn't have access to the cookie jar
				new(http.Client))
			defer regTS.Close()
			viper.Set(configuration.ViperKeyURLsRegistration, regTS.URL)

			body := x.EasyGetBody(t, new(http.Client), public.URL+registration.BrowserRegistrationPath)
			assert.Contains(t, gjson.GetBytes(body, "error").String(), "csrf_token", "%s", body)
		})

		t.Run("case=expired", func(t *testing.T) {
			reg.WithCSRFTokenGenerator(x.FakeCSRFTokenGenerator)
			t.Cleanup(func() {
				reg.WithCSRFTokenGenerator(nosurf.Token)
			})

			j, err := cookiejar.New(nil)
			require.NoError(t, err)
			hc := &http.Client{Jar: j}

			regTS := newRegistrationTS(t, public.URL, hc)
			defer regTS.Close()

			rr := newExpiredRequest()
			require.NoError(t, reg.RegistrationRequestPersister().CreateRegistrationRequest(context.Background(), rr))
			res, body := x.EasyGet(t, admin.Client(), admin.URL+registration.BrowserRegistrationRequestsPath+"?request="+rr.ID.String())
			assertExpiredPayload(t, res, body)
		})
	})
}
