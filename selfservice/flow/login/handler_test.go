package login_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/ory/kratos/corpx"

	"github.com/gobuffalo/httptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/assertx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/x"
)

func init() {
	corpx.RegisterFakes()
}

func TestInitFlow(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	router := x.NewRouterPublic()
	ts, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin())
	loginTS := testhelpers.NewLoginUIFlowEchoServer(t, reg)

	conf.MustSet(config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh")
	conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/login.schema.json")

	assertion := func(body []byte, isForced, isApi bool) {
		r := gjson.GetBytes(body, "forced")
		assert.True(t, r.Exists(), "%s", body)
		assert.Equal(t, isForced, r.Bool(), "%s", body)
		if isApi {
			assert.Equal(t, "api", gjson.GetBytes(body, "type").String())
		} else {
			assert.Equal(t, "browser", gjson.GetBytes(body, "type").String())
		}
	}

	initAuthenticatedFlow := func(t *testing.T, extQuery url.Values, isAPI bool) (*http.Response, []byte) {
		route := login.RouteInitBrowserFlow
		if isAPI {
			route = login.RouteInitAPIFlow
		}
		req := x.NewTestHTTPRequest(t, "GET", ts.URL+route, nil)
		req.URL.RawQuery = extQuery.Encode()
		body, res := testhelpers.MockMakeAuthenticatedRequest(t, reg, conf, router.Router, req)
		if isAPI {
			assert.Len(t, res.Header.Get("Set-Cookie"), 0)
		}
		return res, body
	}

	initFlow := func(t *testing.T, query url.Values, isAPI bool) (*http.Response, []byte) {
		route := login.RouteInitBrowserFlow
		if isAPI {
			route = login.RouteInitAPIFlow
		}
		c := ts.Client()
		res, err := c.Get(ts.URL + route + "?" + query.Encode())
		require.NoError(t, err)
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		return res, body
	}

	t.Run("flow=api", func(t *testing.T) {
		t.Run("case=does not set forced flag on unauthenticated request", func(t *testing.T) {
			res, body := initFlow(t, url.Values{}, true)
			assert.Contains(t, res.Request.URL.String(), login.RouteInitAPIFlow)
			assertion(body, false, true)
		})

		t.Run("case=does not set forced flag on unauthenticated request with refresh=true", func(t *testing.T) {
			res, body := initFlow(t, url.Values{"refresh": {"true"}}, true)
			assert.Contains(t, res.Request.URL.String(), login.RouteInitAPIFlow)
			assertion(body, true, true)
		})

		t.Run("case=does not set forced flag on authenticated request without refresh=true", func(t *testing.T) {
			res, body := initAuthenticatedFlow(t, url.Values{}, true)
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
			assertx.EqualAsJSON(t, login.ErrAlreadyLoggedIn, json.RawMessage(gjson.GetBytes(body, "error").Raw), "%s", body)
		})

		t.Run("case=does not set forced flag on authenticated request with refresh=false", func(t *testing.T) {
			res, body := initAuthenticatedFlow(t, url.Values{"refresh": {"false"}}, true)
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
			assertx.EqualAsJSON(t, login.ErrAlreadyLoggedIn, json.RawMessage(gjson.GetBytes(body, "error").Raw), "%s", body)
		})

		t.Run("case=does set forced flag on authenticated request with refresh=true", func(t *testing.T) {
			res, body := initAuthenticatedFlow(t, url.Values{"refresh": {"true"}}, true)
			assert.Contains(t, res.Request.URL.String(), login.RouteInitAPIFlow)
			assertion(body, true, true)
		})
	})

	t.Run("flow=browser", func(t *testing.T) {
		t.Run("case=does not set forced flag on unauthenticated request", func(t *testing.T) {
			res, body := initFlow(t, url.Values{}, false)
			assertion(body, false, false)
			assert.Contains(t, res.Request.URL.String(), loginTS.URL)
		})

		t.Run("case=does not set forced flag on unauthenticated request with refresh=true", func(t *testing.T) {
			res, body := initFlow(t, url.Values{"refresh": {"true"}}, false)
			assertion(body, true, false)
			assert.Contains(t, res.Request.URL.String(), loginTS.URL)
		})

		t.Run("case=does not set forced flag on authenticated request without refresh=true", func(t *testing.T) {
			res, _ := initAuthenticatedFlow(t, url.Values{}, false)
			assert.Contains(t, res.Request.URL.String(), "https://www.ory.sh")
		})

		t.Run("case=does not set forced flag on authenticated request with refresh=false", func(t *testing.T) {
			res, _ := initAuthenticatedFlow(t, url.Values{"refresh": {"false"}}, false)
			assert.Contains(t, res.Request.URL.String(), "https://www.ory.sh")
		})

		t.Run("case=does set forced flag on authenticated request with refresh=true", func(t *testing.T) {
			res, body := initAuthenticatedFlow(t, url.Values{"refresh": {"true"}}, false)
			assertion(body, true, false)
			assert.Contains(t, res.Request.URL.String(), loginTS.URL)
		})
	})
}

func TestGetFlow(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	public, admin := testhelpers.NewKratosServerWithCSRF(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)
	_ = testhelpers.NewRedirTS(t, "", conf)

	newLoginTS := func(t *testing.T, upstream string, c *http.Client) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if c == nil {
				c = http.DefaultClient
			}
			_, err := w.Write(x.EasyGetBody(t, c, upstream+login.RouteGetFlow+"?id="+r.URL.Query().Get("flow")))
			require.NoError(t, err)
		}))
	}

	assertFlowPayload := func(t *testing.T, body []byte) {
		assert.NotEmpty(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String(), "%s", body)
		assert.NotEmpty(t, gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Empty(t, gjson.GetBytes(body, "headers").Value(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "ui.action").String(), gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "ui.action").String(), public.URL, "%s", body)
	}

	assertExpiredPayload := func(t *testing.T, res *http.Response, body []byte) {
		assert.EqualValues(t, http.StatusGone, res.StatusCode)
		assert.Equal(t, public.URL+login.RouteInitBrowserFlow, gjson.GetBytes(body, "error.details.redirect_to").String(), "%s", body)
	}

	newExpiredFlow := func() *login.Flow {
		return &login.Flow{
			ID:         x.NewUUID(),
			ExpiresAt:  time.Now().Add(-time.Minute),
			IssuedAt:   time.Now().Add(-time.Minute * 2),
			RequestURL: public.URL + login.RouteInitBrowserFlow,
			CSRFToken:  x.FakeCSRFToken,
			Type:       flow.TypeBrowser,
		}
	}

	run := func(t *testing.T, endpoint *httptest.Server) {
		loginTS := newLoginTS(t, endpoint.URL, nil)
		defer loginTS.Close()
		conf.MustSet(config.ViperKeySelfServiceLoginUI, loginTS.URL)
		conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword), map[string]interface{}{
			"enabled": true})

		t.Run("case=valid", func(t *testing.T) {
			assertFlowPayload(t, x.EasyGetBody(t, endpoint.Client(), public.URL+login.RouteInitBrowserFlow))
		})

		t.Run("case=expired", func(t *testing.T) {
			lr := newExpiredFlow()
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
