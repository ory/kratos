// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ory/kratos/ui/node"
	"github.com/ory/x/assertx"
	"github.com/ory/x/snapshotx"

	"github.com/ory/kratos/selfservice/flow"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	oryClient "github.com/ory/kratos/internal/httpclient"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/strategy/code"
)

type state struct {
	flowID         string
	client         *http.Client
	email          string
	testServer     *httptest.Server
	resultIdentity *identity.Identity
	body           string
}

func TestRegistrationCodeStrategyDisabled(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/code.identity.schema.json")
	conf.MustSet(ctx, fmt.Sprintf("%s.%s.enabled", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypePassword.String()), false)
	conf.MustSet(ctx, fmt.Sprintf("%s.%s.enabled", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypeCodeAuth.String()), false)
	conf.MustSet(ctx, fmt.Sprintf("%s.%s.passwordless_enabled", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypeCodeAuth), false)
	conf.MustSet(ctx, "selfservice.flows.registration.enable_legacy_one_step", true)

	_ = testhelpers.NewRegistrationUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)

	public, _, _, _ := testhelpers.NewKratosServerWithCSRFAndRouters(t, reg)

	client := testhelpers.NewClientWithCookies(t)
	resp, err := client.Get(public.URL + registration.RouteInitBrowserFlow)
	require.NoError(t, err)
	require.EqualValues(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Falsef(t, gjson.GetBytes(body, "ui.nodes.#(attributes.value==code)").Exists(), "%s", body)

	// attempt to still submit the code form even though it doesn't exist

	payload := strings.NewReader(url.Values{
		"csrf_token":   {gjson.GetBytes(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String()},
		"method":       {"code"},
		"traits.email": {testhelpers.RandomEmail()},
	}.Encode())
	req, err := http.NewRequestWithContext(ctx, "POST", public.URL+registration.RouteSubmitFlow+"?flow="+gjson.GetBytes(body, "id").String(), payload)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "This endpoint was disabled by system administrator. Please check your url or contact the system administrator to enable it.", gjson.GetBytes(body, "error.reason").String())
}

func TestRegistrationCodeStrategy(t *testing.T) {
	type ApiType string

	const (
		ApiTypeBrowser ApiType = "browser"
		ApiTypeSPA     ApiType = "spa"
		ApiTypeNative  ApiType = "api"
	)

	setup := func(ctx context.Context, t *testing.T) (*config.Config, *driver.RegistryDefault, *httptest.Server) {
		conf, reg := internal.NewFastRegistryWithMocks(t)
		testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/code.identity.schema.json")
		conf.MustSet(ctx, fmt.Sprintf("%s.%s.enabled", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypePassword.String()), false)
		conf.MustSet(ctx, fmt.Sprintf("%s.%s.enabled", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypeCodeAuth.String()), true)
		conf.MustSet(ctx, fmt.Sprintf("%s.%s.passwordless_enabled", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypeCodeAuth), true)
		conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh")
		conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{"https://www.ory.sh"})
		conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationAfter+".code.hooks", []map[string]interface{}{
			{"hook": "session"},
		})
		conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationEnableLegacyOneStep, true)

		_ = testhelpers.NewRegistrationUIFlowEchoServer(t, reg)
		_ = testhelpers.NewErrorTestServer(t, reg)

		public, _, _, _ := testhelpers.NewKratosServerWithCSRFAndRouters(t, reg)

		return conf, reg, public
	}

	createRegistrationFlow := func(ctx context.Context, t *testing.T, public *httptest.Server, apiType ApiType) *state {
		t.Helper()

		var client *http.Client

		if apiType == ApiTypeNative {
			client = &http.Client{}
		} else {
			client = testhelpers.NewClientWithCookies(t)
		}

		client.Transport = testhelpers.NewTransportWithLogger(http.DefaultTransport, t).RoundTripper

		var clientInit *oryClient.RegistrationFlow
		if apiType == ApiTypeNative {
			clientInit = testhelpers.InitializeRegistrationFlowViaAPI(t, client, public)
		} else {
			clientInit = testhelpers.InitializeRegistrationFlowViaBrowser(t, client, public, apiType == ApiTypeSPA, false, false)
		}

		body, err := json.Marshal(clientInit)
		require.NoError(t, err)

		csrfToken := gjson.GetBytes(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String()
		if apiType == ApiTypeNative {
			require.Emptyf(t, csrfToken, "expected an empty value for csrf_token on native api flows but got %s", body)
		} else {
			require.NotEmpty(t, csrfToken)
		}

		require.Truef(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==traits.email)").Exists(), "%s", body)
		require.Truef(t, gjson.GetBytes(body, "ui.nodes.#(attributes.value==code)").Exists(), "%s", body)

		return &state{
			client:     client,
			flowID:     clientInit.GetId(),
			testServer: public,
		}
	}

	type onSubmitAssertion func(ctx context.Context, t *testing.T, s *state, body string, resp *http.Response)

	registerNewUser := func(ctx context.Context, t *testing.T, s *state, apiType ApiType, submitAssertion onSubmitAssertion) *state {
		t.Helper()

		if s.email == "" {
			s.email = testhelpers.RandomEmail()
		}

		rf, resp, err := testhelpers.NewSDKCustomClient(s.testServer, s.client).FrontendAPI.GetRegistrationFlow(context.Background()).Id(s.flowID).Execute()
		require.NoError(t, err)
		require.EqualValues(t, http.StatusOK, resp.StatusCode)

		values := testhelpers.SDKFormFieldsToURLValues(rf.Ui.Nodes)
		values.Set("traits.email", s.email)
		values.Set("traits.tos", "1")
		values.Set("method", "code")

		body, resp := testhelpers.RegistrationMakeRequest(t, apiType == ApiTypeNative, apiType == ApiTypeSPA, rf, s.client, testhelpers.EncodeFormAsJSON(t, apiType == ApiTypeNative, values))
		s.body = body

		if submitAssertion != nil {
			submitAssertion(ctx, t, s, body, resp)
			return s
		}
		t.Logf("%v", body)

		if apiType == ApiTypeBrowser {
			require.EqualValues(t, http.StatusOK, resp.StatusCode)
		} else {
			require.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
		}

		csrfToken := gjson.Get(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String()
		if apiType == ApiTypeNative {
			assert.Emptyf(t, csrfToken, "expected an empty value for csrf_token on native api flows but got %s", body)
		} else {
			assert.NotEmptyf(t, csrfToken, "%s", body)
		}
		require.Equal(t, s.email, gjson.Get(body, "ui.nodes.#(attributes.name==traits.email).attributes.value").String())

		return s
	}

	submitOTP := func(ctx context.Context, t *testing.T, reg *driver.RegistryDefault, s *state, vals func(v *url.Values), apiType ApiType, submitAssertion onSubmitAssertion) *state {
		t.Helper()

		rf, resp, err := testhelpers.NewSDKCustomClient(s.testServer, s.client).FrontendAPI.GetRegistrationFlow(context.Background()).Id(s.flowID).Execute()
		require.NoError(t, err)
		require.EqualValues(t, http.StatusOK, resp.StatusCode)

		values := testhelpers.SDKFormFieldsToURLValues(rf.Ui.Nodes)
		// the sdk to values always adds resend which isn't what we always need here.
		// so we delete it here.
		// the custom vals func can add it again if needed.
		values.Del("resend")
		values.Set("traits.email", s.email)
		values.Set("traits.tos", "1")
		vals(&values)

		body, resp := testhelpers.RegistrationMakeRequest(t, apiType == ApiTypeNative, apiType == ApiTypeSPA, rf, s.client, testhelpers.EncodeFormAsJSON(t, apiType == ApiTypeNative, values))
		s.body = body

		if submitAssertion != nil {
			submitAssertion(ctx, t, s, body, resp)
			return s
		}

		require.Equal(t, http.StatusOK, resp.StatusCode, body)

		verifiableAddress, err := reg.PrivilegedIdentityPool().FindVerifiableAddressByValue(ctx, identity.VerifiableAddressTypeEmail, s.email)
		require.NoError(t, err)
		require.Equal(t, strings.ToLower(s.email), verifiableAddress.Value)

		id, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, verifiableAddress.IdentityID)
		require.NoError(t, err)
		require.NotNil(t, id.ID)

		_, ok := id.GetCredentials(identity.CredentialsTypeCodeAuth)
		require.True(t, ok)

		s.resultIdentity = id
		return s
	}

	t.Run("test=different flows on the same configurations", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		conf, reg, public := setup(ctx, t)

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
			t.Run("flow="+tc.d, func(t *testing.T) {
				t.Run("case=should be able to register with code identity credentials", func(t *testing.T) {
					ctx := context.Background()

					// 1. Initiate flow
					state := createRegistrationFlow(ctx, t, public, tc.apiType)

					// 2. Submit Identifier (email)
					state = registerNewUser(ctx, t, state, tc.apiType, nil)

					message := testhelpers.CourierExpectMessage(ctx, t, reg, state.email, "Use code")
					assert.Contains(t, message.Body, "Complete your account registration with the following code")

					registrationCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
					assert.NotEmpty(t, registrationCode)

					// 3. Submit OTP
					state = submitOTP(ctx, t, reg, state, func(v *url.Values) {
						v.Set("code", registrationCode)
					}, tc.apiType, nil)

					if tc.apiType == ApiTypeSPA {
						assert.EqualValues(t, flow.ContinueWithActionRedirectBrowserToString, gjson.Get(state.body, "continue_with.0.action").String(), "%s", state.body)
						assert.Contains(t, gjson.Get(state.body, "continue_with.0.redirect_browser_to").String(), conf.SelfServiceBrowserDefaultReturnTo(ctx).String(), "%s", state.body)
					} else if tc.apiType == ApiTypeSPA {
						assert.Empty(t, gjson.Get(state.body, "continue_with").Array(), "%s", state.body)
					} else if tc.apiType == ApiTypeNative {
						assert.NotContains(t, gjson.Get(state.body, "continue_with").Raw, string(flow.ContinueWithActionRedirectBrowserToString), "%s", state.body)
					}
				})

				t.Run("case=should normalize email address on sign up", func(t *testing.T) {
					ctx := context.Background()

					// 1. Initiate flow
					state := createRegistrationFlow(ctx, t, public, tc.apiType)
					sourceMail := testhelpers.RandomEmail()
					state.email = strings.ToUpper(sourceMail)
					assert.NotEqual(t, sourceMail, state.email)

					// 2. Submit Identifier (email)
					state = registerNewUser(ctx, t, state, tc.apiType, nil)

					message := testhelpers.CourierExpectMessage(ctx, t, reg, sourceMail, "Use code")
					assert.Contains(t, message.Body, "Complete your account registration with the following code")

					registrationCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
					assert.NotEmpty(t, registrationCode)

					// 3. Submit OTP
					state = submitOTP(ctx, t, reg, state, func(v *url.Values) {
						v.Set("code", registrationCode)
					}, tc.apiType, nil)

					creds, ok := state.resultIdentity.GetCredentials(identity.CredentialsTypeCodeAuth)
					require.True(t, ok)
					require.Len(t, creds.Identifiers, 1)
					assert.Equal(t, sourceMail, creds.Identifiers[0])
				})

				t.Run("case=should be able to resend the code", func(t *testing.T) {
					ctx := context.Background()

					s := createRegistrationFlow(ctx, t, public, tc.apiType)

					s = registerNewUser(ctx, t, s, tc.apiType, func(ctx context.Context, t *testing.T, s *state, body string, resp *http.Response) {
						if tc.apiType == ApiTypeBrowser {
							require.EqualValues(t, http.StatusOK, resp.StatusCode)
						} else {
							require.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
						}

						csrfToken := gjson.Get(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String()
						if tc.apiType == ApiTypeNative {
							require.Empty(t, csrfToken, "expected the csrf_token to be empty but got %s", body)
						} else {
							require.NotEmptyf(t, csrfToken, "expected the csrf_token to exist but got %s", body)
						}
						require.Equal(t, s.email, gjson.Get(body, "ui.nodes.#(attributes.name==traits.email).attributes.value").String())

						attr := gjson.Get(body, "ui.nodes.#(attributes.name==method)#").String()
						require.NotEmpty(t, attr)

						val := gjson.Get(attr, "#(attributes.type==hidden).attributes.value").String()
						require.Equal(t, "code", val, body)
					})

					message := testhelpers.CourierExpectMessage(ctx, t, reg, s.email, "Use code")
					assert.Contains(t, message.Body, "Complete your account registration with the following code")

					registrationCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
					assert.NotEmpty(t, registrationCode)

					// resend code
					s = submitOTP(ctx, t, reg, s, func(v *url.Values) {
						v.Set("resend", "code")
					}, tc.apiType, func(ctx context.Context, t *testing.T, s *state, body string, resp *http.Response) {
						if tc.apiType == ApiTypeBrowser {
							require.Equal(t, http.StatusOK, resp.StatusCode)
						} else {
							require.EqualValues(t, http.StatusBadRequest, resp.StatusCode)
						}
						csrfToken := gjson.Get(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String()
						if tc.apiType == ApiTypeNative {
							assert.Emptyf(t, csrfToken, "expected an empty value for csrf_token on native api flows but got %s", body)
						} else {
							require.NotEmptyf(t, csrfToken, "expected to find the csrf_token but got %s", body)
						}
						require.Containsf(t, gjson.Get(body, "ui.messages").String(), "A code has been sent to the address(es) you provided", "%s", body)
					})

					// get the new code from email
					message = testhelpers.CourierExpectMessage(ctx, t, reg, s.email, "Use code")
					assert.Contains(t, message.Body, "Complete your account registration with the following code")

					registrationCode2 := testhelpers.CourierExpectCodeInMessage(t, message, 1)
					assert.NotEmpty(t, registrationCode2)

					require.NotEqual(t, registrationCode, registrationCode2)

					// try submit old code
					s = submitOTP(ctx, t, reg, s, func(v *url.Values) {
						v.Set("code", registrationCode)
					}, tc.apiType, func(ctx context.Context, t *testing.T, s *state, body string, resp *http.Response) {
						if tc.apiType == ApiTypeBrowser {
							require.Equal(t, http.StatusOK, resp.StatusCode, "%s", body)
						} else {
							require.Equal(t, http.StatusBadRequest, resp.StatusCode, "%s", body)
						}
						require.Contains(t, gjson.Get(body, "ui.messages").String(), "The registration code is invalid or has already been used. Please try again")
					})

					s = submitOTP(ctx, t, reg, s, func(v *url.Values) {
						v.Set("code", registrationCode2)
					}, tc.apiType, nil)
				})

				t.Run("case=swapping out traits should not be possible on code submit", func(t *testing.T) {
					ctx := context.Background()

					// 1. Initiate flow
					s := createRegistrationFlow(ctx, t, public, tc.apiType)

					// 2. Submit Identifier (email)
					s = registerNewUser(ctx, t, s, tc.apiType, nil)

					message := testhelpers.CourierExpectMessage(ctx, t, reg, s.email, "Use code")
					assert.Contains(t, message.Body, "Complete your account registration with the following code")

					registrationCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
					assert.NotEmpty(t, registrationCode)

					s.email = "not-" + s.email // swap out email
					// 3. Submit OTP
					s = submitOTP(ctx, t, reg, s, func(v *url.Values) {
						v.Set("code", registrationCode)
					}, tc.apiType, func(ctx context.Context, t *testing.T, s *state, body string, resp *http.Response) {
						if tc.apiType == ApiTypeBrowser {
							require.Equal(t, http.StatusOK, resp.StatusCode, "%s", body)
						} else {
							require.Equal(t, http.StatusBadRequest, resp.StatusCode, "%s", body)
						}
						require.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "The provided traits do not match the traits previously associated with this flow.")
					})
				})

				t.Run("case=swapping out traits that aren't strings should not be possible on code submit", func(t *testing.T) {
					ctx := context.Background()

					// 1. Initiate flow
					s := createRegistrationFlow(ctx, t, public, tc.apiType)

					// 2. Submit Identifier (email)
					s = registerNewUser(ctx, t, s, tc.apiType, nil)

					message := testhelpers.CourierExpectMessage(ctx, t, reg, s.email, "Use code")
					assert.Contains(t, message.Body, "Complete your account registration")

					registrationCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
					assert.NotEmpty(t, registrationCode)

					// 3. Submit OTP
					s = submitOTP(ctx, t, reg, s, func(v *url.Values) {
						v.Set("code", registrationCode)
						v.Set("traits.tos", "0")
					}, tc.apiType, func(ctx context.Context, t *testing.T, s *state, body string, resp *http.Response) {
						if tc.apiType == ApiTypeBrowser {
							require.Equal(t, http.StatusOK, resp.StatusCode, "%s", body)
						} else {
							require.Equal(t, http.StatusBadRequest, resp.StatusCode, "%s", body)
						}
						require.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "The provided traits do not match the traits previously associated with this flow.")
					})
				})

				t.Run("case=code should not be able to use more than 5 times", func(t *testing.T) {
					ctx := context.Background()

					// 1. Initiate flow
					s := createRegistrationFlow(ctx, t, public, tc.apiType)

					// 2. Submit Identifier (email)
					s = registerNewUser(ctx, t, s, tc.apiType, nil)

					reg.Persister().Transaction(ctx, func(ctx context.Context, connection *pop.Connection) error {
						count, err := connection.RawQuery(fmt.Sprintf("SELECT * FROM %s WHERE selfservice_registration_flow_id = ?", new(code.RegistrationCode).TableName(ctx)), uuid.FromStringOrNil(s.flowID)).Count(new(code.RegistrationCode))
						require.NoError(t, err)
						require.Equal(t, 1, count)
						return nil
					})

					for i := 0; i < 5; i++ {
						s = submitOTP(ctx, t, reg, s, func(v *url.Values) {
							v.Set("code", "111111")
						}, tc.apiType, func(ctx context.Context, t *testing.T, s *state, body string, resp *http.Response) {
							if tc.apiType == ApiTypeBrowser {
								require.Equal(t, http.StatusOK, resp.StatusCode, "%s", body)
							} else {
								require.Equal(t, http.StatusBadRequest, resp.StatusCode, "%s", body)
							}
							require.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "The registration code is invalid or has already been used")
						})
					}

					s = submitOTP(ctx, t, reg, s, func(v *url.Values) {
						v.Set("code", "111111")
					}, tc.apiType, func(ctx context.Context, t *testing.T, s *state, body string, resp *http.Response) {
						if tc.apiType == ApiTypeBrowser {
							require.Equal(t, http.StatusOK, resp.StatusCode, "%s", body)
						} else {
							require.Equal(t, http.StatusBadRequest, resp.StatusCode, "%s", body)
						}
						require.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "The request was submitted too often.")
					})
				})
			})
		}
	})

	t.Run("test=cases with different configs", func(t *testing.T) {
		ctx := context.Background()
		conf, reg, public := setup(ctx, t)

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
				t.Run("case=should fail when schema does not contain the `code` extension", func(t *testing.T) {
					testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/no-code.schema.json")
					conf.MustSet(ctx, config.ViperKeyCodeConfigMissingCredentialFallbackEnabled, false)

					t.Cleanup(func() {
						conf.MustSet(ctx, config.ViperKeyCodeConfigMissingCredentialFallbackEnabled, true)
						testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/code.identity.schema.json")
					})

					// 1. Initiate flow
					s := createRegistrationFlow(ctx, t, public, tc.apiType)

					// 2. Submit Identifier (email)
					s = registerNewUser(ctx, t, s, tc.apiType, func(ctx context.Context, t *testing.T, s *state, body string, resp *http.Response) {
						if tc.apiType == ApiTypeBrowser {
							// we expect a redirect to the registration page with the flow id
							require.Equal(t, http.StatusOK, resp.StatusCode)
							require.Equal(t, conf.SelfServiceFlowRegistrationUI(ctx).Path, resp.Request.URL.Path)
							rf, resp, err := testhelpers.NewSDKCustomClient(public, s.client).FrontendAPI.GetRegistrationFlow(ctx).Id(resp.Request.URL.Query().Get("flow")).Execute()
							require.NoError(t, err)
							require.Equal(t, http.StatusOK, resp.StatusCode)
							body, err := json.Marshal(rf)
							require.NoError(t, err)
							require.Contains(t, gjson.GetBytes(body, "ui.messages").String(), "Could not find any login identifiers")

						} else {
							require.Equal(t, http.StatusBadRequest, resp.StatusCode, "%v", body)
							require.Contains(t, gjson.Get(body, "ui.messages").String(), "Could not find any login identifiers")
						}
					})
				})

				t.Run("case=should have verifiable address even if after session hook is disabled", func(t *testing.T) {
					// disable the after session hook
					conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationAfter+".code.hooks", []map[string]interface{}{})

					t.Cleanup(func() {
						conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationAfter+".code.hooks", []map[string]interface{}{
							{"hook": "session"},
						})
					})

					// 1. Initiate flow
					state := createRegistrationFlow(ctx, t, public, tc.apiType)

					// 2. Submit Identifier (email)
					state = registerNewUser(ctx, t, state, tc.apiType, nil)

					message := testhelpers.CourierExpectMessage(ctx, t, reg, state.email, "Use code")
					assert.Contains(t, message.Body, "Complete your account registration with the following code")

					registrationCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
					assert.NotEmpty(t, registrationCode)

					// 3. Submit OTP
					state = submitOTP(ctx, t, reg, state, func(v *url.Values) {
						v.Set("code", registrationCode)
					}, tc.apiType, nil)
				})

				t.Run("case=code should expire", func(t *testing.T) {
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.config.lifespan", "10ns")
					t.Cleanup(func() {
						conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.config.lifespan", "1h")
					})

					// 1. Initiate flow
					s := createRegistrationFlow(ctx, t, public, tc.apiType)

					// 2. Submit Identifier (email)
					s = registerNewUser(ctx, t, s, tc.apiType, nil)

					message := testhelpers.CourierExpectMessage(ctx, t, reg, s.email, "Use code")
					assert.Contains(t, message.Body, "Complete your account registration with the following code")

					registrationCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
					assert.NotEmpty(t, registrationCode)

					s = submitOTP(ctx, t, reg, s, func(v *url.Values) {
						v.Set("code", registrationCode)
					}, tc.apiType, func(ctx context.Context, t *testing.T, s *state, body string, resp *http.Response) {
						if tc.apiType == ApiTypeBrowser {
							// with browser clients we redirect back to the UI with a new flow id as a query parameter
							require.Equal(t, http.StatusOK, resp.StatusCode)
							require.Equal(t, conf.SelfServiceFlowRegistrationUI(ctx).Path, resp.Request.URL.Path)
							require.NotEqual(t, s.flowID, resp.Request.URL.Query().Get("flow"))
						} else {
							require.Equal(t, http.StatusGone, resp.StatusCode)
							require.Containsf(t, gjson.Get(body, "error.reason").String(), "self-service flow expired 0.00 minutes ago", "%s", body)
						}
					})
				})
			})
		}
	})
}

func TestPopulateRegistrationMethod(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	ctx = testhelpers.WithDefaultIdentitySchema(ctx, "file://stub/code.identity.schema.json")

	conf.MustSet(ctx, fmt.Sprintf("%s.%s.passwordless_enabled", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypeCodeAuth), true)

	s, err := reg.AllRegistrationStrategies().Strategy(identity.CredentialsTypeCodeAuth)
	require.NoError(t, err)

	fh, ok := s.(registration.FormHydrator)
	require.True(t, ok)

	toSnapshot := func(t *testing.T, f node.Nodes) {
		t.Helper()
		// The CSRF token has a unique value that messes with the snapshot - ignore it.
		f.ResetNodes("csrf_token")
		snapshotx.SnapshotT(t, f, snapshotx.ExceptNestedKeys("nonce", "src"))
	}

	newFlow := func(ctx context.Context, t *testing.T) (*http.Request, *registration.Flow) {
		r := httptest.NewRequest("GET", "/self-service/registration/browser", nil)
		r = r.WithContext(ctx)
		t.Helper()
		f, err := registration.NewFlow(conf, time.Minute, "csrf_token", r, flow.TypeBrowser)
		f.UI.Nodes = make(node.Nodes, 0)
		require.NoError(t, err)
		return r, f
	}

	t.Run("method=PopulateRegistrationMethod", func(t *testing.T) {
		r, f := newFlow(ctx, t)
		require.NoError(t, fh.PopulateRegistrationMethod(r, f))
		toSnapshot(t, f.UI.Nodes)
	})

	t.Run("method=PopulateRegistrationMethodProfile", func(t *testing.T) {
		r, f := newFlow(ctx, t)
		require.NoError(t, fh.PopulateRegistrationMethodProfile(r, f))
		toSnapshot(t, f.UI.Nodes)
	})

	t.Run("method=PopulateRegistrationMethodCredentials", func(t *testing.T) {
		r, f := newFlow(ctx, t)
		require.NoError(t, fh.PopulateRegistrationMethodCredentials(r, f))
		toSnapshot(t, f.UI.Nodes)
	})

	t.Run("method=idempotency", func(t *testing.T) {
		r, f := newFlow(ctx, t)

		var snapshots []node.Nodes

		t.Run("case=1", func(t *testing.T) {
			require.NoError(t, fh.PopulateRegistrationMethodProfile(r, f))
			snapshots = append(snapshots, f.UI.Nodes)
			toSnapshot(t, f.UI.Nodes)
		})

		t.Run("case=2", func(t *testing.T) {
			require.NoError(t, fh.PopulateRegistrationMethodCredentials(r, f))
			snapshots = append(snapshots, f.UI.Nodes)
			toSnapshot(t, f.UI.Nodes)
		})

		t.Run("case=3", func(t *testing.T) {
			require.NoError(t, fh.PopulateRegistrationMethodProfile(r, f))
			snapshots = append(snapshots, f.UI.Nodes)
			toSnapshot(t, f.UI.Nodes)
		})

		t.Run("case=4", func(t *testing.T) {
			require.NoError(t, fh.PopulateRegistrationMethodCredentials(r, f))
			snapshots = append(snapshots, f.UI.Nodes)
			toSnapshot(t, f.UI.Nodes)
		})

		t.Run("case=evaluate", func(t *testing.T) {
			assertx.EqualAsJSON(t, snapshots[0], snapshots[2])
			assertx.EqualAsJSON(t, snapshots[1], snapshots[3])
		})
	})
}
