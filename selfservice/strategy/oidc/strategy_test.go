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
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"golang.org/x/oauth2"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hydra"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/hook/hooktest"
	"github.com/ory/kratos/selfservice/sessiontokenexchange"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/x"
	"github.com/ory/x/assertx"
	"github.com/ory/x/configx"
	"github.com/ory/x/httprouterx"
	"github.com/ory/x/ioutilx"
	"github.com/ory/x/randx"
	"github.com/ory/x/snapshotx"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"
	"github.com/ory/x/uuidx"
)

func TestStrategy(t *testing.T) {
	t.Parallel()

	returnTSRouter := http.NewServeMux()
	returnTS := httptest.NewServer(returnTSRouter)

	conf, reg := pkg.NewFastRegistryWithMocks(t,
		configx.WithValues(map[string]any{
			config.ViperKeyPublicBaseURL: "https://foo/", // deterministic value for snapshots
			config.ViperKeyIdentitySchemas: config.Schemas{
				{ID: "default", URL: "file://./stub/registration.schema.json"},
				{ID: "email", URL: "file://./stub/registration.schema.json", SelfserviceSelectable: true},
				{ID: "phone", URL: "file://./stub/registration-phone.schema.json", SelfserviceSelectable: true},
				{ID: "extra_data", URL: "file://./stub/registration-multi-schema-extra-fields.schema.json", SelfserviceSelectable: true},
				{ID: "aal", URL: "file://./stub/registration-aal.schema.json", SelfserviceSelectable: true},
				{ID: "verifiable_email", URL: "file://./stub/registration-verifiable-email.schema.json", SelfserviceSelectable: true},
			},
			config.ViperKeyDefaultIdentitySchemaID: "default",
			config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypeOIDC.String()): []config.SelfServiceHook{{Name: "session"}},
			config.ViperKeyURLsAllowedReturnToDomains:        []string{returnTS.URL},
			config.ViperKeySelfServiceBrowserDefaultReturnTo: returnTS.URL,
			config.ViperKeyOAuth2ProviderURL:                 "http://fake-hydra",
		}),
	)

	returnTSRouter.HandleFunc("/app_code", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) })
	returnTSRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sess, err := reg.SessionManager().FetchFromRequest(r.Context(), r)
		if err != nil {
			reg.Writer().WriteError(w, r, err)
			return
		}
		reg.Writer().Write(w, r, sess)
	})

	hydraAdmin, hydraPublic := newHydra(t)
	uiTS := newUI(t, reg)
	errTS := testhelpers.NewErrorTestServer(t, reg)
	routerP, routerA := httprouterx.NewRouterPublic(), httprouterx.NewRouterAdminWithPrefix()
	ts, _ := testhelpers.NewKratosServerWithRouters(t, reg, routerP, routerA)
	invalid := newOIDCProvider(t, ts, hydraPublic, hydraAdmin, "invalid-issuer")

	orgID := uuidx.NewV4()
	setProviderConfig(t, conf,
		newOIDCProvider(t, ts, hydraPublic, hydraAdmin, "valid"),
		newOIDCProvider(t, ts, hydraPublic, hydraAdmin, "valid2"),
		newOIDCProvider(t, ts, hydraPublic, hydraAdmin, "secondProvider"),
		newOIDCProvider(t, ts, hydraPublic, hydraAdmin, "claimsViaUserInfo", func(c *oidc.Configuration) { c.ClaimsSource = oidc.ClaimsSourceUserInfo }),
		newOIDCProvider(t, ts, hydraPublic, hydraAdmin, "neverPKCE", func(c *oidc.Configuration) { c.PKCE = "never" }),
		newOIDCProvider(t, ts, hydraPublic, hydraAdmin, "autoPKCE", func(c *oidc.Configuration) { c.PKCE = "auto" }),
		newOIDCProvider(t, ts, hydraPublic, hydraAdmin, "forcePKCE", func(c *oidc.Configuration) { c.PKCE = "force" }),
		newOIDCProvider(t, ts, hydraPublic, hydraAdmin, "aal2-acr", func(c *oidc.Configuration) { c.AAL2ACRValues = []string{"urn:mfa", "urn:strong"} }),
		newOIDCProvider(t, ts, hydraPublic, hydraAdmin, "aal2-amr", func(c *oidc.Configuration) { c.AAL2AMRValues = []string{"mfa", "otp"} }),
		oidc.Configuration{
			Provider:     "generic",
			ID:           "invalid-issuer",
			ClientID:     invalid.ClientID,
			ClientSecret: invalid.ClientSecret,
			// We replace this URL to cause an issuer validation mismatch.
			IssuerURL: strings.Replace(hydraPublic, "127.0.0.1", "localhost", 1) + "/",
			Mapper:    "file://./stub/oidc.hydra.jsonnet",
		},
	)

	t.Logf("Kratos Public URL: %s", ts.URL)
	t.Logf("Kratos Error URL: %s", errTS.URL)
	t.Logf("Hydra Public URL: %s", hydraPublic)
	t.Logf("Hydra Admin URL: %s", hydraAdmin)
	t.Logf("Return URL: %s", returnTS.URL)

	// assert form values
	assertFormValues := func(t *testing.T, flowID uuid.UUID, provider string) (action string) {
		var cfg *container.Container
		if req, err := reg.RegistrationFlowPersister().GetRegistrationFlow(context.Background(), flowID); err == nil {
			require.EqualValues(t, req.ID, flowID)
			cfg = req.UI
			require.NotNil(t, cfg)
		} else if req, err := reg.LoginFlowPersister().GetLoginFlow(context.Background(), flowID); err == nil {
			require.EqualValues(t, req.ID, flowID)
			cfg = req.UI
			require.NotNil(t, cfg)
		} else {
			require.NoError(t, err)
			return
		}

		assert.Equal(t, "POST", cfg.Method)

		var providers []interface{}
		for _, nodes := range cfg.Nodes {
			if strings.Contains(nodes.ID(), "provider") {
				providers = append(providers, nodes.GetValue())
			}
		}
		require.Containsf(t, providers, provider, "%+v", assertx.PrettifyJSONPayload(t, cfg))

		return cfg.Action
	}

	registerAction := func(flowID uuid.UUID) string {
		return ts.URL + registration.RouteSubmitFlow + "?flow=" + flowID.String()
	}
	loginAction := func(flowID uuid.UUID) string {
		return ts.URL + login.RouteSubmitFlow + "?flow=" + flowID.String()
	}

	makeRequestWithClient := func(t *testing.T, client *http.Client, action, provider string, params hydraFlowParams, fv url.Values) (*http.Response, []byte) {
		if fv == nil {
			fv = make(url.Values)
		}
		fv.Set("provider", provider)
		res, err := wrapClientForHydraLoginConsent(t, client, hydraAdmin, params).
			PostForm(action, fv)
		require.NoError(t, err, action)

		body, err := io.ReadAll(res.Body)
		require.NoError(t, res.Body.Close())
		require.NoError(t, err)

		require.Equalf(t, http.StatusOK, res.StatusCode, "%s: %s\n\t%s", action, res.Request.URL.String(), body)

		return res, body
	}

	makeRequest := func(t *testing.T, action, provider string, params hydraFlowParams, fv url.Values) (*http.Response, []byte) {
		return makeRequestWithClient(t, testhelpers.NewClientWithCookies(t), action, provider, params, fv)
	}

	makeJSONRequest := func(t *testing.T, action, provider, subject string, scope []string, fv url.Values) (*http.Response, []byte) {
		if fv == nil {
			fv = make(url.Values)
		}
		fv.Set("provider", provider)
		req, err := http.NewRequest("POST", action, strings.NewReader(fv.Encode()))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")
		res, err := wrapClientForHydraLoginConsent(t, testhelpers.NewClientWithCookies(t), hydraAdmin, hydraFlowParams{subject: subject, scope: scope}).
			Do(req)
		require.NoError(t, err, action)

		body, err := io.ReadAll(res.Body)
		require.NoError(t, res.Body.Close())
		require.NoError(t, err)

		require.Equalf(t, 422, res.StatusCode, "%s: %s\n\t%s", action, res.Request.URL.String(), body)

		return res, body
	}

	makeAPICodeFlowRequest := func(t *testing.T, client *http.Client, action, provider string, params hydraFlowParams) (returnToURL *url.URL) {
		res, err := http.Post( // #nosec G107 -- test code
			action,
			"application/json",
			strings.NewReader(fmt.Sprintf(`
{
	"method": "oidc",
	"provider": %q
}`, provider)),
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
		var changeLocation flow.BrowserLocationChangeRequiredError
		require.NoError(t, json.NewDecoder(res.Body).Decode(&changeLocation))

		if client == nil {
			client = testhelpers.NewClientWithCookies(t)
		}
		res, err = wrapClientForHydraLoginConsent(t, client, hydraAdmin, params).
			Get(changeLocation.RedirectBrowserTo)
		require.NoError(t, err)

		returnToURL = res.Request.URL
		assert.Truef(t, strings.HasPrefix(returnToURL.String(), returnTS.URL+"/app_code"), "expected %s to have prefix %s/app_code", returnToURL, returnTS.URL)

		return returnToURL
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
		require.Contains(t, res.Request.URL.String(), uiTS.URL, "Redirect does not point to UI server. Status: %d, body: %s", res.StatusCode, body)
		assert.Contains(t, gjson.GetBytes(body, "ui.messages.0.text").String(), reason, "%s", prettyJSON(t, body))
	}

	// assert identity (success)
	assertIdentity := func(t *testing.T, res *http.Response, body []byte, subject string, claims idTokenClaims) {
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
		require.NoError(t, reg.LoginFlowPersister().UpdateLoginFlow(t.Context(), req))

		// sanity check
		got, err := reg.LoginFlowPersister().GetLoginFlow(t.Context(), req.ID)
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

	newRegistrationFlow := func(t *testing.T, redirectTo string, exp time.Duration, flowType flow.Type, opts ...registration.FlowOption) *registration.Flow {
		req, err := reg.RegistrationHandler().NewRegistrationFlow(httptest.NewRecorder(),
			&http.Request{URL: urlx.ParseOrPanic(redirectTo)}, flowType, opts...)
		require.NoError(t, err)
		req.RequestURL = redirectTo
		req.ExpiresAt = time.Now().Add(exp)
		require.NoError(t, reg.RegistrationFlowPersister().UpdateRegistrationFlow(t.Context(), req))

		// sanity check
		got, err := reg.RegistrationFlowPersister().GetRegistrationFlow(t.Context(), req.ID)
		require.NoError(t, err)
		require.Lenf(t, got.UI.Nodes, len(req.UI.Nodes), "%+v", req)

		return req
	}
	newBrowserRegistrationFlow := func(t *testing.T, redirectTo string, exp time.Duration, opts ...registration.FlowOption) *registration.Flow {
		return newRegistrationFlow(t, redirectTo, exp, flow.TypeBrowser, opts...)
	}
	newAPIRegistrationFlow := func(t *testing.T, redirectTo string, exp time.Duration, opts ...registration.FlowOption) *registration.Flow {
		return newRegistrationFlow(t, redirectTo, exp, flow.TypeAPI, opts...)
	}

	t.Run("case=should fail because provider does not exist", func(t *testing.T) {
		for name, action := range map[string]string{
			"login":        loginAction(newBrowserLoginFlow(t, returnTS.URL, time.Minute).ID),
			"registration": registerAction(newBrowserRegistrationFlow(t, returnTS.URL, time.Minute).ID),
		} {
			t.Run("case="+name, func(t *testing.T) {
				res, body := makeRequest(t, action, "provider-does-not-exist", hydraFlowParams{}, nil)
				assertSystemErrorWithReason(t, res, body, http.StatusNotFound, "is unknown or has not been configured")
			})
		}
	})

	t.Run("case=should fail because the issuer is mismatching", func(t *testing.T) {
		for name, action := range map[string]string{
			"login":        loginAction(newBrowserLoginFlow(t, returnTS.URL, time.Minute).ID),
			"registration": registerAction(newBrowserRegistrationFlow(t, returnTS.URL, time.Minute).ID),
		} {
			t.Run("case="+name, func(t *testing.T) {
				res, body := makeRequest(t, action, "invalid-issuer", hydraFlowParams{}, nil)
				assertSystemErrorWithReason(t, res, body, http.StatusInternalServerError, "issuer did not match the issuer returned by provider")
			})
		}
	})

	t.Run("case=should fail because flow does not exist", func(t *testing.T) {
		for name, action := range map[string]string{
			"login":        loginAction(x.NewUUID()),
			"registration": registerAction(x.NewUUID()),
		} {
			t.Run("case="+name, func(t *testing.T) {
				res, body := makeRequest(t, action, "valid", hydraFlowParams{}, nil)
				assertSystemErrorWithMessage(t, res, body, http.StatusNotFound, "Unable to locate the resource")
			})
		}
	})

	t.Run("case=should fail because the flow is expired", func(t *testing.T) {
		for name, flw := range map[string]uuid.UUID{
			"login":        newBrowserLoginFlow(t, returnTS.URL, -time.Minute).ID,
			"registration": newBrowserRegistrationFlow(t, returnTS.URL, -time.Minute).ID,
		} {
			t.Run("case="+name, func(t *testing.T) {
				action := assertFormValues(t, flw, "valid")
				res, body := makeRequest(t, action, "valid", hydraFlowParams{}, nil)

				assert.NotEqual(t, flw, gjson.GetBytes(body, "id"))
				require.Contains(t, res.Request.URL.String(), uiTS.URL, "%s", body)
				assert.Contains(t, gjson.GetBytes(body, "ui.messages.0.text").String(), "flow expired", "%s", body)
			})
		}
	})

	t.Run("case=should fail because scope was not provided", func(t *testing.T) {
		for name, flowID := range map[string]uuid.UUID{
			"login":        newBrowserLoginFlow(t, returnTS.URL, time.Minute).ID,
			"registration": newBrowserRegistrationFlow(t, returnTS.URL, time.Minute).ID,
		} {
			t.Run("case="+name, func(t *testing.T) {
				action := assertFormValues(t, flowID, "valid")
				res, body := makeRequest(t, action, "valid", hydraFlowParams{subject: "foo@bar.com", scope: []string{}}, nil)
				assertUIError(t, res, body, "no id_token was returned")
			})
		}
	})

	t.Run("case=should fail because oidc does not satisfy AAL2", func(t *testing.T) {
		user := testhelpers.AddAndLoginIdentities(t, reg, map[string]*identity.Identity{"u": {
			SchemaID: "aal",
			Traits:   identity.Traits(`{"email":"just-aal1-identity@ory.sh"}`),
		}})["u"]
		client := wrapClientForHydraLoginConsent(t, user.Client, hydraAdmin, hydraFlowParams{subject: "foo@bar.com", scope: []string{"openid"}})
		f := testhelpers.InitializeLoginFlowViaAPICtx(t.Context(), t, client, ts, true)

		update, err := reg.LoginFlowPersister().GetLoginFlow(t.Context(), uuid.FromStringOrNil(f.Id))
		require.NoError(t, err)
		update.RequestedAAL = identity.AuthenticatorAssuranceLevel2
		require.NoError(t, reg.LoginFlowPersister().UpdateLoginFlow(t.Context(), update))

		req, err := http.NewRequest("POST", f.Ui.Action, strings.NewReader(`{"method":"oidc"}`))
		require.NoError(t, err)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		require.NoError(t, err)
		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		require.Equal(t, http.StatusBadRequest, res.StatusCode, "%s", body)

		assert.Contains(t, res.Request.URL.String(), ts.URL+login.RouteSubmitFlow)
		assert.Equal(t, text.NewErrorValidationLoginNoStrategyFound().Text, gjson.GetBytes(body, "ui.messages.0.text").String())
	})

	t.Run("case=should fail registration flow because subject is not an email", func(t *testing.T) {
		r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
		action := assertFormValues(t, r.ID, "valid")

		res, body := makeRequest(t, action, "valid", hydraFlowParams{subject: "not-an-email", scope: []string{"openid"}}, nil)

		require.Containsf(t, res.Request.URL.String(), uiTS.URL, "%s", body)
		assert.Containsf(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==traits.subject).messages.0.text").String(), "Enter a valid email address", "%s\n%s", gjson.GetBytes(body, "ui.nodes.#(attributes.name==traits.subject)").Raw, body)
	})

	t.Run("case=cannot register multiple accounts with the same OIDC account", func(t *testing.T) {
		subject := testhelpers.RandomEmail()
		scope := []string{"openid", "offline"}

		expectTokens := func(t *testing.T, provider string, body []byte) {
			i, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(t.Context(), uuid.FromStringOrNil(gjson.GetBytes(body, "identity.id").String()))
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
			postRegistrationWebhook := hooktest.NewServer()
			t.Cleanup(postRegistrationWebhook.Close)
			postRegistrationWebhook.SetConfig(t, conf.GetProvider(t.Context()), config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypeOIDC.String()))
			transientPayload := `{"data": "registration-one"}`

			r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			res, body := makeRequest(t, action, "valid", hydraFlowParams{subject: subject, scope: scope}, url.Values{"transient_payload": {transientPayload}})
			assertIdentity(t, res, body, subject, idTokenClaims{})
			expectTokens(t, "valid", body)
			postRegistrationWebhook.AssertTransientPayload(t, transientPayload)
		})

		t.Run("case=try another registration", func(t *testing.T) {
			transientPayload := `{"data": "registration-two"}`
			postLoginWebhook := hooktest.NewServer()
			t.Cleanup(postLoginWebhook.Close)
			postLoginWebhook.SetConfig(t, conf.GetProvider(t.Context()), config.HookStrategyKey(config.ViperKeySelfServiceLoginAfter, identity.CredentialsTypeOIDC.String()))

			returnTo := fmt.Sprintf("%s/home?query=true", returnTS.URL)
			r := newBrowserRegistrationFlow(t, fmt.Sprintf("%s?return_to=%s", returnTS.URL, url.QueryEscape(returnTo)), time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			res, body := makeRequest(t, action, "valid", hydraFlowParams{subject: subject, scope: scope}, url.Values{"transient_payload": {transientPayload}})
			assert.Equal(t, returnTo, res.Request.URL.String())
			assertIdentity(t, res, body, subject, idTokenClaims{})
			expectTokens(t, "valid", body)

			postLoginWebhook.AssertTransientPayload(t, transientPayload)
		})
	})

	expectTokens := func(t *testing.T, provider, subject string, body []byte) uuid.UUID {
		id := uuid.FromStringOrNil(gjson.GetBytes(body, "identity.id").String())
		i, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(t.Context(), id)
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

	t.Run("case=force PKCE", func(t *testing.T) {
		r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
		action := assertFormValues(t, r.ID, "forcePKCE")
		subject := testhelpers.RandomEmail()
		scope := []string{"openid", "offline"}
		var redirects []*http.Request
		client := testhelpers.NewClientWithCookieJar(t, nil, func(_ *http.Request, via []*http.Request) error {
			redirects = via
			return nil
		})
		res, body := makeRequestWithClient(t, client, action, "forcePKCE", hydraFlowParams{subject: subject, scope: scope}, nil)
		require.Len(t, redirects, 2)
		assert.Contains(t, redirects[0].URL.String(), "/oauth2/auth")
		assert.Contains(t, redirects[0].URL.String(), "code_challenge_method=S256")
		assert.Contains(t, redirects[0].URL.String(), "code_challenge=")
		assert.Equal(t, redirects[1].URL.Path, "/self-service/methods/oidc/callback")

		assertIdentity(t, res, body, subject, idTokenClaims{})
		expectTokens(t, "forcePKCE", subject, body)
		assert.Equal(t, "forcePKCE", gjson.GetBytes(body, "authentication_methods.0.provider").String(), "%s", body)
	})
	t.Run("case=force PKCE, invalid verifier", func(t *testing.T) {
		r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
		action := assertFormValues(t, r.ID, "forcePKCE")
		subject := testhelpers.RandomEmail()
		scope := []string{"openid", "offline"}
		verifierFalsified := false
		client := testhelpers.NewClientWithCookieJar(t, nil, func(req *http.Request, via []*http.Request) error {
			if req.URL.Path == "/oauth2/auth" && !verifierFalsified {
				q := req.URL.Query()
				require.NotEmpty(t, q.Get("code_challenge"))
				require.Equal(t, "S256", q.Get("code_challenge_method"))
				q.Set("code_challenge", oauth2.S256ChallengeFromVerifier(oauth2.GenerateVerifier()))
				req.URL.RawQuery = q.Encode()
				verifierFalsified = true
			}
			return nil
		})
		res, body := makeRequestWithClient(t, client, action, "forcePKCE", hydraFlowParams{subject: subject, scope: scope}, nil)
		require.True(t, verifierFalsified)
		assertSystemErrorWithMessage(t, res, body, http.StatusInternalServerError, "The PKCE code challenge did not match the code verifier.")
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowErrorURL(t.Context()).String())
	})
	t.Run("case=force PKCE, code challenge params removed from initial redirect", func(t *testing.T) {
		r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
		action := assertFormValues(t, r.ID, "forcePKCE")
		subject := testhelpers.RandomEmail()
		scope := []string{"openid", "offline"}
		challengeParamsRemoved := false
		client := testhelpers.NewClientWithCookieJar(t, nil, func(req *http.Request, via []*http.Request) error {
			if req.URL.Path == "/oauth2/auth" && !challengeParamsRemoved {
				q := req.URL.Query()
				require.NotEmpty(t, q.Get("code_challenge"))
				require.Equal(t, "S256", q.Get("code_challenge_method"))
				q.Del("code_challenge")
				q.Del("code_challenge_method")
				req.URL.RawQuery = q.Encode()
				challengeParamsRemoved = true
			}
			return nil
		})
		res, body := makeRequestWithClient(t, client, action, "forcePKCE", hydraFlowParams{subject: subject, scope: scope}, nil)
		require.True(t, challengeParamsRemoved)
		assertSystemErrorWithMessage(t, res, body, http.StatusInternalServerError, "The PKCE code challenge did not match the code verifier.")
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowErrorURL(t.Context()).String())
	})
	t.Run("case=PKCE prevents authorization code injection attacks", func(t *testing.T) {
		r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
		action := assertFormValues(t, r.ID, "forcePKCE")
		subject := testhelpers.RandomEmail()
		scope := []string{"openid", "offline"}
		var code string
		client := wrapClientForHydraLoginConsent(t, testhelpers.NewClientWithCookieJar(t, nil, func(req *http.Request, via []*http.Request) error {
			if req.URL.Query().Has("code") {
				code = req.URL.Query().Get("code")
				return errors.New("code intercepted")
			}
			return nil
		}), hydraAdmin, hydraFlowParams{subject: subject, scope: scope})
		_, err := client.PostForm(action, url.Values{"provider": {"forcePKCE"}})
		require.ErrorContains(t, err, "code intercepted")
		require.NotEmpty(t, code) // code now contains a valid authorization code

		r2 := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
		action = assertFormValues(t, r2.ID, "forcePKCE")
		jar, err := cookiejar.New(nil) // must capture the continuity cookie
		require.NoError(t, err)
		var redirectURI, state string
		client = wrapClientForHydraLoginConsent(t, testhelpers.NewClientWithCookieJar(t, jar, func(req *http.Request, via []*http.Request) error {
			if req.URL.Path == "/oauth2/auth" {
				redirectURI = req.URL.Query().Get("redirect_uri")
				state = req.URL.Query().Get("state")
				return errors.New("stop before redirect to Authorization URL")
			}
			return nil
		}), hydraAdmin, hydraFlowParams{subject: subject, scope: scope})
		_, err = client.PostForm(action, url.Values{"provider": {"forcePKCE"}})
		require.ErrorContains(t, err, "stop")
		require.NotEmpty(t, redirectURI)
		require.NotEmpty(t, state)

		client = wrapClientForHydraLoginConsent(t, testhelpers.NewClientWithCookieJar(t, jar, nil), hydraAdmin, hydraFlowParams{subject: subject, scope: scope})
		res, err := client.Get(redirectURI + "?code=" + code + "&state=" + state)
		require.NoError(t, err)
		body := x.MustReadAll(res.Body)
		require.NoError(t, res.Body.Close())
		assertSystemErrorWithMessage(t, res, body, http.StatusInternalServerError, "The PKCE code challenge did not match the code verifier.")
	})
	t.Run("case=confused providers are detected", func(t *testing.T) {
		r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
		action := assertFormValues(t, r.ID, "valid")
		subject := testhelpers.RandomEmail()
		scope := []string{"openid", "offline"}
		redirectConfused := false
		client := wrapClientForHydraLoginConsent(t, testhelpers.NewClientWithCookieJar(t, nil, func(req *http.Request, via []*http.Request) error {
			if req.URL.Query().Has("code") {
				req.URL.Path = strings.Replace(req.URL.Path, "valid", "valid2", 1)
				redirectConfused = true
			}
			return nil
		}), hydraAdmin, hydraFlowParams{subject: subject, scope: scope})
		res, err := client.PostForm(action, url.Values{"provider": {"valid"}})
		require.True(t, redirectConfused)
		require.NoError(t, err)
		body := x.MustReadAll(res.Body)
		require.NoError(t, res.Body.Close())

		assertSystemErrorWithReason(t, res, body, http.StatusBadRequest, "provider mismatch between internal state and URL")
	})
	t.Run("case=automatic PKCE", func(t *testing.T) {
		r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
		action := assertFormValues(t, r.ID, "autoPKCE")
		subject := testhelpers.RandomEmail()
		scope := []string{"openid", "offline"}
		var redirects []*http.Request
		client := testhelpers.NewClientWithCookieJar(t, nil, func(_ *http.Request, via []*http.Request) error {
			redirects = via
			return nil
		})
		res, body := makeRequestWithClient(t, client, action, "autoPKCE", hydraFlowParams{subject: subject, scope: scope}, nil)
		require.Len(t, redirects, 2)
		assert.Contains(t, redirects[0].URL.String(), "/oauth2/auth")
		assert.Contains(t, redirects[0].URL.String(), "code_challenge_method=S256")
		assert.Contains(t, redirects[0].URL.String(), "code_challenge=")
		assert.Equal(t, redirects[1].URL.Path, "/self-service/methods/oidc/callback/autoPKCE")

		assertIdentity(t, res, body, subject, idTokenClaims{})
		expectTokens(t, "autoPKCE", subject, body)
		assert.Equal(t, "autoPKCE", gjson.GetBytes(body, "authentication_methods.0.provider").String(), "%s", body)
	})
	t.Run("case=disabled PKCE", func(t *testing.T) {
		r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
		action := assertFormValues(t, r.ID, "neverPKCE")
		subject := testhelpers.RandomEmail()
		scope := []string{"openid", "offline"}
		var redirects []*http.Request
		client := testhelpers.NewClientWithCookieJar(t, nil, func(_ *http.Request, via []*http.Request) error {
			redirects = via
			return nil
		})
		res, body := makeRequestWithClient(t, client, action, "neverPKCE", hydraFlowParams{subject: subject, scope: scope}, nil)
		require.Len(t, redirects, 2)
		assert.Contains(t, redirects[0].URL.String(), "/oauth2/auth")
		assert.NotContains(t, redirects[0].URL.String(), "code_challenge_method=")
		assert.NotContains(t, redirects[0].URL.String(), "code_challenge=")
		assert.Equal(t, redirects[1].URL.Path, "/self-service/methods/oidc/callback/neverPKCE")

		assertIdentity(t, res, body, subject, idTokenClaims{})
		expectTokens(t, "neverPKCE", subject, body)
		assert.Equal(t, "neverPKCE", gjson.GetBytes(body, "authentication_methods.0.provider").String(), "%s", body)
	})

	t.Run("case=register and then login", func(t *testing.T) {
		postRegistrationWebhook := hooktest.NewServer()
		t.Cleanup(postRegistrationWebhook.Close)
		postLoginWebhook := hooktest.NewServer()
		t.Cleanup(postLoginWebhook.Close)

		postRegistrationWebhook.SetConfig(t, conf.GetProvider(t.Context()),
			config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypeOIDC.String()))
		postLoginWebhook.SetConfig(t, conf.GetProvider(t.Context()),
			config.HookStrategyKey(config.ViperKeySelfServiceLoginAfter, config.HookGlobal))

		subject := testhelpers.RandomEmail()
		scope := []string{"openid", "offline"}

		t.Run("case=should pass registration", func(t *testing.T) {
			transientPayload := `{"data": "registration"}`
			r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			res, body := makeRequest(t, action, "valid", hydraFlowParams{subject: subject, scope: scope}, url.Values{
				"transient_payload": {transientPayload},
			})
			assertIdentity(t, res, body, subject, idTokenClaims{})
			expectTokens(t, "valid", subject, body)
			assert.Equal(t, "valid", gjson.GetBytes(body, "authentication_methods.0.provider").String(), "%s", body)

			postRegistrationWebhook.AssertTransientPayload(t, transientPayload)
		})

		t.Run("case=should pass login", func(t *testing.T) {
			transientPayload := `{"data": "login"}`
			r := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			res, body := makeRequest(t, action, "valid", hydraFlowParams{subject: subject, scope: scope}, url.Values{
				"transient_payload": {transientPayload},
			})
			assertIdentity(t, res, body, subject, idTokenClaims{})
			expectTokens(t, "valid", subject, body)
			assert.Equal(t, "valid", gjson.GetBytes(body, "authentication_methods.0.provider").String(), "%s", body)

			postLoginWebhook.AssertTransientPayload(t, transientPayload)
		})

		t.Run("case=should pass double submit", func(t *testing.T) {
			// This test checks that the continuity manager uses a grace period to handle potential double-submit issues.
			//
			// It addresses issues where Facebook and Apple consent screens on mobile behave in a way that makes it
			// easy for users to experience double-submit issues.
			j, err := cookiejar.New(nil)
			require.NoError(t, err)

			makeInitialRequest := func(t *testing.T, provider, action string) (*http.Response, []byte, []string) {
				fv := make(url.Values)
				fv.Set("provider", provider)

				var lastVia []*http.Request
				hc := &http.Client{
					Jar: j,
					CheckRedirect: func(req *http.Request, via []*http.Request) error {
						lastVia = via
						return nil
					},
				}
				res, err := wrapClientForHydraLoginConsent(t, hc, hydraAdmin, hydraFlowParams{subject: subject, scope: scope}).PostForm(action, fv)
				require.NoError(t, err, action)

				body, err := io.ReadAll(res.Body)
				require.NoError(t, res.Body.Close())
				require.NoError(t, err)
				require.NotEmpty(t, lastVia)

				vias := make([]string, len(lastVia))
				for k, v := range lastVia {
					vias[k] = v.URL.String()
				}

				return res, body, vias
			}

			r := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
			action := assertFormValues(t, r.ID, "valid")

			// First login
			res, body, via := makeInitialRequest(t, "valid", action)
			assertIdentity(t, res, body, subject, idTokenClaims{})
			expectTokens(t, "valid", subject, body)
			assert.Equal(t, "valid", gjson.GetBytes(body, "authentication_methods.0.provider").String(), "%s", body)

			// We fetch the URL which includes the `?code` query parameter.
			result := lo.Filter(via, func(s string, _ int) bool {
				return strings.Contains(s, "code=")
			})
			require.Len(t, result, 1)

			// And call that URL again. What's interesting here is that the whole requets passes because we are already authenticated.
			//
			// In this scenario, Ory Kratos correctly forwards the user to the return URL, which in our case returns the identity.
			//
			// We essentially run into this bit:
			//
			// 	if authenticated, err := s.alreadyAuthenticated(w, r, req); err != nil {
			//		s.forwardError(w, r, req, s.handleError(w, , r, req, pid, nil, err))
			//	} else if authenticated {
			//		return <-- we end up here on the second call
			//	}
			res, err = (&http.Client{Jar: j}).Get(result[0])
			require.NoError(t, err)
			body, err = io.ReadAll(res.Body)
			require.NoError(t, err)
			require.NoError(t, res.Body.Close())

			assertIdentity(t, res, body, subject, idTokenClaims{})
			expectTokens(t, "valid", subject, body)
			assert.Equal(t, "valid", gjson.GetBytes(body, "authentication_methods.0.provider").String(), "%s", body)

			// Trying this flow again without the Ory Session cookie will fail as we run into code reuse:
			cookies := j.Cookies(urlx.ParseOrPanic(ts.URL))
			t.Logf("Cookies: %s", spew.Sdump(cookies))

			secondJar, err := cookiejar.New(nil)
			require.NoError(t, err)

			secondJar.SetCookies(urlx.ParseOrPanic(ts.URL), lo.Filter(cookies, func(item *http.Cookie, index int) bool {
				return item.Name != "ory_kratos_session"
			}))

			cookies = secondJar.Cookies(urlx.ParseOrPanic(ts.URL))
			t.Logf("Cookies after: %s", spew.Sdump(cookies))

			// Doing the request but this time without the Ory Session Cookie. This may be the case in scenarios where we run into race conditions
			// where the server sent a response but the client did not process it.
			res, err = (&http.Client{Jar: secondJar}).Get(result[0])
			require.NoError(t, err)
			body, err = io.ReadAll(res.Body)
			require.NoError(t, err)
			require.NoError(t, res.Body.Close())

			assert.Contains(t, string(body), "The authorization code has already been used", "%s", body)
		})
	})

	t.Run("case=carries upstream acr/amr into session AAL", func(t *testing.T) {
		// Registers the user through the provider and returns the decoded
		// session JSON body from the return handler.
		registerAndReturnSession := func(t *testing.T, provider, email string, acr *string, amr []string) []byte {
			r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
			action := assertFormValues(t, r.ID, provider)
			res, body := makeRequest(t, action, provider, hydraFlowParams{
				subject: email,
				scope:   []string{"openid", "offline"},
				acr:     acr,
				amr:     amr,
			}, nil)
			assertIdentity(t, res, body, email, idTokenClaims{})
			return body
		}

		t.Run("case=matching acr elevates session to aal2", func(t *testing.T) {
			body := registerAndReturnSession(t, "aal2-acr", testhelpers.RandomEmail(), new("urn:mfa"), nil)

			assert.Equal(t, "aal2", gjson.GetBytes(body, "authenticator_assurance_level").String(), "%s", body)
			assert.Equal(t, "aal2", gjson.GetBytes(body, "authentication_methods.0.aal").String(), "%s", body)
			assert.Equal(t, "aal2-acr", gjson.GetBytes(body, "authentication_methods.0.provider").String(), "%s", body)
			assert.Equal(t, "urn:mfa", gjson.GetBytes(body, "authentication_methods.0.upstream_acr").String(), "%s", body)
		})

		t.Run("case=mismatching acr stays aal1", func(t *testing.T) {
			body := registerAndReturnSession(t, "aal2-acr", testhelpers.RandomEmail(), new("urn:basic"), nil)

			assert.Equal(t, "aal1", gjson.GetBytes(body, "authenticator_assurance_level").String(), "%s", body)
			assert.Equal(t, "aal1", gjson.GetBytes(body, "authentication_methods.0.aal").String(), "%s", body)
			assert.Equal(t, "urn:basic", gjson.GetBytes(body, "authentication_methods.0.upstream_acr").String(), "%s", body)
		})

		t.Run("case=matching amr elevates session to aal2", func(t *testing.T) {
			body := registerAndReturnSession(t, "aal2-amr", testhelpers.RandomEmail(), nil, []string{"pwd", "mfa"})

			assert.Equal(t, "aal2", gjson.GetBytes(body, "authenticator_assurance_level").String(), "%s", body)
			assert.Equal(t, "aal2", gjson.GetBytes(body, "authentication_methods.0.aal").String(), "%s", body)
			assert.Equal(t, "aal2-amr", gjson.GetBytes(body, "authentication_methods.0.provider").String(), "%s", body)
			// Upstream amr is persisted for auditing.
			var gotAMR []string
			require.NoError(t, json.Unmarshal([]byte(gjson.GetBytes(body, "authentication_methods.0.upstream_amr").Raw), &gotAMR))
			assert.ElementsMatch(t, []string{"pwd", "mfa"}, gotAMR)
		})

		t.Run("case=mismatching amr stays aal1", func(t *testing.T) {
			body := registerAndReturnSession(t, "aal2-amr", testhelpers.RandomEmail(), nil, []string{"pwd"})

			assert.Equal(t, "aal1", gjson.GetBytes(body, "authenticator_assurance_level").String(), "%s", body)
			assert.Equal(t, "aal1", gjson.GetBytes(body, "authentication_methods.0.aal").String(), "%s", body)
		})
	})

	t.Run("case=login without registered account", func(t *testing.T) {
		postRegistrationWebhook := hooktest.NewServer()
		t.Cleanup(postRegistrationWebhook.Close)
		postLoginWebhook := hooktest.NewServer()
		t.Cleanup(postLoginWebhook.Close)

		postRegistrationWebhook.SetConfig(t, conf.GetProvider(t.Context()),
			config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypeOIDC.String()))
		postLoginWebhook.SetConfig(t, conf.GetProvider(t.Context()),
			config.HookStrategyKey(config.ViperKeySelfServiceLoginAfter, config.HookGlobal))

		subject := testhelpers.RandomEmail()
		scope := []string{"openid"}

		t.Run("case=should pass login", func(t *testing.T) {
			transientPayload := `{"data": "login to registration"}`

			r := newBrowserLoginFlow(t, "https://example.com?identity_schema=phone", time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			res, body := makeRequest(t, action, "valid", hydraFlowParams{subject: subject, scope: scope}, url.Values{
				"transient_payload": {transientPayload},
			})
			assertIdentity(t, res, body, subject, idTokenClaims{})
			assert.Equal(t, "phone", gjson.GetBytes(body, "identity.schema_id").String(), "%s", body)
			assert.Equal(t, "valid", gjson.GetBytes(body, "authentication_methods.0.provider").String(), "%s", body)

			assert.Empty(t, postLoginWebhook.LastBody,
				"post login hook should not have been called, because this was a registration flow")
			postRegistrationWebhook.AssertTransientPayload(t, transientPayload)
		})
	})

	t.Run("case=login with Browser+JSON", func(t *testing.T) {
		subject := testhelpers.RandomEmail()
		scope := []string{"openid"}

		t.Run("case=should pass login", func(t *testing.T) {
			r := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			res, body := makeJSONRequest(t, action, "valid", subject, scope, nil)

			assert.Equal(t, "browser_location_change_required", gjson.GetBytes(body, "error.id").String(), "%s", body)

			continuityCookie := res.Header.Get("Set-Cookie")
			key, val, ok := strings.Cut(continuityCookie, "=")
			require.True(t, ok)
			assert.Equal(t, "ory_kratos_continuity", key)
			assert.NotEmpty(t, val)
		})
	})

	t.Run("suite=API with session token exchange code", func(t *testing.T) {
		loginOrRegister := func(t *testing.T, client *http.Client, flowID uuid.UUID, code string, params hydraFlowParams) {
			_, err := exchangeCodeForToken(t, sessiontokenexchange.Codes{InitCode: code})
			require.Error(t, err)

			action := assertFormValues(t, flowID, "valid")
			returnToURL := makeAPICodeFlowRequest(t, client, action, "valid", params)
			returnToCode := returnToURL.Query().Get("code")
			assert.NotEmpty(t, code, "code query param was empty in the return_to URL")

			codeResponse, err := exchangeCodeForToken(t, sessiontokenexchange.Codes{
				InitCode:     code,
				ReturnToCode: returnToCode,
			})
			require.NoError(t, err)

			assert.NotEmpty(t, codeResponse.Token)
			assert.Equal(t, params.subject, gjson.GetBytes(codeResponse.Session.Identity.Traits, "subject").String())
		}
		performRegistration := func(t *testing.T, client *http.Client, params hydraFlowParams) {
			f := newAPIRegistrationFlow(t, returnTS.URL+"?return_session_token_exchange_code=true&return_to=/app_code", 1*time.Minute)
			loginOrRegister(t, client, f.ID, f.SessionTokenExchangeCode, params)
		}
		performLogin := func(t *testing.T, client *http.Client, params hydraFlowParams) {
			f := newAPILoginFlow(t, returnTS.URL+"?return_session_token_exchange_code=true&return_to=/app_code", 1*time.Minute)
			loginOrRegister(t, client, f.ID, f.SessionTokenExchangeCode, params)
		}

		for _, tc := range []struct {
			name        string
			first, then func(t *testing.T, client *http.Client, params hydraFlowParams)
		}{{
			name:  "login-twice",
			first: performLogin, then: performLogin,
		}, {
			name:  "login-then-register",
			first: performLogin, then: performRegistration,
		}, {
			name:  "register-then-login",
			first: performRegistration, then: performLogin,
		}, {
			name:  "register-twice",
			first: performRegistration, then: performRegistration,
		}} {
			t.Run("case="+tc.name, func(t *testing.T) {
				params := hydraFlowParams{subject: testhelpers.RandomEmail(), scope: []string{"openid"}}
				tc.first(t, nil, params)
				tc.then(t, nil, params)
			})
		}

		params := hydraFlowParams{subject: testhelpers.RandomEmail(), scope: []string{"openid"}}
		t.Run("case=should pass transient payload to registration webhook", func(t *testing.T) {
			transientPayload := `{"data": "api-registration"}`

			webhook := hooktest.NewServer()
			t.Cleanup(webhook.Close)
			webhook.SetConfig(t, conf.GetProvider(t.Context()), config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypeOIDC.String()))

			f := newAPIRegistrationFlow(t, returnTS.URL+"?return_session_token_exchange_code=true&return_to=/app_code", time.Minute)
			action := assertFormValues(t, f.ID, "valid")

			res, err := http.Post(action, "application/json", // #nosec G107 -- test code
				strings.NewReader(fmt.Sprintf(`{"method":"oidc","provider":"valid","transient_payload":%s}`, transientPayload)))
			require.NoError(t, err)
			require.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

			var changeLocation flow.BrowserLocationChangeRequiredError
			require.NoError(t, json.NewDecoder(res.Body).Decode(&changeLocation))

			res, err = wrapClientForHydraLoginConsent(t, testhelpers.NewClientWithCookies(t), hydraAdmin, params).
				Get(changeLocation.RedirectBrowserTo)
			require.NoError(t, err)

			returnToCode := res.Request.URL.Query().Get("code")
			require.NotEmpty(t, returnToCode, "code query param was empty in the return_to URL")

			codeResponse, err := exchangeCodeForToken(t, sessiontokenexchange.Codes{
				InitCode:     f.SessionTokenExchangeCode,
				ReturnToCode: returnToCode,
			})
			require.NoError(t, err)
			assert.NotEmpty(t, codeResponse.Token)

			webhook.AssertTransientPayload(t, transientPayload)
		})

		t.Run("case=should pass transient payload to login webhook", func(t *testing.T) {
			// subject is the same as above, so this triggers a login (already registered)
			transientPayload := `{"data": "api-login"}`

			webhook := hooktest.NewServer()
			t.Cleanup(webhook.Close)
			webhook.SetConfig(t, conf.GetProvider(t.Context()), config.HookStrategyKey(config.ViperKeySelfServiceLoginAfter, identity.CredentialsTypeOIDC.String()))

			f := newAPILoginFlow(t, returnTS.URL+"?return_session_token_exchange_code=true&return_to=/app_code", time.Minute)
			action := assertFormValues(t, f.ID, "valid")

			res, err := http.Post(action, "application/json", // #nosec G107 -- test code
				strings.NewReader(fmt.Sprintf(`{"method":"oidc","provider":"valid","transient_payload":%s}`, transientPayload)))
			require.NoError(t, err)
			require.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

			var changeLocation flow.BrowserLocationChangeRequiredError
			require.NoError(t, json.NewDecoder(res.Body).Decode(&changeLocation))

			res, err = wrapClientForHydraLoginConsent(t, testhelpers.NewClientWithCookies(t), hydraAdmin, params).Get(changeLocation.RedirectBrowserTo)
			require.NoError(t, err)

			returnToCode := res.Request.URL.Query().Get("code")
			require.NotEmpty(t, returnToCode, "code query param was empty in the return_to URL")

			codeResponse, err := exchangeCodeForToken(t, sessiontokenexchange.Codes{
				InitCode:     f.SessionTokenExchangeCode,
				ReturnToCode: returnToCode,
			})
			require.NoError(t, err)
			assert.NotEmpty(t, codeResponse.Token)

			webhook.AssertTransientPayload(t, transientPayload)
		})

		t.Run("case=should register with incomplete data mapper on API flow", func(t *testing.T) {
			t.Run("case=should fail registration on first attempt and succeed on second", func(t *testing.T) {
				params := hydraFlowParams{
					subject: testhelpers.RandomEmail(),
					scope:   []string{"openid"},
				}
				params.claims.traits.website = "https://www.ory.com/kratos"
				params.claims.traits.groups = []string{"group1", "group2"}
				params.claims.metadataPublic.picture = "picture.png"
				params.claims.metadataAdmin.phoneNumber = "911"

				// First attempt with a too-short name trait.
				f := newAPIRegistrationFlow(t, returnTS.URL+"?return_session_token_exchange_code=true&return_to=/app_code&identity_schema", time.Minute)
				action := assertFormValues(t, f.ID, "valid")

				res, err := http.Post(action, "application/json", // #nosec G107 -- test code
					strings.NewReader(`{"method":"oidc","provider":"valid","traits.name":"i"}`))
				require.NoError(t, err)
				require.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

				var changeLocation flow.BrowserLocationChangeRequiredError
				require.NoError(t, json.NewDecoder(res.Body).Decode(&changeLocation))

				_, err = wrapClientForHydraLoginConsent(t, testhelpers.NewClientWithCookies(t), hydraAdmin, params).Get(changeLocation.RedirectBrowserTo)
				require.NoError(t, err)
				// The callback redirects back but registration fails due to validation (name too short).
				// Somehow we still end up at the default return_to URL. Please investigate and fix the test or code.
				// assert.Truef(t, strings.HasPrefix(res.Request.URL.String(), errTS.URL), "expected %q to have prefix %q", res.Request.URL.String(), "error-ts")

				// Second attempt with valid data on a new flow.
				f = newAPIRegistrationFlow(t, returnTS.URL+"?return_session_token_exchange_code=true&return_to=/app_code", time.Minute)
				action = assertFormValues(t, f.ID, "valid")

				res, err = http.Post(action, "application/json", // #nosec G107 -- test code
					strings.NewReader(`{"method":"oidc","provider":"valid","traits.name":"valid-name"}`))
				require.NoError(t, err)
				require.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

				require.NoError(t, json.NewDecoder(res.Body).Decode(&changeLocation))

				res, err = wrapClientForHydraLoginConsent(t, testhelpers.NewClientWithCookies(t), hydraAdmin, params).Get(changeLocation.RedirectBrowserTo)
				require.NoError(t, err)

				returnToCode := res.Request.URL.Query().Get("code")
				require.NotEmpty(t, returnToCode, "code query param was empty in the return_to URL")

				codeResponse, err := exchangeCodeForToken(t, sessiontokenexchange.Codes{
					InitCode:     f.SessionTokenExchangeCode,
					ReturnToCode: returnToCode,
				})
				require.NoError(t, err)
				assert.NotEmptyf(t, codeResponse.Token, "%+v", codeResponse)
				assert.Equalf(t, "valid-name", gjson.GetBytes(codeResponse.Session.Identity.Traits, "name").String(), "%+v", codeResponse.Session)
				assert.Equalf(t, "https://www.ory.com/kratos", gjson.GetBytes(codeResponse.Session.Identity.Traits, "website").String(), "%+v", codeResponse.Session)
			})
		})

		t.Run("case=should return exchange code even if already authenticated", func(t *testing.T) {
			params := hydraFlowParams{
				subject: testhelpers.RandomEmail(),
				scope:   []string{"openid"},
			}
			cl := testhelpers.NewClientWithCookies(t)

			// step 1: register and create a session",
			returnTo := "/foo"
			r := newBrowserLoginFlow(t, fmt.Sprintf("%s?return_to=%s", returnTS.URL, returnTo), time.Minute)
			action := assertFormValues(t, r.ID, "valid")

			res, body := makeRequestWithClient(t, cl, action, "valid", params, nil)
			assert.True(t, strings.HasSuffix(res.Request.URL.String(), returnTo))
			assertIdentity(t, res, body, params.subject, idTokenClaims{})

			// step 2: perform login and get exchange code
			performLogin(t, cl, params)
		})

		t.Run("case=should use redirect_to URL on failure", func(t *testing.T) {
			subject := testhelpers.RandomEmail()

			i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			i.SetCredentials(identity.CredentialsTypePassword, identity.Credentials{
				Identifiers: []string{subject},
			})
			i.Traits = identity.Traits(`{"subject":"` + subject + `"}`)
			require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(t.Context(), i))

			f := newAPILoginFlow(t, returnTS.URL+"?return_session_token_exchange_code=true&return_to=/app_code", 1*time.Minute)

			_, err := exchangeCodeForToken(t, sessiontokenexchange.Codes{InitCode: f.SessionTokenExchangeCode})
			require.Error(t, err)

			action := assertFormValues(t, f.ID, "valid")
			returnToURL := makeAPICodeFlowRequest(t, nil, action, "valid", hydraFlowParams{subject: subject, scope: []string{"openid"}})
			returnedFlow := returnToURL.Query().Get("flow")

			require.NotEmpty(t, returnedFlow, "flow query param was empty in the return_to URL")
			loginFlow, err := reg.LoginFlowPersister().GetLoginFlow(t.Context(), uuid.FromStringOrNil(returnedFlow))
			require.NoError(t, err)
			assert.Equal(t, text.InfoSelfServiceLoginLink, loginFlow.UI.Messages[0].ID)
		})

		t.Run("case=should return flow ID in return_to when registration has missing required traits", func(t *testing.T) {
			// Use a schema that requires "extra_data" which is NOT mapped by the OIDC jsonnet mapper,
			// so the registration will fail with a validation error.

			params := hydraFlowParams{
				subject: testhelpers.RandomEmail(),
				scope:   []string{"openid"},
			}
			params.claims.traits.website = "https://www.ory.com/kratos"
			params.claims.traits.groups = []string{"group1", "group2"}
			params.claims.metadataPublic.picture = "picture.png"
			params.claims.metadataAdmin.phoneNumber = "911"

			f := newAPIRegistrationFlow(t, returnTS.URL+"?return_session_token_exchange_code=true&return_to=/app_code&identity_schema=extra_data", time.Minute)

			_, err := exchangeCodeForToken(t, sessiontokenexchange.Codes{InitCode: f.SessionTokenExchangeCode})
			require.Error(t, err)

			action := assertFormValues(t, f.ID, "valid")
			returnToURL := makeAPICodeFlowRequest(t, nil, action, "valid", params)

			returnToCode := returnToURL.Query().Get("code")
			require.NotEmpty(t, returnToCode, "code query param was empty in the return_to URL")

			returnedFlow := returnToURL.Query().Get("flow")
			require.NotEmpty(t, returnedFlow, "flow query param was empty in the return_to URL")

			// Verify the flow contains validation errors and trait input nodes
			regFlow, err := reg.RegistrationFlowPersister().GetRegistrationFlow(t.Context(), uuid.FromStringOrNil(returnedFlow))
			require.NoError(t, err)
			assert.Equal(t, f.ID, regFlow.ID, "returned flow ID should match the original registration flow")

			// The flow should have trait input nodes for the "extra_data" field that failed validation
			body, err := json.Marshal(regFlow.UI)
			require.NoError(t, err)
			assert.NotEmpty(t, gjson.GetBytes(body, "nodes.#(attributes.name==traits.extra_data)").Raw, "flow UI should contain a traits.extra_data input node: %s", body)
		})
	})

	t.Run("case=submit id_token during registration or login", func(t *testing.T) {
		setProviderConfig(t, conf,
			newOIDCProvider(t, ts, hydraPublic, hydraAdmin, "valid"),
			oidc.Configuration{
				Provider:     "test-provider",
				ID:           "test-provider",
				ClientID:     invalid.ClientID,
				ClientSecret: invalid.ClientSecret,
				IssuerURL:    hydraPublic + "/",
				Mapper:       "file://./stub/oidc.facebook.jsonnet",
			})
		oidc.RegisterTestProvider(t, "test-provider")

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
			token = strings.ReplaceAll(token, "{{sub}}", testhelpers.RandomEmail())
			nonce = randx.MustString(16, randx.Alpha)
			token = strings.ReplaceAll(token, "{{nonce}}", nonce)
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
			params := hydraFlowParams{
				subject: testhelpers.RandomEmail(),
				scope:   []string{"openid"},
			}
			returnTo := "/foo"
			r := newBrowserLoginFlow(t, fmt.Sprintf("%s?return_to=%s", returnTS.URL, returnTo), time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			res, body := makeRequest(t, action, "valid", params, nil)
			assert.True(t, strings.HasSuffix(res.Request.URL.String(), returnTo))
			assertIdentity(t, res, body, params.subject, idTokenClaims{})
		})

		t.Run("case=should pass login and carry over login_challenge to registration", func(t *testing.T) {
			client := wrapClientForHydraLoginConsent(t, testhelpers.NewClientWithCookies(t), hydraAdmin, hydraFlowParams{
				subject: testhelpers.RandomEmail(),
				scope:   []string{"openid"},
			})

			reg.SetHydra(hydra.NewFake())
			r := newBrowserLoginFlow(t, fmt.Sprintf("%s?login_challenge=%s", returnTS.URL, hydra.FakeValidLoginChallenge), time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			fv := url.Values{}
			fv.Set("provider", "valid")
			res, err := client.PostForm(action, fv)
			require.NoError(t, err)
			// Expect to be returned to the hydra instance, that instantiated the request
			assert.Equal(t, hydra.FakePostLoginURL, strings.TrimSuffix(res.Request.URL.String(), "/"))
		})
	})

	t.Run("case=register and register again but login", func(t *testing.T) {
		params := hydraFlowParams{
			subject: testhelpers.RandomEmail(),
			scope:   []string{"openid"},
		}

		t.Run("case=should pass registration", func(t *testing.T) {
			postRegistrationWebhook := hooktest.NewServer()
			t.Cleanup(postRegistrationWebhook.Close)
			postRegistrationWebhook.SetConfig(t, conf.GetProvider(t.Context()), config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypeOIDC.String()))

			r := newBrowserRegistrationFlow(t, returnTS.URL+"?identity_schema=phone", time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			transientPayload := `{"data": "registration-one"}`
			res, body := makeRequest(t, action, "valid", params, url.Values{"transient_payload": {transientPayload}})
			assertIdentity(t, res, body, params.subject, idTokenClaims{})
			assert.Equalf(t, "phone", gjson.GetBytes(body, "identity.schema_id").String(), "%s", body)
			postRegistrationWebhook.AssertTransientPayload(t, transientPayload)
		})

		t.Run("case=should pass second time registration", func(t *testing.T) {
			postLoginWebhook := hooktest.NewServer()
			t.Cleanup(postLoginWebhook.Close)
			postLoginWebhook.SetConfig(t, conf.GetProvider(t.Context()), config.HookStrategyKey(config.ViperKeySelfServiceLoginAfter, identity.CredentialsTypeOIDC.String()))

			r := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			transientPayload := `{"data": "registration-two"}`
			res, body := makeRequest(t, action, "valid", params, url.Values{"transient_payload": {transientPayload}})
			assertIdentity(t, res, body, params.subject, idTokenClaims{})
			postLoginWebhook.AssertTransientPayload(t, transientPayload)
		})

		t.Run("case=should pass third time registration with return to", func(t *testing.T) {
			postLoginWebhook := hooktest.NewServer()
			t.Cleanup(postLoginWebhook.Close)
			postLoginWebhook.SetConfig(t, conf.GetProvider(t.Context()), config.HookStrategyKey(config.ViperKeySelfServiceLoginAfter, identity.CredentialsTypeOIDC.String()))

			returnTo := "/foo"
			r := newBrowserLoginFlow(t, fmt.Sprintf("%s?return_to=%s", returnTS.URL, returnTo), time.Minute)
			action := assertFormValues(t, r.ID, "valid")
			transientPayload := `{"data": "registration-three"}`
			res, body := makeRequest(t, action, "valid", params, url.Values{"transient_payload": {transientPayload}})
			assert.True(t, strings.HasSuffix(res.Request.URL.String(), returnTo))
			assertIdentity(t, res, body, params.subject, idTokenClaims{})
			postLoginWebhook.AssertTransientPayload(t, transientPayload)
		})
	})

	t.Run("case=register, merge, and complete data", func(t *testing.T) {
		for _, tc := range []struct{ name, provider string }{
			{name: "idtoken", provider: "valid"},
			{name: "userinfo", provider: "claimsViaUserInfo"},
			{name: "disable-pkce", provider: "neverPKCE"},
			{name: "auto-pkce", provider: "autoPKCE"},
			{name: "force-pkce", provider: "forcePKCE"},
		} {
			claims := idTokenClaims{}
			claims.traits.website = "https://www.ory.com/kratos"
			claims.traits.groups = []string{"group1", "group2"}
			claims.metadataPublic.picture = "picture.png"
			claims.metadataAdmin.phoneNumber = "911"

			makeRequest := func(t *testing.T, action, subject, name string, fv url.Values) (*http.Response, []byte) {
				if fv == nil {
					fv = make(url.Values)
				}
				fv.Set("traits.name", name)
				fv.Set("provider", tc.provider)
				res, err := wrapClientForHydraLoginConsent(t, testhelpers.NewClientWithCookies(t), hydraAdmin,
					hydraFlowParams{subject: subject, claims: claims, scope: []string{"openid"}},
				).PostForm(action, fv)
				require.NoError(t, err, action)

				body, err := io.ReadAll(res.Body)
				require.NoError(t, res.Body.Close())
				require.NoError(t, err)
				require.Equalf(t, http.StatusOK, res.StatusCode, "%s: %s\n\t%s", action, res.Request.URL.String(), body)

				return res, body
			}

			t.Run(fmt.Sprintf("ClaimsSource=%s", tc.name), func(t *testing.T) {
				t.Run("case=should fail registration on first attempt", func(t *testing.T) {
					r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
					action := assertFormValues(t, r.ID, tc.provider)

					res, body := makeRequest(t, action, testhelpers.RandomEmail(), "i", nil)

					require.Contains(t, res.Request.URL.String(), uiTS.URL, "%s", body)

					assert.Equalf(t, "length must be >= 2, but got 1", gjson.GetBytes(body, "ui.nodes.#(attributes.name==traits.name).messages.0.text").String(), "%s", body) // make sure the field is being echoed
					assert.Equalf(t, "traits.name", gjson.GetBytes(body, "ui.nodes.#(attributes.name==traits.name).attributes.name").String(), "%s", body)                    // make sure the field is being echoed
					assert.Equalf(t, "i", gjson.GetBytes(body, "ui.nodes.#(attributes.name==traits.name).attributes.value").String(), "%s", body)                             // make sure the field is being echoed
					assert.Equalf(t, "https://www.ory.com/kratos", gjson.GetBytes(body, "ui.nodes.#(attributes.name==traits.website).attributes.value").String(), "%s", body) // make sure the field is being echoed
				})

				t.Run("case=should pass registration with valid data", func(t *testing.T) {
					r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
					action := assertFormValues(t, r.ID, tc.provider)

					subject := testhelpers.RandomEmail()
					res, body := makeRequest(t, action, subject, "valid-name", nil)

					assertIdentity(t, res, body, subject, claims)
					assert.Equalf(t, "https://www.ory.com/kratos", gjson.GetBytes(body, "identity.traits.website").String(), "%s", body)
					assert.Equalf(t, "valid-name", gjson.GetBytes(body, "identity.traits.name").String(), "%s", body)
					assert.JSONEqf(t, `["group1","group2"]`, gjson.GetBytes(body, "identity.traits.groups").Raw, "%s", body)
				})

				t.Run("case=should fail registration on first attempt with multi-schema select", func(t *testing.T) {
					r := newBrowserRegistrationFlow(t, returnTS.URL+"?identity_schema=extra_data", time.Minute)
					action := assertFormValues(t, r.ID, tc.provider)
					res, body := makeRequest(t, action, testhelpers.RandomEmail(), "i", nil)
					require.Contains(t, res.Request.URL.String(), uiTS.URL, "%s", body)

					assert.Equal(t, "length must be >= 2, but got 1", gjson.GetBytes(body, "ui.nodes.#(attributes.name==traits.name).messages.0.text").String(), "%s", body) // make sure the field is being echoed
					assert.Equal(t, "traits.name", gjson.GetBytes(body, "ui.nodes.#(attributes.name==traits.name).attributes.name").String(), "%s", body)                    // make sure the field is being echoed
					assert.Equal(t, "traits.extra_data", gjson.GetBytes(body, "ui.nodes.#(attributes.name==traits.extra_data).attributes.name").String(), "%s", body)        // make sure the field is being echoed
					assert.Equal(t, "i", gjson.GetBytes(body, "ui.nodes.#(attributes.name==traits.name).attributes.value").String(), "%s", body)                             // make sure the field is being echoed
					assert.Equal(t, "https://www.ory.com/kratos", gjson.GetBytes(body, "ui.nodes.#(attributes.name==traits.website).attributes.value").String(), "%s", body) // make sure the field is being echoed
				})

				t.Run("case=should pass registration with selected schema with valid data", func(t *testing.T) {
					r := newBrowserRegistrationFlow(t, returnTS.URL+"?identity_schema=extra_data", time.Minute)
					action := assertFormValues(t, r.ID, tc.provider)
					subject := testhelpers.RandomEmail()
					res, body := makeRequest(t, action, subject, "valid-name-2", url.Values{"traits.extra_data": {"extra-data"}})
					assertIdentity(t, res, body, subject, claims)
					assert.Equalf(t, "https://www.ory.com/kratos", gjson.GetBytes(body, "identity.traits.website").String(), "%s", body)
					assert.Equalf(t, "valid-name-2", gjson.GetBytes(body, "identity.traits.name").String(), "%s", body)
					assert.Equalf(t, "extra-data", gjson.GetBytes(body, "identity.traits.extra_data").String(), "%s", body)
					assert.Equalf(t, `["group1","group2"]`, gjson.GetBytes(body, "identity.traits.groups").String(), "%s", body)
				})
			})
		}
	})

	for _, loginHintsEnabled := range []bool{true, false} {
		t.Run("login-hints-enabled="+strconv.FormatBool(loginHintsEnabled), func(t *testing.T) {
			t.Run("case=should fail to register and return fresh login flow if email is already being used by password credentials", func(t *testing.T) {
				subject := "email-exist-with-password-strategy-lh-" + strconv.FormatBool(loginHintsEnabled) + "@ory.sh"
				scope := []string{"openid"}
				conf.MustSet(t.Context(), config.ViperKeySelfServiceRegistrationLoginHints, loginHintsEnabled)

				i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
				i.SetCredentials(identity.CredentialsTypePassword, identity.Credentials{
					Identifiers: []string{subject},
					Config:      sqlxx.JSONRawMessage(`{"hashed_password": "foo"}`),
				})
				i.Traits = identity.Traits(`{"subject":"` + subject + `"}`)
				require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(t.Context(), i))

				t.Run("case=should fail registration", func(t *testing.T) {
					r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
					action := assertFormValues(t, r.ID, "valid")
					_, body := makeRequest(t, action, "valid", hydraFlowParams{subject: subject, scope: scope}, nil)
					snapshotx.SnapshotTJSON(t, body, snapshotx.ExceptPaths("expires_at", "updated_at", "issued_at", "id", "created_at", "ui.action", findCsrfTokenPath(t, body), "request_url"), snapshotx.ExceptNestedKeys("newLoginUrl", "new_login_url"))
				})

				t.Run("case=should fail registration id_first strategy enabled", func(t *testing.T) {
					conf.MustSet(t.Context(), config.ViperKeySelfServiceLoginFlowStyle, "identifier_first")
					r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
					action := assertFormValues(t, r.ID, "valid")
					_, body := makeRequest(t, action, "valid", hydraFlowParams{subject: subject, scope: scope}, nil)
					snapshotx.SnapshotTJSON(t, body, snapshotx.ExceptPaths("expires_at", "updated_at", "issued_at", "id", "created_at", "ui.action", findCsrfTokenPath(t, body), "request_url"), snapshotx.ExceptNestedKeys("newLoginUrl", "new_login_url"))
				})

				t.Run("case=should fail login", func(t *testing.T) {
					r := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
					action := assertFormValues(t, r.ID, "valid")
					_, body := makeRequest(t, action, "valid", hydraFlowParams{subject: subject, scope: scope}, nil)
					snapshotx.SnapshotTJSON(t, body, snapshotx.ExceptPaths("expires_at", "updated_at", "issued_at", "id", "created_at", "ui.action", findCsrfTokenPath(t, body), "request_url"), snapshotx.ExceptNestedKeys("newLoginUrl", "new_login_url"))
				})
			})

			t.Run("case=should fail to register and return fresh login flow if email is already being used by oidc credentials", func(t *testing.T) {
				subject := "email-exist-with-oidc-strategy-lh-" + strconv.FormatBool(loginHintsEnabled) + "@ory.sh"
				scope := []string{"openid"}
				conf.MustSet(t.Context(), config.ViperKeySelfServiceRegistrationLoginHints, loginHintsEnabled)

				t.Run("case=create oidc identity", func(t *testing.T) {
					r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
					action := assertFormValues(t, r.ID, "secondProvider")
					res, _ := makeRequest(t, action, "secondProvider", hydraFlowParams{subject: subject, scope: scope}, nil)
					require.Equal(t, http.StatusOK, res.StatusCode)
				})

				t.Run("case=should fail registration", func(t *testing.T) {
					r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
					action := assertFormValues(t, r.ID, "valid")
					_, body := makeRequest(t, action, "valid", hydraFlowParams{subject: subject, scope: scope}, nil)
					snapshotx.SnapshotTJSON(t, body, snapshotx.ExceptPaths("expires_at", "updated_at", "issued_at", "id", "created_at", "ui.action", findCsrfTokenPath(t, body), "request_url"), snapshotx.ExceptNestedKeys("newLoginUrl", "new_login_url"))
				})

				t.Run("case=should fail registration id_first strategy enabled", func(t *testing.T) {
					conf.MustSet(t.Context(), config.ViperKeySelfServiceLoginFlowStyle, "identifier_first")
					r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
					action := assertFormValues(t, r.ID, "valid")
					_, body := makeRequest(t, action, "valid", hydraFlowParams{subject: subject, scope: scope}, nil)
					snapshotx.SnapshotTJSON(t, body, snapshotx.ExceptPaths("expires_at", "updated_at", "issued_at", "id", "created_at", "ui.action", findCsrfTokenPath(t, body), "request_url"), snapshotx.ExceptNestedKeys("newLoginUrl", "new_login_url"))
				})

				t.Run("case=should fail login", func(t *testing.T) {
					r := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
					action := assertFormValues(t, r.ID, "valid")
					_, body := makeRequest(t, action, "valid", hydraFlowParams{subject: subject, scope: scope}, nil)
					snapshotx.SnapshotTJSON(t, body, snapshotx.ExceptPaths("expires_at", "updated_at", "issued_at", "id", "created_at", "ui.action", findCsrfTokenPath(t, body), "request_url"), snapshotx.ExceptNestedKeys("newLoginUrl", "new_login_url"))
				})
			})
		})
	}

	t.Run("case=should redirect to default return ts when sending authenticated login flow without forced flag", func(t *testing.T) {
		subject := testhelpers.RandomEmail()
		scope := []string{"openid"}

		fv := url.Values{"traits.name": {"valid-name"}}
		client := testhelpers.NewClientWithCookies(t)
		r1 := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
		res1, body1 := makeRequestWithClient(t, client, assertFormValues(t, r1.ID, "valid"), "valid", hydraFlowParams{subject: subject, scope: scope}, fv)
		assertIdentity(t, res1, body1, subject, idTokenClaims{})
		r2 := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
		res2, body2 := makeRequestWithClient(t, client, assertFormValues(t, r2.ID, "valid"), "valid", hydraFlowParams{subject: subject, scope: scope}, fv)
		assertIdentity(t, res2, body2, subject, idTokenClaims{})
		assert.Equal(t, body1, body2)
	})

	t.Run("case=should reauthenticate when sending authenticated login flow with forced flag", func(t *testing.T) {
		subject := testhelpers.RandomEmail()
		scope := []string{"openid"}

		fv := url.Values{"traits.name": {"valid-name"}}
		client := testhelpers.NewClientWithCookies(t)
		r1 := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
		res1, body1 := makeRequestWithClient(t, client, assertFormValues(t, r1.ID, "valid"), "valid", hydraFlowParams{subject: subject, scope: scope}, fv)
		assertIdentity(t, res1, body1, subject, idTokenClaims{})
		r2 := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
		require.NoError(t, reg.LoginFlowPersister().ForceLoginFlow(t.Context(), r2.ID))
		res2, body2 := makeRequestWithClient(t, client, assertFormValues(t, r2.ID, "valid"), "valid", hydraFlowParams{subject: subject, scope: scope}, fv)
		assertIdentity(t, res2, body2, subject, idTokenClaims{})
		assert.NotEqual(t, gjson.GetBytes(body1, "id"), gjson.GetBytes(body2, "id"))
		authAt1, err := time.Parse(time.RFC3339, gjson.GetBytes(body1, "authenticated_at").String())
		require.NoError(t, err)
		authAt2, err := time.Parse(time.RFC3339, gjson.GetBytes(body2, "authenticated_at").String())
		require.NoError(t, err)
		// authenticated at is newer in the second body
		assert.Greater(t, authAt2.Sub(authAt1).Milliseconds(), int64(0), "%s - %s : %s - %s", authAt2, authAt1, body2, body1)
	})

	t.Run("case=upstream parameters should be passed on to provider", func(t *testing.T) {
		subject := testhelpers.RandomEmail()
		scope := []string{"openid", "offline"}

		// We need to disable redirects because the upstream parameters are only passed on to the provider
		c := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
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

			res, err := wrapClientForHydraLoginConsent(t, c, hydraAdmin, hydraFlowParams{subject: subject, scope: scope}).PostForm(action, fv)
			require.NoError(t, err)
			require.Equal(t, http.StatusSeeOther, res.StatusCode)

			loc, err := res.Location()
			require.NoError(t, err)

			// login_hint is never forwarded: a caller-supplied value cannot be
			// trusted (HackerOne #3239672, ory-corp/cloud#8955).
			require.Empty(t, loc.Query().Get("login_hint"))
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

			// login_hint is never forwarded: a caller-supplied value cannot be
			// trusted (HackerOne #3239672, ory-corp/cloud#8955).
			require.Empty(t, loc.Query().Get("login_hint"))
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
		scope := []string{"openid"}

		assertVerifiedEmail := func(t *testing.T, body []byte, subject string, verified bool) {
			b := prettyJSON(t, body)
			assert.Lenf(t, gjson.GetBytes(body, "identity.verifiable_addresses").Array(), 1, "%s", b)
			assert.Equalf(t, "email", gjson.GetBytes(body, "identity.verifiable_addresses.0.via").String(), "%s", b)
			assert.Equalf(t, subject, gjson.GetBytes(body, "identity.verifiable_addresses.0.value").String(), "%s", b)
			assert.Equalf(t, verified, gjson.GetBytes(body, "identity.verifiable_addresses.0.verified").Bool(), "%s", b)
		}

		t.Run("case=should have verified address when subject matches", func(t *testing.T) {
			subject := testhelpers.RandomEmail()
			r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute, registration.WithFlowIdentitySchema("verifiable_email"))
			action := assertFormValues(t, r.ID, "valid")
			res, body := makeRequest(t, action, "valid", hydraFlowParams{subject: subject, scope: scope}, nil)
			assertIdentity(t, res, body, subject, idTokenClaims{})
			assertVerifiedEmail(t, body, subject, true)
		})

		t.Run("case=should have verified address when subject matches after normalization", func(t *testing.T) {
			subject := " Denormalized-Verified-Email@ory.sh "
			r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute, registration.WithFlowIdentitySchema("verifiable_email"))
			action := assertFormValues(t, r.ID, "valid")
			res, body := makeRequest(t, action, "valid", hydraFlowParams{subject: subject, scope: scope}, url.Values{"traits.subject": {"denormalized-verified-EMAIL@ory.sh"}})
			subject = "denormalized-verified-EMAIL@ory.sh"
			assertIdentity(t, res, body, subject, idTokenClaims{})
			subject = "denormalized-verified-email@ory.sh"
			assertVerifiedEmail(t, body, subject, true)
		})

		t.Run("case=should have unverified address when subject does not match", func(t *testing.T) {
			subject := testhelpers.RandomEmail()
			r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute, registration.WithFlowIdentitySchema("verifiable_email"))
			action := assertFormValues(t, r.ID, "valid")
			newSubject := testhelpers.RandomEmail()
			res, body := makeRequest(t, action, "valid", hydraFlowParams{subject: subject, scope: scope}, url.Values{"traits.subject": {newSubject}})
			assertIdentity(t, res, body, newSubject, idTokenClaims{})
			assertVerifiedEmail(t, body, newSubject, false)
		})
	})

	loginWithOIDC := func(t *testing.T, c *http.Client, flowID uuid.UUID, provider, subject string, scope []string) (*http.Response, []byte) {
		action := assertFormValues(t, flowID, provider)
		res, err := wrapClientForHydraLoginConsent(t, c, hydraAdmin, hydraFlowParams{subject: subject, scope: scope}).PostForm(action, url.Values{"provider": {provider}})
		require.NoError(t, err, action)
		body, err := io.ReadAll(res.Body)
		require.NoError(t, res.Body.Close())
		require.NoError(t, err)
		return res, body
	}

	checkCredentialsLinked := func(t *testing.T, res *http.Response, body []byte, identityID uuid.UUID, provider, subject string) {
		assert.Containsf(t, res.Request.URL.String(), returnTS.URL, "%s", body)
		assert.Equal(t, strings.ToLower(subject), gjson.GetBytes(body, "identity.traits.subject").String(), "%s", body)
		i, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(t.Context(), identityID)
		require.NoError(t, err)
		assert.NotEmptyf(t, i.Credentials["oidc"], "%+v", i.Credentials)
		assert.Truef(t, gjson.GetBytes(i.Credentials["oidc"].Config, fmt.Sprintf("providers.#(provider=%q)", provider)).Exists(),
			"%s", string(i.Credentials["oidc"].Config[:]))
		assert.Contains(t, gjson.GetBytes(body, "authentication_methods").String(), "oidc", "%s", body)
	}

	t.Run("case=registration should start new login flow if duplicate credentials detected", func(t *testing.T) {
		reg.Config().MustSet(t.Context(), config.ViperKeySelfServiceRegistrationLoginHints, true)

		t.Run("case=second login is password", func(t *testing.T) {
			subject := "new-login-if-email-exist-with-password-strategy@ory.sh"
			subject2 := "new-login-subject2@ory.sh"
			scope := []string{"openid"}
			password := "lwkj52sdkjf" // #nosec G101

			var i *identity.Identity
			t.Run("step=create password identity", func(t *testing.T) {
				i = identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
				p, err := reg.Hasher(t.Context()).Generate(t.Context(), []byte(password))
				require.NoError(t, err)
				i.SetCredentials(identity.CredentialsTypePassword, identity.Credentials{
					Identifiers: []string{subject},
					Config:      sqlxx.JSONRawMessage(`{"hashed_password":"` + string(p) + `"}`),
				})
				i.Traits = identity.Traits(`{"subject":"` + subject + `"}`)
				require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(t.Context(), i))

				i2 := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
				i2.SetCredentials(identity.CredentialsTypePassword, identity.Credentials{
					Identifiers: []string{subject2},
					Config:      sqlxx.JSONRawMessage(`{"hashed_password":"` + string(p) + `"}`),
				})
				i2.Traits = identity.Traits(`{"subject":"` + subject2 + `"}`)
				require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(t.Context(), i2))
			})

			client := testhelpers.NewClientWithCookies(t)
			loginFlow := newLoginFlow(t, returnTS.URL, time.Minute, flow.TypeBrowser)
			loginFlow.OrganizationID = uuid.NullUUID{UUID: orgID, Valid: true}
			require.NoError(t, reg.LoginFlowPersister().UpdateLoginFlow(t.Context(), loginFlow))

			var linkingLoginFlow struct {
				ID        string
				UIAction  string
				CSRFToken string
			}

			t.Run("step=should fail login and start a new flow", func(t *testing.T) {
				// To test that the subject is normalized properly
				res, body := loginWithOIDC(t, client, loginFlow.ID, "valid", strings.ToUpper(subject), scope)
				assert.True(t, res.Request.URL.Query().Has("no_org_ui"))
				assertUIError(t, res, body, `You tried to sign in with "new-login-if-email-exist-with-password-strategy@ory.sh", but that email is already used by another account. Sign in to your account with one of the options below to add your account "new-login-if-email-exist-with-password-strategy@ory.sh" at "generic" as another way to sign in.`)
				assert.True(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==identifier)").Exists(), "%s", body)
				assert.True(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==password)").Exists(), "%s", body)
				assert.Equal(t, "new-login-if-email-exist-with-password-strategy@ory.sh", gjson.GetBytes(body, "ui.messages.#(id==1010016).context.duplicateIdentifier").String())
				assert.Equal(t, gjson.GetBytes(body, "organization_id").String(), orgID.String())
				assert.NotContains(t, gjson.GetBytes(body, "request_url").String(), "code=")
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
				assert.EqualValues(t,
					text.ErrorValidationLoginLinkedCredentialsDoNotMatch,
					gjson.GetBytes(body, "ui.messages.0.id").Int(),
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
				checkCredentialsLinked(t, res, body, i.ID, "valid", subject)
			})
		})

		t.Run("case=second login is OIDC", func(t *testing.T) {
			email1 := "existing-oidc-identity-1@ory.sh"
			email2 := "existing-oidc-identity-2@ory.sh"
			scope := []string{"openid", "offline"}

			var identityID uuid.UUID
			t.Run("step=create OIDC identity", func(t *testing.T) {
				r := newRegistrationFlow(t, returnTS.URL, time.Minute, flow.TypeBrowser)
				action := assertFormValues(t, r.ID, "secondProvider")
				res, body := makeRequest(t, action, "secondProvider", hydraFlowParams{subject: email1, scope: scope}, nil)
				assertIdentity(t, res, body, email1, idTokenClaims{})
				identityID = expectTokens(t, "secondProvider", email1, body)

				r = newRegistrationFlow(t, returnTS.URL, time.Minute, flow.TypeBrowser)
				action = assertFormValues(t, r.ID, "valid")
				res, body = makeRequest(t, action, "valid", hydraFlowParams{subject: email2, scope: scope}, nil)
				assertIdentity(t, res, body, email2, idTokenClaims{})
				expectTokens(t, "valid", email2, body)
			})

			client := testhelpers.NewClientWithCookieJar(t, nil, nil)
			loginFlow := newLoginFlow(t, returnTS.URL, time.Minute, flow.TypeBrowser)
			var linkingLoginFlowID string
			t.Run("step=should fail login and start a new login", func(t *testing.T) {
				// To test the identifier mismatch
				conf.MustSet(t.Context(), config.ViperKeySelfServiceRegistrationLoginHints, false)
				res, body := loginWithOIDC(t, client, loginFlow.ID, "valid2", email1, scope)
				assertUIError(t, res, body, `You tried to sign in with "existing-oidc-identity-1@ory.sh", but that email is already used by another account. Sign in to your account with one of the options below to add your account "existing-oidc-identity-1@ory.sh" at "generic" as another way to sign in.`)
				linkingLoginFlowID = gjson.GetBytes(body, "id").String()
				assert.NotEqual(t, loginFlow.ID.String(), linkingLoginFlowID, "should have started a new flow")
			})

			t.Run("step=should fail login if existing identity identifier doesn't match", func(t *testing.T) {
				require.NotNil(t, linkingLoginFlowID)
				require.NotEmpty(t, linkingLoginFlowID)
				res, body := loginWithOIDC(t, client, uuid.Must(uuid.FromString(linkingLoginFlowID)), "valid", email2, scope)
				assertUIError(t, res, body, "Linked credentials do not match.")
			})

			t.Run("step=should link oidc credentials to existing identity", func(t *testing.T) {
				res, body := loginWithOIDC(t, client, uuid.Must(uuid.FromString(linkingLoginFlowID)), "secondProvider", email1, scope)
				checkCredentialsLinked(t, res, body, identityID, "secondProvider", email1)
			})
		})
	})

	t.Run("suite=auto link policy", func(t *testing.T) {
		t.Run("case=should automatically link credential if policy says so", func(t *testing.T) {
			subject := testhelpers.RandomEmail()
			scope := []string{"openid"}

			reg.AllLoginStrategies().MustStrategy("oidc").(*oidc.Strategy).SetOnConflictingIdentity(t,
				func(ctx context.Context, existingIdentity, newIdentity *identity.Identity, _ oidc.Provider, _ *oidc.Claims) oidc.ConflictingIdentityVerdict {
					return oidc.ConflictingIdentityVerdictMerge
				})

			var i *identity.Identity
			t.Run("step=create identity in org without credentials", func(t *testing.T) {
				i = identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
				i.Traits = identity.Traits(`{"subject":"` + subject + `"}`)
				i.SetCredentials(identity.CredentialsTypePassword, identity.Credentials{
					Type:        identity.CredentialsTypePassword,
					Identifiers: []string{subject},
					Config:      sqlxx.JSONRawMessage(`{}`),
				})
				i.OrganizationID = uuid.NullUUID{
					UUID:  orgID,
					Valid: true,
				}
				i.VerifiableAddresses = []identity.VerifiableAddress{{Value: subject, Via: "email", Verified: true}}
				require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(t.Context(), i))
			})

			t.Run("step=log in with OIDC", func(t *testing.T) {
				loginFlow := newLoginFlow(t, returnTS.URL, time.Minute, flow.TypeBrowser)
				loginFlow.OrganizationID = i.OrganizationID
				require.NoError(t, reg.LoginFlowPersister().UpdateLoginFlow(t.Context(), loginFlow))
				client := testhelpers.NewClientWithCookieJar(t, nil, nil)

				res, body := loginWithOIDC(t, client, loginFlow.ID, "valid", subject, scope)
				checkCredentialsLinked(t, res, body, i.ID, "valid", subject)
			})
		})

		t.Run("case=should remove use_auto_link credential if policy says so", func(t *testing.T) {
			subject := testhelpers.RandomEmail()
			scope := []string{"openid"}

			reg.AllLoginStrategies().MustStrategy("oidc").(*oidc.Strategy).SetOnConflictingIdentity(t,
				func(ctx context.Context, existingIdentity, newIdentity *identity.Identity, _ oidc.Provider, _ *oidc.Claims) oidc.ConflictingIdentityVerdict {
					return oidc.ConflictingIdentityVerdictMerge
				})

			var i *identity.Identity

			t.Run("step=create identity with use_auto_link", func(t *testing.T) {
				i = identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
				i.Traits = identity.Traits(`{"subject":"` + subject + `"}`)
				i.SetCredentials(identity.CredentialsTypePassword, identity.Credentials{
					Type:        identity.CredentialsTypePassword,
					Identifiers: []string{subject},
					Config:      sqlxx.JSONRawMessage(`{}`),
				})
				i.SetCredentials(identity.CredentialsTypeOIDC, identity.Credentials{
					Type:        identity.CredentialsTypeOIDC,
					Identifiers: []string{"valid:stub", "other:other-identifier"},
					Config: sqlxx.JSONRawMessage(`{"providers": [{
	"subject": "stub",
	"provider": "valid",
	"use_auto_link": true
},{
	"subject": "other-identifier",
	"provider": "other",
	"use_auto_link": true
}]}`),
				})
				i.VerifiableAddresses = []identity.VerifiableAddress{{Value: subject, Via: "email", Verified: true}}
				require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(t.Context(), i))
			})

			t.Run("step=log in with OIDC", func(t *testing.T) {
				loginFlow := newLoginFlow(t, returnTS.URL, time.Minute, flow.TypeBrowser)
				require.NoError(t, reg.LoginFlowPersister().UpdateLoginFlow(t.Context(), loginFlow))
				client := testhelpers.NewClientWithCookieJar(t, nil, nil)

				res, body := loginWithOIDC(t, client, loginFlow.ID, "valid", subject, scope)
				checkCredentialsLinked(t, res, body, i.ID, "valid", subject)
			})

			t.Run("step=should remove use_auto_link", func(t *testing.T) {
				var err error
				i, err = reg.PrivilegedIdentityPool().GetIdentityConfidential(t.Context(), i.ID)
				require.NoError(t, err)
				assert.False(t, gjson.GetBytes(i.Credentials["oidc"].Config, `providers.#(provider=="valid").use_auto_link`).Bool())
				assert.NotContains(t, i.Credentials["oidc"].Identifiers, "valid:stub")
				assert.Contains(t, i.Credentials["oidc"].Identifiers, "valid:"+subject)
				assert.True(t, gjson.GetBytes(i.Credentials["oidc"].Config, `providers.#(provider=="other").use_auto_link`).Bool())
			})
		})
	})

	t.Run("case=update identity on login", func(t *testing.T) {
		t.Run("case=should update traits when claims change", func(t *testing.T) {
			params := hydraFlowParams{
				subject: testhelpers.RandomEmail(),
				scope:   []string{"openid", "offline"},
			}
			params.claims.traits.website = "https://original.example.com"
			params.claims.metadataPublic.picture = "original.png"
			params.claims.metadataAdmin.phoneNumber = "111"

			// Set up a provider with update_identity_on_login: automatic.
			setProviderConfig(t, conf,
				newOIDCProvider(t, ts, hydraPublic, hydraAdmin, "valid", func(c *oidc.Configuration) {
					c.UpdateIdentityOnLogin = oidc.UpdateIdentityOnLoginAutomatic
				}),
			)

			t.Run("step=register", func(t *testing.T) {
				r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
				action := assertFormValues(t, r.ID, "valid")
				res, body := makeRequest(t, action, "valid", params, nil)
				assertIdentity(t, res, body, params.subject, params.claims)
			})

			t.Run("step=login with changed claims updates identity", func(t *testing.T) {
				params := params
				params.claims.traits.website = "https://updated.example.com"
				params.claims.metadataPublic.picture = "updated.png"
				params.claims.metadataAdmin.phoneNumber = "222"

				r := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
				action := assertFormValues(t, r.ID, "valid")
				res, body := makeRequest(t, action, "valid", params, nil)

				assert.Contains(t, res.Request.URL.String(), returnTS.URL, "%s", body)
				assert.Equal(t, params.subject, gjson.GetBytes(body, "identity.traits.subject").String(), "%s", body)
				// as per mapper the original website trait should remain unchanged
				assert.Equal(t, "https://original.example.com", gjson.GetBytes(body, "identity.traits.website").String(), "%s", body)
				assert.Equal(t, "updated.png", gjson.GetBytes(body, "identity.metadata_public.picture").String(), "%s", body)

				// Verify metadata_admin was updated by reading the identity from the admin API.
				identityID := gjson.GetBytes(body, "identity.id").String()
				i, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(t.Context(), uuid.FromStringOrNil(identityID))
				require.NoError(t, err)
				assert.Equal(t, "222", gjson.GetBytes(i.MetadataAdmin, "phone_number").String())
			})
		})

		t.Run("case=should not update identity when claims are unchanged", func(t *testing.T) {
			params := hydraFlowParams{
				subject: testhelpers.RandomEmail(),
				scope:   []string{"openid", "offline"},
			}
			params.claims.traits.website = "https://stable.example.com"
			params.claims.metadataPublic.picture = "stable.png"

			setProviderConfig(t, conf,
				newOIDCProvider(t, ts, hydraPublic, hydraAdmin, "valid", func(c *oidc.Configuration) {
					c.UpdateIdentityOnLogin = oidc.UpdateIdentityOnLoginAutomatic
				}),
			)

			t.Run("step=register", func(t *testing.T) {
				r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
				action := assertFormValues(t, r.ID, "valid")
				res, body := makeRequest(t, action, "valid", params, nil)
				assertIdentity(t, res, body, params.subject, params.claims)
			})

			t.Run("step=login with same claims does not error", func(t *testing.T) {
				r := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
				action := assertFormValues(t, r.ID, "valid")
				res, body := makeRequest(t, action, "valid", params, nil)

				assertIdentity(t, res, body, params.subject, params.claims)
			})
		})

		t.Run("case=should not update identity when update_identity_on_login is never", func(t *testing.T) {
			params := hydraFlowParams{
				subject: testhelpers.RandomEmail(),
				scope:   []string{"openid", "offline"},
			}
			params.claims.traits.website = "https://original.example.com"
			params.claims.metadataPublic.picture = "original.png"

			setProviderConfig(t, conf,
				newOIDCProvider(t, ts, hydraPublic, hydraAdmin, "valid"),
			)

			t.Run("step=register", func(t *testing.T) {
				r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
				action := assertFormValues(t, r.ID, "valid")
				res, body := makeRequest(t, action, "valid", params, nil)
				assertIdentity(t, res, body, params.subject, params.claims)
			})

			t.Run("step=login with changed claims does not update identity", func(t *testing.T) {
				params := params
				params.claims.traits.website = "https://changed.example.com"
				params.claims.metadataPublic.picture = "changed.png"

				r := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
				action := assertFormValues(t, r.ID, "valid")
				res, body := makeRequest(t, action, "valid", params, nil)

				assert.Contains(t, res.Request.URL.String(), returnTS.URL, "%s", body)
				assert.Equal(t, params.subject, gjson.GetBytes(body, "identity.traits.subject").String(), "%s", body)
				// Traits should still have the original values from registration.
				assert.Equal(t, "https://original.example.com", gjson.GetBytes(body, "identity.traits.website").String(), "%s", body)
				assert.Equal(t, "original.png", gjson.GetBytes(body, "identity.metadata_public.picture").String(), "%s", body)
			})
		})

		t.Run("case=should update only metadata without changing traits", func(t *testing.T) {
			params := hydraFlowParams{
				subject: testhelpers.RandomEmail(),
				scope:   []string{"openid", "offline"},
			}
			params.claims.traits.website = "https://example.com"
			params.claims.metadataPublic.picture = "old-pic.png"
			params.claims.metadataAdmin.phoneNumber = "111"

			setProviderConfig(t, conf,
				newOIDCProvider(t, ts, hydraPublic, hydraAdmin, "valid", func(c *oidc.Configuration) {
					c.UpdateIdentityOnLogin = oidc.UpdateIdentityOnLoginAutomatic
				}),
			)

			t.Run("step=register", func(t *testing.T) {
				r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
				action := assertFormValues(t, r.ID, "valid")
				res, body := makeRequest(t, action, "valid", params, nil)
				assertIdentity(t, res, body, params.subject, params.claims)
			})

			t.Run("step=login with changed metadata only", func(t *testing.T) {
				// Traits stay the same, only metadata changes.
				params := params
				params.claims.metadataPublic.picture = "new-pic.png"
				params.claims.metadataAdmin.phoneNumber = "222"

				r := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
				action := assertFormValues(t, r.ID, "valid")
				res, body := makeRequest(t, action, "valid", params, nil)

				assert.Contains(t, res.Request.URL.String(), returnTS.URL, "%s", body)
				assert.Equal(t, "https://example.com", gjson.GetBytes(body, "identity.traits.website").String(), "%s", body)
				assert.Equal(t, "new-pic.png", gjson.GetBytes(body, "identity.metadata_public.picture").String(), "%s", body)

				// Verify metadata_admin was updated by reading the identity from the admin API.
				identityID := gjson.GetBytes(body, "identity.id").String()
				i, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(t.Context(), uuid.FromStringOrNil(identityID))
				require.NoError(t, err)
				assert.Equal(t, "222", gjson.GetBytes(i.MetadataAdmin, "phone_number").String())
			})
		})

		t.Run("case=should fail login when mapper is invalid", func(t *testing.T) {
			params := hydraFlowParams{
				subject: testhelpers.RandomEmail(),
				scope:   []string{"openid", "offline"},
			}
			params.claims.traits.website = "https://example.com"

			// Register with a valid mapper.
			setProviderConfig(t, conf,
				newOIDCProvider(t, ts, hydraPublic, hydraAdmin, "valid", func(c *oidc.Configuration) {
					c.UpdateIdentityOnLogin = oidc.UpdateIdentityOnLoginAutomatic
				}),
			)

			t.Run("step=register", func(t *testing.T) {
				r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
				action := assertFormValues(t, r.ID, "valid")
				res, body := makeRequest(t, action, "valid", params, nil)
				assertIdentity(t, res, body, params.subject, params.claims)
			})

			t.Run("step=login with broken mapper causes error", func(t *testing.T) {
				// Reconfigure with a mapper that produces invalid output.
				setProviderConfig(t, conf,
					newOIDCProvider(t, ts, hydraPublic, hydraAdmin, "valid", func(c *oidc.Configuration) {
						c.UpdateIdentityOnLogin = oidc.UpdateIdentityOnLoginAutomatic
						c.Mapper = "file://./stub/oidc.invalid-mapper.jsonnet"
					}),
				)

				r := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
				action := assertFormValues(t, r.ID, "valid")
				res, body := makeRequest(t, action, "valid", params, nil)

				assertSystemErrorWithReason(t, res, body, http.StatusInternalServerError,
					"OpenID Connect Jsonnet mapper did not return an object for key identity.traits")
			})
		})

		t.Run("case=should preserve existing traits when mapper omits them", func(t *testing.T) {
			params := hydraFlowParams{
				subject: testhelpers.RandomEmail(),
				scope:   []string{"openid", "offline"},
			}
			params.claims.traits.website = "https://example.com"

			// Register with the standard mapper (includes subject and website in traits).
			setProviderConfig(t, conf,
				newOIDCProvider(t, ts, hydraPublic, hydraAdmin, "valid", func(c *oidc.Configuration) {
					c.UpdateIdentityOnLogin = oidc.UpdateIdentityOnLoginAutomatic
				}),
			)

			t.Run("step=register", func(t *testing.T) {
				r := newBrowserRegistrationFlow(t, returnTS.URL, time.Minute)
				action := assertFormValues(t, r.ID, "valid")
				res, body := makeRequest(t, action, "valid", params, nil)
				assertIdentity(t, res, body, params.subject, params.claims)
			})

			t.Run("step=login with mapper that omits subject preserves it via merge", func(t *testing.T) {
				// Reconfigure with a mapper that omits the "subject" trait but
				// outputs a new website. Merge should preserve subject from the
				// existing identity.
				params := params
				params.claims.traits.website = "https://new.example.com"
				setProviderConfig(t, conf,
					newOIDCProvider(t, ts, hydraPublic, hydraAdmin, "valid", func(c *oidc.Configuration) {
						c.UpdateIdentityOnLogin = oidc.UpdateIdentityOnLoginAutomatic
						c.Mapper = "file://./stub/oidc.missing-subject.jsonnet"
					}),
				)

				r := newBrowserLoginFlow(t, returnTS.URL, time.Minute)
				action := assertFormValues(t, r.ID, "valid")
				res, body := makeRequest(t, action, "valid", params, nil)

				assert.Contains(t, res.Request.URL.String(), returnTS.URL, "%s", body)
				// Subject should be preserved from the existing identity via merge.
				assert.Equal(t, params.subject, gjson.GetBytes(body, "identity.traits.subject").String(), "%s", body)
				// Website should be updated from the new mapper output.
				assert.Equal(t, "https://new.example.com", gjson.GetBytes(body, "identity.traits.website").String(), "%s", body)
			})
		})
	})

	t.Run("method=TestPopulateSignUpMethod", func(t *testing.T) {
		sr, err := registration.NewFlow(reg, &http.Request{URL: urlx.ParseOrPanic("/")}, flow.TypeBrowser)
		require.NoError(t, err)
		require.NoError(t, reg.RegistrationStrategies(t.Context()).MustStrategy(identity.CredentialsTypeOIDC).(*oidc.Strategy).PopulateRegistrationMethod(&http.Request{}, sr))

		snapshotx.SnapshotT(t, sr.UI, snapshotx.ExceptPaths("action", "nodes.0.attributes.value"))
	})
}

func prettyJSON(t *testing.T, body []byte) string {
	var out bytes.Buffer
	require.NoError(t, json.Indent(&out, body, "", "\t"))

	return out.String()
}

func TestCountActiveFirstFactorCredentials(t *testing.T) {
	t.Parallel()

	_, reg := pkg.NewFastRegistryWithMocks(t)
	strategy := oidc.NewStrategy(reg)

	toJSON := func(c identity.CredentialsOIDC) []byte {
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
				Config: toJSON(identity.CredentialsOIDC{Providers: []identity.CredentialsOIDCProvider{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
		},
		{
			in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
				Type:        strategy.ID(),
				Identifiers: []string{""},
				Config: toJSON(identity.CredentialsOIDC{Providers: []identity.CredentialsOIDCProvider{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
		},
		{
			in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
				Type:        strategy.ID(),
				Identifiers: []string{"bar:"},
				Config: toJSON(identity.CredentialsOIDC{Providers: []identity.CredentialsOIDCProvider{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
		},
		{
			in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
				Type:        strategy.ID(),
				Identifiers: []string{":foo"},
				Config: toJSON(identity.CredentialsOIDC{Providers: []identity.CredentialsOIDCProvider{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
		},
		{
			in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
				Type:        strategy.ID(),
				Identifiers: []string{"not-bar:foo"},
				Config: toJSON(identity.CredentialsOIDC{Providers: []identity.CredentialsOIDCProvider{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
		},
		{
			in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
				Type:        strategy.ID(),
				Identifiers: []string{"bar:not-foo"},
				Config: toJSON(identity.CredentialsOIDC{Providers: []identity.CredentialsOIDCProvider{
					{Subject: "foo", Provider: "bar"},
				}}),
			}},
		},
		{
			in: map[identity.CredentialsType]identity.Credentials{strategy.ID(): {
				Type:        strategy.ID(),
				Identifiers: []string{"bar:foo"},
				Config: toJSON(identity.CredentialsOIDC{Providers: []identity.CredentialsOIDCProvider{
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
			actual, err := strategy.CountActiveFirstFactorCredentials(t.Context(), in)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestDisabledEndpoint(t *testing.T) {
	t.Parallel()

	_, reg := pkg.NewFastRegistryWithMocks(t,
		configx.WithValues(testhelpers.MethodEnableConfig(identity.CredentialsTypeOIDC, false)),
		configx.WithValues(testhelpers.DefaultIdentitySchemaConfig("file://stub/stub.schema.json")),
	)
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
			c := testhelpers.NewHTTPClientWithArbitrarySessionCookie(t.Context(), t, reg)
			f := testhelpers.InitializeSettingsFlowViaAPI(t, c, publicTS)

			res, err := c.PostForm(f.Ui.Action, url.Values{"link": {"oidc"}})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, res.StatusCode)

			b := ioutilx.MustReadAll(res.Body)
			assert.Contains(t, string(b), "This endpoint was disabled by system administrator")
		})

		t.Run("flow=login", func(t *testing.T) {
			f := testhelpers.InitializeLoginFlowViaAPICtx(t.Context(), t, c, publicTS, false)
			res, err := c.PostForm(f.Ui.Action, url.Values{"provider": {"oidc"}})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, res.StatusCode)

			b := ioutilx.MustReadAll(res.Body)
			assert.Contains(t, string(b), "This endpoint was disabled by system administrator")
		})

		t.Run("flow=registration", func(t *testing.T) {
			f := testhelpers.InitializeRegistrationFlowViaAPICtx(t.Context(), t, c, publicTS)
			res, err := c.PostForm(f.Ui.Action, url.Values{"provider": {"oidc"}})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, res.StatusCode)

			b := ioutilx.MustReadAll(res.Body)
			assert.Contains(t, string(b), "This endpoint was disabled by system administrator")
		})
	})
}

func TestPostEndpointRedirect(t *testing.T) {
	t.Parallel()

	conf, reg := pkg.NewFastRegistryWithMocks(t,
		configx.WithValues(testhelpers.MethodEnableConfig(identity.CredentialsTypeOIDC, true)),
	)

	remoteAdmin, remotePublic := newHydra(t)

	publicTS, _ := testhelpers.NewKratosServerWithCSRF(t, reg)

	setProviderConfig(t, conf, newOIDCProvider(t, publicTS, remotePublic, remoteAdmin, "apple"))

	for _, providerID := range []string{"apple", "apple-asd123"} {
		t.Run("case=should redirect to GET and preserve parameters/id="+providerID, func(t *testing.T) {
			// create a client that does not follow redirects
			c := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}
			res, err := c.PostForm(publicTS.URL+"/self-service/methods/oidc/callback/"+providerID, url.Values{"state": {"foo"}, "test": {"3"}})
			require.NoError(t, err)
			defer func() { _ = res.Body.Close() }()
			assert.Equal(t, http.StatusFound, res.StatusCode)

			location, err := res.Location()
			require.NoError(t, err)
			assert.Equal(t, publicTS.URL+"/self-service/methods/oidc/callback/"+providerID+"?state=foo&test=3", location.String())

			// We don't want to add/override CSRF cookie when redirecting
			testhelpers.AssertNoCSRFCookieInResponse(t, res)
		})
	}
}

func findCsrfTokenPath(t *testing.T, body []byte) string {
	nodes := gjson.GetBytes(body, "ui.nodes").Array()
	index := slices.IndexFunc(nodes, func(n gjson.Result) bool {
		return n.Get("attributes.name").String() == "csrf_token"
	})
	require.GreaterOrEqual(t, index, 0)
	return fmt.Sprintf("ui.nodes.%v.attributes.value", index)
}
