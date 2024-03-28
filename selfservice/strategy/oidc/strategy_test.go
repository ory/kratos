// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ory/kratos/selfservice/hook/hooktest"
	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/hydra"
	"github.com/ory/kratos/selfservice/sessiontokenexchange"
	"github.com/ory/kratos/session"
	"github.com/ory/x/randx"
	"github.com/ory/x/snapshotx"

	"github.com/ory/kratos/text"

	"github.com/ory/x/ioutilx"

	"github.com/ory/x/assertx"

	"github.com/ory/kratos/ui/container"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"

	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/kratos/x"
)

func TestStrategy(t *testing.T) {
	ctx := context.Background()
	if testing.Short() {
		t.Skip()
	}

	var (
		conf, reg = internal.NewFastRegistryWithMocks(t)
		subject   string
		claims    idTokenClaims
		scope     []string
	)
	remoteAdmin, remotePublic, hydraIntegrationTSURL := newHydra(t, &subject, &claims, &scope)
	returnTS := newReturnTs(t, reg)
	conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{returnTS.URL})
	uiTS := newUI(t, reg)
	errTS := testhelpers.NewErrorTestServer(t, reg)
	routerP := x.NewRouterPublic()
	routerA := x.NewRouterAdmin()
	ts, _ := testhelpers.NewKratosServerWithRouters(t, reg, routerP, routerA)
	invalid := newOIDCProvider(t, ts, remotePublic, remoteAdmin, "invalid-issuer")
	viperSetProviderConfig(
		t,
		conf,
		newOIDCProvider(t, ts, remotePublic, remoteAdmin, "valid"),
		newOIDCProvider(t, ts, remotePublic, remoteAdmin, "secondProvider"),
		newOIDCProvider(t, ts, remotePublic, remoteAdmin, "claimsViaUserInfo", func(c *oidc.Configuration) {
			c.ClaimsSource = oidc.ClaimsSourceUserInfo
		}),
		oidc.Configuration{
			Provider:     "generic",
			ID:           "invalid-issuer",
			ClientID:     invalid.ClientID,
			ClientSecret: invalid.ClientSecret,
			// We replace this URL to cause an issuer validation mismatch.
			IssuerURL: strings.Replace(remotePublic, "localhost", "127.0.0.1", 1) + "/",
			Mapper:    "file://./stub/oidc.hydra.jsonnet",
		},
	)

	conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationEnabled, true)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/registration.schema.json")
	conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter,
		identity.CredentialsTypeOIDC.String()), []config.SelfServiceHook{{Name: "session"}})

	t.Logf("Kratos Public URL: %s", ts.URL)
	t.Logf("Kratos Error URL: %s", errTS.URL)
	t.Logf("Hydra Public URL: %s", remotePublic)
	t.Logf("Hydra Admin URL: %s", remoteAdmin)
	t.Logf("Hydra Integration URL: %s", hydraIntegrationTSURL)
	t.Logf("Return URL: %s", returnTS.URL)

	subject = "foo@bar.com"
	scope = []string{}

	// assert form values
	assertFormValues := func(t *testing.T, flowID uuid.UUID, provider string) (action string) {
		var config *container.Container
		if req, err := reg.RegistrationFlowPersister().GetRegistrationFlow(context.Background(), flowID); err == nil {
			require.EqualValues(t, req.ID, flowID)
			config = req.UI
			require.NotNil(t, config)
		} else if req, err := reg.LoginFlowPersister().GetLoginFlow(context.Background(), flowID); err == nil {
			require.EqualValues(t, req.ID, flowID)
			config = req.UI
			require.NotNil(t, config)
		} else {
			require.NoError(t, err)
			return
		}

		assert.Equal(t, "POST", config.Method)

		var found bool
		for _, field := range config.Nodes {
			if strings.Contains(field.ID(), "provider") && field.GetValue() == provider {
				found = true
				break
			}
		}
		require.True(t, found, "%+v", assertx.PrettifyJSONPayload(t, config))

		return config.Action
	}

	registerAction := func(flowID uuid.UUID) string {
		return ts.URL + registration.RouteSubmitFlow + "?flow=" + flowID.String()
	}

	loginAction := func(flowID uuid.UUID) string {
		return ts.URL + login.RouteSubmitFlow + "?flow=" + flowID.String()
	}

	makeRequestWithCookieJar := func(t *testing.T, provider string, action string, fv url.Values, jar *cookiejar.Jar) (*http.Response, []byte) {
		fv.Set("provider", provider)
		res, err := testhelpers.NewClientWithCookieJar(t, jar, false).PostForm(action, fv)
		require.NoError(t, err, action)

		body, err := io.ReadAll(res.Body)
		require.NoError(t, res.Body.Close())
		require.NoError(t, err)

		require.Equal(t, 200, res.StatusCode, "%s: %s\n\t%s", action, res.Request.URL.String(), body)

		return res, body
	}

	makeRequest := func(t *testing.T, provider string, action string, fv url.Values) (*http.Response, []byte) {
		return makeRequestWithCookieJar(t, provider, action, fv, nil)
	}

	makeJSONRequest := func(t *testing.T, provider string, action string, fv url.Values) (*http.Response, []byte) {
		fv.Set("provider", provider)
		client := testhelpers.NewClientWithCookieJar(t, nil, false)
		req, err := http.NewRequest("POST", action, strings.NewReader(fv.Encode()))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")
		res, err := client.Do(req)
		require.NoError(t, err, action)

		body, err := io.ReadAll(res.Body)
		require.NoError(t, res.Body.Close())
		require.NoError(t, err)

		require.Equal(t, 422, res.StatusCode, "%s: %s\n\t%s", action, res.Request.URL.String(), body)

		return res, body
	}

	makeAPICodeFlowRequest := func(t *testing.T, provider, action string) (returnToCode string) {
		res, err := testhelpers.NewDebugClient(t).Post(action, "application/json", strings.NewReader(fmt.Sprintf(`{
	"method": "oidc",
	"provider": %q
}`, provider)))
		require.NoError(t, err)
		require.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
		var changeLocation flow.BrowserLocationChangeRequiredError
		require.NoError(t, json.NewDecoder(res.Body).Decode(&changeLocation))

		res, err = testhelpers.NewClientWithCookieJar(t, nil, true).Get(changeLocation.RedirectBrowserTo)
		require.NoError(t, err)

		returnToURL := res.Request.URL
		assert.True(t, strings.HasPrefix(returnToURL.String(), returnTS.URL+"/app_code"))

		code := returnToURL.Query().Get("code")
		assert.NotEmpty(t, code, "code query param was empty in the return_to URL")

		return code
	}

	exchangeCodeForToken := func(t *testing.T, codes sessiontokenexchange.Codes) (codeResponse session.CodeExchangeResponse, err error) {
		tokenURL := urlx.ParseOrPanic(ts.URL)
		tokenURL.Path = "/sessions/token-exchange"
		tokenURL.RawQuery = fmt.Sprintf("init_code=%s&return_to_code=%s", codes.InitCode, codes.ReturnToCode)
		res, err := ts.Client().Get(tokenURL.String())
		if err != nil {
			return codeResponse, err
		}
		if res.StatusCode != 200 {
			return codeResponse, fmt.Errorf("got status code %d", res.StatusCode)
		}
		require.NoError(t, json.NewDecoder(res.Body).Decode(&codeResponse))

		return
	}

	assertSystemErrorWithReason := func(t *testing.T, res *http.Response, body []byte, code int, reason string) {
		require.Contains(t, res.Request.URL.String(), errTS.URL, "%s", body)

		assert.Equal(t, int64(code), gjson.GetBytes(body, "code").Int(), "%s", prettyJSON(t, body))
		assert.Contains(t, gjson.GetBytes(body, "reason").String(), reason, "%s", prettyJSON(t, body))
	}

	// assert system error (redirect to error endpoint)
	assertSystemErrorWithMessage := func(t *testing.T, res *http.Response, body []byte, code int, message string) {
		require.Contains(t, res.Request.URL.String(), errTS.URL, "%s", body)

		assert.Equal(t, int64(code), gjson.GetBytes(body, "code").Int(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "message").String(), message, "%s", body)
	}

	// assert ui error (redirect to login/registration ui endpoint)
	assertUIError := func(t *testing.T, res *http.Response, body []byte, reason string) {
		require.Contains(t, res.Request.URL.String(), uiTS.URL, "status: %d, body: %s", res.StatusCode, body)
		assert.Contains(t, gjson.GetBytes(body, "ui.messages.0.text").String(), reason, "%s", prettyJSON(t, body))
	}

	// assert identity (success)
	assertIdentity := func(t *testing.T, res *http.Response, body []byte) {
		assert.Contains(t, res.Request.URL.String(), returnTS.URL, "%s", body)
		assert.Equal(t, subject, gjson.GetBytes(body, "identity.traits.subject").String(), "%s", prettyJSON(t, body))
		assert.Equal(t, claims.traits.website, gjson.GetBytes(body, "identity.traits.website").String(), "%s", prettyJSON(t, body))
		assert.Equal(t, claims.metadataPublic.picture, gjson.GetBytes(body, "identity.metadata_public.picture").String(), "%s", prettyJSON(t, body))
	}

	newLoginFlow := func(t *testing.T, requestURL string, exp time.Duration, flowType flow.Type) (req *login.Flow) {
		// Use NewLoginFlow to instantiate the request but change the things we need to control a copy of it.
		req, _, err := reg.LoginHandler().NewLoginFlow(httptest.NewRecorder(),
			&http.Request{URL: urlx.ParseOrPanic(requestURL)}, flowType)
		require.NoError(t, err)
		req.RequestURL = requestURL
		req.ExpiresAt = time.Now().Add(exp)
		require.NoError(t, reg.LoginFlowPersister().UpdateLoginFlow(context.Background(), req))

		// sanity check
		got, err := reg.LoginFlowPersister().GetLoginFlow(context.Background(), req.ID)
		require.NoError(t, err)

		require.Len(t, got.UI.Nodes, len(req.UI.Nodes), "%+v", got)

		return
	}
	newBrowserLoginFlow := func(t *testing.T, redirectTo string, exp time.Duration) (req *login.Flow) {
		return newLoginFlow(t, redirectTo, exp, flow.TypeBrowser)
	}
	newAPILoginFlow := func(t *testing.T, redirectTo string, exp time.Duration) (req *login.Flow) {
		return newLoginFlow(t, redirectTo, exp, flow.TypeAPI)
	}

	newRegistrationFlow := func(t *testing.T, redirectTo string, exp time.Duration, flowType flow.Type) *registration.Flow {
		// Use NewLoginFlow to instantiate the request but change the things we need to control a copy of it.
		req, err := reg.RegistrationHandler().NewRegistrationFlow(httptest.NewRecorder(),
			&http.Request{URL: urlx.ParseOrPanic(redirectTo)}, flowType)
		require.NoError(t, err)
		req.RequestURL = redirectTo
		req.ExpiresAt = time.Now().Add(exp)
		require.NoError(t, reg.RegistrationFlowPersister().UpdateRegistrationFlow(context.Background(), req))

		// sanity check
		got, err := reg.RegistrationFlowPersister().GetRegistrationFlow(context.Background(), req.ID)
		require.NoError(t, err)
		require.Len(t, got.UI.Nodes, len(req.UI.Nodes), "%+v", req)

		return req
	}
	newBrowserRegistrationFlow := func(t *testing.T, redirectTo string, exp time.Duration) *registration.Flow {
		return newRegistrationFlow(t, redirectTo, exp, flow.TypeBrowser)
	}
	newAPIRegistrationFlow := func(t *testing.T, redirectTo string, exp time.Duration) *registration.Flow {
		return newRegistrationFlow(t, redirectTo, exp, flow.TypeAPI)
	}

	t.Run("case=should fail because provider does not exist", func(t *testing.T) {
		for k, v := range []string{
			loginAction(newBrowserLoginFlow(t, returnTS.URL, time.Minute).ID),
			registerAction(newBrowserRegistrationFlow(t, returnTS.URL, time.Minute).ID),
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				res, body := makeRequest(t, "provider-does-not-exist", v, url.Values{})
				assertSystemErrorWithReason(t, res, body, http.StatusNotFound, "is unknown or has not been configured")
			})
		}
	})

	t.Run("case=should fail because the issuer is mismatching", func(t *testing.T) {
		scope = []string{"openid"}
		for k, v := range []string{
			loginAction(newBrowserLoginFlow(t, returnTS.URL, time.Minute).ID),
			registerAction(newBrowserRegistrationFlow(t, returnTS.URL, time.Minute).ID),
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				res, body := makeRequest(t, "invalid-issuer", v, url.Values{})
				assertSystemErrorWithReason(t, res, body, http.StatusInternalServerError, "issuer did not match the issuer returned by provider")
			})
		}
	})

	t.Run("case=should fail because flow does not exist", func(t *testing.T) {
		for k, v := range []string{loginAction(x.NewUUID()), registerAction(x.NewUUID())} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				res, body := makeRequest(t, "valid", v, url.Values{})
				assertSystemErrorWithMessage(t, res, body, http.StatusNotFound, "Unable to locate the resource")
			})
		}
	})

	t.Run("case=should fail because the flow is expired", func(t *testing.T) {
		for k, v := range []uuid.UUID{
			newBrowserLoginFlow(t, returnTS.URL, -time.Minute).ID,
			newBrowserRegistrationFlow(t, returnTS.URL, -time.Minute).ID,
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				action := assertFormValues(t, v, "valid")
				res, body := makeRequest(t, "valid", action, url.Values{})

				assert.NotEqual(t, v, gjson.GetBytes(body, "id"))
				require.Contains(t, res.Request.URL.String(), uiTS.URL, "%s", body)
				assert.Contains(t, gjson.GetBytes(body, "ui.messages.0.text").String(), "flow expired", "%s", body)
			})
		}
	})

	t.Run("case=should fail registration because scope was not provided", func(t *testing.T) {
		subject = "foo@bar.com"
		scope = []string{}

		for k, v := range []uuid.UUID{
			newBrowserLoginFlow(t, returnTS.URL, time.Minute).ID,
			newBrowserRegistrationFlow(t, returnTS.URL, time.Minute).ID,
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				action := assertFormValues(t, v, "valid")
				res, body := makeRequest(t, "valid", action, url.Values{})
				assertUIError(t, res, body, "no id_token was returned")
			})
		}
	})

	t.Run("case=should fail because password can not handle AAL2", func(t *testing.T) {
		testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/registration-aal.schema.json")
		t.Cleanup(func() {
			testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/registration.schema.json")
		})
		bc := testhelpers.NewDebugClient(t)
		f := testhelpers.InitializeLoginFlowViaAPI(t, bc, ts, false)

		update, err := reg.LoginFlowPersister().GetLoginFlow(context.Background(), uuid.FromStringOrNil(f.Id))
		require.NoError(t, err)
		update.RequestedAAL = identity.AuthenticatorAssuranceLevel2
		require.NoError(t, reg.LoginFlowPersister().UpdateLoginFlow(context.Background(), update))

		req, err := http.NewRequest("POST", f.Ui.Action, bytes.NewBufferString(`{"method":"oidc"}`))
		require.NoError(t, err)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		actual, res := testhelpers.MockMakeAuthenticatedRequest(t, reg, conf, routerP.Router, req)
		assert.Contains(t, res.Request.URL.String(), ts.URL+login.RouteSubmitFlow)
		assert.Equal(t, text.NewErrorValidationLoginNoStrategyFound().Text, gjson.GetBytes(actual, "ui.messages.0.text").String())
	})

	t.Run("case=should fail login because scope was not provided", func(t *testing.T) {
		r := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
		action := assertFormValues(t, r.ID, "valid")
		res, body := makeRequest(t, "valid", action, url.Values{})
		assertUIError(t, res, body, "no id_token was returned")
	})

	t.Run("case=should fail registration flow because subject is not an email", func(t *testing.T) {
		subject = "not-an-email"
		scope = []string{"openid"}

		r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
		action := assertFormValues(t, r.ID, "valid")
		res, body := makeRequest(t, "valid", action, url.Values{})

		require.Contains(t, res.Request.URL.String(), uiTS.URL, "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==traits.subject).messages.0.text").String(), "is not valid", "%s\n%s", gjson.GetBytes(body, "ui.nodes.#(attributes.name==traits.subject)").Raw, body)
	})

	t.Run("case=cannot register multiple accounts with the same OIDC account", func(t *testing.T) {
		subject = "oidc-register-then-login@ory.sh"
		scope = []string{"openid", "offline"}

		expectTokens := func(t *testing.T, provider string, body []byte) {
			i, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), uuid.FromStringOrNil(gjson.GetBytes(body, "identity.id").String()))
			require.NoError(t, err)
			c := i.Credentials[identity.CredentialsTypeOIDC].Config
			assert.NotEmpty(t, gjson.GetBytes(c, "providers.0.initial_access_token").String())
			assertx.EqualAsJSONExcept(
				t,
				json.RawMessage(fmt.Sprintf(`{"providers": [{"subject":"%s","provider":"%s"}]}`, subject, provider)),
				json.RawMessage(c),
				[]string{"providers.0.initial_id_token", "providers.0.initial_access_token", "providers.0.initial_refresh_token"},
			)
		}

		t.Run("case=should pass registration", func(t *testing.T) {
			r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			res, body := makeRequest(t, "valid", action, url.Values{})
			assertIdentity(t, res, body)
			expectTokens(t, "valid", body)
		})

		t.Run("case=try another registration", func(t *testing.T) {
			returnTo := fmt.Sprintf("%s/home?query=true", returnTS.URL)
			r := newBrowserRegistrationFlow(t, fmt.Sprintf("%s?return_to=%s", returnTS.URL, url.QueryEscape(returnTo)), time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			res, body := makeRequest(t, "valid", action, url.Values{})
			assert.Equal(t, returnTo, res.Request.URL.String())
			assertIdentity(t, res, body)
			expectTokens(t, "valid", body)
		})
	})

	expectTokens := func(t *testing.T, provider string, body []byte) uuid.UUID {
		id := uuid.FromStringOrNil(gjson.GetBytes(body, "identity.id").String())
		i, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), id)
		require.NoError(t, err)
		c := i.Credentials[identity.CredentialsTypeOIDC].Config
		assert.NotEmpty(t, gjson.GetBytes(c, "providers.0.initial_access_token").String())
		assertx.EqualAsJSONExcept(
			t,
			json.RawMessage(fmt.Sprintf(`{"providers": [{"subject":"%s","provider":"%s"}]}`, subject, provider)),
			json.RawMessage(c),
			[]string{"providers.0.initial_id_token", "providers.0.initial_access_token", "providers.0.initial_refresh_token"},
		)
		return id
	}

	t.Run("case=register and then login", func(t *testing.T) {
		postRegistrationWebhook := hooktest.NewServer()
		t.Cleanup(postRegistrationWebhook.Close)
		postLoginWebhook := hooktest.NewServer()
		t.Cleanup(postLoginWebhook.Close)

		postRegistrationWebhook.SetConfig(t, conf.GetProvider(ctx),
			config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypeOIDC.String()))
		postLoginWebhook.SetConfig(t, conf.GetProvider(ctx),
			config.HookStrategyKey(config.ViperKeySelfServiceLoginAfter, config.HookGlobal))

		subject = "register-then-login@ory.sh"
		scope = []string{"openid", "offline"}

		t.Run("case=should pass registration", func(t *testing.T) {
			transientPayload := `{"data": "registration"}`
			r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			res, body := makeRequest(t, "valid", action, url.Values{
				"transient_payload": {transientPayload},
			})
			assertIdentity(t, res, body)
			expectTokens(t, "valid", body)
			assert.Equal(t, "valid", gjson.GetBytes(body, "authentication_methods.0.provider").String(), "%s", body)

			postRegistrationWebhook.AssertTransientPayload(t, transientPayload)
		})

		t.Run("case=should pass login", func(t *testing.T) {
			transientPayload := `{"data": "login"}`
			r := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			res, body := makeRequest(t, "valid", action, url.Values{
				"transient_payload": {transientPayload},
			})
			assertIdentity(t, res, body)
			expectTokens(t, "valid", body)
			assert.Equal(t, "valid", gjson.GetBytes(body, "authentication_methods.0.provider").String(), "%s", body)

			postLoginWebhook.AssertTransientPayload(t, transientPayload)
		})
	})

	t.Run("case=login without registered account", func(t *testing.T) {
		postRegistrationWebhook := hooktest.NewServer()
		t.Cleanup(postRegistrationWebhook.Close)
		postLoginWebhook := hooktest.NewServer()
		t.Cleanup(postLoginWebhook.Close)

		postRegistrationWebhook.SetConfig(t, conf.GetProvider(ctx),
			config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypeOIDC.String()))
		postLoginWebhook.SetConfig(t, conf.GetProvider(ctx),
			config.HookStrategyKey(config.ViperKeySelfServiceLoginAfter, config.HookGlobal))

		subject = "login-without-register@ory.sh"
		scope = []string{"openid"}

		t.Run("case=should pass login", func(t *testing.T) {
			transientPayload := `{"data": "login to registration"}`

			r := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			res, body := makeRequest(t, "valid", action, url.Values{
				"transient_payload": {transientPayload},
			})
			assertIdentity(t, res, body)
			assert.Equal(t, "valid", gjson.GetBytes(body, "authentication_methods.0.provider").String(), "%s", body)

			assert.Empty(t, postLoginWebhook.LastBody,
				"post login hook should not have been called, because this was a registration flow")
			postRegistrationWebhook.AssertTransientPayload(t, transientPayload)
		})
	})

	t.Run("case=login with Browser+JSON", func(t *testing.T) {
		subject = "login-with-browser-json@ory.sh"
		scope = []string{"openid"}

		t.Run("case=should pass login", func(t *testing.T) {
			r := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			res, body := makeJSONRequest(t, "valid", action, url.Values{})

			assert.Equal(t, "browser_location_change_required", gjson.GetBytes(body, "error.id").String(), "%s", body)

			continuityCookie := res.Header.Get("Set-Cookie")
			key, val, ok := strings.Cut(continuityCookie, "=")
			require.True(t, ok)
			assert.Equal(t, "ory_kratos_continuity", key)
			assert.NotEmpty(t, val)
		})
	})

	t.Run("suite=API with session token exchange code", func(t *testing.T) {
		scope = []string{"openid"}

		loginOrRegister := func(t *testing.T, id uuid.UUID, code string) {
			_, err := exchangeCodeForToken(t, sessiontokenexchange.Codes{InitCode: code})
			require.Error(t, err)

			action := assertFormValues(t, id, "valid")
			returnToCode := makeAPICodeFlowRequest(t, "valid", action)
			codeResponse, err := exchangeCodeForToken(t, sessiontokenexchange.Codes{
				InitCode:     code,
				ReturnToCode: returnToCode,
			})
			require.NoError(t, err)

			assert.NotEmpty(t, codeResponse.Token)
			assert.Equal(t, subject, gjson.GetBytes(codeResponse.Session.Identity.Traits, "subject").String())
		}
		register := func(t *testing.T) {
			f := newAPIRegistrationFlow(t, returnTS.URL+"?return_session_token_exchange_code=true&return_to=/app_code", 1*time.Minute)
			loginOrRegister(t, f.ID, f.SessionTokenExchangeCode)
		}
		login := func(t *testing.T) {
			f := newAPILoginFlow(t, returnTS.URL+"?return_session_token_exchange_code=true&return_to=/app_code", 1*time.Minute)
			loginOrRegister(t, f.ID, f.SessionTokenExchangeCode)
		}

		for _, tc := range []struct {
			name        string
			first, then func(*testing.T)
		}{{
			name:  "login-twice",
			first: login, then: login,
		}, {
			name:  "login-then-register",
			first: login, then: register,
		}, {
			name:  "register-then-login",
			first: register, then: login,
		}, {
			name:  "register-twice",
			first: register, then: register,
		}} {
			t.Run("case="+tc.name, func(t *testing.T) {
				subject = tc.name + "-api-code-testing@ory.sh"
				tc.first(t)
				tc.then(t)
			})
		}
	})

	t.Run("case=submit id_token during registration or login", func(t *testing.T) {
		viperSetProviderConfig(
			t,
			conf,
			newOIDCProvider(t, ts, remotePublic, remoteAdmin, "valid"),
			oidc.Configuration{
				Provider:     "test-provider",
				ID:           "test-provider",
				ClientID:     invalid.ClientID,
				ClientSecret: invalid.ClientSecret,
				IssuerURL:    remotePublic + "/",
				Mapper:       "file://./stub/oidc.facebook.jsonnet",
			},
		)
		t.Cleanup(oidc.RegisterTestProvider("test-provider"))

		cl := http.Client{}

		type testCase struct {
			name     string
			idToken  string
			provider string
			v        func(string, string, string) url.Values
			expect   func(t *testing.T, res *http.Response, body []byte)
		}

		prep := func(tc *testCase) (provider string, token string, nonce string) {
			provider = tc.provider
			if provider == "" {
				provider = "test-provider"
			}
			token = tc.idToken
			token = strings.Replace(token, "{{sub}}", testhelpers.RandomEmail(), -1)
			nonce = randx.MustString(16, randx.Alpha)
			token = strings.Replace(token, "{{nonce}}", nonce, -1)
			return
		}

		for _, tc := range []testCase{
			{
				name:     "should fail if provider does not support id_token submission",
				idToken:  "error",
				provider: "valid",
				expect: func(t *testing.T, res *http.Response, body []byte) {
					require.Equal(t, "The provider generic does not support id_token verification", gjson.GetBytes(body, "error.reason").String(), "%s", body)
				},
			},
			{
				name:    "should fail because id_token is invalid",
				idToken: "error",
				expect: func(t *testing.T, res *http.Response, body []byte) {
					require.Equal(t, "Could not verify id_token", gjson.GetBytes(body, "error.reason").String(), "%s", body)
					require.Equal(t, "stub error", gjson.GetBytes(body, "error.message").String(), "%s", body)
				},
			},
			{
				name:    "should fail because claims are invalid",
				idToken: "{}",
				expect: func(t *testing.T, res *http.Response, body []byte) {
					require.Equal(t, "The id_token claims were invalid", gjson.GetBytes(body, "error.reason").String(), "%s", body)
				},
			},
			{
				name: "should fail if no nonce is included in the id_token",
				idToken: `{
					"iss": "https://appleid.apple.com",
					"sub": "{{sub}}"
				}`,
				expect: func(t *testing.T, res *http.Response, body []byte) {
					require.Equal(t, "No nonce was included in the id_token but is required by the provider", gjson.GetBytes(body, "error.reason").String(), "%s", body)
				},
			},
			{
				name: "should fail if no nonce is supplied in request",
				idToken: `{
					"iss": "https://appleid.apple.com",
					"sub": "{{sub}}",
					"nonce": "{{nonce}}"
				}`,
				v: func(provider, token, _ string) url.Values {
					return url.Values{
						"id_token": {token},
						"provider": {provider},
					}
				},
				expect: func(t *testing.T, res *http.Response, body []byte) {
					require.Equal(t, "No nonce was provided but is required by the provider", gjson.GetBytes(body, "error.reason").String(), "%s", body)
				},
			},
			{
				name: "should pass if claims are valid",
				idToken: `{
					"iss": "https://appleid.apple.com",
					"sub": "{{sub}}",
					"nonce": "{{nonce}}"
				}`,
				expect: func(t *testing.T, res *http.Response, body []byte) {
					require.NotEmpty(t, gjson.GetBytes(body, "session_token").String(), "%s", body)
				},
			},
			{
				name: "nonce mismatch",
				idToken: `{
					"iss": "https://appleid.apple.com",
					"sub": "{{sub}}",
					"nonce": "random-nonce"
				}`,
				expect: func(t *testing.T, res *http.Response, body []byte) {
					require.Equal(t, "The supplied nonce does not match the nonce from the id_token", gjson.GetBytes(body, "error.reason").String(), "%s", body)
				},
			},
		} {
			tc := tc
			t.Run(fmt.Sprintf("flow=registration/case=%s", tc.name), func(t *testing.T) {
				f := newAPIRegistrationFlow(t, returnTS.URL, time.Minute)
				provider, token, nonce := prep(&tc)
				action := assertFormValues(t, f.ID, "test-provider")
				v := url.Values{
					"id_token":       {token},
					"provider":       {provider},
					"id_token_nonce": {nonce},
				}
				if tc.v != nil {
					v = tc.v(provider, token, nonce)
				}
				res, err := cl.PostForm(action, v)
				require.NoError(t, err)
				body := ioutilx.MustReadAll(res.Body)
				tc.expect(t, res, body)
			})

			t.Run(fmt.Sprintf("flow=login/case=%s", tc.name), func(t *testing.T) {
				provider, token, nonce := prep(&tc)
				rf := newAPIRegistrationFlow(t, returnTS.URL, time.Minute)
				action := assertFormValues(t, rf.ID, "test-provider")
				v := url.Values{
					"id_token":       {token},
					"provider":       {provider},
					"id_token_nonce": {nonce},
				}
				if tc.v != nil {
					v = tc.v(provider, token, nonce)
				}
				_, err := cl.PostForm(action, v)
				require.NoError(t, err)

				lf := newAPILoginFlow(t, returnTS.URL, time.Minute)
				action = assertFormValues(t, lf.ID, "test-provider")

				res, err := cl.PostForm(action, v)
				require.NoError(t, err)
				body := ioutilx.MustReadAll(res.Body)
				tc.expect(t, res, body)
			})

			t.Run(fmt.Sprintf("flow=login_without_registration/case=%s", tc.name), func(t *testing.T) {
				provider, token, nonce := prep(&tc)
				rf := newAPIRegistrationFlow(t, returnTS.URL, time.Minute)
				action := assertFormValues(t, rf.ID, "test-provider")

				v := url.Values{
					"id_token":       {token},
					"provider":       {provider},
					"id_token_nonce": {nonce},
				}
				if tc.v != nil {
					v = tc.v(provider, token, nonce)
				}
				_, err := cl.PostForm(action, v)
				require.NoError(t, err)

				lf := newAPIRegistrationFlow(t, returnTS.URL, time.Minute)
				action = assertFormValues(t, lf.ID, "test-provider")

				res, err := cl.PostForm(action, v)
				require.NoError(t, err)
				body := ioutilx.MustReadAll(res.Body)
				tc.expect(t, res, body)
			})

			t.Run(fmt.Sprintf("flow=login_with_return_session_token_exchange_code/case=%s", tc.name), func(t *testing.T) {
				provider, token, nonce := prep(&tc)
				lf := newAPILoginFlow(t, returnTS.URL+"?return_session_token_exchange_code=true&return_to=/app_code", time.Minute)
				action := assertFormValues(t, lf.ID, "test-provider")
				v := url.Values{
					"id_token":       {token},
					"provider":       {provider},
					"id_token_nonce": {nonce},
				}
				if tc.v != nil {
					v = tc.v(provider, token, nonce)
				}
				res, err := cl.PostForm(action, v)
				require.NoError(t, err)
				body := ioutilx.MustReadAll(res.Body)
				tc.expect(t, res, body)
			})

			t.Run(fmt.Sprintf("flow=registration_with_return_session_token_exchange_code/case=%s", tc.name), func(t *testing.T) {
				provider, token, nonce := prep(&tc)
				lf := newAPIRegistrationFlow(t, returnTS.URL+"?return_session_token_exchange_code=true&return_to=/app_code", time.Minute)
				action := assertFormValues(t, lf.ID, "test-provider")
				v := url.Values{
					"id_token":       {token},
					"provider":       {provider},
					"id_token_nonce": {nonce},
				}
				if tc.v != nil {
					v = tc.v(provider, token, nonce)
				}
				res, err := cl.PostForm(action, v)
				require.NoError(t, err)
				body := ioutilx.MustReadAll(res.Body)
				tc.expect(t, res, body)
			})

		}
	})

	t.Run("case=login without registered account with return_to", func(t *testing.T) {
		t.Run("case=should pass login", func(t *testing.T) {
			subject = "login-without-register-return-to@ory.sh"
			scope = []string{"openid"}
			returnTo := "/foo"
			r := newBrowserLoginFlow(t, fmt.Sprintf("%s?return_to=%s", returnTS.URL, returnTo), time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			res, body := makeRequest(t, "valid", action, url.Values{})
			assert.True(t, strings.HasSuffix(res.Request.URL.String(), returnTo))
			assertIdentity(t, res, body)
		})

		t.Run("case=should pass login and carry over login_challenge to registration", func(t *testing.T) {
			subject = "login_challenge_carry_over@ory.sh"
			scope = []string{"openid"}
			conf.MustSet(ctx, config.ViperKeyOAuth2ProviderURL, "http://fake-hydra")

			reg.WithHydra(hydra.NewFake())
			r := newBrowserLoginFlow(t, fmt.Sprintf("%s?login_challenge=%s", returnTS.URL, hydra.FakeValidLoginChallenge), time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			fv := url.Values{}
			fv.Set("provider", "valid")
			res, err := testhelpers.NewClientWithCookieJar(t, nil, false).PostForm(action, fv)
			require.NoError(t, err)
			// Expect to be returned to the hydra instance, that instantiated the request
			assert.Equal(t, hydra.FakePostLoginURL, res.Request.URL.String())
		})
	})

	t.Run("case=register and register again but login", func(t *testing.T) {
		subject = "register-twice@ory.sh"
		scope = []string{"openid"}

		t.Run("case=should pass registration", func(t *testing.T) {
			r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			res, body := makeRequest(t, "valid", action, url.Values{})
			assertIdentity(t, res, body)
		})

		t.Run("case=should pass second time registration", func(t *testing.T) {
			r := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			res, body := makeRequest(t, "valid", action, url.Values{})
			assertIdentity(t, res, body)
		})

		t.Run("case=should pass third time registration with return to", func(t *testing.T) {
			returnTo := "/foo"
			r := newBrowserLoginFlow(t, fmt.Sprintf("%s?return_to=%s", returnTS.URL, returnTo), time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			res, body := makeRequest(t, "valid", action, url.Values{})
			assert.True(t, strings.HasSuffix(res.Request.URL.String(), returnTo))
			assertIdentity(t, res, body)
		})
	})

	t.Run("case=register, merge, and complete data", func(t *testing.T) {

		for _, tc := range []struct{ name, provider string }{
			{name: "idtoken", provider: "valid"},
			{name: "userinfo", provider: "claimsViaUserInfo"},
		} {
			subject = fmt.Sprintf("incomplete-data@%s.ory.sh", tc.name)
			scope = []string{"openid"}
			claims = idTokenClaims{}
			claims.traits.website = "https://www.ory.sh/kratos"
			claims.traits.groups = []string{"group1", "group2"}
			claims.metadataPublic.picture = "picture.png"
			claims.metadataAdmin.phoneNumber = "911"

			t.Run(fmt.Sprintf("ClaimsSource=%s", tc.name), func(t *testing.T) {
				t.Run("case=should fail registration on first attempt", func(t *testing.T) {
					r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
					action := assertFormValues(t, r.ID, tc.provider)
					res, body := makeRequest(t, tc.provider, action, url.Values{"traits.name": {"i"}})
					require.Contains(t, res.Request.URL.String(), uiTS.URL, "%s", body)

					assert.Equal(t, "length must be >= 2, but got 1", gjson.GetBytes(body, "ui.nodes.#(attributes.name==traits.name).messages.0.text").String(), "%s", body) // make sure the field is being echoed
					assert.Equal(t, "traits.name", gjson.GetBytes(body, "ui.nodes.#(attributes.name==traits.name).attributes.name").String(), "%s", body)                    // make sure the field is being echoed
					assert.Equal(t, "i", gjson.GetBytes(body, "ui.nodes.#(attributes.name==traits.name).attributes.value").String(), "%s", body)                             // make sure the field is being echoed
					assert.Equal(t, "https://www.ory.sh/kratos", gjson.GetBytes(body, "ui.nodes.#(attributes.name==traits.website).attributes.value").String(), "%s", body)  // make sure the field is being echoed
				})

				t.Run("case=should pass registration with valid data", func(t *testing.T) {
					r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
					action := assertFormValues(t, r.ID, tc.provider)
					res, body := makeRequest(t, tc.provider, action, url.Values{"traits.name": {"valid-name"}})
					assertIdentity(t, res, body)
					assert.Equal(t, "https://www.ory.sh/kratos", gjson.GetBytes(body, "identity.traits.website").String(), "%s", body)
					assert.Equal(t, "valid-name", gjson.GetBytes(body, "identity.traits.name").String(), "%s", body)
					assert.Equal(t, "[\"group1\",\"group2\"]", gjson.GetBytes(body, "identity.traits.groups").String(), "%s", body)
				})
			})
		}

	})

	t.Run("case=should fail to register and return fresh login flow if email is already being used by password credentials", func(t *testing.T) {
		subject = "email-exist-with-password-strategy@ory.sh"
		scope = []string{"openid"}

		t.Run("case=create password identity", func(t *testing.T) {
			i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			i.SetCredentials(identity.CredentialsTypePassword, identity.Credentials{
				Identifiers: []string{subject},
			})
			i.Traits = identity.Traits(`{"subject":"` + subject + `"}`)

			require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))
		})

		t.Run("case=should fail registration", func(t *testing.T) {
			r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			res, body := makeRequest(t, "valid", action, url.Values{})
			assertUIError(t, res, body, "An account with the same identifier (email, phone, username, ...) exists already.")
			require.Contains(t, gjson.GetBytes(body, "ui.action").String(), "/self-service/login")
		})

		t.Run("case=should fail login", func(t *testing.T) {
			r := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			res, body := makeRequest(t, "valid", action, url.Values{})
			assertUIError(t, res, body, "An account with the same identifier (email, phone, username, ...) exists already.")
		})
	})

	t.Run("case=should redirect to default return ts when sending authenticated login flow without forced flag", func(t *testing.T) {
		subject = "no-reauth-login@ory.sh"
		scope = []string{"openid"}

		fv := url.Values{"traits.name": {"valid-name"}}
		jar, _ := cookiejar.New(nil)
		r1 := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
		res1, body1 := makeRequestWithCookieJar(t, "valid", assertFormValues(t, r1.ID, "valid"), fv, jar)
		assertIdentity(t, res1, body1)
		r2 := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
		res2, body2 := makeRequestWithCookieJar(t, "valid", assertFormValues(t, r2.ID, "valid"), fv, jar)
		assertIdentity(t, res2, body2)
		assert.Equal(t, body1, body2)
	})

	t.Run("case=should reauthenticate when sending authenticated login flow with forced flag", func(t *testing.T) {
		subject = "reauth-login@ory.sh"
		scope = []string{"openid"}

		fv := url.Values{"traits.name": {"valid-name"}}
		jar, _ := cookiejar.New(nil)
		r1 := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
		res1, body1 := makeRequestWithCookieJar(t, "valid", assertFormValues(t, r1.ID, "valid"), fv, jar)
		assertIdentity(t, res1, body1)
		r2 := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
		require.NoError(t, reg.LoginFlowPersister().ForceLoginFlow(context.Background(), r2.ID))
		res2, body2 := makeRequestWithCookieJar(t, "valid", assertFormValues(t, r2.ID, "valid"), fv, jar)
		assertIdentity(t, res2, body2)
		assert.NotEqual(t, gjson.GetBytes(body1, "id"), gjson.GetBytes(body2, "id"))
		authAt1, err := time.Parse(time.RFC3339, gjson.GetBytes(body1, "authenticated_at").String())
		require.NoError(t, err)
		authAt2, err := time.Parse(time.RFC3339, gjson.GetBytes(body2, "authenticated_at").String())
		require.NoError(t, err)
		// authenticated at is newer in the second body
		assert.Greater(t, authAt2.Sub(authAt1).Milliseconds(), int64(0), "%s - %s : %s - %s", authAt2, authAt1, body2, body1)
	})

	t.Run("case=upstream parameters should be passed on to provider", func(t *testing.T) {
		subject = "oidc-upstream-parameters@ory.sh"
		scope = []string{"openid", "offline"}

		// We need to disable redirects because the upstream parameters are only passed on to the provider
		c := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		t.Run("case=should pass when registering", func(t *testing.T) {
			f := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
			action := assertFormValues(t, f.ID, "valid")

			fv := url.Values{}

			fv.Set("provider", "valid")
			fv.Set("upstream_parameters.login_hint", "oidc-upstream-parameters@ory.sh")
			fv.Set("upstream_parameters.hd", "ory.sh")
			fv.Set("upstream_parameters.prompt", "select_account")
			fv.Set("upstream_parameters.auth_type", "reauthenticate")

			res, err := c.PostForm(action, fv)
			require.NoError(t, err)
			require.Equal(t, http.StatusSeeOther, res.StatusCode)

			loc, err := res.Location()
			require.NoError(t, err)

			require.Equal(t, "oidc-upstream-parameters@ory.sh", loc.Query().Get("login_hint"))
			require.Equal(t, "ory.sh", loc.Query().Get("hd"))
			require.Equal(t, "select_account", loc.Query().Get("prompt"))
			require.Equal(t, "reauthenticate", loc.Query().Get("auth_type"))
		})

		t.Run("case=should pass when logging in", func(t *testing.T) {
			f := newBrowserLoginFlow(t, returnTS.URL, time.Minute)

			action := assertFormValues(t, f.ID, "valid")

			fv := url.Values{}

			fv.Set("provider", "valid")
			fv.Set("upstream_parameters.login_hint", "oidc-upstream-parameters@ory.sh")
			fv.Set("upstream_parameters.hd", "ory.sh")
			fv.Set("upstream_parameters.prompt", "select_account")

			res, err := c.PostForm(action, fv)
			require.NoError(t, err)
			require.Equal(t, http.StatusSeeOther, res.StatusCode)

			loc, err := res.Location()
			require.NoError(t, err)

			require.Equal(t, "oidc-upstream-parameters@ory.sh", loc.Query().Get("login_hint"))
			require.Equal(t, "ory.sh", loc.Query().Get("hd"))
			require.Equal(t, "select_account", loc.Query().Get("prompt"))
		})

		t.Run("case=should ignore invalid parameters when logging in", func(t *testing.T) {
			f := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
			action := assertFormValues(t, f.ID, "valid")

			fv := url.Values{}
			fv.Set("upstream_parameters.lol", "invalid")

			res, err := c.PostForm(action, fv)
			require.NoError(t, err)
			require.Equal(t, http.StatusSeeOther, res.StatusCode)

			loc, err := res.Location()
			require.NoError(t, err)

			// upstream parameters that are not on the allow list will be ignored and not passed on to the upstream provider.
			require.Empty(t, loc.Query().Get("lol"))
		})

		t.Run("case=should ignore invalid parameters when registering", func(t *testing.T) {
			f := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
			action := assertFormValues(t, f.ID, "valid")

			fv := url.Values{}
			fv.Set("upstream_parameters.lol", "invalid")

			res, err := c.PostForm(action, fv)
			require.NoError(t, err)
			require.Equal(t, http.StatusFound, res.StatusCode)

			loc, err := res.Location()
			require.NoError(t, err)

			// upstream parameters that are not on the allow list will be ignored and not passed on to the upstream provider.
			require.Empty(t, loc.Query().Get("lol"))
		})
	})

	t.Run("case=verified addresses should be respected", func(t *testing.T) {
		scope = []string{"openid"}

		testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/registration-verifiable-email.schema.json")

		assertVerifiedEmail := func(t *testing.T, body []byte, verified bool) {
			assert.Len(t, gjson.GetBytes(body, "identity.verifiable_addresses").Array(), 1, "%s", body)
			assert.Equal(t, "email", gjson.GetBytes(body, "identity.verifiable_addresses.0.via").String(), "%s", body)
			assert.Equal(t, subject, gjson.GetBytes(body, "identity.verifiable_addresses.0.value").String(), "%s", body)
			assert.Equal(t, verified, gjson.GetBytes(body, "identity.verifiable_addresses.0.verified").Bool(), "%s", body)
		}

		t.Run("case=should have verified address when subject matches", func(t *testing.T) {
			subject = "verified-email@ory.sh"
			r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			res, body := makeRequest(t, "valid", action, url.Values{})
			assertIdentity(t, res, body)
			assertVerifiedEmail(t, body, true)
		})

		t.Run("case=should have verified address when subject matches after normalization", func(t *testing.T) {
			subject = " Denormalized-Verified-Email@ory.sh "
			r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			res, body := makeRequest(t, "valid", action, url.Values{"traits.subject": {"denormalized-verified-EMAIL@ory.sh"}})
			subject = "denormalized-verified-EMAIL@ory.sh"
			assertIdentity(t, res, body)
			subject = "denormalized-verified-email@ory.sh"
			assertVerifiedEmail(t, body, true)
		})

		t.Run("case=should have unverified address when subject does not match", func(t *testing.T) {
			subject = "changed-verified-email@ory.sh"
			r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			res, body := makeRequest(t, "valid", action, url.Values{"traits.subject": {"unverified-email@ory.sh"}})
			subject = "unverified-email@ory.sh"
			assertIdentity(t, res, body)
			assertVerifiedEmail(t, body, false)
		})
	})

	t.Run("case=registration should start new login flow if duplicate credentials detected", func(t *testing.T) {
		require.NoError(t, reg.Config().Set(ctx, config.ViperKeySelfServiceRegistrationLoginHints, true))
		loginWithOIDC := func(t *testing.T, c *http.Client, flowID uuid.UUID, provider string) (*http.Response, []byte) {
			action := assertFormValues(t, flowID, provider)
			res, err := c.PostForm(action, url.Values{"provider": {provider}})
			require.NoError(t, err, action)
			body, err := io.ReadAll(res.Body)
			require.NoError(t, res.Body.Close())
			require.NoError(t, err)
			return res, body
		}

		checkCredentialsLinked := func(res *http.Response, body []byte, identityID uuid.UUID, provider string) {
			assert.Contains(t, res.Request.URL.String(), returnTS.URL, "%s", body)
			assert.Equal(t, strings.ToLower(subject), gjson.GetBytes(body, "identity.traits.subject").String(), "%s", body)
			i, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, identityID)
			require.NoError(t, err)
			assert.NotEmpty(t, i.Credentials["oidc"], "%+v", i.Credentials)
			assert.Equal(t, provider, gjson.GetBytes(i.Credentials["oidc"].Config, "providers.0.provider").String(),
				"%s", string(i.Credentials["oidc"].Config[:]))
			assert.Contains(t, gjson.GetBytes(body, "authentication_methods").String(), "oidc", "%s", body)
		}

		t.Run("case=second login is password", func(t *testing.T) {
			subject = "new-login-if-email-exist-with-password-strategy@ory.sh"
			subject2 := "new-login-subject2@ory.sh"
			scope = []string{"openid"}
			password := "lwkj52sdkjf"

			var i *identity.Identity
			t.Run("step=create password identity", func(t *testing.T) {
				i = identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
				p, err := reg.Hasher(ctx).Generate(ctx, []byte(password))
				require.NoError(t, err)
				i.SetCredentials(identity.CredentialsTypePassword, identity.Credentials{
					Identifiers: []string{subject},
					Config:      sqlxx.JSONRawMessage(`{"hashed_password":"` + string(p) + `"}`),
				})
				i.Traits = identity.Traits(`{"subject":"` + subject + `"}`)
				require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))

				i2 := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
				i2.SetCredentials(identity.CredentialsTypePassword, identity.Credentials{
					Identifiers: []string{subject2},
					Config:      sqlxx.JSONRawMessage(`{"hashed_password":"` + string(p) + `"}`),
				})
				i2.Traits = identity.Traits(`{"subject":"` + subject2 + `"}`)
				require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i2))
			})

			client := testhelpers.NewClientWithCookieJar(t, nil, false)
			loginFlow := newLoginFlow(t, returnTS.URL, time.Minute, flow.TypeBrowser)

			var linkingLoginFlow struct {
				ID        string
				UIAction  string
				CSRFToken string
			}

			// To test that the subject is normalized properly
			subject = strings.ToUpper(subject)

			t.Run("step=should fail login and start a new flow", func(t *testing.T) {
				res, body := loginWithOIDC(t, client, loginFlow.ID, "valid")
				assert.True(t, res.Request.URL.Query().Has("no_org_ui"))
				assertUIError(t, res, body, "You tried signing in with new-login-if-email-exist-with-password-strategy@ory.sh which is already in use by another account. You can sign in using your password.")
				assert.Equal(t, "password", gjson.GetBytes(body, "ui.messages.#(id==4000028).context.available_credential_types.0").String())
				assert.Equal(t, "new-login-if-email-exist-with-password-strategy@ory.sh", gjson.GetBytes(body, "ui.messages.#(id==4000028).context.credential_identifier_hint").String())
				linkingLoginFlow.ID = gjson.GetBytes(body, "id").String()
				linkingLoginFlow.UIAction = gjson.GetBytes(body, "ui.action").String()
				linkingLoginFlow.CSRFToken = gjson.GetBytes(body, `ui.nodes.#(attributes.name=="csrf_token").attributes.value`).String()
				assert.NotEqual(t, loginFlow.ID.String(), linkingLoginFlow.ID, "should have started a new flow")
			})

			t.Run("step=should fail login if existing identity identifier doesn't match", func(t *testing.T) {
				res, err := client.PostForm(linkingLoginFlow.UIAction, url.Values{
					"csrf_token": {linkingLoginFlow.CSRFToken},
					"method":     {"password"},
					"identifier": {subject2},
					"password":   {password},
				})
				require.NoError(t, err, linkingLoginFlow.UIAction)
				body, err := io.ReadAll(res.Body)
				require.NoError(t, res.Body.Close())
				require.NoError(t, err)
				assert.Equal(t,
					strconv.Itoa(int(text.ErrorValidationLoginLinkedCredentialsDoNotMatch)),
					gjson.GetBytes(body, "ui.messages.0.id").String(),
					prettyJSON(t, body),
				)
			})

			t.Run("step=should link oidc credentials to existing identity", func(t *testing.T) {
				res, err := client.PostForm(linkingLoginFlow.UIAction, url.Values{
					"csrf_token": {linkingLoginFlow.CSRFToken},
					"method":     {"password"},
					"identifier": {subject},
					"password":   {password},
				})
				require.NoError(t, err, linkingLoginFlow.UIAction)
				body, err := io.ReadAll(res.Body)
				require.NoError(t, res.Body.Close())
				require.NoError(t, err)
				checkCredentialsLinked(res, body, i.ID, "valid")
			})
		})

		t.Run("case=second login is OIDC", func(t *testing.T) {
			email1 := "existing-oidc-identity-1@ory.sh"
			email2 := "existing-oidc-identity-2@ory.sh"
			scope = []string{"openid", "offline"}

			var identityID uuid.UUID
			t.Run("step=create OIDC identity", func(t *testing.T) {
				subject = email1
				r := newRegistrationFlow(t, returnTS.URL, time.Minute, flow.TypeBrowser)
				action := assertFormValues(t, r.ID, "secondProvider")
				res, body := makeRequest(t, "secondProvider", action, url.Values{})
				assertIdentity(t, res, body)
				identityID = expectTokens(t, "secondProvider", body)

				subject = email2
				r = newRegistrationFlow(t, returnTS.URL, time.Minute, flow.TypeBrowser)
				action = assertFormValues(t, r.ID, "valid")
				res, body = makeRequest(t, "valid", action, url.Values{})
				assertIdentity(t, res, body)
				expectTokens(t, "valid", body)
			})

			subject = email1
			client := testhelpers.NewClientWithCookieJar(t, nil, false)
			loginFlow := newLoginFlow(t, returnTS.URL, time.Minute, flow.TypeBrowser)
			var linkingLoginFlow struct{ ID string }
			t.Run("step=should fail login and start a new login", func(t *testing.T) {
				res, body := loginWithOIDC(t, client, loginFlow.ID, "valid")
				assertUIError(t, res, body, "You tried signing in with existing-oidc-identity-1@ory.sh which is already in use by another account. You can sign in using social sign in. You can sign in using one of the following social sign in providers: Secondprovider.")
				linkingLoginFlow.ID = gjson.GetBytes(body, "id").String()
				assert.NotEqual(t, loginFlow.ID.String(), linkingLoginFlow.ID, "should have started a new flow")
			})

			subject = email2
			t.Run("step=should fail login if existing identity identifier doesn't match", func(t *testing.T) {
				res, body := loginWithOIDC(t, client, uuid.Must(uuid.FromString(linkingLoginFlow.ID)), "valid")
				assertUIError(t, res, body, "Linked credentials do not match.")
			})

			subject = email1
			t.Run("step=should link oidc credentials to existing identity", func(t *testing.T) {
				res, body := loginWithOIDC(t, client, uuid.Must(uuid.FromString(linkingLoginFlow.ID)), "secondProvider")
				checkCredentialsLinked(res, body, identityID, "secondProvider")
			})
		})
	})

	t.Run("method=TestPopulateSignUpMethod", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeyPublicBaseURL, "https://foo/")

		sr, err := registration.NewFlow(conf, time.Minute, "nosurf", &http.Request{URL: urlx.ParseOrPanic("/")}, flow.TypeBrowser)
		require.NoError(t, err)
		require.NoError(t, reg.RegistrationStrategies(context.Background()).MustStrategy(identity.CredentialsTypeOIDC).(*oidc.Strategy).PopulateRegistrationMethod(&http.Request{}, sr))

		snapshotx.SnapshotTExcept(t, sr.UI, []string{"action", "nodes.0.attributes.value"})
	})

	t.Run("method=TestPopulateLoginMethod", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeyPublicBaseURL, "https://foo/")

		sr, err := login.NewFlow(conf, time.Minute, "nosurf", &http.Request{URL: urlx.ParseOrPanic("/")}, flow.TypeBrowser)
		require.NoError(t, err)
		require.NoError(t, reg.LoginStrategies(context.Background()).MustStrategy(identity.CredentialsTypeOIDC).(*oidc.Strategy).PopulateLoginMethod(&http.Request{}, identity.AuthenticatorAssuranceLevel1, sr))

		snapshotx.SnapshotTExcept(t, sr.UI, []string{"action", "nodes.0.attributes.value"})
	})
}

func prettyJSON(t *testing.T, body []byte) string {
	var out bytes.Buffer
	require.NoError(t, json.Indent(&out, body, "", "\t"))

	return out.String()
}

func TestCountActiveFirstFactorCredentials(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	strategy := oidc.NewStrategy(reg)

	toJson := func(c identity.CredentialsOIDC) []byte {
		out, err := json.Marshal(&c)
		require.NoError(t, err)
		return out
	}

	for k, tc := range []struct {
		in       map[identity.CredentialsType]identity.Credentials
		expected int
	}{
		{
			in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
				Type:   strategy.ID(),
				Config: []byte{},
			}},
		},
		{
			in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
				Type: strategy.ID(),
				Config: toJson(identity.CredentialsOIDC{Providers: []identity.CredentialsOIDCProvider{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
		},
		{
			in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
				Type:        strategy.ID(),
				Identifiers: []string{""},
				Config: toJson(identity.CredentialsOIDC{Providers: []identity.CredentialsOIDCProvider{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
		},
		{
			in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
				Type:        strategy.ID(),
				Identifiers: []string{"bar:"},
				Config: toJson(identity.CredentialsOIDC{Providers: []identity.CredentialsOIDCProvider{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
		},
		{
			in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
				Type:        strategy.ID(),
				Identifiers: []string{":foo"},
				Config: toJson(identity.CredentialsOIDC{Providers: []identity.CredentialsOIDCProvider{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
		},
		{
			in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
				Type:        strategy.ID(),
				Identifiers: []string{"not-bar:foo"},
				Config: toJson(identity.CredentialsOIDC{Providers: []identity.CredentialsOIDCProvider{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
		},
		{
			in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
				Type:        strategy.ID(),
				Identifiers: []string{"bar:not-foo"},
				Config: toJson(identity.CredentialsOIDC{Providers: []identity.CredentialsOIDCProvider{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
		},
		{
			in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
				Type:        strategy.ID(),
				Identifiers: []string{"bar:foo"},
				Config: toJson(identity.CredentialsOIDC{Providers: []identity.CredentialsOIDCProvider{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
			expected: 1,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			in := make(map[identity.CredentialsType]identity.Credentials)
			for _, v := range tc.in {
				in[v.Type] = v
			}
			actual, err := strategy.CountActiveFirstFactorCredentials(in)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestDisabledEndpoint(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.StrategyEnable(t, conf, identity.CredentialsTypeOIDC.String(), false)

	publicTS, _ := testhelpers.NewKratosServer(t, reg)

	t.Run("case=should not callback when oidc method is disabled", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)
		res, err := c.Get(publicTS.URL + oidc.RouteCallback)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, res.StatusCode)

		b := ioutilx.MustReadAll(res.Body)
		assert.Contains(t, string(b), "This endpoint was disabled by system administrator")
	})

	t.Run("case=should not auth when oidc method is disabled", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)

		t.Run("flow=settings", func(t *testing.T) {
			testhelpers.SetDefaultIdentitySchema(conf, "file://stub/stub.schema.json")
			c := testhelpers.NewHTTPClientWithArbitrarySessionCookie(t, reg)
			f := testhelpers.InitializeSettingsFlowViaAPI(t, c, publicTS)

			res, err := c.PostForm(f.Ui.Action, url.Values{"link": {"oidc"}})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, res.StatusCode)

			b := ioutilx.MustReadAll(res.Body)
			assert.Contains(t, string(b), "This endpoint was disabled by system administrator")
		})

		t.Run("flow=login", func(t *testing.T) {
			f := testhelpers.InitializeLoginFlowViaAPI(t, c, publicTS, false)
			res, err := c.PostForm(f.Ui.Action, url.Values{"provider": {"oidc"}})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, res.StatusCode)

			b := ioutilx.MustReadAll(res.Body)
			assert.Contains(t, string(b), "This endpoint was disabled by system administrator")
		})

		t.Run("flow=registration", func(t *testing.T) {
			f := testhelpers.InitializeRegistrationFlowViaAPI(t, c, publicTS)
			res, err := c.PostForm(f.Ui.Action, url.Values{"provider": {"oidc"}})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, res.StatusCode)

			b := ioutilx.MustReadAll(res.Body)
			assert.Contains(t, string(b), "This endpoint was disabled by system administrator")
		})
	})
}

func TestPostEndpointRedirect(t *testing.T) {
	var (
		conf, reg = internal.NewFastRegistryWithMocks(t)
		subject   string
		claims    idTokenClaims
		scope     []string
	)
	testhelpers.StrategyEnable(t, conf, identity.CredentialsTypeOIDC.String(), true)

	remoteAdmin, remotePublic, _ := newHydra(t, &subject, &claims, &scope)

	publicTS, _, _, _ := testhelpers.NewKratosServerWithCSRFAndRouters(t, reg)

	viperSetProviderConfig(
		t,
		conf,
		newOIDCProvider(t, publicTS, remotePublic, remoteAdmin, "apple"),
	)

	t.Run("case=should redirect to GET and preserve parameters"+publicTS.URL, func(t *testing.T) {
		// create a client that does not follow redirects
		c := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		res, err := c.PostForm(publicTS.URL+"/self-service/methods/oidc/callback/apple", url.Values{"state": {"foo"}, "test": {"3"}})
		require.NoError(t, err)
		defer res.Body.Close()
		assert.Equal(t, http.StatusFound, res.StatusCode)

		location, err := res.Location()
		require.NoError(t, err)
		assert.Equal(t, publicTS.URL+"/self-service/methods/oidc/callback/apple?state=foo&test=3", location.String())

		// We don't want to add/override CSRF cookie when redirecting
		testhelpers.AssertNoCSRFCookieInResponse(t, publicTS, c, res)
	})
}
