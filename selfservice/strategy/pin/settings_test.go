package pin_test

import (
	"context"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/x"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"net/http"
	"net/url"
	"testing"
)

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

func TestSettings(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh/")
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/profile.schema.json")
	testhelpers.StrategyEnable(t, conf, identity.CredentialsTypePassword.String(), true)
	testhelpers.StrategyEnable(t, conf, identity.CredentialsTypePin.String(), true)
	testhelpers.StrategyEnable(t, conf, settings.StrategyProfile, true)

	_ = testhelpers.NewSettingsUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)
	_ = testhelpers.NewLoginUIWith401Response(t, conf)
	conf.MustSet(ctx, config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1m")

	browserIdentity1 := newIdentityWithPassword("john-browser@doe.com")
	apiIdentity1 := newIdentityWithPassword("john-api@doe.com")

	publicTS, _ := testhelpers.NewKratosServer(t, reg)

	browserUser1 := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, browserIdentity1)
	apiUser1 := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, apiIdentity1)

	t.Run("description=should update the pin and clear errors if everything is ok", func(t *testing.T) {
		var check = func(t *testing.T, actual string) {
			assert.Equal(t, "success", gjson.Get(actual, "state").String(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==pin).value").String(), "%s", actual)
			assert.Empty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==pin).messages.0.text").String(), actual)
		}

		var payload = func(v url.Values) {
			v.Set("method", "pin")
			v.Set("pin", x.NewUUID().String())
		}

		t.Run("type=api", func(t *testing.T) {
			actual := testhelpers.SubmitSettingsForm(t, true, false, apiUser1, publicTS, payload, http.StatusOK, publicTS.URL+settings.RouteSubmitFlow)
			check(t, actual)
		})

		t.Run("type=spa", func(t *testing.T) {
			actual := testhelpers.SubmitSettingsForm(t, false, true, browserUser1, publicTS, payload, http.StatusOK, publicTS.URL+settings.RouteSubmitFlow)
			check(t, actual)
		})

		t.Run("type=browser", func(t *testing.T) {
			check(t, testhelpers.SubmitSettingsForm(t, false, false, browserUser1, publicTS, payload, http.StatusOK, conf.SelfServiceFlowSettingsUI(ctx).String()))
		})
	})
}
