package password_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/ory/x/assertx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/ory/x/randx"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/httpclient/models"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/strategy/password"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

func init() {
	internal.RegisterFakes()
}

func TestSettings(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	viper.Set(configuration.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh/")
	viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/profile.schema.json")
	testhelpers.StrategyEnable(identity.CredentialsTypePassword.String(), true)
	testhelpers.StrategyEnable(settings.StrategyProfile, true)

	_ = testhelpers.NewSettingsUITestServer(t)
	_ = testhelpers.NewErrorTestServer(t, reg)
	viper.Set(configuration.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1m")

	browserIdentity1 := &identity.Identity{
		ID: x.NewUUID(), Credentials: map[identity.CredentialsType]identity.Credentials{
			"password": {Type: "password", Identifiers: []string{"john-browser@doe.com"}, Config:
			[]byte(`{"hashed_password":"foo"}`)}},
		Traits: identity.Traits(`{"email":"john-browser@doe.com"}`), SchemaID: configuration.DefaultIdentityTraitsSchemaID}
	apiIdentity1 := &identity.Identity{
		ID: x.NewUUID(), Credentials: map[identity.CredentialsType]identity.Credentials{
			"password": {Type: "password", Identifiers: []string{"john-api@doe.com"}, Config:
			[]byte(`{"hashed_password":"foo"}`)}},
		Traits: identity.Traits(`{"email":"john-api@doe.com"}`), SchemaID: configuration.DefaultIdentityTraitsSchemaID}
	browserIdentity2 := &identity.Identity{
		ID: x.NewUUID(), Credentials: map[identity.CredentialsType]identity.Credentials{},
		Traits: identity.Traits(`{}`), SchemaID: configuration.DefaultIdentityTraitsSchemaID}
	apiIdentity2 := &identity.Identity{
		ID: x.NewUUID(), Credentials: map[identity.CredentialsType]identity.Credentials{},
		Traits: identity.Traits(`{}`), SchemaID: configuration.DefaultIdentityTraitsSchemaID}

	publicTS, adminTS := testhelpers.NewKratosServer(t, reg)

	browserUser1 := testhelpers.NewHTTPClientWithSessionCookie(t, reg, session.NewActiveSession(browserIdentity1, testhelpers.NewSessionLifespanProvider(time.Hour), time.Now()))
	browserUser2 := testhelpers.NewHTTPClientWithSessionCookie(t, reg, session.NewActiveSession(browserIdentity2, testhelpers.NewSessionLifespanProvider(time.Hour), time.Now()))
	apiUser1 := testhelpers.NewHTTPClientWithSessionToken(t, reg, session.NewActiveSession(apiIdentity1, testhelpers.NewSessionLifespanProvider(time.Hour), time.Now()))
	apiUser2 := testhelpers.NewHTTPClientWithSessionToken(t, reg, session.NewActiveSession(apiIdentity2, testhelpers.NewSessionLifespanProvider(time.Hour), time.Now()))

	adminClient := testhelpers.NewSDKClient(adminTS)

	var encodeForm = func(t *testing.T, isApi bool, values url.Values) (payload string) {
		if !isApi {
			return values.Encode()
		}
		payload = "{}"
		for k := range values {
			var err error
			payload, err = sjson.Set(payload, k, values.Get(k))
			require.NoError(t, err)
		}
		return payload
	}

	t.Run("description=should update the password and clear errors after input error occurred", func(t *testing.T) {
		t.Run("description=should fail if password violates policy", func(t *testing.T) {
			var run = func(t *testing.T, isAPI bool, client *http.Client, ec int) string {
				var form *models.FlowMethodConfig
				if isAPI {
					rs := testhelpers.InitializeSettingsFlowViaAPI(t, client, publicTS)
					form = rs.Payload.Methods[string(identity.CredentialsTypePassword)].Config
				} else {
					rs := testhelpers.InitializeSettingsFlowViaBrowser(t, client, publicTS)
					form = rs.Payload.Methods[string(identity.CredentialsTypePassword)].Config
				}

				values := testhelpers.SDKFormFieldsToURLValues(form.Fields)
				values.Set("password", "123456")

				actual, _ := testhelpers.SettingsSubmitForm(t, isAPI, form, client, encodeForm(t, isAPI, values), ec)
				assert.Equal(t, *form.Action, gjson.Get(actual, "methods.password.config.action").String(), "%s", actual)
				return actual
			}

			t.Run("type=api", func(t *testing.T) {
				t.Run("session=with privileged session", func(t *testing.T) {
					viper.Set(configuration.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "5m")

					actual := run(t, true, apiUser1, http.StatusBadRequest)
					assert.Empty(t, gjson.Get(actual, "methods.password.config.fields.#(name==password).value").String(), "%s", actual)
					assert.NotEmpty(t, gjson.Get(actual, "methods.password.config.fields.#(name==csrf_token).value").String(), "%s", actual)
					assert.Contains(t, gjson.Get(actual, "methods.password.config.fields.#(name==password).messages.0.text").String(), "password can not be used because", "%s", actual)
				})

				t.Run("session=needs reauthentication", func(t *testing.T) {
					viper.Set(configuration.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1ns")
					_ = testhelpers.NewSettingsLoginAcceptAPIServer(t, adminClient)
					defer viper.Set(configuration.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "5m")

					actual := run(t, true, apiUser1, http.StatusForbidden)
					assert.Empty(t, gjson.Get(actual, "methods.password.config.fields.#(name==password).value").String(), "%s", actual)
				})
			})

			t.Run("type=browser", func(t *testing.T) {
				var runInner = func(t *testing.T) {
					actual := run(t, false, browserUser1, http.StatusNoContent)
					assert.Empty(t, gjson.Get(actual, "methods.password.config.fields.#(name==password).value").String(), "%s", actual)
					assert.NotEmpty(t, gjson.Get(actual, "methods.password.config.fields.#(name==csrf_token).value").String(), "%s", actual)
					assert.Contains(t, gjson.Get(actual, "methods.password.config.fields.#(name==password).messages.0.text").String(), "password can not be used because", "%s", actual)
				}

				t.Run("session=with privileged session", func(t *testing.T) {
					viper.Set(configuration.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "5m")
					runInner(t)
				})

				t.Run("session=needs reauthentication", func(t *testing.T) {
					viper.Set(configuration.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1ns")
					_ = testhelpers.NewSettingsLoginAcceptAPIServer(t, adminClient)
					defer viper.Set(configuration.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "5m")

					runInner(t)
				})
			})
		})
	})

	t.Run("description=should not be able to make requests for another user", func(t *testing.T) {
		t.Run("type=api", func(t *testing.T) {
			rs := testhelpers.InitializeSettingsFlowViaAPI(t, apiUser1, publicTS)
			f := rs.Payload.Methods[string(identity.CredentialsTypePassword)].Config
			values := testhelpers.SDKFormFieldsToURLValues(f.Fields)
			values.Set("password", x.NewUUID().String())
			actual, res := testhelpers.SettingsMakeRequest(t, true, f, apiUser2, encodeForm(t, true, values))
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
			assert.Contains(t, gjson.Get(actual, "error.reason").String(), "initiated by another person", "%s", actual)
		})

		t.Run("type=browser", func(t *testing.T) {
			rs := testhelpers.InitializeSettingsFlowViaBrowser(t, browserUser1, publicTS)
			f := rs.Payload.Methods[string(identity.CredentialsTypePassword)].Config
			values := testhelpers.SDKFormFieldsToURLValues(f.Fields)
			values.Set("password", x.NewUUID().String())
			actual, res := testhelpers.SettingsMakeRequest(t, false, f, browserUser2, values.Encode())
			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Contains(t, gjson.Get(actual, "0.reason").String(), "initiated by another person", "%s", actual)
		})
	})

	t.Run("description=should update the password and clear errors if everything is ok", func(t *testing.T) {
		var run = func(t *testing.T, f *models.FlowMethodConfig, isAPI bool, c *http.Client, id *identity.Identity) {
			values := testhelpers.SDKFormFieldsToURLValues(f.Fields)
			values.Set("password", x.NewUUID().String())
			actual, _ := testhelpers.SettingsSubmitForm(t, isAPI, f, c, encodeForm(t, isAPI, values), expectStatusCode(isAPI, http.StatusOK, http.StatusNoContent))

			assert.Equal(t, "success", gjson.Get(actual, "state").String(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "methods.password.fields.#(name==password).value").String(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "methods.password.config.fields.#(name==password).messages.0.text").String(), actual)

			actualIdentity, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), id.ID)
			require.NoError(t, err)
			assert.NotEqualValues(t, browserIdentity1.Credentials[identity.CredentialsTypePassword].Config, actualIdentity.Credentials[identity.CredentialsTypePassword].Config)
		}

		t.Run("type=api", func(t *testing.T) {
			rs := testhelpers.InitializeSettingsFlowViaAPI(t, apiUser1, publicTS)
			f := rs.Payload.Methods[string(identity.CredentialsTypePassword)].Config
			run(t, f, true, apiUser1, apiIdentity1)
		})

		t.Run("type=browser", func(t *testing.T) {
			rs := testhelpers.InitializeSettingsFlowViaBrowser(t, browserUser1, publicTS)
			f := rs.Payload.Methods[string(identity.CredentialsTypePassword)].Config
			run(t, f, false, browserUser1, browserIdentity1)
		})
	})

	t.Run("case=should fail because of missing CSRF token/type=browser", func(t *testing.T) {
		rs := testhelpers.InitializeSettingsFlowViaBrowser(t, browserUser1, publicTS)
		f := rs.Payload.Methods[string(identity.CredentialsTypePassword)].Config
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
		f := rs.Payload.Methods[string(identity.CredentialsTypePassword)].Config
		values := testhelpers.SDKFormFieldsToURLValues(f.Fields)
		values.Set("password", x.NewUUID().String())
		values.Set("csrf_token", "invalid_token")
		actual, res := testhelpers.SettingsMakeRequest(t, true, f, apiUser1, encodeForm(t, true, values))
		assert.Equal(t, http.StatusOK, res.StatusCode)

		assert.Contains(t, res.Request.URL.String(), publicTS.URL+password.RouteSettings)
		assert.NotEmpty(t, gjson.Get(actual, "identity.id").String(), "%s", actual)
	})

	t.Run("description=should update the password even if no password was set before", func(t *testing.T) {
		var run = func(t *testing.T, f *models.FlowMethodConfig, isAPI bool, c *http.Client, id *identity.Identity) {
			values := testhelpers.SDKFormFieldsToURLValues(f.Fields)
			values.Set("password", randx.MustString(16, randx.AlphaNum))
			actual, _ := testhelpers.SettingsSubmitForm(t, isAPI, f, c, encodeForm(t, isAPI, values),
				expectStatusCode(isAPI, http.StatusOK, http.StatusNoContent))

			assert.Equal(t, "success", gjson.Get(actual, "state").String(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "methods.password.fields.#(name==password).value").String(), "%s", actual)

			actualIdentity, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), id.ID)
			require.NoError(t, err)
			cfg := string(actualIdentity.Credentials[identity.CredentialsTypePassword].Config)
			assert.Contains(t, cfg, "hashed_password")
			require.Len(t, actualIdentity.Credentials[identity.CredentialsTypePassword].Identifiers, 1)
			assert.Contains(t, actualIdentity.Credentials[identity.CredentialsTypePassword].Identifiers[0], "-4")
		}

		t.Run("type=api", func(t *testing.T) {
			rs := testhelpers.InitializeSettingsFlowViaBrowser(t, apiUser2, publicTS)
			f := rs.Payload.Methods[string(identity.CredentialsTypePassword)].Config
			run(t, f, true, apiUser2, apiIdentity2)
		})

		t.Run("type=browser", func(t *testing.T) {
			rs := testhelpers.InitializeSettingsFlowViaBrowser(t, browserUser2, publicTS)
			f := rs.Payload.Methods[string(identity.CredentialsTypePassword)].Config
			run(t, f, false, browserUser2, browserIdentity2)
		})
	})

	t.Run("description=should update the password and execute hooks", func(t *testing.T) {
		var returned bool
		router := httprouter.New()
		router.GET("/return-ts", func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
			returned = true
		})
		rts := httptest.NewServer(router)
		defer rts.Close()

		viper.Set(configuration.ViperKeySelfServiceSettingsAfter+"."+configuration.DefaultBrowserReturnURL, rts.URL+"/return-ts")
		t.Cleanup(func() {
			viper.Set(configuration.ViperKeySelfServiceSettingsAfter, nil)
		})

		var run = func(t *testing.T, f *models.FlowMethodConfig, isAPI bool, c *http.Client, id *identity.Identity) {
			returned = false

			values := testhelpers.SDKFormFieldsToURLValues(f.Fields)
			values.Set("password", randx.MustString(16, randx.AlphaNum))
			_, _ = testhelpers.SettingsMakeRequest(t, isAPI, f, c, encodeForm(t, isAPI, values))

			assert.True(t, returned)
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
