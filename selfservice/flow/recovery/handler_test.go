// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package recovery_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/corpx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/assertx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

func init() {
	corpx.RegisterFakes()
}

func TestHandlerRedirectOnAuthenticated(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryEnabled, true)

	router := x.NewRouterPublic()
	ts, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin())

	redirTS := testhelpers.NewRedirTS(t, "already authenticated", conf)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/identity.schema.json")

	t.Run("does redirect to default on authenticated request", func(t *testing.T) {
		body, res := testhelpers.MockMakeAuthenticatedRequest(t, reg, conf, router.Router, x.NewTestHTTPRequest(t, "GET", ts.URL+recovery.RouteInitBrowserFlow, nil))
		assert.Contains(t, res.Request.URL.String(), redirTS.URL, "%+v", res)
		assert.EqualValues(t, "already authenticated", string(body))
	})

	t.Run("does redirect to default on authenticated request", func(t *testing.T) {
		body, res := testhelpers.MockMakeAuthenticatedRequest(t, reg, conf, router.Router, x.NewTestHTTPRequest(t, "GET", ts.URL+recovery.RouteInitAPIFlow, nil))
		assert.Contains(t, res.Request.URL.String(), recovery.RouteInitAPIFlow)
		assert.EqualValues(t, text.ErrIDAlreadyLoggedIn, gjson.GetBytes(body, "error.id").Str)
		assertx.EqualAsJSON(t, recovery.ErrAlreadyLoggedIn, json.RawMessage(gjson.GetBytes(body, "error").Raw))
	})
}

func TestInitFlow(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryEnabled, true)
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(recovery.RecoveryStrategyLink),
		map[string]interface{}{"enabled": true})
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(recovery.RecoveryStrategyCode),
		map[string]interface{}{"enabled": true})

	router := x.NewRouterPublic()
	publicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin())
	recoveryTS := testhelpers.NewRecoveryUIFlowEchoServer(t, reg)

	conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh")
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/identity.schema.json")

	assertion := func(body []byte, isForced, isApi bool) {
		if isApi {
			assert.Equal(t, "api", gjson.GetBytes(body, "type").String())
		} else {
			assert.Equal(t, "browser", gjson.GetBytes(body, "type").String())
		}
	}

	initAuthenticatedFlow := func(t *testing.T, isAPI, isSPA bool) (*http.Response, []byte) {
		route := recovery.RouteInitBrowserFlow
		if isAPI {
			route = recovery.RouteInitAPIFlow
		}
		req := x.NewTestHTTPRequest(t, "GET", publicTS.URL+route, nil)
		if isSPA {
			req.Header.Set("Accept", "application/json")
		}
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
		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		return res, body
	}

	initSPAFlow := func(t *testing.T, hc *http.Client, isSPA bool) (*http.Response, []byte) {
		route := recovery.RouteInitBrowserFlow
		c := publicTS.Client()
		req := x.NewTestHTTPRequest(t, "GET", publicTS.URL+route, nil)
		if isSPA {
			req.Header.Set("Accept", "application/json")
		}
		res, err := c.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
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
			res, body := initAuthenticatedFlow(t, true, false)
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
			assertx.EqualAsJSON(t, recovery.ErrAlreadyLoggedIn, json.RawMessage(gjson.GetBytes(body, "error").Raw), "%s", body)
		})
	})

	t.Run("flow=spa", func(t *testing.T) {
		t.Run("case=creates a new flow on unauthenticated request", func(t *testing.T) {
			res, body := initSPAFlow(t, new(http.Client), true)
			assert.Contains(t, res.Request.URL.String(), recovery.RouteInitBrowserFlow)
			assertion(body, false, false)
		})

		t.Run("case=fails on authenticated request", func(t *testing.T) {
			res, body := initAuthenticatedFlow(t, false, true)
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

		t.Run("case=fails on authenticated request", func(t *testing.T) {
			res, _ := initAuthenticatedFlow(t, false, false)
			assert.Contains(t, res.Request.URL.String(), "https://www.ory.sh")
		})

		t.Run("case=relative redirect when self-service recovery ui is a relative URL", func(t *testing.T) {
			reg.Config().MustSet(ctx, config.ViperKeySelfServiceRecoveryUI, "/recovery-ts")
			assert.Regexp(
				t,
				"^/recovery-ts.*$",
				testhelpers.GetSelfServiceRedirectLocation(t, publicTS.URL+recovery.RouteInitBrowserFlow),
			)
		})

		t.Run("case=redirects with 303", func(t *testing.T) {
			c := &http.Client{}
			// don't get the reference, instead copy the values, so we don't alter the client directly.
			*c = *publicTS.Client()
			// prevent the redirect
			c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}
			req, err := http.NewRequest("GET", publicTS.URL+recovery.RouteInitBrowserFlow, nil)
			require.NoError(t, err)

			res, err := c.Do(req)
			require.NoError(t, err)
			defer res.Body.Close()
			// here we check that the redirect status is 303
			require.Equal(t, http.StatusSeeOther, res.StatusCode)
		})
	})
}

func TestGetFlow(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryEnabled, true)
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(recovery.RecoveryStrategyLink),
		map[string]interface{}{"enabled": true})
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(recovery.RecoveryStrategyCode),
		map[string]interface{}{"enabled": true})
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/identity.schema.json")

	public, _ := testhelpers.NewKratosServerWithCSRF(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)
	_ = testhelpers.NewRedirTS(t, "", conf)

	setupRecoveryTS := func(t *testing.T, c *http.Client) *httptest.Server {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write(x.EasyGetBody(t, c, public.URL+recovery.RouteGetFlow+"?id="+r.URL.Query().Get("flow")))
			require.NoError(t, err)
		}))
		t.Cleanup(ts.Close)
		conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryUI, ts.URL)
		return ts
	}

	t.Run("case=csrf cookie missing", func(t *testing.T) {
		client := http.DefaultClient
		setupRecoveryTS(t, client)
		body := x.EasyGetBody(t, client, public.URL+recovery.RouteInitBrowserFlow)

		assert.EqualValues(t, x.ErrInvalidCSRFToken.ReasonField, gjson.GetBytes(body, "error.reason").String(), "%s", body)
	})

	t.Run("case=valid", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		setupRecoveryTS(t, client)
		body := x.EasyGetBody(t, client, public.URL+recovery.RouteInitBrowserFlow)
		assert.NotEmpty(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String(), "%s", body)
		assert.NotEmpty(t, gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Empty(t, gjson.GetBytes(body, "headers").Value(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "ui.action").String(), gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "ui.action").String(), public.URL, "%s", body)
	})

	t.Run("case=expired", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		setupRecoveryTS(t, client)
		body := x.EasyGetBody(t, client, public.URL+recovery.RouteInitBrowserFlow)

		// Expire the flow
		f, err := reg.RecoveryFlowPersister().GetRecoveryFlow(context.Background(), uuid.FromStringOrNil(gjson.GetBytes(body, "id").String()))
		require.NoError(t, err)
		f.ExpiresAt = time.Now().Add(-time.Second)
		require.NoError(t, reg.RecoveryFlowPersister().UpdateRecoveryFlow(context.Background(), f))

		res, body := x.EasyGet(t, client, public.URL+recovery.RouteGetFlow+"?id="+f.ID.String())
		assert.EqualValues(t, http.StatusGone, res.StatusCode)
		assert.Equal(t, public.URL+recovery.RouteInitBrowserFlow, gjson.GetBytes(body, "error.details.redirect_to").String(), "%s", body)
	})

	t.Run("case=expired with return_to", func(t *testing.T) {
		returnTo := "https://www.ory.sh"
		conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{returnTo})
		client := testhelpers.NewClientWithCookies(t)
		setupRecoveryTS(t, client)
		body := x.EasyGetBody(t, client, public.URL+recovery.RouteInitBrowserFlow+"?return_to="+returnTo)

		// Expire the flow
		f, err := reg.RecoveryFlowPersister().GetRecoveryFlow(context.Background(), uuid.FromStringOrNil(gjson.GetBytes(body, "id").String()))
		require.NoError(t, err)
		f.ExpiresAt = time.Now().Add(-time.Second)
		require.NoError(t, reg.RecoveryFlowPersister().UpdateRecoveryFlow(context.Background(), f))

		// Retrieve the flow and verify that return_to is in the response
		getURL := fmt.Sprintf("%s%s?id=%s&return_to=%s", public.URL, recovery.RouteGetFlow, f.ID, returnTo)
		getBody := x.EasyGetBody(t, client, getURL)
		assert.Equal(t, gjson.GetBytes(getBody, "error.details.return_to").String(), returnTo)

		// submit the flow but it is expired
		u := public.URL + recovery.RouteSubmitFlow + "?flow=" + f.ID.String()
		res, err := client.PostForm(u, url.Values{"email": {"email@ory.sh"}, "csrf_token": {f.CSRFToken}, "method": {"link"}})
		require.NoError(t, err)
		resBody, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		f, err = reg.RecoveryFlowPersister().GetRecoveryFlow(context.Background(), uuid.FromStringOrNil(gjson.GetBytes(resBody, "id").String()))
		require.NoError(t, err)
		assert.Equal(t, public.URL+recovery.RouteInitBrowserFlow+"?return_to="+returnTo, f.RequestURL)
	})

	t.Run("case=not found", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		setupRecoveryTS(t, client)

		res, _ := x.EasyGet(t, client, public.URL+recovery.RouteGetFlow+"?id="+x.NewUUID().String())
		assert.EqualValues(t, http.StatusNotFound, res.StatusCode)
	})
}
