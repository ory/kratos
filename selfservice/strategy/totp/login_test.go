package totp_test

import (
	"context"
	"fmt"
	"github.com/ory/kratos/text"
	"testing"
	"time"

	"github.com/pquerna/otp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/strategy/totp"
	"github.com/ory/kratos/x"
	"github.com/ory/x/sqlxx"
)

func createIdentity(t *testing.T, reg driver.Registry) (*identity.Identity, *otp.Key) {
	identifier := x.NewUUID().String() + "@ory.sh"
	password := x.NewUUID().String()
	key, err := totp.NewKey(context.Background(), "foo", reg)
	require.NoError(t, err)
	p, err := reg.Hasher().Generate(context.Background(), []byte(password))
	require.NoError(t, err)
	i := &identity.Identity{
		Traits: identity.Traits(fmt.Sprintf(`{"subject":"%s"}`, identifier)),
		VerifiableAddresses: []identity.VerifiableAddress{
			{
				Value:     identifier,
				Verified:  false,
				CreatedAt: time.Now(),
			},
		},
	}
	require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))
	i.Credentials = map[identity.CredentialsType]identity.Credentials{
		identity.CredentialsTypePassword: {
			Type:        identity.CredentialsTypePassword,
			Identifiers: []string{identifier},
			Config:      sqlxx.JSONRawMessage(`{"hashed_password":"` + string(p) + `"}`),
		},
		identity.CredentialsTypeTOTP: {
			Type:        identity.CredentialsTypeTOTP,
			Identifiers: []string{i.ID.String()},
			Config:      sqlxx.JSONRawMessage(`{"totp_url":"` + string(key.URL()) + `"}`),
		},
	}
	require.NoError(t, reg.PrivilegedIdentityPool().UpdateIdentity(context.Background(), i))
	return i, key
}

func TestCompleteLogin(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword), map[string]interface{}{"enabled": true})
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeTOTP), map[string]interface{}{"enabled": true})

	router := x.NewRouterPublic()
	publicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin())

	errTS := testhelpers.NewErrorTestServer(t, reg)
	uiTS := testhelpers.NewLoginUIFlowEchoServer(t, reg)
	//redirTS := testhelpers.NewRedirSessionEchoTS(t, reg)

	// Overwrite these two to make it more explicit when tests fail
	conf.MustSet(config.ViperKeySelfServiceErrorUI, errTS.URL+"/error-ts")
	conf.MustSet(config.ViperKeySelfServiceLoginUI, uiTS.URL+"/login-ts")

	conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/login.schema.json")
	conf.MustSet(config.ViperKeySecretsDefault, []string{"not-a-secure-session-key"})

	t.Run("case=should show the error ui because the request payload is malformed", func(t *testing.T) {
		id, _ := createIdentity(t, reg)
		t.Run("type=api", func(t *testing.T) {
			apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)
			f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))

			body, res := testhelpers.LoginMakeRequest(t, true, false, f, apiClient, "14=)=!(%)$/ZP()GHIÖ")
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, body, `Expected JSON sent in request body to be an object but got: Number`)
		})

		t.Run("type=browser", func(t *testing.T) {
			browserClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, false, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))

			body, res := testhelpers.LoginMakeRequest(t, false, false, f, browserClient, "14=)=!(%)$/ZP()GHIÖ")
			assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/login-ts")
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "invalid URL escape", "%s", body)
		})

		t.Run("type=spa", func(t *testing.T) {
			browserClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, true, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))

			body, res := testhelpers.LoginMakeRequest(t, false, true, f, browserClient, "14=)=!(%)$/ZP()GHIÖ")
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "invalid URL escape", "%s", body)
		})
	})

	t.Run("case=should fail if code is empty", func(t *testing.T) {
		id, _ := createIdentity(t, reg)
		t.Run("type=api", func(t *testing.T) {
			apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)
			f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))

			body, res := testhelpers.LoginMakeRequest(t, true, false, f, apiClient, `{"method":"totp","totp_code":""}`)
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Equal(t, "length must be >= 6, but got 0", gjson.Get(body, "ui.nodes.#(attributes.name==totp_code).messages.0.text").String(), "%s", body)
		})

		t.Run("type=browser", func(t *testing.T) {
			browserClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, false, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))

			vals := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
			vals.Set("totp_code", "")

			body, res := testhelpers.LoginMakeRequest(t, false, false, f, browserClient, vals.Encode())
			assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/login-ts")
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Equal(t, "length must be >= 6, but got 0", gjson.Get(body, "ui.nodes.#(attributes.name==totp_code).messages.0.text").String(), "%s", body)
		})

		t.Run("type=spa", func(t *testing.T) {
			browserClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, true, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))

			vals := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
			vals.Set("totp_code", "")

			body, res := testhelpers.LoginMakeRequest(t, false, true, f, browserClient, vals.Encode())
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Equal(t, "length must be >= 6, but got 0", gjson.Get(body, "ui.nodes.#(attributes.name==totp_code).messages.0.text").String(), "%s", body)
		})
	})

	t.Run("case=should fail if code is empty", func(t *testing.T) {
		id, _ := createIdentity(t, reg)
		t.Run("type=api", func(t *testing.T) {
			apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)
			f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))

			body, res := testhelpers.LoginMakeRequest(t, true, false, f, apiClient, `{"method":"totp","totp_code":"111111"}`)
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Equal(t, text.NewErrorValidationInvalidTOTPCode().Text, gjson.Get(body, "ui.messages.0.text").String(), "%s", body)
		})

		t.Run("type=browser", func(t *testing.T) {
			browserClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, false, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))

			vals := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
			vals.Set("totp_code", "111111")

			body, res := testhelpers.LoginMakeRequest(t, false, false, f, browserClient, vals.Encode())
			assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/login-ts")
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Equal(t, text.NewErrorValidationInvalidTOTPCode().Text, gjson.Get(body, "ui.messages.0.text").String(), "%s", body)
		})

		t.Run("type=spa", func(t *testing.T) {
			browserClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, true, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))

			vals := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
			vals.Set("totp_code", "111111")

			body, res := testhelpers.LoginMakeRequest(t, false, true, f, browserClient, vals.Encode())
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Equal(t, text.NewErrorValidationInvalidTOTPCode().Text, gjson.Get(body, "ui.messages.0.text").String(), "%s", body)
		})
	})

	// check what happens if identity has no totp set up

	// check what happens if good code is sent

	// One test where we check AAL1 with TOTP (need to fake the flow)
}
