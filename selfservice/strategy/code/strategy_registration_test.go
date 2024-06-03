// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/ory/kratos/courier/template/sms"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
	"github.com/ory/x/assertx"
	"github.com/ory/x/snapshotx"
	"github.com/ory/x/urlx"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

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

		rf, resp, err := testhelpers.NewSDKCustomClient(s.testServer, s.client).FrontendApi.GetRegistrationFlow(context.Background()).Id(s.flowID).Execute()
		require.NoError(t, err)
		require.EqualValues(t, http.StatusOK, resp.StatusCode)

		values := testhelpers.SDKFormFieldsToURLValues(rf.Ui.Nodes)
		values.Set("traits.email", s.email)
		values.Set("traits.tos", "1")
		values.Set("method", "code")

		body, resp := testhelpers.RegistrationMakeRequest(t, apiType == ApiTypeNative, apiType == ApiTypeSPA, rf, s.client, testhelpers.EncodeFormAsJSON(t, apiType == ApiTypeNative, values))

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

		rf, resp, err := testhelpers.NewSDKCustomClient(s.testServer, s.client).FrontendApi.GetRegistrationFlow(context.Background()).Id(s.flowID).Execute()
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

		if submitAssertion != nil {
			submitAssertion(ctx, t, s, body, resp)
			return s
		}

		require.Equal(t, http.StatusOK, resp.StatusCode)

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
		_, reg, public := setup(ctx, t)

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

					message := testhelpers.CourierExpectMessage(ctx, t, reg, state.email, "Complete your account registration")
					assert.Contains(t, message.Body, "please complete your account registration by entering the following code")

					registrationCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
					assert.NotEmpty(t, registrationCode)

					// 3. Submit OTP
					state = submitOTP(ctx, t, reg, state, func(v *url.Values) {
						v.Set("code", registrationCode)
					}, tc.apiType, nil)
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

					message := testhelpers.CourierExpectMessage(ctx, t, reg, sourceMail, "Complete your account registration")
					assert.Contains(t, message.Body, "please complete your account registration by entering the following code")

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
						require.Equal(t, "code", val)
					})

					message := testhelpers.CourierExpectMessage(ctx, t, reg, s.email, "Complete your account registration")
					assert.Contains(t, message.Body, "please complete your account registration by entering the following code")

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
						require.Containsf(t, gjson.Get(body, "ui.messages").String(), "An email containing a code has been sent to the email address you provided.", "%s", body)
					})

					// get the new code from email
					message = testhelpers.CourierExpectMessage(ctx, t, reg, s.email, "Complete your account registration")
					assert.Contains(t, message.Body, "please complete your account registration by entering the following code")

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

					message := testhelpers.CourierExpectMessage(ctx, t, reg, s.email, "Complete your account registration")
					assert.Contains(t, message.Body, "please complete your account registration by entering the following code")

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

					message := testhelpers.CourierExpectMessage(ctx, t, reg, s.email, "Complete your account registration")
					assert.Contains(t, message.Body, "please complete your account registration by entering the following code")

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
					testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/default.schema.json")
					t.Cleanup(func() {
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
							rf, resp, err := testhelpers.NewSDKCustomClient(public, s.client).FrontendApi.GetRegistrationFlow(ctx).Id(resp.Request.URL.Query().Get("flow")).Execute()
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

					message := testhelpers.CourierExpectMessage(ctx, t, reg, state.email, "Complete your account registration")
					assert.Contains(t, message.Body, "please complete your account registration by entering the following code")

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

					message := testhelpers.CourierExpectMessage(ctx, t, reg, s.email, "Complete your account registration")
					assert.Contains(t, message.Body, "please complete your account registration by entering the following code")

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

func TestRegistrationCode_SMS(t *testing.T) {
	ctx := context.Background()

	cleanCourierQueue := func(reg *driver.RegistryDefault) {
		for {
			_, err := reg.CourierPersister().NextMessages(context.Background(), 10)
			if err != nil {
				return
			}
		}
	}

	//requestCode := func(t *testing.T, publicTS *httptest.Server, identifier string, statusCode int) {
	//	hc := new(http.Client)
	//	f := testhelpers.InitializeRegistrationFlow(t, true, hc, publicTS, false)
	//	var values = func(v url.Values) {
	//		v.Set("method", "code")
	//		v.Set("traits.phone", identifier)
	//	}
	//	testhelpers.SubmitRegistrationFormWithFlow(t, true, hc, values,
	//		false, statusCode, publicTS.URL+registration.RouteSubmitFlow, f)
	//}

	getRegistrationNode := func(f *oryClient.RegistrationFlow, nodeName string) *oryClient.UiNode {
		for _, n := range f.Ui.Nodes {
			if n.Attributes.UiNodeInputAttributes.Name == nodeName {
				return &n
			}
		}
		return nil
	}

	t.Run("case=registration", func(t *testing.T) {
		conf, reg := internal.NewFastRegistryWithMocks(t)

		router := x.NewRouterPublic()
		admin := x.NewRouterAdmin()
		conf.MustSet(ctx, fmt.Sprintf("%s.%s.enabled", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypeCodeAuth.String()), true)
		conf.MustSet(ctx, fmt.Sprintf("%s.%s.passwordless_enabled", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypeCodeAuth.String()), true)

		publicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, admin)
		//errTS := testhelpers.NewErrorTestServer(t, reg)
		//uiTS := testhelpers.NewRegistrationUIFlowEchoServer(t, reg)
		redirTS := testhelpers.NewRedirSessionEchoTS(t, reg)

		// Overwrite these two to ensure that they run
		conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, redirTS.URL+"/default-return-to")
		conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationAfter+"."+config.DefaultBrowserReturnURL, redirTS.URL+"/registration-return-ts")
		testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/default.schema.json")
		//conf.MustSet(ctx, config.CodeMaxAttempts, 5)

		t.Run("case=should fail if identifier changed when submitted with code", func(t *testing.T) {
			identifier := "+4550050000"

			cleanCourierQueue(reg)

			hc := new(http.Client)

			f := testhelpers.InitializeRegistrationFlow(t, true, hc, publicTS, false)

			var values = func(v url.Values) {
				v.Set("method", "code")
				v.Set("traits.phone", identifier)
			}
			body := testhelpers.SubmitRegistrationFormWithFlow(t, true, hc, values,
				false, http.StatusBadRequest, publicTS.URL+registration.RouteSubmitFlow, f)
			assertx.EqualAsJSON(t,
				text.NewRegistrationEmailWithCodeSent(),
				json.RawMessage(gjson.Get(body, "ui.messages.0").Raw),
				"%s", testhelpers.PrettyJSON(t, []byte(body)),
			)
			assert.Equal(t, flow.StateSMSSent.String(), gjson.Get(body, "state").String(),
				"%s", testhelpers.PrettyJSON(t, []byte(body)))

			message := testhelpers.CourierExpectMessage(ctx, t, reg, identifier, "")
			var smsModel sms.RegistrationCodeValidModel
			err := json.Unmarshal(message.TemplateData, &smsModel)
			assert.NoError(t, err)

			values = func(v url.Values) {
				v.Set("method", "code")
				v.Set("traits.phone", identifier+"2")
				v.Set("code", smsModel.RegistrationCode)
			}

			body = testhelpers.SubmitRegistrationFormWithFlow(t, true, hc, values,
				false, http.StatusBadRequest, publicTS.URL+registration.RouteSubmitFlow, f)
			assertx.EqualAsJSON(t,
				text.NewErrorValidationTraitsMismatch(),
				json.RawMessage(gjson.Get(body, "ui.messages.0").Raw),
				"%s", testhelpers.PrettyJSON(t, []byte(body)),
			)
		})

		//t.Run("case=should fail if spam detected", func(t *testing.T) {
		//	identifier := "+4550050001"
		//	conf.MustSet(ctx, config.CodeSMSSpamProtectionEnabled, true)
		//
		//	for i := 0; i <= 50; i++ {
		//		requestCode(t, publicTS, identifier, http.StatusOK)
		//	}
		//
		//	requestCode(t, publicTS, identifier, http.StatusBadRequest)
		//
		//	identifier = "+456005"
		//
		//	for i := 0; i <= 100; i++ {
		//		requestCode(t, publicTS, identifier+fmt.Sprintf("%04d", i), http.StatusOK)
		//	}
		//
		//	requestCode(t, publicTS, identifier+"0101", http.StatusBadRequest)
		//})

		var expectSuccessfulLogin = func(
			t *testing.T, isAPI, isSPA bool, hc *http.Client,
			expectReturnTo string,
			identifier string,
		) string {
			if hc == nil {
				if isAPI {
					hc = new(http.Client)
				} else {
					hc = testhelpers.NewClientWithCookies(t)
				}
			}

			cleanCourierQueue(reg)

			f := testhelpers.InitializeRegistrationFlow(t, isAPI, hc, publicTS, isSPA)

			assert.Empty(t, getRegistrationNode(f, "code"))
			assert.NotEmpty(t, getRegistrationNode(f, "traits.phone"))

			var values = func(v url.Values) {
				v.Set("method", "code")
				v.Set("traits.phone", identifier)
			}
			body := testhelpers.SubmitRegistrationFormWithFlow(t, isAPI, hc, values,
				isSPA, testhelpers.ExpectStatusCode(isAPI, http.StatusBadRequest, http.StatusOK), expectReturnTo, f)

			messages, err := reg.CourierPersister().NextMessages(context.Background(), 10)
			assert.NoError(t, err, "Courier queue should not be empty.")
			assert.Equal(t, 1, len(messages))
			var smsModel sms.RegistrationCodeValidModel
			err = json.Unmarshal(messages[0].TemplateData, &smsModel)
			assert.NoError(t, err)

			st := gjson.Get(body, "session_token").String()
			assert.Empty(t, st, "Response body: %s", body) //No session token as we have not presented the code yet

			values = func(v url.Values) {
				v.Set("method", "code")
				v.Set("traits.phone", identifier)
				v.Set("code", smsModel.RegistrationCode)
			}

			body = testhelpers.SubmitRegistrationFormWithFlow(t, isAPI, hc, values,
				isSPA, http.StatusOK, expectReturnTo, f)

			assert.Equal(t, identifier, gjson.Get(body, "session.identity.traits.phone").String(),
				"%s", body)
			identityID, err := uuid.FromString(gjson.Get(body, "identity.id").String())
			assert.NoError(t, err)
			i, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), identityID)
			assert.NoError(t, err)
			assert.NotEmpty(t, i.Credentials, "%s", body)
			assert.Equal(t, identifier, i.Credentials["code"].Identifiers[0], "%s", body)
			assert.NotEmpty(t, gjson.Get(body, "session_token").String(), "%s", body)
			assert.Equal(t, identifier, gjson.Get(body, "identity.verifiable_addresses.0.value").String())
			assert.Equal(t, "true", gjson.Get(body, "identity.verifiable_addresses.0.verified").String())

			return body
		}

		t.Run("case=should pass and set up a session", func(t *testing.T) {
			testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/default.schema.json")
			conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypeCodeAuth.String()), []config.SelfServiceHook{{Name: "session"}})
			t.Cleanup(func() {
				conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypeCodeAuth.String()), nil)
			})

			identifier := "+4570050001"

			t.Run("type=api", func(t *testing.T) {
				expectSuccessfulLogin(t, true, false, nil,
					publicTS.URL+registration.RouteSubmitFlow, identifier)
			})

			//t.Run("type=spa", func(t *testing.T) {
			//	hc := testhelpers.NewClientWithCookies(t)
			//	body := expectSuccessfulLogin(t, false, true, hc, func(v url.Values) {
			//		v.Set("traits.username", "registration-identifier-8-spa")
			//		v.Set("password", x.NewUUID().String())
			//		v.Set("traits.foobar", "bar")
			//	})
			//	assert.Equal(t, `registration-identifier-8-spa`, gjson.Get(body, "identity.traits.username").String(), "%s", body)
			//	assert.Empty(t, gjson.Get(body, "session_token").String(), "%s", body)
			//	assert.NotEmpty(t, gjson.Get(body, "session.id").String(), "%s", body)
			//})
			//
			//t.Run("type=browser", func(t *testing.T) {
			//	body := expectSuccessfulLogin(t, false, false, nil, func(v url.Values) {
			//		v.Set("traits.username", "registration-identifier-8-browser")
			//		v.Set("password", x.NewUUID().String())
			//		v.Set("traits.foobar", "bar")
			//	})
			//	assert.Equal(t, `registration-identifier-8-browser`, gjson.Get(body, "identity.traits.username").String(), "%s", body)
			//})
		})

		t.Run("case=should create verifiable address", func(t *testing.T) {
			identifier := "+1234567890"
			conf.MustSet(ctx, config.CodeTestNumbers, []string{identifier})
			createdIdentity := &identity.Identity{
				SchemaID: "default",
				Traits:   identity.Traits(fmt.Sprintf(`{"phone":"%s"}`, identifier)),
				State:    identity.StateActive}
			err := reg.IdentityManager().Create(context.Background(), createdIdentity)
			assert.NoError(t, err)

			i, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), createdIdentity.ID)
			assert.NoError(t, err)
			assert.Equal(t, identifier, i.VerifiableAddresses[0].Value)
			assert.False(t, i.VerifiableAddresses[0].Verified)
			assert.Equal(t, identity.VerifiableAddressStatusPending, i.VerifiableAddresses[0].Status)
		})

		t.Run("method=TestPopulateSignUpMethod", func(t *testing.T) {
			conf.MustSet(ctx, config.ViperKeyPublicBaseURL, "https://foo/")
			t.Cleanup(func() {
				conf.MustSet(ctx, config.ViperKeyPublicBaseURL, publicTS.URL)
			})

			sr, err := registration.NewFlow(conf, time.Minute, "nosurf", &http.Request{URL: urlx.ParseOrPanic("/")}, flow.TypeBrowser)
			require.NoError(t, err)
			require.NoError(t, reg.RegistrationStrategies(context.Background()).
				MustStrategy(identity.CredentialsTypeCodeAuth).(*code.Strategy).PopulateRegistrationMethod(&http.Request{}, sr))

			snapshotx.SnapshotTExcept(t, sr.UI, []string{"action", "nodes.3.attributes.value"})
		})

		//t.Run("case=should use standby sender", func(t *testing.T) {
		//	senderMessagesCount := 0
		//	standbySenderMessagesCount := 0
		//	senderSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//		senderMessagesCount++
		//	}))
		//	standbySenderSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//		standbySenderMessagesCount++
		//	}))
		//
		//	//configTemplate := `{
		//	//	"url": "%s",
		//	//	"method": "POST",
		//	//	"body": "file://./stub/request.config.twilio.jsonnet"
		//	//}`
		//
		//	//conf.MustSet(ctx, config.ViperKeyCourierSMSRequestConfig, fmt.Sprintf(configTemplate, senderSrv.URL))
		//	//conf.MustSet(ctx, config.ViperKeyCourierSMSStandbyRequestConfig, fmt.Sprintf(configTemplate, standbySenderSrv.URL))
		//	//conf.MustSet(ctx, config.ViperKeyCourierSMSFrom, "test sender")
		//	//conf.MustSet(ctx, config.ViperKeyCourierSMSStandbyFrom, "test standby sender")
		//	//conf.MustSet(ctx, config.ViperKeyCourierSMSEnabled, true)
		//	conf.MustSet(ctx, config.ViperKeyCourierSMTPURL, "http://foo.url")
		//
		//	c, err := reg.Courier(ctx)
		//	require.NoError(t, err)
		//
		//	ctx, cancel := context.WithCancel(ctx)
		//	defer t.Cleanup(cancel)
		//
		//	identifier := "+4550050005"
		//	f := testhelpers.InitializeRegistrationFlow(t, true, nil, publicTS, false)
		//	var values = func(v url.Values) {
		//		v.Set("method", "code")
		//		v.Set("traits.phone", identifier)
		//	}
		//	testhelpers.SubmitRegistrationFormWithFlow(t, true, nil, values,
		//		false, http.StatusOK, publicTS.URL+registration.RouteSubmitFlow, f)
		//	testhelpers.SubmitRegistrationFormWithFlow(t, true, nil, values,
		//		false, http.StatusOK, publicTS.URL+registration.RouteSubmitFlow, f)
		//
		//	go func() {
		//		require.NoError(t, c.Work(ctx))
		//	}()
		//
		//	require.NoError(t, resilience.Retry(reg.Logger(), time.Millisecond*250, time.Second*10, func() error {
		//		if senderMessagesCount+standbySenderMessagesCount >= 2 {
		//			return nil
		//		}
		//		return errors.New("messages not sent yet")
		//	}))
		//
		//	assert.Equal(t, 1, senderMessagesCount)
		//	assert.Equal(t, 1, standbySenderMessagesCount)
		//
		//	senderSrv.Close()
		//	standbySenderSrv.Close()
		//
		//})

	})
}
