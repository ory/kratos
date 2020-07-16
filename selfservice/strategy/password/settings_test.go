package password_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/randx"

	"github.com/ory/viper"
	"github.com/ory/x/pointerx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/settings"
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
		ID: x.NewUUID(),
		Credentials: map[identity.CredentialsType]identity.Credentials{
			"password": {Type: "password", Identifiers: []string{"john@doe.com"}, Config: []byte(`{"hashed_password":"foo"}`)},
		},
		Traits:   identity.Traits(`{"email":"john@doe.com"}`),
		SchemaID: configuration.DefaultIdentityTraitsSchemaID,
	}
	secondaryIdentity := &identity.Identity{
		ID:          x.NewUUID(),
		Credentials: map[identity.CredentialsType]identity.Credentials{},
		Traits:      identity.Traits(`{}`),
		SchemaID:    configuration.DefaultIdentityTraitsSchemaID,
	}
	publicTS, adminTS, clients := testhelpers.NewSettingsAPIServer(t, reg, map[string]*identity.Identity{
		"primary": primaryIdentity, "secondary": secondaryIdentity})

	primaryUser := clients["primary"]
	secondaryUser := clients["secondary"]
	adminClient := testhelpers.NewSDKClient(adminTS)

	t.Run("description=should update the password and clear errors after input error occurred", func(t *testing.T) {
		rs := testhelpers.GetSettingsRequest(t, primaryUser, publicTS)

		form := rs.Payload.Methods[string(identity.CredentialsTypePassword)].Config
		values := testhelpers.SDKFormFieldsToURLValues(form.Fields)

		t.Run("description=should fail if password violates policy", func(t *testing.T) {
			var run = func(t *testing.T) {
				values.Set("password", "123456")
				actual, _ := testhelpers.SettingsSubmitForm(t, form, primaryUser, values)

				assert.Equal(t, *form.Action, gjson.Get(actual, "methods.password.config.action").String(), "%s", actual)
				assert.Empty(t, gjson.Get(actual, "methods.password.config.fields.#(name==password).value").String(), "%s", actual)
				assert.NotEmpty(t, gjson.Get(actual, "methods.password.config.fields.#(name==csrf_token).value").String(), "%s", actual)
				assert.Contains(t, gjson.Get(actual, "methods.password.config.fields.#(name==password).messages.0.text").String(), "password can not be used because", "%s", actual)
			}

			t.Run("session=with privileged session", run)

			t.Run("session=needs reauthentication", func(t *testing.T) {
				viper.Set(configuration.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1ns")

				_ = testhelpers.NewSettingsLoginAcceptAPIServer(t, adminClient)
				t.Cleanup(func() {
					viper.Set(configuration.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "5m")
				})
				run(t)
			})
		})

		t.Run("description=should update the password and clear errors if everything is ok", func(t *testing.T) {
			values.Set("password", uuid.New().String())
			actual, _ := testhelpers.SettingsSubmitForm(t, form, primaryUser, values)

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

	t.Run("description=should update the password even if no password was set before", func(t *testing.T) {
		rs := testhelpers.GetSettingsRequest(t, secondaryUser, publicTS)

		form := rs.Payload.Methods[string(identity.CredentialsTypePassword)].Config
		values := testhelpers.SDKFormFieldsToURLValues(form.Fields)
		values.Set("password", randx.MustString(16, randx.AlphaNum))
		actual, _ := testhelpers.SettingsSubmitForm(t, form, secondaryUser, values)

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

		rs := testhelpers.GetSettingsRequest(t, primaryUser, publicTS)

		form := rs.Payload.Methods[string(identity.CredentialsTypePassword)].Config
		values := testhelpers.SDKFormFieldsToURLValues(form.Fields)
		values.Set("password", randx.MustString(16, randx.AlphaNum))

		res, err := primaryUser.PostForm(pointerx.StringR(form.Action), values)
		require.NoError(t, err)
		defer res.Body.Close()

		assert.True(t, returned)

		actualIdentity, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), primaryIdentity.ID)
		require.NoError(t, err)
		cfg := string(actualIdentity.Credentials[identity.CredentialsTypePassword].Config)
		assert.NotContains(t, cfg, "foo")
		assert.NotEqual(t, `{"hashed_password":"foo"}`, cfg)
	})
}
