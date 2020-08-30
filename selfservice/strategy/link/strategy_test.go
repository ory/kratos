package link_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/urlx"

	"github.com/ory/x/sqlxx"

	"github.com/ory/x/assertx"

	"github.com/ory/viper"
	"github.com/ory/x/pointerx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/httpclient/client/admin"
	"github.com/ory/kratos/internal/httpclient/client/common"
	"github.com/ory/kratos/internal/httpclient/models"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/strategy/link"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

func init() {
	internal.RegisterFakes()
}

var identityToRecover = &identity.Identity{
	Credentials: map[identity.CredentialsType]identity.Credentials{
		"password": {Type: "password", Identifiers: []string{"recoverme@ory.sh"}, Config: sqlxx.JSONRawMessage(`{"hashed_password":"foo"}`)}},
	Traits:   identity.Traits(`{"email":"recoverme@ory.sh"}`),
	SchemaID: configuration.DefaultIdentityTraitsSchemaID,
}
var recoveryEmail = gjson.GetBytes(identityToRecover.Traits, "email").String()

func initViper() {
	viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/default.schema.json")
	viper.Set(configuration.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh")
	viper.Set(configuration.ViperKeySelfServiceStrategyConfig+"."+identity.CredentialsTypePassword.String()+".enabled", true)
	viper.Set(configuration.ViperKeySelfServiceStrategyConfig+"."+recovery.StrategyRecoveryTokenName+".enabled", true)
	viper.Set(configuration.ViperKeySelfServiceRecoveryEnabled, true)
}

func TestAdminStrategy(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	initViper()

	_ = testhelpers.NewRecoveryUIFlowEchoServer(t, reg)
	_ = testhelpers.NewSettingsUIFlowEchoServer(t, reg)
	_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)

	publicTS, adminTS := testhelpers.NewKratosServer(t, reg)
	adminSDK := testhelpers.NewSDKClient(adminTS)

	checkLink := func(t *testing.T, l *admin.CreateRecoveryLinkOK, isBefore time.Time) {
		require.Contains(t, *l.Payload.RecoveryLink, publicTS.URL+link.PublicPath)
		rl := urlx.ParseOrPanic(*l.Payload.RecoveryLink)
		assert.NotEmpty(t, rl.Query().Get("token"))
		assert.NotEmpty(t, rl.Query().Get("flow"))
		require.True(t, time.Time(l.Payload.ExpiresAt).Before(isBefore))
	}

	t.Run("description=should not be able to recover an account that does not exist", func(t *testing.T) {
		_, err := adminSDK.Admin.CreateRecoveryLink(admin.NewCreateRecoveryLinkParams().WithBody(
			&models.CreateRecoveryLink{IdentityID: models.UUID(x.NewUUID().String())}))
		require.IsType(t, err, new(admin.CreateRecoveryLinkNotFound), "%T", err)
	})

	t.Run("description=should not be able to recover an account that does not have a recovery email", func(t *testing.T) {
		id := identity.Identity{Traits: identity.Traits(`{}`)}
		require.NoError(t, reg.IdentityManager().Create(context.Background(),
			&id, identity.ManagerAllowWriteProtectedTraits))

		_, err := adminSDK.Admin.CreateRecoveryLink(admin.NewCreateRecoveryLinkParams().WithBody(
			&models.CreateRecoveryLink{IdentityID: models.UUID(id.ID.String())}))
		require.IsType(t, err, new(admin.CreateRecoveryLinkBadRequest), "%T", err)
	})

	t.Run("description=should create a valid recovery link and set the expiry time and not be able to recover the account", func(t *testing.T) {
		id := identity.Identity{Traits: identity.Traits(`{"email":"recover.expired@ory.sh"}`)}

		require.NoError(t, reg.IdentityManager().Create(context.Background(),
			&id, identity.ManagerAllowWriteProtectedTraits))

		rl, err := adminSDK.Admin.CreateRecoveryLink(admin.NewCreateRecoveryLinkParams().
			WithBody(&models.CreateRecoveryLink{
				IdentityID: models.UUID(id.ID.String()),
				ExpiresIn:  "100ms",
			}))
		require.NoError(t, err)

		time.Sleep(time.Millisecond * 100)
		checkLink(t, rl, time.Now().Add(conf.SelfServiceFlowRecoveryRequestLifespan()))
		res, err := publicTS.Client().Get(*rl.Payload.RecoveryLink)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, res.StatusCode)

		// We end up here because the link is expired.
		assert.Contains(t, res.Request.URL.Path, "/recover")
	})

	t.Run("description=should create a valid recovery link and set the expiry time as well and recover the account", func(t *testing.T) {
		id := identity.Identity{Traits: identity.Traits(`{"email":"` + recoveryEmail + `"}`)}

		require.NoError(t, reg.IdentityManager().Create(context.Background(),
			&id, identity.ManagerAllowWriteProtectedTraits))

		rl, err := adminSDK.Admin.CreateRecoveryLink(admin.NewCreateRecoveryLinkParams().
			WithBody(&models.CreateRecoveryLink{IdentityID: models.UUID(id.ID.String())}))
		require.NoError(t, err)

		checkLink(t, rl, time.Now().Add(conf.SelfServiceFlowRecoveryRequestLifespan()+time.Second))
		res, err := publicTS.Client().Get(*rl.Payload.RecoveryLink)
		require.NoError(t, err)

		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowSettingsUI().String())
		assert.Equal(t, http.StatusOK, res.StatusCode)
		testhelpers.LogJSON(t, rl.Payload)

		sr, err := adminSDK.Common.GetSelfServiceSettingsFlow(
			common.NewGetSelfServiceSettingsFlowParams().
				WithID(res.Request.URL.Query().Get("flow")))
		require.NoError(t, err, "%s", res.Request.URL.String())

		require.Len(t, sr.Payload.Messages, 1)
		assert.Equal(t, "You successfully recovered your account. Please change your password or set up an alternative login method (e.g. social sign in) within the next 60.00 minutes.", sr.Payload.Messages[0].Text)
	})
}

func TestStrategy(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	initViper()

	_ = testhelpers.NewRecoveryUIFlowEchoServer(t, reg)
	_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
	_ = testhelpers.NewSettingsUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)

	public, _ := testhelpers.NewKratosServer(t, reg)
	sdk := testhelpers.NewSDKClient(public)

	require.NoError(t, reg.IdentityManager().Create(context.Background(), identityToRecover,
		identity.ManagerAllowWriteProtectedTraits))

	var csrfField = &models.FormField{Name: pointerx.String("csrf_token"), Required: true,
		Type: pointerx.String("hidden"), Value: x.FakeCSRFToken}

	var expect = func(t *testing.T, isAPI bool, values func(url.Values), c int) string {
		hc := testhelpers.NewDebugClient(t)
		if !isAPI {
			hc = testhelpers.NewDebugClient(t)
		}
		return testhelpers.SubmitRecoveryForm(t, isAPI, hc, public, values, recovery.StrategyRecoveryTokenName, c,
			testhelpers.ExpectURL(isAPI, public.URL+link.PublicPath, conf.SelfServiceFlowRecoveryUI().String()))
	}

	var expectValidationError = func(t *testing.T, isAPI bool, values func(url.Values)) string {
		return expect(t, isAPI, values, testhelpers.ExpectStatusCode(isAPI, http.StatusBadRequest, http.StatusOK))
	}

	var expectSuccess = func(t *testing.T, isAPI bool, values func(url.Values)) string {
		return expect(t, isAPI, values, http.StatusOK)
	}

	t.Run("description=should set all the correct recovery payloads", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)
		rs := testhelpers.GetRecoveryRequest(t, c, public)
		assert.Contains(t, rs.Payload.Methods, recovery.StrategyRecoveryTokenName)
		method := rs.Payload.Methods[recovery.StrategyRecoveryTokenName]

		assert.EqualValues(t, models.FormFields{csrfField,
			{Name: pointerx.String("email"), Required: true, Type: pointerx.String("email")},
		}, method.Config.Fields)
		assert.EqualValues(t, public.URL+link.PublicPath+"?flow="+string(rs.Payload.ID), *method.Config.Action)
		assert.Empty(t, method.Config.Messages)
		assert.Empty(t, rs.Payload.Messages)
	})

	t.Run("description=should require an email to be sent", func(t *testing.T) {
		var check = func(t *testing.T, actual string) {
			assert.EqualValues(t, recovery.StrategyRecoveryTokenName, gjson.Get(actual, "active").String(), "%s", actual)
			assert.EqualValues(t, "Property email is missing.",
				gjson.Get(actual, "methods.link.config.fields.#(name==email).messages.0.text").String(),
				"%s", actual)
		}

		var values = func(v url.Values) {
			v.Del("email")
		}

		t.Run("type=browser", func(t *testing.T) {
			check(t, expectValidationError(t, false, values))
		})

		t.Run("type=api", func(t *testing.T) {
			check(t, expectValidationError(t, true, values))
		})
	})

	t.Run("description=should try to recover an email that does not exist", func(t *testing.T) {
		var email string
		var check = func(t *testing.T, actual string) {
			assert.EqualValues(t, recovery.StrategyRecoveryTokenName, gjson.Get(actual, "active").String(), "%s", actual)
			assert.EqualValues(t, email, gjson.Get(actual, "methods.link.config.fields.#(name==email).value").String(), "%s", actual)
			assertx.EqualAsJSON(t, text.NewRecoveryEmailSent(), json.RawMessage(gjson.Get(actual, "messages.0").Raw))

			message := testhelpers.CourierExpectMessage(t, reg, email, "Account access attempted")
			assert.Contains(t, message.Body, "If this was you, check if you signed up using a different address.")
		}

		var values = func(v url.Values) {
			v.Set("email", email)
		}

		t.Run("type=browser", func(t *testing.T) {
			email = x.NewUUID().String() + "@ory.sh"
			check(t, expectSuccess(t, false, values))
		})

		t.Run("type=api", func(t *testing.T) {
			email = x.NewUUID().String() + "@ory.sh"
			check(t, expectSuccess(t, true, values))
		})
	})

	t.Run("description=should recover an account", func(t *testing.T) {

		var check = func(t *testing.T, actual string) {
			assert.EqualValues(t, recovery.StrategyRecoveryTokenName, gjson.Get(actual, "active").String(), "%s", actual)
			assert.EqualValues(t, recoveryEmail, gjson.Get(actual, "methods.link.config.fields.#(name==email).value").String(), "%s", actual)
			assertx.EqualAsJSON(t, text.NewRecoveryEmailSent(), json.RawMessage(gjson.Get(actual, "messages.0").Raw))

			message := testhelpers.CourierExpectMessage(t, reg, recoveryEmail, "Recover access to your account")
			assert.Contains(t, message.Body, "please recover access to your account by clicking the following link")

			recoveryLink := testhelpers.CourierExpectLinkInMessage(t, message, 1)

			assert.Contains(t, recoveryLink, public.URL+link.PublicPath)
			assert.Contains(t, recoveryLink, "token=")

			cl := testhelpers.NewClientWithCookies(t)
			res, err := cl.Get(recoveryLink)
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowSettingsUI().String())

			body := x.MustReadAll(res.Body)
			assert.Equal(t, text.NewRecoverySuccessful(time.Now().Add(time.Hour)).Text,
				gjson.GetBytes(body, "messages.0.text").String())
		}

		var values = func(v url.Values) {
			v.Set("email", recoveryEmail)
		}

		t.Run("type=browser", func(t *testing.T) {
			check(t, expectSuccess(t, false, values))
		})

		t.Run("type=api", func(t *testing.T) {
			check(t, expectSuccess(t, true, values))
		})
	})

	t.Run("description=should not be able to use an invalid link", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)
		res, err := c.Get(public.URL + link.PublicPath + "?token=i-do-not-exist")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowRecoveryUI().String()+"?flow=")

		sr, err := sdk.Common.GetSelfServiceRecoveryFlow(
			common.NewGetSelfServiceRecoveryFlowParams().WithHTTPClient(c).
				WithID(res.Request.URL.Query().Get("flow")))
		require.NoError(t, err)

		require.Len(t, sr.Payload.Messages, 1)
		assert.Equal(t, "The recovery token is invalid or has already been used. Please retry the flow.", sr.Payload.Messages[0].Text)
	})

	t.Run("description=should not be able to use an outdated link", func(t *testing.T) {
		viper.Set(configuration.ViperKeySelfServiceRecoveryRequestLifespan, time.Millisecond*200)
		t.Cleanup(func() {
			viper.Set(configuration.ViperKeySelfServiceRecoveryRequestLifespan, time.Minute)
		})

		c := testhelpers.NewClientWithCookies(t)
		rs := testhelpers.GetRecoveryRequest(t, c, public)
		method := rs.Payload.Methods[recovery.StrategyRecoveryTokenName].Config

		time.Sleep(time.Millisecond * 201)

		res, err := c.PostForm(pointerx.StringR(method.Action), url.Values{"email": {recoveryEmail}})
		require.NoError(t, err)
		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.NotContains(t, res.Request.URL.String(), "flow="+rs.Payload.ID)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowRecoveryUI().String())
	})

	t.Run("description=should not be able to use an outdated request", func(t *testing.T) {
		viper.Set(configuration.ViperKeySelfServiceRecoveryRequestLifespan, time.Millisecond*200)
		t.Cleanup(func() {
			viper.Set(configuration.ViperKeySelfServiceRecoveryRequestLifespan, time.Minute)
		})

		body := expectSuccess(t, false, func(v url.Values) {
			v.Set("email", recoveryEmail)
		})

		message := testhelpers.CourierExpectMessage(t, reg, recoveryEmail, "Recover access to your account")
		assert.Contains(t, message.Body, "please recover access to your account by clicking the following link")

		recoveryLink := testhelpers.CourierExpectLinkInMessage(t, message, 1)

		time.Sleep(time.Millisecond * 201)

		c := testhelpers.NewClientWithCookies(t)
		res, err := c.Get(recoveryLink)
		require.NoError(t, err)

		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowRecoveryUI().String())
		assert.NotContains(t, res.Request.URL.String(), gjson.Get(body, "id").String())

		sr, err := sdk.Common.GetSelfServiceRecoveryFlow(
			common.NewGetSelfServiceRecoveryFlowParams().WithHTTPClient(c).
				WithID(res.Request.URL.Query().Get("flow")))
		require.NoError(t, err)

		require.Len(t, sr.Payload.Messages, 1)
		assert.Contains(t, sr.Payload.Messages[0].Text, "The recovery flow expired")
	})
}
