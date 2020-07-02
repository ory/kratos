package recovery_test

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
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/x"
)

func init() {
	internal.RegisterFakes()
}

func TestHandlerRedirectOnAuthenticated(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.RecoveryFlowEnable(true)

	testhelpers.NewRecoveryUITestServer(t)
	redirTS := testhelpers.NewRedirTS(t, "already authenticated")
	viper.Set(configuration.ViperKeySelfServiceLoginUI, redirTS.URL)

	router := x.NewRouterPublic()
	testhelpers.NewErrorTestServer(t, reg)
	public, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin())

	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/identity.schema.json")

	t.Run("does redirect to default on authenticated request", func(t *testing.T) {
		body, _ := testhelpers.MockMakeAuthenticatedRequest(t, reg, conf, router.Router, x.NewTestHTTPRequest(t, "GET", public.URL+recovery.PublicRecoveryInitPath, nil))
		assert.EqualValues(t, "already authenticated", string(body))
	})
}

func TestRecoveryHandler(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.RecoveryFlowEnable(true)

	testhelpers.NewRedirTS(t, "")
	testhelpers.NewLoginUIRequestEchoServer(t, reg)
	testhelpers.NewErrorTestServer(t, reg)

	public, admin := testhelpers.NewKratosServerWithCSRF(t, reg)
	newRecoveryTS := func(t *testing.T, upstream string, c *http.Client) *httptest.Server {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if c == nil {
				c = http.DefaultClient
			}
			_, _ = w.Write(x.EasyGetBody(t, c, upstream+recovery.PublicRecoveryRequestPath+"?request="+r.URL.Query().Get("request")))
		}))
		viper.Set(configuration.ViperKeySelfServiceRecoveryUI, ts.URL)
		t.Cleanup(ts.Close)
		return ts
	}

	assertRequestPayload := func(t *testing.T, body []byte) {
		// assert.Equal(t, "password", gjson.GetBytes(body, "methods.password.method").String(), "%s", body)
		// assert.NotEmpty(t, gjson.GetBytes(body, "methods.password.config.fields.#(name==csrf_token).value").String(), "%s", body)
		assert.NotEmpty(t, gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Empty(t, gjson.GetBytes(body, "headers").Value(), "%s", body)
		assert.Equal(t, "choose_method", gjson.GetBytes(body, "state").Value(), "%s", body)
		// assert.Contains(t, gjson.GetBytes(body, "methods.password.config.action").String(), gjson.GetBytes(body, "id").String(), "%s", body)
		// assert.Contains(t, gjson.GetBytes(body, "methods.password.config.action").String(), public.URL, "%s", body)
	}

	assertExpiredPayload := func(t *testing.T, res *http.Response, body []byte) {
		assert.EqualValues(t, http.StatusGone, res.StatusCode)
		assert.Equal(t, public.URL+recovery.PublicRecoveryInitPath, gjson.GetBytes(body, "error.details.redirect_to").String(), "%s", body)
	}

	newExpiredRequest := func() *recovery.Request {
		return &recovery.Request{
			ID:         x.NewUUID(),
			ExpiresAt:  time.Now().Add(-time.Minute),
			IssuedAt:   time.Now().Add(-time.Minute * 2),
			RequestURL: public.URL + recovery.PublicRecoveryInitPath,
			CSRFToken:  x.FakeCSRFToken,
		}
	}

	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/recovery.schema.json")

	t.Run("daemon=admin", func(t *testing.T) {
		regTS := newRecoveryTS(t, admin.URL, nil)
		defer regTS.Close()
		viper.Set(configuration.ViperKeySelfServiceRecoveryUI, regTS.URL)

		t.Run("case=valid", func(t *testing.T) {
			assertRequestPayload(t, x.EasyGetBody(t, public.Client(), public.URL+recovery.PublicRecoveryInitPath))
		})

		t.Run("case=expired", func(t *testing.T) {
			rr := newExpiredRequest()
			require.NoError(t, reg.RecoveryRequestPersister().CreateRecoveryRequest(context.Background(), rr))
			res, body := x.EasyGet(t, admin.Client(), admin.URL+recovery.PublicRecoveryRequestPath+"?request="+rr.ID.String())
			assertExpiredPayload(t, res, body)
		})
	})

	t.Run("daemon=public", func(t *testing.T) {
		t.Run("case=with_csrf", func(t *testing.T) {
			j, err := cookiejar.New(nil)
			require.NoError(t, err)
			hc := &http.Client{Jar: j}

			newRecoveryTS(t, public.URL, hc)

			body := x.EasyGetBody(t, hc, public.URL+recovery.PublicRecoveryInitPath)
			assertRequestPayload(t, body)
		})

		t.Run("case=without_csrf", func(t *testing.T) {
			newRecoveryTS(t, public.URL,
				// using a different client because it doesn't have access to the cookie jar
				new(http.Client))

			body := x.EasyGetBody(t, new(http.Client), public.URL+recovery.PublicRecoveryInitPath)
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

			regTS := newRecoveryTS(t, public.URL, hc)
			defer regTS.Close()

			rr := newExpiredRequest()
			require.NoError(t, reg.RecoveryRequestPersister().CreateRecoveryRequest(context.Background(), rr))
			res, body := x.EasyGet(t, admin.Client(), admin.URL+recovery.PublicRecoveryRequestPath+"?request="+rr.ID.String())
			assertExpiredPayload(t, res, body)
		})
	})
}
