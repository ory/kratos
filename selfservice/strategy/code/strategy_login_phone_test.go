// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	oryClient "github.com/ory/kratos/internal/httpclient"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/x/ioutilx"
	"github.com/ory/x/sqlxx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ory/kratos/internal/testhelpers"
)

func TestLoginCodeStrategy_SMS(t *testing.T) {
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

	var externalVerifyResult string
	var externalVerifyRequestBody string
	initExternalSMSVerifier(t, ctx, conf, "file://./stub/request.config.login.jsonnet",
		&externalVerifyRequestBody, &externalVerifyResult)

	createIdentity := func(ctx context.Context, t *testing.T) *identity.Identity {
		t.Helper()
		i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		email := testhelpers.RandomEmail()
		phone := testhelpers.RandomPhone()

		i.Traits = identity.Traits(fmt.Sprintf(`{"tos": true, "email": "%s", "phone": "%s"}`, email, phone))

		i.Credentials[identity.CredentialsTypeCodeAuth] = identity.Credentials{Type: identity.CredentialsTypeCodeAuth, Identifiers: []string{phone}, Config: sqlxx.JSONRawMessage("{\"address_type\": \"phone\", \"used_at\": \"2023-07-26T16:59:06+02:00\"}")}

		var va []identity.VerifiableAddress
		va = append(va, identity.VerifiableAddress{
			Value:    phone,
			Via:      identity.AddressTypeSMS,
			Verified: true,
			Status:   identity.VerifiableAddressStatusCompleted,
		})

		i.VerifiableAddresses = va

		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(ctx, i))
		return i
	}

	type state struct {
		flowID     string
		client     *http.Client
		loginCode  string
		testServer *httptest.Server
		body       string
	}

	type ApiType string

	const (
		ApiTypeBrowser ApiType = "browser"
		ApiTypeSPA     ApiType = "spa"
		ApiTypeNative  ApiType = "api"
	)

	createLoginFlow := func(ctx context.Context, t *testing.T, public *httptest.Server, apiType ApiType) *state {
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
			client:     client,
			testServer: public,
		}
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
			require.EqualValues(t, http.StatusOK, resp.StatusCode, "%s", string(ioutilx.MustReadAll(resp.Body)))
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

	setNotifyUnknownRecipientsToTrue := func(t *testing.T) {
		conf.MustSet(ctx, fmt.Sprintf("%s.%s.notify_unknown_recipients", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypeCodeAuth.String()), true)
		t.Cleanup(func() {
			conf.MustSet(ctx, fmt.Sprintf("%s.%s.notify_unknown_recipients", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypeCodeAuth.String()), false)
		})
	}

	t.Run("test=notify_unknown_recipients false", func(t *testing.T) {
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
				t.Run("case=should be able to log in with code", func(t *testing.T) {
					externalVerifyResult = ""
					i := createIdentity(ctx, t)
					loginPhone := gjson.Get(i.Traits.String(), "phone").String()
					require.NotEmptyf(t, loginPhone, "could not find the phone trait inside the identity: %s", i.Traits.String())

					// create login flow
					s := createLoginFlow(ctx, t, public, tc.apiType)

					// submit phone
					s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
						v.Set("identifier", loginPhone)
					}, false, nil)

					assert.Contains(t, externalVerifyResult, "code has been sent")

					// 3. Submit OTP
					submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
						v.Set("code", "0000")
					}, true, nil)

					assert.Contains(t, externalVerifyResult, "code valid")
				})

				t.Run("case=should not be able to use valid code after 5 attempts", func(t *testing.T) {
					externalVerifyResult = ""
					i := createIdentity(ctx, t)
					loginPhone := gjson.Get(i.Traits.String(), "phone").String()
					require.NotEmptyf(t, loginPhone, "could not find the phone trait inside the identity: %s", i.Traits.String())
					s := createLoginFlow(ctx, t, public, tc.apiType)

					// submit email
					s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
						v.Set("identifier", loginPhone)
					}, false, nil)

					assert.Contains(t, externalVerifyResult, "code has been sent")

					for i := 0; i < 5; i++ {
						// 3. Submit OTP
						s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
							v.Set("code", "111111")
							v.Set("identifier", loginPhone)
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
						v.Set("code", "0000")
						v.Set("identifier", loginPhone)
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
			})
		}
	})

	t.Run("test=notify_unknown_recipients true", func(t *testing.T) {
		setNotifyUnknownRecipientsToTrue(t)
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
				t.Run("case=should be able to log in with code", func(t *testing.T) {
					externalVerifyResult = ""
					i := createIdentity(ctx, t)
					loginPhone := gjson.Get(i.Traits.String(), "phone").String()
					require.NotEmptyf(t, loginPhone, "could not find the phone trait inside the identity: %s", i.Traits.String())
					// create login flow
					s := createLoginFlow(ctx, t, public, tc.apiType)

					// submit phone
					s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
						v.Set("identifier", loginPhone)
					}, false, nil)

					assert.Contains(t, externalVerifyResult, "code has been sent")

					// 3. Submit OTP
					submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
						v.Set("code", "0000")
					}, true, nil)

					assert.Contains(t, externalVerifyResult, "code valid")
				})
				t.Run("case=respond with code sent but send info message instead", func(t *testing.T) {
					externalVerifyResult = ""
					// create login flow
					s := createLoginFlow(ctx, t, public, tc.apiType)

					// submit phone
					s = submitLogin(ctx, t, s, tc.apiType, func(v *url.Values) {
						v.Set("identifier", "+1234567890")
					}, false, nil)

					assert.Equal(t, externalVerifyResult, "")
					assert.Equal(t, int64(text.InfoSelfServiceLoginEmailWithCodeSent), gjson.GetBytes([]byte(s.body), "ui.messages.0.id").Int(), "%s", s.body)
					message := testhelpers.CourierExpectMessage(ctx, t, reg, "+1234567890", "")
					assert.Contains(t, message.Body, "we couldn’t find an account linked to this phone")
				})
			})
		}
	})
}
