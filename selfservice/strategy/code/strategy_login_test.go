// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/x/assertx"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

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

	createIdentity := func(ctx context.Context, t *testing.T, withoutCodeCredential bool, moreIdentifiers ...string) *identity.Identity {
		t.Helper()
		i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
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
			credentials[identity.CredentialsTypeCodeAuth] = identity.Credentials{Type: identity.CredentialsTypeCodeAuth, Identifiers: append([]string{email}, moreIdentifiers...), Config: sqlxx.JSONRawMessage("{\"address_type\": \"email\", \"used_at\": \"2023-07-26T16:59:06+02:00\"}")}
		}
		i.Credentials = credentials

		var va []identity.VerifiableAddress
		for _, identifier := range moreIdentifiers {
			va = append(va, identity.VerifiableAddress{Value: identifier, Verified: false, Status: identity.VerifiableAddressStatusCompleted})
		}

		va = append(va, identity.VerifiableAddress{Value: email, Verified: true, Status: identity.VerifiableAddressStatusCompleted})

		i.VerifiableAddresses = va

		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(ctx, i))
		return i
	}

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

	createLoginFlow := func(ctx context.Context, t *testing.T, public *httptest.Server, apiType ApiType, withoutCodeCredential bool, moreIdentifiers ...string) *state {
		t.Helper()

		identity := createIdentity(ctx, t, withoutCodeCredential, moreIdentifiers...)

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

		loginEmail := gjson.Get(identity.Traits.String(), "email").String()
		require.NotEmptyf(t, loginEmail, "could not find the email trait inside the identity: %s", identity.Traits.String())

		return &state{
			flowID:        clientInit.GetId(),
			identity:      identity,
			identityEmail: loginEmail,
			client:        client,
			testServer:    public,
		}
	}

	type onSubmitAssertion func(t *testing.T, s *state, body string, res *http.Response)

	submitLogin := func(ctx context.Context, t *testing.T, s *state, apiType ApiType, vals func(v *url.Values), mustHaveSession bool, submitAssertion onSubmitAssertion) *state {
		t.Helper()

		lf, resp, err := testhelpers.NewSDKCustomClient(s.testServer, s.client).FrontendApi.GetLoginFlow(ctx).Id(s.flowID).Execute()
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
			require.EqualValues(t, http.StatusOK, resp.StatusCode)
			body = string(ioutilx.MustReadAll(resp.Body))
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

				message := testhelpers.CourierExpectMessage(ctx, t, reg, s.identityEmail, "Login to your account")
				assert.Contains(t, message.Body, "please login to your account by entering the following code")

				loginCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
				assert.NotEmpty(t, loginCode)

				// 3. Submit OTP
				submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("code", loginCode)
				}, true, nil)
			})

			t.Run("case=should be able to log in with code", func(t *testing.T) {
				// create login flow
				s := createLoginFlow(ctx, t, public, tc.apiType, false)

				// submit email
				s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("identifier", s.identityEmail)
				}, false, nil)

				message := testhelpers.CourierExpectMessage(ctx, t, reg, s.identityEmail, "Login to your account")
				assert.Contains(t, message.Body, "please login to your account by entering the following code")

				loginCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
				assert.NotEmpty(t, loginCode)

				// 3. Submit OTP
				submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("code", loginCode)
				}, true, nil)
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

				message := testhelpers.CourierExpectMessage(ctx, t, reg, s.identityEmail, "Login to your account")
				assert.Contains(t, message.Body, "please login to your account by entering the following code")

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

				message := testhelpers.CourierExpectMessage(ctx, t, reg, s.identityEmail, "Login to your account")
				assert.Contains(t, message.Body, "please login to your account by entering the following code")

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

				message := testhelpers.CourierExpectMessage(ctx, t, reg, s.identityEmail, "Login to your account")
				assert.Contains(t, message.Body, "please login to your account by entering the following code")

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
						lf, resp, err := testhelpers.NewSDKCustomClient(public, s.client).FrontendApi.GetLoginFlow(ctx).Id(s.flowID).Execute()
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

						lf, resp, err := testhelpers.NewSDKCustomClient(public, s.client).FrontendApi.GetLoginFlow(ctx).Id(s.flowID).Execute()
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

				message := testhelpers.CourierExpectMessage(ctx, t, reg, s.identityEmail, "Login to your account")
				assert.Contains(t, message.Body, "please login to your account by entering the following code")
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

				message := testhelpers.CourierExpectMessage(ctx, t, reg, s.identityEmail, "Login to your account")
				assert.Contains(t, message.Body, "please login to your account by entering the following code")
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
						lf, _, err := testhelpers.NewSDKCustomClient(public, s.client).FrontendApi.GetLoginFlow(ctx).Id(resp.Request.URL.Query().Get("flow")).Execute()
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

				message := testhelpers.CourierExpectMessage(ctx, t, reg, s.identityEmail, "Login to your account")
				assert.Contains(t, message.Body, "please login to your account by entering the following code")
				loginCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
				assert.NotEmpty(t, loginCode)

				// resend code
				s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
					v.Set("resend", "code")
					v.Set("identifier", s.identityEmail)
				}, false, nil)

				message = testhelpers.CourierExpectMessage(ctx, t, reg, s.identityEmail, "Login to your account")
				assert.Contains(t, message.Body, "please login to your account by entering the following code")
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

				message := testhelpers.CourierExpectMessage(ctx, t, reg, loginEmail, "Login to your account")
				require.Contains(t, message.Body, "please login to your account by entering the following code")

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
					t.Cleanup(testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/default.schema.json")) // doesn't have the code credential
					identity := createIdentity(ctx, t, true)
					var cl *http.Client
					var f *oryClient.LoginFlow
					if tc.apiType == ApiTypeNative {
						cl = testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, identity)
						f = testhelpers.InitializeLoginFlowViaAPI(t, cl, public, false, testhelpers.InitFlowWithAAL("aal2"), testhelpers.InitFlowWithVia("email"))
					} else {
						cl = testhelpers.NewHTTPClientWithIdentitySessionCookieLocalhost(t, reg, identity)
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
					s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
						v.Set("identifier", s.identityEmail)
					}, true, nil)

					message := testhelpers.CourierExpectMessage(ctx, t, reg, s.identityEmail, "Login to your account")
					assert.Contains(t, message.Body, "please login to your account by entering the following code")
					loginCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
					assert.NotEmpty(t, loginCode)

					s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
						v.Set("code", loginCode)
					}, true, nil)

					testhelpers.EnsureAAL(t, cl, public, "aal2", "code")
				})
				t.Run("case=cannot use different identifier", func(t *testing.T) {
					identity := createIdentity(ctx, t, false)
					var cl *http.Client
					var f *oryClient.LoginFlow
					if tc.apiType == ApiTypeNative {
						cl = testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, identity)
						f = testhelpers.InitializeLoginFlowViaAPI(t, cl, public, false, testhelpers.InitFlowWithAAL("aal2"), testhelpers.InitFlowWithVia("email"))
					} else {
						cl = testhelpers.NewHTTPClientWithIdentitySessionCookieLocalhost(t, reg, identity)
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
					}, true, nil)

					require.Equal(t, "The address you entered does not match any known addresses in the current account.", gjson.Get(s.body, "ui.messages.0.text").String(), "%s", body)
				})

				t.Run("case=verify initial payload", func(t *testing.T) {
					fixedEmail := fmt.Sprintf("fixed_mfa_test_%s@ory.sh", tc.apiType)
					identity := createIdentity(ctx, t, false, fixedEmail)
					var cl *http.Client
					var f *oryClient.LoginFlow
					if tc.apiType == ApiTypeNative {
						cl = testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, identity)
						f = testhelpers.InitializeLoginFlowViaAPI(t, cl, public, false, testhelpers.InitFlowWithAAL("aal2"), testhelpers.InitFlowWithVia("email_1"))
					} else {
						cl = testhelpers.NewHTTPClientWithIdentitySessionCookieLocalhost(t, reg, identity)
						f = testhelpers.InitializeLoginFlowViaBrowser(t, cl, public, false, tc.apiType == ApiTypeSPA, false, false, testhelpers.InitFlowWithAAL("aal2"), testhelpers.InitFlowWithVia("email_1"))
					}

					body, err := json.Marshal(f)
					require.NoError(t, err)
					snapshotx.SnapshotTJSON(t, body, snapshotx.ExceptPaths("ui.nodes.0.attributes.value", "id", "created_at", "expires_at", "updated_at", "issued_at", "request_url", "ui.action"))
				})

				t.Run("case=using a non existing identity trait results in an error", func(t *testing.T) {
					identity := createIdentity(ctx, t, false)
					var cl *http.Client
					var res *http.Response
					var err error
					if tc.apiType == ApiTypeNative {
						cl = testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, identity)
						res, err = cl.Get(public.URL + "/self-service/login/api?aal=aal2&via=doesnt_exist")
					} else {
						cl = testhelpers.NewHTTPClientWithIdentitySessionCookieLocalhost(t, reg, identity)
						res, err = cl.Get(public.URL + "/self-service/login/browser?aal=aal2&via=doesnt_exist")
					}
					require.NoError(t, err)

					body := ioutilx.MustReadAll(res.Body)
					if tc.apiType == ApiTypeNative {
						body = []byte(gjson.GetBytes(body, "error").Raw)
					}
					require.Equal(t, "Trait does not exist in identity schema", gjson.GetBytes(body, "reason").String(), "%s", body)
				})

				t.Run("case=missing via parameter results results in an error", func(t *testing.T) {
					identity := createIdentity(ctx, t, false)
					var cl *http.Client
					var res *http.Response
					var err error
					if tc.apiType == ApiTypeNative {
						cl = testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, identity)
						res, err = cl.Get(public.URL + "/self-service/login/api?aal=aal2")
					} else {
						cl = testhelpers.NewHTTPClientWithIdentitySessionCookieLocalhost(t, reg, identity)
						res, err = cl.Get(public.URL + "/self-service/login/browser?aal=aal2")
					}
					require.NoError(t, err)

					body := ioutilx.MustReadAll(res.Body)
					if tc.apiType == ApiTypeNative {
						body = []byte(gjson.GetBytes(body, "error").Raw)
					}
					require.Equal(t, "AAL2 login via code requires the `via` query parameter", gjson.GetBytes(body, "reason").String(), "%s", body)
				})
				t.Run("case=unset trait in identity should lead to an error", func(t *testing.T) {
					identity := createIdentity(ctx, t, false)
					var cl *http.Client
					var res *http.Response
					var err error
					if tc.apiType == ApiTypeNative {
						cl = testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, identity)
						res, err = cl.Get(public.URL + "/self-service/login/api?aal=aal2&via=email_1")
					} else {
						cl = testhelpers.NewHTTPClientWithIdentitySessionCookieLocalhost(t, reg, identity)
						res, err = cl.Get(public.URL + "/self-service/login/browser?aal=aal2&via=email_1")
					}
					require.NoError(t, err)

					body := ioutilx.MustReadAll(res.Body)
					if tc.apiType == ApiTypeNative {
						body = []byte(gjson.GetBytes(body, "error").Raw)
					}
					require.Equal(t, "No value found for trait email_1 in the current identity", gjson.GetBytes(body, "reason").String(), "%s", body)
				})
			})
		})
	}
}

func TestLoginCodeStrategy_SMS(t *testing.T) {
	newReturnTs := func(t *testing.T, reg interface {
		session.ManagementProvider
		x.WriterProvider
		config.Provider
	}) *httptest.Server {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess, err := reg.SessionManager().FetchFromRequest(r.Context(), r)
			require.NoError(t, err)
			reg.Writer().Write(w, r, sess)
		}))
		t.Cleanup(ts.Close)
		reg.Config().MustSet(context.Background(), config.ViperKeySelfServiceBrowserDefaultReturnTo, ts.URL+"/return-ts")
		return ts
	}

	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, fmt.Sprintf("%s.%s.enabled", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypeCodeAuth.String()), true)
	conf.MustSet(ctx, fmt.Sprintf("%s.%s.passwordless_enabled", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypeCodeAuth.String()), true)
	conf.MustSet(ctx, fmt.Sprintf("%s.%s.enabled", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypePassword.String()), false)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/default.schema.json")
	conf.MustSet(ctx, config.ViperKeyCourierSMTPURL, "smtp://foo@bar@dev.null/")
	//conf.MustSet(ctx, config.CodeMaxAttempts, 5)
	//conf.MustSet(ctx, config.CodeLifespan, "1h")

	publicTS, _ := testhelpers.NewKratosServer(t, reg)
	redirTS := newReturnTs(t, reg)

	uiTS := testhelpers.NewLoginUIFlowEchoServer(t, reg)

	conf.MustSet(ctx, config.ViperKeySelfServiceLoginUI, uiTS.URL+"/login-ts")

	var expectValidationError = func(t *testing.T, isAPI, forced, isSPA bool, values func(url.Values)) string {
		return testhelpers.SubmitLoginForm(t, isAPI, nil, publicTS, values,
			isSPA, forced,
			testhelpers.ExpectStatusCode(isAPI || isSPA, http.StatusBadRequest, http.StatusOK),
			testhelpers.ExpectURL(isAPI || isSPA, publicTS.URL+login.RouteSubmitFlow, conf.SelfServiceFlowLoginUI(ctx).String()),
		)
	}

	createIdentity := func(identifier string) (error, *identity.Identity) {
		stateChangedAt := sqlxx.NullTime(time.Now())

		i := &identity.Identity{
			SchemaID:       "default",
			Traits:         identity.Traits(fmt.Sprintf(`{"phone":"%s"}`, identifier)),
			State:          identity.StateActive,
			StateChangedAt: &stateChangedAt}
		if err := reg.IdentityManager().Create(ctx, i); err != nil {
			return err, nil
		}

		return nil, i
	}

	getLoginNode := func(f *oryClient.LoginFlow, nodeName string) *oryClient.UiNode {
		for _, n := range f.Ui.Nodes {
			if n.Attributes.UiNodeInputAttributes.Name == nodeName {
				return &n
			}
		}
		return nil
	}

	var loginWithPhone = func(
		t *testing.T, isAPI, refresh, isSPA bool,
		expectedStatusCode int, expectedURL string,
		values func(url.Values),
	) string {
		f := testhelpers.InitializeLoginFlow(t, isAPI, nil, publicTS, false, false)

		assert.Empty(t, getLoginNode(f, "code"))
		assert.NotEmpty(t, getLoginNode(f, "identifier"))

		body := testhelpers.SubmitLoginFormWithFlow(t, isAPI, nil, values,
			false, testhelpers.ExpectStatusCode(isAPI, http.StatusBadRequest, http.StatusOK),
			testhelpers.ExpectURL(isAPI || isSPA, publicTS.URL+login.RouteSubmitFlow, conf.SelfServiceFlowLoginUI(ctx).String()),
			f)

		assertx.EqualAsJSON(t,
			text.NewLoginEmailWithCodeSent(),
			json.RawMessage(gjson.Get(body, "ui.messages.0").Raw),
			"%s", body,
		)
		assert.Equal(t, flow.StateSMSSent.String(), gjson.Get(body, "state").String(),
			"%s", testhelpers.PrettyJSON(t, []byte(body)))
		assert.Equal(t, "code", gjson.Get(body, "active").String(), "%s", body)
		assert.NotEmpty(t, gjson.Get(body, "ui.nodes.#(attributes.name==code)"), "%s", body)
		assert.Empty(t, gjson.Get(body, "ui.nodes.#(attributes.name==code).attirbutes.value"), "%s", body)

		st := gjson.Get(body, "session_token").String()
		assert.Empty(t, st, "Response body: %s", body) //No session token as we have not presented the code yet

		values = func(v url.Values) {
			v.Del("resend")
			if isAPI {
				v.Set("method", "code")
			}
			v.Set("code", "0000")
		}

		publicClient := testhelpers.NewSDKCustomClient(publicTS, &http.Client{})
		f, _, err := publicClient.FrontendApi.GetLoginFlow(ctx).Id(f.Id).Execute()
		assert.NoError(t, err)
		body = testhelpers.SubmitLoginFormWithFlow(t, isAPI, nil, values, false, expectedStatusCode, expectedURL, f)

		return body
	}

	t.Run("should return an error because no phone is set", func(t *testing.T) {
		var check = func(t *testing.T, body string) {
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.action").String(), publicTS.URL+login.RouteSubmitFlow, "%s", body)

			assert.Equal(t, "Property identifier is missing.",
				gjson.Get(body, "ui.nodes.#(attributes.name==identifier).messages.0.text").String(), "%s", body)

			// The code value should not be returned!
			assert.Empty(t, gjson.Get(body, "ui.nodes.#(attributes.name==code).attributes.value").String())
		}

		t.Run("type=api", func(t *testing.T) {
			var values = func(v url.Values) {
				v.Set("method", "code")
				v.Del("identifier")
			}

			check(t, expectValidationError(t, true, false, false, values))
		})
	})

	t.Run("should not send code as user was not registered", func(t *testing.T) {
		var check = func(t *testing.T, isAPI bool, values func(url.Values)) {
			f := testhelpers.InitializeLoginFlow(t, isAPI, nil, publicTS, false, false)
			body := testhelpers.SubmitLoginFormWithFlow(t, isAPI, nil, values,
				false, testhelpers.ExpectStatusCode(isAPI, http.StatusBadRequest, http.StatusOK),
				testhelpers.ExpectURL(isAPI, publicTS.URL+login.RouteSubmitFlow, conf.SelfServiceFlowLoginUI(ctx).String()),
				f)

			assertx.EqualAsJSON(t,
				text.NewErrorValidationNoCodeUser(),
				json.RawMessage(gjson.Get(body, "ui.messages.0").Raw),
				"%s", body,
			)
		}

		t.Run("type=api", func(t *testing.T) {
			var values = func(v url.Values) {
				v.Set("method", "code")
				v.Set("identifier", "+99999999999")
			}

			check(t, true, values)
		})
		t.Run("type=browser", func(t *testing.T) {
			var values = func(v url.Values) {
				v.Set("identifier", "+99999999999")
			}

			check(t, false, values)
		})
	})

	t.Run("should pass with registered user", func(t *testing.T) {
		identifier := fmt.Sprintf("+452%s", fmt.Sprint(rand.Int())[0:7])
		conf.MustSet(ctx, config.CodeTestNumbers, []string{identifier})
		err, createdIdentity := createIdentity(identifier)
		assert.NoError(t, err)

		var values = func(v url.Values) {
			v.Set("method", "code")
			v.Set("identifier", identifier)
		}

		t.Run("type=api", func(t *testing.T) {
			body := loginWithPhone(t, true, false, false, http.StatusOK, publicTS.URL+login.RouteSubmitFlow, values)
			assert.Equal(t, identifier, gjson.Get(body, "session.identity.traits.phone").String(), "%s", body)
			assert.NotEmpty(t, gjson.Get(body, "session_token").String(), "%s", body)
			i, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, createdIdentity.ID)
			assert.NoError(t, err)
			assert.NotEmpty(t, i.VerifiableAddresses, "%s", body)
			assert.Equal(t, identifier, i.VerifiableAddresses[0].Value)
			assert.True(t, i.VerifiableAddresses[0].Verified)

			assert.Equal(t, identifier, gjson.Get(body, "session.identity.verifiable_addresses.0.value").String())
			assert.Equal(t, "true", gjson.Get(body, "session.identity.verifiable_addresses.0.verified").String(), "%s", body)
		})
		t.Run("type=browser", func(t *testing.T) {
			body := loginWithPhone(t, false, false, false, http.StatusOK, redirTS.URL, values)
			assert.Equal(t, identifier, gjson.Get(body, "identity.traits.phone").String(), "%s", body)
			assert.True(t, gjson.Get(body, "active").Bool(), "%s", body)
			i, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, createdIdentity.ID)
			assert.NoError(t, err)
			assert.NotEmpty(t, i.VerifiableAddresses, "%s", body)
			assert.Equal(t, identifier, i.VerifiableAddresses[0].Value)
			assert.True(t, i.VerifiableAddresses[0].Verified)

			assert.Equal(t, identifier, gjson.Get(body, "identity.verifiable_addresses.0.value").String())
			assert.Equal(t, "true", gjson.Get(body, "identity.verifiable_addresses.0.verified").String())
		})
	})

	t.Run("should save transient payload to SMS template data", func(t *testing.T) {
		identifier := fmt.Sprintf("+452%s", fmt.Sprint(rand.Int())[0:7])
		err, _ := createIdentity(identifier)
		assert.NoError(t, err)

		var values = func(v url.Values) {
			v.Set("method", "code")
			v.Set("identifier", identifier)
			v.Set("transient_payload", `{"branding": "brand-1"}`)
		}

		var doTest = func(t *testing.T, isAPI bool) {
			f := testhelpers.InitializeLoginFlow(t, isAPI, nil, publicTS, false, false)
			testhelpers.SubmitLoginFormWithFlow(t, isAPI, nil, values,
				false, testhelpers.ExpectStatusCode(isAPI, http.StatusBadRequest, http.StatusOK),
				testhelpers.ExpectURL(isAPI, publicTS.URL+login.RouteSubmitFlow, conf.SelfServiceFlowLoginUI(ctx).String()),
				f)

			message := testhelpers.CourierExpectMessage(ctx, t, reg, identifier, "")
			assert.Equal(t, "brand-1", gjson.GetBytes(message.TemplateData, "transient_payload.branding").String(), "%s", message.TemplateData)
		}

		t.Run("type=browser", func(t *testing.T) {
			doTest(t, false)
		})

		t.Run("type=api", func(t *testing.T) {
			doTest(t, true)
		})
	})

}
