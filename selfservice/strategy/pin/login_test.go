package pin_test

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/registrationhelpers"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/x"
	"github.com/ory/x/errorsx"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"net/http"
	"net/url"
	"testing"
	"time"
)

//go:embed stub/login.schema.json
var loginSchema []byte

func TestCompleteLogin(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePin),
		map[string]interface{}{"enabled": true})
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword),
		map[string]interface{}{"enabled": true})
	router := x.NewRouterPublic()
	publicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin())

	errTS := testhelpers.NewErrorTestServer(t, reg)
	uiTS := testhelpers.NewLoginUIFlowEchoServer(t, reg)
	redirTS := testhelpers.NewRedirSessionEchoTS(t, reg)

	// Overwrite these two:
	conf.MustSet(ctx, config.ViperKeySelfServiceErrorUI, errTS.URL+"/error-ts")
	conf.MustSet(ctx, config.ViperKeySelfServiceLoginUI, uiTS.URL+"/login-ts")

	testhelpers.SetDefaultIdentitySchemaFromRaw(conf, loginSchema)
	conf.MustSet(ctx, config.ViperKeySecretsDefault, []string{"not-a-secure-session-key"})

	ensureFieldsExist := func(t *testing.T, body []byte) {
		registrationhelpers.CheckFormContent(t, body, "pin", "csrf_token")
	}

	createIdentity := func(identifier, password string, pin string) *identity.Identity {
		hashedPassword, _ := reg.Hasher(ctx).Generate(ctx, []byte(password))
		hashedPin, _ := reg.Hasher(ctx).Generate(ctx, []byte(pin))
		iId := x.NewUUID()
		i := &identity.Identity{
			ID:     iId,
			Traits: identity.Traits(fmt.Sprintf(`{"subject":"%s"}`, identifier)),
			Credentials: map[identity.CredentialsType]identity.Credentials{
				identity.CredentialsTypePassword: {
					Type:        identity.CredentialsTypePassword,
					Identifiers: []string{identifier},
					Config:      sqlxx.JSONRawMessage(`{"hashed_password":"` + string(hashedPassword) + `"}`),
				},
				identity.CredentialsTypePin: {
					Type:        identity.CredentialsTypePin,
					Identifiers: []string{},
					Config:      sqlxx.JSONRawMessage(`{"hashed_pin":"` + string(hashedPin) + `"}`),
				},
			},
			VerifiableAddresses: []identity.VerifiableAddress{
				{
					ID:         x.NewUUID(),
					Value:      identifier,
					Verified:   false,
					CreatedAt:  time.Now(),
					IdentityID: iId,
				},
			},
		}
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(ctx, i))
		return i
	}

	var expectValidationError = func(t *testing.T, isAPI, isSPA bool,
		hc *http.Client, values func(url.Values)) string {
		return testhelpers.SubmitLoginForm(t, isAPI, hc, publicTS, values,
			isSPA, false,
			testhelpers.ExpectStatusCode(isAPI || isSPA, http.StatusBadRequest, http.StatusOK),
			testhelpers.ExpectURL(isAPI || isSPA, publicTS.URL+login.RouteSubmitFlow, conf.SelfServiceFlowLoginUI(ctx).String()),
			testhelpers.InitFlowWithAAL(identity.NoAuthenticatorAssuranceLevel))
	}
	var expectUnauthorizedError = func(t *testing.T, isAPI, refresh, isSPA bool, values func(url.Values)) string {
		return testhelpers.SubmitLoginForm(t, isAPI, nil, publicTS, values,
			isSPA, refresh,
			testhelpers.ExpectStatusCode(isAPI || isSPA, http.StatusUnauthorized, http.StatusOK),
			testhelpers.ExpectURL(isAPI || isSPA, publicTS.URL+login.RouteSubmitFlow, conf.SelfServiceFlowErrorURL(ctx).String()))
	}

	t.Run("should return an error because no session", func(t *testing.T) {
		var check = func(t *testing.T, isAPI bool, body string) {
			var path = ""
			if isAPI {
				path = "error.id"
			} else {
				path = "id"
			}
			assert.Equal(t, "session_inactive", gjson.Get(body, path).String(), "%s", body)
		}

		var values = func(v url.Values) {
			v.Set("method", "pin")
			v.Set("pin", "not-pin")
		}

		t.Run("type=browser", func(t *testing.T) {
			check(t, false, expectUnauthorizedError(t, false, false, false, values))
		})

		t.Run("type=api", func(t *testing.T) {
			check(t, true, expectUnauthorizedError(t, true, false, false, values))
		})
		t.Run("type=spa", func(t *testing.T) {
			check(t, true, expectUnauthorizedError(t, true, false, true, values))
		})
	})

	t.Run("should return an error because no pin is set", func(t *testing.T) {
		identifier, pwd, pin := x.NewUUID().String(), "password", "1234"
		i := createIdentity(identifier, pwd, pin)

		var check = func(t *testing.T, body string) {
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.action").String(), publicTS.URL+login.RouteSubmitFlow, "%s", body)

			ensureFieldsExist(t, []byte(body))
			assert.Equal(t, "Property pin is missing.", gjson.Get(body, "ui.nodes.#(attributes.name==pin).messages.0.text").String(), "%s", body)
			assert.Len(t, gjson.Get(body, "ui.nodes").Array(), 5)

			// This must not include the pin!
			assert.Empty(t, gjson.Get(body, "ui.nodes.#(attributes.name==pin).attributes.value").String())
		}

		var values = func(v url.Values) {
			v.Set("method", "pin")
			v.Del("pin")
		}

		t.Run("type=browser", func(t *testing.T) {
			hc := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, i)
			check(t, expectValidationError(t, false, true, hc, values))
		})

		t.Run("type=api", func(t *testing.T) {
			hc := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, i)
			check(t, expectValidationError(t, true, true, hc, values))
		})
	})

	t.Run("should return an error because the credentials are invalid (pin not correct)", func(t *testing.T) {
		var check = func(t *testing.T, body string) {
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.action").String(), publicTS.URL+login.RouteSubmitFlow, "%s", body)

			ensureFieldsExist(t, []byte(body))
			assert.Equal(t,
				errorsx.Cause(schema.NewInvalidPinError()).(*schema.ValidationError).Messages[0].Text,
				gjson.Get(body, "ui.messages.0.text").String(),
				"%s", body,
			)

			// This must not include the pin!
			assert.Empty(t, gjson.Get(body, "ui.nodes.#(attributes.name==pin).attributes.value").String())
		}

		identifier, pwd, pin := x.NewUUID().String(), "password", "1234"
		i := createIdentity(identifier, pwd, pin)

		var values = func(v url.Values) {
			v.Set("method", "pin")
			v.Set("pin", "not-pin")
		}

		t.Run("type=api", func(t *testing.T) {
			hc := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, i)
			check(t, expectValidationError(t, true, true, hc, values))
		})

		t.Run("type=browser", func(t *testing.T) {
			hc := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, i)
			check(t, expectValidationError(t, false, true, hc, values))
		})

		t.Run("type=spa", func(t *testing.T) {
			hc := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, i)
			check(t, expectValidationError(t, false, true, hc, values))
		})

	})

	t.Run("should pass with real request", func(t *testing.T) {
		identifier, pwd, pin := x.NewUUID().String(), "password", "1234"
		i := createIdentity(identifier, pwd, pin)

		var values = func(v url.Values) {
			v.Set("method", "pin")
			v.Set("pin", "1234")
		}

		t.Run("type=api", func(t *testing.T) {
			hc := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, i)
			body := testhelpers.SubmitLoginForm(t, true, hc, publicTS, values,
				false, false, http.StatusOK, publicTS.URL+login.RouteSubmitFlow,
				testhelpers.InitFlowWithAAL(identity.NoAuthenticatorAssuranceLevel))

			assert.Equal(t, identifier, gjson.Get(body, "session.identity.traits.subject").String(), "%s", body)
			st := gjson.Get(body, "session_token").String()
			assert.NotEmpty(t, st, "%s", body)
			assert.Equal(t, int64(2), gjson.Get(body, "session.authentication_methods.#").Int(), "%s", body)
			assert.Equal(t, "pin", gjson.Get(body, "session.authentication_methods.1.method").String(), "%s", body)
			assert.Equal(t, "aal0", gjson.Get(body, "session.authentication_methods.1.aal").String(), "%s", body)
		})

		t.Run("type=browser", func(t *testing.T) {
			hc := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, i)
			body := testhelpers.SubmitLoginForm(t, false, hc, publicTS, values,
				false, false, http.StatusOK, redirTS.URL,
				testhelpers.InitFlowWithAAL(identity.NoAuthenticatorAssuranceLevel))

			assert.Equal(t, identifier, gjson.Get(body, "identity.traits.subject").String(), "%s", body)
		})

		t.Run("type=spa", func(t *testing.T) {
			hc := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, i)
			body := testhelpers.SubmitLoginForm(t, false, hc, publicTS, values,
				true, false, http.StatusOK, publicTS.URL+login.RouteSubmitFlow,
				testhelpers.InitFlowWithAAL(identity.NoAuthenticatorAssuranceLevel))

			assert.Equal(t, identifier, gjson.Get(body, "session.identity.traits.subject").String(), "%s", body)
			assert.Empty(t, gjson.Get(body, "session_token").String(), "%s", body)
			assert.Equal(t, int64(2), gjson.Get(body, "session.authentication_methods.#").Int(), "%s", body)
			assert.Equal(t, "pin", gjson.Get(body, "session.authentication_methods.1.method").String(), "%s", body)
			assert.Equal(t, "aal0", gjson.Get(body, "session.authentication_methods.1.aal").String(), "%s", body)

			// Was the session cookie set?
			require.NotEmpty(t, hc.Jar.Cookies(urlx.ParseOrPanic(publicTS.URL)), "%+v", hc.Jar)
		})

	})

}
