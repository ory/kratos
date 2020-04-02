package password_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

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

func TestProfile(t *testing.T) {
	_, reg := internal.NewRegistryDefault(t)
	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/profile.schema.json")

	_ = testhelpers.NewSettingsUITestServer(t)
	_ = testhelpers.NewErrorTestServer(t, reg)
	viper.Set(configuration.ViperKeySelfServicePrivilegedAuthenticationAfter, "1m")

	primaryIdentity := &identity.Identity{
		ID: x.NewUUID(),
		Credentials: map[identity.CredentialsType]identity.Credentials{
			"password": {Type: "password", Identifiers: []string{"john@doe.com"}, Config: json.RawMessage(`{"hashed_password":"foo"}`)},
		},
		Traits:         identity.Traits(`{"email":"john@doe.com"}`),
		TraitsSchemaID: configuration.DefaultIdentityTraitsSchemaID,
	}
	secondaryIdentity := &identity.Identity{
		ID:             x.NewUUID(),
		Credentials:    map[identity.CredentialsType]identity.Credentials{},
		Traits:         identity.Traits(`{}`),
		TraitsSchemaID: configuration.DefaultIdentityTraitsSchemaID,
	}
	publicTS, _ := testhelpers.NewSettingsAPIServer(t, reg, []identity.Identity{*primaryIdentity, *secondaryIdentity})
	primaryUser := testhelpers.NewSessionClient(t, publicTS.URL+"/sessions/set/0")
	secondaryUser := testhelpers.NewSessionClient(t, publicTS.URL+"/sessions/set/1")

	t.Run("description=should fail to update when session is too old", func(t *testing.T) {
		viper.Set(configuration.ViperKeySelfServicePrivilegedAuthenticationAfter, "1ns")
		t.Cleanup(func() {
			viper.Set(configuration.ViperKeySelfServicePrivilegedAuthenticationAfter, "1m")
		})
		rs := testhelpers.GetSettingsRequest(t, primaryUser, publicTS)

		form := rs.Payload.Methods[string(identity.CredentialsTypePassword)].Config
		values := testhelpers.SDKFormFieldsToURLValues(form.Fields)
		values.Set("password", "123456")

		res, err := primaryUser.PostForm(pointerx.StringR(form.Action), values)
		require.NoError(t, err)
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)

		assert.Contains(t, gjson.GetBytes(body, "0.reason").String(), "session is too old and thus not allowed to update these fields. Please re-authenticate")
		assert.Equal(t, int64(http.StatusForbidden), gjson.GetBytes(body, "0.code").Int())
	})

	t.Run("description=should come back with form errors if the password data is invalid", func(t *testing.T) {
		rs := testhelpers.GetSettingsRequest(t, primaryUser, publicTS)

		form := rs.Payload.Methods[string(identity.CredentialsTypePassword)].Config
		values := testhelpers.SDKFormFieldsToURLValues(form.Fields)
		values.Set("password", "123456")
		actual, _ := testhelpers.SettingsSubmitForm(t, form, primaryUser, values)

		assert.Equal(t, *form.Action, gjson.Get(actual, "methods.password.config.action").String(), "%s", actual)
		assert.Empty(t, gjson.Get(actual, "methods.password.config.fields.#(name==password).value").String(), "%s", actual)
		assert.NotEmpty(t, gjson.Get(actual, "methods.password.config.fields.#(name==csrf_token).value").String(), "%s", actual)
		assert.Contains(t, gjson.Get(actual, "methods.password.config.fields.#(name==password).errors.0.message").String(), "the password does not fulfill the password policy because", "%s", actual)
	})

	t.Run("description=should update the password if everything is ok", func(t *testing.T) {
		rs := testhelpers.GetSettingsRequest(t, primaryUser, publicTS)

		form := rs.Payload.Methods[string(identity.CredentialsTypePassword)].Config
		values := testhelpers.SDKFormFieldsToURLValues(form.Fields)
		values.Set("password", uuid.New().String())
		actual, _ := testhelpers.SettingsSubmitForm(t, form, primaryUser, values)

		assert.Equal(t, true, gjson.Get(actual, "update_successful").Bool(), "%s", actual)
		assert.Empty(t, gjson.Get(actual, "methods.password.fields.#(name==password).value").String(), "%s", actual)

		actualIdentity, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), primaryIdentity.ID)
		require.NoError(t, err)
		cfg := string(actualIdentity.Credentials[identity.CredentialsTypePassword].Config)
		assert.NotContains(t, cfg, "foo")
		assert.NotEqual(t, `{"hashed_password":"foo"}`, cfg)
	})

	t.Run("description=should update the password even if no password was set before", func(t *testing.T) {
		rs := testhelpers.GetSettingsRequest(t, secondaryUser, publicTS)

		form := rs.Payload.Methods[string(identity.CredentialsTypePassword)].Config
		values := testhelpers.SDKFormFieldsToURLValues(form.Fields)
		values.Set("password", uuid.New().String())
		actual, _ := testhelpers.SettingsSubmitForm(t, form, secondaryUser, values)

		assert.Equal(t, true, gjson.Get(actual, "update_successful").Bool(), "%s", actual)
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

		viper.Set(configuration.ViperKeySelfServiceSettingsAfterConfig+"."+settings.StrategyTraitsID, testhelpers.HookConfigRedirectTo(t, rts.URL+"/return-ts"))
		t.Cleanup(func() {
			viper.Set(configuration.ViperKeySelfServiceLoginAfterConfig+"."+string(identity.CredentialsTypePassword), nil)
		})

		rs := testhelpers.GetSettingsRequest(t, primaryUser, publicTS)

		form := rs.Payload.Methods[string(identity.CredentialsTypePassword)].Config
		values := testhelpers.SDKFormFieldsToURLValues(form.Fields)
		values.Set("password", uuid.New().String())

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
