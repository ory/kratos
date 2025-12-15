// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package passkey_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/sjson"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	kratos "github.com/ory/kratos/internal/httpclient"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/assertx"
	"github.com/ory/x/sqlxx"
)

func newRegistrationRegistry(t *testing.T) *driver.RegistryDefault {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword)+".enabled", true)
	conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationEnableLegacyOneStep, true)
	enablePasskeyStrategy(conf)
	conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationLoginHints, true)
	return reg
}

func newLoginRegistry(t *testing.T) *driver.RegistryDefault {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword)+".enabled", true)
	enablePasskeyStrategy(conf)
	conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationLoginHints, true)
	return reg
}

func enablePasskeyStrategy(conf *config.Config) {
	ctx := context.Background()
	key := config.ViperKeySelfServiceStrategyConfig + "." + string(identity.CredentialsTypePasskey)
	conf.MustSet(ctx, key+".enabled", true)
	conf.MustSet(ctx, key+".config.rp.display_name", "Ory Corp")
	conf.MustSet(ctx, key+".config.rp.id", "localhost")
	conf.MustSet(ctx, key+".config.rp.origins", []string{"http://localhost:4455"})
}

type fixture struct {
	ctx  context.Context
	conf *config.Config
	reg  *driver.RegistryDefault

	publicTS         *httptest.Server
	redirTS          *httptest.Server
	redirNoSessionTS *httptest.Server
	uiTS             *httptest.Server
	errTS            *httptest.Server
	loginTS          *httptest.Server
}

func newRegistrationFixture(t *testing.T) *fixture {
	fix := new(fixture)
	fix.ctx = context.Background()
	fix.reg = newRegistrationRegistry(t)
	fix.conf = fix.reg.Config()
	ctx := fix.ctx

	router := x.NewRouterPublic(fix.reg)
	fix.publicTS, _ = testhelpers.NewKratosServerWithRouters(t, fix.reg, router, x.NewRouterAdmin(fix.reg))

	_ = testhelpers.NewErrorTestServer(t, fix.reg)
	_ = testhelpers.NewRegistrationUIFlowEchoServer(t, fix.reg)
	_ = testhelpers.NewRedirSessionEchoTS(t, fix.reg)

	testhelpers.SetDefaultIdentitySchema(fix.conf, "file://./stub/registration.schema.json")
	fix.conf.MustSet(ctx, config.ViperKeySecretsDefault, []string{"not-a-secure-session-key"})

	fix.redirTS = testhelpers.NewRedirSessionEchoTS(t, fix.reg)
	fix.redirNoSessionTS = testhelpers.NewRedirNoSessionTS(t, fix.reg)

	fix.useReturnToFromTS(fix.redirTS)

	return fix
}

func newLoginFixture(t *testing.T) *fixture {
	fix := new(fixture)
	fix.ctx = context.Background()
	fix.reg = newLoginRegistry(t)
	fix.conf = fix.reg.Config()
	ctx := fix.ctx

	fix.conf.MustSet(ctx,
		config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword)+".enabled",
		false)

	router := x.NewRouterPublic(fix.reg)
	fix.publicTS, _ = testhelpers.NewKratosServerWithRouters(t, fix.reg, router, x.NewRouterAdmin(fix.reg))

	fix.errTS = testhelpers.NewErrorTestServer(t, fix.reg)
	fix.uiTS = testhelpers.NewLoginUIFlowEchoServer(t, fix.reg)
	fix.loginTS = fix.uiTS

	// Overwrite these two to make it more explicit when tests fail
	fix.conf.MustSet(ctx, config.ViperKeySelfServiceErrorUI, fix.errTS.URL+"/error-ts")
	fix.conf.MustSet(ctx, config.ViperKeySelfServiceLoginUI, fix.uiTS.URL+"/login-ts")

	testhelpers.SetDefaultIdentitySchema(fix.conf, "file://./stub/login.schema.json")
	fix.conf.MustSet(ctx, config.ViperKeySecretsDefault, []string{"not-a-secure-session-key"})

	fix.redirTS = testhelpers.NewRedirSessionEchoTS(t, fix.reg)
	fix.redirNoSessionTS = testhelpers.NewRedirNoSessionTS(t, fix.reg)

	fix.useReturnToFromTS(fix.redirTS)

	return fix
}

func newSettingsFixture(t *testing.T) *fixture {
	fix := newLoginFixture(t)
	fix.uiTS = testhelpers.NewSettingsUIFlowEchoServer(t, fix.reg)
	fix.conf.MustSet(ctx, config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1m")
	testhelpers.SetDefaultIdentitySchema(fix.conf, "file://./stub/settings.schema.json")
	fix.conf.MustSet(ctx, config.ViperKeySecretsDefault, []string{"not-a-secure-session-key"})
	fix.conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, "aal1")
	fix.conf.MustSet(fix.ctx, config.ViperKeySessionWhoAmIAAL, "aal1")
	fix.conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".profile.enabled", false)

	return fix
}

func (fix *fixture) checkURL(t *testing.T, shouldRedirect bool, res *http.Response) {
	if shouldRedirect {
		assert.Contains(t, res.Request.URL.String(), fix.uiTS.URL+"/login-ts")
	} else {
		assert.Contains(t, res.Request.URL.String(), fix.publicTS.URL+login.RouteSubmitFlow)
	}
}

func (fix *fixture) loginViaBrowser(t *testing.T, spa bool, cb func(url.Values), browserClient *http.Client, opts ...testhelpers.InitFlowWithOption) (string, *http.Response) {
	f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, fix.publicTS, false, spa, false, false, opts...)
	values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
	cb(values)
	return testhelpers.LoginMakeRequest(t, false, spa, f, browserClient, values.Encode())
}

func (fix *fixture) createIdentityWithPasskey(t *testing.T, c identity.Credentials) *identity.Identity {
	var id identity.Identity
	require.NoError(t, json.Unmarshal(loginSuccessIdentity, &id))

	id.SetCredentials(identity.CredentialsTypePasskey, identity.Credentials{
		Identifiers: []string{"some-random-user-handle"},
		Config:      c.Config,
		Type:        identity.CredentialsTypePasskey,
		Version:     c.Version,
	})

	// We clean up the identity in case it has been created before
	_ = fix.reg.PrivilegedIdentityPool().DeleteIdentity(fix.ctx, id.ID)

	require.NoError(t, fix.reg.PrivilegedIdentityPool().CreateIdentity(fix.ctx, &id))

	return &id
}

func (fix *fixture) submitWebAuthnLoginFlowWithClient(t *testing.T, isSPA bool, f *kratos.LoginFlow, contextFixture []byte, client *http.Client, cb func(values url.Values)) (string, *http.Response, *kratos.LoginFlow) {
	// We inject the session to replay
	interim, err := fix.reg.LoginFlowPersister().GetLoginFlow(fix.ctx, uuid.FromStringOrNil(f.Id))
	require.NoError(t, err)
	interim.InternalContext = contextFixture
	require.NoError(t, fix.reg.LoginFlowPersister().UpdateLoginFlow(fix.ctx, interim))

	values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
	cb(values)

	// We use the response replay
	body, res := testhelpers.LoginMakeRequest(t, false, isSPA, f, client, values.Encode())
	return body, res, f
}

func (fix *fixture) submitWebAuthnLoginWithClient(t *testing.T, isSPA bool, contextFixture []byte, client *http.Client, cb func(values url.Values), opts ...testhelpers.InitFlowWithOption) (string, *http.Response, *kratos.LoginFlow) {
	f := testhelpers.InitializeLoginFlowViaBrowser(t, client, fix.publicTS, false, isSPA, false, false, opts...)
	return fix.submitWebAuthnLoginFlowWithClient(t, isSPA, f, contextFixture, client, cb)
}

func (fix *fixture) submitWebAuthnLogin(t *testing.T, ctx context.Context, isSPA bool, id *identity.Identity, contextFixture []byte, cb func(values url.Values), opts ...testhelpers.InitFlowWithOption) (string, *http.Response, *kratos.LoginFlow) {
	browserClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, ctx, fix.reg, id)
	return fix.submitWebAuthnLoginWithClient(t, isSPA, contextFixture, browserClient, cb, opts...)
}

// useReturnToFromTS sets the "return to" server, which will assert the session
// state (redirTS: enforce that a session exists, redirNoSessionTS: enforce that
// no session exists)
func (fix *fixture) useReturnToFromTS(ts *httptest.Server) {
	fix.conf.MustSet(fix.ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, ts.URL+"/default-return-to")
	fix.conf.MustSet(fix.ctx, config.ViperKeySelfServiceRegistrationAfter+"."+config.DefaultBrowserReturnURL, ts.URL+"/registration-return-ts")
}
func (fix *fixture) useRedirTS()          { fix.useReturnToFromTS(fix.redirTS) }
func (fix *fixture) useRedirNoSessionTS() { fix.useReturnToFromTS(fix.redirNoSessionTS) }

func (fix *fixture) disableSessionAfterRegistration() {
	fix.conf.MustSet(fix.ctx, config.HookStrategyKey(
		config.ViperKeySelfServiceRegistrationAfter,
		identity.CredentialsTypePasskey.String(),
	), nil)
}

func (fix *fixture) enableSessionAfterRegistration() {
	fix.conf.MustSet(fix.ctx, config.HookStrategyKey(
		config.ViperKeySelfServiceRegistrationAfter,
		identity.CredentialsTypePasskey.String(),
	), []config.SelfServiceHook{{Name: "session"}})
}

type submitPasskeyOpt struct {
	initFlowOpts    []testhelpers.InitFlowWithOption
	userID          string
	internalContext sqlxx.JSONRawMessage
}

type submitPasskeyOption func(o *submitPasskeyOpt)

func withUserID(id string) submitPasskeyOption {
	return func(o *submitPasskeyOpt) {
		o.userID = base64.StdEncoding.EncodeToString([]byte(id))
	}
}

func withInternalContext(ic sqlxx.JSONRawMessage) submitPasskeyOption {
	return func(o *submitPasskeyOpt) {
		o.internalContext = ic
	}
}

func withInitFlowWithOption(ifo []testhelpers.InitFlowWithOption) submitPasskeyOption {
	return func(o *submitPasskeyOpt) {
		o.initFlowOpts = ifo
	}
}

func (fix *fixture) submitPasskeyBrowserRegistration(
	t *testing.T,
	flowType string,
	client *http.Client,
	cb func(values url.Values),
	opts ...submitPasskeyOption,
) (string, *http.Response, *kratos.RegistrationFlow) {
	return fix.submitPasskeyRegistration(t, flowType, client, cb, append([]submitPasskeyOption{withInternalContext(registrationFixtureSuccessBrowserInternalContext)}, opts...)...)
}

func (fix *fixture) submitPasskeyAndroidRegistration(
	t *testing.T,
	flowType string,
	client *http.Client,
	cb func(values url.Values),
	opts ...submitPasskeyOption,
) (string, *http.Response, *kratos.RegistrationFlow) {
	return fix.submitPasskeyRegistration(t, flowType, client, cb,
		append([]submitPasskeyOption{withInternalContext(
			registrationFixtureSuccessAndroidInternalContext,
		)}, opts...)...)
}

func (fix *fixture) submitPasskeyRegistration(
	t *testing.T,
	flowType string,
	client *http.Client,
	cb func(values url.Values),
	opts ...submitPasskeyOption,
) (string, *http.Response, *kratos.RegistrationFlow) {
	o := &submitPasskeyOpt{}
	for _, fn := range opts {
		fn(o)
	}

	isSPA := flowType == "spa"
	regFlow := testhelpers.InitializeRegistrationFlowViaBrowser(t, client, fix.publicTS, isSPA, false, false, o.initFlowOpts...)

	// First step: fill out traits and click on "sign up with passkey"
	values := testhelpers.SDKFormFieldsToURLValues(regFlow.Ui.Nodes)
	cb(values)
	passkeyRegisterVal := values.Get(node.PasskeyRegister) // needed in the second step
	values.Del(node.PasskeyRegister)
	values.Set("method", "passkey")
	_, _ = testhelpers.RegistrationMakeRequest(t, false, isSPA, regFlow, client, values.Encode())

	// We inject the session to replay
	interim, err := fix.reg.RegistrationFlowPersister().GetRegistrationFlow(fix.ctx, uuid.FromStringOrNil(regFlow.Id))
	require.NoError(t, err)
	interim.InternalContext = o.internalContext
	if o.userID != "" {
		interim.InternalContext, err = sjson.SetBytes(interim.InternalContext, "passkey_session_data.user_id", o.userID)
		require.NoError(t, err)
	}
	require.NoError(t, fix.reg.RegistrationFlowPersister().UpdateRegistrationFlow(fix.ctx, interim))

	// Second step: fill out passkey response
	values.Set(node.PasskeyRegister, passkeyRegisterVal)
	body, res := testhelpers.RegistrationMakeRequest(t, false, isSPA, regFlow, client, values.Encode())

	return body, res, regFlow
}

func (fix *fixture) makeRegistration(t *testing.T, flowType string, values func(v url.Values), opts ...submitPasskeyOption) (actual string, res *http.Response, fetchedFlow *registration.Flow) {
	actual, res, actualFlow := fix.submitPasskeyBrowserRegistration(t, flowType, testhelpers.NewClientWithCookies(t), values, opts...)
	fetchedFlow, err := fix.reg.RegistrationFlowPersister().GetRegistrationFlow(fix.ctx, uuid.FromStringOrNil(actualFlow.Id))
	require.NoError(t, err)

	return actual, res, fetchedFlow
}

func (fix *fixture) makeSuccessfulRegistration(t *testing.T, flowType string, expectReturnTo string, values func(v url.Values), opts ...submitPasskeyOption) (actual string) {
	actual, res, _ := fix.makeRegistration(t, flowType, values, opts...)
	if flowType == "spa" {
		expectReturnTo = fix.publicTS.URL
	}
	assert.Contains(t, res.Request.URL.String(), expectReturnTo, "%+v\n\t%s", res.Request, assertx.PrettifyJSONPayload(t, actual))
	return actual
}

func (fix *fixture) makeUnsuccessfulRegistration(t *testing.T, flowType string, expectReturnTo string, values func(v url.Values), opts ...submitPasskeyOption) (actual string, res *http.Response) {
	actual, res, _ = fix.makeRegistration(t, flowType, values, opts...)
	return actual, res
}

func (fix *fixture) createIdentityWithoutPasskey(t *testing.T) *identity.Identity {
	id := fix.createIdentity(t)
	delete(id.Credentials, identity.CredentialsTypePasskey)
	require.NoError(t, fix.reg.PrivilegedIdentityPool().UpdateIdentity(fix.ctx, id))
	return id
}

func (fix *fixture) createIdentityAndReturnIdentifier(t *testing.T, conf []byte) (*identity.Identity, string) {
	identifier := x.NewUUID().String() + "@ory.sh"
	password := x.NewUUID().String()
	p, err := fix.reg.Hasher(fix.ctx).Generate(fix.ctx, []byte(password))
	require.NoError(t, err)
	i := &identity.Identity{
		Traits: identity.Traits(fmt.Sprintf(`{"email":"%s"}`, identifier)),
		VerifiableAddresses: []identity.VerifiableAddress{
			{
				Value:     identifier,
				Verified:  false,
				CreatedAt: time.Now(),
			},
		},
	}
	if conf == nil {
		conf = []byte(`{"credentials":[{"id":"Zm9vZm9v","display_name":"foo"},{"id":"YmFyYmFy","display_name":"bar"}]}`)
	}
	require.NoError(t, fix.reg.PrivilegedIdentityPool().CreateIdentity(fix.ctx, i))
	i.Credentials = map[identity.CredentialsType]identity.Credentials{
		identity.CredentialsTypePassword: {
			Type:        identity.CredentialsTypePassword,
			Identifiers: []string{identifier},
			Config:      sqlxx.JSONRawMessage(`{"hashed_password":"` + string(p) + `"}`),
		},
		identity.CredentialsTypePasskey: {
			Type:        identity.CredentialsTypePasskey,
			Identifiers: []string{identifier},
			Config:      conf,
		},
	}
	require.NoError(t, fix.reg.PrivilegedIdentityPool().UpdateIdentity(fix.ctx, i))
	return i, identifier
}

func (fix *fixture) createIdentity(t *testing.T) *identity.Identity {
	id, _ := fix.createIdentityAndReturnIdentifier(t, nil)
	return id
}
