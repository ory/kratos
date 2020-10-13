package password_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/assertx"
	"github.com/ory/x/httpx"
	"github.com/ory/x/randx"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/httpclient/models"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/strategy/password"
	"github.com/ory/kratos/selfservice/strategy/profile"
	"github.com/ory/kratos/x"
)

func init() {
	internal.RegisterFakes()
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
		Traits:   identity.Traits(`{"email":"` + email + `"}`),
		SchemaID: configuration.DefaultIdentityTraitsSchemaID,
	}
}

func newEmptyIdentity() *identity.Identity {
	return &identity.Identity{
		ID:       x.NewUUID(),
		Traits:   identity.Traits(`{}`),
		SchemaID: configuration.DefaultIdentityTraitsSchemaID,
	}
}

func TestSettings(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	viper.Set(configuration.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh/")
	viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/profile.schema.json")
	testhelpers.StrategyEnable(identity.CredentialsTypePassword.String(), true)
	testhelpers.StrategyEnable(settings.StrategyProfile, true)

	_ = testhelpers.NewSettingsUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)
	_ = testhelpers.NewLoginUIWith401Response(t)
	viper.Set(configuration.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1m")

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
			res, err := c.Do(httpx.MustNewRequest("POST", publicTS.URL+profile.RouteSettings, strings.NewReader(url.Values{"foo": {"bar"}}.Encode()), "application/x-www-form-urlencoded"))
			require.NoError(t, err)
			defer res.Body.Close()
			assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode, "%+v", res.Request)
			assert.Contains(t, res.Request.URL.String(), viper.GetString(configuration.ViperKeySelfServiceLoginUI))
		})

		t.Run("type=api", func(t *testing.T) {
			res, err := c.Do(httpx.MustNewRequest("POST", publicTS.URL+profile.RouteSettings, strings.NewReader(`{"foo":"bar"}`), "application/json"))
			require.NoError(t, err)
			defer res.Body.Close()
			assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode)
		})
	})

	var expectValidationError = func(t *testing.T, isAPI bool, hc *http.Client, values func(url.Values)) string {
		return testhelpers.SubmitSettingsForm(t, isAPI, hc, publicTS, values,
			identity.CredentialsTypePassword.String(),
			testhelpers.ExpectStatusCode(isAPI, http.StatusBadRequest, http.StatusOK),
			testhelpers.ExpectURL(isAPI, publicTS.URL+password.RouteSettings, conf.SelfServiceFlowSettingsUI().String()))
	}

	t.Run("description=should fail if password violates policy", func(t *testing.T) {
		var check = func(t *testing.T, actual string) {
			assert.Empty(t, gjson.Get(actual, "methods.password.config.fields.#(name==password).value").String(), "%s", actual)
			assert.NotEmpty(t, gjson.Get(actual, "methods.password.config.fields.#(name==csrf_token).value").String(), "%s", actual)
			assert.Contains(t, gjson.Get(actual, "methods.password.config.fields.#(name==password).messages.0.text").String(), "password can not be used because", "%s", actual)
		}

		t.Run("session=with privileged session", func(t *testing.T) {
			viper.Set(configuration.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "5m")

			var payload = func(v url.Values) {
				v.Set("password", "123456")
			}

			t.Run("type=api", func(t *testing.T) {
				check(t, expectValidationError(t, true, apiUser1, payload))
			})

			t.Run("type=browser", func(t *testing.T) {
				check(t, expectValidationError(t, false, browserUser1, payload))
			})
		})

		t.Run("session=needs reauthentication", func(t *testing.T) {
			viper.Set(configuration.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1ns")
			_ = testhelpers.NewSettingsLoginAcceptAPIServer(t, adminClient)
			defer testhelpers.NewLoginUIWith401Response(t)
			defer viper.Set(configuration.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "5m")

			var payload = func(v url.Values) {
				v.Set("password", "123456")
			}

			t.Run("type=api/expected=an error because reauth can not be initialized for API clients", func(t *testing.T) {
				actual := testhelpers.SubmitSettingsForm(t, true, apiUser1, publicTS, payload,
					identity.CredentialsTypePassword.String(), http.StatusForbidden, publicTS.URL+password.RouteSettings)
				assertx.EqualAsJSON(t, settings.NewFlowNeedsReAuth(), json.RawMessage(gjson.Get(actual, "error").Raw))
			})

			t.Run("type=browser", func(t *testing.T) {
				check(t, expectValidationError(t, false, browserUser1, payload))
			})
		})
	})

	t.Run("description=should not be able to make requests for another user", func(t *testing.T) {
		t.Run("type=api", func(t *testing.T) {
			rs := testhelpers.InitializeSettingsFlowViaAPI(t, apiUser1, publicTS)
			f := testhelpers.GetSettingsFlowMethodConfig(t, rs.Payload, identity.CredentialsTypePassword.String())
			values := testhelpers.SDKFormFieldsToURLValues(f.Fields)
			values.Set("password", x.NewUUID().String())
			actual, res := testhelpers.SettingsMakeRequest(t, true, f, apiUser2, testhelpers.EncodeFormAsJSON(t, true, values))
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
			assert.Contains(t, gjson.Get(actual, "error.reason").String(), "initiated by another person", "%s", actual)
		})

		t.Run("type=browser", func(t *testing.T) {
			rs := testhelpers.InitializeSettingsFlowViaBrowser(t, browserUser1, publicTS)
			f := testhelpers.GetSettingsFlowMethodConfig(t, rs.Payload, identity.CredentialsTypePassword.String())
			values := testhelpers.SDKFormFieldsToURLValues(f.Fields)
			values.Set("password", x.NewUUID().String())
			actual, res := testhelpers.SettingsMakeRequest(t, false, f, browserUser2, values.Encode())
			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Contains(t, gjson.Get(actual, "0.reason").String(), "initiated by another person", "%s", actual)
		})
	})

	t.Run("description=should update the password and clear errors if everything is ok", func(t *testing.T) {
		var check = func(t *testing.T, actual string) {
			assert.Equal(t, "success", gjson.Get(actual, "state").String(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "methods.password.fields.#(name==password).value").String(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "methods.password.config.fields.#(name==password).messages.0.text").String(), actual)
		}

		var payload = func(v url.Values) {
			v.Set("password", x.NewUUID().String())
		}

		t.Run("type=api", func(t *testing.T) {
			actual := testhelpers.SubmitSettingsForm(t, true, apiUser1, publicTS, payload,
				identity.CredentialsTypePassword.String(), http.StatusOK, publicTS.URL+password.RouteSettings)
			check(t, gjson.Get(actual, "flow").Raw)
		})

		t.Run("type=browser", func(t *testing.T) {
			check(t, testhelpers.SubmitSettingsForm(t, false, browserUser1, publicTS, payload,
				identity.CredentialsTypePassword.String(), http.StatusOK, conf.SelfServiceFlowSettingsUI().String()))
		})
	})

	t.Run("case=should fail because of missing CSRF token/type=browser", func(t *testing.T) {
		rs := testhelpers.InitializeSettingsFlowViaBrowser(t, browserUser1, publicTS)
		f := testhelpers.GetSettingsFlowMethodConfig(t, rs.Payload, identity.CredentialsTypePassword.String())
		values := testhelpers.SDKFormFieldsToURLValues(f.Fields)
		values.Set("password", x.NewUUID().String())
		values.Set("csrf_token", "invalid_token")

		actual, res := testhelpers.SettingsMakeRequest(t, false, f, browserUser1, values.Encode())
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, res.Request.URL.String(), viper.Get(configuration.ViperKeySelfServiceErrorUI))

		assertx.EqualAsJSON(t, x.ErrInvalidCSRFToken, json.RawMessage(gjson.Get(actual, "0").Raw), "%s", actual)
	})

	t.Run("case=should pass even without CSRF token/type=api", func(t *testing.T) {
		rs := testhelpers.InitializeSettingsFlowViaAPI(t, apiUser1, publicTS)
		f := testhelpers.GetSettingsFlowMethodConfig(t, rs.Payload, identity.CredentialsTypePassword.String())
		values := testhelpers.SDKFormFieldsToURLValues(f.Fields)
		values.Set("password", x.NewUUID().String())
		values.Set("csrf_token", "invalid_token")
		actual, res := testhelpers.SettingsMakeRequest(t, true, f, apiUser1, testhelpers.EncodeFormAsJSON(t, true, values))
		assert.Equal(t, http.StatusOK, res.StatusCode)

		assert.Contains(t, res.Request.URL.String(), publicTS.URL+password.RouteSettings)
		assert.NotEmpty(t, gjson.Get(actual, "identity.id").String(), "%s", actual)
	})

	var expectSuccess = func(t *testing.T, isAPI bool, hc *http.Client, values func(url.Values)) string {
		return testhelpers.SubmitSettingsForm(t, isAPI, hc, publicTS, values,
			identity.CredentialsTypePassword.String(), http.StatusOK,
			testhelpers.ExpectURL(isAPI, publicTS.URL+password.RouteSettings, conf.SelfServiceFlowSettingsUI().String()))
	}

	t.Run("description=should update the password even if no password was set before", func(t *testing.T) {
		var check = func(t *testing.T, actual string, id *identity.Identity) {
			assert.Equal(t, "success", gjson.Get(actual, "state").String(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "methods.password.fields.#(name==password).value").String(), "%s", actual)

			actualIdentity, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), id.ID)
			require.NoError(t, err)
			cfg := string(actualIdentity.Credentials[identity.CredentialsTypePassword].Config)
			assert.Contains(t, cfg, "hashed_password")
			require.Len(t, actualIdentity.Credentials[identity.CredentialsTypePassword].Identifiers, 1)
			assert.Contains(t, actualIdentity.Credentials[identity.CredentialsTypePassword].Identifiers[0], "-4")
		}

		var payload = func(v url.Values) {
			v.Set("password", randx.MustString(16, randx.AlphaNum))
		}

		t.Run("type=api", func(t *testing.T) {
			actual := expectSuccess(t, true, apiUser2, payload)
			check(t, gjson.Get(actual, "flow").Raw, apiIdentity2)
		})

		t.Run("type=browser", func(t *testing.T) {
			actual := expectSuccess(t, false, browserUser2, payload)
			check(t, actual, browserIdentity2)
		})
	})

	t.Run("description=should update the password and perform the correct redirection", func(t *testing.T) {
		rts := testhelpers.NewRedirTS(t, "")
		viper.Set(configuration.ViperKeySelfServiceSettingsAfter+"."+configuration.DefaultBrowserReturnURL, rts.URL+"/return-ts")
		defer viper.Set(configuration.ViperKeySelfServiceSettingsAfter, nil)

		var run = func(t *testing.T, f *models.SettingsFlowMethodConfig, isAPI bool, c *http.Client, id *identity.Identity) {
			values := testhelpers.SDKFormFieldsToURLValues(f.Fields)
			values.Set("password", randx.MustString(16, randx.AlphaNum))
			_, res := testhelpers.SettingsMakeRequest(t, isAPI, f, c, testhelpers.EncodeFormAsJSON(t, isAPI, values))
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

		t.Run("type=browser", func(t *testing.T) {
			rs := testhelpers.InitializeSettingsFlowViaBrowser(t, browserUser1, publicTS)
			form := rs.Payload.Methods[string(identity.CredentialsTypePassword)].Config
			run(t, form, false, browserUser1, browserIdentity1)
		})
	})
}
