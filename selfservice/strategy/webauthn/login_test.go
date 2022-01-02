package webauthn_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/strategy/webauthn"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/assertx"
)

//go:embed fixtures/login/success/identity.json
var loginFixtureSuccessIdentity []byte

//go:embed fixtures/login/success/response.json
var loginFixtureSuccessResponse []byte

//go:embed fixtures/login/success/internal_context.json
var loginFixtureSuccessInternalContext []byte

//go:embed fixtures/login/success/credentials.json
var loginFixtureSuccessCredentials []byte

func TestCompleteLogin(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword)+".enabled", false)
	enableWebAuthn(conf)

	router := x.NewRouterPublic()
	publicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin())

	errTS := testhelpers.NewErrorTestServer(t, reg)
	uiTS := testhelpers.NewLoginUIFlowEchoServer(t, reg)
	redirTS := testhelpers.NewRedirSessionEchoTS(t, reg)

	// Overwrite these two to make it more explicit when tests fail
	conf.MustSet(config.ViperKeySelfServiceErrorUI, errTS.URL+"/error-ts")
	conf.MustSet(config.ViperKeySelfServiceLoginUI, uiTS.URL+"/login-ts")

	conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/login.schema.json")
	conf.MustSet(config.ViperKeySecretsDefault, []string{"not-a-secure-session-key"})

	t.Run("case=webauthn payload is set when identity has webauthn", func(t *testing.T) {
		id := createIdentity(t, reg)

		apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)
		f := testhelpers.InitializeLoginFlowViaBrowser(t, apiClient, publicTS, false, true, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))
		testhelpers.SnapshotTExcept(t, f.Ui.Nodes, []string{
			"0.attributes.value",
			"1.attributes.onclick",
			"1.attributes.onload",
			"3.attributes.src",
			"3.attributes.nonce",
		})
		ensureReplacement(t, "1", f.Ui, "allowCredentials")
	})

	t.Run("case=webauthn payload is not set when identity has no webauthn", func(t *testing.T) {
		id := createIdentityWithoutWebAuthn(t, reg)
		apiClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id)
		f := testhelpers.InitializeLoginFlowViaBrowser(t, apiClient, publicTS, false, true, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))

		testhelpers.SnapshotTExcept(t, f.Ui.Nodes, []string{
			"0.attributes.value",
		})
	})

	t.Run("case=webauthn payload is not set for API clients", func(t *testing.T) {
		id := createIdentity(t, reg)

		apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)
		f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))
		assertx.EqualAsJSON(t, nil, f.Ui.Nodes)
	})

	doAPIFlow := func(t *testing.T, v func(url.Values), id *identity.Identity) (string, *http.Response) {
		apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)
		f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))
		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		v(values)
		payload := testhelpers.EncodeFormAsJSON(t, true, values)
		return testhelpers.LoginMakeRequest(t, true, false, f, apiClient, payload)
	}

	doBrowserFlow := func(t *testing.T, spa bool, v func(url.Values), id *identity.Identity) (string, *http.Response) {
		browserClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id)
		f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, spa, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))
		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		v(values)
		return testhelpers.LoginMakeRequest(t, false, spa, f, browserClient, values.Encode())
	}

	checkURL := func(t *testing.T, shouldRedirect bool, res *http.Response) {
		if shouldRedirect {
			assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/login-ts")
		} else {
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
		}
	}

	t.Run("case=should refuse to execute api flow", func(t *testing.T) {
		id := createIdentity(t, reg)
		payload := func(v url.Values) {
			v.Set(node.WebAuthnLogin, "{}")
		}

		body, res := doAPIFlow(t, payload, id)
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
		assert.Equal(t, "Could not find a strategy to log you in with. Did you fill out the form correctly?", gjson.Get(body, "ui.messages.0.text").String(), "%s", body)
	})

	t.Run("case=should fail if webauthn login is invalid", func(t *testing.T) {
		id := createIdentity(t, reg)
		payload := func(v url.Values) {
			v.Set(node.WebAuthnLogin, "{}")
		}

		check := func(t *testing.T, shouldRedirect bool, body string, res *http.Response) {
			checkURL(t, shouldRedirect, res)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Equal(t, "Unable to parse WebAuthn response.", gjson.Get(body, "ui.messages.0.text").String(), "%s", body)
		}

		t.Run("type=browser", func(t *testing.T) {
			body, res := doBrowserFlow(t, false, payload, id)
			check(t, true, body, res)
		})

		t.Run("type=spa", func(t *testing.T) {
			body, res := doBrowserFlow(t, true, payload, id)
			check(t, false, body, res)
		})
	})

	t.Run("case=login with a security key", func(t *testing.T) {
		run := func(t *testing.T, spa bool) {
			// We load our identity which we will use to replay the webauth session
			var id identity.Identity
			require.NoError(t, json.Unmarshal(loginFixtureSuccessIdentity, &id))
			id.Credentials = map[identity.CredentialsType]identity.Credentials{
				identity.CredentialsTypeWebAuthn: {
					Identifiers: []string{id.ID.String()},
					Config:      loginFixtureSuccessCredentials,
					Type:        identity.CredentialsTypeWebAuthn,
				}}
			_ = reg.PrivilegedIdentityPool().DeleteIdentity(context.Background(), id.ID)
			browserClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, &id)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, spa, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))

			// We inject the session to replay
			interim, err := reg.LoginFlowPersister().GetLoginFlow(context.Background(), uuid.FromStringOrNil(f.Id))
			require.NoError(t, err)
			interim.InternalContext = loginFixtureSuccessInternalContext
			require.NoError(t, reg.LoginFlowPersister().UpdateLoginFlow(context.Background(), interim))

			values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
			values.Set(node.WebAuthnLogin, string(loginFixtureSuccessResponse))

			// We use the response replay
			body, res := testhelpers.LoginMakeRequest(t, false, spa, f, browserClient, values.Encode())

			prefix := ""
			if spa {
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
				prefix = "session."
			} else {
				assert.Contains(t, res.Request.URL.String(), redirTS.URL)
			}

			assert.True(t, gjson.Get(body, prefix+"active").Bool(), "%s", body)
			assert.EqualValues(t, identity.AuthenticatorAssuranceLevel2, gjson.Get(body, prefix+"authenticator_assurance_level").String(), "%s", body)
			assert.EqualValues(t, identity.CredentialsTypeWebAuthn, gjson.Get(body, prefix+"authentication_methods.#(method==webauthn).method").String(), "%s", body)
			assert.EqualValues(t, id.ID.String(), gjson.Get(body, prefix+"identity.id").String(), "%s", body)

			actualFlow, err := reg.LoginFlowPersister().GetLoginFlow(context.Background(), uuid.FromStringOrNil(f.Id))
			require.NoError(t, err)
			assert.Empty(t, gjson.GetBytes(actualFlow.InternalContext, flow.PrefixInternalContextKey(identity.CredentialsTypeWebAuthn, webauthn.InternalContextKeySessionData)))
		}

		t.Run("type=browser", func(t *testing.T) {
			run(t, false)
		})

		t.Run("type=spa", func(t *testing.T) {
			run(t, true)
		})
	})
}
