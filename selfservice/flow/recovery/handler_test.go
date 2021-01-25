package recovery_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/assertx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/x"
)

func init() {
	internal.RegisterFakes()
}

func TestHandlerRedirectOnAuthenticated(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(config.ViperKeySelfServiceRecoveryEnabled, true)

	router := x.NewRouterPublic()
	ts, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin())

	redirTS := testhelpers.NewRedirTS(t, "already authenticated", conf)
	conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/identity.schema.json")

	t.Run("does redirect to default on authenticated request", func(t *testing.T) {
		body, res := testhelpers.MockMakeAuthenticatedRequest(t, reg, conf, router.Router, x.NewTestHTTPRequest(t, "GET", ts.URL+recovery.RouteInitBrowserFlow, nil))
		assert.Contains(t, res.Request.URL.String(), redirTS.URL, "%+v", res)
		assert.EqualValues(t, "already authenticated", string(body))
	})

	t.Run("does redirect to default on authenticated request", func(t *testing.T) {
		body, res := testhelpers.MockMakeAuthenticatedRequest(t, reg, conf, router.Router, x.NewTestHTTPRequest(t, "GET", ts.URL+recovery.RouteInitAPIFlow, nil))
		assert.Contains(t, res.Request.URL.String(), recovery.RouteInitAPIFlow)
		assertx.EqualAsJSON(t, recovery.ErrAlreadyLoggedIn, json.RawMessage(gjson.GetBytes(body, "error").Raw))
	})
}

func TestInitFlow(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(config.ViperKeySelfServiceRecoveryEnabled, true)
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+recovery.StrategyRecoveryLinkName,
		map[string]interface{}{"enabled": true})

	router := x.NewRouterPublic()
	publicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin())
	recoveryTS := testhelpers.NewRecoveryUIFlowEchoServer(t, reg)

	conf.MustSet(config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh")
	conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/identity.schema.json")

	assertion := func(body []byte, isForced, isApi bool) {
		if isApi {
			assert.Equal(t, "api", gjson.GetBytes(body, "type").String())
		} else {
			assert.Equal(t, "browser", gjson.GetBytes(body, "type").String())
		}
	}

	initAuthenticatedFlow := func(t *testing.T, isAPI bool) (*http.Response, []byte) {
		route := recovery.RouteInitBrowserFlow
		if isAPI {
			route = recovery.RouteInitAPIFlow
		}
		req := x.NewTestHTTPRequest(t, "GET", publicTS.URL+route, nil)
		body, res := testhelpers.MockMakeAuthenticatedRequest(t, reg, conf, router.Router, req)
		if isAPI {
			assert.Len(t, res.Header.Get("Set-Cookie"), 0)
		}
		return res, body
	}

	initFlow := func(t *testing.T, isAPI bool) (*http.Response, []byte) {
		route := recovery.RouteInitBrowserFlow
		if isAPI {
			route = recovery.RouteInitAPIFlow
		}
		c := publicTS.Client()
		res, err := c.Get(publicTS.URL + route)
		require.NoError(t, err)
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		return res, body
	}

	t.Run("flow=api", func(t *testing.T) {
		t.Run("case=creates a new flow on unauthenticated request", func(t *testing.T) {
			res, body := initFlow(t, true)
			assert.Contains(t, res.Request.URL.String(), recovery.RouteInitAPIFlow)
			assertion(body, false, true)
		})

		t.Run("case=fails on authenticated request", func(t *testing.T) {
			res, body := initAuthenticatedFlow(t, true)
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
			assertx.EqualAsJSON(t, recovery.ErrAlreadyLoggedIn, json.RawMessage(gjson.GetBytes(body, "error").Raw), "%s", body)
		})
	})

	t.Run("flow=browser", func(t *testing.T) {
		t.Run("case=does not set forced flag on unauthenticated request", func(t *testing.T) {
			res, body := initFlow(t, false)
			assertion(body, false, false)
			assert.Contains(t, res.Request.URL.String(), recoveryTS.URL)
		})
		t.Run("case=does not set forced flag on authenticated request with refresh=false", func(t *testing.T) {
			res, _ := initAuthenticatedFlow(t, false)
			assert.Contains(t, res.Request.URL.String(), "https://www.ory.sh")
		})
	})
}

func TestGetFlow(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(config.ViperKeySelfServiceRecoveryEnabled, true)
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+recovery.StrategyRecoveryLinkName,
		map[string]interface{}{"enabled": true})
	conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/identity.schema.json")

	public, admin := testhelpers.NewKratosServerWithCSRF(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)
	_ = testhelpers.NewRedirTS(t, "", conf)

	newRecoveryTS := func(t *testing.T, upstream string, c *http.Client) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if c == nil {
				c = http.DefaultClient
			}
			_, err := w.Write(x.EasyGetBody(t, c, upstream+recovery.RouteGetFlow+"?id="+r.URL.Query().Get("flow")))
			require.NoError(t, err)
		}))
	}

	assertFlowPayload := func(t *testing.T, body []byte) {
		assert.Equal(t, "link", gjson.GetBytes(body, "methods.link.method").String(), "%s", body)
		assert.NotEmpty(t, gjson.GetBytes(body, "methods.link.config.nodes.#(attributes.name==csrf_token).attributes.value").String(), "%s", body)
		assert.NotEmpty(t, gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Empty(t, gjson.GetBytes(body, "headers").Value(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "methods.link.config.action").String(), gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "methods.link.config.action").String(), public.URL, "%s", body)
	}

	assertExpiredPayload := func(t *testing.T, res *http.Response, body []byte) {
		assert.EqualValues(t, http.StatusGone, res.StatusCode)
		assert.Equal(t, public.URL+recovery.RouteInitBrowserFlow, gjson.GetBytes(body, "error.details.redirect_to").String(), "%s", body)
	}

	newExpiredFlow := func() *recovery.Flow {
		return &recovery.Flow{
			ID:         x.NewUUID(),
			ExpiresAt:  time.Now().Add(-time.Minute),
			IssuedAt:   time.Now().Add(-time.Minute * 2),
			RequestURL: public.URL + recovery.RouteInitBrowserFlow,
			CSRFToken:  x.FakeCSRFToken,
			Type:       flow.TypeBrowser,
		}
	}

	run := func(t *testing.T, endpoint *httptest.Server) {
		recoveryTS := newRecoveryTS(t, endpoint.URL, nil)
		defer recoveryTS.Close()
		conf.MustSet(config.ViperKeySelfServiceRecoveryUI, recoveryTS.URL)
		conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword),
			map[string]interface{}{"enabled": true})

		t.Run("case=valid", func(t *testing.T) {
			assertFlowPayload(t, x.EasyGetBody(t, endpoint.Client(), public.URL+recovery.RouteInitBrowserFlow))
		})

		t.Run("case=expired", func(t *testing.T) {
			lr := newExpiredFlow()
			require.NoError(t, reg.RecoveryFlowPersister().CreateRecoveryFlow(context.Background(), lr))
			res, body := x.EasyGet(t, admin.Client(), endpoint.URL+recovery.RouteGetFlow+"?id="+lr.ID.String())
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
