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

	kratos "github.com/ory/kratos-client-go"

	"github.com/ory/kratos/corpx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/x"
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
		ID: x.NewUUID(),
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
		State:    identity.StateActive,
		Traits:   identity.Traits(`{}`),
		SchemaID: config.DefaultIdentityTraitsSchemaID,
	}
}

func TestSettings(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh/")
	conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/profile.schema.json")
	testhelpers.StrategyEnable(t, conf, identity.CredentialsTypePassword.String(), true)
	testhelpers.StrategyEnable(t, conf, settings.StrategyProfile, true)

	_ = testhelpers.NewSettingsUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)
	_ = testhelpers.NewLoginUIWith401Response(t, conf)
	conf.MustSet(config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1m")

	browserIdentity1 := newIdentityWithPassword("john-browser@doe.com")
	apiIdentity1 := newIdentityWithPassword("john-api@doe.com")
	browserIdentity2 := newEmptyIdentity()
	apiIdentity2 := newEmptyIdentity()

	publicTS, adminTS := testhelpers.NewKratosServer(t, reg)

	browserUser1 := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, browserIdentity1)
	browserUser2 := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, browserIdentity2)
	apiUser1 := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, apiIdentity1)
	apiUser2 := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, apiIdentity2)

	adminClient := testhelpers.NewSDKClient(adminTS)

	t.Run("description=not authorized to call endpoints without a session", func(t *testing.T) {
		c := testhelpers.NewDebugClient(t)
		t.Run("type=browser", func(t *testing.T) {
			res, err := c.Do(httpx.MustNewRequest("POST", publicTS.URL+settings.RouteSubmitFlow, strings.NewReader(url.Values{"foo": {"bar"}}.Encode()), "application/x-www-form-urlencoded"))
			require.NoError(t, err)
			defer res.Body.Close()
			assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode, "%+v", res.Request)
			assert.Contains(t, res.Request.URL.String(), conf.Source().String(config.ViperKeySelfServiceLoginUI))
		})

		t.Run("type=spa", func(t *testing.T) {
			res, err := c.Do(httpx.MustNewRequest("POST", publicTS.URL+settings.RouteSubmitFlow, strings.NewReader(url.Values{"foo": {"bar"}}.Encode()), "application/json"))
			require.NoError(t, err)
			defer res.Body.Close()
			assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode, "%+v", res.Request)
			assert.Contains(t, res.Request.URL.String(), settings.RouteSubmitFlow)
		})

		t.Run("type=api", func(t *testing.T) {
			res, err := c.Do(httpx.MustNewRequest("POST", publicTS.URL+settings.RouteSubmitFlow, strings.NewReader(`{"foo":"bar"}`), "application/json"))
			require.NoError(t, err)
			assert.Len(t, res.Cookies(), 0)
			defer res.Body.Close()
			assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode)
		})
	})

	var expectValidationError = func(t *testing.T, isAPI, isSPA bool, hc *http.Client, values func(url.Values)) string {
		return testhelpers.SubmitSettingsForm(t, isAPI, isSPA, hc, publicTS, values,
			testhelpers.ExpectStatusCode(isAPI || isSPA, http.StatusBadRequest, http.StatusOK),
			testhelpers.ExpectURL(isAPI || isSPA, publicTS.URL+settings.RouteSubmitFlow, conf.SelfServiceFlowSettingsUI().String()))
	}

	t.Run("description=should fail if password violates policy", func(t *testing.T) {
		var check = func(t *testing.T, actual string) {
			assert.Empty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==password).attributes.value").String(), "%s", actual)
			assert.NotEmpty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String(), "%s", actual)
			assert.Contains(t, gjson.Get(actual, "ui.nodes.#(attributes.name==password).messages.0.text").String(), "password can not be used because", "%s", actual)
		}

		t.Run("session=with privileged session", func(t *testing.T) {
			conf.MustSet(config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "5m")

			var payload = func(v url.Values) {
				v.Set("password", "123456")
				v.Set("method", "password")
			}

			t.Run("type=api", func(t *testing.T) {
				check(t, expectValidationError(t, true, false, apiUser1, payload))
			})

			t.Run("spa=spa", func(t *testing.T) {
				check(t, expectValidationError(t, false, true, browserUser1, payload))
			})

			t.Run("type=browser", func(t *testing.T) {
				check(t, expectValidationError(t, false, false, browserUser1, payload))
			})
		})

		t.Run("session=needs reauthentication", func(t *testing.T) {
			conf.MustSet(config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1ns")
			_ = testhelpers.NewSettingsLoginAcceptAPIServer(t, adminClient, conf)
			defer testhelpers.NewLoginUIWith401Response(t, conf)
			t.Cleanup(func() {
				conf.MustSet(config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "5m")
			})

			var payload = func(v url.Values) {
				v.Set("method", "password")
				v.Set("password", "123456")
			}

			t.Run("type=api/expected=an error because reauth can not be initialized for API clients", func(t *testing.T) {
				actual := testhelpers.SubmitSettingsForm(t, true, false, apiUser1, publicTS, payload,
					http.StatusForbidden, publicTS.URL+settings.RouteSubmitFlow)
				assertx.EqualAsJSON(t, settings.NewFlowNeedsReAuth(), json.RawMessage(gjson.Get(actual, "error").Raw))
			})

			t.Run("type=spa", func(t *testing.T) {
				actual := testhelpers.SubmitSettingsForm(t, false, true, browserUser1, publicTS, payload,
					http.StatusForbidden, publicTS.URL+settings.RouteSubmitFlow)
				assertx.EqualAsJSON(t, settings.NewFlowNeedsReAuth(), json.RawMessage(gjson.Get(actual, "error").Raw))
			})

			t.Run("type=browser", func(t *testing.T) {
				check(t, expectValidationError(t, false, false, browserUser1, payload))
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
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
			assert.Contains(t, gjson.Get(actual, "ui.messages.0.text").String(), "initiated by another person", "%s", actual)
		})

		t.Run("type=spa", func(t *testing.T) {
			f := testhelpers.InitializeSettingsFlowViaBrowser(t, browserUser1, true, publicTS)
			values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
			values.Set("method", "password")
			values.Set("password", x.NewUUID().String())
			actual, res := testhelpers.SettingsMakeRequest(t, false, true, f, browserUser2, values.Encode())
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
			assert.Contains(t, gjson.Get(actual, "ui.messages.0.text").String(), "initiated by another person", "%s", actual)
		})

		t.Run("type=browser", func(t *testing.T) {
			f := testhelpers.InitializeSettingsFlowViaBrowser(t, browserUser1, false, publicTS)
			values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
			values.Set("method", "password")
			values.Set("password", x.NewUUID().String())
			actual, res := testhelpers.SettingsMakeRequest(t, false, false, f, browserUser2, values.Encode())
			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Contains(t, gjson.Get(actual, "ui.messages.0.text").String(), "initiated by another person", "%s", actual)
		})
	})

	t.Run("description=should update the password and clear errors if everything is ok", func(t *testing.T) {
		var check = func(t *testing.T, actual string) {
			assert.Equal(t, "success", gjson.Get(actual, "state").String(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==password).value").String(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==password).messages.0.text").String(), actual)
		}

		var payload = func(v url.Values) {
			v.Set("method", "password")
			v.Set("password", x.NewUUID().String())
		}

		t.Run("type=api", func(t *testing.T) {
			actual := testhelpers.SubmitSettingsForm(t, true, false, apiUser1, publicTS, payload, http.StatusOK, publicTS.URL+settings.RouteSubmitFlow)
			check(t, gjson.Get(actual, "flow").Raw)
		})

		t.Run("type=spa", func(t *testing.T) {
			actual := testhelpers.SubmitSettingsForm(t, false, true, browserUser1, publicTS, payload, http.StatusOK, publicTS.URL+settings.RouteSubmitFlow)
			check(t, gjson.Get(actual, "flow").Raw)
		})

		t.Run("type=browser", func(t *testing.T) {
			check(t, testhelpers.SubmitSettingsForm(t, false, false, browserUser1, publicTS, payload, http.StatusOK, conf.SelfServiceFlowSettingsUI().String()))
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
		assert.Contains(t, res.Request.URL.String(), conf.Source().String(config.ViperKeySelfServiceErrorUI))

		assertx.EqualAsJSON(t, x.ErrInvalidCSRFToken, json.RawMessage(gjson.Get(actual, "0").Raw), "%s", actual)
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
		assertx.EqualAsJSON(t, x.ErrInvalidCSRFToken, json.RawMessage(gjson.Get(actual, "error").Raw), "%s", actual)
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
				defer res.Body.Close()

				actual := string(ioutilx.MustReadAll(res.Body))
				assert.EqualValues(t, http.StatusBadRequest, res.StatusCode)
				assert.Contains(t, actual, tc.exp)
			})
		}
	})

	var expectSuccess = func(t *testing.T, isAPI, isSPA bool, hc *http.Client, values func(url.Values)) string {
		return testhelpers.SubmitSettingsForm(t, isAPI, isSPA, hc, publicTS, values, http.StatusOK,
			testhelpers.ExpectURL(isAPI || isSPA, publicTS.URL+settings.RouteSubmitFlow, conf.SelfServiceFlowSettingsUI().String()))
	}

	t.Run("description=should update the password even if no password was set before", func(t *testing.T) {
		var check = func(t *testing.T, actual string, id *identity.Identity) {
			assert.Equal(t, "success", gjson.Get(actual, "state").String(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "ui.nodes.#(name==password).attributes.value").String(), "%s", actual)

			actualIdentity, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), id.ID)
			require.NoError(t, err)
			cfg := string(actualIdentity.Credentials[identity.CredentialsTypePassword].Config)
			assert.Contains(t, cfg, "hashed_password", "%+v", actualIdentity.Credentials)
			require.Len(t, actualIdentity.Credentials[identity.CredentialsTypePassword].Identifiers, 1)
			assert.Contains(t, actualIdentity.Credentials[identity.CredentialsTypePassword].Identifiers[0], "-4")
		}

		var payload = func(v url.Values) {
			v.Set("method", "password")
			v.Set("password", randx.MustString(16, randx.AlphaNum))
		}

		t.Run("type=api", func(t *testing.T) {
			actual := expectSuccess(t, true, false, apiUser2, payload)
			check(t, gjson.Get(actual, "flow").Raw, apiIdentity2)
		})

		t.Run("type=spa", func(t *testing.T) {
			actual := expectSuccess(t, false, true, browserUser2, payload)
			check(t, gjson.Get(actual, "flow").Raw, browserIdentity2)
		})

		t.Run("type=browser", func(t *testing.T) {
			actual := expectSuccess(t, false, false, browserUser2, payload)
			check(t, actual, browserIdentity2)
		})
	})

	t.Run("description=should update the password and perform the correct redirection", func(t *testing.T) {
		rts := testhelpers.NewRedirTS(t, "", conf)
		conf.MustSet(config.ViperKeySelfServiceSettingsAfter+"."+config.DefaultBrowserReturnURL, rts.URL+"/return-ts")
		t.Cleanup(func() {
			conf.MustSet(config.ViperKeySelfServiceSettingsAfter, nil)
		})

		var run = func(t *testing.T, f *kratos.SettingsFlow, isAPI bool, c *http.Client, id *identity.Identity) {
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
}
