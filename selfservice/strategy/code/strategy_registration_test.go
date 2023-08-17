// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	_ "embed"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	kratos "github.com/ory/kratos/internal/httpclient"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/strategy/code"
)

type state struct {
	flowID    string
	csrfToken string
	client    *http.Client
	email     string

	registrationFlow *kratos.RegistrationFlow
}

func TestRegistrationCodeStrategyDisabled(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/code.identity.schema.json")
	conf.MustSet(ctx, fmt.Sprintf("%s.%s.enabled", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypePassword.String()), false)
	conf.MustSet(ctx, fmt.Sprintf("%s.%s.enabled", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypeCodeAuth.String()), false)
	conf.MustSet(ctx, fmt.Sprintf("%s.%s.registration_enabled", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypeCodeAuth), false)

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
	setup := func(ctx context.Context, t *testing.T) (*config.Config, *driver.RegistryDefault, *httptest.Server) {
		conf, reg := internal.NewFastRegistryWithMocks(t)
		testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/code.identity.schema.json")
		conf.MustSet(ctx, fmt.Sprintf("%s.%s.enabled", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypePassword.String()), false)
		conf.MustSet(ctx, fmt.Sprintf("%s.%s.enabled", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypeCodeAuth.String()), false)
		conf.MustSet(ctx, fmt.Sprintf("%s.%s.registration_enabled", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypeCodeAuth), true)
		conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh")
		conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{"https://www.ory.sh"})
		conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationAfter+".code.hooks", []map[string]interface{}{
			{"hook": "session"},
		})

		_ = testhelpers.NewRegistrationUIFlowEchoServer(t, reg)
		_ = testhelpers.NewErrorTestServer(t, reg)

		public, _, _, _ := testhelpers.NewKratosServerWithCSRFAndRouters(t, reg)

		return conf, reg, public
	}

	createRegistrationFlow := func(ctx context.Context, t *testing.T, public *httptest.Server, isSPA bool) *state {
		t.Helper()

		client := testhelpers.NewClientWithCookies(t)
		clientInit := testhelpers.InitializeRegistrationFlowViaBrowser(t, client, public, isSPA, false, false)

		body, err := json.Marshal(clientInit)
		require.NoError(t, err)

		csrfToken := gjson.GetBytes(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String()
		require.NotEmpty(t, csrfToken)

		require.Truef(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==traits.email)").Exists(), "%s", body)
		require.Truef(t, gjson.GetBytes(body, "ui.nodes.#(attributes.value==code)").Exists(), "%s", body)

		return &state{
			csrfToken:        csrfToken,
			client:           client,
			flowID:           clientInit.GetId(),
			registrationFlow: clientInit,
		}
	}

	type onSubmitAssertion func(ctx context.Context, t *testing.T, s *state, body string, resp *http.Response)

	registerNewUser := func(ctx context.Context, t *testing.T, s *state, isSPA bool, submitAssertion onSubmitAssertion) *state {
		t.Helper()

		email := testhelpers.RandomEmail()

		s.email = email

		values := testhelpers.SDKFormFieldsToURLValues(s.registrationFlow.Ui.Nodes)
		values.Set("traits.email", email)
		values.Set("csrf_token", s.csrfToken)
		values.Set("method", "code")

		// TODO: API client
		body, resp := testhelpers.RegistrationMakeRequest(t, false, isSPA, s.registrationFlow, s.client, testhelpers.EncodeFormAsJSON(t, false, values))

		// 2. Submit Identifier (email)
		if submitAssertion != nil {
			submitAssertion(ctx, t, s, body, resp)
		} else {
			require.Equal(t, http.StatusOK, resp.StatusCode)
			csrfToken := gjson.Get(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String()
			assert.NotEmptyf(t, csrfToken, "%s", body)
			require.Equal(t, email, gjson.Get(body, "ui.nodes.#(attributes.name==traits.email).attributes.value").String())
		}

		var rf *kratos.RegistrationFlow
		require.NoError(t, json.Unmarshal([]byte(body), &rf))
		s.registrationFlow = rf

		return s
	}

	submitOTP := func(ctx context.Context, t *testing.T, reg *driver.RegistryDefault, s *state, otp string, isSPA bool, submitAssertion onSubmitAssertion) *state {
		t.Helper()

		values := testhelpers.SDKFormFieldsToURLValues(s.registrationFlow.Ui.Nodes)
		values.Set("csrf_token", s.csrfToken)
		values.Set("method", "code")
		values.Set("code", otp)
		values.Set("traits.email", s.email)

		// TODO: API clients
		body, resp := testhelpers.RegistrationMakeRequest(t, false, isSPA, s.registrationFlow, s.client, testhelpers.EncodeFormAsJSON(t, false, values))

		if submitAssertion != nil {
			submitAssertion(ctx, t, s, body, resp)
			return s
		}

		require.Equal(t, http.StatusOK, resp.StatusCode)

		verifiableAddress, err := reg.PrivilegedIdentityPool().FindVerifiableAddressByValue(ctx, identity.VerifiableAddressTypeEmail, s.email)
		require.NoError(t, err)
		require.Equal(t, s.email, verifiableAddress.Value)

		id, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, verifiableAddress.IdentityID)
		require.NoError(t, err)
		require.NotNil(t, id.ID)

		_, ok := id.GetCredentials(identity.CredentialsTypeCodeAuth)
		require.True(t, ok)

		return s
	}

	t.Run("test=different flows on the same configurations", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		_, reg, public := setup(ctx, t)

		t.Run("test=SPA client", func(t *testing.T) {
			t.Run("case=should be able to register with code identity credentials", func(t *testing.T) {
				ctx := context.Background()

				// 1. Initiate flow
				state := createRegistrationFlow(ctx, t, public, true)

				// 2. Submit Identifier (email)
				state = registerNewUser(ctx, t, state, true, nil)

				message := testhelpers.CourierExpectMessage(ctx, t, reg, state.email, "Complete your account registration")
				assert.Contains(t, message.Body, "please complete your account registration by entering the following code")

				registrationCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
				assert.NotEmpty(t, registrationCode)

				// 3. Submit OTP
				state = submitOTP(ctx, t, reg, state, registrationCode, true, nil)
			})

			t.Run("case=should be able to resend the code", func(t *testing.T) {
				ctx := context.Background()

				s := createRegistrationFlow(ctx, t, public, true)

				s = registerNewUser(ctx, t, s, true, func(ctx context.Context, t *testing.T, s *state, body string, resp *http.Response) {
					require.Equal(t, http.StatusOK, resp.StatusCode)
					csrfToken := gjson.Get(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String()
					require.NotEmptyf(t, csrfToken, "%s", body)
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
				req, err := http.NewRequestWithContext(ctx, "POST", public.URL+registration.RouteSubmitFlow+"?flow="+s.flowID, strings.NewReader(url.Values{
					"csrf_token":   {s.csrfToken},
					"method":       {"code"},
					"resend":       {"code"},
					"traits.email": {s.email},
				}.Encode()))
				require.NoError(t, err)

				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				req.Header.Set("Accept", "application/json")

				resp, err := s.client.Do(req)
				require.NoError(t, err)

				require.Equal(t, http.StatusOK, resp.StatusCode)

				// get the new code from email
				message = testhelpers.CourierExpectMessage(ctx, t, reg, s.email, "Complete your account registration")
				assert.Contains(t, message.Body, "please complete your account registration by entering the following code")

				registrationCode2 := testhelpers.CourierExpectCodeInMessage(t, message, 1)
				assert.NotEmpty(t, registrationCode2)

				require.NotEqual(t, registrationCode, registrationCode2)

				// try submit old code
				s = submitOTP(ctx, t, reg, s, registrationCode, true, func(ctx context.Context, t *testing.T, s *state, body string, resp *http.Response) {
					require.Equal(t, http.StatusBadRequest, resp.StatusCode)
					require.Contains(t, gjson.Get(body, "ui.messages").String(), "The registration code is invalid or has already been used. Please try again")
				})

				s = submitOTP(ctx, t, reg, s, registrationCode2, true, nil)
			})

			t.Run("case=swapping out traits should not be possible on code submit", func(t *testing.T) {
				ctx := context.Background()

				// 1. Initiate flow
				s := createRegistrationFlow(ctx, t, public, true)

				// 2. Submit Identifier (email)
				s = registerNewUser(ctx, t, s, true, nil)

				message := testhelpers.CourierExpectMessage(ctx, t, reg, s.email, "Complete your account registration")
				assert.Contains(t, message.Body, "please complete your account registration by entering the following code")

				registrationCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
				assert.NotEmpty(t, registrationCode)

				s.email = "not-" + s.email // swap out email

				// 3. Submit OTP
				s = submitOTP(ctx, t, reg, s, registrationCode, true, func(ctx context.Context, t *testing.T, s *state, body string, resp *http.Response) {
					require.Equal(t, http.StatusBadRequest, resp.StatusCode)
					require.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "The provided traits do not match the traits previously associated with this flow.")
				})
			})

			t.Run("case=code should not be able to use more than 5 times", func(t *testing.T) {
				ctx := context.Background()

				// 1. Initiate flow
				s := createRegistrationFlow(ctx, t, public, true)

				// 2. Submit Identifier (email)
				s = registerNewUser(ctx, t, s, true, nil)

				reg.Persister().Transaction(ctx, func(ctx context.Context, connection *pop.Connection) error {
					count, err := connection.RawQuery(fmt.Sprintf("SELECT * FROM %s WHERE selfservice_registration_flow_id = ?", new(code.RegistrationCode).TableName(ctx)), uuid.FromStringOrNil(s.flowID)).Count(new(code.RegistrationCode))
					require.NoError(t, err)
					require.Equal(t, 1, count)
					return nil
				})

				for i := 0; i < 5; i++ {
					s = submitOTP(ctx, t, reg, s, "111111", true, func(ctx context.Context, t *testing.T, s *state, body string, resp *http.Response) {
						require.Equal(t, http.StatusBadRequest, resp.StatusCode)
						require.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "The registration code is invalid or has already been used")
					})
				}

				s = submitOTP(ctx, t, reg, s, "111111", true, func(ctx context.Context, t *testing.T, s *state, body string, resp *http.Response) {
					require.Equal(t, http.StatusBadRequest, resp.StatusCode)
					require.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "The request was submitted too often.")
				})
			})

		})

		t.Run("test=Browser client", func(t *testing.T) {
			t.Run("case=should be able to register with code identity credentials", func(t *testing.T) {
				ctx := context.Background()

				// 1. Initiate flow
				state := createRegistrationFlow(ctx, t, public, false)

				// 2. Submit Identifier (email)
				state = registerNewUser(ctx, t, state, false, nil)

				message := testhelpers.CourierExpectMessage(ctx, t, reg, state.email, "Complete your account registration")
				assert.Contains(t, message.Body, "please complete your account registration by entering the following code")

				registrationCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
				assert.NotEmpty(t, registrationCode)

				// 3. Submit OTP
				state = submitOTP(ctx, t, reg, state, registrationCode, false, nil)
			})

			t.Run("case=should be able to resend the code", func(t *testing.T) {
				ctx := context.Background()

				s := createRegistrationFlow(ctx, t, public, false)

				s = registerNewUser(ctx, t, s, false, func(ctx context.Context, t *testing.T, s *state, body string, resp *http.Response) {
					require.Equal(t, http.StatusOK, resp.StatusCode)

					csrfToken := gjson.Get(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String()
					require.NotEmptyf(t, csrfToken, "%s", body)
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
				req, err := http.NewRequestWithContext(ctx, "POST", public.URL+registration.RouteSubmitFlow+"?flow="+s.flowID, strings.NewReader(url.Values{
					"csrf_token":   {s.csrfToken},
					"method":       {"code"},
					"resend":       {"code"},
					"traits.email": {s.email},
				}.Encode()))
				require.NoError(t, err)

				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				req.Header.Set("Accept", "application/json")

				resp, err := s.client.Do(req)
				require.NoError(t, err)

				require.Equal(t, http.StatusOK, resp.StatusCode)

				// get the new code from email
				message = testhelpers.CourierExpectMessage(ctx, t, reg, s.email, "Complete your account registration")
				assert.Contains(t, message.Body, "please complete your account registration by entering the following code")

				registrationCode2 := testhelpers.CourierExpectCodeInMessage(t, message, 1)
				assert.NotEmpty(t, registrationCode2)

				require.NotEqual(t, registrationCode, registrationCode2)

				// try submit old code
				s = submitOTP(ctx, t, reg, s, registrationCode, false, func(ctx context.Context, t *testing.T, s *state, body string, resp *http.Response) {
					require.Equal(t, http.StatusBadRequest, resp.StatusCode)
					require.Contains(t, gjson.Get(body, "ui.messages").String(), "The registration code is invalid or has already been used. Please try again")
				})

				s = submitOTP(ctx, t, reg, s, registrationCode2, false, nil)
			})

			t.Run("case=swapping out traits should not be possible on code submit", func(t *testing.T) {
				ctx := context.Background()

				// 1. Initiate flow
				s := createRegistrationFlow(ctx, t, public, false)

				// 2. Submit Identifier (email)
				s = registerNewUser(ctx, t, s, false, nil)

				message := testhelpers.CourierExpectMessage(ctx, t, reg, s.email, "Complete your account registration")
				assert.Contains(t, message.Body, "please complete your account registration by entering the following code")

				registrationCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
				assert.NotEmpty(t, registrationCode)

				s.email = "not-" + s.email // swap out email

				// 3. Submit OTP
				s = submitOTP(ctx, t, reg, s, registrationCode, false, func(ctx context.Context, t *testing.T, s *state, body string, resp *http.Response) {
					require.Equal(t, http.StatusOK, resp.Status)
					require.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "The provided traits do not match the traits previously associated with this flow.")
				})
			})

			t.Run("case=code should not be able to use more than 5 times", func(t *testing.T) {
				ctx := context.Background()

				// 1. Initiate flow
				s := createRegistrationFlow(ctx, t, public, false)

				// 2. Submit Identifier (email)
				s = registerNewUser(ctx, t, s, false, nil)

				reg.Persister().Transaction(ctx, func(ctx context.Context, connection *pop.Connection) error {
					count, err := connection.RawQuery(fmt.Sprintf("SELECT * FROM %s WHERE selfservice_registration_flow_id = ?", new(code.RegistrationCode).TableName(ctx)), uuid.FromStringOrNil(s.flowID)).Count(new(code.RegistrationCode))
					require.NoError(t, err)
					require.Equal(t, 1, count)
					return nil
				})

				for i := 0; i < 5; i++ {
					s = submitOTP(ctx, t, reg, s, "111111", false, func(ctx context.Context, t *testing.T, s *state, body string, resp *http.Response) {
						require.Equal(t, http.StatusBadRequest, resp.StatusCode)
						require.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "The registration code is invalid or has already been used")
					})
				}

				s = submitOTP(ctx, t, reg, s, "111111", false, func(ctx context.Context, t *testing.T, s *state, body string, resp *http.Response) {
					require.Equal(t, http.StatusBadRequest, resp.StatusCode)
					require.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "The request was submitted too often.")
				})
			})

		})
	})

	t.Run("test=cases with different configs", func(t *testing.T) {
		ctx := context.Background()
		conf, reg, public := setup(ctx, t)

		t.Run("test=SPA client", func(t *testing.T) {
			t.Run("case=should fail when schema does not contain the `code` extension", func(t *testing.T) {
				testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/default.schema.json")
				t.Cleanup(func() {
					testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/code.identity.schema.json")
				})

				// 1. Initiate flow
				s := createRegistrationFlow(ctx, t, public, true)

				// 2. Submit Identifier (email)
				s = registerNewUser(ctx, t, s, true, func(ctx context.Context, t *testing.T, s *state, body string, resp *http.Response) {
					require.Equal(t, http.StatusBadRequest, resp.StatusCode)
					require.Contains(t, gjson.Get(body, "ui.messages").String(), "Could not find any login identifiers")
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
				state := createRegistrationFlow(ctx, t, public, true)

				// 2. Submit Identifier (email)
				state = registerNewUser(ctx, t, state, true, nil)

				message := testhelpers.CourierExpectMessage(ctx, t, reg, state.email, "Complete your account registration")
				assert.Contains(t, message.Body, "please complete your account registration by entering the following code")

				registrationCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
				assert.NotEmpty(t, registrationCode)

				// 3. Submit OTP
				state = submitOTP(ctx, t, reg, state, registrationCode, true, nil)
			})

			t.Run("case=code should expire", func(t *testing.T) {
				conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.config.lifespan", "10ns")
				t.Cleanup(func() {
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.config.lifespan", "1h")
				})

				// 1. Initiate flow
				s := createRegistrationFlow(ctx, t, public, true)

				// 2. Submit Identifier (email)
				s = registerNewUser(ctx, t, s, true, nil)

				message := testhelpers.CourierExpectMessage(ctx, t, reg, s.email, "Complete your account registration")
				assert.Contains(t, message.Body, "please complete your account registration by entering the following code")

				registrationCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
				assert.NotEmpty(t, registrationCode)

				s = submitOTP(ctx, t, reg, s, registrationCode, false, func(ctx context.Context, t *testing.T, s *state, body string, resp *http.Response) {
					require.Equal(t, http.StatusGone, resp.StatusCode)
					require.Contains(t, gjson.Get(body, "error.reason").String(), "self-service flow expired 0.00 minutes ago")
				})
			})

		})

		t.Run("test=Browser client", func(t *testing.T) {
			t.Run("case=should fail when schema does not contain the `code` extension", func(t *testing.T) {
				testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/default.schema.json")
				t.Cleanup(func() {
					testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/code.identity.schema.json")
				})

				// 1. Initiate flow
				s := createRegistrationFlow(ctx, t, public, false)

				// 2. Submit Identifier (email)
				s = registerNewUser(ctx, t, s, false, func(ctx context.Context, t *testing.T, s *state, body string, resp *http.Response) {
					require.Equal(t, http.StatusBadRequest, resp.StatusCode)
					require.Contains(t, gjson.Get(body, "ui.messages").String(), "Could not find any login identifiers")
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
				state := createRegistrationFlow(ctx, t, public, false)

				// 2. Submit Identifier (email)
				state = registerNewUser(ctx, t, state, false, nil)

				message := testhelpers.CourierExpectMessage(ctx, t, reg, state.email, "Complete your account registration")
				assert.Contains(t, message.Body, "please complete your account registration by entering the following code")

				registrationCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
				assert.NotEmpty(t, registrationCode)

				// 3. Submit OTP
				state = submitOTP(ctx, t, reg, state, registrationCode, false, nil)
			})

			t.Run("case=code should expire", func(t *testing.T) {
				conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.config.lifespan", "10ns")
				t.Cleanup(func() {
					conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.config.lifespan", "1h")
				})

				// 1. Initiate flow
				s := createRegistrationFlow(ctx, t, public, false)

				// 2. Submit Identifier (email)
				s = registerNewUser(ctx, t, s, false, nil)

				message := testhelpers.CourierExpectMessage(ctx, t, reg, s.email, "Complete your account registration")
				assert.Contains(t, message.Body, "please complete your account registration by entering the following code")

				registrationCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
				assert.NotEmpty(t, registrationCode)

				s = submitOTP(ctx, t, reg, s, registrationCode, false, func(ctx context.Context, t *testing.T, s *state, body string, resp *http.Response) {
					require.Equal(t, http.StatusGone, resp.StatusCode)
					require.Contains(t, gjson.Get(body, "error.reason").String(), "self-service flow expired 0.00 minutes ago")
				})
			})
		})

	})
}
