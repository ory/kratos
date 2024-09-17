// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ory/kratos/courier"

	"github.com/ory/kratos/selfservice/strategy/idfirst"

	configtesthelpers "github.com/ory/kratos/driver/config/testhelpers"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/selfservice/flow/login"

	"github.com/ory/kratos/selfservice/flow"

	"github.com/ory/x/ioutilx"
	"github.com/ory/x/snapshotx"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/stringsx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	oryClient "github.com/ory/kratos/internal/httpclient"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
	"github.com/ory/x/sqlxx"
)

func createIdentity(ctx context.Context, t *testing.T, reg driver.Registry, withoutCodeCredential bool, moreIdentifiers ...string) *identity.Identity {
	t.Helper()
	i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
	i.NID = x.NewUUID()
	email := testhelpers.RandomEmail()

	ids := fmt.Sprintf(`"email":"%s"`, email)
	for i, identifier := range moreIdentifiers {
		ids = fmt.Sprintf(`%s,"email_%d":"%s"`, ids, i+1, identifier)
	}

	i.Traits = identity.Traits(fmt.Sprintf(`{"tos": true, %s}`, ids))

	credentials := map[identity.CredentialsType]identity.Credentials{
		identity.CredentialsTypePassword: {Identifiers: append([]string{email}, moreIdentifiers...), Type: identity.CredentialsTypePassword, Config: sqlxx.JSONRawMessage("{\"some\" : \"secret\"}")},
		identity.CredentialsTypeOIDC:     {Type: identity.CredentialsTypeOIDC, Identifiers: append([]string{email}, moreIdentifiers...), Config: sqlxx.JSONRawMessage("{\"some\" : \"secret\"}")},
		identity.CredentialsTypeWebAuthn: {Type: identity.CredentialsTypeWebAuthn, Identifiers: append([]string{email}, moreIdentifiers...), Config: sqlxx.JSONRawMessage("{\"some\" : \"secret\", \"user_handle\": \"rVIFaWRcTTuQLkXFmQWpgA==\"}")},
	}
	if !withoutCodeCredential {
		credentials[identity.CredentialsTypeCodeAuth] = identity.Credentials{Type: identity.CredentialsTypeCodeAuth, Identifiers: append([]string{email}, moreIdentifiers...), Config: sqlxx.JSONRawMessage(`{"addresses":[{"channel":"email","address":"` + email + `"}]}`)}
	}
	i.Credentials = credentials

	var va []identity.VerifiableAddress
	for _, identifier := range moreIdentifiers {
		va = append(va, identity.VerifiableAddress{Value: identifier, Verified: false, Status: identity.VerifiableAddressStatusCompleted})
	}

	va = append(va, identity.VerifiableAddress{Value: email, Verified: true, Status: identity.VerifiableAddressStatusCompleted})

	i.VerifiableAddresses = va

	require.NoError(t, reg.IdentityManager().Create(ctx, i))
	return i
}

func TestLoginCodeStrategy(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/code.identity.schema.json")
	conf.MustSet(ctx, fmt.Sprintf("%s.%s.enabled", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypeCodeAuth.String()), true)
	conf.MustSet(ctx, fmt.Sprintf("%s.%s.passwordless_enabled", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypeCodeAuth.String()), true)
	conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh")
	conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{"https://www.ory.sh"})

	_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)

	public, _, _, _ := testhelpers.NewKratosServerWithCSRFAndRouters(t, reg)

	type state struct {
		flowID        string
		identity      *identity.Identity
		client        *http.Client
		loginCode     string
		identityEmail string
		testServer    *httptest.Server
		body          string
	}

	type ApiType string

	const (
		ApiTypeBrowser ApiType = "browser"
		ApiTypeSPA     ApiType = "spa"
		ApiTypeNative  ApiType = "api"
	)

	createLoginFlowWithIdentity := func(ctx context.Context, t *testing.T, public *httptest.Server, apiType ApiType, user *identity.Identity) *state {
		t.Helper()

		var client *http.Client
		if apiType == ApiTypeNative {
			client = &http.Client{}
		} else {
			client = testhelpers.NewClientWithCookies(t)
		}

		client.Transport = testhelpers.NewTransportWithLogger(http.DefaultTransport, t).RoundTripper

		var clientInit *oryClient.LoginFlow
		if apiType == ApiTypeNative {
			clientInit = testhelpers.InitializeLoginFlowViaAPI(t, client, public, false)
		} else {
			clientInit = testhelpers.InitializeLoginFlowViaBrowser(t, client, public, false, apiType == ApiTypeSPA, false, false)
		}

		body, err := json.Marshal(clientInit)
		require.NoError(t, err)

		csrfToken := gjson.GetBytes(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String()
		if apiType == ApiTypeNative {
			require.Emptyf(t, csrfToken, "csrf_token should be empty in native flows, but was found in: %s", body)
		} else {
			require.NotEmptyf(t, csrfToken, "could not find csrf_token in: %s", body)
		}

		return &state{
			flowID:     clientInit.GetId(),
			identity:   user,
			client:     client,
			testServer: public,
		}
	}

	createLoginFlow := func(ctx context.Context, t *testing.T, public *httptest.Server, apiType ApiType, withoutCodeCredential bool, moreIdentifiers ...string) *state {
		t.Helper()
		s := createLoginFlowWithIdentity(ctx, t, public, apiType, createIdentity(ctx, t, reg, withoutCodeCredential, moreIdentifiers...))
		loginEmail := gjson.Get(s.identity.Traits.String(), "email").String()
		require.NotEmptyf(t, loginEmail, "could not find the email trait inside the identity: %s", s.identity.Traits.String())
		s.identityEmail = loginEmail
		return s
	}

	type onSubmitAssertion func(t *testing.T, s *state, body string, res *http.Response)

	submitLogin := func(ctx context.Context, t *testing.T, s *state, apiType ApiType, vals func(v *url.Values), mustHaveSession bool, submitAssertion onSubmitAssertion) *state {
		t.Helper()

		lf, resp, err := testhelpers.NewSDKCustomClient(s.testServer, s.client).FrontendAPI.GetLoginFlow(ctx).Id(s.flowID).Execute()
		require.NoError(t, err)
		require.EqualValues(t, http.StatusOK, resp.StatusCode)

		values := testhelpers.SDKFormFieldsToURLValues(lf.Ui.Nodes)
		// we need to remove resend here
		// since it is not required for the first request
		// subsequent requests might need it later
		values.Del("resend")
		values.Set("method", "code")
		vals(&values)

		body, resp := testhelpers.LoginMakeRequest(t, apiType == ApiTypeNative, apiType == ApiTypeSPA, lf, s.client, testhelpers.EncodeFormAsJSON(t, apiType == ApiTypeNative, values))

		if submitAssertion != nil {
			submitAssertion(t, s, body, resp)
			return s
		}
		s.body = body

		if mustHaveSession {
			req, err := http.NewRequest("GET", s.testServer.URL+session.RouteWhoami, nil)
			require.NoError(t, err)

			if apiType == ApiTypeNative {
				req.Header.Set("Authorization", "Bearer "+gjson.Get(body, "session_token").String())
			}

			resp, err = s.client.Do(req)
			require.NoError(t, err)
			body = string(ioutilx.MustReadAll(resp.Body))
			require.EqualValues(t, http.StatusOK, resp.StatusCode, "%s", body)
		} else {
			// SPAs need to be informed that the login has not yet completed using status 400.
			// Browser clients will redirect back to the login URL.
			if apiType == ApiTypeBrowser {
				require.EqualValues(t, http.StatusOK, resp.StatusCode, "%s", body)
			} else {
				require.EqualValues(t, http.StatusBadRequest, resp.StatusCode, "%s", body)
			}
		}

		return s
	}

	for _, tc := range []struct {
		d       string
		apiType ApiType
	}{
		{
			d:       "SPA client",
			apiType: ApiTypeSPA,
		},
		{
			d:       "Browser client",
			apiType: ApiTypeBrowser,
		},
		{
			d:       "Native client",
			apiType: ApiTypeNative,
		},
	} {
		t.Run("test="+tc.d, func(t *testing.T) {
			t.Run("case=email identifier should be case insensitive", func(t *testing.T) {
				// create login flow
				s := createLoginFlow(ctx, t, public, tc.apiType, false)

				// submit email
				s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("identifier", stringsx.ToUpperInitial(s.identityEmail))
				}, false, nil)

				message := testhelpers.CourierExpectMessage(ctx, t, reg, s.identityEmail, "Use code")
				assert.Contains(t, message.Body, "Login to your account with the following code")

				loginCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
				assert.NotEmpty(t, loginCode)

				// 3. Submit OTP
				submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("code", loginCode)
				}, true, nil)
			})

			t.Run("case=should be able to log in with code sent to email", func(t *testing.T) {
				// create login flow
				s := createLoginFlow(ctx, t, public, tc.apiType, false)

				// submit email
				s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("identifier", s.identityEmail)
				}, false, nil)

				message := testhelpers.CourierExpectMessage(ctx, t, reg, s.identityEmail, "Use code")
				assert.Contains(t, message.Body, "Login to your account with the following code")

				loginCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
				assert.NotEmpty(t, loginCode)

				// 3. Submit OTP
				state := submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("code", loginCode)
				}, true, nil)
				if tc.apiType == ApiTypeSPA {
					assert.EqualValues(t, flow.ContinueWithActionRedirectBrowserToString, gjson.Get(state.body, "continue_with.0.action").String(), "%s", state.body)
					assert.Contains(t, gjson.Get(state.body, "continue_with.0.redirect_browser_to").String(), conf.SelfServiceBrowserDefaultReturnTo(ctx).String(), "%s", state.body)
				} else {
					assert.Empty(t, gjson.Get(state.body, "continue_with").Array(), "%s", state.body)
				}
			})

			t.Run("case=should be able to log in legacy cases", func(t *testing.T) {
				run := func(t *testing.T, s *state) {
					// submit email
					s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
						v.Set("identifier", s.identityEmail)
					}, false, nil)

					t.Logf("s.body: %s", s.body)

					message := testhelpers.CourierExpectMessage(ctx, t, reg, s.identityEmail, "Use code")
					assert.Contains(t, message.Body, "Login to your account with the following code")

					loginCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
					assert.NotEmpty(t, loginCode)

					// 3. Submit OTP
					state := submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
						v.Set("code", loginCode)
					}, true, nil)
					if tc.apiType == ApiTypeSPA {
						assert.Contains(t, gjson.Get(state.body, "continue_with.0.redirect_browser_to").String(), conf.SelfServiceBrowserDefaultReturnTo(ctx).String(), "%s", state.body)
					} else {
						assert.Empty(t, gjson.Get(state.body, "continue_with").Array(), "%s", state.body)
					}
				}

				initDefault := func(t *testing.T, cf string) *state {
					i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
					i.NID = x.NewUUID()

					// valid fake phone number for libphonenumber
					email := testhelpers.RandomEmail()
					i.Traits = identity.Traits(fmt.Sprintf(`{"tos": true, "email": "%s"}`, email))
					i.Credentials = map[identity.CredentialsType]identity.Credentials{
						identity.CredentialsTypeCodeAuth: {
							Type:        identity.CredentialsTypeCodeAuth,
							Identifiers: []string{email},
							Version:     0,
							Config:      sqlxx.JSONRawMessage(cf),
						},
					}
					require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentities(ctx, i)) // We explicitly bypass identity validation to test the legacy code path
					s := createLoginFlowWithIdentity(ctx, t, public, tc.apiType, i)
					s.identityEmail = email
					return s
				}

				t.Run("case=should be able to send address type with spaces", func(t *testing.T) {
					run(t,
						initDefault(t, `{"address_type": "email                               ", "used_at": {"Time": "0001-01-01T00:00:00Z", "Valid": false}}`),
					)
				})

				t.Run("case=should be able to send to empty address type", func(t *testing.T) {
					run(t,
						initDefault(t, `{"address_type": "", "used_at": {"Time": "0001-01-01T00:00:00Z", "Valid": false}}`),
					)
				})

				t.Run("case=should be able to send to empty credentials config", func(t *testing.T) {
					run(t,
						initDefault(t, `{}`),
					)
				})

				t.Run("case=should be able to send to identity with no credentials at all when fallback is enabled", func(t *testing.T) {
					conf.MustSet(ctx, config.ViperKeyCodeConfigMissingCredentialFallbackEnabled, true)
					t.Cleanup(func() {
						conf.MustSet(ctx, config.ViperKeyCodeConfigMissingCredentialFallbackEnabled, nil)
					})

					i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
					i.NID = x.NewUUID()
					email := testhelpers.RandomEmail()
					i.Traits = identity.Traits(fmt.Sprintf(`{"tos": true, "email": "%s"}`, email))
					i.Credentials = map[identity.CredentialsType]identity.Credentials{
						// This makes it possible for our code to find the identity identifier here.
						identity.CredentialsTypePassword: {Type: identity.CredentialsTypePassword, Identifiers: []string{email}, Config: sqlxx.JSONRawMessage(`{}`)},
					}

					require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentities(ctx, i)) // We explicitly bypass identity validation to test the legacy code path
					s := createLoginFlowWithIdentity(ctx, t, public, tc.apiType, i)
					s.identityEmail = email
					run(t, s)
				})

				t.Run("case=should fail to send to identity with no credentials at all when fallback is disabled", func(t *testing.T) {
					conf.MustSet(ctx, config.ViperKeyCodeConfigMissingCredentialFallbackEnabled, false)
					t.Cleanup(func() {
						conf.MustSet(ctx, config.ViperKeyCodeConfigMissingCredentialFallbackEnabled, nil)
					})

					i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
					i.NID = x.NewUUID()
					email := testhelpers.RandomEmail()
					i.Traits = identity.Traits(fmt.Sprintf(`{"tos": true, "email": "%s"}`, email))
					require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentities(ctx, i)) // We explicitly bypass identity validation to test the legacy code path
					s := createLoginFlowWithIdentity(ctx, t, public, tc.apiType, i)
					s.identityEmail = email
					// submit email
					s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
						v.Set("identifier", s.identityEmail)
					}, false, nil)
					assert.Contains(t, s.body, "4000035", "Should not find the account")
				})
			})

			t.Run("case=should be able to log in with code to sms and normalize the number", func(t *testing.T) {
				i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
				i.NID = x.NewUUID()

				// valid fake phone number for libphonenumber
				phone := "+1 (415) 55526-71"
				i.Traits = identity.Traits(fmt.Sprintf(`{"tos": true, "phone_1": "%s"}`, phone))
				require.NoError(t, reg.IdentityManager().Create(ctx, i))
				t.Cleanup(func() {
					require.NoError(t, reg.PrivilegedIdentityPool().DeleteIdentity(ctx, i.ID))
				})

				s := createLoginFlowWithIdentity(ctx, t, public, tc.apiType, i)

				// submit email
				s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("identifier", phone)
				}, false, nil)

				message := testhelpers.CourierExpectMessage(ctx, t, reg, x.GracefulNormalization(phone), "Your login code is:")
				loginCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
				assert.NotEmpty(t, loginCode)

				// 3. Submit OTP
				state := submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("code", loginCode)
				}, true, nil)
				if tc.apiType == ApiTypeSPA {
					assert.EqualValues(t, flow.ContinueWithActionRedirectBrowserToString, gjson.Get(state.body, "continue_with.0.action").String(), "%s", state.body)
					assert.Contains(t, gjson.Get(state.body, "continue_with.0.redirect_browser_to").String(), conf.SelfServiceBrowserDefaultReturnTo(ctx).String(), "%s", state.body)
				} else {
					assert.Empty(t, gjson.Get(state.body, "continue_with").Array(), "%s", state.body)
				}
			})

			t.Run("case=new identities automatically have login with code", func(t *testing.T) {
				ctx := context.Background()

				conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationEnabled, true)
				conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".password.enabled", true)

				client := testhelpers.NewClientWithCookies(t)
				client.Transport = testhelpers.NewTransportWithLogger(http.DefaultTransport, t).RoundTripper

				registrationFlow := testhelpers.InitializeRegistrationFlowViaBrowser(t, client, public, tc.apiType == ApiTypeNative, false, false)

				email := testhelpers.RandomEmail()

				values := testhelpers.SDKFormFieldsToURLValues(registrationFlow.Ui.Nodes)
				values.Set("traits.email", email)
				values.Set("method", "password")
				values.Set("traits.tos", "1")
				values.Set("password", x.NewUUID().String())

				_, resp := testhelpers.RegistrationMakeRequest(t, tc.apiType == ApiTypeNative, tc.apiType == ApiTypeSPA, registrationFlow, client, testhelpers.EncodeFormAsJSON(t, tc.apiType == ApiTypeNative, values))
				require.EqualValues(t, http.StatusOK, resp.StatusCode)

				_, _, err := reg.PrivilegedIdentityPool().FindByCredentialsIdentifier(ctx, identity.CredentialsTypeCodeAuth, email)
				require.NoError(t, err, sqlcon.ErrNoRows)

				s := createLoginFlow(ctx, t, public, tc.apiType, true)

				s.identityEmail = email
				// submit email
				s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("identifier", s.identityEmail)
				}, false, nil)

				message := testhelpers.CourierExpectMessage(ctx, t, reg, s.identityEmail, "Use code")
				assert.Contains(t, message.Body, "Login to your account with the following code")

				loginCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
				assert.NotEmpty(t, loginCode)

				// submit code
				s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("code", loginCode)
				}, true, nil)

				// assert that the identity contains a code credential
				identity, cred, err := reg.PrivilegedIdentityPool().FindByCredentialsIdentifier(ctx, identity.CredentialsTypeCodeAuth, s.identityEmail)
				require.NoError(t, err)
				require.NotNil(t, cred)
				assert.Equal(t, identity.ID, cred.IdentityID)
			})

			t.Run("case=old identities should be able to login with code", func(t *testing.T) {
				// createLoginFlow uses the persister layer to create the identity
				// we pass in `true` to not do automatic code credential creation
				s := createLoginFlow(ctx, t, public, tc.apiType, true)

				// submit email
				s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("identifier", s.identityEmail)
				}, false, nil)

				message := testhelpers.CourierExpectMessage(ctx, t, reg, s.identityEmail, "Use code")
				assert.Contains(t, message.Body, "Login to your account with the following code")

				loginCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
				assert.NotEmpty(t, loginCode)

				// submit code
				s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("code", loginCode)
				}, true, nil)

				// assert that the identity contains a code credential
				identity, cred, err := reg.PrivilegedIdentityPool().FindByCredentialsIdentifier(ctx, identity.CredentialsTypeCodeAuth, s.identityEmail)
				require.NoError(t, err)
				require.NotNil(t, cred)
				assert.Equal(t, identity.ID, cred.IdentityID)
			})

			t.Run("case=should not be able to change submitted id on code submit", func(t *testing.T) {
				// create login flow
				s := createLoginFlow(ctx, t, public, tc.apiType, false)

				// submit email
				s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("identifier", s.identityEmail)
				}, false, nil)

				message := testhelpers.CourierExpectMessage(ctx, t, reg, s.identityEmail, "Use code")
				assert.Contains(t, message.Body, "Login to your account with the following code")

				loginCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
				assert.NotEmpty(t, loginCode)

				// 3. Submit OTP
				s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("identifier", "not-"+s.identityEmail)
					v.Set("code", loginCode)
				}, false, func(t *testing.T, s *state, body string, resp *http.Response) {
					if tc.apiType == ApiTypeBrowser {
						require.EqualValues(t, http.StatusOK, resp.StatusCode)
						require.EqualValues(t, conf.SelfServiceFlowLoginUI(ctx).Path, resp.Request.URL.Path)
						lf, resp, err := testhelpers.NewSDKCustomClient(public, s.client).FrontendAPI.GetLoginFlow(ctx).Id(s.flowID).Execute()
						require.NoError(t, err)
						require.EqualValues(t, http.StatusOK, resp.StatusCode)
						body, err := json.Marshal(lf)
						require.NoError(t, err)
						assert.Contains(t, gjson.GetBytes(body, "ui.messages.0.text").String(), "account does not exist or has not setup sign in with code")
					} else {
						require.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
						assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "account does not exist or has not setup sign in with code")
					}
				})
			})

			t.Run("case=should not be able to proceed to code entry when the account is unknown", func(t *testing.T) {
				s := createLoginFlow(ctx, t, public, tc.apiType, false)

				// submit email
				s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("identifier", testhelpers.RandomEmail())
				}, false, func(t *testing.T, s *state, body string, resp *http.Response) {
					if tc.apiType == ApiTypeBrowser {
						require.EqualValues(t, http.StatusOK, resp.StatusCode)
						require.EqualValues(t, conf.SelfServiceFlowLoginUI(ctx).Path, resp.Request.URL.Path)

						lf, resp, err := testhelpers.NewSDKCustomClient(public, s.client).FrontendAPI.GetLoginFlow(ctx).Id(s.flowID).Execute()
						require.NoError(t, err)
						require.EqualValues(t, http.StatusOK, resp.StatusCode)
						body, err := json.Marshal(lf)
						require.NoError(t, err)
						assert.Contains(t, gjson.GetBytes(body, "ui.messages.0.text").String(), "account does not exist or has not setup sign in with code")
					} else {
						require.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
						assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "account does not exist or has not setup sign in with code")
					}
				})
			})

			t.Run("case=should not be able to use valid code after 5 attempts", func(t *testing.T) {
				s := createLoginFlow(ctx, t, public, tc.apiType, false)

				// submit email
				s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("identifier", s.identityEmail)
				}, false, nil)

				message := testhelpers.CourierExpectMessage(ctx, t, reg, s.identityEmail, "Use code")
				assert.Contains(t, message.Body, "Login to your account with the following code")
				loginCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
				assert.NotEmpty(t, loginCode)

				for i := 0; i < 5; i++ {
					// 3. Submit OTP
					s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
						v.Set("code", "111111")
						v.Set("identifier", s.identityEmail)
					}, false, func(t *testing.T, s *state, body string, resp *http.Response) {
						if tc.apiType == ApiTypeBrowser {
							// in browser flows we redirect back to the login ui
							require.Equal(t, http.StatusOK, resp.StatusCode, "%s", body)
						} else {
							require.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
						}
						assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "The login code is invalid or has already been used")
					})
				}

				// 3. Submit OTP
				s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("code", loginCode)
					v.Set("identifier", s.identityEmail)
				}, false, func(t *testing.T, s *state, body string, resp *http.Response) {
					if tc.apiType == ApiTypeBrowser {
						// in browser flows we redirect back to the login ui
						require.Equal(t, http.StatusOK, resp.StatusCode, "%s", body)
					} else {
						require.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
					}
					assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "The request was submitted too often.")
				})
			})

			t.Run("case=code should expire", func(t *testing.T) {
				ctx := context.Background()

				conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.config.lifespan", "1ns")

				t.Cleanup(func() {
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.config.lifespan", "1h")
				})

				s := createLoginFlow(ctx, t, public, tc.apiType, false)

				// submit email
				s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("identifier", s.identityEmail)
				}, false, nil)

				message := testhelpers.CourierExpectMessage(ctx, t, reg, s.identityEmail, "Use code")
				assert.Contains(t, message.Body, "Login to your account with the following code")
				loginCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
				assert.NotEmpty(t, loginCode)

				submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("code", loginCode)
					v.Set("identifier", s.identityEmail)
				}, false, func(t *testing.T, s *state, body string, resp *http.Response) {
					if tc.apiType == ApiTypeBrowser {
						// with browser clients we redirect back to the UI with a new flow id as a query parameter
						require.Equal(t, http.StatusOK, resp.StatusCode)
						require.Equal(t, conf.SelfServiceFlowLoginUI(ctx).Path, resp.Request.URL.Path)
						lf, _, err := testhelpers.NewSDKCustomClient(public, s.client).FrontendAPI.GetLoginFlow(ctx).Id(resp.Request.URL.Query().Get("flow")).Execute()
						require.NoError(t, err)
						require.EqualValues(t, http.StatusOK, resp.StatusCode)

						body, err := json.Marshal(lf)
						require.NoError(t, err)
						assert.Contains(t, gjson.GetBytes(body, "ui.messages.0.text").String(), "flow expired 0.00 minutes ago")
					} else {
						require.EqualValues(t, http.StatusGone, resp.StatusCode)
						require.Contains(t, gjson.Get(body, "error.reason").String(), "self-service flow expired 0.00 minutes ago")
					}
				})
			})

			t.Run("case=resend code should invalidate previous code", func(t *testing.T) {
				ctx := context.Background()

				s := createLoginFlow(ctx, t, public, tc.apiType, false)

				s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("identifier", s.identityEmail)
				}, false, nil)

				message := testhelpers.CourierExpectMessage(ctx, t, reg, s.identityEmail, "Use code")
				assert.Contains(t, message.Body, "Login to your account with the following code")
				loginCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
				assert.NotEmpty(t, loginCode)

				// resend code
				s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("resend", "code")
					v.Set("identifier", s.identityEmail)
				}, false, nil)

				message = testhelpers.CourierExpectMessage(ctx, t, reg, s.identityEmail, "Use code")
				assert.Contains(t, message.Body, "Login to your account with the following code")
				loginCode2 := testhelpers.CourierExpectCodeInMessage(t, message, 1)
				assert.NotEmpty(t, loginCode2)

				assert.NotEqual(t, loginCode, loginCode2)
				s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("code", loginCode)
					v.Set("identifier", s.identityEmail)
				}, false, func(t *testing.T, s *state, body string, res *http.Response) {
					if tc.apiType == ApiTypeBrowser {
						require.EqualValues(t, http.StatusOK, res.StatusCode)
					} else {
						require.EqualValues(t, http.StatusBadRequest, res.StatusCode)
					}
					require.Contains(t, gjson.Get(body, "ui.messages").String(), "The login code is invalid or has already been used. Please try again")
				})

				s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("code", loginCode2)
					v.Set("identifier", s.identityEmail)
				}, true, nil)
			})

			t.Run("case=on login with un-verified address, should verify it", func(t *testing.T) {
				s := createLoginFlow(ctx, t, public, tc.apiType, false, testhelpers.RandomEmail())

				// we need to fetch only the first email
				loginEmail := gjson.Get(s.identity.Traits.String(), "email_1").String()
				require.NotEmpty(t, loginEmail)

				s.identityEmail = loginEmail

				var va *identity.VerifiableAddress

				for _, v := range s.identity.VerifiableAddresses {
					if v.Value == loginEmail {
						va = &v
						break
					}
				}

				require.NotNil(t, va)
				require.False(t, va.Verified)

				// submit email
				s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("identifier", s.identityEmail)
				}, false, nil)

				message := testhelpers.CourierExpectMessage(ctx, t, reg, loginEmail, "Use code")
				require.Contains(t, message.Body, "Login to your account with the following code")

				loginCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
				require.NotEmpty(t, loginCode)

				// Submit OTP
				s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("code", loginCode)
					v.Set("identifier", s.identityEmail)
				}, true, nil)

				id, err := reg.PrivilegedIdentityPool().GetIdentity(ctx, s.identity.ID, identity.ExpandEverything)
				require.NoError(t, err)

				va = nil

				for _, v := range id.VerifiableAddresses {
					if v.Value == loginEmail {
						va = &v
						break
					}
				}

				require.NotNil(t, va)
				require.True(t, va.Verified)
			})

			t.Run("suite=mfa", func(t *testing.T) {
				ctx := context.Background()
				conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.passwordless_enabled", false)
				conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.mfa_enabled", true)
				t.Cleanup(func() {
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.passwordless_enabled", true)
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.mfa_enabled", false)
				})

				t.Run("case=should be able to get AAL2 session", func(t *testing.T) {
					run := func(t *testing.T, withoutCodeCredential bool, overrideCodeCredential *identity.Credentials, overrideAllCredentials map[identity.CredentialsType]identity.Credentials) (*state, *http.Client) {
						user := createIdentity(ctx, t, reg, withoutCodeCredential)
						if overrideCodeCredential != nil {
							toUpdate := user.Credentials[identity.CredentialsTypeCodeAuth]
							if overrideCodeCredential.Config != nil {
								toUpdate.Config = overrideCodeCredential.Config
							}
							if overrideCodeCredential.Identifiers != nil {
								toUpdate.Identifiers = overrideCodeCredential.Identifiers
							}
							user.Credentials[identity.CredentialsTypeCodeAuth] = toUpdate
						}
						if overrideAllCredentials != nil {
							user.Credentials = overrideAllCredentials
						}
						require.NoError(t, reg.PrivilegedIdentityPool().UpdateIdentity(ctx, user))

						var cl *http.Client
						var f *oryClient.LoginFlow
						if tc.apiType == ApiTypeNative {
							cl = testhelpers.NewHTTPClientWithIdentitySessionToken(t, ctx, reg, user)
							f = testhelpers.InitializeLoginFlowViaAPI(t, cl, public, false, testhelpers.InitFlowWithAAL("aal2"), testhelpers.InitFlowWithVia("email"))
						} else {
							cl = testhelpers.NewHTTPClientWithIdentitySessionCookieLocalhost(t, ctx, reg, user)
							f = testhelpers.InitializeLoginFlowViaBrowser(t, cl, public, false, tc.apiType == ApiTypeSPA, false, false, testhelpers.InitFlowWithAAL("aal2"), testhelpers.InitFlowWithVia("email"))
						}

						body, err := json.Marshal(f)
						require.NoError(t, err)
						require.Len(t, gjson.GetBytes(body, "ui.nodes.#(group==code)").Array(), 1, "%s", body)
						require.Len(t, gjson.GetBytes(body, "ui.messages").Array(), 1, "%s", body)
						require.EqualValues(t, gjson.GetBytes(body, "ui.messages.0.id").Int(), text.InfoSelfServiceLoginMFA, "%s", body)

						s := &state{
							flowID:        f.GetId(),
							identity:      user,
							client:        cl,
							testServer:    public,
							identityEmail: gjson.Get(user.Traits.String(), "email").String(),
						}
						s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
							v.Set("identifier", s.identityEmail)
						}, false, nil)

						message := testhelpers.CourierExpectMessage(ctx, t, reg, s.identityEmail, "Use code")
						assert.Contains(t, message.Body, "Login to your account with the following code")
						loginCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
						assert.NotEmpty(t, loginCode)

						return submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
							v.Set("code", loginCode)
						}, true, nil), cl
					}

					t.Run("case=correct code credential without fallback works", func(t *testing.T) {
						testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/code.identity.schema.json") // has code identifier
						conf.MustSet(ctx, config.ViperKeyCodeConfigMissingCredentialFallbackEnabled, false)   // fallback enabled

						_, cl := run(t, true, nil, nil)
						testhelpers.EnsureAAL(t, cl, public, "aal2", "code")
					})

					t.Run("case=disabling mfa does not lock out the users", func(t *testing.T) {
						testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/code.identity.schema.json") // has code identifier

						s, cl := run(t, true, nil, nil)
						testhelpers.EnsureAAL(t, cl, public, "aal2", "code")

						email := gjson.GetBytes(s.identity.Traits, "email").String()
						s.identityEmail = email

						// We change now disable code mfa and enable passwordless instead.
						conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.mfa_enabled", false)
						conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.passwordless_enabled", true)

						t.Cleanup(func() {
							conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.passwordless_enabled", false)
							conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.mfa_enabled", true)
						})

						s = createLoginFlowWithIdentity(ctx, t, public, tc.apiType, s.identity)
						s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
							v.Set("identifier", email)
							v.Set("method", "code")
						}, false, nil)

						message := testhelpers.CourierExpectMessage(ctx, t, reg, email, "Use code")
						assert.Contains(t, message.Body, "Login to your account with the following code")
						loginCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
						assert.NotEmpty(t, loginCode)

						loginResult := submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
							v.Set("code", loginCode)
						}, true, nil)

						if tc.apiType == ApiTypeNative {
							assert.EqualValues(t, "aal1", gjson.Get(loginResult.body, "session.authenticator_assurance_level").String())
							assert.EqualValues(t, "code", gjson.Get(loginResult.body, "session.authentication_methods.#(method==code).method").String())
						} else {
							// The user should be able to sign in correctly even though, probably, the internal state was aal2 for available AAL.
							res, err := s.client.Get(public.URL + session.RouteWhoami)
							require.NoError(t, err)
							assert.EqualValues(t, http.StatusOK, res.StatusCode, loginResult.body)
							sess := x.MustReadAll(res.Body)
							require.NoError(t, res.Body.Close())

							assert.EqualValues(t, "aal1", gjson.GetBytes(sess, "authenticator_assurance_level").String())
							assert.EqualValues(t, "code", gjson.GetBytes(sess, "authentication_methods.#(method==code).method").String())
						}
					})

					t.Run("case=missing code credential with fallback works when identity schema has the code identifier set", func(t *testing.T) {
						testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/code.identity.schema.json") // has code identifier
						conf.MustSet(ctx, config.ViperKeyCodeConfigMissingCredentialFallbackEnabled, true)    // fallback enabled
						t.Cleanup(func() {
							conf.MustSet(ctx, config.ViperKeyCodeConfigMissingCredentialFallbackEnabled, false)
						})

						_, cl := run(t, false, nil, nil)
						testhelpers.EnsureAAL(t, cl, public, "aal2", "code")
					})

					t.Run("case=missing code credential with fallback works even when identity schema has no code identifier set", func(t *testing.T) {
						testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/no-code.schema.json")    // missing the code identifier
						conf.MustSet(ctx, config.ViperKeyCodeConfigMissingCredentialFallbackEnabled, true) // fallback enabled
						t.Cleanup(func() {
							testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/code.identity.schema.json")
							conf.MustSet(ctx, config.ViperKeyCodeConfigMissingCredentialFallbackEnabled, false)
						})

						_, cl := run(t, false, nil, nil)
						testhelpers.EnsureAAL(t, cl, public, "aal2", "code")
					})

					t.Run("case=missing code credential with fallback works even when identity schema has no code identifier set", func(t *testing.T) {
						testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/no-code-id.schema.json") // missing the code identifier
						conf.MustSet(ctx, config.ViperKeyCodeConfigMissingCredentialFallbackEnabled, true) // fallback enabled
						t.Cleanup(func() {
							testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/code.identity.schema.json")
							conf.MustSet(ctx, config.ViperKeyCodeConfigMissingCredentialFallbackEnabled, false)
						})

						_, cl := run(t, true, &identity.Credentials{}, map[identity.CredentialsType]identity.Credentials{})
						testhelpers.EnsureAAL(t, cl, public, "aal2", "code")
					})

					t.Run("case=legacy code credential with fallback works when identity schema has the code identifier not set", func(t *testing.T) {
						testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/no-code.schema.json")    // has code identifier
						conf.MustSet(ctx, config.ViperKeyCodeConfigMissingCredentialFallbackEnabled, true) // fallback enabled
						t.Cleanup(func() {
							testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/code.identity.schema.json") // has code identifier
							conf.MustSet(ctx, config.ViperKeyCodeConfigMissingCredentialFallbackEnabled, false)
						})

						_, cl := run(t, false, &identity.Credentials{Config: []byte(`{"via":""}`)}, nil)
						testhelpers.EnsureAAL(t, cl, public, "aal2", "code")
					})

					t.Run("case=legacy code credential with fallback works when identity schema has the code identifier not set", func(t *testing.T) {
						testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/no-code.schema.json")    // has code identifier
						conf.MustSet(ctx, config.ViperKeyCodeConfigMissingCredentialFallbackEnabled, true) // fallback enabled
						t.Cleanup(func() {
							testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/code.identity.schema.json") // has code identifier
							conf.MustSet(ctx, config.ViperKeyCodeConfigMissingCredentialFallbackEnabled, false)
						})

						for k, credentialsConfig := range []string{
							`{"address_type": "email                               ", "used_at": {"Time": "0001-01-01T00:00:00Z", "Valid": false}}`,
							`{"address_type": "email", "used_at": {"Time": "0001-01-01T00:00:00Z", "Valid": false}}`,
							`{"address_type": "", "used_at": {"Time": "0001-01-01T00:00:00Z", "Valid": false}}`,
							`{"address_type": ""}`,
							`{"address_type": "phone"}`,
							`{}`,
						} {
							t.Run(fmt.Sprintf("config=%d", k), func(t *testing.T) {
								_, cl := run(t, false, &identity.Credentials{Config: []byte(credentialsConfig)}, nil)
								testhelpers.EnsureAAL(t, cl, public, "aal2", "code")
							})
						}
					})
				})

				t.Run("case=without via parameter all options are shown", func(t *testing.T) {
					testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/code-mfa.identity.schema.json")
					conf.MustSet(ctx, config.ViperKeyCodeConfigMissingCredentialFallbackEnabled, false)
					t.Cleanup(func() {
						testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/code.identity.schema.json")
					})

					var cl *http.Client
					var f *oryClient.LoginFlow

					user := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
					user.NID = x.NewUUID()
					email1 := "code-mfa-1" + string(tc.apiType) + "@ory.sh"
					email2 := "code-mfa-2" + string(tc.apiType) + "@ory.sh"
					phone1 := 4917613213110
					if tc.apiType == ApiTypeNative {
						phone1 += 1
					} else if tc.apiType == ApiTypeSPA {
						phone1 += 2
					}
					user.Traits = identity.Traits(fmt.Sprintf(`{"email1":"%s","email2":"%s","phone1":"+%d"}`, email1, email2, phone1))
					require.NoError(t, reg.IdentityManager().Create(ctx, user))

					run := func(t *testing.T, identifierField string, identifier string) {
						if tc.apiType == ApiTypeNative {
							cl = testhelpers.NewHTTPClientWithIdentitySessionToken(t, ctx, reg, user)
							f = testhelpers.InitializeLoginFlowViaAPI(t, cl, public, false, testhelpers.InitFlowWithAAL("aal2"))
						} else {
							cl = testhelpers.NewHTTPClientWithIdentitySessionCookieLocalhost(t, ctx, reg, user)
							f = testhelpers.InitializeLoginFlowViaBrowser(t, cl, public, false, tc.apiType == ApiTypeSPA, false, false, testhelpers.InitFlowWithAAL("aal2"))
						}

						body, err := json.Marshal(f)
						require.NoError(t, err)

						snapshotx.SnapshotT(t, json.RawMessage(gjson.GetBytes(body, "ui.nodes.#(group==code)#").Raw))
						require.Len(t, gjson.GetBytes(body, "ui.messages").Array(), 1, "%s", body)
						require.EqualValues(t, gjson.GetBytes(body, "ui.messages.0.id").Int(), text.InfoSelfServiceLoginMFA, "%s", body)

						s := &state{
							flowID:        f.GetId(),
							identity:      user,
							client:        cl,
							testServer:    public,
							identityEmail: gjson.Get(user.Traits.String(), "email").String(),
						}

						s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
							v.Del("method")
							v.Set(identifierField, identifier)
						}, false, nil)

						var message *courier.Message
						if !strings.HasPrefix(identifier, "+") {
							// email
							message = testhelpers.CourierExpectMessage(ctx, t, reg, x.GracefulNormalization(identifier), "Use code")
							assert.Contains(t, message.Body, "Login to your account with the following code")
						} else {
							// SMS
							message = testhelpers.CourierExpectMessage(ctx, t, reg, x.GracefulNormalization(identifier), "Your login code is:")
						}
						loginCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
						assert.NotEmpty(t, loginCode)

						t.Logf("loginCode: %s", loginCode)

						s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
							v.Set("code", loginCode)
							v.Set(identifierField, identifier)
						}, true, nil)

						testhelpers.EnsureAAL(t, cl, public, "aal2", "code")
					}

					t.Run("field=identifier-email", func(t *testing.T) {
						run(t, "identifier", email1)
					})

					t.Run("field=address-email", func(t *testing.T) {
						run(t, "address", email2)
					})

					t.Run("field=address-phone", func(t *testing.T) {
						run(t, "address", fmt.Sprintf("+%d", phone1))
					})
				})

				t.Run("case=cannot use different identifier", func(t *testing.T) {
					identity := createIdentity(ctx, t, reg, false)
					var cl *http.Client
					var f *oryClient.LoginFlow
					if tc.apiType == ApiTypeNative {
						cl = testhelpers.NewHTTPClientWithIdentitySessionToken(t, ctx, reg, identity)
						f = testhelpers.InitializeLoginFlowViaAPI(t, cl, public, false, testhelpers.InitFlowWithAAL("aal2"), testhelpers.InitFlowWithVia("email"))
					} else {
						cl = testhelpers.NewHTTPClientWithIdentitySessionCookieLocalhost(t, ctx, reg, identity)
						f = testhelpers.InitializeLoginFlowViaBrowser(t, cl, public, false, tc.apiType == ApiTypeSPA, false, false, testhelpers.InitFlowWithAAL("aal2"), testhelpers.InitFlowWithVia("email"))
					}

					body, err := json.Marshal(f)
					require.NoError(t, err)
					require.Len(t, gjson.GetBytes(body, "ui.nodes.#(group==code)").Array(), 1)
					require.Len(t, gjson.GetBytes(body, "ui.messages").Array(), 1, "%s", body)
					require.EqualValues(t, gjson.GetBytes(body, "ui.messages.0.id").Int(), text.InfoSelfServiceLoginMFA, "%s", body)

					s := &state{
						flowID:        f.GetId(),
						identity:      identity,
						client:        cl,
						testServer:    public,
						identityEmail: gjson.Get(identity.Traits.String(), "email").String(),
					}
					email := testhelpers.RandomEmail()
					s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
						v.Set("identifier", email)
					}, false, nil)

					require.Equal(t, "This account does not exist or has not setup sign in with code.", gjson.Get(s.body, "ui.messages.0.text").String(), "%s", body)
				})

				t.Run("case=verify initial payload", func(t *testing.T) {
					fixedEmail := fmt.Sprintf("fixed_mfa_test_%s@ory.sh", tc.apiType)
					identity := createIdentity(ctx, t, reg, false, fixedEmail)
					var cl *http.Client
					var f *oryClient.LoginFlow
					if tc.apiType == ApiTypeNative {
						cl = testhelpers.NewHTTPClientWithIdentitySessionToken(t, ctx, reg, identity)
						f = testhelpers.InitializeLoginFlowViaAPI(t, cl, public, false, testhelpers.InitFlowWithAAL("aal2"), testhelpers.InitFlowWithVia("email_1"))
					} else {
						cl = testhelpers.NewHTTPClientWithIdentitySessionCookieLocalhost(t, ctx, reg, identity)
						f = testhelpers.InitializeLoginFlowViaBrowser(t, cl, public, false, tc.apiType == ApiTypeSPA, false, false, testhelpers.InitFlowWithAAL("aal2"), testhelpers.InitFlowWithVia("email_1"))
					}

					body, err := json.Marshal(f)
					require.NoError(t, err)
					snapshotx.SnapshotTJSON(t, body, snapshotx.ExceptPaths("ui.nodes.0.attributes.value", "id", "created_at", "expires_at", "updated_at", "issued_at", "request_url", "ui.action"))
				})

				t.Run("case=using a non existing identity trait results in an error", func(t *testing.T) {
					identity := createIdentity(ctx, t, reg, false)
					var cl *http.Client
					var res *http.Response
					var err error
					if tc.apiType == ApiTypeNative {
						cl = testhelpers.NewHTTPClientWithIdentitySessionToken(t, ctx, reg, identity)
						res, err = cl.Get(public.URL + "/self-service/login/api?aal=aal2&via=doesnt_exist")
					} else {
						cl = testhelpers.NewHTTPClientWithIdentitySessionCookieLocalhost(t, ctx, reg, identity)
						res, err = cl.Get(public.URL + "/self-service/login/browser?aal=aal2&via=doesnt_exist")
					}
					require.NoError(t, err)

					body := ioutilx.MustReadAll(res.Body)
					if tc.apiType == ApiTypeNative {
						body = []byte(gjson.GetBytes(body, "error").Raw)
					}
					require.Equal(t, "No value found for trait doesnt_exist in the current identity.", gjson.GetBytes(body, "reason").String(), "%s", body)
				})

				t.Run("case=unset trait in identity should lead to an error", func(t *testing.T) {
					identity := createIdentity(ctx, t, reg, false)
					var cl *http.Client
					var res *http.Response
					var err error
					if tc.apiType == ApiTypeNative {
						cl = testhelpers.NewHTTPClientWithIdentitySessionToken(t, ctx, reg, identity)
						res, err = cl.Get(public.URL + "/self-service/login/api?aal=aal2&via=email_1")
					} else {
						cl = testhelpers.NewHTTPClientWithIdentitySessionCookieLocalhost(t, ctx, reg, identity)
						res, err = cl.Get(public.URL + "/self-service/login/browser?aal=aal2&via=email_1")
					}
					require.NoError(t, err)

					body := ioutilx.MustReadAll(res.Body)
					if tc.apiType == ApiTypeNative {
						body = []byte(gjson.GetBytes(body, "error").Raw)
					}
					require.Equal(t, "No value found for trait email_1 in the current identity.", gjson.GetBytes(body, "reason").String(), "%s", body)
				})
			})
		})
	}
}

func TestFormHydration(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	ctx = configtesthelpers.WithConfigValue(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeCodeAuth), map[string]interface{}{
		"enabled":              true,
		"passwordless_enabled": true,
	})
	ctx = testhelpers.WithDefaultIdentitySchema(ctx, "file://./stub/code.identity.schema.json")

	s, err := reg.AllLoginStrategies().Strategy(identity.CredentialsTypeCodeAuth)
	require.NoError(t, err)
	fh, ok := s.(login.FormHydrator)
	require.True(t, ok)

	toSnapshot := func(t *testing.T, f *login.Flow) {
		t.Helper()
		// The CSRF token has a unique value that messes with the snapshot - ignore it.
		f.UI.Nodes.ResetNodes("csrf_token")
		snapshotx.SnapshotT(t, f.UI.Nodes)
	}
	newFlow := func(ctx context.Context, t *testing.T) (*http.Request, *login.Flow) {
		r := httptest.NewRequest("GET", "/self-service/login/browser", nil)
		r = r.WithContext(ctx)
		t.Helper()
		f, err := login.NewFlow(conf, time.Minute, "csrf_token", r, flow.TypeBrowser)
		require.NoError(t, err)
		return r, f
	}

	passwordlessEnabled := configtesthelpers.WithConfigValue(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeCodeAuth), map[string]interface{}{
		"enabled":              true,
		"passwordless_enabled": true,
		"mfa_enabled":          false,
	})

	mfaEnabled := configtesthelpers.WithConfigValue(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeCodeAuth), map[string]interface{}{
		"enabled":              true,
		"passwordless_enabled": false,
		"mfa_enabled":          true,
	})

	toMFARequest := func(t *testing.T, r *http.Request, f *login.Flow, traits string) {
		f.RequestedAAL = identity.AuthenticatorAssuranceLevel2
		r.URL = &url.URL{Path: "/", RawQuery: "via=email"}
		// I only fear god.
		r.Header = testhelpers.NewHTTPClientWithArbitrarySessionTokenAndTraits(t, ctx, reg, []byte(traits)).Transport.(*testhelpers.TransportWithHeader).GetHeader()
	}

	t.Run("method=PopulateLoginMethodFirstFactor", func(t *testing.T) {
		t.Run("case=code is used for 2fa but request is 1fa", func(t *testing.T) {
			r, f := newFlow(mfaEnabled, t)
			f.RequestedAAL = identity.AuthenticatorAssuranceLevel1
			require.NoError(t, fh.PopulateLoginMethodFirstFactor(r, f))
			toSnapshot(t, f)
		})

		t.Run("case=code is used for passwordless login and request is 1fa", func(t *testing.T) {
			r, f := newFlow(passwordlessEnabled, t)
			f.RequestedAAL = identity.AuthenticatorAssuranceLevel1
			require.NoError(t, fh.PopulateLoginMethodFirstFactor(r, f))
			toSnapshot(t, f)
		})
	})

	t.Run("method=PopulateLoginMethodFirstFactorRefresh", func(t *testing.T) {
		t.Run("case=code is used for passwordless login and request is 1fa with refresh", func(t *testing.T) {
			r, f := newFlow(passwordlessEnabled, t)
			f.RequestedAAL = identity.AuthenticatorAssuranceLevel1
			f.Refresh = true
			require.NoError(t, fh.PopulateLoginMethodFirstFactorRefresh(r, f))
			toSnapshot(t, f)
		})

		t.Run("case=code is used for 2fa and request is 1fa with refresh", func(t *testing.T) {
			r, f := newFlow(mfaEnabled, t)
			f.RequestedAAL = identity.AuthenticatorAssuranceLevel1
			f.Refresh = true
			require.NoError(t, fh.PopulateLoginMethodFirstFactorRefresh(r, f))
			toSnapshot(t, f)
		})
	})

	t.Run("method=PopulateLoginMethodSecondFactor", func(t *testing.T) {
		t.Run("using via", func(t *testing.T) {
			test := func(t *testing.T, ctx context.Context, email string) {
				r, f := newFlow(ctx, t)
				toMFARequest(t, r, f, `{"email":"`+email+`"}`)

				// We still use the legacy hydrator under the hood here and thus need to set this correctly.
				f.RequestedAAL = identity.AuthenticatorAssuranceLevel2
				r.URL = &url.URL{Path: "/", RawQuery: "via=email"}

				require.NoError(t, fh.PopulateLoginMethodSecondFactor(r, f))
				toSnapshot(t, f)
			}

			t.Run("case=code is used for 2fa", func(t *testing.T) {
				test(t, mfaEnabled, "PopulateLoginMethodSecondFactor-code-mfa-via-2fa@ory.sh")
			})

			t.Run("case=code is used for passwordless login", func(t *testing.T) {
				test(t, passwordlessEnabled, "PopulateLoginMethodSecondFactor-code-mfa-via-passwordless@ory.sh")
			})
		})

		t.Run("without via", func(t *testing.T) {
			test := func(t *testing.T, ctx context.Context, traits string) {
				r, f := newFlow(ctx, t)
				toMFARequest(t, r, f, traits)

				// We still use the legacy hydrator under the hood here and thus need to set this correctly.
				f.RequestedAAL = identity.AuthenticatorAssuranceLevel2
				r.URL = &url.URL{Path: "/"}

				require.NoError(t, fh.PopulateLoginMethodSecondFactor(r, f))
				toSnapshot(t, f)
			}

			t.Run("case=code is used for 2fa", func(t *testing.T) {
				ctx = testhelpers.WithDefaultIdentitySchema(mfaEnabled, "file://./stub/code-mfa.identity.schema.json")
				test(t, ctx, `{"email1":"PopulateLoginMethodSecondFactor-no-via-2fa-0@ory.sh","email2":"PopulateLoginMethodSecondFactor-no-via-2fa-1@ory.sh","phone1":"+4917655138291"}`)
			})

			t.Run("case=code is used for passwordless login", func(t *testing.T) {
				ctx = testhelpers.WithDefaultIdentitySchema(passwordlessEnabled, "file://./stub/code-mfa.identity.schema.json")
				test(t, ctx, `{"email1":"PopulateLoginMethodSecondFactor-no-via-passwordless-0@ory.sh","email2":"PopulateLoginMethodSecondFactor-no-via-passwordless-1@ory.sh","phone1":"+4917655138292"}`)
			})
		})

		t.Run("case=code is used for 2fa and request is 2fa", func(t *testing.T) {
			r, f := newFlow(mfaEnabled, t)
			toMFARequest(t, r, f, `{"email":"foo@ory.sh"}`)
			require.NoError(t, fh.PopulateLoginMethodSecondFactor(r, f))
			toSnapshot(t, f)
		})

		t.Run("case=code is used for passwordless login and request is 2fa", func(t *testing.T) {
			r, f := newFlow(passwordlessEnabled, t)
			toMFARequest(t, r, f, `{"email":"foo@ory.sh"}`)
			require.NoError(t, fh.PopulateLoginMethodSecondFactor(r, f))
			toSnapshot(t, f)
		})
	})

	t.Run("method=PopulateLoginMethodSecondFactorRefresh", func(t *testing.T) {
		t.Run("case=code is used for 2fa and request is 2fa with refresh", func(t *testing.T) {
			r, f := newFlow(mfaEnabled, t)
			toMFARequest(t, r, f, `{"email":"foo@ory.sh"}`)
			f.Refresh = true
			require.NoError(t, fh.PopulateLoginMethodSecondFactorRefresh(r, f))
			toSnapshot(t, f)
		})

		t.Run("case=code is used for passwordless login and request is 2fa with refresh", func(t *testing.T) {
			r, f := newFlow(passwordlessEnabled, t)
			toMFARequest(t, r, f, `{"email":"foo@ory.sh"}`)
			f.Refresh = true
			require.NoError(t, fh.PopulateLoginMethodSecondFactorRefresh(r, f))
			toSnapshot(t, f)
		})
	})

	t.Run("method=PopulateLoginMethodIdentifierFirstCredentials", func(t *testing.T) {
		t.Run("case=no options", func(t *testing.T) {
			t.Run("case=code is used for 2fa", func(t *testing.T) {
				r, f := newFlow(mfaEnabled, t)
				require.NoError(t, fh.PopulateLoginMethodFirstFactor(r, f))
				toSnapshot(t, f)
			})

			t.Run("case=code is used for passwordless login", func(t *testing.T) {
				r, f := newFlow(passwordlessEnabled, t)
				require.NoError(t, fh.PopulateLoginMethodFirstFactor(r, f))
				toSnapshot(t, f)
			})
		})

		t.Run("case=WithIdentityHint", func(t *testing.T) {
			t.Run("case=account enumeration mitigation enabled", func(t *testing.T) {
				t.Run("case=code is used for 2fa", func(t *testing.T) {
					r, f := newFlow(
						configtesthelpers.WithConfigValue(mfaEnabled, config.ViperKeySecurityAccountEnumerationMitigate, true),
						t,
					)
					require.ErrorIs(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f, login.WithIdentifier("foo@bar.com")), idfirst.ErrNoCredentialsFound)
					toSnapshot(t, f)
				})

				t.Run("case=code is used for passwordless login", func(t *testing.T) {
					r, f := newFlow(
						configtesthelpers.WithConfigValue(passwordlessEnabled, config.ViperKeySecurityAccountEnumerationMitigate, true),
						t,
					)
					require.NoError(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f, login.WithIdentifier("foo@bar.com")))
					toSnapshot(t, f)
				})
			})

			t.Run("case=account enumeration mitigation disabled", func(t *testing.T) {
				t.Run("case=with no identity", func(t *testing.T) {
					t.Run("case=code is used for 2fa", func(t *testing.T) {
						r, f := newFlow(mfaEnabled, t)
						require.ErrorIs(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f), idfirst.ErrNoCredentialsFound)
						toSnapshot(t, f)
					})

					t.Run("case=code is used for passwordless login", func(t *testing.T) {
						r, f := newFlow(passwordlessEnabled, t)
						require.ErrorIs(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f), idfirst.ErrNoCredentialsFound)
						toSnapshot(t, f)
					})
				})
				t.Run("case=identity has code method", func(t *testing.T) {
					identifier := x.NewUUID().String()
					id := createIdentity(ctx, t, reg, false, identifier)

					t.Run("case=code is used for 2fa", func(t *testing.T) {
						r, f := newFlow(mfaEnabled, t)
						require.ErrorIs(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f, login.WithIdentityHint(id)), idfirst.ErrNoCredentialsFound)
						toSnapshot(t, f)
					})

					t.Run("case=code is used for passwordless login", func(t *testing.T) {
						r, f := newFlow(passwordlessEnabled, t)
						require.NoError(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f, login.WithIdentityHint(id)))
						toSnapshot(t, f)
					})
				})

				t.Run("case=identity does not have a code method", func(t *testing.T) {
					id := identity.NewIdentity("default")

					t.Run("case=code is used for 2fa", func(t *testing.T) {
						r, f := newFlow(mfaEnabled, t)
						require.ErrorIs(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f, login.WithIdentityHint(id)), idfirst.ErrNoCredentialsFound)
						toSnapshot(t, f)
					})

					t.Run("case=code is used for passwordless login", func(t *testing.T) {
						r, f := newFlow(passwordlessEnabled, t)
						require.NoError(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f, login.WithIdentityHint(id)))
						toSnapshot(t, f)
					})
				})
			})
		})
	})

	t.Run("method=PopulateLoginMethodIdentifierFirstIdentification", func(t *testing.T) {
		r, f := newFlow(ctx, t)
		require.NoError(t, fh.PopulateLoginMethodIdentifierFirstIdentification(r, f))
		toSnapshot(t, f)
	})
}
