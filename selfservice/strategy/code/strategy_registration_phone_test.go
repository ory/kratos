// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/nyaruka/phonenumbers"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/x/randx"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	oryClient "github.com/ory/kratos/internal/httpclient"
	"github.com/ory/kratos/internal/testhelpers"
)

func TestRegistrationCodeStrategy_SMS(t *testing.T) {
	type ApiType string

	const (
		ApiTypeBrowser ApiType = "browser"
		ApiTypeSPA     ApiType = "spa"
		ApiTypeNative  ApiType = "api"
	)

	type state struct {
		flowID         string
		client         *http.Client
		phone          string
		testServer     *httptest.Server
		resultIdentity *identity.Identity
	}

	setup := func(ctx context.Context, t *testing.T) (*config.Config, *driver.RegistryDefault, *httptest.Server) {
		conf, reg := internal.NewFastRegistryWithMocks(t)
		testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/code.phone.identity.schema.json")
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

	ctx := context.Background()
	conf, reg, public := setup(ctx, t)

	var externalVerifyResult string
	var externalVerifyRequestBody string
	initExternalSMSVerifier(t, ctx, conf, "file://./stub/request.config.registration.jsonnet",
		&externalVerifyRequestBody, &externalVerifyResult)

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

		if s.phone == "" {
			s.phone = testhelpers.RandomPhone()
		}

		rf, resp, err := testhelpers.NewSDKCustomClient(s.testServer, s.client).FrontendAPI.GetRegistrationFlow(context.Background()).Id(s.flowID).Execute()
		require.NoError(t, err)
		require.EqualValues(t, http.StatusOK, resp.StatusCode)

		values := testhelpers.SDKFormFieldsToURLValues(rf.Ui.Nodes)
		values.Set("traits.phone", s.phone)
		values.Set("traits.tos", "1")
		values.Set("method", "code")

		body, resp := testhelpers.RegistrationMakeRequest(t, apiType == ApiTypeNative, apiType == ApiTypeSPA, rf, s.client, testhelpers.EncodeFormAsJSON(t, apiType == ApiTypeNative, values))

		if submitAssertion != nil {
			submitAssertion(ctx, t, s, body, resp)
			return s
		}
		t.Logf("%v", body)

		if apiType == ApiTypeBrowser {
			require.EqualValues(t, http.StatusOK, resp.StatusCode, "%s", body)
		} else {
			require.EqualValues(t, http.StatusBadRequest, resp.StatusCode, "%s", body)
		}

		csrfToken := gjson.Get(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String()
		if apiType == ApiTypeNative {
			assert.Emptyf(t, csrfToken, "expected an empty value for csrf_token on native api flows but got %s", body)
		} else {
			assert.NotEmptyf(t, csrfToken, "%s", body)
		}
		require.Equal(t, s.phone, gjson.Get(body, "ui.nodes.#(attributes.name==traits.phone).attributes.value").String())

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
		values.Set("traits.phone", s.phone)
		values.Set("traits.tos", "1")
		vals(&values)

		body, resp := testhelpers.RegistrationMakeRequest(t, apiType == ApiTypeNative, apiType == ApiTypeSPA, rf, s.client, testhelpers.EncodeFormAsJSON(t, apiType == ApiTypeNative, values))

		if submitAssertion != nil {
			submitAssertion(ctx, t, s, body, resp)
			return s
		}

		require.Equal(t, http.StatusOK, resp.StatusCode, "%s", body)

		phoneNumber, err := phonenumbers.Parse(fmt.Sprintf("%s", s.phone), "")
		require.NoError(t, err)
		e164 := fmt.Sprintf("+%d%d", *phoneNumber.CountryCode, *phoneNumber.NationalNumber)

		verifiableAddress, err := reg.PrivilegedIdentityPool().FindVerifiableAddressByValue(ctx, identity.VerifiableAddressTypePhone, e164)
		require.NoError(t, err)
		require.Equal(t, strings.ToLower(e164), verifiableAddress.Value)

		id, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, verifiableAddress.IdentityID)
		require.NoError(t, err)
		require.NotNil(t, id.ID)

		_, ok := id.GetCredentials(identity.CredentialsTypeCodeAuth)
		require.True(t, ok)

		s.resultIdentity = id
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
		t.Run("flow="+tc.d, func(t *testing.T) {
			t.Run("case=should be able to register with code identity credentials", func(t *testing.T) {
				ctx := context.Background()

				// 1. Initiate flow
				state := createRegistrationFlow(ctx, t, public, tc.apiType)

				// 2. Submit Identifier (phone)
				state = registerNewUser(ctx, t, state, tc.apiType, nil)

				assert.Contains(t, externalVerifyResult, "code has been sent")

				// 3. Submit OTP
				state = submitOTP(ctx, t, reg, state, func(v *url.Values) {
					v.Set("code", "0000")
				}, tc.apiType, nil)

				assert.Contains(t, externalVerifyResult, "code valid")

			})

			t.Run("case=should normalize phone address on sign up", func(t *testing.T) {
				ctx := context.Background()

				// 1. Initiate flow
				state := createRegistrationFlow(ctx, t, public, tc.apiType)
				random := strings.ToLower(randx.MustString(9, randx.Numeric))
				sourcePhone := "+441" + random
				state.phone = "+4401" + random
				assert.NotEqual(t, sourcePhone, state.phone)

				// 2. Submit Identifier (email)
				state = registerNewUser(ctx, t, state, tc.apiType, nil)

				// 3. Submit OTP
				state = submitOTP(ctx, t, reg, state, func(v *url.Values) {
					v.Set("code", "0000")
				}, tc.apiType, nil)

				creds, ok := state.resultIdentity.GetCredentials(identity.CredentialsTypeCodeAuth)
				require.True(t, ok)
				require.Len(t, creds.Identifiers, 1)
				assert.Equal(t, sourcePhone, creds.Identifiers[0])
			})

			t.Run("case=code should not be able to use more than 5 times", func(t *testing.T) {
				ctx := context.Background()

				// 1. Initiate flow
				s := createRegistrationFlow(ctx, t, public, tc.apiType)

				// 2. Submit Identifier (phone)
				s = registerNewUser(ctx, t, s, tc.apiType, nil)

				reg.Persister().Transaction(ctx, func(ctx context.Context, connection *pop.Connection) error {
					count, err := connection.RawQuery(fmt.Sprintf("SELECT * FROM %s WHERE selfservice_registration_flow_id = ?", new(code.RegistrationCode).TableName(ctx)), uuid.FromStringOrNil(s.flowID)).Count(new(code.RegistrationCode))
					require.NoError(t, err)
					require.Equal(t, 1, count)
					return nil
				})

				for i := 0; i < 5; i++ {
					s = submitOTP(ctx, t, reg, s, func(v *url.Values) {
						v.Set("code", "1111")
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
					v.Set("code", "0000")
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
}
