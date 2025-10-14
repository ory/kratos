// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package registration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/corpx"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hydra"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/kratos/selfservice/strategy/password"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/nosurfx"
	"github.com/ory/x/assertx"
	"github.com/ory/x/ioutilx"
	"github.com/ory/x/urlx"
)

func init() {
	corpx.RegisterFakes()
}

func TestHandlerRedirectOnAuthenticated(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	fakeHydra := hydra.NewFake()
	reg.SetHydra(fakeHydra)

	router := x.NewRouterPublic(reg)
	ts, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin(reg))

	// Set it first as otherwise it will overwrite the ViperKeySelfServiceBrowserDefaultReturnTo key;
	returnToTS := testhelpers.NewRedirTS(t, "return_to", conf)
	conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{returnToTS.URL})

	redirTS := testhelpers.NewRedirTS(t, "already authenticated", conf)
	conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationEnabled, true)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/identity.schema.json")

	t.Run("does redirect to default on authenticated request", func(t *testing.T) {
		body, res := testhelpers.MockMakeAuthenticatedRequest(t, reg, conf, router, testhelpers.NewTestHTTPRequest(t, "GET", ts.URL+registration.RouteInitBrowserFlow, nil))
		assert.Contains(t, res.Request.URL.String(), redirTS.URL)
		assert.EqualValues(t, "already authenticated", string(body))
	})

	t.Run("does redirect to default on authenticated request", func(t *testing.T) {
		body, res := testhelpers.MockMakeAuthenticatedRequest(t, reg, conf, router, testhelpers.NewTestHTTPRequest(t, "GET", ts.URL+registration.RouteInitAPIFlow, nil))
		assert.Contains(t, res.Request.URL.String(), registration.RouteInitAPIFlow)
		assertx.EqualAsJSON(t, registration.ErrAlreadyLoggedIn, json.RawMessage(gjson.GetBytes(body, "error").Raw))
	})

	t.Run("does redirect to return_to url on authenticated request", func(t *testing.T) {
		body, res := testhelpers.MockMakeAuthenticatedRequest(t, reg, conf, router, testhelpers.NewTestHTTPRequest(t, "GET", ts.URL+registration.RouteInitBrowserFlow+"?return_to="+returnToTS.URL, nil))
		assert.Contains(t, res.Request.URL.String(), returnToTS.URL)
		assert.EqualValues(t, "return_to", string(body))
	})

	t.Run("oauth2 with session and skip=false is redirected to login", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeyOAuth2ProviderURL, "https://fake-hydra")

		fakeHydra.RequestURL = "https://www.ory.sh/oauth2/auth?audience=&client_id=foo&login_verifier="
		fakeHydra.Skip = false

		client := testhelpers.NewClientWithCookies(t)
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
		_, res := testhelpers.MockMakeAuthenticatedRequestWithClient(t, reg, conf, router, testhelpers.NewTestHTTPRequest(t, "GET", ts.URL+registration.RouteInitBrowserFlow+"?login_challenge="+hydra.FakeValidLoginChallenge, nil), client)
		assert.Contains(t, res.Header.Get("location"), login.RouteInitBrowserFlow)
	})

	t.Run("oauth2 with session and skip=true is accepted", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeyOAuth2ProviderURL, "https://fake-hydra")

		fakeHydra.Skip = true
		fakeHydra.RequestURL = "https://www.ory.sh/oauth2/auth?audience=&client_id=foo&login_verifier="

		client := testhelpers.NewClientWithCookies(t)
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
		_, res := testhelpers.MockMakeAuthenticatedRequestWithClient(t, reg, conf, router, testhelpers.NewTestHTTPRequest(t, "GET", ts.URL+registration.RouteInitBrowserFlow+"?login_challenge="+hydra.FakeValidLoginChallenge, nil), client)
		assert.Contains(t, res.Header.Get("location"), hydra.FakePostLoginURL)
	})
}

func TestInitFlow(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword),
		map[string]interface{}{"enabled": true})

	router := x.NewRouterPublic(reg)
	publicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin(reg))
	registrationTS := testhelpers.NewRegistrationUIFlowEchoServer(t, reg)
	returnToTS := testhelpers.NewRedirTS(t, "return_to", conf)

	conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationEnabled, true)
	conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, returnToTS.URL)

	conf.MustSet(ctx, config.ViperKeyIdentitySchemas, config.Schemas{
		{ID: "default", URL: "file://./stub/registration.schema.json"},
		{ID: "email", URL: "file://./stub/registration.schema.json", SelfserviceSelectable: true},
		{ID: "phone", URL: "file://./stub/registration.phone.schema.json", SelfserviceSelectable: true},
		{ID: "not-allowed", URL: "file://./stub/registration.schema.json"},
	})
	conf.MustSet(ctx, config.ViperKeyDefaultIdentitySchemaID, "email")
	errTS := testhelpers.NewErrorTestServer(t, reg)

	assertion := func(body []byte, isForced, isApi bool) {
		if isApi {
			assert.Equal(t, "api", gjson.GetBytes(body, "type").String())
		} else {
			assert.Equal(t, "browser", gjson.GetBytes(body, "type").String())
		}
	}

	initAuthenticatedFlow := func(t *testing.T, isAPI bool, isSPA bool) (*http.Response, []byte) {
		route := registration.RouteInitBrowserFlow
		if isAPI {
			route = registration.RouteInitAPIFlow
		}
		req := testhelpers.NewTestHTTPRequest(t, "GET", publicTS.URL+route, nil)
		if isSPA {
			req.Header.Set("Accept", "application/json")
		}
		body, res := testhelpers.MockMakeAuthenticatedRequest(t, reg, conf, router, req)
		if isAPI {
			assert.Len(t, res.Header.Get("Set-Cookie"), 0)
		}
		return res, body
	}

	initFlowWithAccept := func(t *testing.T, query url.Values, isAPI bool, accept string) (*http.Response, []byte) {
		route := registration.RouteInitBrowserFlow
		if isAPI {
			route = registration.RouteInitAPIFlow
		}
		c := publicTS.Client()
		req, err := http.NewRequest("GET", publicTS.URL+route+"?"+query.Encode(), nil)
		require.NoError(t, err)
		if accept != "" {
			req.Header.Set("Accept", accept)
		}

		res, err := c.Do(req)
		require.NoError(t, err)
		defer func() { _ = res.Body.Close() }()
		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		return res, body
	}

	initFlow := func(t *testing.T, query url.Values, isAPI bool) (*http.Response, []byte) {
		return initFlowWithAccept(t, query, isAPI, "")
	}

	initSPAFlow := func(t *testing.T) (*http.Response, []byte) {
		return initFlowWithAccept(t, url.Values{}, false, "application/json")
	}

	t.Run("suite=identity schema in query", func(t *testing.T) {
		for _, tc := range []struct {
			name    string
			query   url.Values
			wantErr bool
			assert  func(*testing.T, []byte)
		}{{
			name:    "not-allowed",
			query:   url.Values{"identity_schema": {"not-allowed"}},
			wantErr: true,
		}, {
			name:    "not-found",
			query:   url.Values{"identity_schema": {"not-found"}},
			wantErr: true,
		}, {
			name:  "phone",
			query: url.Values{"identity_schema": {"phone"}},
			assert: func(t *testing.T, body []byte) {
				assert.True(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==traits.phone)").Exists(), "%s", body)
				assert.False(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==traits.email)").Exists(), "%s", body)
			},
		}, {
			name:  "email",
			query: url.Values{"identity_schema": {"email"}},
			assert: func(t *testing.T, body []byte) {
				assert.False(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==traits.phone)").Exists(), "%s", body)
				assert.True(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==traits.email)").Exists(), "%s", body)
			},
		}} {
			t.Run("case="+tc.name, func(t *testing.T) {
				t.Run("flow=api", func(t *testing.T) {
					res, body := initFlow(t, tc.query, true)
					if tc.wantErr {
						assert.Equal(t, http.StatusBadRequest, res.StatusCode)
						return
					}
					tc.assert(t, body)
				})

				t.Run("flow=browser", func(t *testing.T) {
					res, body := initFlow(t, tc.query, false)
					if tc.wantErr {
						require.Contains(t, res.Request.URL.String(), errTS.URL, "%s", body)
						assert.EqualValues(t, "Bad Request", gjson.GetBytes(body, "status").String(), "%s", body)
						return
					}
					tc.assert(t, body)
				})
			})
		}
	})

	t.Run("flow=api", func(t *testing.T) {
		t.Run("case=creates a new flow on unauthenticated request", func(t *testing.T) {
			res, body := initFlow(t, url.Values{}, true)
			assert.Contains(t, res.Request.URL.String(), registration.RouteInitAPIFlow)
			assertion(body, false, true)
			assert.Empty(t, gjson.GetBytes(body, "session_token_exchange_code").String())
		})

		t.Run("case=returns a session token exchange code", func(t *testing.T) {
			res, body := initFlow(t, urlx.ParseOrPanic("/?return_session_token_exchange_code=true").Query(), true)
			assert.Contains(t, res.Request.URL.String(), registration.RouteInitAPIFlow)
			assertion(body, false, true)
			assert.NotEmpty(t, gjson.GetBytes(body, "session_token_exchange_code").String())
		})

		t.Run("case=fails on authenticated request", func(t *testing.T) {
			res, body := initAuthenticatedFlow(t, true, false)
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
			assertx.EqualAsJSON(t, registration.ErrAlreadyLoggedIn, json.RawMessage(gjson.GetBytes(body, "error").Raw), "%s", body)
		})
	})

	t.Run("flow=browser", func(t *testing.T) {
		t.Run("case=does not set forced flag on unauthenticated request", func(t *testing.T) {
			res, body := initFlow(t, url.Values{}, false)
			assertion(body, false, false)
			assert.Contains(t, res.Request.URL.String(), registrationTS.URL)
			assert.Empty(t, gjson.GetBytes(body, "session_token_exchange_code").String())
		})

		t.Run("case=never returns a session token exchange code", func(t *testing.T) {
			_, body := initFlow(t, urlx.ParseOrPanic("/?return_session_token_exchange_code=true").Query(), false)
			assertion(body, false, false)
			assert.Empty(t, gjson.GetBytes(body, "session_token_exchange_code").String())
		})

		t.Run("case=makes request with JSON", func(t *testing.T) {
			res, body := initSPAFlow(t)
			assertion(body, false, false)
			assert.NotContains(t, res.Request.URL.String(), registrationTS.URL)
		})

		t.Run("case=redirects when already authenticated", func(t *testing.T) {
			res, _ := initAuthenticatedFlow(t, false, false)
			assert.Contains(t, res.Request.URL.String(), returnToTS.URL)
		})

		t.Run("case=responds with error if already authenticated and SPA", func(t *testing.T) {
			res, body := initAuthenticatedFlow(t, false, true)
			assert.NotContains(t, res.Request.URL.String(), returnToTS.URL)
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
			assertx.EqualAsJSON(t, registration.ErrAlreadyLoggedIn, json.RawMessage(gjson.GetBytes(body, "error").Raw), "%s", body)
		})

		t.Run("case=relative redirect when self-service registration ui is a relative URL", func(t *testing.T) {
			reg.Config().MustSet(ctx, config.ViperKeySelfServiceRegistrationUI, "/registration-ts")
			assert.Regexp(
				t,
				"^/registration-ts.*$",
				testhelpers.GetSelfServiceRedirectLocation(t, publicTS.URL+registration.RouteInitBrowserFlow),
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
			req, err := http.NewRequest("GET", publicTS.URL+registration.RouteInitBrowserFlow, nil)
			require.NoError(t, err)

			res, err := c.Do(req)
			require.NoError(t, err)
			defer func() { _ = res.Body.Close() }()
			// here we check that the redirect status is 303
			require.Equal(t, http.StatusSeeOther, res.StatusCode)
		})
	})
}

func TestDisabledFlow(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)

	conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationEnabled, false)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/login.schema.json")
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword),
		map[string]interface{}{"enabled": true})

	publicTS, _ := testhelpers.NewKratosServerWithCSRF(t, reg)
	errTS := testhelpers.NewErrorTestServer(t, reg)

	makeRequest := func(t *testing.T, route string, isSPA bool) (*http.Response, []byte) {
		c := publicTS.Client()
		req, err := http.NewRequest("GET", publicTS.URL+route, nil)
		require.NoError(t, err)

		if isSPA {
			req.Header.Set("Accept", "application/json")
		}

		res, err := c.Do(req)
		require.NoError(t, err)
		defer func() { _ = res.Body.Close() }()
		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		return res, body
	}

	t.Run("flow=api", func(t *testing.T) {
		t.Run("case=init fails when flow disabled", func(t *testing.T) {
			res, body := makeRequest(t, registration.RouteInitAPIFlow, false)
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
			assertx.EqualAsJSON(t, registration.ErrRegistrationDisabled, json.RawMessage(gjson.GetBytes(body, "error").Raw), "%s", body)
		})

		t.Run("case=get flow fails when flow disabled", func(t *testing.T) {
			res, body := makeRequest(t, registration.RouteGetFlow+"?id="+x.NewUUID().String(), false)
			require.Contains(t, res.Request.URL.String(), errTS.URL, "%s", body)
			assert.EqualValues(t, registration.ErrRegistrationDisabled.ReasonField, gjson.GetBytes(body, "reason").String(), "%s", body)
		})
	})

	t.Run("flow=browser", func(t *testing.T) {
		t.Run("case=init responds with error if flow disabled and SPA", func(t *testing.T) {
			res, body := makeRequest(t, registration.RouteInitBrowserFlow, true)
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
			assertx.EqualAsJSON(t, registration.ErrRegistrationDisabled, json.RawMessage(gjson.GetBytes(body, "error").Raw), "%s", body)
		})

		t.Run("case=get flow responds with error if flow disabled and SPA", func(t *testing.T) {
			res, body := makeRequest(t, registration.RouteGetFlow+"?id="+x.NewUUID().String(), true)
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
			assertx.EqualAsJSON(t, registration.ErrRegistrationDisabled, json.RawMessage(gjson.GetBytes(body, "error").Raw), "%s", body)
		})
	})
}

func TestGetFlow(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationEnabled, true)
	returnToTS := testhelpers.NewRedirTS(t, "return_to", conf)

	conf.MustSet(ctx, config.ViperKeyIdentitySchemas, config.Schemas{
		{ID: "email", URL: "file://./stub/registration.schema.json", SelfserviceSelectable: true},
		{ID: "phone", URL: "file://./stub/registration.phone.schema.json", SelfserviceSelectable: true},
		{ID: "not-allowed", URL: "file://./stub/registration.schema.json"},
	})
	conf.MustSet(ctx, config.ViperKeyDefaultIdentitySchemaID, "email")

	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword),
		map[string]interface{}{"enabled": true})

	public, _ := testhelpers.NewKratosServerWithCSRF(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)
	_ = testhelpers.NewRedirTS(t, "", conf)

	setupRegistrationUI := func(t *testing.T, c *http.Client) *httptest.Server {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write(testhelpers.EasyGetBody(t, c, public.URL+registration.RouteGetFlow+"?id="+r.URL.Query().Get("flow")))
			require.NoError(t, err)
		}))
		t.Cleanup(ts.Close)
		conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationUI, ts.URL)
		return ts
	}

	t.Run("case=valid", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		_ = setupRegistrationUI(t, client)
		body := testhelpers.EasyGetBody(t, client, public.URL+registration.RouteInitBrowserFlow)

		assert.NotEmpty(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String(), "%s", body)
		assert.NotEmpty(t, gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Empty(t, gjson.GetBytes(body, "headers").Value(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "ui.action").String(), gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "ui.action").String(), public.URL, "%s", body)
	})

	t.Run("case=csrf cookie missing", func(t *testing.T) {
		client := http.DefaultClient
		_ = setupRegistrationUI(t, client)
		body := testhelpers.EasyGetBody(t, client, public.URL+registration.RouteInitBrowserFlow)

		assert.EqualValues(t, nosurfx.ErrInvalidCSRFToken.ReasonField, gjson.GetBytes(body, "error.reason").String(), "%s", body)
	})

	t.Run("case=expired", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		setupRegistrationUI(t, client)
		body := testhelpers.EasyGetBody(t, client, public.URL+registration.RouteInitBrowserFlow)

		// Expire the flow
		f, err := reg.RegistrationFlowPersister().GetRegistrationFlow(context.Background(), uuid.FromStringOrNil(gjson.GetBytes(body, "id").String()))
		require.NoError(t, err)
		f.ExpiresAt = time.Now().Add(-time.Second)
		require.NoError(t, reg.RegistrationFlowPersister().UpdateRegistrationFlow(context.Background(), f))

		res, body := testhelpers.EasyGet(t, client, public.URL+registration.RouteGetFlow+"?id="+f.ID.String())
		assert.EqualValues(t, http.StatusGone, res.StatusCode)
		assert.Equal(t, public.URL+registration.RouteInitBrowserFlow, gjson.GetBytes(body, "error.details.redirect_to").String(), "%s", body)
	})

	t.Run("case=expired with return_to and identity_schema", func(t *testing.T) {
		returnTo := returnToTS.URL
		conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{returnTo})

		client := testhelpers.NewClientWithCookies(t)
		setupRegistrationUI(t, client)
		body := testhelpers.EasyGetBody(t, client, public.URL+registration.RouteInitBrowserFlow+
			"?return_to="+returnTo+
			"&identity_schema=email")

		// Expire the flow
		f, err := reg.RegistrationFlowPersister().GetRegistrationFlow(context.Background(), uuid.FromStringOrNil(gjson.GetBytes(body, "id").String()))
		require.NoError(t, err)
		f.ExpiresAt = time.Now().Add(-time.Second)
		require.NoError(t, reg.RegistrationFlowPersister().UpdateRegistrationFlow(context.Background(), f))

		// Retrieve the flow and verify that return_to is in the response
		getURL := fmt.Sprintf("%s%s?id=%s&return_to=%s", public.URL, registration.RouteGetFlow, f.ID, returnTo)
		getBody := testhelpers.EasyGetBody(t, client, getURL)
		assert.Equal(t, gjson.GetBytes(getBody, "error.details.return_to").String(), returnTo)

		// submit the flow but it is expired
		u := public.URL + registration.RouteSubmitFlow + "?flow=" + f.ID.String()
		res, err := client.PostForm(u, url.Values{"method": {"password"}, "csrf_token": {f.CSRFToken}, "password": {"password"}, "traits.email": {"email@ory.sh"}})
		require.NoError(t, err)
		resBody, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		f, err = reg.RegistrationFlowPersister().GetRegistrationFlow(context.Background(), uuid.FromStringOrNil(gjson.GetBytes(resBody, "id").String()))
		require.NoError(t, err)
		assert.Equal(t, public.URL+registration.RouteInitBrowserFlow+"?return_to="+returnTo+"&identity_schema=email", f.RequestURL)
	})

	t.Run("case=not found", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		setupRegistrationUI(t, client)

		res, _ := testhelpers.EasyGet(t, client, public.URL+registration.RouteGetFlow+"?id="+x.NewUUID().String())
		assert.EqualValues(t, http.StatusNotFound, res.StatusCode)
	})
}

// This test verifies that the password method is still executed even if the
// oidc strategy is ordered before the password strategy
// when submitting the form with both `method=password` and `provider=google`.
func TestOIDCStrategyOrder(t *testing.T) {
	t.Logf("This test has been set up to validate the current incorrect `oidc` behaviour. When submitting the form, the `oidc` strategy is executed first, even if the method is set to `password`.")

	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)

	// reorder the strategies
	reg.WithSelfserviceStrategies(t, []any{
		oidc.NewStrategy(reg, oidc.ForCredentialType(identity.CredentialsTypeOIDC)),
		password.NewStrategy(reg),
	})

	conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationEnabled, true)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/registration.schema.json")
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword),
		map[string]interface{}{"enabled": true})
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeOIDC),
		map[string]interface{}{"enabled": true})
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeCodeAuth),
		map[string]interface{}{"passwordless_enabled": true})
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeOIDC)+".config", &oidc.ConfigurationCollection{Providers: []oidc.Configuration{
		{
			ID:           "google",
			Provider:     "google",
			ClientID:     "1234",
			ClientSecret: "1234",
		},
	}})

	public, _ := testhelpers.NewKratosServerWithCSRF(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)
	_ = testhelpers.NewRedirTS(t, "", conf)

	setupRegistrationUI := func(t *testing.T, c *http.Client) *httptest.Server {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write(testhelpers.EasyGetBody(t, c, public.URL+registration.RouteGetFlow+"?id="+r.URL.Query().Get("flow")))
			require.NoError(t, err)
		}))
		t.Cleanup(ts.Close)
		conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationUI, ts.URL)
		return ts
	}

	t.Run("case=accept `password` method while including `provider:google`", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		_ = setupRegistrationUI(t, client)
		body := testhelpers.EasyGetBody(t, client, public.URL+registration.RouteInitBrowserFlow)

		flow := gjson.GetBytes(body, "id").String()

		csrfToken := gjson.GetBytes(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String()
		email := faker.Email()
		payload := json.RawMessage(`{"traits": {"email": "` + email + `"},"method": "password","password": "asdasdasdsa21312@#!@%","provider": "google","csrf_token": "` + csrfToken + `"}`)

		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, public.URL+registration.RouteSubmitFlow+"?flow="+flow, bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		require.NoError(t, err)
		resp, err := client.Do(req)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode, "%s", ioutilx.MustReadAll(resp.Body))

		verifiableAddress, err := reg.PrivilegedIdentityPool().FindVerifiableAddressByValue(ctx, identity.VerifiableAddressTypeEmail, email)
		require.NoError(t, err)
		require.Equal(t, strings.ToLower(email), verifiableAddress.Value)

		id, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, verifiableAddress.IdentityID)
		require.NoError(t, err)
		require.NotNil(t, id.ID)

		_, ok := id.GetCredentials(identity.CredentialsTypePassword)
		require.True(t, ok)
	})

	t.Run("case=accept oidc flow with just `provider:google`", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		_ = setupRegistrationUI(t, client)
		body := testhelpers.EasyGetBody(t, client, public.URL+registration.RouteInitBrowserFlow)

		flow := gjson.GetBytes(body, "id").String()

		csrfToken := gjson.GetBytes(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String()

		payload := json.RawMessage(`{"provider": "google","csrf_token": "` + csrfToken + `"}`)

		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, public.URL+registration.RouteSubmitFlow+"?flow="+flow, bytes.NewBuffer(payload))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Containsf(t,
			gjson.GetBytes(b, "error.reason").String(),
			"In order to complete this flow please redirect the browser to: https://accounts.google.com/o/oauth2/v2/auth",
			"%s", b,
		)
	})
}
