// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package password_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/client-go"
	"github.com/ory/kratos/corpx"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	kratos "github.com/ory/kratos/internal/httpclient"
	"github.com/ory/kratos/internal/settingshelpers"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/nosurfx"
	"github.com/ory/x/assertx"
	"github.com/ory/x/httpx"
	"github.com/ory/x/ioutilx"
	"github.com/ory/x/randx"
)

func init() {
	corpx.RegisterFakes()
}

func newIdentityWithPassword(email string) *identity.Identity {
	return &identity.Identity{
		ID:  x.NewUUID(),
		NID: x.NewUUID(),
		Credentials: map[identity.CredentialsType]identity.Credentials{
			"password": {
				Type:        "password",
				Identifiers: []string{email},
				Config:      []byte(`{"hashed_password":"foo"}`),
			},
		},
		State:    identity.StateActive,
		Traits:   identity.Traits(`{"email":"` + email + `"}`),
		SchemaID: config.DefaultIdentityTraitsSchemaID,
	}
}

func newEmptyIdentity() *identity.Identity {
	return &identity.Identity{
		ID:       x.NewUUID(),
		NID:      x.NewUUID(),
		State:    identity.StateActive,
		Traits:   identity.Traits(`{}`),
		SchemaID: config.DefaultIdentityTraitsSchemaID,
	}
}

func newIdentityWithoutCredentials(email string) *identity.Identity {
	return &identity.Identity{
		ID:       x.NewUUID(),
		NID:      x.NewUUID(),
		State:    identity.StateActive,
		Traits:   identity.Traits(`{"email":"` + email + `"}`),
		SchemaID: config.DefaultIdentityTraitsSchemaID,
	}
}

func TestSettings(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh/")
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/profile.schema.json")
	testhelpers.StrategyEnable(t, conf, identity.CredentialsTypePassword.String(), true)
	testhelpers.StrategyEnable(t, conf, settings.StrategyProfile, true)

	settingsUI := testhelpers.NewSettingsUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)
	_ = testhelpers.NewLoginUIWith401Response(t, conf)
	conf.MustSet(ctx, config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1m")

	browserIdentity1 := newIdentityWithPassword("john-browser@doe.com")
	apiIdentity1 := newIdentityWithPassword("john-api@doe.com")
	browserIdentity2 := newEmptyIdentity()
	apiIdentity2 := newEmptyIdentity()

	publicTS, _ := testhelpers.NewKratosServer(t, reg)

	browserUser1 := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, ctx, reg, browserIdentity1)
	browserUser2 := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, ctx, reg, browserIdentity2)
	apiUser1 := testhelpers.NewHTTPClientWithIdentitySessionToken(t, ctx, reg, apiIdentity1)
	apiUser2 := testhelpers.NewHTTPClientWithIdentitySessionToken(t, ctx, reg, apiIdentity2)

	t.Run("case=should reject a new password if it is the same as the old one", func(t *testing.T) {
		conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()), []config.SelfServiceHook{{Name: "session"}})
		t.Cleanup(func() {
			conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()), nil)
		})

		cfg := client.NewConfiguration()
		u, err := url.Parse(publicTS.URL)
		require.NoError(t, err)
		cfg.Scheme = u.Scheme
		cfg.Host = u.Host
		cl := client.NewAPIClient(cfg)
		api := cl.FrontendAPI

		t.Run("type=api", func(t *testing.T) {
			// Create a new account.
			password := uuid.Must(uuid.NewV4()).String()
			var sessionToken string
			{
				registrationFlow, _, err := api.CreateNativeRegistrationFlow(t.Context()).Execute()
				require.NoError(t, err)
				require.NotNil(t, registrationFlow)

				registrationBody := client.UpdateRegistrationFlowBody{
					UpdateRegistrationFlowWithPasswordMethod: &client.UpdateRegistrationFlowWithPasswordMethod{
						Method:   "password",
						Password: password,
						Traits: map[string]any{
							"email": uuid.Must(uuid.NewV4()).String() + "@ory.dev",
						},
					},
				}
				registration, _, err := api.UpdateRegistrationFlow(t.Context()).Flow(registrationFlow.Id).UpdateRegistrationFlowBody(registrationBody).Execute()
				require.NoError(t, err)
				require.NotNil(t, registration)
				require.NotNil(t, registration.SessionToken)

				sessionToken = *registration.SessionToken
				require.NotEmpty(t, sessionToken)
			}

			// Create a settings flow.
			var settingsFlow *client.SettingsFlow
			{

				var err error
				settingsFlow, _, err = api.CreateNativeSettingsFlow(t.Context()).XSessionToken(sessionToken).Execute()
				require.NoError(t, err)
				require.NotNil(t, settingsFlow)
			}

			// Try to set the same password: fails.
			{
				update := client.UpdateSettingsFlowBody{
					UpdateSettingsFlowWithPasswordMethod: &client.UpdateSettingsFlowWithPasswordMethod{
						Method:   "password",
						Password: password,
					},
				}
				req := api.UpdateSettingsFlow(t.Context()).UpdateSettingsFlowBody(update).Flow(settingsFlow.Id).XSessionToken(sessionToken)
				settingsFlow, httpResp, err := api.UpdateSettingsFlowExecute(req)
				require.Error(t, err)
				require.Nil(t, settingsFlow)
				require.NotNil(t, httpResp)
				require.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
			}

			// Try to set a different password: succeeds.
			{
				update := client.UpdateSettingsFlowBody{
					UpdateSettingsFlowWithPasswordMethod: &client.UpdateSettingsFlowWithPasswordMethod{
						Method:   "password",
						Password: uuid.Must(uuid.NewV4()).String(),
					},
				}
				req := api.UpdateSettingsFlow(t.Context()).UpdateSettingsFlowBody(update).Flow(settingsFlow.Id).XSessionToken(sessionToken)
				settingsFlow, httpResp, err := api.UpdateSettingsFlowExecute(req)
				require.NoError(t, err)
				require.NotNil(t, settingsFlow)
				require.NotNil(t, httpResp)
				require.Equal(t, http.StatusOK, httpResp.StatusCode)
			}
		})

		t.Run("type=browser", func(t *testing.T) {
			// Create a new account.
			password := uuid.Must(uuid.NewV4()).String()
			var cookie string
			{
				registrationFlow, _, err := api.CreateBrowserRegistrationFlow(t.Context()).Execute()
				require.NoError(t, err)
				require.NotNil(t, registrationFlow)

				csrfToken := registrationFlow.Ui.Nodes[0].Attributes.UiNodeInputAttributes.Value.(string)
				require.NotEmpty(t, csrfToken)

				registrationBody := client.UpdateRegistrationFlowBody{
					UpdateRegistrationFlowWithPasswordMethod: &client.UpdateRegistrationFlowWithPasswordMethod{
						Method:   "password",
						Password: password,
						Traits: map[string]any{
							"email": uuid.Must(uuid.NewV4()).String() + "@ory.dev",
						},
						CsrfToken: &csrfToken,
					},
				}

				registration, httpResp, err := api.UpdateRegistrationFlow(t.Context()).Flow(registrationFlow.Id).UpdateRegistrationFlowBody(registrationBody).Execute()
				require.NoError(t, err)
				require.NotNil(t, httpResp)
				require.NotNil(t, registration)
				cookie = httpResp.Header.Get("Set-Cookie")
				require.NotEmpty(t, cookie)
			}

			// Create a settings flow.
			var settingsFlow *client.SettingsFlow
			var csrfToken string
			{

				var err error
				settingsFlow, _, err = api.CreateBrowserSettingsFlow(t.Context()).Cookie(cookie).Execute()
				require.NoError(t, err)
				require.NotNil(t, settingsFlow)

				csrfToken = settingsFlow.Ui.Nodes[0].Attributes.UiNodeInputAttributes.Value.(string)
				require.NotEmpty(t, csrfToken)
			}

			// Try to set the same password: fails.
			{
				update := client.UpdateSettingsFlowBody{
					UpdateSettingsFlowWithPasswordMethod: &client.UpdateSettingsFlowWithPasswordMethod{
						Method:    "password",
						Password:  password,
						CsrfToken: &csrfToken,
					},
				}
				req := api.UpdateSettingsFlow(t.Context()).UpdateSettingsFlowBody(update).Flow(settingsFlow.Id).Cookie(cookie)
				settingsFlow, httpResp, err := api.UpdateSettingsFlowExecute(req)
				require.Error(t, err)
				require.Nil(t, settingsFlow)
				require.NotNil(t, httpResp)
				require.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
			}

			// Try to set a different password: succeeds.
			{
				update := client.UpdateSettingsFlowBody{
					UpdateSettingsFlowWithPasswordMethod: &client.UpdateSettingsFlowWithPasswordMethod{
						Method:    "password",
						Password:  uuid.Must(uuid.NewV4()).String(),
						CsrfToken: &csrfToken,
					},
				}
				req := api.UpdateSettingsFlow(t.Context()).UpdateSettingsFlowBody(update).Flow(settingsFlow.Id).Cookie(cookie)
				settingsFlow, httpResp, err := api.UpdateSettingsFlowExecute(req)
				require.NoError(t, err)
				require.NotNil(t, settingsFlow)
				require.NotNil(t, httpResp)
				require.Equal(t, http.StatusOK, httpResp.StatusCode)
			}
		})
	})

	t.Run("description=not authorized to call endpoints without a session", func(t *testing.T) {
		c := testhelpers.NewDebugClient(t)
		t.Run("type=browser", func(t *testing.T) {
			res, err := c.Do(httpx.MustNewRequest("POST", publicTS.URL+settings.RouteSubmitFlow, strings.NewReader(url.Values{"foo": {"bar"}}.Encode()), "application/x-www-form-urlencoded"))
			require.NoError(t, err)
			defer func() { _ = res.Body.Close() }()
			assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode, "%+v", res.Request)
			assert.Contains(t, res.Request.URL.String(), conf.GetProvider(ctx).String(config.ViperKeySelfServiceLoginUI))
		})

		t.Run("type=spa", func(t *testing.T) {
			res, err := c.Do(httpx.MustNewRequest("POST", publicTS.URL+settings.RouteSubmitFlow, strings.NewReader(url.Values{"foo": {"bar"}}.Encode()), "application/json"))
			require.NoError(t, err)
			defer func() { _ = res.Body.Close() }()
			assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode, "%+v", res.Request)
			assert.Contains(t, res.Request.URL.String(), settings.RouteSubmitFlow)
		})

		t.Run("type=api", func(t *testing.T) {
			res, err := c.Do(httpx.MustNewRequest("POST", publicTS.URL+settings.RouteSubmitFlow, strings.NewReader(`{"foo":"bar"}`), "application/json"))
			require.NoError(t, err)
			assert.Len(t, res.Cookies(), 0)
			defer func() { _ = res.Body.Close() }()
			assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode)
		})
	})

	expectValidationError := func(t *testing.T, isAPI, isSPA bool, hc *http.Client, values func(url.Values)) string {
		return testhelpers.SubmitSettingsForm(t, isAPI, isSPA, hc, publicTS, values,
			testhelpers.ExpectStatusCode(isAPI || isSPA, http.StatusBadRequest, http.StatusOK),
			testhelpers.ExpectURL(isAPI || isSPA, publicTS.URL+settings.RouteSubmitFlow, conf.SelfServiceFlowSettingsUI(ctx).String()))
	}

	t.Run("description=should fail if password violates policy", func(t *testing.T) {
		check := func(t *testing.T, reason, actual string) {
			assert.Empty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==password).attributes.value").String(), "%s", actual)
			assert.NotEmpty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String(), "%s", actual)
			assert.Equal(t, reason, gjson.Get(actual, "ui.nodes.#(attributes.name==password).messages.0.text").String(), "%s", actual)
		}

		t.Run("session=with privileged session", func(t *testing.T) {
			conf.MustSet(ctx, config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "5m")

			payload := func(v url.Values) {
				v.Set("password", "123456")
				v.Set("method", "password")
			}

			t.Run("type=api", func(t *testing.T) {
				check(t, "The password must be at least 8 characters long, but got 6.", expectValidationError(t, true, false, apiUser1, payload))
			})

			t.Run("spa=spa", func(t *testing.T) {
				check(t, "The password must be at least 8 characters long, but got 6.", expectValidationError(t, false, true, browserUser1, payload))
			})

			t.Run("type=browser", func(t *testing.T) {
				check(t, "The password must be at least 8 characters long, but got 6.", expectValidationError(t, false, false, browserUser1, payload))
			})
		})

		t.Run("session=needs reauthentication", func(t *testing.T) {
			conf.MustSet(ctx, config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1ns")
			defer testhelpers.NewLoginUIWith401Response(t, conf)
			t.Cleanup(func() {
				conf.MustSet(ctx, config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "5m")
			})

			payload := func(v url.Values) {
				v.Set("method", "password")
				v.Set("password", "123456")
			}

			t.Run("type=api/expected=an error because reauth can not be initialized for API clients", func(t *testing.T) {
				_ = testhelpers.NewSettingsLoginAcceptAPIServer(t, testhelpers.NewSDKCustomClient(publicTS, apiUser1), conf)
				actual := testhelpers.SubmitSettingsForm(t, true, false, apiUser1, publicTS, payload,
					http.StatusForbidden, publicTS.URL+settings.RouteSubmitFlow)
				assertx.EqualAsJSONExcept(t, settings.NewFlowNeedsReAuth(), json.RawMessage(actual), []string{"redirect_browser_to"})
				assert.NotEmpty(t, json.RawMessage(gjson.Get(actual, "redirect_browser_to").String()))
			})

			t.Run("type=spa", func(t *testing.T) {
				_ = testhelpers.NewSettingsLoginAcceptAPIServer(t, testhelpers.NewSDKCustomClient(publicTS, browserUser1), conf)
				actual := testhelpers.SubmitSettingsForm(t, false, true, browserUser1, publicTS, payload,
					http.StatusForbidden, publicTS.URL+settings.RouteSubmitFlow)
				assertx.EqualAsJSON(t, settings.NewFlowNeedsReAuth().DefaultError, json.RawMessage(gjson.Get(actual, "error").Raw))
				assert.NotEmpty(t, json.RawMessage(gjson.Get(actual, "redirect_browser_to").String()))
			})

			t.Run("type=browser", func(t *testing.T) {
				_ = testhelpers.NewSettingsLoginAcceptAPIServer(t, testhelpers.NewSDKCustomClient(publicTS, browserUser1), conf)
				check(t, "The password must be at least 8 characters long, but got 6.", expectValidationError(t, false, false, browserUser1, payload))
			})
		})
	})

	t.Run("description=should not be able to make requests for another user", func(t *testing.T) {
		t.Run("type=api", func(t *testing.T) {
			f := testhelpers.InitializeSettingsFlowViaAPI(t, apiUser1, publicTS)
			values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
			values.Set("method", "password")
			values.Set("password", x.NewUUID().String())
			actual, res := testhelpers.SettingsMakeRequest(t, true, false, f, apiUser2, testhelpers.EncodeFormAsJSON(t, true, values))
			assert.Equal(t, http.StatusForbidden, res.StatusCode)
			assert.Contains(t, gjson.Get(actual, "error.reason").String(), "initiated by someone else", "%s", actual)
		})

		t.Run("type=spa", func(t *testing.T) {
			f := testhelpers.InitializeSettingsFlowViaBrowser(t, browserUser1, true, publicTS)
			values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
			values.Set("method", "password")
			values.Set("password", x.NewUUID().String())
			actual, res := testhelpers.SettingsMakeRequest(t, false, true, f, browserUser2, values.Encode())
			assert.Equal(t, http.StatusForbidden, res.StatusCode)
			assert.Contains(t, gjson.Get(actual, "error.reason").String(), "initiated by someone else", "%s", actual)
		})

		t.Run("type=browser", func(t *testing.T) {
			f := testhelpers.InitializeSettingsFlowViaBrowser(t, browserUser1, false, publicTS)
			values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
			values.Set("method", "password")
			values.Set("password", x.NewUUID().String())
			actual, res := testhelpers.SettingsMakeRequest(t, false, false, f, browserUser2, values.Encode())
			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Contains(t, gjson.Get(actual, "reason").String(), "initiated by someone else", "%s", actual)
		})
	})

	t.Run("description=should update the password and clear errors if everything is ok", func(t *testing.T) {
		check := func(t *testing.T, actual string) {
			assert.Equal(t, "success", gjson.Get(actual, "state").String(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==password).value").String(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==password).messages.0.text").String(), actual)
		}

		payload := func(v url.Values) {
			v.Set("method", "password")
			v.Set("password", x.NewUUID().String())
		}

		t.Run("type=api", func(t *testing.T) {
			actual := testhelpers.SubmitSettingsForm(t, true, false, apiUser1, publicTS, payload, http.StatusOK, publicTS.URL+settings.RouteSubmitFlow)
			check(t, actual)
			assert.Empty(t, gjson.Get(actual, "continue_with").Array(), "%s", actual)
		})

		t.Run("type=spa", func(t *testing.T) {
			actual := testhelpers.SubmitSettingsForm(t, false, true, browserUser1, publicTS, payload, http.StatusOK, publicTS.URL+settings.RouteSubmitFlow)
			check(t, actual)
			assert.EqualValues(t, flow.ContinueWithActionRedirectBrowserToString, gjson.Get(actual, "continue_with.0.action").String(), "%s", actual)
			assert.Contains(t, gjson.Get(actual, "continue_with.0.redirect_browser_to").String(), settingsUI.URL, "%s", actual)
		})

		t.Run("type=browser", func(t *testing.T) {
			actual := testhelpers.SubmitSettingsForm(t, false, false, browserUser1, publicTS, payload, http.StatusOK, conf.SelfServiceFlowSettingsUI(ctx).String())
			check(t, actual)
			assert.Empty(t, gjson.Get(actual, "continue_with").Array(), "%s", actual)
		})
	})

	t.Run("case=should fail because of missing CSRF token/type=browser", func(t *testing.T) {
		f := testhelpers.InitializeSettingsFlowViaBrowser(t, browserUser1, false, publicTS)
		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		values.Set("method", "password")
		values.Set("password", x.NewUUID().String())
		values.Set("csrf_token", "invalid_token")

		actual, res := testhelpers.SettingsMakeRequest(t, false, false, f, browserUser1, values.Encode())
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, res.Request.URL.String(), conf.GetProvider(ctx).String(config.ViperKeySelfServiceErrorUI))

		assertx.EqualAsJSON(t, nosurfx.ErrInvalidCSRFToken, json.RawMessage(actual), "%s", actual)
	})

	t.Run("case=should pass even without CSRF token/type=spa", func(t *testing.T) {
		f := testhelpers.InitializeSettingsFlowViaBrowser(t, browserUser1, true, publicTS)
		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		values.Set("method", "password")
		values.Set("password", x.NewUUID().String())
		values.Set("csrf_token", "invalid_token")
		actual, res := testhelpers.SettingsMakeRequest(t, false, true, f, browserUser1, testhelpers.EncodeFormAsJSON(t, true, values))
		assert.Equal(t, http.StatusForbidden, res.StatusCode)

		assert.Contains(t, res.Request.URL.String(), publicTS.URL+settings.RouteSubmitFlow)
		assertx.EqualAsJSON(t, nosurfx.ErrInvalidCSRFToken, json.RawMessage(gjson.Get(actual, "error").Raw), "%s", actual)
	})

	t.Run("case=should pass even without CSRF token/type=api", func(t *testing.T) {
		f := testhelpers.InitializeSettingsFlowViaAPI(t, apiUser1, publicTS)
		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		values.Set("method", "password")
		values.Set("password", x.NewUUID().String())
		values.Set("csrf_token", "invalid_token")
		actual, res := testhelpers.SettingsMakeRequest(t, true, false, f, apiUser1, testhelpers.EncodeFormAsJSON(t, true, values))
		assert.Equal(t, http.StatusOK, res.StatusCode)

		assert.Contains(t, res.Request.URL.String(), publicTS.URL+settings.RouteSubmitFlow)
		assert.NotEmpty(t, gjson.Get(actual, "identity.id").String(), "%s", actual)
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
				f := testhelpers.InitializeSettingsFlowViaAPI(t, apiUser1, publicTS)
				values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
				values.Set("password", x.NewUUID().String())

				req := testhelpers.NewRequest(t, true, "POST", f.Ui.Action, bytes.NewBufferString(testhelpers.EncodeFormAsJSON(t, true, values)))
				tc.mod(req.Header)

				res, err := apiUser1.Do(req)
				require.NoError(t, err)
				defer func() { _ = res.Body.Close() }()

				actual := string(ioutilx.MustReadAll(res.Body))
				assert.EqualValues(t, http.StatusBadRequest, res.StatusCode)
				assert.Contains(t, actual, tc.exp)
			})
		}
	})

	expectSuccess := func(t *testing.T, isAPI, isSPA bool, hc *http.Client, values func(url.Values)) string {
		return testhelpers.SubmitSettingsForm(t, isAPI, isSPA, hc, publicTS, values, http.StatusOK,
			testhelpers.ExpectURL(isAPI || isSPA, publicTS.URL+settings.RouteSubmitFlow, conf.SelfServiceFlowSettingsUI(ctx).String()))
	}

	t.Run("description=should update the password even if no password was set before", func(t *testing.T) {
		bi := newIdentityWithoutCredentials(x.NewUUID().String() + "@ory.sh")
		si := newIdentityWithoutCredentials(x.NewUUID().String() + "@ory.sh")
		ai := newIdentityWithoutCredentials(x.NewUUID().String() + "@ory.sh")
		browserUser := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, ctx, reg, bi)
		spaUser := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, ctx, reg, si)
		apiUser := testhelpers.NewHTTPClientWithIdentitySessionToken(t, ctx, reg, ai)

		check := func(t *testing.T, actual string, id *identity.Identity) {
			assert.Equal(t, "success", gjson.Get(actual, "state").String(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "ui.nodes.#(name==password).attributes.value").String(), "%s", actual)

			actualIdentity, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), id.ID)
			require.NoError(t, err)
			cfg := string(actualIdentity.Credentials[identity.CredentialsTypePassword].Config)
			assert.Contains(t, cfg, "hashed_password", "%+v", actualIdentity.Credentials)
			require.Len(t, actualIdentity.Credentials[identity.CredentialsTypePassword].Identifiers, 1)
			assert.Contains(t, actualIdentity.Credentials[identity.CredentialsTypePassword].Identifiers[0], "-4")
		}

		payload := func(v url.Values) {
			v.Set("method", "password")
			v.Set("password", randx.MustString(16, randx.AlphaNum))
		}

		t.Run("type=api", func(t *testing.T) {
			actual := expectSuccess(t, true, false, apiUser, payload)
			check(t, actual, ai)
		})

		t.Run("type=spa", func(t *testing.T) {
			actual := expectSuccess(t, false, true, spaUser, payload)
			check(t, actual, si)
		})

		t.Run("type=browser", func(t *testing.T) {
			actual := expectSuccess(t, false, false, browserUser, payload)
			check(t, actual, bi)
		})
	})

	t.Run("description=should update the password and perform the correct redirection", func(t *testing.T) {
		rts := testhelpers.NewRedirTS(t, "", conf)
		conf.MustSet(ctx, config.ViperKeySelfServiceSettingsAfter+"."+config.DefaultBrowserReturnURL, rts.URL+"/return-ts")
		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceSettingsAfter, nil)
		})

		run := func(t *testing.T, f *kratos.SettingsFlow, isAPI bool, c *http.Client, _ *identity.Identity) {
			values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
			values.Set("method", "password")
			values.Set("password", randx.MustString(16, randx.AlphaNum))
			_, res := testhelpers.SettingsMakeRequest(t, isAPI, false, f, c, testhelpers.EncodeFormAsJSON(t, isAPI, values))
			require.EqualValues(t, rts.URL+"/return-ts", res.Request.URL.String())

			actualIdentity, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), browserIdentity1.ID)
			require.NoError(t, err)

			cfg := string(actualIdentity.Credentials[identity.CredentialsTypePassword].Config)
			assert.NotContains(t, cfg, "foo")
			assert.NotEqual(t, `{"hashed_password":"foo"}`, cfg)
		}

		t.Run("type=api", func(t *testing.T) {
			t.Skip("Post-registration redirects do not work for API flows and are thus not tested here.")
		})

		t.Run("type=spa", func(t *testing.T) {
			t.Skip("Post-registration redirects do not work for API flows and are thus not tested here.")
		})

		t.Run("type=browser", func(t *testing.T) {
			rs := testhelpers.InitializeSettingsFlowViaBrowser(t, browserUser1, false, publicTS)
			run(t, rs, false, browserUser1, browserIdentity1)
		})
	})

	t.Run("description=should update the password and revoke other user sessions", func(t *testing.T) {
		conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceSettingsAfter, "password"), []config.SelfServiceHook{{Name: "revoke_active_sessions"}})
		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceSettingsAfter, nil)
		})

		check := func(t *testing.T, actual string, id *identity.Identity) {
			assert.Equal(t, "success", gjson.Get(actual, "state").String(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "ui.nodes.#(name==password).attributes.value").String(), "%s", actual)

			actualIdentity, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), id.ID)
			require.NoError(t, err)
			cfg := string(actualIdentity.Credentials[identity.CredentialsTypePassword].Config)
			assert.Contains(t, cfg, "hashed_password", "%+v", actualIdentity.Credentials)
			require.Len(t, actualIdentity.Credentials[identity.CredentialsTypePassword].Identifiers, 1)
			assert.Contains(t, actualIdentity.Credentials[identity.CredentialsTypePassword].Identifiers[0], "-4")
		}

		initClients := func(isAPI bool, id *identity.Identity) (client1, client2 *http.Client) {
			if isAPI {
				client1 = testhelpers.NewHTTPClientWithIdentitySessionToken(t, ctx, reg, id)
				client2 = testhelpers.NewHTTPClientWithIdentitySessionToken(t, ctx, reg, id)
				return client1, client2
			}
			client1 = testhelpers.NewHTTPClientWithIdentitySessionCookie(t, ctx, reg, id)
			client2 = testhelpers.NewHTTPClientWithIdentitySessionCookie(t, ctx, reg, id)

			return client1, client2
		}

		run := func(t *testing.T, isAPI, isSPA bool, id *identity.Identity) {
			payload := func(v url.Values) {
				v.Set("method", "password")
				v.Set("password", randx.MustString(16, randx.AlphaNum))
			}

			user1, user2 := initClients(isAPI, id)

			actual := expectSuccess(t, isAPI, isSPA, user1, payload)
			check(t, actual, id)

			// second client should be logged out
			res, err := user2.Do(httpx.MustNewRequest("POST", publicTS.URL+settings.RouteSubmitFlow, strings.NewReader(url.Values{"foo": {"bar"}}.Encode()), "application/json"))
			require.NoError(t, err)
			require.NoError(t, res.Body.Close())
			assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode, "%+v", res.Request)

			// again change password via first client
			actual = expectSuccess(t, isAPI, isSPA, user1, payload)
			check(t, actual, id)
		}

		t.Run("type=api", func(t *testing.T) {
			id := newIdentityWithoutCredentials(x.NewUUID().String() + "@ory.sh")
			run(t, true, false, id)
		})

		t.Run("type=spa", func(t *testing.T) {
			id := newIdentityWithoutCredentials(x.NewUUID().String() + "@ory.sh")
			run(t, false, true, id)
		})

		t.Run("type=browser", func(t *testing.T) {
			id := newIdentityWithoutCredentials(x.NewUUID().String() + "@ory.sh")
			run(t, false, false, id)
		})
	})

	t.Run("case=should fail if no identifier was set in the schema", func(t *testing.T) {
		testhelpers.SetDefaultIdentitySchema(conf, "file://stub/missing-identifier.schema.json")

		id := newIdentityWithoutCredentials(testhelpers.RandomEmail())
		browser := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, ctx, reg, id)
		api := testhelpers.NewHTTPClientWithIdentitySessionToken(t, ctx, reg, id)

		for _, f := range []string{"spa", "api", "browser"} {
			t.Run("type="+f, func(t *testing.T) {
				hc := browser
				if f == "api" {
					hc = api
				}
				actual := settingshelpers.ExpectValidationError(t, publicTS, hc, conf, f, func(v url.Values) {
					v.Set("password", x.NewUUID().String())
					v.Set("method", "password")
				})
				assert.Equal(t, text.NewErrorValidationIdentifierMissing().Text, gjson.Get(actual, "ui.messages.0.text").String(), "%s", actual)
			})
		}
	})
}
