package recoverytoken_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	"github.com/ory/kratos/selfservice/strategy/recoverytoken"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

func init() {
	internal.RegisterFakes()
}

var identityToRecover = &identity.Identity{
	Credentials: map[identity.CredentialsType]identity.Credentials{
		"password": {Type: "password", Identifiers: []string{"recover@ory.sh"}, Config: sqlxx.JSONRawMessage(`{"hashed_password":"foo"}`)}},
	Traits:   identity.Traits(`{"email":"recover@ory.sh"}`),
	SchemaID: configuration.DefaultIdentityTraitsSchemaID,
}

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

	_ = testhelpers.NewRecoveryUITestServer(t)
	_ = testhelpers.NewLoginUIRequestEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)

	publicTS, adminTS := testhelpers.NewKratosServer(t, reg)
	adminSDK := testhelpers.NewSDKClient(adminTS)

	checkLink := func(t *testing.T, link *admin.CreateRecoveryLinkOK, isBefore time.Time) {
		require.Contains(t, *link.Payload.RecoveryLink, publicTS.URL+recoverytoken.PublicPath)
		rl := urlx.ParseOrPanic(*link.Payload.RecoveryLink)
		assert.NotEmpty(t, rl.Query().Get("token"))
		assert.NotEmpty(t, rl.Query().Get("request"))
		require.True(t, time.Time(link.Payload.ExpiresAt).Before(isBefore))
	}

	t.Run("description=should not be able to recover an account that does not exist", func(t *testing.T) {
		_, err := adminSDK.Admin.CreateRecoveryLink(admin.NewCreateRecoveryLinkParams().WithBody(
			admin.CreateRecoveryLinkBody{IdentityID: models.UUID(x.NewUUID().String())}))
		require.IsType(t, err, new(admin.CreateRecoveryLinkNotFound), "%T", err)
	})

	t.Run("description=should not be able to recover an account that does not have a recovery email", func(t *testing.T) {
		id := identity.Identity{Traits: identity.Traits(`{}`)}
		require.NoError(t, reg.IdentityManager().Create(context.Background(),
			&id, identity.ManagerAllowWriteProtectedTraits))

		_, err := adminSDK.Admin.CreateRecoveryLink(admin.NewCreateRecoveryLinkParams().WithBody(
			admin.CreateRecoveryLinkBody{IdentityID: models.UUID(id.ID.String())}))
		require.IsType(t, err, new(admin.CreateRecoveryLinkBadRequest), "%T", err)
	})

	t.Run("description=should create a valid recovery link and set the expiry time and not e able to recover the account", func(t *testing.T) {
		id := identity.Identity{Traits: identity.Traits(`{"email":"recover.expired@ory.sh"}`)}

		require.NoError(t, reg.IdentityManager().Create(context.Background(),
			&id, identity.ManagerAllowWriteProtectedTraits))

		link, err := adminSDK.Admin.CreateRecoveryLink(admin.NewCreateRecoveryLinkParams().
			WithBody(admin.CreateRecoveryLinkBody{
				IdentityID: models.UUID(id.ID.String()),
				ExpiresIn:  "100ms",
			}))
		require.NoError(t, err)

		time.Sleep(time.Millisecond * 100)
		checkLink(t, link, time.Now().Add(conf.SelfServiceFlowRecoveryRequestLifespan()))
		res, err := publicTS.Client().Get(*link.Payload.RecoveryLink)
		require.NoError(t, err)

		require.Equal(t, http.StatusNoContent, res.StatusCode)

		// We end up here because the link is expired.
		assert.Contains(t, res.Request.URL.Path, "/recover")
	})

	t.Run("description=should create a valid recovery link and set the expiry time as well and recover the account", func(t *testing.T) {
		id := identity.Identity{Traits: identity.Traits(`{"email":"recover@ory.sh"}`)}

		require.NoError(t, reg.IdentityManager().Create(context.Background(),
			&id, identity.ManagerAllowWriteProtectedTraits))

		link, err := adminSDK.Admin.CreateRecoveryLink(admin.NewCreateRecoveryLinkParams().
			WithBody(admin.CreateRecoveryLinkBody{IdentityID: models.UUID(id.ID.String())}))
		require.NoError(t, err)

		checkLink(t, link, time.Now().Add(conf.SelfServiceFlowRecoveryRequestLifespan()+time.Second))
		res, err := publicTS.Client().Get(*link.Payload.RecoveryLink)
		require.NoError(t, err)

		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowSettingsUI().String())
		assert.Equal(t, http.StatusAccepted, res.StatusCode)
		testhelpers.LogJSON(t, link.Payload)

		sr, err := adminSDK.Common.GetSelfServiceBrowserSettingsRequest(
			common.NewGetSelfServiceBrowserSettingsRequestParams().
				WithRequest(res.Request.URL.Query().Get("request")))
		require.NoError(t, err)

		require.Len(t, sr.Payload.Messages, 1)
		assert.Equal(t, "You successfully recovered your account. Please change your password or set up an alternative login method (e.g. social sign in) within the next 60.00 minutes.", sr.Payload.Messages[0].Text)
	})
}

func TestStrategy(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	initViper()

	_ = testhelpers.NewRecoveryUITestServer(t)
	_ = testhelpers.NewLoginUIRequestEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)

	public, _ := testhelpers.NewKratosServer(t, reg)
	sdk := testhelpers.NewSDKClient(public)

	require.NoError(t, reg.IdentityManager().Create(context.Background(), identityToRecover,
		identity.ManagerAllowWriteProtectedTraits))

	var csrfField = &models.FormField{Name: pointerx.String("csrf_token"), Required: true,
		Type: pointerx.String("hidden"), Value: "nosurf"}

	t.Run("description=should set all the correct recovery payloads", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)
		rs := testhelpers.GetRecoveryRequest(t, c, public)
		assert.Contains(t, rs.Payload.Methods, recovery.StrategyRecoveryTokenName)
		method := rs.Payload.Methods[recovery.StrategyRecoveryTokenName]

		assert.EqualValues(t, models.FormFields{csrfField,
			{Name: pointerx.String("email"), Required: true, Type: pointerx.String("email")},
		}, method.Config.Fields)
		assert.EqualValues(t, public.URL+recoverytoken.PublicPath+"?request="+string(rs.Payload.ID), *method.Config.Action)
		assert.Empty(t, method.Config.Messages)
		assert.Empty(t, rs.Payload.Messages)
	})

	t.Run("description=should require an email to be sent", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)
		rs := testhelpers.GetRecoveryRequest(t, c, public)

		f := rs.Payload.Methods[recovery.StrategyRecoveryTokenName].Config

		_, rs = testhelpers.RecoverySubmitForm(t, f, c, url.Values{"email": {""}})
		assert.EqualValues(t, recovery.StrategyRecoveryTokenName, rs.Payload.Active)
		assert.Contains(t, rs.Payload.Methods, recovery.StrategyRecoveryTokenName)
		method := rs.Payload.Methods[recovery.StrategyRecoveryTokenName]
		assert.EqualValues(t, models.FormFields{csrfField,
			{Name: pointerx.String("email"), Required: true, Type: pointerx.String("email"), Value: "",
				Messages: models.Messages{{ID: 4000002, Type: "error", Text: "Property email is missing.", Context: map[string]interface{}{"property": "email"}}}},
		}, method.Config.Fields)
	})

	t.Run("description=should try to recover an email that does not exist", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)
		rs := testhelpers.GetRecoveryRequest(t, c, public)

		f := rs.Payload.Methods[recovery.StrategyRecoveryTokenName].Config

		_, rs = testhelpers.RecoverySubmitForm(t, f, c, url.Values{"email": {"i-do-not-exist@ory.sh"}})
		assert.EqualValues(t, recovery.StrategyRecoveryTokenName, rs.Payload.Active)
		assert.Contains(t, rs.Payload.Methods, recovery.StrategyRecoveryTokenName)
		method := rs.Payload.Methods[recovery.StrategyRecoveryTokenName]

		assert.EqualValues(t, models.FormFields{csrfField, {
			Name: pointerx.String("email"), Required: true, Type: pointerx.String("email"),
			Value: "i-do-not-exist@ory.sh",
		}}, method.Config.Fields)
		assertx.EqualAsJSON(t, text.Messages{*text.NewRecoveryEmailSent()}, rs.Payload.Messages)

		message := testhelpers.CourierExpectMessage(t, reg, "i-do-not-exist@ory.sh", "Account access attempted")
		assert.Contains(t, message.Body, "If this was you, check if you signed up using a different address.")
	})

	t.Run("description=should recover an account", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)
		rs := testhelpers.GetRecoveryRequest(t, c, public)

		f := rs.Payload.Methods[recovery.StrategyRecoveryTokenName].Config

		_, rs = testhelpers.RecoverySubmitForm(t, f, c, url.Values{"email": {"recover@ory.sh"}})

		assert.EqualValues(t, recovery.StrategyRecoveryTokenName, rs.Payload.Active)
		assert.Contains(t, rs.Payload.Methods, recovery.StrategyRecoveryTokenName)
		method := rs.Payload.Methods[recovery.StrategyRecoveryTokenName]
		assert.EqualValues(t, models.FormFields{csrfField, {
			Name: pointerx.String("email"), Required: true, Type: pointerx.String("email"),
			Value: "recover@ory.sh",
		}}, method.Config.Fields)
		assertx.EqualAsJSON(t, text.Messages{*text.NewRecoveryEmailSent()}, rs.Payload.Messages)

		message := testhelpers.CourierExpectMessage(t, reg, "recover@ory.sh", "Recover access to your account")
		assert.Contains(t, message.Body, "please recover access to your account by clicking the following link")

		recoveryLink := testhelpers.CourierExpectLinkInMessage(t, message, 1)

		assert.Contains(t, recoveryLink, public.URL+recoverytoken.PublicPath)
		assert.Contains(t, recoveryLink, "token=")
		res, err := c.Get(recoveryLink)
		require.NoError(t, err)

		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowSettingsUI().String())
		assert.Equal(t, http.StatusAccepted, res.StatusCode)

		sr, err := sdk.Common.GetSelfServiceBrowserSettingsRequest(
			common.NewGetSelfServiceBrowserSettingsRequestParams().WithHTTPClient(c).
				WithRequest(res.Request.URL.Query().Get("request")),
		)
		require.NoError(t, err)

		require.Len(t, sr.Payload.Messages, 1)
		assert.Equal(t, "You successfully recovered your account. Please change your password or set up an alternative login method (e.g. social sign in) within the next 60.00 minutes.", sr.Payload.Messages[0].Text)
	})

	t.Run("description=should not be able to use an invalid link", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)
		res, err := c.Get(public.URL + recoverytoken.PublicPath + "?token=i-do-not-exist")
		require.NoError(t, err)

		assert.Equal(t, http.StatusNoContent, res.StatusCode)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowRecoveryUI().String()+"?request=")

		sr, err := sdk.Common.GetSelfServiceBrowserRecoveryRequest(
			common.NewGetSelfServiceBrowserRecoveryRequestParams().WithHTTPClient(c).
				WithRequest(res.Request.URL.Query().Get("request")),
		)
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

		res, err := c.PostForm(pointerx.StringR(method.Action), url.Values{"email": {"recovery@ory.sh"}})
		require.NoError(t, err)
		assert.EqualValues(t, http.StatusNoContent, res.StatusCode)
		assert.NotContains(t, res.Request.URL.String(), "request="+rs.Payload.ID)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowRecoveryUI().String())
	})

	t.Run("description=should not be able to use an outdated request", func(t *testing.T) {
		viper.Set(configuration.ViperKeySelfServiceRecoveryRequestLifespan, time.Millisecond*200)
		t.Cleanup(func() {
			viper.Set(configuration.ViperKeySelfServiceRecoveryRequestLifespan, time.Minute)
		})

		c := testhelpers.NewClientWithCookies(t)
		rs := testhelpers.GetRecoveryRequest(t, c, public)
		_, rs = testhelpers.RecoverySubmitForm(t, rs.Payload.Methods[recovery.StrategyRecoveryTokenName].Config,
			c, url.Values{"email": {"recover@ory.sh"}})

		message := testhelpers.CourierExpectMessage(t, reg, "recover@ory.sh", "Recover access to your account")
		assert.Contains(t, message.Body, "please recover access to your account by clicking the following link")

		recoveryLink := testhelpers.CourierExpectLinkInMessage(t, message, 1)

		time.Sleep(time.Millisecond * 201)

		res, err := c.Get(recoveryLink)
		require.NoError(t, err)

		assert.EqualValues(t, http.StatusNoContent, res.StatusCode)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowRecoveryUI().String())
		assert.NotContains(t, res.Request.URL.String(), string(rs.Payload.ID))

		sr, err := sdk.Common.GetSelfServiceBrowserRecoveryRequest(
			common.NewGetSelfServiceBrowserRecoveryRequestParams().WithHTTPClient(c).
				WithRequest(res.Request.URL.Query().Get("request")),
		)
		require.NoError(t, err)

		require.Len(t, sr.Payload.Messages, 1)
		assert.Equal(t, "The recovery token is invalid or has already been used. Please retry the flow.", sr.Payload.Messages[0].Text)
	})
}
