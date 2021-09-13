package login_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/ory/kratos/text"

	"github.com/gobuffalo/httptest"
	"github.com/gofrs/uuid"

	"github.com/ory/kratos/corpx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/assertx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
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

	initFlowWithAccept := func(t *testing.T, query url.Values, isAPI bool, accept string) (*http.Response, []byte) {
		route := login.RouteInitBrowserFlow
		if isAPI {
			route = login.RouteInitAPIFlow
		}
		c := ts.Client()
		req, err := http.NewRequest("GET", ts.URL+route+"?"+query.Encode(), nil)
		require.NoError(t, err)
		if accept != "" {
			req.Header.Set("Accept", accept)
		}

		res, err := c.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		return res, body
	}

	initFlow := func(t *testing.T, query url.Values, isAPI bool) (*http.Response, []byte) {
		return initFlowWithAccept(t, query, isAPI, "")
	}

	initSPAFlow := func(t *testing.T, query url.Values) (*http.Response, []byte) {
		return initFlowWithAccept(t, query, false, "application/json")
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

		t.Run("case=check info message on authenticated request with refresh=true", func(t *testing.T) {
			res, body := initAuthenticatedFlow(t, url.Values{"refresh": {"true"}}, true)
			assert.Contains(t, res.Request.URL.String(), login.RouteInitAPIFlow)
			assertion(body, true, true)
			assert.Equal(t, gjson.GetBytes(body, "ui.messages.0.text").String(), text.NewInfoLoginReAuth().Text)
		})
	})

	t.Run("flow=browser", func(t *testing.T) {
		t.Run("case=does not set forced flag on unauthenticated request", func(t *testing.T) {
			res, body := initFlow(t, url.Values{}, false)
			assertion(body, false, false)
			assert.Contains(t, res.Request.URL.String(), loginTS.URL)
		})

		t.Run("case=makes request with JSON", func(t *testing.T) {
			res, body := initSPAFlow(t, url.Values{})
			assertion(body, false, false)
			assert.NotContains(t, res.Request.URL.String(), loginTS.URL)
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
	public, _ := testhelpers.NewKratosServerWithCSRF(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)
	_ = testhelpers.NewRedirTS(t, "", conf)

	setupLoginUI := func(t *testing.T, c *http.Client) *httptest.Server {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// It is important that we use a HTTP request to fetch the flow because that will show us if CSRF works or not
			_, err := w.Write(x.EasyGetBody(t, c, public.URL+login.RouteGetFlow+"?id="+r.URL.Query().Get("flow")))
			require.NoError(t, err)
		}))
		conf.MustSet(config.ViperKeySelfServiceLoginUI, ts.URL)
		conf.MustSet(config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh")
		t.Cleanup(ts.Close)
		return ts
	}

	_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword), map[string]interface{}{
		"enabled": true})

	t.Run("case=fetching successful", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		setupLoginUI(t, client)
		body := x.EasyGetBody(t, client, public.URL+login.RouteInitBrowserFlow)

		assert.NotEmpty(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String(), "%s", body)
		assert.NotEmpty(t, gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Empty(t, gjson.GetBytes(body, "headers").Value(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "ui.action").String(), gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "ui.action").String(), public.URL, "%s", body)
	})

	t.Run("case=csrf cookie missing", func(t *testing.T) {
		client := http.DefaultClient
		setupLoginUI(t, client)
		body := x.EasyGetBody(t, client, public.URL+login.RouteInitBrowserFlow)

		assert.EqualValues(t, x.ErrInvalidCSRFToken.ReasonField, gjson.GetBytes(body, "error.reason").String(), "%s", body)
	})

	t.Run("case=expired", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		setupLoginUI(t, client)
		body := x.EasyGetBody(t, client, public.URL+login.RouteInitBrowserFlow)

		// Expire the flow
		f, err := reg.LoginFlowPersister().GetLoginFlow(context.Background(), uuid.FromStringOrNil(gjson.GetBytes(body, "id").String()))
		require.NoError(t, err)
		f.ExpiresAt = time.Now().Add(-time.Second)
		require.NoError(t, reg.LoginFlowPersister().UpdateLoginFlow(context.Background(), f))

		// Try the flow but it is expired
		res, body := x.EasyGet(t, client, public.URL+login.RouteGetFlow+"?id="+f.ID.String())
		assert.EqualValues(t, http.StatusGone, res.StatusCode)
		assert.Equal(t, public.URL+login.RouteInitBrowserFlow, gjson.GetBytes(body, "error.details.redirect_to").String(), "%s", body)
	})

	t.Run("case=expired with return_to", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		setupLoginUI(t, client)
		body := x.EasyGetBody(t, client, public.URL+login.RouteInitBrowserFlow+"?return_to=https://www.ory.sh")

		// Expire the flow
		f, err := reg.LoginFlowPersister().GetLoginFlow(context.Background(), uuid.FromStringOrNil(gjson.GetBytes(body, "id").String()))
		require.NoError(t, err)
		f.ExpiresAt = time.Now().Add(-time.Second)
		require.NoError(t, reg.LoginFlowPersister().UpdateLoginFlow(context.Background(), f))

		// submit the flow but it is expired
		u := public.URL + login.RouteSubmitFlow + "?flow=" + f.ID.String()
		res, err := client.PostForm(u, url.Values{"password_identifier": {"email@ory.sh"}, "csrf_token": {f.CSRFToken}, "password": {"password"}, "method": {"password"}})
		resBody, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		f, err = reg.LoginFlowPersister().GetLoginFlow(context.Background(), uuid.FromStringOrNil(gjson.GetBytes(resBody, "id").String()))
		require.NoError(t, err)
		assert.Equal(t, public.URL+login.RouteInitBrowserFlow+"?return_to=https://www.ory.sh", f.RequestURL)
	})
}
