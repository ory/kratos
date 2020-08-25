package password_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/ory/x/randx"

	"github.com/ory/viper"
	"github.com/ory/x/pointerx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/httpclient/models"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/settings"
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

	primaryIdentity := &identity.Identity{
		ID: x.NewUUID(), Credentials: map[identity.CredentialsType]identity.Credentials{
			"password": {Type: "password", Identifiers: []string{"john@doe.com"}, Config:
			[]byte(`{"hashed_password":"foo"}`)}},
		Traits: identity.Traits(`{"email":"john@doe.com"}`), SchemaID: configuration.DefaultIdentityTraitsSchemaID}
	secondaryIdentity := &identity.Identity{
		ID: x.NewUUID(), Credentials: map[identity.CredentialsType]identity.Credentials{},
		Traits: identity.Traits(`{}`), SchemaID: configuration.DefaultIdentityTraitsSchemaID}

	publicTS, adminTS := testhelpers.NewKratosServer(t, reg)

	sessionUser1 := session.NewActiveSession(primaryIdentity, testhelpers.NewSessionLifespanProvider(time.Hour), time.Now())
	sessionUser2 := session.NewActiveSession(secondaryIdentity, testhelpers.NewSessionLifespanProvider(time.Hour), time.Now())
	browserUser1 := testhelpers.NewHTTPClientWithSessionCookie(t, reg, sessionUser1)
	browserUser2 := testhelpers.NewHTTPClientWithSessionCookie(t, reg, sessionUser2)
	apiUser1 := testhelpers.NewHTTPClientWithSessionToken(t, reg, sessionUser1)

	adminClient := testhelpers.NewSDKClient(adminTS)

	t.Run("description=should update the password and clear errors after input error occurred", func(t *testing.T) {
		t.Run("description=should fail if password violates policy", func(t *testing.T) {
			var run = func(t *testing.T, isAPI bool, client *http.Client) string {
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
				payload := values.Encode()

				if isAPI {
					payload = `{}`
					for k := range values {
						var err error
						payload, err = sjson.Set(payload, k, values.Get(k))
						require.NoError(t, err)
					}
				}

				t.Logf("Submitting payload: %s", payload)
				actual, _ := testhelpers.SettingsSubmitForm(t, isAPI, form, client, payload,
					expectStatusCodeBetter(isAPI, http.StatusBadRequest, http.StatusNoContent))
				assert.Equal(t, *form.Action, gjson.Get(actual, "methods.password.config.action").String(), "%s", actual)
				return actual
			}

			t.Run("type=api", func(t *testing.T) {
				t.Run("session=with privileged session", func(t *testing.T) {
					viper.Set(configuration.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "5m")

					actual := run(t, true, apiUser1)
					assert.Empty(t, gjson.Get(actual, "methods.password.config.fields.#(name==password).value").String(), "%s", actual)
					assert.NotEmpty(t, gjson.Get(actual, "methods.password.config.fields.#(name==csrf_token).value").String(), "%s", actual)
					assert.Contains(t, gjson.Get(actual, "methods.password.config.fields.#(name==password).messages.0.text").String(), "password can not be used because", "%s", actual)
				})
			})

			t.Run("type=browser", func(t *testing.T) {
				var runInner = func(t *testing.T) {
					actual := run(t, false, browserUser1)
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

			t.Run("description=should update the password and clear errors if everything is ok", func(t *testing.T) {
				t.Run("type=browser", func(t *testing.T) {
					rs := testhelpers.InitializeSettingsFlowViaBrowser(t, browserUser1, publicTS)
					f := rs.Payload.Methods[string(identity.CredentialsTypePassword)].Config
					values := testhelpers.SDKFormFieldsToURLValues(f.Fields)
					values.Set("password", uuid.New().String())
					actual, _ := testhelpers.SettingsSubmitForm(t, false, f, browserUser1, values.Encode(), http.StatusNoContent)

					assert.Equal(t, "success", gjson.Get(actual, "state").String(), "%s", actual)
					assert.Empty(t, gjson.Get(actual, "methods.password.fields.#(name==password).value").String(), "%s", actual)
					assert.Empty(t, gjson.Get(actual, "methods.password.config.fields.#(name==password).messages.0.text").String(), actual)

					actualIdentity, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), primaryIdentity.ID)
					require.NoError(t, err)
					cfg := string(actualIdentity.Credentials[identity.CredentialsTypePassword].Config)
					assert.NotContains(t, cfg, "foo")
					assert.NotEqual(t, `{"hashed_password":"foo"}`, cfg)
				})
			})
		})
	})

	t.Run("description=should update the password even if no password was set before", func(t *testing.T) {
		rs := testhelpers.InitializeSettingsFlowViaBrowser(t, browserUser2, publicTS)

		form := rs.Payload.Methods[string(identity.CredentialsTypePassword)].Config
		values := testhelpers.SDKFormFieldsToURLValues(form.Fields)
		values.Set("password", randx.MustString(16, randx.AlphaNum))
		actual, _ := testhelpers.SettingsSubmitForm(t, false, form, browserUser2, values.Encode(), http.StatusNoContent)

		assert.Equal(t, "success", gjson.Get(actual, "state").String(), "%s", actual)
		assert.Empty(t, gjson.Get(actual, "methods.password.fields.#(name==password).value").String(), "%s", actual)

		actualIdentity, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), secondaryIdentity.ID)
		require.NoError(t, err)
		cfg := string(actualIdentity.Credentials[identity.CredentialsTypePassword].Config)
		assert.Contains(t, cfg, "hashed_password")
		require.Len(t, actualIdentity.Credentials[identity.CredentialsTypePassword].Identifiers, 1)
		assert.Contains(t, actualIdentity.Credentials[identity.CredentialsTypePassword].Identifiers[0], "-4")
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

		rs := testhelpers.InitializeSettingsFlowViaBrowser(t, browserUser1, publicTS)

		form := rs.Payload.Methods[string(identity.CredentialsTypePassword)].Config
		values := testhelpers.SDKFormFieldsToURLValues(form.Fields)
		values.Set("password", randx.MustString(16, randx.AlphaNum))

		res, err := browserUser1.PostForm(pointerx.StringR(form.Action), values)
		require.NoError(t, err)
		defer res.Body.Close()

		assert.True(t, returned)

		actualIdentity, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), primaryIdentity.ID)
		require.NoError(t, err)
		cfg := string(actualIdentity.Credentials[identity.CredentialsTypePassword].Config)
		assert.NotContains(t, cfg, "foo")
		assert.NotEqual(t, `{"hashed_password":"foo"}`, cfg)
	})

	// 	t.Run("case=should fail because of missing CSRF token/type=browser", func(t *testing.T) {
	// 		rr := newRegistrationRequest(t, time.Minute, false)
	// 		body, _ := makeRequest(t, rr.ID, false, url.Values{
	// 			"csrf_token":      {"invalid_token"},
	// 			"traits.username": {"registration-identifier-csrf-browser"},
	// 			"password":        {x.NewUUID().String()},
	// 			"traits.foobar":   {"bar"},
	// 		}.Encode(), http.StatusOK)
	// 		assertx.EqualAsJSON(t, x.ErrInvalidCSRFToken,
	// 			json.RawMessage(gjson.GetBytes(body, "0").Raw), "%s", body)
	// 	})
	//
	// 	t.Run("case=should pass even without CSRF token/type=api", func(t *testing.T) {
	// 		rr := newRegistrationRequest(t, time.Minute, true)
	// 		body, _ := makeRequest(t, rr.ID, true, `{
	//   "csrf_token": "invalid_token",
	//   "traits.username": "registration-identifier-csrf-api",
	//   "traits.foobar": "bar",
	//   "password": "5216f2ef-f14b-4c92-bd91-08c2c2fe1448"
	// }`, http.StatusOK)
	// 		assert.NotEmpty(t, gjson.GetBytes(body, "identity.id").Raw, "%s", body) // registration successful
	// 	})
}
