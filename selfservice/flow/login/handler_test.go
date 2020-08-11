package login_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gobuffalo/httptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/x"
)

func init() {
	internal.RegisterFakes()
}

func TestHandlerSettingForced(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)

	router := x.NewRouterPublic()
	ts, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin())

	loginTS := testhelpers.NewLoginUIRequestEchoServer(t, reg)

	viper.Set(configuration.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh")
	viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/login.schema.json")

	// assert bool
	ab := func(body []byte, exp bool) {
		r := gjson.GetBytes(body, "forced")
		assert.True(t, r.Exists(), "%s", body)
		assert.Equal(t, exp, r.Bool(), "%s", body)
	}

	// make authenticated request
	mar := func(t *testing.T, extQuery url.Values) (*http.Response, []byte) {
		flowID := x.NewUUID()
		req := x.NewTestHTTPRequest(t, "GET", ts.URL+login.RouteInitBrowserFlow, nil)
		loginFlow := login.NewFlow(time.Minute, x.FakeCSRFToken, req, flow.TypeBrowser)
		loginFlow.ID = flowID
		for _, s := range reg.LoginStrategies() {
			require.NoError(t, s.PopulateLoginMethod(req, loginFlow))
		}
		require.NoError(t, reg.LoginFlowPersister().CreateLoginFlow(context.TODO(), loginFlow), "%+v", loginFlow)

		q := url.Values{"id": {flowID.String()}}
		for key := range extQuery {
			q.Set(key, extQuery.Get(key))
		}
		req.URL.RawQuery = q.Encode()

		body, res := testhelpers.MockMakeAuthenticatedRequest(t, reg, conf, router.Router, req)
		return res, body
	}

	// make unauthenticated request
	mur := func(t *testing.T, query url.Values) (*http.Response, []byte) {
		c := ts.Client()
		res, err := c.Get(ts.URL + login.RouteInitBrowserFlow + "?" + query.Encode())
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

	t.Run("case=does not set forced flag on unauthenticated request with refresh=true", func(t *testing.T) {
		res, body := mur(t, url.Values{"refresh": {"true"}})
		ab(body, true)
		assert.Contains(t, res.Request.URL.String(), loginTS.URL)
	})

	t.Run("case=does not set forced flag on authenticated request without refresh=true", func(t *testing.T) {
		res, _ := mar(t, url.Values{})
		assert.Contains(t, res.Request.URL.String(), "https://www.ory.sh")
	})

	t.Run("case=does not set forced flag on authenticated request with refresh=false", func(t *testing.T) {
		res, _ := mar(t, url.Values{
			"refresh": {"false"},
		})
		assert.Contains(t, res.Request.URL.String(), "https://www.ory.sh")
	})

	t.Run("case=does set forced flag on authenticated request with refresh=true", func(t *testing.T) {
		res, body := mar(t, url.Values{
			"refresh": {"true"},
		})
		ab(body, true)
		assert.Contains(t, res.Request.URL.String(), loginTS.URL)
	})
}

func TestLoginHandler(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	public, admin := testhelpers.NewKratosServerWithCSRF(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)
	_ = testhelpers.NewRedirTS(t, "")

	newLoginTS := func(t *testing.T, upstream string, c *http.Client) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if c == nil {
				c = http.DefaultClient
			}
			_, err := w.Write(x.EasyGetBody(t, c, upstream+login.RouteGetFlow+"?id="+r.URL.Query().Get("flow")))
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
		assert.Equal(t, public.URL+login.RouteInitBrowserFlow, gjson.GetBytes(body, "error.details.redirect_to").String(), "%s", body)
	}

	newExpiredRequest := func() *login.Flow {
		return &login.Flow{
			ID:         x.NewUUID(),
			ExpiresAt:  time.Now().Add(-time.Minute),
			IssuedAt:   time.Now().Add(-time.Minute * 2),
			RequestURL: public.URL + login.RouteInitBrowserFlow,
			CSRFToken:  x.FakeCSRFToken,
		}
	}

	run := func(t *testing.T, endpoint *httptest.Server) {
		loginTS := newLoginTS(t, endpoint.URL, nil)
		defer loginTS.Close()
		viper.Set(configuration.ViperKeySelfServiceLoginUI, loginTS.URL)
		viper.Set(configuration.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword), map[string]interface{}{
			"enabled": true})

		t.Run("case=valid", func(t *testing.T) {
			assertRequestPayload(t, x.EasyGetBody(t, endpoint.Client(), public.URL+login.RouteInitBrowserFlow))
		})

		t.Run("case=expired", func(t *testing.T) {
			lr := newExpiredRequest()
			require.NoError(t, reg.LoginFlowPersister().CreateLoginFlow(context.Background(), lr))
			res, body := x.EasyGet(t, admin.Client(), endpoint.URL+login.RouteGetFlow+"?id="+lr.ID.String())
			assertExpiredPayload(t, res, body)
		})
	}

	t.Run("daemon=admin", func(t *testing.T) {
		run(t, admin)
	})

	t.Run("daemon=public", func(t *testing.T) {
		run(t, public)
	})
}
