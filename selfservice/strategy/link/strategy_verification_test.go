package link_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/ory/kratos/ui/node"

	"github.com/ory/x/pointerx"

	"github.com/ory/kratos-client-go"

	"github.com/ory/x/ioutilx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/assertx"
	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

func TestVerification(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	initViper(t, conf)

	var identityToVerify = &identity.Identity{
		ID:       x.NewUUID(),
		Traits:   identity.Traits(`{"email":"verifyme@ory.sh"}`),
		SchemaID: config.DefaultIdentityTraitsSchemaID,
		Credentials: map[identity.CredentialsType]identity.Credentials{
			"password": {Type: "password", Identifiers: []string{"recoverme@ory.sh"}, Config: sqlxx.JSONRawMessage(`{"hashed_password":"foo"}`)}},
	}

	var verificationEmail = gjson.GetBytes(identityToVerify.Traits, "email").String()

	_ = testhelpers.NewVerificationUIFlowEchoServer(t, reg)
	_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
	_ = testhelpers.NewSettingsUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)
	_ = testhelpers.NewRedirTS(t, "returned", conf)

	public, _ := testhelpers.NewKratosServer(t, reg)
	require.NoError(t, reg.IdentityManager().Create(context.Background(), identityToVerify,
		identity.ManagerAllowWriteProtectedTraits))

	var csrfField = testhelpers.NewFakeCSRFNode()

	var expect = func(t *testing.T, isAPI bool, values func(url.Values), c int) string {
		hc := testhelpers.NewDebugClient(t)
		if !isAPI {
			hc = testhelpers.NewDebugClient(t)
		}
		return testhelpers.SubmitVerificationForm(t, isAPI, hc, public, values, c,
			testhelpers.ExpectURL(isAPI, public.URL+verification.RouteSubmitFlow, conf.SelfServiceFlowVerificationUI().String()))
	}

	var expectValidationError = func(t *testing.T, isAPI bool, values func(url.Values)) string {
		return expect(t, isAPI, values, testhelpers.ExpectStatusCode(isAPI, http.StatusBadRequest, http.StatusOK))
	}

	var expectSuccess = func(t *testing.T, isAPI bool, values func(url.Values)) string {
		return expect(t, isAPI, values, http.StatusOK)
	}

	t.Run("description=should set all the correct verification payloads", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)
		rs := testhelpers.GetVerificationFlow(t, c, public)

		assert.EqualValues(t, []kratos.UiNode{
			*csrfField,
			{
				Type:  "input",
				Group: node.VerificationLinkGroup.String(),
				Attributes: kratos.UiNodeInputAttributesAsUiNodeAttributes(&kratos.UiNodeInputAttributes{
					Name:     "email",
					Type:     "email",
					Required: pointerx.Bool(true),
				}),
			},
			*testhelpers.NewMethodSubmit(string(node.VerificationLinkGroup), verification.StrategyVerificationLinkName),
		}, rs.Ui.Nodes)
		assert.EqualValues(t, public.URL+verification.RouteSubmitFlow+"?flow="+rs.Id, rs.Ui.Action)
		assert.Empty(t, rs.Ui.Messages)
	})

	t.Run("description=should not execute submit without correct method set", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)
		rs := testhelpers.GetVerificationFlow(t, c, public)

		res, err := c.PostForm(rs.Ui.Action, url.Values{"method": {"not-link"}, "email": {verificationEmail}})
		require.NoError(t, err)
		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowVerificationUI().String())

		body := ioutilx.MustReadAll(res.Body)
		require.NoError(t, res.Body.Close())
		assert.Equal(t, "Could not find a strategy to verify your account with. Did you fill out the form correctly?", gjson.GetBytes(body, "ui.messages.0.text").String(), "%s", body)
	})

	t.Run("description=should require an email to be sent", func(t *testing.T) {
		var check = func(t *testing.T, actual string) {
			assert.EqualValues(t, string(node.VerificationLinkGroup), gjson.Get(actual, "active").String(), "%s", actual)
			assert.EqualValues(t, "Property email is missing.",
				gjson.Get(actual, "ui.nodes.#(attributes.name==email).messages.0.text").String(),
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

	t.Run("description=should try to verify an email that does not exist", func(t *testing.T) {
		var email string
		var check = func(t *testing.T, actual string) {
			assert.EqualValues(t, string(node.VerificationLinkGroup), gjson.Get(actual, "active").String(), "%s", actual)
			assert.EqualValues(t, email, gjson.Get(actual, "ui.nodes.#(attributes.name==email).attributes.value").String(), "%s", actual)
			assertx.EqualAsJSON(t, text.NewVerificationEmailSent(), json.RawMessage(gjson.Get(actual, "ui.messages.0").Raw))

			message := testhelpers.CourierExpectMessage(t, reg, email, "Someone tried to verify this email address")
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

	t.Run("description=should not be able to use an invalid link", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)
		f := testhelpers.InitializeVerificationFlowViaBrowser(t, c, public)
		res, err := c.Get(public.URL + verification.RouteSubmitFlow + "?flow=" + f.Id + "&token=i-do-not-exist")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowVerificationUI().String()+"?flow=")

		sr, _, err := testhelpers.NewSDKCustomClient(public, c).PublicApi.GetSelfServiceVerificationFlow(context.Background()).Id(res.Request.URL.Query().Get("flow")).Execute()
		require.NoError(t, err)

		require.Len(t, sr.Ui.Messages, 1)
		assert.Equal(t, "The verification token is invalid or has already been used. Please retry the flow.", sr.Ui.Messages[0].Text)
	})

	t.Run("description=should not be able to use an outdated link", func(t *testing.T) {
		conf.MustSet(config.ViperKeySelfServiceVerificationRequestLifespan, time.Millisecond*200)
		t.Cleanup(func() {
			conf.MustSet(config.ViperKeySelfServiceVerificationRequestLifespan, time.Minute)
		})

		c := testhelpers.NewClientWithCookies(t)
		rs := testhelpers.GetVerificationFlow(t, c, public)

		time.Sleep(time.Millisecond * 201)

		res, err := c.PostForm(rs.Ui.Action, url.Values{"method": {"link"}, "email": {verificationEmail}})
		require.NoError(t, err)
		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.NotContains(t, res.Request.URL.String(), "flow="+rs.Id)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowVerificationUI().String())
	})

	t.Run("description=should not be able to use an outdated flow", func(t *testing.T) {
		conf.MustSet(config.ViperKeySelfServiceVerificationRequestLifespan, time.Millisecond*200)
		t.Cleanup(func() {
			conf.MustSet(config.ViperKeySelfServiceVerificationRequestLifespan, time.Minute)
		})

		body := expectSuccess(t, false, func(v url.Values) {
			v.Set("email", verificationEmail)
		})

		message := testhelpers.CourierExpectMessage(t, reg, verificationEmail, "Please verify your email address")
		assert.Contains(t, message.Body, "Hi, please verify your account by clicking the following link")

		verificationLink := testhelpers.CourierExpectLinkInMessage(t, message, 1)

		time.Sleep(time.Millisecond * 201)

		c := testhelpers.NewClientWithCookies(t)
		res, err := c.Get(verificationLink)
		require.NoError(t, err)

		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowVerificationUI().String())
		assert.NotContains(t, res.Request.URL.String(), gjson.Get(body, "id").String())

		sr, _, err := testhelpers.NewSDKCustomClient(public, c).PublicApi.GetSelfServiceVerificationFlow(context.Background()).Id(res.Request.URL.Query().Get("flow")).Execute()
		require.NoError(t, err)

		require.Len(t, sr.Ui.Messages, 1)
		assert.Contains(t, sr.Ui.Messages[0].Text, "The verification flow expired")
	})

	t.Run("description=should verify an email address", func(t *testing.T) {
		var check = func(t *testing.T, actual string) {
			assert.EqualValues(t, string(node.VerificationLinkGroup), gjson.Get(actual, "active").String(), "%s", actual)
			assert.EqualValues(t, verificationEmail, gjson.Get(actual, "ui.nodes.#(attributes.name==email).attributes.value").String(), "%s", actual)
			assertx.EqualAsJSON(t, text.NewVerificationEmailSent(), json.RawMessage(gjson.Get(actual, "ui.messages.0").Raw))

			message := testhelpers.CourierExpectMessage(t, reg, verificationEmail, "Please verify your email address")
			assert.Contains(t, message.Body, "please verify your account by clicking the following link")

			verificationLink := testhelpers.CourierExpectLinkInMessage(t, message, 1)

			assert.Contains(t, verificationLink, public.URL+verification.RouteSubmitFlow)
			assert.Contains(t, verificationLink, "token=")

			cl := testhelpers.NewClientWithCookies(t)
			res, err := cl.Get(verificationLink)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowVerificationUI().String())
			body := string(ioutilx.MustReadAll(res.Body))
			assert.EqualValues(t, "passed_challenge", gjson.Get(body, "state").String())

			id, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), identityToVerify.ID)
			require.NoError(t, err)
			require.Len(t, id.VerifiableAddresses, 1)

			address := id.VerifiableAddresses[0]
			assert.EqualValues(t, verificationEmail, address.Value)
			assert.True(t, address.Verified)
			assert.EqualValues(t, identity.VerifiableAddressStatusCompleted, address.Status)
			assert.True(t, time.Time(address.VerifiedAt).Add(time.Second*5).After(time.Now()))
		}

		var values = func(v url.Values) {
			v.Set("email", verificationEmail)
		}

		t.Run("type=browser", func(t *testing.T) {
			check(t, expectSuccess(t, false, values))
		})

		t.Run("type=api", func(t *testing.T) {
			check(t, expectSuccess(t, true, values))
		})
	})
}
