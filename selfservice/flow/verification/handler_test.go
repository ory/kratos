// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package verification_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gobuffalo/httptest"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hydra"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/x"
)

func TestGetFlow(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, config.ViperKeySelfServiceVerificationEnabled, true)
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(verification.VerificationStrategyLink),
		map[string]interface{}{"enabled": true})
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/identity.schema.json")

	public, _ := testhelpers.NewKratosServerWithCSRF(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)
	_ = testhelpers.NewRedirTS(t, "", conf)

	setupVerificationUI := func(t *testing.T, c *http.Client) *httptest.Server {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write(testhelpers.EasyGetBody(t, c, public.URL+verification.RouteGetFlow+"?id="+r.URL.Query().Get("flow")))
			require.NoError(t, err)
		}))
		t.Cleanup(ts.Close)
		conf.MustSet(ctx, config.ViperKeySelfServiceVerificationUI, ts.URL)
		return ts
	}

	assertFlowPayload := func(t *testing.T, body []byte, isApi bool) {
		if isApi {
			assert.Equal(t, "api", gjson.GetBytes(body, "type").String(), "%s", body)
			assert.Empty(t, gjson.GetBytes(body, "ui.fields.#(attributes.name==csrf_token).attributes.value").String(), "%s", body)
		} else {
			assert.Equal(t, "browser", gjson.GetBytes(body, "type").String(), "%s", body)
			assert.NotEmpty(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String(), "%s", body)
		}

		assert.NotEmpty(t, gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Empty(t, gjson.GetBytes(body, "headers").Value(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "ui.action").String(), gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "ui.action").String(), public.URL, "%s", body)
	}

	t.Run("case=valid", func(t *testing.T) {
		t.Run("type=browser", func(t *testing.T) {
			client := testhelpers.NewClientWithCookies(t)
			_ = setupVerificationUI(t, client)
			res, body := testhelpers.EasyGet(t, client, public.URL+verification.RouteInitBrowserFlow)
			require.NotEqualValues(t, res.Request.URL.String(), public.URL+verification.RouteInitBrowserFlow)
			assertFlowPayload(t, body, false)
		})

		t.Run("type=spa", func(t *testing.T) {
			client := testhelpers.NewClientWithCookies(t)
			_ = setupVerificationUI(t, client)
			res, body := testhelpers.EasyGetJSON(t, client, public.URL+verification.RouteInitBrowserFlow)
			require.EqualValues(t, res.Request.URL.String(), public.URL+verification.RouteInitBrowserFlow)
			assertFlowPayload(t, body, false)
		})

		t.Run("type=api", func(t *testing.T) {
			client := testhelpers.NewClientWithCookies(t)
			_ = setupVerificationUI(t, client)
			res, body := testhelpers.EasyGet(t, client, public.URL+verification.RouteInitAPIFlow)
			assert.Len(t, res.Header.Get("Set-Cookie"), 0)
			assertFlowPayload(t, body, true)
		})
	})

	t.Run("case=csrf cookie missing", func(t *testing.T) {
		client := http.DefaultClient
		_ = setupVerificationUI(t, client)
		body := testhelpers.EasyGetBody(t, client, public.URL+verification.RouteInitBrowserFlow)

		assert.EqualValues(t, x.ErrInvalidCSRFToken.ReasonField, gjson.GetBytes(body, "error.reason").String(), "%s", body)
	})

	t.Run("case=expired", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		_ = setupVerificationUI(t, client)
		body := testhelpers.EasyGetBody(t, client, public.URL+verification.RouteInitBrowserFlow)

		// Expire the flow
		f, err := reg.VerificationFlowPersister().GetVerificationFlow(context.Background(), uuid.FromStringOrNil(gjson.GetBytes(body, "id").String()))
		require.NoError(t, err)
		f.ExpiresAt = time.Now().Add(-time.Second)
		require.NoError(t, reg.VerificationFlowPersister().UpdateVerificationFlow(context.Background(), f))

		res, body := testhelpers.EasyGet(t, client, public.URL+verification.RouteGetFlow+"?id="+f.ID.String())
		assert.EqualValues(t, http.StatusGone, res.StatusCode)
		assert.Equal(t, public.URL+verification.RouteInitBrowserFlow, gjson.GetBytes(body, "error.details.redirect_to").String(), "%s", body)
	})

	t.Run("case=expired with return_to", func(t *testing.T) {
		returnTo := "https://www.ory.sh"
		conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{returnTo})

		client := testhelpers.NewClientWithCookies(t)
		_ = setupVerificationUI(t, client)
		body := testhelpers.EasyGetBody(t, client, public.URL+verification.RouteInitBrowserFlow+"?return_to="+returnTo)

		// Expire the flow
		f, err := reg.VerificationFlowPersister().GetVerificationFlow(context.Background(), uuid.FromStringOrNil(gjson.GetBytes(body, "id").String()))
		require.NoError(t, err)
		f.ExpiresAt = time.Now().Add(-time.Second)
		require.NoError(t, reg.VerificationFlowPersister().UpdateVerificationFlow(context.Background(), f))

		// Retrieve the flow and verify that return_to is in the response
		getURL := fmt.Sprintf("%s%s?id=%s&return_to=%s", public.URL, verification.RouteGetFlow, f.ID, returnTo)
		getBody := testhelpers.EasyGetBody(t, client, getURL)
		assert.Equal(t, gjson.GetBytes(getBody, "error.details.return_to").String(), returnTo)

		// submit the flow but it is expired
		u := public.URL + verification.RouteSubmitFlow + "?flow=" + f.ID.String()
		res, err := client.PostForm(u, url.Values{"method": {"link"}, "csrf_token": {f.CSRFToken}, "email": {"email@ory.sh"}})
		require.NoError(t, err)
		resBody, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		f, err = reg.VerificationFlowPersister().GetVerificationFlow(context.Background(), uuid.FromStringOrNil(gjson.GetBytes(resBody, "id").String()))
		require.NoError(t, err)
		assert.Equal(t, public.URL+verification.RouteInitBrowserFlow+"?return_to="+returnTo, f.RequestURL)
	})

	t.Run("case=relative redirect when self-service verification ui is a relative URL", func(t *testing.T) {
		router := x.NewRouterPublic()
		ts, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin())
		reg.Config().MustSet(ctx, config.ViperKeySelfServiceVerificationUI, "/verification-ts")
		assert.Regexp(
			t,
			"^/verification-ts.*$",
			testhelpers.GetSelfServiceRedirectLocation(t, ts.URL+verification.RouteInitBrowserFlow),
		)
	})

	t.Run("case=not found", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		_ = setupVerificationUI(t, client)

		res, _ := testhelpers.EasyGet(t, client, public.URL+verification.RouteGetFlow+"?id="+x.NewUUID().String())
		assert.EqualValues(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("case=redirects with 303", func(t *testing.T) {
		router := x.NewRouterPublic()
		ts, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin())

		// prevent the redirect
		ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
		req, err := http.NewRequest("GET", ts.URL+verification.RouteInitBrowserFlow, nil)
		require.NoError(t, err)

		res, err := ts.Client().Do(req)
		require.NoError(t, err)
		defer res.Body.Close()
		// here we check that the redirect status is 303
		require.Equal(t, http.StatusSeeOther, res.StatusCode)
	})
}

func TestPostFlow(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	reg.WithSelfserviceStrategies(t, []any{&verification.FakeStrategy{}})
	reg.WithHydra(hydra.NewFake())
	conf.MustSet(ctx, config.ViperKeySelfServiceVerificationEnabled, true)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/identity.schema.json")

	public, _ := testhelpers.NewKratosServerWithCSRF(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)
	_ = testhelpers.NewRedirTS(t, "", conf)

	t.Run("case=valid", func(t *testing.T) {
		f := &verification.Flow{
			ID:        uuid.Must(uuid.NewV4()),
			Type:      "browser",
			ExpiresAt: time.Now().Add(1 * time.Hour),
			IssuedAt:  time.Now(),
			State:     flow.StateChooseMethod,
		}
		require.NoError(t, reg.VerificationFlowPersister().CreateVerificationFlow(ctx, f))

		client := testhelpers.NewClientWithCookies(t)

		u := public.URL + verification.RouteSubmitFlow + "?flow=" + f.ID.String()
		resp, err := client.PostForm(u, url.Values{"method": {"fake"}})
		require.NoError(t, err)
		assert.EqualValues(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("suite=with OIDC login challenge", func(t *testing.T) {
		t.Run("case=succeeds with a session", func(t *testing.T) {
			s := testhelpers.CreateSession(t, reg)

			f := &verification.Flow{
				ID:                   uuid.Must(uuid.NewV4()),
				Type:                 "browser",
				ExpiresAt:            time.Now().Add(1 * time.Hour),
				IssuedAt:             time.Now(),
				OAuth2LoginChallenge: hydra.FakeValidLoginChallenge,
				OAuth2LoginChallengeParams: verification.OAuth2LoginChallengeParams{
					SessionID:  uuid.NullUUID{UUID: s.ID, Valid: true},
					IdentityID: uuid.NullUUID{UUID: s.IdentityID, Valid: true},
					AMR:        s.AMR,
				},
				State: flow.StatePassedChallenge,
			}
			require.NoError(t, reg.VerificationFlowPersister().CreateVerificationFlow(ctx, f))

			client := testhelpers.NewNoRedirectClientWithCookies(t)

			u := public.URL + verification.RouteSubmitFlow + "?flow=" + f.ID.String()
			resp, err := client.PostForm(u, url.Values{"method": {"fake"}})
			require.NoError(t, err)
			assert.Equal(t, http.StatusSeeOther, resp.StatusCode)
			assert.Equal(t, hydra.FakePostLoginURL, resp.Header.Get("Location"))
		})

		t.Run("case=fails without a session", func(t *testing.T) {
			client := testhelpers.NewClientWithCookies(t)
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, err := w.Write(testhelpers.EasyGetBody(t, client, public.URL+verification.RouteGetFlow+"?id="+r.URL.Query().Get("flow")))
				require.NoError(t, err)
			}))
			t.Cleanup(ts.Close)
			conf.MustSet(ctx, config.ViperKeySelfServiceVerificationUI, ts.URL)

			f := &verification.Flow{
				ID:                   uuid.Must(uuid.NewV4()),
				Type:                 "browser",
				ExpiresAt:            time.Now().Add(1 * time.Hour),
				IssuedAt:             time.Now(),
				OAuth2LoginChallenge: hydra.FakeValidLoginChallenge,
				State:                flow.StateChooseMethod,
			}
			require.NoError(t, reg.VerificationFlowPersister().CreateVerificationFlow(ctx, f))

			u := public.URL + verification.RouteSubmitFlow + "?flow=" + f.ID.String()
			resp, err := client.PostForm(u, url.Values{"method": {"fake"}})
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, f.ID.String(), resp.Request.URL.Query().Get("flow"))
		})
	})
}
