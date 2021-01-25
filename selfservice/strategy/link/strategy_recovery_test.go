package link_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/ory/kratos-client-go"

	"github.com/ory/kratos/corpx"

	"github.com/ory/x/ioutilx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/urlx"

	"github.com/ory/x/sqlxx"

	"github.com/ory/x/assertx"

	"github.com/ory/x/pointerx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/strategy/link"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

func init() {
	corpx.RegisterFakes()
}

func TestAdminStrategy(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	initViper(t, conf)

	_ = testhelpers.NewRecoveryUIFlowEchoServer(t, reg)
	_ = testhelpers.NewSettingsUIFlowEchoServer(t, reg)
	_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)

	publicTS, adminTS := testhelpers.NewKratosServer(t, reg)
	adminSDK := testhelpers.NewSDKClient(adminTS)

	checkLink := func(t *testing.T, l *kratos.RecoveryLink, isBefore time.Time) {
		require.Contains(t, l.RecoveryLink, publicTS.URL+link.RouteRecovery)
		rl := urlx.ParseOrPanic(l.RecoveryLink)
		assert.NotEmpty(t, rl.Query().Get("token"))
		assert.NotEmpty(t, rl.Query().Get("flow"))
		require.True(t, (*l.ExpiresAt).Before(isBefore))
	}

	t.Run("description=should not be able to recover an account that does not exist", func(t *testing.T) {
		_, _, err := adminSDK.AdminApi.CreateRecoveryLink(context.Background()).CreateRecoveryLink(kratos.CreateRecoveryLink{
			IdentityId: x.NewUUID().String(),
		}).Execute()
		require.IsType(t, err, new(kratos.GenericOpenAPIError), "%T", err)
		assert.EqualError(t, err.(*kratos.GenericOpenAPIError), "404 Not Found")
	})

	t.Run("description=should not be able to recover an account that does not have a recovery email", func(t *testing.T) {
		id := identity.Identity{Traits: identity.Traits(`{}`)}
		require.NoError(t, reg.IdentityManager().Create(context.Background(),
			&id, identity.ManagerAllowWriteProtectedTraits))

		_, _, err := adminSDK.AdminApi.CreateRecoveryLink(context.Background()).CreateRecoveryLink(kratos.CreateRecoveryLink{
			IdentityId: id.ID.String(),
		}).Execute()
		require.IsType(t, err, new(kratos.GenericOpenAPIError), "%T", err)
		assert.EqualError(t, err.(*kratos.GenericOpenAPIError), "400 Bad Request")
	})

	t.Run("description=should create a valid recovery link and set the expiry time and not be able to recover the account", func(t *testing.T) {
		id := identity.Identity{Traits: identity.Traits(`{"email":"recover.expired@ory.sh"}`)}

		require.NoError(t, reg.IdentityManager().Create(context.Background(),
			&id, identity.ManagerAllowWriteProtectedTraits))

		rl, _, err := adminSDK.AdminApi.CreateRecoveryLink(context.Background()).CreateRecoveryLink(kratos.CreateRecoveryLink{
			IdentityId: id.ID.String(),
			ExpiresIn:  pointerx.String("100ms"),
		}).Execute()
		require.NoError(t, err)

		time.Sleep(time.Millisecond * 100)
		checkLink(t, rl, time.Now().Add(conf.SelfServiceFlowRecoveryRequestLifespan()))
		res, err := publicTS.Client().Get(rl.RecoveryLink)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, res.StatusCode)

		// We end up here because the link is expired.
		assert.Contains(t, res.Request.URL.Path, "/recover")
	})

	t.Run("description=should create a valid recovery link and set the expiry time as well and recover the account", func(t *testing.T) {
		id := identity.Identity{Traits: identity.Traits(`{"email":"recoverme@ory.sh"}`)}

		require.NoError(t, reg.IdentityManager().Create(context.Background(),
			&id, identity.ManagerAllowWriteProtectedTraits))

		rl, _, err := adminSDK.AdminApi.CreateRecoveryLink(context.Background()).CreateRecoveryLink(kratos.CreateRecoveryLink{
			IdentityId: id.ID.String(),
		}).Execute()
		require.NoError(t, err)

		checkLink(t, rl, time.Now().Add(conf.SelfServiceFlowRecoveryRequestLifespan()+time.Second))
		res, err := publicTS.Client().Get(rl.RecoveryLink)
		require.NoError(t, err)

		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowSettingsUI().String())
		assert.Equal(t, http.StatusOK, res.StatusCode)
		testhelpers.LogJSON(t, rl)

		sr, _, err := adminSDK.PublicApi.GetSelfServiceSettingsFlow(context.Background()).Id(res.Request.URL.Query().Get("flow")).Execute()
		require.NoError(t, err, "%s", res.Request.URL.String())

		require.Len(t, sr.Messages, 1)
		assert.Equal(t, "You successfully recovered your account. Please change your password or set up an alternative login method (e.g. social sign in) within the next 60.00 minutes.", sr.Messages[0].Text)
	})
}

func TestRecovery(t *testing.T) {
	var identityToRecover = &identity.Identity{
		Credentials: map[identity.CredentialsType]identity.Credentials{
			"password": {Type: "password", Identifiers: []string{"recoverme@ory.sh"}, Config: sqlxx.JSONRawMessage(`{"hashed_password":"foo"}`)}},
		Traits:   identity.Traits(`{"email":"recoverme@ory.sh"}`),
		SchemaID: config.DefaultIdentityTraitsSchemaID,
	}
	var recoveryEmail = gjson.GetBytes(identityToRecover.Traits, "email").String()

	conf, reg := internal.NewFastRegistryWithMocks(t)
	initViper(t, conf)

	_ = testhelpers.NewRecoveryUIFlowEchoServer(t, reg)
	_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
	_ = testhelpers.NewSettingsUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)

	public, _ := testhelpers.NewKratosServer(t, reg)

	require.NoError(t, reg.IdentityManager().Create(context.Background(), identityToRecover,
		identity.ManagerAllowWriteProtectedTraits))

	var csrfField = testhelpers.NewFakeCSRFNode()

	var expect = func(t *testing.T, isAPI bool, values func(url.Values), c int) string {
		hc := testhelpers.NewDebugClient(t)
		if !isAPI {
			hc = testhelpers.NewDebugClient(t)
		}
		return testhelpers.SubmitRecoveryForm(t, isAPI, hc, public, values, recovery.StrategyRecoveryLinkName, c,
			testhelpers.ExpectURL(isAPI, public.URL+link.RouteRecovery, conf.SelfServiceFlowRecoveryUI().String()))
	}

	var expectValidationError = func(t *testing.T, isAPI bool, values func(url.Values)) string {
		return expect(t, isAPI, values, testhelpers.ExpectStatusCode(isAPI, http.StatusBadRequest, http.StatusOK))
	}

	var expectSuccess = func(t *testing.T, isAPI bool, values func(url.Values)) string {
		return expect(t, isAPI, values, http.StatusOK)
	}

	t.Run("description=should set all the correct recovery payloads", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)
		rs := testhelpers.GetRecoveryFlow(t, c, public)
		assert.Contains(t, rs.Methods, recovery.StrategyRecoveryLinkName)
		method := rs.Methods[recovery.StrategyRecoveryLinkName]

		assert.EqualValues(t, []kratos.UiNode{
			*csrfField,
			{
				Type:  "input",
				Group: "recovery_link",
				Attributes: kratos.UiNodeInputAttributesAsUiNodeAttributes(&kratos.UiNodeInputAttributes{
					Name:     "email",
					Required: pointerx.Bool(true),
					Type:     "email",
				}),
			},
		}, method.Config.Nodes)
		assert.EqualValues(t, public.URL+link.RouteRecovery+"?flow="+rs.Id, method.Config.Action)
		assert.Empty(t, method.Config.Messages)
		assert.Empty(t, rs.Messages)
	})

	t.Run("description=should require an email to be sent", func(t *testing.T) {
		var check = func(t *testing.T, actual string) {
			assert.EqualValues(t, recovery.StrategyRecoveryLinkName, gjson.Get(actual, "active").String(), "%s", actual)
			assert.EqualValues(t, "Property email is missing.",
				gjson.Get(actual, "methods.link.config.nodes.#(attributes.name==email).messages.0.text").String(),
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
			assert.EqualValues(t, recovery.StrategyRecoveryLinkName, gjson.Get(actual, "active").String(), "%s", actual)
			assert.EqualValues(t, email, gjson.Get(actual, "methods.link.config.nodes.#(attributes.name==email).attributes.value").String(), "%s", actual)
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
			assert.EqualValues(t, recovery.StrategyRecoveryLinkName, gjson.Get(actual, "active").String(), "%s", actual)
			assert.EqualValues(t, recoveryEmail, gjson.Get(actual, "methods.link.config.nodes.#(attributes.name==email).attributes.value").String(), "%s", actual)
			assertx.EqualAsJSON(t, text.NewRecoveryEmailSent(), json.RawMessage(gjson.Get(actual, "messages.0").Raw))

			message := testhelpers.CourierExpectMessage(t, reg, recoveryEmail, "Recover access to your account")
			assert.Contains(t, message.Body, "please recover access to your account by clicking the following link")

			recoveryLink := testhelpers.CourierExpectLinkInMessage(t, message, 1)

			assert.Contains(t, recoveryLink, public.URL+link.RouteRecovery)
			assert.Contains(t, recoveryLink, "token=")

			cl := testhelpers.NewClientWithCookies(t)
			res, err := cl.Get(recoveryLink)
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowSettingsUI().String())

			body := ioutilx.MustReadAll(res.Body)
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
		res, err := c.Get(public.URL + link.RouteRecovery + "?token=i-do-not-exist")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowRecoveryUI().String()+"?flow=")

		sr, _, err := testhelpers.NewSDKCustomClient(public, c).PublicApi.GetSelfServiceRecoveryFlow(context.Background()).Id(res.Request.URL.Query().Get("flow")).Execute()
		require.NoError(t, err)

		require.Len(t, sr.Messages, 1)
		assert.Equal(t, "The recovery token is invalid or has already been used. Please retry the flow.", sr.Messages[0].Text)
	})

	t.Run("description=should not be able to use an outdated link", func(t *testing.T) {
		conf.MustSet(config.ViperKeySelfServiceRecoveryRequestLifespan, time.Millisecond*200)
		t.Cleanup(func() {
			conf.MustSet(config.ViperKeySelfServiceRecoveryRequestLifespan, time.Minute)
		})

		c := testhelpers.NewClientWithCookies(t)
		rs := testhelpers.GetRecoveryFlow(t, c, public)
		method := rs.Methods[recovery.StrategyRecoveryLinkName].Config

		time.Sleep(time.Millisecond * 201)

		res, err := c.PostForm(method.Action, url.Values{"email": {recoveryEmail}})
		require.NoError(t, err)
		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.NotContains(t, res.Request.URL.String(), "flow="+rs.Id)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowRecoveryUI().String())
	})

	t.Run("description=should not be able to use an outdated flow", func(t *testing.T) {
		conf.MustSet(config.ViperKeySelfServiceRecoveryRequestLifespan, time.Millisecond*200)
		t.Cleanup(func() {
			conf.MustSet(config.ViperKeySelfServiceRecoveryRequestLifespan, time.Minute)
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

		sr, _, err := testhelpers.NewSDKCustomClient(public, c).PublicApi.GetSelfServiceRecoveryFlow(context.Background()).Id(res.Request.URL.Query().Get("flow")).Execute()
		require.NoError(t, err)

		require.Len(t, sr.Messages, 1)
		assert.Contains(t, sr.Messages[0].Text, "The recovery flow expired")
	})
}

func TestDisabledEndpoint(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	initViper(t, conf)
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+recovery.StrategyRecoveryLinkName+".enabled", false)

	publicTS, adminTS := testhelpers.NewKratosServer(t, reg)
	adminSDK := testhelpers.NewSDKClient(adminTS)

	t.Run("role=admin", func(t *testing.T) {
		t.Run("description=can not create recovery link when link method is disabled", func(t *testing.T) {
			id := identity.Identity{Traits: identity.Traits(`{"email":"recovery-endpoint-disabled@ory.sh"}`)}

			require.NoError(t, reg.IdentityManager().Create(context.Background(),
				&id, identity.ManagerAllowWriteProtectedTraits))

			uuid := models.UUID(id.ID.String())
			rl, err := adminSDK.Admin.CreateRecoveryLink(admin.NewCreateRecoveryLinkParams().
				WithBody(&models.CreateRecoveryLink{IdentityID: &uuid}))
			assert.Nil(t, rl)
			require.IsType(t, new(admin.CreateRecoveryLinkNotFound), err, "%s", err)

			br, _ := err.(*admin.CreateRecoveryLinkNotFound)
			assert.Contains(t, br.Payload.Error.Reason, "This endpoint was disabled by system administrator")
		})
	})

	t.Run("role=public", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)
		t.Run("description=can not recover an account by get request when link method is disabled", func(t *testing.T) {
			u := publicTS.URL + link.RouteRecovery + "?token=endpoint-disabled"
			res, err := c.Get(u)
			require.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, res.StatusCode)

			b := make([]byte, res.ContentLength)
			_, _ = res.Body.Read(b)
			assert.Contains(t, string(b), "This endpoint was disabled by system administrator")
		})

		t.Run("description=can not recover an account by post request when link method is disabled", func(t *testing.T) {
			u := publicTS.URL + link.RouteRecovery
			res, err := c.PostForm(u, url.Values{"email": {"email@ory.sh"}})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, res.StatusCode)

			b := make([]byte, res.ContentLength)
			_, _ = res.Body.Read(b)
			assert.Contains(t, string(b), "This endpoint was disabled by system administrator")
		})
	})
}
