package totp_test

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/ory/x/assertx"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/text"

	"github.com/pquerna/otp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	stdtotp "github.com/pquerna/otp/totp"

	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/strategy/totp"
	"github.com/ory/kratos/x"
)

const totpCodeGJSONQuery = "ui.nodes.#(attributes.name==totp_code)"

func createIdentityWithoutTOTP(t *testing.T, reg driver.Registry) *identity.Identity {
	id, _, _ := createIdentity(t, reg)
	delete(id.Credentials, identity.CredentialsTypeTOTP)
	require.NoError(t, reg.PrivilegedIdentityPool().UpdateIdentity(context.Background(), id))
	return id
}

func createIdentity(t *testing.T, reg driver.Registry) (*identity.Identity, string, *otp.Key) {
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
	return i, password, key
}

func TestCompleteLogin(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword), map[string]interface{}{"enabled": true})
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeTOTP), map[string]interface{}{"enabled": true})
	conf.MustSet(config.ViperKeyURLsWhitelistedReturnToDomains, []string{"https://www.ory.sh"})

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

	t.Run("case=totp payload is set when identity has totp", func(t *testing.T) {
		id, _, _ := createIdentity(t, reg)

		apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)
		f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))
		testhelpers.SnapshotTExcept(t, f.Ui.Nodes, []string{
			"0.attributes.value",
		})
	})

	t.Run("case=totp payload is not set when identity has no totp", func(t *testing.T) {
		id := createIdentityWithoutTOTP(t, reg)

		apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)
		f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))
		assertx.EqualAsJSON(t, nil, f.Ui.Nodes)
	})

	t.Run("case=should show the error ui because the request payload is malformed", func(t *testing.T) {
		id, _, _ := createIdentity(t, reg)

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

	doAPIFlow := func(t *testing.T, v func(url.Values), id *identity.Identity) (string, *http.Response) {
		apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)
		f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))
		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		values.Set("method", "totp")
		v(values)
		payload := testhelpers.EncodeFormAsJSON(t, true, values)
		return testhelpers.LoginMakeRequest(t, true, false, f, apiClient, payload)
	}

	doBrowserFlow := func(t *testing.T, spa bool, v func(url.Values), id *identity.Identity, returnTo string) (string, *http.Response) {
		browserClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id)

		opts := []testhelpers.InitFlowWithOption{testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2)}
		if len(returnTo) > 0 {
			opts = append(opts, testhelpers.InitFlowWithReturnTo(returnTo))
		}

		f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, spa, opts...)
		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		values.Set("method", "totp")
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

	t.Run("case=should fail if code is empty", func(t *testing.T) {
		id, _, _ := createIdentity(t, reg)
		payload := func(v url.Values) {
			v.Set("totp_code", "")
		}

		check := func(t *testing.T, shouldRedirect bool, body string, res *http.Response) {
			checkURL(t, shouldRedirect, res)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Equal(t, "Property totp_code is missing.", gjson.Get(body, totpCodeGJSONQuery+".messages.0.text").String(), "%s", body)
		}

		t.Run("type=api", func(t *testing.T) {
			body, res := doAPIFlow(t, payload, id)
			check(t, false, body, res)
		})

		t.Run("type=browser", func(t *testing.T) {
			body, res := doBrowserFlow(t, false, payload, id, "")
			check(t, true, body, res)
		})

		t.Run("type=spa", func(t *testing.T) {
			body, res := doBrowserFlow(t, true, payload, id, "")
			check(t, false, body, res)
		})
	})

	t.Run("case=should fail if code is invalid", func(t *testing.T) {
		id, _, _ := createIdentity(t, reg)
		payload := func(v url.Values) {
			v.Set("totp_code", "111111")
		}

		check := func(t *testing.T, shouldRedirect bool, body string, res *http.Response) {
			checkURL(t, shouldRedirect, res)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Equal(t, text.NewErrorValidationTOTPVerifierWrong().Text, gjson.Get(body, "ui.messages.0.text").String(), "%s", body)
		}

		t.Run("type=api", func(t *testing.T) {
			body, res := doAPIFlow(t, payload, id)
			check(t, false, body, res)
		})

		t.Run("type=browser", func(t *testing.T) {
			body, res := doBrowserFlow(t, false, payload, id, "")
			check(t, true, body, res)
		})

		t.Run("type=spa", func(t *testing.T) {
			body, res := doBrowserFlow(t, true, payload, id, "")
			check(t, false, body, res)
		})
	})

	t.Run("case=should fail if code is too long", func(t *testing.T) {
		id, _, _ := createIdentity(t, reg)
		payload := func(v url.Values) {
			v.Set("totp_code", "1111111111")
		}

		check := func(t *testing.T, shouldRedirect bool, body string, res *http.Response) {
			checkURL(t, shouldRedirect, res)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Equal(t, text.NewErrorValidationTOTPVerifierWrong().Text, gjson.Get(body, "ui.messages.0.text").String(), "%s", body)
		}

		t.Run("type=api", func(t *testing.T) {
			body, res := doAPIFlow(t, payload, id)
			check(t, false, body, res)
		})

		t.Run("type=browser", func(t *testing.T) {
			body, res := doBrowserFlow(t, false, payload, id, "")
			check(t, true, body, res)
		})

		t.Run("type=spa", func(t *testing.T) {
			body, res := doBrowserFlow(t, true, payload, id, "")
			check(t, false, body, res)
		})
	})

	t.Run("case=should fail if TOTP was not set up for identity", func(t *testing.T) {
		id := createIdentityWithoutTOTP(t, reg)

		payload := func(v url.Values) {
			v.Set("totp_code", "111111")
		}

		check := func(t *testing.T, shouldRedirect bool, body string, res *http.Response) {
			checkURL(t, shouldRedirect, res)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Equal(t, text.NewErrorValidationNoTOTPDevice().Text, gjson.Get(body, "ui.messages.0.text").String(), "%s", body)
		}

		t.Run("type=api", func(t *testing.T) {
			body, res := doAPIFlow(t, payload, id)
			check(t, false, body, res)
		})

		t.Run("type=browser", func(t *testing.T) {
			body, res := doBrowserFlow(t, false, payload, id, "")
			check(t, true, body, res)
		})

		t.Run("type=spa", func(t *testing.T) {
			body, res := doBrowserFlow(t, true, payload, id, "")
			check(t, false, body, res)
		})
	})

	t.Run("case=should pass when TOTP is supplied correctly", func(t *testing.T) {
		id, _, key := createIdentity(t, reg)
		code, err := stdtotp.GenerateCode(key.Secret(), time.Now())
		require.NoError(t, err)
		payload := func(v url.Values) {
			v.Set("totp_code", code)
		}

		startAt := time.Now()
		check := func(t *testing.T, shouldRedirect bool, body string, res *http.Response) {
			prefix := "session."
			if shouldRedirect {
				assert.Contains(t, res.Request.URL.String(), redirTS.URL+"/return-ts")
				prefix = ""
			} else {
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
			}
			assert.True(t, gjson.Get(body, prefix+"active").Bool(), "%s", body)
			assert.EqualValues(t, identity.AuthenticatorAssuranceLevel2, gjson.Get(body, prefix+"authenticator_assurance_level").String())
			require.Len(t, gjson.Get(body, prefix+"authentication_methods").Array(), 2)
			assert.EqualValues(t, identity.CredentialsTypePassword, gjson.Get(body, prefix+"authentication_methods.0.method").String(), 2)
			assert.True(t, gjson.Get(body, prefix+"authentication_methods.0.completed_at").Time().After(startAt), 2)
			assert.EqualValues(t, identity.CredentialsTypeTOTP, gjson.Get(body, prefix+"authentication_methods.1.method").String(), 2)
			assert.True(t, gjson.Get(body, prefix+"authentication_methods.1.completed_at").Time().After(startAt), 2)
			assert.True(t, gjson.Get(body, prefix+"authentication_methods.1.completed_at").Time().After(gjson.Get(body, prefix+"authentication_methods.0.completed_at").Time()), 2)
			assert.Equal(t, gjson.Get(body, prefix+"authentication_methods.1.completed_at").Time().Unix(), gjson.Get(body, prefix+"authenticated_at").Time().Unix(), "%s", body)
		}

		t.Run("type=api", func(t *testing.T) {
			body, res := doAPIFlow(t, payload, id)
			check(t, false, body, res)
		})

		t.Run("type=browser", func(t *testing.T) {
			body, res := doBrowserFlow(t, false, payload, id, "")
			check(t, true, body, res)
		})

		t.Run("type=browser set return_to", func(t *testing.T) {
			returnTo := "https://www.ory.sh"
			_, res := doBrowserFlow(t, false, payload, id, returnTo)
			t.Log(res.Request.URL.String())
			assert.Contains(t, res.Request.URL.String(), returnTo)
		})

		t.Run("type=spa", func(t *testing.T) {
			body, res := doBrowserFlow(t, true, payload, id, "")
			check(t, false, body, res)
		})
	})

	t.Run("case=should fail because totp can not handle AAL1", func(t *testing.T) {
		apiClient := testhelpers.NewDebugClient(t)
		f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false)

		update, err := reg.LoginFlowPersister().GetLoginFlow(context.Background(), uuid.FromStringOrNil(f.Id))
		require.NoError(t, err)
		update.RequestedAAL = identity.AuthenticatorAssuranceLevel1
		require.NoError(t, reg.LoginFlowPersister().UpdateLoginFlow(context.Background(), update))

		req, err := http.NewRequest("POST", f.Ui.Action, bytes.NewBufferString(`{"method":"totp"}`))
		require.NoError(t, err)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		body := x.MustReadAll(res.Body)
		require.NoError(t, res.Body.Close())
		assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
		assert.Equal(t, text.NewErrorValidationLoginNoStrategyFound().Text, gjson.GetBytes(body, "ui.messages.0.text").String())
	})

	t.Run("case=should pass without csrf if API flow", func(t *testing.T) {
		id, _, _ := createIdentity(t, reg)
		body, res := doAPIFlow(t, func(v url.Values) {
			v.Del("csrf_token")
			v.Set("totp_code", "111111")
		}, id)

		assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
		assert.Equal(t, text.NewErrorValidationTOTPVerifierWrong().Text, gjson.Get(body, "ui.messages.0.text").String(), "%s", body)
	})

	t.Run("case=should fail if CSRF token is invalid", func(t *testing.T) {
		id, _, _ := createIdentity(t, reg)
		t.Run("type=browser", func(t *testing.T) {
			body, res := doBrowserFlow(t, false, func(v url.Values) {
				v.Del("csrf_token")
				v.Set("totp_code", "111111")
			}, id, "")

			assert.Contains(t, res.Request.URL.String(), errTS.URL)
			assert.Equal(t, x.ErrInvalidCSRFToken.Reason(), gjson.Get(body, "reason").String(), body)
		})

		t.Run("type=spa", func(t *testing.T) {
			body, res := doBrowserFlow(t, true, func(v url.Values) {
				v.Del("csrf_token")
				v.Set("totp_code", "111111")
			}, id, "")

			assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
			assert.Equal(t, x.ErrInvalidCSRFToken.Reason(), gjson.Get(body, "error.reason").String(), body)
		})
	})

	t.Run("case=should pass return_to URL after login", func(t *testing.T) {
		id, pwd, _ := createIdentity(t, reg)

		t.Run("type=browser", func(t *testing.T) {
			returnTo := "https://www.ory.sh"
			browserClient := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, false, testhelpers.InitFlowWithReturnTo(returnTo))

			cred, ok := id.GetCredentials(identity.CredentialsTypePassword)
			require.True(t, ok)
			values := url.Values{"method": {"password"}, "password_identifier": {cred.Identifiers[0]},
				"password": {pwd}, "csrf_token": {x.FakeCSRFToken}}.Encode()

			body, res := testhelpers.LoginMakeRequest(t, false, false, f, browserClient, values)
			require.Contains(t, res.Request.URL.Path, "login", "%s", res.Request.URL.String())
			assert.Equal(t, gjson.Get(body, "requested_aal").String(), "aal2", "%s", body)
			assert.Equal(t, gjson.Get(body, "return_to").String(), returnTo, "%s", body)
		})
	})
}
