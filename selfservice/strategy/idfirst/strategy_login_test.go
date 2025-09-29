// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package idfirst_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	kratos "github.com/ory/kratos/internal/httpclient"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/strategy/idfirst"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/nosurfx"
	"github.com/ory/x/assertx"
	"github.com/ory/x/contextx"
	"github.com/ory/x/ioutilx"
	"github.com/ory/x/snapshotx"
	"github.com/ory/x/urlx"
)

//go:embed stub/default.schema.json
var loginSchema []byte

func TestCompleteLogin(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)

	// We enable the password method to test the identifier first strategy

	// ctx = contextx.WithConfigValue(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword), map[string]interface{}{"enabled": true})
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword), map[string]interface{}{"enabled": true})

	// ctx = contextx.WithConfigValue(ctx, config.ViperKeySelfServiceLoginFlowStyle, "identifier_first")
	conf.MustSet(ctx, config.ViperKeySelfServiceLoginFlowStyle, "identifier_first")

	router := x.NewRouterPublic(reg)
	publicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin(reg))

	errTS := testhelpers.NewErrorTestServer(t, reg)
	uiTS := testhelpers.NewLoginUIFlowEchoServer(t, reg)
	redirTS := testhelpers.NewRedirSessionEchoTS(t, reg)

	// Overwrite these two:
	// ctx = contextx.WithConfigValue(ctx, config.ViperKeySelfServiceErrorUI, errTS.URL+"/error-ts")
	conf.MustSet(ctx, config.ViperKeySelfServiceErrorUI, errTS.URL+"/error-ts")

	// ctx = contextx.WithConfigValue(ctx, config.ViperKeySelfServiceLoginUI, uiTS.URL+"/login-ts")
	conf.MustSet(ctx, config.ViperKeySelfServiceLoginUI, uiTS.URL+"/login-ts")

	// ctx = testhelpers.WithDefaultIdentitySchemaFromRaw(ctx, loginSchema)
	testhelpers.SetDefaultIdentitySchemaFromRaw(conf, loginSchema)

	// ctx = contextx.WithConfigValue(ctx, config.ViperKeySecretsDefault, []string{"not-a-secure-session-key"})
	conf.MustSet(ctx, config.ViperKeySecretsDefault, []string{"not-a-secure-session-key"})

	//ensureFieldsExist := func(t *testing.T, body []byte) {
	//	registrationhelpers.CheckFormContent(t, body, "identifier",
	//		"password",
	//		"csrf_token")
	//}

	apiClient := testhelpers.NewDebugClient(t)

	t.Run("case=should show the error ui because the request payload is malformed", func(t *testing.T) {
		t.Run("type=api", func(t *testing.T) {
			f := testhelpers.InitializeLoginFlowViaAPIWithContext(t, ctx, apiClient, publicTS, false)

			body, res := testhelpers.LoginMakeRequestWithContext(t, ctx, true, false, f, apiClient, "14=)=!(%)$/ZP()GHIÖ")
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, body, `Expected JSON sent in request body to be an object but got: Number`)
		})

		t.Run("type=browser", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, false, false, false, testhelpers.InitFlowWithContext(ctx))

			body, res := testhelpers.LoginMakeRequestWithContext(t, ctx, false, false, f, browserClient, "14=)=!(%)$/ZP()GHIÖ")
			assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/login-ts")
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "invalid URL escape", "%s", body)
		})

		t.Run("type=spa", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, true, false, false, testhelpers.InitFlowWithContext(ctx))

			body, res := testhelpers.LoginMakeRequestWithContext(t, ctx, false, true, f, browserClient, "14=)=!(%)$/ZP()GHIÖ")
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "invalid URL escape", "%s", body)
		})
	})

	t.Run("case=should fail because identifier first can not handle AAL2", func(t *testing.T) {
		f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false)

		update, err := reg.LoginFlowPersister().GetLoginFlow(context.Background(), uuid.FromStringOrNil(f.Id))
		require.NoError(t, err)
		update.RequestedAAL = identity.AuthenticatorAssuranceLevel2
		require.NoError(t, reg.LoginFlowPersister().UpdateLoginFlow(context.Background(), update))

		req, err := http.NewRequest("POST", f.Ui.Action, bytes.NewBufferString(`{"method":"identifier_first"}`))
		require.NoError(t, err)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		actual, res := testhelpers.MockMakeAuthenticatedRequest(t, reg, conf, router, req)
		assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
		assert.Equal(t, text.NewErrorValidationLoginNoStrategyFound().Text, gjson.GetBytes(actual, "ui.messages.0.text").String())
	})

	t.Run("should return an error because the request does not exist", func(t *testing.T) {
		check := func(t *testing.T, actual string) {
			assert.Equal(t, int64(http.StatusNotFound), gjson.Get(actual, "code").Int(), "%s", actual)
			assert.Equal(t, "Not Found", gjson.Get(actual, "status").String(), "%s", actual)
			assert.Contains(t, gjson.Get(actual, "message").String(), "Unable to locate the resource", "%s", actual)
		}

		fakeFlow := &kratos.LoginFlow{
			Ui: kratos.UiContainer{
				Action: publicTS.URL + login.RouteSubmitFlow + "?flow=" + x.NewUUID().String(),
			},
		}

		t.Run("type=api", func(t *testing.T) {
			actual, res := testhelpers.LoginMakeRequestWithContext(t, ctx, true, false, fakeFlow, apiClient, "{}")
			assert.Len(t, res.Cookies(), 0)
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
			check(t, gjson.Get(actual, "error").Raw)
		})

		t.Run("type=browser", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			actual, res := testhelpers.LoginMakeRequestWithContext(t, ctx, false, false, fakeFlow, browserClient, "")
			assert.Contains(t, res.Request.URL.String(), errTS.URL)
			check(t, actual)
		})

		t.Run("type=api", func(t *testing.T) {
			actual, res := testhelpers.LoginMakeRequestWithContext(t, ctx, false, true, fakeFlow, apiClient, "{}")
			assert.Len(t, res.Cookies(), 0)
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
			check(t, gjson.Get(actual, "error").Raw)
		})
	})

	t.Run("case=should return an error because the request is expired", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceLoginRequestLifespan, time.Millisecond*30)
		conf.MustSet(ctx, config.ViperKeySecurityAccountEnumerationMitigate, true)
		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceLoginRequestLifespan, time.Hour)
			conf.MustSet(ctx, config.ViperKeySecurityAccountEnumerationMitigate, nil)
		})

		values := url.Values{
			"csrf_token": {nosurfx.FakeCSRFToken},
			"identifier": {"identifier"},
			"method":     {"identifier_first"},
		}

		t.Run("type=api", func(t *testing.T) {
			f := testhelpers.InitializeLoginFlowViaAPIWithContext(t, ctx, apiClient, publicTS, false)

			time.Sleep(time.Millisecond * 60)
			actual, res := testhelpers.LoginMakeRequestWithContext(t, ctx, true, false, f, apiClient, testhelpers.EncodeFormAsJSON(t, true, values))
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
			assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", gjson.Get(actual, "use_flow_id").String())
			assertx.EqualAsJSONExcept(t, flow.NewFlowExpiredError(time.Now()), json.RawMessage(actual), []string{"use_flow_id", "since", "expired_at"}, "expired", "%s", actual)
		})

		t.Run("type=browser", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, false, false, false)

			time.Sleep(time.Millisecond * 60)
			actual, res := testhelpers.LoginMakeRequestWithContext(t, ctx, false, false, f, browserClient, values.Encode())
			assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/login-ts")
			assert.NotEqual(t, f.Id, gjson.Get(actual, "id").String(), "%s", actual)
			assert.Contains(t, gjson.Get(actual, "ui.messages.0.text").String(), "expired", "%s", actual)
		})

		t.Run("type=SPA", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, true, false, false)

			time.Sleep(time.Millisecond * 60)
			actual, res := testhelpers.LoginMakeRequestWithContext(t, ctx, false, true, f, apiClient, testhelpers.EncodeFormAsJSON(t, true, values))
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
			assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", gjson.Get(actual, "use_flow_id").String())
			assertx.EqualAsJSONExcept(t, flow.NewFlowExpiredError(time.Now()), json.RawMessage(actual), []string{"use_flow_id", "since", "expired_at"}, "expired", "%s", actual)
		})
	})

	t.Run("case=should have correct CSRF behavior", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySecurityAccountEnumerationMitigate, true)
		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySecurityAccountEnumerationMitigate, nil)
		})

		values := url.Values{
			"method":     {"identifier_first"},
			"csrf_token": {"invalid_token"},
			"identifier": {"login-identifier-csrf-browser"},
		}

		t.Run("case=should fail because of missing CSRF token/type=browser", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, false, false, false)

			actual, res := testhelpers.LoginMakeRequest(t, false, false, f, browserClient, values.Encode())
			assert.EqualValues(t, http.StatusOK, res.StatusCode)
			assertx.EqualAsJSON(t, nosurfx.ErrInvalidCSRFToken,
				json.RawMessage(actual), "%s", actual)
		})

		t.Run("case=should fail because of missing CSRF token/type=spa", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, true, false, false)

			actual, res := testhelpers.LoginMakeRequest(t, false, true, f, browserClient, values.Encode())
			assert.EqualValues(t, http.StatusForbidden, res.StatusCode)
			assertx.EqualAsJSON(t, nosurfx.ErrInvalidCSRFToken,
				json.RawMessage(gjson.Get(actual, "error").Raw), "%s", actual)
		})

		t.Run("case=should pass even without CSRF token/type=api", func(t *testing.T) {
			f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false)

			actual, res := testhelpers.LoginMakeRequest(t, true, false, f, apiClient, testhelpers.EncodeFormAsJSON(t, true, values))
			assert.EqualValues(t, http.StatusBadRequest, res.StatusCode)
			assert.Contains(t, actual, "1010022")
		})

		t.Run("case=should fail with correct CSRF error cause/type=api", func(t *testing.T) {
			for k, tc := range []struct {
				mod func(http.Header)
				exp string
			}{
				{
					mod: func(h http.Header) {
						h.Add("Cookie", "name=bar")
					},
					exp: "The HTTP Request Header included the \\\"Cookie\\\" key",
				},
				{
					mod: func(h http.Header) {
						h.Add("Origin", "www.bar.com")
					},
					exp: "The HTTP Request Header included the \\\"Origin\\\" key",
				},
			} {
				t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
					f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false)

					req := testhelpers.NewRequest(t, true, "POST", f.Ui.Action, bytes.NewBufferString(testhelpers.EncodeFormAsJSON(t, true, values)))
					tc.mod(req.Header)

					res, err := apiClient.Do(req)
					require.NoError(t, err)
					defer func() { _ = res.Body.Close() }()

					actual := string(ioutilx.MustReadAll(res.Body))
					assert.EqualValues(t, http.StatusBadRequest, res.StatusCode)
					assert.Contains(t, actual, tc.exp)
				})
			}
		})
	})

	expectValidationError := func(t *testing.T, isAPI, refresh, isSPA bool, values func(url.Values)) string {
		return testhelpers.SubmitLoginForm(t, isAPI, nil, publicTS, values,
			isSPA, refresh,
			testhelpers.ExpectStatusCode(isAPI || isSPA, http.StatusBadRequest, http.StatusOK),
			testhelpers.ExpectURL(isAPI || isSPA, publicTS.URL+login.RouteSubmitFlow, conf.SelfServiceFlowLoginUI(ctx).String()))
	}

	t.Run("should return an error because the user does not exist", func(t *testing.T) {
		// In this test we check if the account mitigation behaves correctly by enabling all login strategies EXCEPT
		// for the passwordless code strategy. That is because this strategy always shows the login button.

		testhelpers.StrategyEnable(t, conf, identity.CredentialsTypePassword.String(), true)

		testhelpers.StrategyEnable(t, conf, identity.CredentialsTypeOIDC.String(), true)
		conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeOIDC)+".config", &oidc.ConfigurationCollection{Providers: []oidc.Configuration{
			{
				ID:           "google",
				Provider:     "google",
				Label:        "Google",
				ClientID:     "a",
				ClientSecret: "b",
				Mapper:       "file://",
			},
		}})

		testhelpers.StrategyEnable(t, conf, identity.CredentialsTypeWebAuthn.String(), true)
		conf.MustSet(ctx, config.ViperKeyWebAuthnPasswordless, true)
		conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeWebAuthn)+".config.rp.display_name", "Ory Corp")
		conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeWebAuthn)+".config.rp.id", "localhost")
		conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeWebAuthn)+".config.rp.origin", "http://localhost:4455")

		testhelpers.StrategyEnable(t, conf, identity.CredentialsTypePasskey.String(), true)
		conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePasskey)+".enabled", true)
		conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePasskey)+".config.rp.display_name", "Ory Corp")
		conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePasskey)+".config.rp.id", "localhost")
		conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePasskey)+".config.rp.origins", []string{"http://localhost:4455"})

		t.Cleanup(func() {
			conf.MustSet(ctx, "selfservice.methods.password", nil)
			conf.MustSet(ctx, "selfservice.methods.oidc", nil)
			conf.MustSet(ctx, "selfservice.methods.passkey", nil)
			conf.MustSet(ctx, "selfservice.methods.webauthn", nil)
			conf.MustSet(ctx, "selfservice.methods.code", nil)
		})

		t.Run("account enumeration mitigation enabled", func(t *testing.T) {
			conf.MustSet(ctx, config.ViperKeySecurityAccountEnumerationMitigate, true)

			t.Cleanup(func() {
				conf.MustSet(ctx, config.ViperKeySecurityAccountEnumerationMitigate, nil)
			})

			check := func(t *testing.T, body string, isAPI bool) {
				t.Logf("%s", body)
				if !isAPI {
					assert.Contains(t, body, fmt.Sprintf("%d", text.InfoSelfServiceLoginWebAuthn), "we do expect to see a webauthn trigger:\n%s", body)
					assert.Contains(t, body, fmt.Sprintf("%d", text.InfoSelfServiceLoginPasskey), "we do expect to see a passkey trigger button:\n%s", body)
				}

				assert.Equal(t, "hidden", gjson.Get(body, "ui.nodes.#(attributes.name==identifier).attributes.type").String(), "identifier is hidden to appear that we found an identity even though we did not")

				assert.NotContains(t, body, text.NewErrorValidationAccountNotFound().Text, "we do not expect to see an account not found error:\n%s", body)

				assert.Contains(t, body, fmt.Sprintf("%d", text.InfoSelfServiceLoginPassword), "we do expect to see a password trigger:\n%s", body)

				// We do expect to see the same social sign in buttons that were on the first page:
				assert.Contains(t, body, fmt.Sprintf("%d", text.InfoSelfServiceLoginWith), "we do expect to see a oidc trigger:\n%s", body)
				assert.Contains(t, body, "google", "we do expect to see a google trigger:\n%s", body)
			}

			values := func(v url.Values) {
				v.Set("identifier", "identifier")
				v.Set("method", "identifier_first")
			}

			t.Run("type=browser", func(t *testing.T) {
				check(t, expectValidationError(t, false, false, false, values), false)
			})

			t.Run("type=SPA", func(t *testing.T) {
				check(t, expectValidationError(t, false, false, true, values), false)
			})

			t.Run("type=api", func(t *testing.T) {
				check(t, expectValidationError(t, true, false, false, values), true)
			})
		})

		t.Run("account enumeration mitigation disabled", func(t *testing.T) {
			conf.MustSet(ctx, config.ViperKeySecurityAccountEnumerationMitigate, false)
			t.Cleanup(func() {
				conf.MustSet(ctx, config.ViperKeySecurityAccountEnumerationMitigate, nil)
			})

			check := func(t *testing.T, body string) {
				t.Logf("aaxx %s", body)
				assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
				assert.Contains(t, gjson.Get(body, "ui.action").String(), publicTS.URL+login.RouteSubmitFlow, "%s", body)
				assert.Contains(t, body, text.NewErrorValidationAccountNotFound().Text, "we do expect to see an error that the account does not exist: %s", body)

				assert.Equal(t, "text", gjson.Get(body, "ui.nodes.#(attributes.name==identifier).attributes.type").String(), "identifier is not hidden and we can see the input field as well")
				assert.Equal(t, "google", gjson.Get(body, "ui.nodes.#(attributes.name==provider).attributes.value").String(), "google oidc button is not hidden")

				assert.NotContains(t, body, fmt.Sprintf("%d", text.InfoSelfServiceLoginPasskey), "we do not expect to see a passkey trigger button: %s", body)
				assert.NotContains(t, body, fmt.Sprintf("%d", text.InfoSelfServiceLoginWebAuthn), "we do not expect to see a webauthn trigger: %s", body)
				assert.NotContains(t, body, fmt.Sprintf("%d", text.InfoSelfServiceLoginPassword), "we do not expect to see a password trigger: %s", body)
			}

			values := func(v url.Values) {
				v.Set("identifier", "identifier")
				v.Set("method", "identifier_first")
			}

			t.Run("type=browser", func(t *testing.T) {
				check(t, expectValidationError(t, false, false, false, values))
			})

			t.Run("type=SPA", func(t *testing.T) {
				check(t, expectValidationError(t, false, false, true, values))
			})

			t.Run("type=api", func(t *testing.T) {
				check(t, expectValidationError(t, true, false, false, values))
			})
		})
	})

	t.Run("should pass with real request", func(t *testing.T) {
		identifier, pwd := x.NewUUID().String(), "password"
		createIdentity(ctx, reg, t, identifier, pwd)

		firstValues := func(v url.Values) {
			v.Set("identifier", identifier)
			v.Set("method", "identifier_first")
		}

		secondValues := func(v url.Values) {
			v.Set("identifier", identifier)
			v.Set("password", pwd)
			v.Set("method", "password")
		}

		t.Run("type=browser", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)

			secondStep := testhelpers.SubmitLoginForm(t, false, browserClient, publicTS, firstValues,
				true, false, http.StatusBadRequest, publicTS.URL+login.RouteSubmitFlow)
			t.Logf("secondStep: %s", secondStep)
			assert.Contains(t, secondStep, "current-password")
			assert.Contains(t, secondStep, `"value":"password"`)

			body := testhelpers.SubmitLoginForm(t, false, browserClient, publicTS, secondValues,
				false, false, http.StatusOK, redirTS.URL)

			assert.Equal(t, identifier, gjson.Get(body, "identity.traits.subject").String(), "%s", body)
		})

		t.Run("type=spa", func(t *testing.T) {
			hc := testhelpers.NewClientWithCookies(t)

			secondStep := testhelpers.SubmitLoginForm(t, false, hc, publicTS, firstValues,
				true, false, http.StatusBadRequest, publicTS.URL+login.RouteSubmitFlow)
			t.Logf("secondStep: %s", secondStep)
			assert.Contains(t, secondStep, "current-password")
			assert.Contains(t, secondStep, `"value":"password"`)

			body := testhelpers.SubmitLoginForm(t, false, hc, publicTS, secondValues,
				true, false, http.StatusOK, publicTS.URL+login.RouteSubmitFlow)

			assert.Equal(t, identifier, gjson.Get(body, "session.identity.traits.subject").String(), "%s", body)
			assert.Empty(t, gjson.Get(body, "session_token").String(), "%s", body)
			assert.Empty(t, gjson.Get(body, "session.token").String(), "%s", body)

			// Was the session cookie set?
			require.NotEmpty(t, hc.Jar.Cookies(urlx.ParseOrPanic(publicTS.URL)), "%+v", hc.Jar)
		})

		t.Run("type=api", func(t *testing.T) {
			secondStep := testhelpers.SubmitLoginForm(t, true, nil, publicTS, firstValues,
				false, false, http.StatusBadRequest, publicTS.URL+login.RouteSubmitFlow)
			t.Logf("secondStep: %s", secondStep)
			assert.Contains(t, secondStep, "current-password")
			assert.Contains(t, secondStep, `"value":"password"`)

			body := testhelpers.SubmitLoginForm(t, true, nil, publicTS, secondValues,
				false, false, http.StatusOK, publicTS.URL+login.RouteSubmitFlow)

			assert.Equal(t, identifier, gjson.Get(body, "session.identity.traits.subject").String(), "%s", body)
			st := gjson.Get(body, "session_token").String()
			assert.NotEmpty(t, st, "%s", body)
		})
	})
}

func TestFormHydration(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	ctx = contextx.WithConfigValue(ctx, config.ViperKeySelfServiceLoginFlowStyle, "identifier_first")
	ctx = contextx.WithConfigValue(ctx, config.ViperKeyDefaultIdentitySchemaID, "default")
	ctx = contextx.WithConfigValue(ctx, config.ViperKeyIdentitySchemas, config.Schemas{
		{ID: "default", URL: "base64://" + base64.URLEncoding.EncodeToString(loginSchema), SelfserviceSelectable: true},
		{ID: "not-default", URL: "file://stub/doesnotexist.schema.json", SelfserviceSelectable: true},
	})

	s, err := reg.AllLoginStrategies().Strategy(identity.CredentialsType(node.IdentifierFirstGroup))
	require.NoError(t, err)
	fh, ok := s.(login.FormHydrator)
	require.True(t, ok)

	toSnapshot := func(t *testing.T, f *login.Flow) {
		t.Helper()
		// The CSRF token has a unique value that messes with the snapshot - ignore it.
		f.UI.Nodes.ResetNodes("csrf_token")
		snapshotx.SnapshotT(t, f.UI.Nodes)
	}
	newFlow := func(ctx context.Context, t *testing.T) (*http.Request, *login.Flow) {
		t.Helper()
		query := ""

		r := httptest.NewRequest("GET", "/self-service/login/browser"+query, nil)
		r = r.WithContext(ctx)
		f, err := login.NewFlow(conf, time.Minute, "csrf_token", r, flow.TypeBrowser)
		require.NoError(t, err)
		return r, f
	}
	t.Run("method=PopulateLoginMethodSecondFactor", func(t *testing.T) {
		r, f := newFlow(ctx, t)
		f.RequestedAAL = identity.AuthenticatorAssuranceLevel2
		require.NoError(t, fh.PopulateLoginMethodSecondFactor(r, f))
		toSnapshot(t, f)
	})

	t.Run("method=PopulateLoginMethodFirstFactor", func(t *testing.T) {
		r, f := newFlow(ctx, t)
		require.NoError(t, fh.PopulateLoginMethodFirstFactor(r, f))
		toSnapshot(t, f)
	})

	t.Run("method=PopulateLoginMethodFirstFactorRefresh", func(t *testing.T) {
		r, f := newFlow(ctx, t)
		require.NoError(t, fh.PopulateLoginMethodFirstFactorRefresh(r, f, nil))
		toSnapshot(t, f)
	})

	t.Run("method=PopulateLoginMethodRefresh", func(t *testing.T) {
		r, f := newFlow(ctx, t)
		require.NoError(t, fh.PopulateLoginMethodSecondFactorRefresh(r, f))
		toSnapshot(t, f)
	})

	t.Run("method=PopulateLoginMethodIdentifierFirstCredentials", func(t *testing.T) {
		t.Run("case=no options", func(t *testing.T) {
			r, f := newFlow(ctx, t)
			require.ErrorIs(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f), idfirst.ErrNoCredentialsFound)
			toSnapshot(t, f)
		})

		t.Run("case=WithIdentifier", func(t *testing.T) {
			r, f := newFlow(ctx, t)
			require.ErrorIs(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f, login.WithIdentifier("foo@bar.com")), idfirst.ErrNoCredentialsFound)
			toSnapshot(t, f)
		})

		t.Run("case=WithIdentityHint", func(t *testing.T) {
			t.Run("case=account enumeration mitigation enabled", func(t *testing.T) {
				ctx := contextx.WithConfigValue(ctx, config.ViperKeySecurityAccountEnumerationMitigate, true)

				id := identity.NewIdentity("default")
				r, f := newFlow(ctx, t)
				require.ErrorIs(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f, login.WithIdentityHint(id)), idfirst.ErrNoCredentialsFound)
				toSnapshot(t, f)
			})

			t.Run("case=account enumeration mitigation disabled", func(t *testing.T) {
				ctx := contextx.WithConfigValue(ctx, config.ViperKeySecurityAccountEnumerationMitigate, false)

				t.Run("case=identity has password", func(t *testing.T) {
					id := identity.NewIdentity("default")

					r, f := newFlow(ctx, t)
					require.ErrorIs(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f, login.WithIdentityHint(id)), idfirst.ErrNoCredentialsFound)
					toSnapshot(t, f)
				})

				t.Run("case=identity does not have a password", func(t *testing.T) {
					id := identity.NewIdentity("default")
					r, f := newFlow(ctx, t)
					require.ErrorIs(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f, login.WithIdentityHint(id)), idfirst.ErrNoCredentialsFound)
					toSnapshot(t, f)
				})
			})
		})
	})

	t.Run("method=PopulateLoginMethodIdentifierFirstIdentification", func(t *testing.T) {
		r, f := newFlow(ctx, t)
		require.NoError(t, fh.PopulateLoginMethodIdentifierFirstIdentification(r, f))
		toSnapshot(t, f)
	})
}
