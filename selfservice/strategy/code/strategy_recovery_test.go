// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/corpx"
	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	kratos "github.com/ory/kratos/internal/httpclient"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/hook/hooktest"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/assertx"
	"github.com/ory/x/contextx"
	"github.com/ory/x/ioutilx"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"
)

func init() {
	corpx.RegisterFakes()
}

func extractCsrfToken(body []byte) string {
	return gjson.GetBytes(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String()
}

type ClientType string

const (
	RecoveryClientTypeBrowser ClientType = "browser"
	RecoveryClientTypeSPA     ClientType = "spa"
	RecoveryClientTypeAPI     ClientType = "api"
)

func (c ClientType) String() string {
	return string(c)
}

func apiHttpClient(*testing.T) *http.Client {
	return &http.Client{}
}

func spaHttpClient(t *testing.T) *http.Client {
	return testhelpers.NewClientWithCookies(t)
}

func browserHttpClient(t *testing.T) *http.Client {
	return testhelpers.NewClientWithCookies(t)
}

var flowTypes = []ClientType{RecoveryClientTypeBrowser, RecoveryClientTypeAPI, RecoveryClientTypeSPA}

var flowTypeCases = []struct {
	FlowType        flow.Type
	ClientType      ClientType
	GetClient       func(*testing.T) *http.Client
	FormContentType string
}{
	{
		FlowType:        flow.TypeBrowser,
		ClientType:      RecoveryClientTypeBrowser,
		GetClient:       testhelpers.NewClientWithCookies,
		FormContentType: "application/x-www-form-urlencoded",
	},
	{
		FlowType:   flow.TypeAPI,
		ClientType: RecoveryClientTypeAPI,
		GetClient: func(_ *testing.T) *http.Client {
			return &http.Client{}
		},
		FormContentType: "application/json",
	},
	{
		FlowType:        flow.TypeBrowser,
		ClientType:      RecoveryClientTypeSPA,
		GetClient:       testhelpers.NewClientWithCookies,
		FormContentType: "application/json",
	},
}

func withCSRFToken(t *testing.T, clientType ClientType, body string, v url.Values) string {
	t.Helper()
	csrfToken := gjson.Get(body, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String()
	if csrfToken != "" && clientType != RecoveryClientTypeAPI {
		v.Set("csrf_token", csrfToken)
	}
	if clientType == RecoveryClientTypeBrowser {
		return v.Encode()
	}
	return testhelpers.EncodeFormAsJSON(t, true, v)
}

func createIdentityToRecover(t *testing.T, reg *driver.RegistryDefault, email string) *identity.Identity {
	t.Helper()
	id := &identity.Identity{
		Credentials: map[identity.CredentialsType]identity.Credentials{
			"password": {
				Type:        "password",
				Identifiers: []string{email},
				Config:      sqlxx.JSONRawMessage(`{"hashed_password":"$2a$08$.cOYmAd.vCpDOoiVJrO5B.hjTLKQQ6cAK40u8uB.FnZDyPvVvQ9Q."}`),
			},
		},
		Traits:   identity.Traits(fmt.Sprintf(`{"email":"%s"}`, email)),
		SchemaID: config.DefaultIdentityTraitsSchemaID,
		State:    identity.StateActive,
	}
	require.NoError(t, reg.IdentityManager().Create(context.Background(), id, identity.ManagerAllowWriteProtectedTraits))

	addr, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, email)
	assert.NoError(t, err)
	assert.False(t, addr.Verified)
	assert.Nil(t, addr.VerifiedAt)
	assert.Equal(t, identity.VerifiableAddressStatusPending, addr.Status)
	return id
}

func TestRecovery(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.StrategyEnable(t, conf, string(recovery.RecoveryStrategyCode), true)
	testhelpers.StrategyEnable(t, conf, string(recovery.RecoveryStrategyLink), false)

	initViper(t, ctx, conf)

	_ = testhelpers.NewRecoveryUIFlowEchoServer(t, reg)
	_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
	_ = testhelpers.NewSettingsUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)

	public, _, _, _ := testhelpers.NewKratosServerWithCSRFAndRouters(t, reg)

	submitRecovery := func(t *testing.T, client *http.Client, flowType ClientType, values func(url.Values), code int) string {
		isSPA := flowType == RecoveryClientTypeSPA
		isAPI := flowType == RecoveryClientTypeAPI
		if client == nil {
			client = testhelpers.NewDebugClient(t)
			if !isAPI {
				client = testhelpers.NewClientWithCookies(t)
				client.Transport = testhelpers.NewTransportWithLogger(http.DefaultTransport, t).RoundTripper
			}
		}

		expectedUrl := testhelpers.ExpectURL(isAPI || isSPA, public.URL+recovery.RouteSubmitFlow, conf.SelfServiceFlowRecoveryUI(ctx).String())
		return testhelpers.SubmitRecoveryForm(t, isAPI, isSPA, client, public, values, code, expectedUrl)
	}

	submitRecoveryCode := func(t *testing.T, client *http.Client, flow string, flowType ClientType, recoveryCode string, statusCode int) string {
		action := gjson.Get(flow, "ui.action").String()
		assert.NotEmpty(t, action)

		values := withCSRFToken(t, flowType, flow, url.Values{
			"code":   {recoveryCode},
			"method": {"code"},
		})

		contentType := "application/json"
		if flowType == RecoveryClientTypeBrowser {
			contentType = "application/x-www-form-urlencoded"
		}

		res, err := client.Post(action, contentType, bytes.NewBufferString(values))
		require.NoError(t, err)
		assert.Equal(t, statusCode, res.StatusCode)

		return string(ioutilx.MustReadAll(res.Body))
	}

	resendRecoveryCode := func(t *testing.T, client *http.Client, flow string, flowType ClientType, statusCode int) string {
		action := gjson.Get(flow, "ui.action").String()
		assert.NotEmpty(t, action)

		email := gjson.Get(flow, "ui.nodes.#(attributes.name==email).attributes.value").String()

		values := withCSRFToken(t, flowType, flow, url.Values{
			"method": {"code"},
			"email":  {email},
		})

		contentType := "application/json"
		if flowType == RecoveryClientTypeBrowser {
			contentType = "application/x-www-form-urlencoded"
		}

		res, err := client.Post(action, contentType, bytes.NewBufferString(values))
		require.NoError(t, err)
		assert.Equal(t, statusCode, res.StatusCode)

		return string(ioutilx.MustReadAll(res.Body))
	}

	expectValidationError := func(t *testing.T, hc *http.Client, flowType ClientType, values func(url.Values)) string {
		code := testhelpers.ExpectStatusCode(flowType == RecoveryClientTypeAPI || flowType == RecoveryClientTypeSPA, http.StatusBadRequest, http.StatusOK)
		return submitRecovery(t, hc, flowType, values, code)
	}

	expectSuccessfulRecovery := func(t *testing.T, hc *http.Client, flowType ClientType, values func(url.Values)) string {
		code := testhelpers.ExpectStatusCode(flowType == RecoveryClientTypeAPI || flowType == RecoveryClientTypeSPA, http.StatusUnprocessableEntity, http.StatusOK)
		return submitRecovery(t, hc, flowType, values, code)
	}

	ExpectVerfiableAddressStatus := func(t *testing.T, email string, status identity.VerifiableAddressStatus) {
		addr, err := reg.IdentityPool().
			FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, email)
		assert.NoError(t, err)
		assert.Equal(t, status, addr.Status, "verifiable address %s was not %s. instead %s", email, status, addr.Status)
	}

	t.Run("description=should recover an account", func(t *testing.T) {
		checkRecovery := func(t *testing.T, client *http.Client, flowType ClientType, recoveryEmail, recoverySubmissionResponse string) string {
			ExpectVerfiableAddressStatus(t, recoveryEmail, identity.VerifiableAddressStatusPending)

			assert.EqualValues(t, node.CodeGroup, gjson.Get(recoverySubmissionResponse, "active").String(), "%s", recoverySubmissionResponse)
			assert.True(t, gjson.Get(recoverySubmissionResponse, "ui.nodes.#(attributes.name==code)").Exists(), "%s", recoverySubmissionResponse)
			assert.Len(t, gjson.Get(recoverySubmissionResponse, "ui.messages").Array(), 1, "%s", recoverySubmissionResponse)
			assertx.EqualAsJSON(t, text.NewRecoveryEmailWithCodeSent(), json.RawMessage(gjson.Get(recoverySubmissionResponse, "ui.messages.0").Raw))

			message := testhelpers.CourierExpectMessage(ctx, t, reg, recoveryEmail, "Use code")
			assert.Contains(t, message.Body, "Recover access to your account by entering")

			recoveryCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
			assert.NotEmpty(t, recoveryCode)

			statusCode := testhelpers.ExpectStatusCode(flowType == RecoveryClientTypeAPI || flowType == RecoveryClientTypeSPA, http.StatusUnprocessableEntity, http.StatusOK)
			return submitRecoveryCode(t, client, recoverySubmissionResponse, flowType, recoveryCode, statusCode)
		}

		t.Run("type=browser", func(t *testing.T) {
			var wg sync.WaitGroup
			wg.Add(1)
			testhelpers.NewRecoveryAfterHookWebHookTarget(ctx, t, conf, func(t *testing.T, msg []byte) {
				defer wg.Done()
				assert.EqualValues(t, "recoverme1@ory.sh", gjson.GetBytes(msg, "identity.verifiable_addresses.0.value").String(), string(msg))
				assert.EqualValues(t, true, gjson.GetBytes(msg, "identity.verifiable_addresses.0.verified").Bool(), string(msg))
				assert.EqualValues(t, "completed", gjson.GetBytes(msg, "identity.verifiable_addresses.0.status").String(), string(msg))
			})

			client := testhelpers.NewClientWithCookies(t)
			email := "recoverme1@ory.sh"
			createIdentityToRecover(t, reg, email)
			recoverySubmissionResponse := submitRecovery(t, client, RecoveryClientTypeBrowser, func(v url.Values) {
				v.Set("email", email)
			}, http.StatusOK)
			body := checkRecovery(t, client, RecoveryClientTypeBrowser, email, recoverySubmissionResponse)

			assert.Equal(t, text.NewRecoverySuccessful(time.Now().Add(time.Hour)).Text,
				gjson.Get(body, "ui.messages.0.text").String())

			res, err := client.Get(public.URL + session.RouteWhoami)
			require.NoError(t, err)
			body = string(x.MustReadAll(res.Body))
			require.NoError(t, res.Body.Close())
			assert.Equal(t, "code_recovery", gjson.Get(body, "authentication_methods.0.method").String(), "%s", body)
			assert.Equal(t, "aal1", gjson.Get(body, "authenticator_assurance_level").String(), "%s", body)

			wg.Wait()
		})

		t.Run("type=spa", func(t *testing.T) {
			client := testhelpers.NewClientWithCookies(t)
			email := "recoverme3@ory.sh"
			createIdentityToRecover(t, reg, email)
			recoverySubmissionResponse := submitRecovery(t, client, RecoveryClientTypeSPA, func(v url.Values) {
				v.Set("email", email)
			}, http.StatusOK)
			body := checkRecovery(t, client, RecoveryClientTypeSPA, email, recoverySubmissionResponse)
			assert.Equal(t, "browser_location_change_required", gjson.Get(body, "error.id").String())
			assert.Contains(t, gjson.Get(body, "redirect_browser_to").String(), "settings-ts?")
		})

		t.Run("type=api", func(t *testing.T) {
			client := &http.Client{}
			email := "recoverme4@ory.sh"
			createIdentityToRecover(t, reg, email)
			recoverySubmissionResponse := submitRecovery(t, client, RecoveryClientTypeAPI, func(v url.Values) {
				v.Set("email", email)
			}, http.StatusOK)
			body := checkRecovery(t, client, RecoveryClientTypeAPI, email, recoverySubmissionResponse)
			assert.Equal(t, "browser_location_change_required", gjson.Get(body, "error.id").String())
			assert.Contains(t, gjson.Get(body, "redirect_browser_to").String(), "settings-ts?")
		})

		t.Run("description=should pass transient data to email template and webhooks", func(t *testing.T) {
			webhookTS := hooktest.NewServer()
			t.Cleanup(webhookTS.Close)

			conf.MustSet(ctx, "selfservice.flows.recovery.after.hooks", []config.SelfServiceHook{webhookTS.HookConfig()})
			t.Cleanup(func() { conf.MustSet(ctx, "selfservice.flows.recovery.after.hooks", nil) })

			client := testhelpers.NewClientWithCookies(t)
			email := testhelpers.RandomEmail()
			createIdentityToRecover(t, reg, email)
			templatePayload := `{"payload":"template data"}`
			webhookPayload := `{"payload":"webhook data"}`

			f := testhelpers.InitializeRecoveryFlowViaBrowser(t, client, false, public, nil)

			formPayload := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
			formPayload.Set("email", email)
			formPayload.Set("transient_payload", templatePayload)

			body, _ := testhelpers.RecoveryMakeRequest(t, false, f, client, formPayload.Encode())
			message := testhelpers.CourierExpectMessage(ctx, t, reg, email, "Use code")
			assert.Equal(t, templatePayload, gjson.GetBytes(message.TemplateData, "transient_payload").String(),
				"should pass transient payload to email template")

			recoveryCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
			assert.NotEmpty(t, recoveryCode)

			action := gjson.Get(body, "ui.action").String()
			assert.NotEmpty(t, action)

			_, err := client.Post(action, "application/x-www-form-urlencoded", bytes.NewBufferString(
				withCSRFToken(t, RecoveryClientTypeBrowser, body, url.Values{
					"code":              {recoveryCode},
					"method":            {"code"},
					"transient_payload": {webhookPayload},
				})))
			require.NoError(t, err)

			assert.JSONEq(t, webhookPayload, gjson.GetBytes(webhookTS.LastBody, "flow.transient_payload").String(),
				"should pass transient payload to webhook")
		})

		t.Run("description=should return browser to return url", func(t *testing.T) {
			returnTo := public.URL + "/return-to"
			conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{returnTo})

			for _, tc := range []struct {
				desc        string
				returnTo    string
				f           func(t *testing.T, client *http.Client, identity *identity.Identity) *kratos.RecoveryFlow
				expectedAAL string
			}{
				{
					desc:     "should use return_to from recovery flow",
					returnTo: returnTo,
					f: func(t *testing.T, client *http.Client, identity *identity.Identity) *kratos.RecoveryFlow {
						return testhelpers.InitializeRecoveryFlowViaBrowser(t, client, false, public, url.Values{"return_to": []string{returnTo}})
					},
				},
				{
					desc:     "should use return_to from config",
					returnTo: returnTo,
					f: func(t *testing.T, client *http.Client, identity *identity.Identity) *kratos.RecoveryFlow {
						conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryBrowserDefaultReturnTo, returnTo)
						t.Cleanup(func() {
							conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryBrowserDefaultReturnTo, "")
						})
						return testhelpers.InitializeRecoveryFlowViaBrowser(t, client, false, public, nil)
					},
				},
				{
					desc:     "no return to",
					returnTo: "",
					f: func(t *testing.T, client *http.Client, identity *identity.Identity) *kratos.RecoveryFlow {
						return testhelpers.InitializeRecoveryFlowViaBrowser(t, client, false, public, nil)
					},
				},
				{
					desc:     "should use return_to with an account that has 2fa enabled",
					returnTo: returnTo,
					f: func(t *testing.T, client *http.Client, id *identity.Identity) *kratos.RecoveryFlow {
						conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, config.HighestAvailableAAL)
						conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, config.HighestAvailableAAL)
						conf.MustSet(ctx, config.ViperKeyWebAuthnRPDisplayName, "Kratos")
						conf.MustSet(ctx, config.ViperKeyWebAuthnRPID, "ory.sh")

						t.Cleanup(func() {
							conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, identity.AuthenticatorAssuranceLevel1)
							conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, identity.AuthenticatorAssuranceLevel1)
						})
						testhelpers.StrategyEnable(t, conf, identity.CredentialsTypeWebAuthn.String(), true)

						id.SetCredentials(identity.CredentialsTypeWebAuthn, identity.Credentials{
							Type:        identity.CredentialsTypeWebAuthn,
							Config:      []byte(`{"credentials":[{"is_passwordless":false, "display_name":"test"}]}`),
							Identifiers: []string{testhelpers.RandomEmail()},
						})

						require.NoError(t, reg.IdentityManager().Update(ctx, id, identity.ManagerAllowWriteProtectedTraits))
						return testhelpers.InitializeRecoveryFlowViaBrowser(t, client, false, public, url.Values{"return_to": []string{returnTo}})
					},
					expectedAAL: "aal2",
				},
			} {
				t.Run(tc.desc, func(t *testing.T) {
					client := testhelpers.NewClientWithCookies(t)
					email := testhelpers.RandomEmail()
					i := createIdentityToRecover(t, reg, email)

					client.Transport = testhelpers.NewTransportWithLogger(http.DefaultTransport, t).RoundTripper
					f := tc.f(t, client, i)

					formPayload := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
					formPayload.Set("email", email)

					body, res := testhelpers.RecoveryMakeRequest(t, false, f, client, formPayload.Encode())
					assert.EqualValues(t, http.StatusOK, res.StatusCode, "%s", body)
					expectedURL := testhelpers.ExpectURL(false, public.URL+recovery.RouteSubmitFlow, conf.SelfServiceFlowRecoveryUI(ctx).String())
					assert.Contains(t, res.Request.URL.String(), expectedURL, "%+v\n\t%s", res.Request, body)

					body = checkRecovery(t, client, RecoveryClientTypeBrowser, email, body)

					require.Equal(t, text.NewRecoverySuccessful(time.Now().Add(time.Hour)).Text,
						gjson.Get(body, "ui.messages.0.text").String())

					settingsId := gjson.Get(body, "id").String()

					sf, err := reg.SettingsFlowPersister().GetSettingsFlow(ctx, uuid.Must(uuid.FromString(settingsId)))
					require.NoError(t, err)

					u, err := url.Parse(public.URL)
					require.NoError(t, err)
					require.Len(t, client.Jar.Cookies(u), 2)
					found := false
					for _, cookie := range client.Jar.Cookies(u) {
						if cookie.Name == "ory_kratos_session" {
							found = true
						}
					}
					require.True(t, found)

					require.Equal(t, tc.returnTo, sf.ReturnTo)
					res, err = client.Get(public.URL + session.RouteWhoami)
					require.NoError(t, err)
					body = string(x.MustReadAll(res.Body))
					require.NoError(t, res.Body.Close())

					if tc.expectedAAL == "aal2" {
						require.Equal(t, http.StatusForbidden, res.StatusCode)
						require.Equalf(t, session.NewErrAALNotSatisfied("").Reason(), gjson.Get(body, "error.reason").String(), "%s", body)
						require.Equalf(t, "session_aal2_required", gjson.Get(body, "error.id").String(), "%s", body)
					} else {
						assert.Equal(t, "code_recovery", gjson.Get(body, "authentication_methods.0.method").String(), "%s", body)
						assert.Equal(t, "aal1", gjson.Get(body, "authenticator_assurance_level").String(), "%s", body)
					}
				})
			}
		})
	})

	t.Run("description=should set all the correct recovery payloads after submission", func(t *testing.T) {
		body := expectSuccessfulRecovery(t, nil, RecoveryClientTypeBrowser, func(v url.Values) {
			v.Set("email", "test@ory.sh")
		})
		testhelpers.SnapshotTExcept(t, json.RawMessage(gjson.Get(body, "ui.nodes").String()), []string{"0.attributes.value"})
	})

	t.Run("description=should set all the correct recovery payloads", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)
		rs := testhelpers.GetRecoveryFlow(t, c, public)

		testhelpers.SnapshotTExcept(t, rs.Ui.Nodes, []string{"0.attributes.value"})
		assert.EqualValues(t, public.URL+recovery.RouteSubmitFlow+"?flow="+rs.Id, rs.Ui.Action)
		assert.Empty(t, rs.Ui.Messages)
	})

	t.Run("description=should require an email to be sent", func(t *testing.T) {
		for _, flowType := range flowTypes {
			t.Run("type="+string(flowType), func(t *testing.T) {
				body := expectValidationError(t, nil, flowType, func(v url.Values) {
					v.Del("email")
				})
				assert.EqualValues(t, node.CodeGroup, gjson.Get(body, "active").String(), "%s", body)
				assert.EqualValues(t, "Property email is missing.",
					gjson.Get(body, "ui.nodes.#(attributes.name==email).messages.0.text").String(),
					"%s", body)
			})
		}
	})

	t.Run("description=should require a valid email to be sent", func(t *testing.T) {
		for _, flowType := range flowTypes {
			for _, email := range []string{"\\", "asdf", "...", "aiacobelli.sec@gmail.com,alejandro.iacobelli@mercadolibre.com"} {
				t.Run("type="+string(flowType), func(t *testing.T) {
					responseJSON := expectValidationError(t, nil, flowType, func(v url.Values) {
						v.Set("email", email)
					})
					activeMethod := gjson.Get(responseJSON, "active").String()
					assert.EqualValues(t, node.CodeGroup, activeMethod, "expected method to be %s got %s", node.CodeGroup, activeMethod)
					expectedMessage := fmt.Sprintf("%q is not valid \"email\"", email)
					actualMessage := gjson.Get(responseJSON, "ui.nodes.#(attributes.name==email).messages.0.text").String()
					assert.EqualValues(t, expectedMessage, actualMessage, "%s", responseJSON)
				})
			}
		}
	})

	t.Run("description=should try to submit the form while authenticated", func(t *testing.T) {
		for _, flowType := range flowTypes {
			t.Run("type="+string(flowType), func(t *testing.T) {
				isSPA := flowType == "spa"
				isAPI := flowType == "api"
				client := testhelpers.NewDebugClient(t)
				if !isAPI {
					client = testhelpers.NewClientWithCookies(t)
					client.Transport = testhelpers.NewTransportWithLogger(http.DefaultTransport, t).RoundTripper
				}

				var f *kratos.RecoveryFlow
				if isAPI {
					f = testhelpers.InitializeRecoveryFlowViaAPI(t, client, public)
				} else {
					f = testhelpers.InitializeRecoveryFlowViaBrowser(t, client, isSPA, public, nil)
				}
				req := httptest.NewRequest("GET", "/sessions/whoami", nil).WithContext(contextx.WithConfigValue(ctx, config.ViperKeySessionLifespan, time.Hour))
				session, err := testhelpers.NewActiveSession(req,
					reg,
					&identity.Identity{ID: x.NewUUID(), State: identity.StateActive, NID: x.NewUUID()},
					time.Now(),
					identity.CredentialsTypePassword,
					identity.AuthenticatorAssuranceLevel1,
				)

				require.NoError(t, err)

				// Add the authentication to the request
				client.Transport = testhelpers.NewTransportWithLogger(testhelpers.NewAuthorizedTransport(t, ctx, reg, session), t).RoundTripper

				v := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
				v.Set("email", "some-email@example.org")
				v.Set("method", "code")

				body, res := testhelpers.RecoveryMakeRequest(t, isAPI || isSPA, f, client, testhelpers.EncodeFormAsJSON(t, isAPI || isSPA, v))

				if isAPI || isSPA {
					assert.EqualValues(t, http.StatusBadRequest, res.StatusCode, "%s", body)
					assert.Contains(t, res.Request.URL.String(), recovery.RouteSubmitFlow, "%+v\n\t%s", res.Request, body)
					assertx.EqualAsJSONExcept(t, recovery.ErrAlreadyLoggedIn, json.RawMessage(gjson.Get(body, "error").Raw), nil)
				} else {
					assert.EqualValues(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.Contains(t, res.Request.URL.String(), conf.SelfServiceBrowserDefaultReturnTo(ctx).String(), "%+v\n\t%s", res.Request, body)
				}
			})
		}
	})

	t.Run("description=should not be able to recover account that does not exist", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryNotifyUnknownRecipients, true)

		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryNotifyUnknownRecipients, false)
		})

		check := func(t *testing.T, c *http.Client, flowType ClientType, email string) {
			withValues := func(v url.Values) {
				v.Set("email", email)
			}
			body := submitRecovery(t, c, flowType, withValues, http.StatusOK)
			assert.EqualValues(t, node.CodeGroup, gjson.Get(body, "active").String(), "%s", body)
			assert.Empty(t, gjson.Get(body, "ui.nodes.#(attributes.name==code).attributes.value").String(), "%s", body)
			assertx.EqualAsJSON(t, text.NewRecoveryEmailWithCodeSent(), json.RawMessage(gjson.Get(body, "ui.messages.0").Raw))

			message := testhelpers.CourierExpectMessage(ctx, t, reg, email, "Account access attempted")
			assert.Contains(t, message.Body, "If this was you, check if you signed up using a different address.")
		}

		t.Run("type=browser", func(t *testing.T) {
			email := "recover_browser@ory.sh"
			c := browserHttpClient(t)
			check(t, c, RecoveryClientTypeBrowser, email)
		})

		t.Run("type=spa", func(t *testing.T) {
			email := "recover_spa@ory.sh"
			c := spaHttpClient(t)
			check(t, c, RecoveryClientTypeSPA, email)
		})

		t.Run("type=api", func(t *testing.T) {
			email := "recover_api@ory.sh"
			c := apiHttpClient(t)
			check(t, c, RecoveryClientTypeAPI, email)
		})
	})

	t.Run("description=should not be able to recover an inactive account", func(t *testing.T) {
		for _, flowType := range flowTypeCases {
			t.Run("type="+string(flowType.ClientType), func(t *testing.T) {
				email := "recoverinactive_" + string(flowType.ClientType) + "@ory.sh"
				createIdentityToRecover(t, reg, email)
				values := func(v url.Values) {
					v.Set("email", email)
				}
				cl := testhelpers.NewClientWithCookies(t)

				body := submitRecovery(t, cl, flowType.ClientType, values, http.StatusOK)
				addr, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, email)
				assert.NoError(t, err)

				emailText := testhelpers.CourierExpectMessage(ctx, t, reg, email, "Use code")
				recoveryCode := testhelpers.CourierExpectCodeInMessage(t, emailText, 1)

				// Deactivate the identity
				require.NoError(t, reg.Persister().GetConnection(context.Background()).RawQuery("UPDATE identities SET state=? WHERE id = ?", identity.StateInactive, addr.IdentityID).Exec())

				if flowType.ClientType == RecoveryClientTypeAPI || flowType.ClientType == RecoveryClientTypeSPA {
					body = submitRecoveryCode(t, cl, body, flowType.ClientType, recoveryCode, http.StatusUnauthorized)
					assertx.EqualAsJSON(t, session.ErrIdentityDisabled.WithDetail("identity_id", addr.IdentityID), json.RawMessage(gjson.Get(body, "error").Raw), "%s", body)
				} else {
					body = submitRecoveryCode(t, cl, body, flowType.ClientType, recoveryCode, http.StatusOK)
					assertx.EqualAsJSON(t, session.ErrIdentityDisabled.WithDetail("identity_id", addr.IdentityID), json.RawMessage(body), "%s", body)
				}
			})
		}
	})

	t.Run("description=should recover and invalidate all other sessions if hook is set", func(t *testing.T) {
		conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal), []config.SelfServiceHook{{Name: "revoke_active_sessions"}})
		t.Cleanup(func() {
			conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()), nil)
		})

		email := testhelpers.RandomEmail()
		id := createIdentityToRecover(t, reg, email)

		req := httptest.NewRequest("GET", "/sessions/whoami", nil)
		sess, err := testhelpers.NewActiveSession(req, reg, id, time.Now(), identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
		require.NoError(t, err)
		require.NoError(t, reg.SessionPersister().UpsertSession(context.Background(), sess))

		actualSession, err := reg.SessionPersister().GetSession(context.Background(), sess.ID, session.ExpandNothing)
		require.NoError(t, err)
		assert.True(t, actualSession.IsActive())

		cl := testhelpers.NewClientWithCookies(t)
		actual := expectSuccessfulRecovery(t, cl, RecoveryClientTypeBrowser, func(v url.Values) {
			v.Set("email", email)
		})
		message := testhelpers.CourierExpectMessage(ctx, t, reg, email, "Use code")
		recoveryCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)

		cl.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}

		action := gjson.Get(actual, "ui.action").String()
		require.NotEmpty(t, action)
		csrf_token := gjson.Get(actual, "ui.nodes.#(attributes.name==csrf_token).attributes.value").String()
		require.NotEmpty(t, csrf_token)

		submitRecoveryCode(t, cl, actual, RecoveryClientTypeBrowser, recoveryCode, http.StatusSeeOther)

		require.Len(t, cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 2)
		cookies := spew.Sdump(cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)))
		assert.Contains(t, cookies, "ory_kratos_session")

		actualSession, err = reg.SessionPersister().GetSession(context.Background(), sess.ID, session.ExpandNothing)
		require.NoError(t, err)
		assert.False(t, actualSession.IsActive())
	})

	t.Run("description=should not be able to use an invalid code more than 5 times", func(t *testing.T) {
		email := testhelpers.RandomEmail()
		createIdentityToRecover(t, reg, email)
		c := testhelpers.NewClientWithCookies(t)
		body := submitRecovery(t, c, RecoveryClientTypeBrowser, func(v url.Values) {
			v.Set("email", email)
		}, http.StatusOK)
		initialFlowId := gjson.Get(body, "id")

		for submitTry := 0; submitTry < 5; submitTry++ {
			body := submitRecoveryCode(t, c, body, RecoveryClientTypeBrowser, "12312312", http.StatusOK)

			testhelpers.AssertMessage(t, []byte(body), "The recovery code is invalid or has already been used. Please try again.")
		}

		// submit an invalid code for the 6th time
		body = submitRecoveryCode(t, c, body, RecoveryClientTypeBrowser, "12312312", http.StatusOK)

		require.Len(t, gjson.Get(body, "ui.messages").Array(), 1)
		assert.Equal(t, "The request was submitted too often. Please request another code.", gjson.Get(body, "ui.messages.0.text").String())

		// check that a new flow has been created
		assert.NotEqual(t, gjson.Get(body, "id"), initialFlowId)

		assert.True(t, gjson.Get(body, "ui.nodes.#(attributes.name==email)").Exists())
	})

	t.Run("description=should be able to recover after using invalid code", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				c := testCase.GetClient(t)
				recoveryEmail := testhelpers.RandomEmail()
				_ = createIdentityToRecover(t, reg, recoveryEmail)

				actual := submitRecovery(t, c, testCase.ClientType, func(v url.Values) {
					v.Set("email", recoveryEmail)
				}, http.StatusOK)

				message := testhelpers.CourierExpectMessage(ctx, t, reg, recoveryEmail, "Use code")
				recoveryCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)

				form := withCSRFToken(t, testCase.ClientType, actual, url.Values{
					"code": {"12312312"},
				})

				action := gjson.Get(actual, "ui.action").String()
				require.NotEmpty(t, action)

				res, err := c.Post(action, testCase.FormContentType, bytes.NewBufferString(form))
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, res.StatusCode)

				flowId := gjson.Get(actual, "id").String()
				require.NotEmpty(t, flowId)

				rs, res, err := testhelpers.
					NewSDKCustomClient(public, c).
					FrontendAPI.GetRecoveryFlow(context.Background()).
					Id(flowId).
					Execute()

				require.NoError(t, err)
				body := ioutilx.MustReadAll(res.Body)
				require.NotEmpty(t, body)

				require.Len(t, rs.Ui.Messages, 1)
				assert.Equal(t, "The recovery code is invalid or has already been used. Please try again.", rs.Ui.Messages[0].Text)

				form = withCSRFToken(t, testCase.ClientType, actual, url.Values{
					"code": {recoveryCode},
				})
				// Now submit the correct code
				res, err = c.Post(action, testCase.FormContentType, bytes.NewBufferString(form))
				require.NoError(t, err)
				switch testCase.ClientType {
				case RecoveryClientTypeBrowser:
					assert.Equal(t, http.StatusOK, res.StatusCode)

					json := ioutilx.MustReadAll(res.Body)

					assert.Len(t, gjson.GetBytes(json, "ui.messages").Array(), 1)
					assert.Contains(t, gjson.GetBytes(json, "ui.messages.0.text").String(), "You successfully recovered your account.")
				case RecoveryClientTypeSPA:
					assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

					json := ioutilx.MustReadAll(res.Body)

					assert.Equal(t, gjson.GetBytes(json, "error.id").String(), "browser_location_change_required")
					assert.Contains(t, gjson.GetBytes(json, "redirect_browser_to").String(), "settings-ts?")
				}
			})
		}
	})

	t.Run("description=should not be able to use an invalid code", func(t *testing.T) {
		email := "recoverme+invalid_code@ory.sh"
		createIdentityToRecover(t, reg, email)
		c := testhelpers.NewClientWithCookies(t)

		body := submitRecovery(t, c, RecoveryClientTypeBrowser, func(v url.Values) {
			v.Set("email", email)
		}, http.StatusOK)

		body = submitRecoveryCode(t, c, body, RecoveryClientTypeBrowser, "12312312", http.StatusOK)

		testhelpers.AssertMessage(t, []byte(body), "The recovery code is invalid or has already been used. Please try again.")
	})

	t.Run("description=should not be able to submit recover address after flow expired", func(t *testing.T) {
		recoveryEmail := "recoverme5@ory.sh"
		createIdentityToRecover(t, reg, recoveryEmail)
		conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryRequestLifespan, time.Millisecond*200)
		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryRequestLifespan, time.Minute)
		})

		c := testhelpers.NewClientWithCookies(t)
		rs := testhelpers.GetRecoveryFlow(t, c, public)

		time.Sleep(time.Millisecond * 201)

		res, err := c.PostForm(rs.Ui.Action, url.Values{"email": {recoveryEmail}})
		require.NoError(t, err)
		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.NotContains(t, res.Request.URL.String(), "flow="+rs.Id)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowRecoveryUI(ctx).String())

		addr, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, recoveryEmail)
		assert.NoError(t, err)
		assert.False(t, addr.Verified)
		assert.Nil(t, addr.VerifiedAt)
		assert.Equal(t, identity.VerifiableAddressStatusPending, addr.Status)
	})

	t.Run("description=should not be able to submit code after flow expired", func(t *testing.T) {
		recoveryEmail := "recoverme6@ory.sh"
		createIdentityToRecover(t, reg, recoveryEmail)
		conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryRequestLifespan, time.Millisecond*200)
		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryRequestLifespan, time.Minute)
		})

		c := testhelpers.NewClientWithCookies(t)

		body := expectSuccessfulRecovery(t, c, RecoveryClientTypeBrowser, func(v url.Values) {
			v.Set("email", recoveryEmail)
		})

		initialFlowId := gjson.Get(body, "id")

		message := testhelpers.CourierExpectMessage(ctx, t, reg, recoveryEmail, "Use code")
		assert.Contains(t, message.Body, "Recover access to your account by entering")

		recoveryCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)

		time.Sleep(time.Millisecond * 201)

		body = submitRecoveryCode(t, c, body, RecoveryClientTypeBrowser, recoveryCode, http.StatusOK)

		assert.NotEqual(t, gjson.Get(body, "id"), initialFlowId)

		testhelpers.AssertMessage(t, []byte(body), "The recovery flow expired 0.00 minutes ago, please try again.")

		addr, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, recoveryEmail)
		require.NoError(t, err)
		assert.False(t, addr.Verified)
		assert.Nil(t, addr.VerifiedAt)
		assert.Equal(t, identity.VerifiableAddressStatusPending, addr.Status)
	})

	t.Run("description=should not break ui if empty code is submitted", func(t *testing.T) {
		recoveryEmail := "recoverme7@ory.sh"
		createIdentityToRecover(t, reg, recoveryEmail)

		c := testhelpers.NewClientWithCookies(t)
		body := expectSuccessfulRecovery(t, c, RecoveryClientTypeBrowser, func(v url.Values) {
			v.Set("email", recoveryEmail)
		})

		action := gjson.Get(body, "ui.action").String()
		require.NotEmpty(t, action)

		body = submitRecoveryCode(t, c, body, RecoveryClientTypeBrowser, "", http.StatusOK)

		assert.NotContains(t, gjson.Get(body, "ui.nodes").String(), "Property email is missing.")
		testhelpers.AssertMessage(t, []byte(body), "The recovery code is invalid or has already been used. Please try again.")
	})

	t.Run("description=should be able to re-send the recovery code", func(t *testing.T) {
		recoveryEmail := testhelpers.RandomEmail()
		createIdentityToRecover(t, reg, recoveryEmail)

		c := testhelpers.NewClientWithCookies(t)
		body := expectSuccessfulRecovery(t, c, RecoveryClientTypeBrowser, func(v url.Values) {
			v.Set("email", recoveryEmail)
		})

		action := gjson.Get(body, "ui.action").String()
		require.NotEmpty(t, action)
		assert.Equal(t, recoveryEmail, gjson.Get(body, "ui.nodes.#(attributes.name==email).attributes.value").String())

		body = resendRecoveryCode(t, c, body, RecoveryClientTypeBrowser, http.StatusOK)
		assert.True(t, gjson.Get(body, "ui.nodes.#(attributes.name==code)").Exists())
		assert.Equal(t, recoveryEmail, gjson.Get(body, "ui.nodes.#(attributes.name==email).attributes.value").String())

		message := testhelpers.CourierExpectMessage(ctx, t, reg, recoveryEmail, "Use code")
		recoveryCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)

		submitRecoveryCode(t, c, body, RecoveryClientTypeBrowser, recoveryCode, http.StatusOK)
	})

	t.Run("description=should not be able to use first code after re-sending email", func(t *testing.T) {
		recoveryEmail := testhelpers.RandomEmail()
		createIdentityToRecover(t, reg, recoveryEmail)

		c := testhelpers.NewClientWithCookies(t)
		body := expectSuccessfulRecovery(t, c, RecoveryClientTypeBrowser, func(v url.Values) {
			v.Set("email", recoveryEmail)
		})

		action := gjson.Get(body, "ui.action").String()
		require.NotEmpty(t, action)
		assert.Equal(t, recoveryEmail, gjson.Get(body, "ui.nodes.#(attributes.name==email).attributes.value").String())

		message1 := testhelpers.CourierExpectMessage(ctx, t, reg, recoveryEmail, "Use code")
		recoveryCode1 := testhelpers.CourierExpectCodeInMessage(t, message1, 1)

		body = resendRecoveryCode(t, c, body, RecoveryClientTypeBrowser, http.StatusOK)
		assert.True(t, gjson.Get(body, "ui.nodes.#(attributes.name==code)").Exists())
		assert.Equal(t, recoveryEmail, gjson.Get(body, "ui.nodes.#(attributes.name==email).attributes.value").String())

		message2 := testhelpers.CourierExpectMessage(ctx, t, reg, recoveryEmail, "Use code")
		recoveryCode2 := testhelpers.CourierExpectCodeInMessage(t, message2, 1)

		body = submitRecoveryCode(t, c, body, RecoveryClientTypeBrowser, recoveryCode1, http.StatusOK)
		testhelpers.AssertMessage(t, []byte(body), "The recovery code is invalid or has already been used. Please try again.")

		// For good measure, check that the second code works!
		body = submitRecoveryCode(t, c, body, RecoveryClientTypeBrowser, recoveryCode2, http.StatusOK)
		testhelpers.AssertMessage(t, []byte(body), "You successfully recovered your account. Please change your password or set up an alternative login method (e.g. social sign in) within the next 60.00 minutes.")
	})

	t.Run("description=should not show outdated validation message if newer message appears #2799", func(t *testing.T) {
		recoveryEmail := testhelpers.RandomEmail()
		createIdentityToRecover(t, reg, recoveryEmail)

		c := testhelpers.NewClientWithCookies(t)
		body := expectSuccessfulRecovery(t, c, RecoveryClientTypeBrowser, func(v url.Values) {
			v.Set("email", recoveryEmail)
		})

		action := gjson.Get(body, "ui.action").String()
		require.NotEmpty(t, action)
		assert.Equal(t, recoveryEmail, gjson.Get(body, "ui.nodes.#(attributes.name==email).attributes.value").String())

		body = submitRecoveryCode(t, c, body, RecoveryClientTypeBrowser, "12312312", http.StatusOK) // Now send a wrong code that triggers "global" validation error

		assert.Empty(t, gjson.Get(body, "ui.nodes.#(attributes.name==code).messages").Array())
		testhelpers.AssertMessage(t, []byte(body), "The recovery code is invalid or has already been used. Please try again.")
	})

	t.Run("description=should recover if post recovery hook is successful", func(t *testing.T) {
		conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal), []config.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})
		t.Cleanup(func() {
			conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal), nil)
		})

		recoveryEmail := testhelpers.RandomEmail()
		createIdentityToRecover(t, reg, recoveryEmail)

		cl := testhelpers.NewClientWithCookies(t)
		body := expectSuccessfulRecovery(t, cl, RecoveryClientTypeBrowser, func(v url.Values) {
			v.Set("email", recoveryEmail)
		})

		message := testhelpers.CourierExpectMessage(ctx, t, reg, recoveryEmail, "Use code")
		recoveryCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)

		action := gjson.Get(body, "ui.action").String()
		require.NotEmpty(t, action)
		assert.Equal(t, recoveryEmail, gjson.Get(body, "ui.nodes.#(attributes.name==email).attributes.value").String())

		cl.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}

		submitRecoveryCode(t, cl, body, RecoveryClientTypeBrowser, recoveryCode, http.StatusSeeOther)

		require.Len(t, cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 2)
		cookies := spew.Sdump(cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)))
		assert.Contains(t, cookies, "ory_kratos_session")
	})

	t.Run("description=should not be able to recover if post recovery hook fails", func(t *testing.T) {
		conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal), []config.SelfServiceHook{{Name: "err", Config: []byte(`{"ExecutePostRecoveryHook": "err"}`)}})
		t.Cleanup(func() {
			conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal), nil)
		})

		recoveryEmail := testhelpers.RandomEmail()
		createIdentityToRecover(t, reg, recoveryEmail)

		cl := testhelpers.NewClientWithCookies(t)
		body := expectSuccessfulRecovery(t, cl, RecoveryClientTypeBrowser, func(v url.Values) {
			v.Set("email", recoveryEmail)
		})

		message := testhelpers.CourierExpectMessage(ctx, t, reg, recoveryEmail, "Use code")
		recoveryCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)

		action := gjson.Get(body, "ui.action").String()
		require.NotEmpty(t, action)
		assert.Equal(t, recoveryEmail, gjson.Get(body, "ui.nodes.#(attributes.name==email).attributes.value").String())

		cl.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}

		initialFlowId := gjson.Get(body, "id")
		body = submitRecoveryCode(t, cl, body, RecoveryClientTypeBrowser, recoveryCode, http.StatusSeeOther)
		assert.NotEqual(t, gjson.Get(body, "id"), initialFlowId)

		require.Len(t, cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 1) // No session
		cookies := spew.Sdump(cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)))
		assert.NotContains(t, cookies, "ory_kratos_session")
	})
}

func TestRecovery_WithContinueWith(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.StrategyEnable(t, conf, string(recovery.RecoveryStrategyCode), true)
	testhelpers.StrategyEnable(t, conf, string(recovery.RecoveryStrategyLink), false)
	conf.MustSet(ctx, config.ViperKeyUseContinueWithTransitions, true)

	initViper(t, ctx, conf)

	_ = testhelpers.NewRecoveryUIFlowEchoServer(t, reg)
	_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
	_ = testhelpers.NewSettingsUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)

	public, _, _, _ := testhelpers.NewKratosServerWithCSRFAndRouters(t, reg)

	submitRecoveryForm := func(t *testing.T, client *http.Client, clientType ClientType, values func(url.Values), code int) string {
		isSPA := clientType == RecoveryClientTypeSPA
		isAPI := clientType == RecoveryClientTypeAPI
		if client == nil {
			client = testhelpers.NewDebugClient(t)
			if !isAPI {
				client = testhelpers.NewClientWithCookies(t)
				client.Transport = testhelpers.NewTransportWithLogger(http.DefaultTransport, t).RoundTripper
			}
		}

		expectedUrl := testhelpers.ExpectURL(isAPI || isSPA, public.URL+recovery.RouteSubmitFlow, conf.SelfServiceFlowRecoveryUI(ctx).String())
		return testhelpers.SubmitRecoveryForm(t, isAPI, isSPA, client, public, values, code, expectedUrl)
	}

	submitRecoveryCode := func(t *testing.T, client *http.Client, flow string, flowType ClientType, recoveryCode string, statusCode int) string {
		t.Helper()
		action := gjson.Get(flow, "ui.action").String()
		assert.NotEmpty(t, action)

		values := withCSRFToken(t, flowType, flow, url.Values{
			"code":   {recoveryCode},
			"method": {"code"},
		})

		contentType := "application/json"
		if flowType == RecoveryClientTypeBrowser {
			contentType = "application/x-www-form-urlencoded"
		}

		res, err := client.Post(action, contentType, bytes.NewBufferString(values))
		require.NoError(t, err)
		assert.Equal(t, statusCode, res.StatusCode)

		return string(ioutilx.MustReadAll(res.Body))
	}

	resendRecoveryCode := func(t *testing.T, client *http.Client, flow string, flowType ClientType, statusCode int) string {
		action := gjson.Get(flow, "ui.action").String()
		assert.NotEmpty(t, action)

		email := gjson.Get(flow, "ui.nodes.#(attributes.name==email).attributes.value").String()

		values := withCSRFToken(t, flowType, flow, url.Values{
			"method": {"code"},
			"email":  {email},
		})

		contentType := "application/json"
		if flowType == RecoveryClientTypeBrowser {
			contentType = "application/x-www-form-urlencoded"
		}

		res, err := client.Post(action, contentType, bytes.NewBufferString(values))
		require.NoError(t, err)
		assert.Equal(t, statusCode, res.StatusCode)

		return string(ioutilx.MustReadAll(res.Body))
	}

	expectValidationError := func(t *testing.T, hc *http.Client, flowType ClientType, values func(url.Values)) string {
		code := testhelpers.ExpectStatusCode(flowType == RecoveryClientTypeAPI || flowType == RecoveryClientTypeSPA, http.StatusBadRequest, http.StatusOK)
		return submitRecoveryForm(t, hc, flowType, values, code)
	}

	expectVerfiableAddressStatus := func(t *testing.T, email string, status identity.VerifiableAddressStatus) {
		addr, err := reg.IdentityPool().
			FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, email)
		assert.NoError(t, err)
		assert.Equal(t, status, addr.Status, "verifiable address %s was not %s. instead %s", email, status, addr.Status)
	}

	submitCodeAndExpectRedirectToSettings := func(t *testing.T, c *http.Client, clientType ClientType, recoveryCode, body string) {
		t.Helper()
		switch clientType {
		case RecoveryClientTypeBrowser:
			body = submitRecoveryCode(t, c, body, clientType, recoveryCode, http.StatusOK)
			require.Len(t, c.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 2)
			cookies := spew.Sdump(c.Jar.Cookies(urlx.ParseOrPanic(public.URL)))
			assert.Contains(t, cookies, "ory_kratos_session")
			require.Contains(t, body, "You successfully recovered your account. Please change your password or set up an alternative login method (e.g. social sign in) within the next 60.00 minutes.")
		case RecoveryClientTypeSPA:
			body = submitRecoveryCode(t, c, body, clientType, recoveryCode, http.StatusOK)
			// assert.Equal(t, "browser_location_change_required", gjson.Get(body, "error.id").String())
			require.Len(t, c.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 2)
			cookies := spew.Sdump(c.Jar.Cookies(urlx.ParseOrPanic(public.URL)))
			assert.Contains(t, cookies, "ory_kratos_session")

			require.NotEmpty(t, gjson.Get(body, "continue_with.#(action==show_settings_ui).flow").String(), "%s", body)
		case RecoveryClientTypeAPI:
			body = submitRecoveryCode(t, c, body, clientType, recoveryCode, http.StatusOK)
			require.NotEmpty(t, gjson.Get(body, "continue_with.#(action==show_settings_ui).flow").String(), "%s", body)
			require.NotEmpty(t, gjson.Get(body, "continue_with.#(action==set_ory_session_token).ory_session_token").String(), "%s", body)
		}
	}

	t.Run("description=should recover an account", func(t *testing.T) {
		checkRecovery := func(t *testing.T, client *http.Client, flowType ClientType, recoveryEmail, recoverySubmissionResponse string) string {
			expectVerfiableAddressStatus(t, recoveryEmail, identity.VerifiableAddressStatusPending)

			assert.EqualValues(t, node.CodeGroup, gjson.Get(recoverySubmissionResponse, "active").String(), "%s", recoverySubmissionResponse)
			assert.True(t, gjson.Get(recoverySubmissionResponse, "ui.nodes.#(attributes.name==code)").Exists(), "%s", recoverySubmissionResponse)
			assert.Len(t, gjson.Get(recoverySubmissionResponse, "ui.messages").Array(), 1, "%s", recoverySubmissionResponse)
			assertx.EqualAsJSON(t, text.NewRecoveryEmailWithCodeSent(), json.RawMessage(gjson.Get(recoverySubmissionResponse, "ui.messages.0").Raw))

			message := testhelpers.CourierExpectMessage(ctx, t, reg, recoveryEmail, "Use code")
			assert.Contains(t, message.Body, "Recover access to your account by entering")

			recoveryCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
			assert.NotEmpty(t, recoveryCode)

			// statusCode := testhelpers.ExpectStatusCode(flowType == RecoveryClientTypeSPA, http.StatusUnprocessableEntity, http.StatusOK)
			return submitRecoveryCode(t, client, recoverySubmissionResponse, flowType, recoveryCode, http.StatusOK)
		}

		t.Run("type=browser", func(t *testing.T) {
			client := testhelpers.NewClientWithCookies(t)
			email := testhelpers.RandomEmail()
			createIdentityToRecover(t, reg, email)
			recoverySubmissionResponse := submitRecoveryForm(t, client, RecoveryClientTypeBrowser, func(v url.Values) {
				v.Set("email", email)
			}, http.StatusOK)
			body := checkRecovery(t, client, RecoveryClientTypeBrowser, email, recoverySubmissionResponse)

			assert.Equal(t, text.NewRecoverySuccessful(time.Now().Add(time.Hour)).Text,
				gjson.Get(body, "ui.messages.0.text").String())

			res, err := client.Get(public.URL + session.RouteWhoami)
			require.NoError(t, err)
			body = string(x.MustReadAll(res.Body))
			require.NoError(t, res.Body.Close())
			assert.Equal(t, "code_recovery", gjson.Get(body, "authentication_methods.0.method").String(), "%s", body)
			assert.Equal(t, "aal1", gjson.Get(body, "authenticator_assurance_level").String(), "%s", body)
		})

		t.Run("type=spa", func(t *testing.T) {
			client := testhelpers.NewClientWithCookies(t)
			email := testhelpers.RandomEmail()
			createIdentityToRecover(t, reg, email)
			recoverySubmissionResponse := submitRecoveryForm(t, client, RecoveryClientTypeSPA, func(v url.Values) {
				v.Set("email", email)
			}, http.StatusOK)
			body := checkRecovery(t, client, RecoveryClientTypeSPA, email, recoverySubmissionResponse)
			assert.Equal(t, "passed_challenge", gjson.Get(body, "state").String())
			assert.Len(t, gjson.Get(body, "continue_with").Array(), 1)
			sfId := gjson.Get(body, "continue_with.#(action==show_settings_ui).flow.id").String()
			assert.NotEmpty(t, uuid.Must(uuid.FromString(sfId)))
		})

		t.Run("type=api", func(t *testing.T) {
			client := &http.Client{}
			email := testhelpers.RandomEmail()
			createIdentityToRecover(t, reg, email)
			recoverySubmissionResponse := submitRecoveryForm(t, client, RecoveryClientTypeAPI, func(v url.Values) {
				v.Set("email", email)
			}, http.StatusOK)
			body := checkRecovery(t, client, RecoveryClientTypeAPI, email, recoverySubmissionResponse)
			assert.Equal(t, "passed_challenge", gjson.Get(body, "state").String())
			assert.Len(t, gjson.Get(body, "continue_with").Array(), 2)
			assert.NotEmpty(t, gjson.Get(body, "continue_with.#(action==set_ory_session_token).ory_session_token").String())
			sfId := gjson.Get(body, "continue_with.#(action==show_settings_ui).flow.id").String()
			assert.NotEmpty(t, uuid.Must(uuid.FromString(sfId)))
		})

		t.Run("description=should return browser to return url", func(t *testing.T) {
			returnTo := public.URL + "/return-to"
			conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{returnTo})
			for _, tc := range []struct {
				desc        string
				returnTo    string
				f           func(t *testing.T, client *http.Client, identity *identity.Identity) *kratos.RecoveryFlow
				expectedAAL string
			}{
				{
					desc:     "should use return_to from recovery flow",
					returnTo: returnTo,
					f: func(t *testing.T, client *http.Client, identity *identity.Identity) *kratos.RecoveryFlow {
						return testhelpers.InitializeRecoveryFlowViaBrowser(t, client, false, public, url.Values{"return_to": []string{returnTo}})
					},
				},
				{
					desc:     "should use return_to from config",
					returnTo: returnTo,
					f: func(t *testing.T, client *http.Client, identity *identity.Identity) *kratos.RecoveryFlow {
						conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryBrowserDefaultReturnTo, returnTo)
						t.Cleanup(func() {
							conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryBrowserDefaultReturnTo, "")
						})
						return testhelpers.InitializeRecoveryFlowViaBrowser(t, client, false, public, nil)
					},
				},
				{
					desc:     "no return to",
					returnTo: "",
					f: func(t *testing.T, client *http.Client, identity *identity.Identity) *kratos.RecoveryFlow {
						return testhelpers.InitializeRecoveryFlowViaBrowser(t, client, false, public, nil)
					},
				},
				{
					desc:     "should use return_to with an account that has 2fa enabled",
					returnTo: returnTo,
					f: func(t *testing.T, client *http.Client, id *identity.Identity) *kratos.RecoveryFlow {
						conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, config.HighestAvailableAAL)
						conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, config.HighestAvailableAAL)
						conf.MustSet(ctx, config.ViperKeyWebAuthnRPDisplayName, "Kratos")
						conf.MustSet(ctx, config.ViperKeyWebAuthnRPID, "ory.sh")

						t.Cleanup(func() {
							conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, identity.AuthenticatorAssuranceLevel1)
							conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, identity.AuthenticatorAssuranceLevel1)
						})
						testhelpers.StrategyEnable(t, conf, identity.CredentialsTypeWebAuthn.String(), true)

						id.SetCredentials(identity.CredentialsTypeWebAuthn, identity.Credentials{
							Type:        identity.CredentialsTypeWebAuthn,
							Config:      []byte(`{"credentials":[{"is_passwordless":false, "display_name":"test"}]}`),
							Identifiers: []string{testhelpers.RandomEmail()},
						})

						require.NoError(t, reg.IdentityManager().Update(ctx, id, identity.ManagerAllowWriteProtectedTraits))
						return testhelpers.InitializeRecoveryFlowViaBrowser(t, client, false, public, url.Values{"return_to": []string{returnTo}})
					},
					expectedAAL: "aal2",
				},
			} {
				t.Run(tc.desc, func(t *testing.T) {
					client := testhelpers.NewClientWithCookies(t)
					email := testhelpers.RandomEmail()
					i := createIdentityToRecover(t, reg, email)

					client.Transport = testhelpers.NewTransportWithLogger(http.DefaultTransport, t).RoundTripper
					f := tc.f(t, client, i)

					formPayload := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
					formPayload.Set("email", email)

					body, res := testhelpers.RecoveryMakeRequest(t, false, f, client, formPayload.Encode())
					assert.EqualValues(t, http.StatusOK, res.StatusCode, "%s", body)
					expectedURL := testhelpers.ExpectURL(false, public.URL+recovery.RouteSubmitFlow, conf.SelfServiceFlowRecoveryUI(ctx).String())
					assert.Contains(t, res.Request.URL.String(), expectedURL, "%+v\n\t%s", res.Request, body)

					body = checkRecovery(t, client, RecoveryClientTypeBrowser, email, body)

					require.Equal(t, text.NewRecoverySuccessful(time.Now().Add(time.Hour)).Text,
						gjson.Get(body, "ui.messages.0.text").String())

					settingsId := gjson.Get(body, "id").String()

					sf, err := reg.SettingsFlowPersister().GetSettingsFlow(ctx, uuid.Must(uuid.FromString(settingsId)))
					require.NoError(t, err)

					u, err := url.Parse(public.URL)
					require.NoError(t, err)
					require.Len(t, client.Jar.Cookies(u), 2)
					found := false
					for _, cookie := range client.Jar.Cookies(u) {
						if cookie.Name == "ory_kratos_session" {
							found = true
						}
					}
					require.True(t, found)

					require.Equal(t, tc.returnTo, sf.ReturnTo)
					res, err = client.Get(public.URL + session.RouteWhoami)
					require.NoError(t, err)
					body = string(x.MustReadAll(res.Body))
					require.NoError(t, res.Body.Close())

					if tc.expectedAAL == "aal2" {
						require.Equal(t, http.StatusForbidden, res.StatusCode)
						require.Equalf(t, session.NewErrAALNotSatisfied("").Reason(), gjson.Get(body, "error.reason").String(), "%s", body)
						require.Equalf(t, "session_aal2_required", gjson.Get(body, "error.id").String(), "%s", body)
					} else {
						assert.Equal(t, "code_recovery", gjson.Get(body, "authentication_methods.0.method").String(), "%s", body)
						assert.Equal(t, "aal1", gjson.Get(body, "authenticator_assurance_level").String(), "%s", body)
					}
				})
			}
		})
	})

	t.Run("description=should set all the correct recovery payloads after submission", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				body := submitRecoveryForm(t, testCase.GetClient(t), testCase.ClientType, func(v url.Values) {
					v.Set("email", "test@ory.sh")
				}, http.StatusOK)
				testhelpers.SnapshotTExcept(t, json.RawMessage(gjson.Get(body, "ui.nodes").String()), []string{"0.attributes.value"})
			})
		}
	})

	t.Run("description=should set all the correct recovery payloads", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				c := testCase.GetClient(t)
				rs := testhelpers.GetRecoveryFlowForType(t, c, public, testCase.FlowType)

				testhelpers.SnapshotTExcept(t, rs.Ui.Nodes, []string{"0.attributes.value"})
				assert.EqualValues(t, public.URL+recovery.RouteSubmitFlow+"?flow="+rs.Id, rs.Ui.Action)
				assert.Empty(t, rs.Ui.Messages)
			})
		}
	})

	t.Run("description=should require an email to be sent", func(t *testing.T) {
		for _, flowType := range flowTypes {
			t.Run("type="+flowType.String(), func(t *testing.T) {
				body := expectValidationError(t, nil, flowType, func(v url.Values) {
					v.Del("email")
				})
				assert.EqualValues(t, node.CodeGroup, gjson.Get(body, "active").String(), "%s", body)
				assert.EqualValues(t, "Property email is missing.",
					gjson.Get(body, "ui.nodes.#(attributes.name==email).messages.0.text").String(),
					"%s", body)
			})
		}
	})

	t.Run("description=should require a valid email to be sent", func(t *testing.T) {
		for _, flowType := range flowTypes {
			t.Run("type="+flowType.String(), func(t *testing.T) {
				for _, email := range []string{"\\", "asdf", "...", "aiacobelli.sec@gmail.com,alejandro.iacobelli@mercadolibre.com"} {
					responseJSON := expectValidationError(t, nil, flowType, func(v url.Values) {
						v.Set("email", email)
					})
					activeMethod := gjson.Get(responseJSON, "active").String()
					assert.EqualValues(t, node.CodeGroup, activeMethod, "expected method to be %s got %s", node.CodeGroup, activeMethod)
					expectedMessage := fmt.Sprintf("%q is not valid \"email\"", email)
					actualMessage := gjson.Get(responseJSON, "ui.nodes.#(attributes.name==email).messages.0.text").String()
					assert.EqualValues(t, expectedMessage, actualMessage, "%s", responseJSON)
				}
			})
		}
	})

	t.Run("description=should try to submit the form while authenticated", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				isSPA := testCase.ClientType == "spa"
				isAPI := testCase.ClientType == "api"
				client := testCase.GetClient(t)

				var f *kratos.RecoveryFlow
				if isAPI {
					f = testhelpers.InitializeRecoveryFlowViaAPI(t, client, public)
				} else {
					f = testhelpers.InitializeRecoveryFlowViaBrowser(t, client, isSPA, public, nil)
				}
				req := httptest.NewRequest("GET", "/sessions/whoami", nil)
				req = req.WithContext(contextx.WithConfigValue(ctx, config.ViperKeySessionLifespan, time.Hour))

				session, err := testhelpers.NewActiveSession(
					req,
					reg,
					&identity.Identity{ID: x.NewUUID(), State: identity.StateActive, NID: x.NewUUID()},
					time.Now(),
					identity.CredentialsTypePassword,
					identity.AuthenticatorAssuranceLevel1,
				)

				require.NoError(t, err)

				// Add the authentication to the request
				client.Transport = testhelpers.NewTransportWithLogger(testhelpers.NewAuthorizedTransport(t, ctx, reg, session), t).RoundTripper

				v := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
				v.Set("email", "some-email@example.org")
				v.Set("method", "code")

				body, res := testhelpers.RecoveryMakeRequest(t, isAPI || isSPA, f, client, testhelpers.EncodeFormAsJSON(t, isAPI || isSPA, v))

				if isAPI || isSPA {
					assert.EqualValues(t, http.StatusBadRequest, res.StatusCode, "%s", body)
					assert.Contains(t, res.Request.URL.String(), recovery.RouteSubmitFlow, "%+v\n\t%s", res.Request, body)
					assertx.EqualAsJSONExcept(t, recovery.ErrAlreadyLoggedIn, json.RawMessage(gjson.Get(body, "error").Raw), nil)
				} else {
					assert.EqualValues(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.Contains(t, res.Request.URL.String(), conf.SelfServiceBrowserDefaultReturnTo(ctx).String(), "%+v\n\t%s", res.Request, body)
				}
			})
		}
	})

	t.Run("description=should not be able to recover account that does not exist", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryNotifyUnknownRecipients, true)

		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryNotifyUnknownRecipients, false)
		})

		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				email := x.NewUUID().String() + "@ory.sh"
				c := testCase.GetClient(t)
				withValues := func(v url.Values) {
					v.Set("email", email)
				}
				body := submitRecoveryForm(t, c, testCase.ClientType, withValues, http.StatusOK)
				assert.EqualValues(t, node.CodeGroup, gjson.Get(body, "active").String(), "%s", body)
				assert.Empty(t, gjson.Get(body, "ui.nodes.#(attributes.name==code).attributes.value").String(), "%s", body)
				assertx.EqualAsJSON(t, text.NewRecoveryEmailWithCodeSent(), json.RawMessage(gjson.Get(body, "ui.messages.0").Raw))

				message := testhelpers.CourierExpectMessage(ctx, t, reg, email, "Account access attempted")
				assert.Contains(t, message.Body, "If this was you, check if you signed up using a different address.")
			})
		}
	})

	t.Run("description=should not be able to recover an inactive account", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				email := "recoverinactive_" + testCase.ClientType.String() + "@ory.sh"
				createIdentityToRecover(t, reg, email)
				values := func(v url.Values) {
					v.Set("email", email)
				}
				cl := testhelpers.NewClientWithCookies(t)

				body := submitRecoveryForm(t, cl, testCase.ClientType, values, http.StatusOK)
				addr, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, email)
				assert.NoError(t, err)

				emailText := testhelpers.CourierExpectMessage(ctx, t, reg, email, "Use code")
				recoveryCode := testhelpers.CourierExpectCodeInMessage(t, emailText, 1)

				// Deactivate the identity
				require.NoError(t, reg.Persister().GetConnection(context.Background()).RawQuery("UPDATE identities SET state=? WHERE id = ?", identity.StateInactive, addr.IdentityID).Exec())

				switch testCase.ClientType {
				case RecoveryClientTypeAPI:
					fallthrough
				case RecoveryClientTypeSPA:
					body = submitRecoveryCode(t, cl, body, testCase.ClientType, recoveryCode, http.StatusUnauthorized)
					assertx.EqualAsJSON(t, session.ErrIdentityDisabled.WithDetail("identity_id", addr.IdentityID), json.RawMessage(gjson.Get(body, "error").Raw), "%s", body)
				default:
					body = submitRecoveryCode(t, cl, body, testCase.ClientType, recoveryCode, http.StatusOK)
					assertx.EqualAsJSON(t, session.ErrIdentityDisabled.WithDetail("identity_id", addr.IdentityID), json.RawMessage(body), "%s", body)
				}
			})
		}
	})

	t.Run("description=should recover and invalidate all other sessions if hook is set", func(t *testing.T) {
		conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal), []config.SelfServiceHook{{Name: "revoke_active_sessions"}})
		t.Cleanup(func() {
			conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()), nil)
		})

		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				email := testhelpers.RandomEmail()
				id := createIdentityToRecover(t, reg, email)

				otherSession, err := testhelpers.NewActiveSession(httptest.NewRequest("GET", "/sessions/whoami", nil), reg, id, time.Now(), identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
				require.NoError(t, err)
				require.NoError(t, reg.SessionPersister().UpsertSession(ctx, otherSession))

				refetchedOtherSession, err := reg.SessionPersister().GetSession(ctx, otherSession.ID, session.ExpandNothing)
				require.NoError(t, err)
				assert.True(t, refetchedOtherSession.IsActive())

				cl := testCase.GetClient(t)
				actual := submitRecoveryForm(t, cl, testCase.ClientType, func(v url.Values) {
					v.Set("email", email)
				}, http.StatusOK)
				message := testhelpers.CourierExpectMessage(ctx, t, reg, email, "Use code")
				recoveryCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)

				submitCodeAndExpectRedirectToSettings(t, cl, testCase.ClientType, recoveryCode, actual)

				refetchedOtherSession, err = reg.SessionPersister().GetSession(ctx, otherSession.ID, session.ExpandNothing)
				require.NoError(t, err)
				assert.False(t, refetchedOtherSession.IsActive())
			})
		}
	})

	t.Run("description=should not be able to use an invalid code more than 5 times", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				email := testhelpers.RandomEmail()
				createIdentityToRecover(t, reg, email)
				c := testCase.GetClient(t)
				body := submitRecoveryForm(t, c, testCase.ClientType, func(v url.Values) {
					v.Set("email", email)
				}, http.StatusOK)

				initialFlowId := gjson.Get(body, "id")

				for submitTry := 0; submitTry < 5; submitTry++ {
					body := submitRecoveryCode(t, c, body, testCase.ClientType, "12312312", http.StatusOK)

					testhelpers.AssertMessage(t, []byte(body), "The recovery code is invalid or has already been used. Please try again.")
				}

				switch testCase.ClientType {
				case RecoveryClientTypeBrowser:
					// submit an invalid code for the 6th time
					body = submitRecoveryCode(t, c, body, testCase.ClientType, "12312312", http.StatusOK)

					require.Len(t, gjson.Get(body, "ui.messages").Array(), 1, "%s", body)
					assert.Equal(t, "The request was submitted too often. Please request another code.", gjson.Get(body, "ui.messages.0.text").String())

					// check that a new flow has been created
					assert.NotEqual(t, gjson.Get(body, "id"), initialFlowId)

					assert.True(t, gjson.Get(body, "ui.nodes.#(attributes.name==email)").Exists())
				case RecoveryClientTypeSPA:
					fallthrough
				case RecoveryClientTypeAPI:
					// submit an invalid code for the 6th time
					body = submitRecoveryCode(t, c, body, testCase.ClientType, "12312312", http.StatusBadRequest)

					assert.Equal(t, "Bad Request", gjson.Get(body, "error.status").String(), "%s", body)
					assert.Equal(t, "The request was submitted too often. Please request another code.", gjson.Get(body, "error.reason").String(), "%s", body)
					continueWith := gjson.Get(body, "error.details.continue_with").Array()
					assert.Len(t, continueWith, 1, "%s", body)
					assert.Equal(t, "show_recovery_ui", continueWith[0].Get("action").String(), "%s", body)
					flowId := continueWith[0].Get("flow.id").String()
					assert.NotEmpty(t, flowId, "%s", body)
					require.NotEqual(t, flowId, initialFlowId, "%s", body)

					flow, err := reg.Persister().GetRecoveryFlow(ctx, uuid.Must(uuid.FromString(flowId)))
					require.NoError(t, err)
					assert.Len(t, flow.UI.Messages, 1, "%+v", flow)
					assert.Equal(t, "The request was submitted too often. Please request another code.", flow.UI.Messages[0].Text)
				}
			})
		}
	})

	t.Run("description=should be able to recover after using invalid code", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				c := testCase.GetClient(t)
				recoveryEmail := testhelpers.RandomEmail()
				_ = createIdentityToRecover(t, reg, recoveryEmail)

				actual := submitRecoveryForm(t, c, testCase.ClientType, func(v url.Values) {
					v.Set("email", recoveryEmail)
				}, http.StatusOK)

				message := testhelpers.CourierExpectMessage(ctx, t, reg, recoveryEmail, "Use code")
				recoveryCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)

				form := withCSRFToken(t, testCase.ClientType, actual, url.Values{
					"code": {"12312312"},
				})

				action := gjson.Get(actual, "ui.action").String()
				require.NotEmpty(t, action)

				res, err := c.Post(action, testCase.FormContentType, bytes.NewBufferString(form))
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, res.StatusCode)

				flowId := gjson.Get(actual, "id").String()
				require.NotEmpty(t, flowId)

				rs, res, err := testhelpers.
					NewSDKCustomClient(public, c).
					FrontendAPI.GetRecoveryFlow(context.Background()).
					Id(flowId).
					Execute()

				require.NoError(t, err)
				body := ioutilx.MustReadAll(res.Body)
				require.NotEmpty(t, body)

				require.Len(t, rs.Ui.Messages, 1)
				assert.Equal(t, "The recovery code is invalid or has already been used. Please try again.", rs.Ui.Messages[0].Text)

				form = withCSRFToken(t, testCase.ClientType, actual, url.Values{
					"code": {recoveryCode},
				})
				// Now submit the correct code
				res, err = c.Post(action, testCase.FormContentType, bytes.NewBufferString(form))
				require.NoError(t, err)
				switch testCase.ClientType {
				case RecoveryClientTypeBrowser:
					assert.Equal(t, http.StatusOK, res.StatusCode)

					json := ioutilx.MustReadAll(res.Body)

					assert.Len(t, gjson.GetBytes(json, "ui.messages").Array(), 1)
					assert.Contains(t, gjson.GetBytes(json, "ui.messages.0.text").String(), "You successfully recovered your account.")
				case RecoveryClientTypeSPA:
					assert.Equal(t, http.StatusOK, res.StatusCode)

					json := ioutilx.MustReadAll(res.Body)

					require.Len(t, c.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 2)
					cookies := spew.Sdump(c.Jar.Cookies(urlx.ParseOrPanic(public.URL)))
					assert.Contains(t, cookies, "ory_kratos_session")

					require.NotEmpty(t, gjson.GetBytes(json, "continue_with.#(action==show_settings_ui).flow").String(), "%s", json)
				case RecoveryClientTypeAPI:
					assert.Equal(t, http.StatusOK, res.StatusCode)

					json := ioutilx.MustReadAll(res.Body)

					require.NotEmpty(t, gjson.GetBytes(json, "continue_with.#(action==show_settings_ui).flow").String(), "%s", json)
					require.NotEmpty(t, gjson.GetBytes(json, "continue_with.#(action==set_ory_session_token).ory_session_token").String(), "%s", json)
				}
			})
		}
	})

	t.Run("description=does not issue csrf cookie when submitting API flow", func(t *testing.T) {
		t.Run("type="+RecoveryClientTypeAPI.String(), func(t *testing.T) {
			c := new(http.Client)
			recoveryEmail := testhelpers.RandomEmail()
			_ = createIdentityToRecover(t, reg, recoveryEmail)

			actual := submitRecoveryForm(t, c, RecoveryClientTypeAPI, func(v url.Values) {
				v.Set("email", recoveryEmail)
			}, http.StatusOK)

			message := testhelpers.CourierExpectMessage(ctx, t, reg, recoveryEmail, "Use code")
			recoveryCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)

			action := gjson.Get(actual, "ui.action").String()
			require.NotEmpty(t, action)

			flowId := gjson.Get(actual, "id").String()
			require.NotEmpty(t, flowId)

			form := withCSRFToken(t, RecoveryClientTypeAPI, actual, url.Values{
				"code": {recoveryCode},
			})

			// Now submit the correct code
			res, err := c.Post(action, "application/json", bytes.NewBufferString(form))
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, res.StatusCode)

			assert.Empty(t, res.Header.Get("Set-Cookie"))

			json := ioutilx.MustReadAll(res.Body)
			require.NotEmpty(t, gjson.GetBytes(json, "continue_with.#(action==show_settings_ui).flow").String(), "%s", json)
			require.NotEmpty(t, gjson.GetBytes(json, "continue_with.#(action==set_ory_session_token).ory_session_token").String(), "%s", json)
		})
	})

	t.Run("description=should not be able to use an invalid code", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				email := testhelpers.RandomEmail()
				createIdentityToRecover(t, reg, email)
				c := testCase.GetClient(t)

				body := submitRecoveryForm(t, c, testCase.ClientType, func(v url.Values) {
					v.Set("email", email)
				}, http.StatusOK)

				body = submitRecoveryCode(t, c, body, RecoveryClientTypeBrowser, "12312312", http.StatusOK)

				testhelpers.AssertMessage(t, []byte(body), "The recovery code is invalid or has already been used. Please try again.")
			})
		}
	})

	t.Run("description=should not be able to submit recover address after flow expired", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				recoveryEmail := testhelpers.RandomEmail()
				createIdentityToRecover(t, reg, recoveryEmail)
				conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryRequestLifespan, time.Millisecond*100)
				t.Cleanup(func() {
					conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryRequestLifespan, time.Minute)
				})

				c := testCase.GetClient(t)
				var rs *kratos.RecoveryFlow
				var res *http.Response
				var err error
				switch testCase.ClientType {
				case RecoveryClientTypeBrowser:
					fallthrough
				case RecoveryClientTypeSPA:
					rs = testhelpers.GetRecoveryFlow(t, c, public)
					time.Sleep(time.Millisecond * 110)
					res, err = c.PostForm(rs.Ui.Action, url.Values{"email": {recoveryEmail}, "method": {"code"}})
					require.NoError(t, err)
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.NotContains(t, res.Request.URL.String(), "flow="+rs.Id)
					assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowRecoveryUI(ctx).String())
				case RecoveryClientTypeAPI:
					rs = testhelpers.InitializeRecoveryFlowViaAPI(t, c, public)
					time.Sleep(time.Millisecond * 110)
					form := testhelpers.EncodeFormAsJSON(t, true, url.Values{"email": {recoveryEmail}, "method": {"code"}})
					res, err = c.Post(rs.Ui.Action, "application/json", bytes.NewBufferString(form))
					require.NoError(t, err)
					body := ioutilx.MustReadAll(res.Body)
					assert.Equal(t, http.StatusGone, res.StatusCode, "%s", body)
					assert.Equal(t, "self_service_flow_expired", gjson.GetBytes(body, "error.id").String(), "%s", body)
					continueWith := gjson.GetBytes(body, "error.details.continue_with").Array()
					assert.Len(t, continueWith, 1, "%s", body)
					assert.Equal(t, "show_recovery_ui", continueWith[0].Get("action").String(), "%s", continueWith)
					flowId := continueWith[0].Get("flow.id").String()
					require.NotEmpty(t, flowId, "%s", body)
				}

				addr, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, recoveryEmail)
				assert.NoError(t, err)
				assert.False(t, addr.Verified)
				assert.Nil(t, addr.VerifiedAt)
				assert.Equal(t, identity.VerifiableAddressStatusPending, addr.Status)
			})
		}
	})

	t.Run("description=should not be able to submit code after flow expired", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryRequestLifespan, time.Millisecond*200)
		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryRequestLifespan, time.Minute)
		})
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				recoveryEmail := testhelpers.RandomEmail()
				createIdentityToRecover(t, reg, recoveryEmail)

				c := testhelpers.NewClientWithCookies(t)

				body := submitRecoveryForm(t, c, testCase.ClientType, func(v url.Values) {
					v.Set("email", recoveryEmail)
				}, http.StatusOK)

				initialFlowId := gjson.Get(body, "id")

				message := testhelpers.CourierExpectMessage(ctx, t, reg, recoveryEmail, "Use code")
				assert.Contains(t, message.Body, "Recover access to your account by entering")

				recoveryCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)

				time.Sleep(time.Millisecond * 201)

				if testCase.FlowType == "browser" {
					body = submitRecoveryCode(t, c, body, testCase.ClientType, recoveryCode, http.StatusOK)
					assert.NotEqual(t, gjson.Get(body, "id"), initialFlowId)

					testhelpers.AssertMessage(t, []byte(body), "The recovery flow expired 0.00 minutes ago, please try again.")
				} else {
					body = submitRecoveryCode(t, c, body, testCase.ClientType, recoveryCode, http.StatusGone)
					assert.NotEqual(t, gjson.Get(body, "id"), initialFlowId)
					assert.Equal(t, "self_service_flow_expired", gjson.Get(body, "error.id").String())
				}

				addr, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, recoveryEmail)
				require.NoError(t, err)
				assert.False(t, addr.Verified)
				assert.Nil(t, addr.VerifiedAt)
				assert.Equal(t, identity.VerifiableAddressStatusPending, addr.Status)
			})
		}
	})

	t.Run("description=should not break ui if empty code is submitted", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				recoveryEmail := testhelpers.RandomEmail()
				createIdentityToRecover(t, reg, recoveryEmail)

				c := testCase.GetClient(t)
				body := submitRecoveryForm(t, c, testCase.ClientType, func(v url.Values) {
					v.Set("email", recoveryEmail)
				}, http.StatusOK)

				action := gjson.Get(body, "ui.action").String()
				require.NotEmpty(t, action)

				body = submitRecoveryCode(t, c, body, testCase.ClientType, "", http.StatusOK)

				assert.NotContains(t, gjson.Get(body, "ui.nodes").String(), "Property email is missing.")
				testhelpers.AssertMessage(t, []byte(body), "The recovery code is invalid or has already been used. Please try again.")
			})
		}
	})

	t.Run("description=should be able to resend the recovery code", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				recoveryEmail := testhelpers.RandomEmail()
				createIdentityToRecover(t, reg, recoveryEmail)

				c := testCase.GetClient(t)
				body := submitRecoveryForm(t, c, testCase.ClientType, func(v url.Values) {
					v.Set("email", recoveryEmail)
				}, http.StatusOK)

				action := gjson.Get(body, "ui.action").String()
				require.NotEmpty(t, action)
				assert.Equal(t, recoveryEmail, gjson.Get(body, "ui.nodes.#(attributes.name==email).attributes.value").String())

				body = resendRecoveryCode(t, c, body, testCase.ClientType, http.StatusOK)
				assert.True(t, gjson.Get(body, "ui.nodes.#(attributes.name==code)").Exists())
				assert.Equal(t, recoveryEmail, gjson.Get(body, "ui.nodes.#(attributes.name==email).attributes.value").String())

				message := testhelpers.CourierExpectMessage(ctx, t, reg, recoveryEmail, "Use code")
				recoveryCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)

				submitCodeAndExpectRedirectToSettings(t, c, testCase.ClientType, recoveryCode, body)
			})
		}
	})

	t.Run("description=should not be able to use first code after re-sending email", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				recoveryEmail := testhelpers.RandomEmail()
				createIdentityToRecover(t, reg, recoveryEmail)

				c := testCase.GetClient(t)
				body := submitRecoveryForm(t, c, testCase.ClientType, func(v url.Values) {
					v.Set("email", recoveryEmail)
				}, http.StatusOK)

				action := gjson.Get(body, "ui.action").String()
				require.NotEmpty(t, action)
				assert.Equal(t, recoveryEmail, gjson.Get(body, "ui.nodes.#(attributes.name==email).attributes.value").String())

				message1 := testhelpers.CourierExpectMessage(ctx, t, reg, recoveryEmail, "Use code")
				recoveryCode1 := testhelpers.CourierExpectCodeInMessage(t, message1, 1)

				body = resendRecoveryCode(t, c, body, testCase.ClientType, http.StatusOK)
				assert.True(t, gjson.Get(body, "ui.nodes.#(attributes.name==code)").Exists())
				assert.Equal(t, recoveryEmail, gjson.Get(body, "ui.nodes.#(attributes.name==email).attributes.value").String())

				message2 := testhelpers.CourierExpectMessage(ctx, t, reg, recoveryEmail, "Use code")
				recoveryCode2 := testhelpers.CourierExpectCodeInMessage(t, message2, 1)

				body = submitRecoveryCode(t, c, body, testCase.ClientType, recoveryCode1, http.StatusOK)
				testhelpers.AssertMessage(t, []byte(body), "The recovery code is invalid or has already been used. Please try again.")

				submitCodeAndExpectRedirectToSettings(t, c, testCase.ClientType, recoveryCode2, body)
			})
		}
	})

	t.Run("description=should not show outdated validation message if newer message appears #2799", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				recoveryEmail := testhelpers.RandomEmail()
				createIdentityToRecover(t, reg, recoveryEmail)

				c := testCase.GetClient(t)
				body := submitRecoveryForm(t, c, testCase.ClientType, func(v url.Values) {
					v.Set("email", recoveryEmail)
				}, http.StatusOK)

				action := gjson.Get(body, "ui.action").String()
				require.NotEmpty(t, action)
				assert.Equal(t, recoveryEmail, gjson.Get(body, "ui.nodes.#(attributes.name==email).attributes.value").String())

				body = submitRecoveryCode(t, c, body, testCase.ClientType, "12312312", http.StatusOK) // Now send a wrong code that triggers "global" validation error

				assert.Empty(t, gjson.Get(body, "ui.nodes.#(attributes.name==code).messages").Array())
				testhelpers.AssertMessage(t, []byte(body), "The recovery code is invalid or has already been used. Please try again.")
			})
		}
	})

	t.Run("description=should recover if post recovery hook is successful", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal), []config.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})
				t.Cleanup(func() {
					conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal), nil)
				})

				recoveryEmail := testhelpers.RandomEmail()
				createIdentityToRecover(t, reg, recoveryEmail)

				cl := testCase.GetClient(t)
				body := submitRecoveryForm(t, cl, testCase.ClientType, func(v url.Values) {
					v.Set("email", recoveryEmail)
				}, http.StatusOK)

				message := testhelpers.CourierExpectMessage(ctx, t, reg, recoveryEmail, "Use code")
				recoveryCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)

				action := gjson.Get(body, "ui.action").String()
				require.NotEmpty(t, action)
				assert.Equal(t, recoveryEmail, gjson.Get(body, "ui.nodes.#(attributes.name==email).attributes.value").String())

				submitCodeAndExpectRedirectToSettings(t, cl, testCase.ClientType, recoveryCode, body)
			})
		}
	})

	t.Run("description=should not be able to recover if post recovery hook fails", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal), []config.SelfServiceHook{{Name: "err", Config: []byte(`{"ExecutePostRecoveryHook": "err"}`)}})
				t.Cleanup(func() {
					conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal), nil)
				})

				recoveryEmail := testhelpers.RandomEmail()
				createIdentityToRecover(t, reg, recoveryEmail)

				cl := testhelpers.NewClientWithCookies(t)
				body := submitRecoveryForm(t, cl, testCase.ClientType, func(v url.Values) {
					v.Set("email", recoveryEmail)
				}, http.StatusOK)

				message := testhelpers.CourierExpectMessage(ctx, t, reg, recoveryEmail, "Use code")
				recoveryCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)

				action := gjson.Get(body, "ui.action").String()
				require.NotEmpty(t, action)
				assert.Equal(t, recoveryEmail, gjson.Get(body, "ui.nodes.#(attributes.name==email).attributes.value").String())

				cl.CheckRedirect = func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				}

				initialFlowId := gjson.Get(body, "id")
				switch testCase.ClientType {
				case RecoveryClientTypeBrowser:
					body = submitRecoveryCode(t, cl, body, testCase.ClientType, recoveryCode, http.StatusSeeOther)
					assert.NotEqual(t, gjson.Get(body, "id"), initialFlowId)

					require.Len(t, cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 1)
					cookies := spew.Sdump(cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)))
					assert.NotContains(t, cookies, "ory_kratos_session")
				case RecoveryClientTypeSPA:
					body = submitRecoveryCode(t, cl, body, testCase.ClientType, recoveryCode, http.StatusBadRequest)
					assert.NotEqual(t, gjson.Get(body, "id"), initialFlowId)

					require.Len(t, cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 1)
					cookies := spew.Sdump(cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)))
					assert.NotContains(t, cookies, "ory_kratos_session")
				case RecoveryClientTypeAPI:
					body = submitRecoveryCode(t, cl, body, testCase.ClientType, recoveryCode, http.StatusBadRequest)
					assert.NotEqual(t, gjson.Get(body, "id"), initialFlowId)
					require.Equal(t, "err", gjson.Get(body, "error.message").String(), "%s", body)
				}
			})
		}
	})
}

// Recovery V2 is only tested with `ContinueWith`.
func TestRecovery_V2_WithContinueWith_OneAddress_Email(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.StrategyEnable(t, conf, string(recovery.RecoveryStrategyCode), true)
	testhelpers.StrategyEnable(t, conf, string(recovery.RecoveryStrategyLink), false)
	conf.MustSet(ctx, config.ViperKeyUseContinueWithTransitions, true)
	conf.MustSet(ctx, config.ViperKeyChooseRecoveryAddress, true)

	initViper(t, ctx, conf)

	_ = testhelpers.NewRecoveryUIFlowEchoServer(t, reg)
	_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
	_ = testhelpers.NewSettingsUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)

	public, _, _, _ := testhelpers.NewKratosServerWithCSRFAndRouters(t, reg)

	submitRecoveryFormInitial := func(t *testing.T, client *http.Client, flowType ClientType, values func(url.Values), code int) string {
		isSPA := flowType == RecoveryClientTypeSPA
		isAPI := flowType == RecoveryClientTypeAPI
		if client == nil {
			client = testhelpers.NewDebugClient(t)
			if !isAPI {
				client = testhelpers.NewClientWithCookies(t)
				client.Transport = testhelpers.NewTransportWithLogger(http.DefaultTransport, t).RoundTripper
			}
		}

		expectedUrl := testhelpers.ExpectURL(isAPI || isSPA, public.URL+recovery.RouteSubmitFlow, conf.SelfServiceFlowRecoveryUI(ctx).String())
		return testhelpers.SubmitRecoveryForm(t, isAPI, isSPA, client, public, values, code, expectedUrl)
	}

	submitRecoveryFormSubsequent := func(t *testing.T, client *http.Client, flow string, flowType ClientType, urlValuesFn func(url.Values), statusCode int) string {
		t.Helper()
		action := gjson.Get(flow, "ui.action").String()
		assert.NotEmpty(t, action)

		urlValues := url.Values{}
		urlValuesFn(urlValues)
		values := withCSRFToken(t, flowType, flow, urlValues)

		contentType := "application/json"
		if flowType == RecoveryClientTypeBrowser {
			contentType = "application/x-www-form-urlencoded"
		}

		res, err := client.Post(action, contentType, bytes.NewBufferString(values))
		require.NoError(t, err)
		assert.Equal(t, statusCode, res.StatusCode)

		return string(ioutilx.MustReadAll(res.Body))
	}

	expectVerifiableAddressStatus := func(t *testing.T, email string, status identity.VerifiableAddressStatus) {
		addr, err := reg.IdentityPool().
			FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, email)
		assert.NoError(t, err)
		assert.Equal(t, status, addr.Status, "verifiable address %s was not %s. instead %s", email, status, addr.Status)
	}

	checkRecoveryScreenAskForCode := func(t *testing.T, chosenRecoveryConfirmAddress, recoverySubmissionResponse string) {
		expectVerifiableAddressStatus(t, chosenRecoveryConfirmAddress, identity.VerifiableAddressStatusPending)

		assert.True(t, gjson.Get(recoverySubmissionResponse, "ui.nodes.#(attributes.name==code)").Exists(), "%s", recoverySubmissionResponse)
		assert.Len(t, gjson.Get(recoverySubmissionResponse, "ui.messages").Array(), 1, "%s", recoverySubmissionResponse)
		assertx.EqualAsJSON(t, text.NewRecoveryCodeRecoverySelectAddressSent(code.MaskAddress(chosenRecoveryConfirmAddress)), json.RawMessage(gjson.Get(recoverySubmissionResponse, "ui.messages.0").Raw))
	}

	extractCodeFromCourierAndSubmit := func(t *testing.T, client *http.Client, flowType ClientType, chosenRecoveryConfirmAddress string, recoverySubmissionResponse string, expectedCode int) string {
		message := testhelpers.CourierExpectMessage(ctx, t, reg, chosenRecoveryConfirmAddress, "Use code")
		assert.Contains(t, message.Body, "Recover access to your account by entering")

		recoveryCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
		assert.NotEmpty(t, recoveryCode)

		return submitRecoveryFormSubsequent(t, client, recoverySubmissionResponse, flowType, func(v url.Values) { v.Set("code", recoveryCode) }, expectedCode)
	}

	recoverHappyPath := func(t *testing.T, client *http.Client, clientType ClientType, anyAddress string) string {
		recoverySubmissionResponse := submitRecoveryFormInitial(t, client, clientType, func(v url.Values) {
			v.Set("recovery_address", anyAddress)
		}, http.StatusOK)

		checkRecoveryScreenAskForCode(t, anyAddress, recoverySubmissionResponse)

		body := extractCodeFromCourierAndSubmit(t, client, clientType, anyAddress, recoverySubmissionResponse, http.StatusOK)
		return body
	}

	expectRedirectToSettings := func(t *testing.T, client *http.Client, clientType ClientType, body string) {
		switch clientType {
		case RecoveryClientTypeBrowser:
			require.Len(t, client.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 2)
			cookies := spew.Sdump(client.Jar.Cookies(urlx.ParseOrPanic(public.URL)))
			assert.Contains(t, cookies, "ory_kratos_session")
			require.Contains(t, body, "You successfully recovered your account. Please change your password or set up an alternative login method (e.g. social sign in) within the next 60.00 minutes.")
		case RecoveryClientTypeSPA:
			require.Len(t, client.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 2)
			cookies := spew.Sdump(client.Jar.Cookies(urlx.ParseOrPanic(public.URL)))
			assert.Contains(t, cookies, "ory_kratos_session")

			require.NotEmpty(t, gjson.Get(body, "continue_with.#(action==show_settings_ui).flow").String(), "%s", body)
		case RecoveryClientTypeAPI:
			require.NotEmpty(t, gjson.Get(body, "continue_with.#(action==show_settings_ui).flow").String(), "%s", body)
			require.NotEmpty(t, gjson.Get(body, "continue_with.#(action==set_ory_session_token).ory_session_token").String(), "%s", body)
		}
	}

	t.Run("description=should recover an account", func(t *testing.T) {
		t.Run("type=browser", func(t *testing.T) {
			client := testhelpers.NewClientWithCookies(t)
			email := testhelpers.RandomEmail()
			createIdentityToRecover(t, reg, email)

			body := recoverHappyPath(t, client, RecoveryClientTypeBrowser, email)

			assert.Equal(t, text.NewRecoverySuccessful(time.Now().Add(time.Hour)).Text,
				gjson.Get(body, "ui.messages.0.text").String())

			res, err := client.Get(public.URL + session.RouteWhoami)
			require.NoError(t, err)
			body = string(x.MustReadAll(res.Body))
			require.NoError(t, res.Body.Close())
			assert.Equal(t, "code_recovery", gjson.Get(body, "authentication_methods.0.method").String(), "%s", body)
			assert.Equal(t, "aal1", gjson.Get(body, "authenticator_assurance_level").String(), "%s", body)
		})

		t.Run("type=spa", func(t *testing.T) {
			client := testhelpers.NewClientWithCookies(t)
			email := testhelpers.RandomEmail()
			createIdentityToRecover(t, reg, email)

			body := recoverHappyPath(t, client, RecoveryClientTypeSPA, email)

			assert.Equal(t, "passed_challenge", gjson.Get(body, "state").String())
			assert.Len(t, gjson.Get(body, "continue_with").Array(), 1)
			sfId := gjson.Get(body, "continue_with.#(action==show_settings_ui).flow.id").String()
			assert.NotEmpty(t, uuid.Must(uuid.FromString(sfId)))
		})

		t.Run("type=api", func(t *testing.T) {
			client := &http.Client{}
			email := testhelpers.RandomEmail()
			createIdentityToRecover(t, reg, email)

			body := recoverHappyPath(t, client, RecoveryClientTypeAPI, email)

			assert.Equal(t, "passed_challenge", gjson.Get(body, "state").String())
			assert.Len(t, gjson.Get(body, "continue_with").Array(), 2)
			assert.NotEmpty(t, gjson.Get(body, "continue_with.#(action==set_ory_session_token).ory_session_token").String())
			sfId := gjson.Get(body, "continue_with.#(action==show_settings_ui).flow.id").String()
			assert.NotEmpty(t, uuid.Must(uuid.FromString(sfId)))
		})

		t.Run("description=should return browser to return url", func(t *testing.T) {
			returnTo := public.URL + "/return-to"
			conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{returnTo})
			for _, tc := range []struct {
				desc        string
				returnTo    string
				f           func(t *testing.T, client *http.Client, identity *identity.Identity) *kratos.RecoveryFlow
				expectedAAL string
			}{
				{
					desc:     "should use return_to from recovery flow",
					returnTo: returnTo,
					f: func(t *testing.T, client *http.Client, identity *identity.Identity) *kratos.RecoveryFlow {
						return testhelpers.InitializeRecoveryFlowViaBrowser(t, client, false, public, url.Values{"return_to": []string{returnTo}})
					},
				},
				{
					desc:     "should use return_to from config",
					returnTo: returnTo,
					f: func(t *testing.T, client *http.Client, identity *identity.Identity) *kratos.RecoveryFlow {
						conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryBrowserDefaultReturnTo, returnTo)
						t.Cleanup(func() {
							conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryBrowserDefaultReturnTo, "")
						})
						return testhelpers.InitializeRecoveryFlowViaBrowser(t, client, false, public, nil)
					},
				},
				{
					desc:     "no return to",
					returnTo: "",
					f: func(t *testing.T, client *http.Client, identity *identity.Identity) *kratos.RecoveryFlow {
						return testhelpers.InitializeRecoveryFlowViaBrowser(t, client, false, public, nil)
					},
				},
				{
					desc:     "should use return_to with an account that has 2fa enabled",
					returnTo: returnTo,
					f: func(t *testing.T, client *http.Client, id *identity.Identity) *kratos.RecoveryFlow {
						conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, config.HighestAvailableAAL)
						conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, config.HighestAvailableAAL)
						conf.MustSet(ctx, config.ViperKeyWebAuthnRPDisplayName, "Kratos")
						conf.MustSet(ctx, config.ViperKeyWebAuthnRPID, "ory.sh")

						t.Cleanup(func() {
							conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, identity.AuthenticatorAssuranceLevel1)
							conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, identity.AuthenticatorAssuranceLevel1)
						})
						testhelpers.StrategyEnable(t, conf, identity.CredentialsTypeWebAuthn.String(), true)

						id.SetCredentials(identity.CredentialsTypeWebAuthn, identity.Credentials{
							Type:        identity.CredentialsTypeWebAuthn,
							Config:      []byte(`{"credentials":[{"is_passwordless":false, "display_name":"test"}]}`),
							Identifiers: []string{testhelpers.RandomEmail()},
						})

						require.NoError(t, reg.IdentityManager().Update(ctx, id, identity.ManagerAllowWriteProtectedTraits))
						return testhelpers.InitializeRecoveryFlowViaBrowser(t, client, false, public, url.Values{"return_to": []string{returnTo}})
					},
					expectedAAL: "aal2",
				},
			} {
				t.Run(tc.desc, func(t *testing.T) {
					client := testhelpers.NewClientWithCookies(t)
					email := testhelpers.RandomEmail()
					i := createIdentityToRecover(t, reg, email)

					client.Transport = testhelpers.NewTransportWithLogger(http.DefaultTransport, t).RoundTripper
					f := tc.f(t, client, i)

					formPayload := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
					formPayload.Set("recovery_address", email)

					body, res := testhelpers.RecoveryMakeRequest(t, false, f, client, formPayload.Encode())
					assert.EqualValues(t, http.StatusOK, res.StatusCode, "%s", body)

					body = extractCodeFromCourierAndSubmit(t, client, RecoveryClientTypeBrowser, email, body, http.StatusOK)

					require.Equal(t, text.NewRecoverySuccessful(time.Now().Add(time.Hour)).Text,
						gjson.Get(body, "ui.messages.0.text").String())

					settingsId := gjson.Get(body, "id").String()

					sf, err := reg.SettingsFlowPersister().GetSettingsFlow(ctx, uuid.Must(uuid.FromString(settingsId)))
					require.NoError(t, err)

					u, err := url.Parse(public.URL)
					require.NoError(t, err)
					require.Len(t, client.Jar.Cookies(u), 2)
					found := false
					for _, cookie := range client.Jar.Cookies(u) {
						if cookie.Name == "ory_kratos_session" {
							found = true
						}
					}
					require.True(t, found)

					require.Equal(t, tc.returnTo, sf.ReturnTo)
					res, err = client.Get(public.URL + session.RouteWhoami)
					require.NoError(t, err)
					body = string(x.MustReadAll(res.Body))
					require.NoError(t, res.Body.Close())

					if tc.expectedAAL == "aal2" {
						require.Equal(t, http.StatusForbidden, res.StatusCode)
						require.Equalf(t, session.NewErrAALNotSatisfied("").Reason(), gjson.Get(body, "error.reason").String(), "%s", body)
						require.Equalf(t, "session_aal2_required", gjson.Get(body, "error.id").String(), "%s", body)
					} else {
						assert.Equal(t, "code_recovery", gjson.Get(body, "authentication_methods.0.method").String(), "%s", body)
						assert.Equal(t, "aal1", gjson.Get(body, "authenticator_assurance_level").String(), "%s", body)
					}
				})
			}
		})
	})

	t.Run("description=should set all the correct recovery payloads after submission", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				address := fmt.Sprintf("test-%s@ory.sh", testCase.ClientType)
				createIdentityToRecover(t, reg, address)
				body := submitRecoveryFormInitial(t, testCase.GetClient(t), testCase.ClientType, func(u url.Values) { u.Set("recovery_address", address) }, http.StatusOK)
				testhelpers.SnapshotTExcept(t, json.RawMessage(gjson.Get(body, "ui.nodes").String()), []string{"0.attributes.value"})
			})
		}
	})

	t.Run("description=should set all the correct recovery payloads", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				c := testCase.GetClient(t)
				rs := testhelpers.GetRecoveryFlowForType(t, c, public, testCase.FlowType)

				testhelpers.SnapshotTExcept(t, rs.Ui.Nodes, []string{"0.attributes.value"})
				assert.EqualValues(t, public.URL+recovery.RouteSubmitFlow+"?flow="+rs.Id, rs.Ui.Action)
				assert.Empty(t, rs.Ui.Messages)
			})
		}
	})

	t.Run("description=should require an address to be sent", func(t *testing.T) {
		for _, flowType := range flowTypes {
			t.Run("type="+flowType.String(), func(t *testing.T) {
				code := testhelpers.ExpectStatusCode(flowType == RecoveryClientTypeAPI || flowType == RecoveryClientTypeSPA, http.StatusBadRequest, http.StatusOK)
				body := submitRecoveryFormInitial(t, nil, flowType, func(v url.Values) {
					v.Del("recovery_address")
				}, code)
				assert.EqualValues(t, node.CodeGroup, gjson.Get(body, "active").String(), "%s", body)
				assert.EqualValues(t, "Property recovery_address is missing.",
					gjson.Get(body, "ui.nodes.#(attributes.name==recovery_address).messages.0.text").String(),
					"%s", body)
			})
		}
	})

	t.Run("description=should pretend the address exists when it does not", func(t *testing.T) {
		for _, flowType := range flowTypes {
			t.Run("type="+flowType.String(), func(t *testing.T) {
				for _, address := range []string{"\\@", "asdf@", "...@", "aiacobelli.sec@gmail.com,alejandro.iacobelli@mercadolibre.com"} {
					body := submitRecoveryFormInitial(t, nil, flowType, func(v url.Values) {
						v.Set("recovery_address", address)
					}, http.StatusOK)

					activeMethod := gjson.Get(body, "active").String()
					assert.EqualValues(t, node.CodeGroup, activeMethod, "expected method to be %s got %s", node.CodeGroup, activeMethod)
					expectedMessage := text.NewRecoveryCodeRecoverySelectAddressSent(code.MaskAddress(address)).Text
					actualMessage := gjson.Get(body, "ui.messages.0.text").String()
					assert.EqualValues(t, expectedMessage, actualMessage, "%s", body)
				}
			})
		}
	})

	t.Run("description=should try to submit the form while authenticated", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				isSPA := testCase.ClientType == "spa"
				isAPI := testCase.ClientType == "api"
				client := testCase.GetClient(t)

				var f *kratos.RecoveryFlow
				if isAPI {
					f = testhelpers.InitializeRecoveryFlowViaAPI(t, client, public)
				} else {
					f = testhelpers.InitializeRecoveryFlowViaBrowser(t, client, isSPA, public, nil)
				}
				req := httptest.NewRequest("GET", "/sessions/whoami", nil).WithContext(contextx.WithConfigValue(ctx, config.ViperKeySessionLifespan, time.Hour))

				session, err := testhelpers.NewActiveSession(
					req,
					reg,
					&identity.Identity{ID: x.NewUUID(), State: identity.StateActive, NID: x.NewUUID()},
					time.Now(),
					identity.CredentialsTypePassword,
					identity.AuthenticatorAssuranceLevel1,
				)

				require.NoError(t, err)

				// Add the authentication to the request
				client.Transport = testhelpers.NewTransportWithLogger(testhelpers.NewAuthorizedTransport(t, ctx, reg, session), t).RoundTripper

				v := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
				v.Set("recovery_address", "some-email@example.org")
				v.Set("method", "code")

				body, res := testhelpers.RecoveryMakeRequest(t, isAPI || isSPA, f, client, testhelpers.EncodeFormAsJSON(t, isAPI || isSPA, v))

				if isAPI || isSPA {
					assert.EqualValues(t, http.StatusBadRequest, res.StatusCode, "%s", body)
					assert.Contains(t, res.Request.URL.String(), recovery.RouteSubmitFlow, "%+v\n\t%s", res.Request, body)
					assertx.EqualAsJSONExcept(t, recovery.ErrAlreadyLoggedIn, json.RawMessage(gjson.Get(body, "error").Raw), nil)
				} else {
					assert.EqualValues(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.Contains(t, res.Request.URL.String(), conf.SelfServiceBrowserDefaultReturnTo(ctx).String(), "%+v\n\t%s", res.Request, body)
				}
			})
		}
	})

	t.Run("description=should not be able to recover account that does not exist", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryNotifyUnknownRecipients, true)

		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryNotifyUnknownRecipients, false)
		})

		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				email := x.NewUUID().String() + "@ory.sh"
				c := testCase.GetClient(t)
				withValues := func(v url.Values) {
					v.Set("recovery_address", email)
				}
				body := submitRecoveryFormInitial(t, c, testCase.ClientType, withValues, http.StatusOK)
				assert.EqualValues(t, node.CodeGroup, gjson.Get(body, "active").String(), "%s", body)
				assert.Empty(t, gjson.Get(body, "ui.nodes.#(attributes.name==code).attributes.value").String(), "%s", body)
				assertx.EqualAsJSON(t, text.NewRecoveryCodeRecoverySelectAddressSent(code.MaskAddress(email)), json.RawMessage(gjson.Get(body, "ui.messages.0").Raw))

				message := testhelpers.CourierExpectMessage(ctx, t, reg, email, "Account access attempted")
				assert.Contains(t, message.Body, "If this was you, check if you signed up using a different address.")
			})
		}
	})

	t.Run("description=should not be able to recover an inactive account", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				address := "recoverinactive_" + testCase.ClientType.String() + "@ory.sh"
				createIdentityToRecover(t, reg, address)
				values := func(v url.Values) {
					v.Set("recovery_address", address)
				}
				cl := testhelpers.NewClientWithCookies(t)

				body := submitRecoveryFormInitial(t, cl, testCase.ClientType, values, http.StatusOK)
				addr, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, address)
				assert.NoError(t, err)

				checkRecoveryScreenAskForCode(t, address, body)

				// Deactivate the identity
				require.NoError(t, reg.Persister().GetConnection(context.Background()).RawQuery("UPDATE identities SET state=? WHERE id = ?", identity.StateInactive, addr.IdentityID).Exec())

				code := testhelpers.ExpectStatusCode(testCase.ClientType == RecoveryClientTypeAPI || testCase.ClientType == RecoveryClientTypeSPA, http.StatusUnauthorized, http.StatusOK)
				body = extractCodeFromCourierAndSubmit(t, cl, testCase.ClientType, address, body, code)

				switch testCase.ClientType {
				case RecoveryClientTypeAPI:
					fallthrough
				case RecoveryClientTypeSPA:
					assertx.EqualAsJSON(t, session.ErrIdentityDisabled.WithDetail("identity_id", addr.IdentityID), json.RawMessage(gjson.Get(body, "error").Raw), "%s", body)
				default:
					assertx.EqualAsJSON(t, session.ErrIdentityDisabled.WithDetail("identity_id", addr.IdentityID), json.RawMessage(body), "%s", body)
				}
			})
		}
	})

	t.Run("description=should recover and invalidate all other sessions if hook is set", func(t *testing.T) {
		conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal), []config.SelfServiceHook{{Name: "revoke_active_sessions"}})
		t.Cleanup(func() {
			conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()), nil)
		})

		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				email := testhelpers.RandomEmail()
				id := createIdentityToRecover(t, reg, email)

				otherSession, err := testhelpers.NewActiveSession(httptest.NewRequest("GET", "/sessions/whoami", nil), reg, id, time.Now(), identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
				require.NoError(t, err)
				require.NoError(t, reg.SessionPersister().UpsertSession(ctx, otherSession))

				refetchedOtherSession, err := reg.SessionPersister().GetSession(ctx, otherSession.ID, session.ExpandNothing)
				require.NoError(t, err)
				assert.True(t, refetchedOtherSession.IsActive())

				cl := testCase.GetClient(t)
				body := recoverHappyPath(t, cl, testCase.ClientType, email)

				expectRedirectToSettings(t, cl, testCase.ClientType, body)

				refetchedOtherSession, err = reg.SessionPersister().GetSession(ctx, otherSession.ID, session.ExpandNothing)
				require.NoError(t, err)
				assert.False(t, refetchedOtherSession.IsActive())
			})
		}
	})

	t.Run("description=should not be able to use an invalid code more than 5 times", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				email := testhelpers.RandomEmail()
				createIdentityToRecover(t, reg, email)
				c := testCase.GetClient(t)

				body := submitRecoveryFormInitial(t, c, testCase.ClientType, func(v url.Values) {
					v.Set("recovery_address", email)
				}, http.StatusOK)

				checkRecoveryScreenAskForCode(t, email, body)

				initialFlowId := gjson.Get(body, "id")

				for range 5 {
					body := submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) { v.Set("code", "12312312") }, http.StatusOK)

					testhelpers.AssertMessage(t, []byte(body), "The recovery code is invalid or has already been used. Please try again.")
				}

				switch testCase.ClientType {
				case RecoveryClientTypeBrowser:
					// submit an invalid code for the 6th time
					body := submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) { v.Set("code", "12312312") }, http.StatusOK)

					require.Len(t, gjson.Get(body, "ui.messages").Array(), 1, "%s", body)
					assert.Equal(t, "The request was submitted too often. Please request another code.", gjson.Get(body, "ui.messages.0.text").String())

					// check that a new flow has been created
					assert.NotEqual(t, gjson.Get(body, "id"), initialFlowId)

					assert.True(t, gjson.Get(body, "ui.nodes.#(attributes.name==recovery_address)").Exists())
				case RecoveryClientTypeSPA:
					fallthrough
				case RecoveryClientTypeAPI:
					// submit an invalid code for the 6th time
					body := submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) { v.Set("code", "12312312") }, http.StatusBadRequest)

					assert.Equal(t, "Bad Request", gjson.Get(body, "error.status").String(), "%s", body)
					assert.Equal(t, "The request was submitted too often. Please request another code.", gjson.Get(body, "error.reason").String(), "%s", body)
					continueWith := gjson.Get(body, "error.details.continue_with").Array()
					assert.Len(t, continueWith, 1, "%s", body)
					assert.Equal(t, "show_recovery_ui", continueWith[0].Get("action").String(), "%s", body)
					flowId := continueWith[0].Get("flow.id").String()
					assert.NotEmpty(t, flowId, "%s", body)
					require.NotEqual(t, flowId, initialFlowId, "%s", body)

					flow, err := reg.Persister().GetRecoveryFlow(ctx, uuid.Must(uuid.FromString(flowId)))
					require.NoError(t, err)
					assert.Len(t, flow.UI.Messages, 1, "%+v", flow)
					assert.Equal(t, "The request was submitted too often. Please request another code.", flow.UI.Messages[0].Text)
				}
			})
		}
	})

	t.Run("description=should be able to recover after using invalid code", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				c := testCase.GetClient(t)
				recoveryEmail := testhelpers.RandomEmail()
				_ = createIdentityToRecover(t, reg, recoveryEmail)

				body := submitRecoveryFormInitial(t, c, testCase.ClientType, func(v url.Values) {
					v.Set("recovery_address", recoveryEmail)
				}, http.StatusOK)

				checkRecoveryScreenAskForCode(t, recoveryEmail, body)

				// Submit invalid code.
				body = submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) { v.Set("code", "12312312") }, http.StatusOK)
				flowId := gjson.Get(body, "id").String()
				require.NotEmpty(t, flowId)

				rs, res, err := testhelpers.
					NewSDKCustomClient(public, c).
					FrontendAPI.GetRecoveryFlow(context.Background()).
					Id(flowId).
					Execute()

				require.NoError(t, err)
				getBody := ioutilx.MustReadAll(res.Body)
				require.NotEmpty(t, getBody)

				require.Len(t, rs.Ui.Messages, 1)
				assert.Equal(t, "The recovery code is invalid or has already been used. Please try again.", rs.Ui.Messages[0].Text)

				// Now submit the correct code
				body = extractCodeFromCourierAndSubmit(t, c, testCase.ClientType, recoveryEmail, body, http.StatusOK)

				switch testCase.ClientType {
				case RecoveryClientTypeBrowser:
					assert.Len(t, gjson.Get(body, "ui.messages").Array(), 1)
					assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "You successfully recovered your account.")
				case RecoveryClientTypeSPA:
					require.Len(t, c.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 2)
					cookies := spew.Sdump(c.Jar.Cookies(urlx.ParseOrPanic(public.URL)))
					assert.Contains(t, cookies, "ory_kratos_session")

					require.NotEmpty(t, gjson.Get(body, "continue_with.#(action==show_settings_ui).flow").String(), "%s", body)
				case RecoveryClientTypeAPI:
					require.NotEmpty(t, gjson.Get(body, "continue_with.#(action==show_settings_ui).flow").String(), "%s", body)
					require.NotEmpty(t, gjson.Get(body, "continue_with.#(action==set_ory_session_token).ory_session_token").String(), "%s", body)
				}
			})
		}
	})

	t.Run("description=should not break ui if empty code is submitted", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				recoveryEmail := testhelpers.RandomEmail()
				createIdentityToRecover(t, reg, recoveryEmail)

				c := testCase.GetClient(t)
				body := submitRecoveryFormInitial(t, c, testCase.ClientType, func(v url.Values) {
					v.Set("recovery_address", recoveryEmail)
				}, http.StatusOK)

				checkRecoveryScreenAskForCode(t, recoveryEmail, body)

				body = submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) {
					v.Set("code", "")
					v.Set("recovery_confirm_address", recoveryEmail)
				}, http.StatusOK)

				// Not an error, just handle it as a code resend.
				testhelpers.AssertMessage(t, []byte(body), text.NewRecoveryCodeRecoverySelectAddressSent(code.MaskAddress(recoveryEmail)).Text)
			})
		}
	})

	t.Run("description=should be able to resend the recovery code", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				recoveryEmail := testhelpers.RandomEmail()
				createIdentityToRecover(t, reg, recoveryEmail)

				c := testCase.GetClient(t)

				body := submitRecoveryFormInitial(t, c, testCase.ClientType, func(v url.Values) {
					v.Set("recovery_address", recoveryEmail)
				}, http.StatusOK)

				checkRecoveryScreenAskForCode(t, recoveryEmail, body)

				body = submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) {
					v.Set("code", "")
					v.Set("recovery_confirm_address", recoveryEmail) // Trigger resend.
				}, http.StatusOK)

				action := gjson.Get(body, "ui.action").String()
				require.NotEmpty(t, action)
				assert.Equal(t, recoveryEmail, gjson.Get(body, "ui.nodes.#(attributes.name==recovery_confirm_address).attributes.value").String())
				assert.True(t, gjson.Get(body, "ui.nodes.#(attributes.name==code)").Exists())

				body = extractCodeFromCourierAndSubmit(t, c, testCase.ClientType, recoveryEmail, body, http.StatusOK)
				expectRedirectToSettings(t, c, testCase.ClientType, body)
			})
		}
	})

	t.Run("description=should not be able to use first code after re-sending email", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				recoveryEmail := testhelpers.RandomEmail()
				createIdentityToRecover(t, reg, recoveryEmail)

				c := testCase.GetClient(t)

				body := submitRecoveryFormInitial(t, c, testCase.ClientType, func(v url.Values) {
					v.Set("recovery_address", recoveryEmail)
				}, http.StatusOK)

				checkRecoveryScreenAskForCode(t, recoveryEmail, body)

				message1 := testhelpers.CourierExpectMessage(ctx, t, reg, recoveryEmail, "Use code")
				assert.Contains(t, message1.Body, "Recover access to your account by entering")
				recoveryCode1 := testhelpers.CourierExpectCodeInMessage(t, message1, 1)
				assert.NotEmpty(t, recoveryCode1)

				body = submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) {
					v.Set("code", "")
					v.Set("recovery_confirm_address", recoveryEmail) // Trigger resend.
				}, http.StatusOK)

				action := gjson.Get(body, "ui.action").String()
				require.NotEmpty(t, action)
				assert.Equal(t, recoveryEmail, gjson.Get(body, "ui.nodes.#(attributes.name==recovery_confirm_address).attributes.value").String())
				assert.True(t, gjson.Get(body, "ui.nodes.#(attributes.name==code)").Exists())

				// Try to submit the old (expired) code.
				body = submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) {
					v.Set("code", recoveryCode1)
				}, http.StatusOK)
				testhelpers.AssertMessage(t, []byte(body), "The recovery code is invalid or has already been used. Please try again.")

				// Send the right code.
				body = extractCodeFromCourierAndSubmit(t, c, testCase.ClientType, recoveryEmail, body, http.StatusOK)
				expectRedirectToSettings(t, c, testCase.ClientType, body)
			})
		}
	})

	t.Run("description=should recover if post recovery hook is successful", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal), []config.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})
				t.Cleanup(func() {
					conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal), nil)
				})

				recoveryEmail := testhelpers.RandomEmail()
				createIdentityToRecover(t, reg, recoveryEmail)

				cl := testCase.GetClient(t)

				body := recoverHappyPath(t, cl, testCase.ClientType, recoveryEmail)
				expectRedirectToSettings(t, cl, testCase.ClientType, body)
			})
		}
	})

	t.Run("description=should not be able to recover if post recovery hook fails", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal), []config.SelfServiceHook{{Name: "err", Config: []byte(`{"ExecutePostRecoveryHook": "err"}`)}})
				t.Cleanup(func() {
					conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal), nil)
				})

				recoveryEmail := testhelpers.RandomEmail()
				createIdentityToRecover(t, reg, recoveryEmail)

				cl := testhelpers.NewClientWithCookies(t)

				body := submitRecoveryFormInitial(t, cl, testCase.ClientType, func(v url.Values) {
					v.Set("recovery_address", recoveryEmail)
				}, http.StatusOK)

				checkRecoveryScreenAskForCode(t, recoveryEmail, body)

				assert.Equal(t, recoveryEmail, gjson.Get(body, "ui.nodes.#(attributes.name==recovery_confirm_address).attributes.value").String())

				cl.CheckRedirect = func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				}

				initialFlowId := gjson.Get(body, "id")
				switch testCase.ClientType {
				case RecoveryClientTypeBrowser:
					body = extractCodeFromCourierAndSubmit(t, cl, testCase.ClientType, recoveryEmail, body, http.StatusSeeOther)
					assert.NotEqual(t, gjson.Get(body, "id"), initialFlowId)

					require.Len(t, cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 1)
					cookies := spew.Sdump(cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)))
					assert.NotContains(t, cookies, "ory_kratos_session")
				case RecoveryClientTypeSPA:
					body = extractCodeFromCourierAndSubmit(t, cl, testCase.ClientType, recoveryEmail, body, http.StatusBadRequest)
					assert.NotEqual(t, gjson.Get(body, "id"), initialFlowId)

					require.Len(t, cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 1)
					cookies := spew.Sdump(cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)))
					assert.NotContains(t, cookies, "ory_kratos_session")
				case RecoveryClientTypeAPI:
					body = extractCodeFromCourierAndSubmit(t, cl, testCase.ClientType, recoveryEmail, body, http.StatusBadRequest)
					assert.NotEqual(t, gjson.Get(body, "id"), initialFlowId)
					require.Equal(t, "err", gjson.Get(body, "error.message").String(), "%s", body)
				}
			})
		}
	})

	t.Run("choose different address", func(t *testing.T) {
		recoveryEmail := testhelpers.RandomEmail()
		wrongEmail := testhelpers.RandomEmail()

		createIdentityToRecover(t, reg, recoveryEmail)

		client := testhelpers.NewClientWithCookies(t)

		recoverySubmissionResponse := submitRecoveryFormInitial(t, client, RecoveryClientTypeBrowser, func(v url.Values) {
			v.Set("recovery_address", wrongEmail)
		}, http.StatusOK)
		assert.True(t, gjson.Get(recoverySubmissionResponse, "ui.nodes.#(attributes.name==code)").Exists(), "%s", recoverySubmissionResponse)

		submitRecoveryFormSubsequent(t, client, recoverySubmissionResponse, RecoveryClientTypeBrowser, func(v url.Values) { v.Set("screen", "previous") }, http.StatusOK)
		recoverHappyPath(t, client, RecoveryClientTypeBrowser, recoveryEmail)
	})
}

func createIdentityToRecoverPhone(t *testing.T, reg *driver.RegistryDefault, address string) *identity.Identity {
	t.Helper()
	id := &identity.Identity{
		Credentials: map[identity.CredentialsType]identity.Credentials{
			"password": {
				Type:        "password",
				Identifiers: []string{address},
				Config:      sqlxx.JSONRawMessage(`{"hashed_password":"$2a$08$.cOYmAd.vCpDOoiVJrO5B.hjTLKQQ6cAK40u8uB.FnZDyPvVvQ9Q."}`),
			},
		},
		Traits:   identity.Traits(fmt.Sprintf(`{"phone":"%s"}`, address)),
		SchemaID: config.DefaultIdentityTraitsSchemaID,
		State:    identity.StateActive,
	}
	require.NoError(t, reg.IdentityManager().Create(context.Background(), id, identity.ManagerAllowWriteProtectedTraits))

	return id
}

// Recovery V2 is only tested with `ContinueWith`.
func TestRecovery_V2_WithContinueWith_OneAddress_Phone(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.StrategyEnable(t, conf, string(recovery.RecoveryStrategyCode), true)
	testhelpers.StrategyEnable(t, conf, string(recovery.RecoveryStrategyLink), false)
	conf.MustSet(ctx, config.ViperKeyUseContinueWithTransitions, true)
	conf.MustSet(ctx, config.ViperKeyChooseRecoveryAddress, true)

	initViper(t, ctx, conf)

	_ = testhelpers.NewRecoveryUIFlowEchoServer(t, reg)
	_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
	_ = testhelpers.NewSettingsUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)

	public, _, _, _ := testhelpers.NewKratosServerWithCSRFAndRouters(t, reg)

	submitRecoveryFormInitial := func(t *testing.T, client *http.Client, flowType ClientType, values func(url.Values), code int) string {
		isSPA := flowType == RecoveryClientTypeSPA
		isAPI := flowType == RecoveryClientTypeAPI
		if client == nil {
			client = testhelpers.NewDebugClient(t)
			if !isAPI {
				client = testhelpers.NewClientWithCookies(t)
				client.Transport = testhelpers.NewTransportWithLogger(http.DefaultTransport, t).RoundTripper
			}
		}

		expectedUrl := testhelpers.ExpectURL(isAPI || isSPA, public.URL+recovery.RouteSubmitFlow, conf.SelfServiceFlowRecoveryUI(ctx).String())
		return testhelpers.SubmitRecoveryForm(t, isAPI, isSPA, client, public, values, code, expectedUrl)
	}

	submitRecoveryFormSubsequent := func(t *testing.T, client *http.Client, flow string, flowType ClientType, urlValuesFn func(url.Values), statusCode int) string {
		t.Helper()
		action := gjson.Get(flow, "ui.action").String()
		assert.NotEmpty(t, action)

		urlValues := url.Values{}
		urlValuesFn(urlValues)
		values := withCSRFToken(t, flowType, flow, urlValues)

		contentType := "application/json"
		if flowType == RecoveryClientTypeBrowser {
			contentType = "application/x-www-form-urlencoded"
		}

		res, err := client.Post(action, contentType, bytes.NewBufferString(values))
		require.NoError(t, err)
		assert.Equal(t, statusCode, res.StatusCode)

		return string(ioutilx.MustReadAll(res.Body))
	}

	extractCodeFromCourierAndSubmit := func(t *testing.T, client *http.Client, flowType ClientType, chosenRecoveryConfirmAddress string, recoverySubmissionResponse string, expectedCode int) string {
		message := testhelpers.CourierExpectMessage(ctx, t, reg, chosenRecoveryConfirmAddress, "")
		assert.Contains(t, message.Body, "Your recovery code is")

		recoveryCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
		assert.NotEmpty(t, recoveryCode)

		return submitRecoveryFormSubsequent(t, client, recoverySubmissionResponse, flowType, func(v url.Values) { v.Set("code", recoveryCode) }, expectedCode)
	}

	recoverHappyPath := func(t *testing.T, client *http.Client, clientType ClientType, anyAddress string) string {
		recoverySubmissionResponse := submitRecoveryFormInitial(t, client, clientType, func(v url.Values) {
			v.Set("recovery_address", anyAddress)
		}, http.StatusOK)

		body := extractCodeFromCourierAndSubmit(t, client, clientType, anyAddress, recoverySubmissionResponse, http.StatusOK)
		return body
	}

	expectRedirectToSettings := func(t *testing.T, client *http.Client, clientType ClientType, body string) {
		switch clientType {
		case RecoveryClientTypeBrowser:
			require.Len(t, client.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 2)
			cookies := spew.Sdump(client.Jar.Cookies(urlx.ParseOrPanic(public.URL)))
			assert.Contains(t, cookies, "ory_kratos_session")
			require.Contains(t, body, "You successfully recovered your account. Please change your password or set up an alternative login method (e.g. social sign in) within the next 60.00 minutes.")
		case RecoveryClientTypeSPA:
			require.Len(t, client.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 2)
			cookies := spew.Sdump(client.Jar.Cookies(urlx.ParseOrPanic(public.URL)))
			assert.Contains(t, cookies, "ory_kratos_session")

			require.NotEmpty(t, gjson.Get(body, "continue_with.#(action==show_settings_ui).flow").String(), "%s", body)
		case RecoveryClientTypeAPI:
			require.NotEmpty(t, gjson.Get(body, "continue_with.#(action==show_settings_ui).flow").String(), "%s", body)
			require.NotEmpty(t, gjson.Get(body, "continue_with.#(action==set_ory_session_token).ory_session_token").String(), "%s", body)
		}
	}

	t.Run("description=should recover an account", func(t *testing.T) {
		t.Run("type=browser", func(t *testing.T) {
			client := testhelpers.NewClientWithCookies(t)
			address := testhelpers.RandomPhone()
			createIdentityToRecoverPhone(t, reg, address)

			body := recoverHappyPath(t, client, RecoveryClientTypeBrowser, address)

			assert.Equal(t, text.NewRecoverySuccessful(time.Now().Add(time.Hour)).Text,
				gjson.Get(body, "ui.messages.0.text").String())

			res, err := client.Get(public.URL + session.RouteWhoami)
			require.NoError(t, err)
			body = string(x.MustReadAll(res.Body))
			require.NoError(t, res.Body.Close())
			assert.Equal(t, "code_recovery", gjson.Get(body, "authentication_methods.0.method").String(), "%s", body)
			assert.Equal(t, "aal1", gjson.Get(body, "authenticator_assurance_level").String(), "%s", body)
		})

		t.Run("type=spa", func(t *testing.T) {
			client := testhelpers.NewClientWithCookies(t)
			address := testhelpers.RandomPhone()
			createIdentityToRecoverPhone(t, reg, address)

			body := recoverHappyPath(t, client, RecoveryClientTypeSPA, address)

			assert.Equal(t, "passed_challenge", gjson.Get(body, "state").String())
			assert.Len(t, gjson.Get(body, "continue_with").Array(), 1)
			sfId := gjson.Get(body, "continue_with.#(action==show_settings_ui).flow.id").String()
			assert.NotEmpty(t, uuid.Must(uuid.FromString(sfId)))
		})

		t.Run("type=api", func(t *testing.T) {
			client := &http.Client{}
			address := testhelpers.RandomPhone()
			createIdentityToRecoverPhone(t, reg, address)

			body := recoverHappyPath(t, client, RecoveryClientTypeAPI, address)

			assert.Equal(t, "passed_challenge", gjson.Get(body, "state").String())
			assert.Len(t, gjson.Get(body, "continue_with").Array(), 2)
			assert.NotEmpty(t, gjson.Get(body, "continue_with.#(action==set_ory_session_token).ory_session_token").String())
			sfId := gjson.Get(body, "continue_with.#(action==show_settings_ui).flow.id").String()
			assert.NotEmpty(t, uuid.Must(uuid.FromString(sfId)))
		})

		t.Run("description=should return browser to return url", func(t *testing.T) {
			returnTo := public.URL + "/return-to"
			conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{returnTo})
			for _, tc := range []struct {
				desc        string
				returnTo    string
				f           func(t *testing.T, client *http.Client, identity *identity.Identity) *kratos.RecoveryFlow
				expectedAAL string
			}{
				{
					desc:     "should use return_to from recovery flow",
					returnTo: returnTo,
					f: func(t *testing.T, client *http.Client, identity *identity.Identity) *kratos.RecoveryFlow {
						return testhelpers.InitializeRecoveryFlowViaBrowser(t, client, false, public, url.Values{"return_to": []string{returnTo}})
					},
				},
				{
					desc:     "should use return_to from config",
					returnTo: returnTo,
					f: func(t *testing.T, client *http.Client, identity *identity.Identity) *kratos.RecoveryFlow {
						conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryBrowserDefaultReturnTo, returnTo)
						t.Cleanup(func() {
							conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryBrowserDefaultReturnTo, "")
						})
						return testhelpers.InitializeRecoveryFlowViaBrowser(t, client, false, public, nil)
					},
				},
				{
					desc:     "no return to",
					returnTo: "",
					f: func(t *testing.T, client *http.Client, identity *identity.Identity) *kratos.RecoveryFlow {
						return testhelpers.InitializeRecoveryFlowViaBrowser(t, client, false, public, nil)
					},
				},
				{
					desc:     "should use return_to with an account that has 2fa enabled",
					returnTo: returnTo,
					f: func(t *testing.T, client *http.Client, id *identity.Identity) *kratos.RecoveryFlow {
						conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, config.HighestAvailableAAL)
						conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, config.HighestAvailableAAL)
						conf.MustSet(ctx, config.ViperKeyWebAuthnRPDisplayName, "Kratos")
						conf.MustSet(ctx, config.ViperKeyWebAuthnRPID, "ory.sh")

						t.Cleanup(func() {
							conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, identity.AuthenticatorAssuranceLevel1)
							conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, identity.AuthenticatorAssuranceLevel1)
						})
						testhelpers.StrategyEnable(t, conf, identity.CredentialsTypeWebAuthn.String(), true)

						id.SetCredentials(identity.CredentialsTypeWebAuthn, identity.Credentials{
							Type:        identity.CredentialsTypeWebAuthn,
							Config:      []byte(`{"credentials":[{"is_passwordless":false, "display_name":"test"}]}`),
							Identifiers: []string{testhelpers.RandomPhone()},
						})

						require.NoError(t, reg.IdentityManager().Update(ctx, id, identity.ManagerAllowWriteProtectedTraits))
						return testhelpers.InitializeRecoveryFlowViaBrowser(t, client, false, public, url.Values{"return_to": []string{returnTo}})
					},
					expectedAAL: "aal2",
				},
			} {
				t.Run(tc.desc, func(t *testing.T) {
					client := testhelpers.NewClientWithCookies(t)
					address := testhelpers.RandomPhone()
					i := createIdentityToRecoverPhone(t, reg, address)

					client.Transport = testhelpers.NewTransportWithLogger(http.DefaultTransport, t).RoundTripper
					f := tc.f(t, client, i)

					formPayload := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
					formPayload.Set("recovery_address", address)

					body, res := testhelpers.RecoveryMakeRequest(t, false, f, client, formPayload.Encode())
					assert.EqualValues(t, http.StatusOK, res.StatusCode, "%s", body)

					body = extractCodeFromCourierAndSubmit(t, client, RecoveryClientTypeBrowser, address, body, http.StatusOK)

					require.Equal(t, text.NewRecoverySuccessful(time.Now().Add(time.Hour)).Text,
						gjson.Get(body, "ui.messages.0.text").String())

					settingsId := gjson.Get(body, "id").String()

					sf, err := reg.SettingsFlowPersister().GetSettingsFlow(ctx, uuid.Must(uuid.FromString(settingsId)))
					require.NoError(t, err)

					u, err := url.Parse(public.URL)
					require.NoError(t, err)
					require.Len(t, client.Jar.Cookies(u), 2)
					found := false
					for _, cookie := range client.Jar.Cookies(u) {
						if cookie.Name == "ory_kratos_session" {
							found = true
						}
					}
					require.True(t, found)

					require.Equal(t, tc.returnTo, sf.ReturnTo)
					res, err = client.Get(public.URL + session.RouteWhoami)
					require.NoError(t, err)
					body = string(x.MustReadAll(res.Body))
					require.NoError(t, res.Body.Close())

					if tc.expectedAAL == "aal2" {
						require.Equal(t, http.StatusForbidden, res.StatusCode)
						require.Equalf(t, session.NewErrAALNotSatisfied("").Reason(), gjson.Get(body, "error.reason").String(), "%s", body)
						require.Equalf(t, "session_aal2_required", gjson.Get(body, "error.id").String(), "%s", body)
					} else {
						assert.Equal(t, "code_recovery", gjson.Get(body, "authentication_methods.0.method").String(), "%s", body)
						assert.Equal(t, "aal1", gjson.Get(body, "authenticator_assurance_level").String(), "%s", body)
					}
				})
			}
		})
	})

	t.Run("description=should set all the correct recovery payloads after submission", func(t *testing.T) {
		fakes := []string{"+491705550176", "+491705550177", "+491705550178"}
		fakeIdx := 0

		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				address := fakes[fakeIdx]
				fakeIdx += 1

				createIdentityToRecoverPhone(t, reg, address)
				body := submitRecoveryFormInitial(t, testCase.GetClient(t), testCase.ClientType, func(u url.Values) { u.Set("recovery_address", address) }, http.StatusOK)
				testhelpers.SnapshotTExcept(t, json.RawMessage(gjson.Get(body, "ui.nodes").String()), []string{"0.attributes.value"})
			})
		}
	})

	t.Run("description=should set all the correct recovery payloads", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				c := testCase.GetClient(t)
				rs := testhelpers.GetRecoveryFlowForType(t, c, public, testCase.FlowType)

				testhelpers.SnapshotTExcept(t, rs.Ui.Nodes, []string{"0.attributes.value"})
				assert.EqualValues(t, public.URL+recovery.RouteSubmitFlow+"?flow="+rs.Id, rs.Ui.Action)
				assert.Empty(t, rs.Ui.Messages)
			})
		}
	})

	t.Run("description=should require an address to be sent", func(t *testing.T) {
		for _, flowType := range flowTypes {
			t.Run("type="+flowType.String(), func(t *testing.T) {
				code := testhelpers.ExpectStatusCode(flowType == RecoveryClientTypeAPI || flowType == RecoveryClientTypeSPA, http.StatusBadRequest, http.StatusOK)
				body := submitRecoveryFormInitial(t, nil, flowType, func(v url.Values) {
					v.Del("recovery_address")
				}, code)
				assert.EqualValues(t, node.CodeGroup, gjson.Get(body, "active").String(), "%s", body)
				assert.EqualValues(t, "Property recovery_address is missing.",
					gjson.Get(body, "ui.nodes.#(attributes.name==recovery_address).messages.0.text").String(),
					"%s", body)
			})
		}
	})

	t.Run("description=should require an existing address to be sent", func(t *testing.T) {
		for _, flowType := range flowTypes {
			t.Run("type="+flowType.String(), func(t *testing.T) {
				for _, address := range []string{"\\", "asdf", "...", testhelpers.RandomPhone() + "," + testhelpers.RandomPhone()} {
					body := submitRecoveryFormInitial(t, nil, flowType, func(v url.Values) {
						v.Set("recovery_address", address)
					}, http.StatusOK)

					activeMethod := gjson.Get(body, "active").String()
					assert.EqualValues(t, node.CodeGroup, activeMethod, "expected method to be %s got %s", node.CodeGroup, activeMethod)
					expectedMessage := text.NewRecoveryCodeRecoverySelectAddressSent(code.MaskAddress(address)).Text
					actualMessage := gjson.Get(body, "ui.messages.0.text").String()
					assert.EqualValues(t, expectedMessage, actualMessage, "%s", body)
				}
			})
		}
	})

	t.Run("description=should try to submit the form while authenticated", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				isSPA := testCase.ClientType == "spa"
				isAPI := testCase.ClientType == "api"
				client := testCase.GetClient(t)

				var f *kratos.RecoveryFlow
				if isAPI {
					f = testhelpers.InitializeRecoveryFlowViaAPI(t, client, public)
				} else {
					f = testhelpers.InitializeRecoveryFlowViaBrowser(t, client, isSPA, public, nil)
				}
				req := httptest.NewRequest("GET", "/sessions/whoami", nil).WithContext(contextx.WithConfigValue(ctx, config.ViperKeySessionLifespan, time.Hour))

				session, err := testhelpers.NewActiveSession(
					req,
					reg,
					&identity.Identity{ID: x.NewUUID(), State: identity.StateActive, NID: x.NewUUID()},
					time.Now(),
					identity.CredentialsTypePassword,
					identity.AuthenticatorAssuranceLevel1,
				)

				require.NoError(t, err)

				// Add the authentication to the request
				client.Transport = testhelpers.NewTransportWithLogger(testhelpers.NewAuthorizedTransport(t, ctx, reg, session), t).RoundTripper

				v := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
				v.Set("recovery_address", testhelpers.RandomPhone())
				v.Set("method", "code")

				body, res := testhelpers.RecoveryMakeRequest(t, isAPI || isSPA, f, client, testhelpers.EncodeFormAsJSON(t, isAPI || isSPA, v))

				if isAPI || isSPA {
					assert.EqualValues(t, http.StatusBadRequest, res.StatusCode, "%s", body)
					assert.Contains(t, res.Request.URL.String(), recovery.RouteSubmitFlow, "%+v\n\t%s", res.Request, body)
					assertx.EqualAsJSONExcept(t, recovery.ErrAlreadyLoggedIn, json.RawMessage(gjson.Get(body, "error").Raw), nil)
				} else {
					assert.EqualValues(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.Contains(t, res.Request.URL.String(), conf.SelfServiceBrowserDefaultReturnTo(ctx).String(), "%+v\n\t%s", res.Request, body)
				}
			})
		}
	})

	t.Run("description=should not be able to recover account that does not exist", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryNotifyUnknownRecipients, true)

		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryNotifyUnknownRecipients, false)
		})

		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				address := testhelpers.RandomPhone()
				c := testCase.GetClient(t)
				withValues := func(v url.Values) {
					v.Set("recovery_address", address)
				}
				body := submitRecoveryFormInitial(t, c, testCase.ClientType, withValues, http.StatusOK)
				assert.EqualValues(t, node.CodeGroup, gjson.Get(body, "active").String(), "%s", body)
				assert.Empty(t, gjson.Get(body, "ui.nodes.#(attributes.name==code).attributes.value").String(), "%s", body)
				assertx.EqualAsJSON(t, text.NewRecoveryCodeRecoverySelectAddressSent(code.MaskAddress(address)), json.RawMessage(gjson.Get(body, "ui.messages.0").Raw))
			})
		}
	})

	t.Run("description=should not be able to recover an inactive account", func(t *testing.T) {
		fakes := []string{"+491705550173", "+491705550174", "+491705550175"}
		fakeIdx := 0

		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				address := fakes[fakeIdx]
				fakeIdx += 1

				i := createIdentityToRecoverPhone(t, reg, address)
				values := func(v url.Values) {
					v.Set("recovery_address", address)
				}
				cl := testhelpers.NewClientWithCookies(t)

				body := submitRecoveryFormInitial(t, cl, testCase.ClientType, values, http.StatusOK)

				// Deactivate the identity
				require.NoError(t, reg.Persister().GetConnection(context.Background()).RawQuery("UPDATE identities SET state=? WHERE id = ?", identity.StateInactive, i.ID).Exec())

				code := testhelpers.ExpectStatusCode(testCase.ClientType == RecoveryClientTypeAPI || testCase.ClientType == RecoveryClientTypeSPA, http.StatusUnauthorized, http.StatusOK)
				body = extractCodeFromCourierAndSubmit(t, cl, testCase.ClientType, address, body, code)

				switch testCase.ClientType {
				case RecoveryClientTypeAPI:
					fallthrough
				case RecoveryClientTypeSPA:
					assertx.EqualAsJSON(t, session.ErrIdentityDisabled.WithDetail("identity_id", i.ID), json.RawMessage(gjson.Get(body, "error").Raw), "%s", body)
				default:
					assertx.EqualAsJSON(t, session.ErrIdentityDisabled.WithDetail("identity_id", i.ID), json.RawMessage(body), "%s", body)
				}
			})
		}
	})

	t.Run("description=should recover and invalidate all other sessions if hook is set", func(t *testing.T) {
		conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal), []config.SelfServiceHook{{Name: "revoke_active_sessions"}})
		t.Cleanup(func() {
			conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()), nil)
		})

		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				address := testhelpers.RandomPhone()
				id := createIdentityToRecoverPhone(t, reg, address)

				otherSession, err := testhelpers.NewActiveSession(httptest.NewRequest("GET", "/sessions/whoami", nil), reg, id, time.Now(), identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
				require.NoError(t, err)
				require.NoError(t, reg.SessionPersister().UpsertSession(ctx, otherSession))

				refetchedOtherSession, err := reg.SessionPersister().GetSession(ctx, otherSession.ID, session.ExpandNothing)
				require.NoError(t, err)
				assert.True(t, refetchedOtherSession.IsActive())

				cl := testCase.GetClient(t)
				body := recoverHappyPath(t, cl, testCase.ClientType, address)

				expectRedirectToSettings(t, cl, testCase.ClientType, body)

				refetchedOtherSession, err = reg.SessionPersister().GetSession(ctx, otherSession.ID, session.ExpandNothing)
				require.NoError(t, err)
				assert.False(t, refetchedOtherSession.IsActive())
			})
		}
	})

	t.Run("description=should not be able to use an invalid code more than 5 times", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				address := testhelpers.RandomPhone()
				createIdentityToRecoverPhone(t, reg, address)
				c := testCase.GetClient(t)

				body := submitRecoveryFormInitial(t, c, testCase.ClientType, func(v url.Values) {
					v.Set("recovery_address", address)
				}, http.StatusOK)

				initialFlowId := gjson.Get(body, "id")

				for range 5 {
					body := submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) { v.Set("code", "12312312") }, http.StatusOK)

					testhelpers.AssertMessage(t, []byte(body), "The recovery code is invalid or has already been used. Please try again.")
				}

				switch testCase.ClientType {
				case RecoveryClientTypeBrowser:
					// submit an invalid code for the 6th time
					body := submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) { v.Set("code", "12312312") }, http.StatusOK)

					require.Len(t, gjson.Get(body, "ui.messages").Array(), 1, "%s", body)
					assert.Equal(t, "The request was submitted too often. Please request another code.", gjson.Get(body, "ui.messages.0.text").String())

					// check that a new flow has been created
					assert.NotEqual(t, gjson.Get(body, "id"), initialFlowId)

					assert.True(t, gjson.Get(body, "ui.nodes.#(attributes.name==recovery_address)").Exists())
				case RecoveryClientTypeSPA:
					fallthrough
				case RecoveryClientTypeAPI:
					// submit an invalid code for the 6th time
					body := submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) { v.Set("code", "12312312") }, http.StatusBadRequest)

					assert.Equal(t, "Bad Request", gjson.Get(body, "error.status").String(), "%s", body)
					assert.Equal(t, "The request was submitted too often. Please request another code.", gjson.Get(body, "error.reason").String(), "%s", body)
					continueWith := gjson.Get(body, "error.details.continue_with").Array()
					assert.Len(t, continueWith, 1, "%s", body)
					assert.Equal(t, "show_recovery_ui", continueWith[0].Get("action").String(), "%s", body)
					flowId := continueWith[0].Get("flow.id").String()
					assert.NotEmpty(t, flowId, "%s", body)
					require.NotEqual(t, flowId, initialFlowId, "%s", body)

					flow, err := reg.Persister().GetRecoveryFlow(ctx, uuid.Must(uuid.FromString(flowId)))
					require.NoError(t, err)
					assert.Len(t, flow.UI.Messages, 1, "%+v", flow)
					assert.Equal(t, "The request was submitted too often. Please request another code.", flow.UI.Messages[0].Text)
				}
			})
		}
	})

	t.Run("description=should be able to recover after using invalid code", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				c := testCase.GetClient(t)
				recoveryAddress := testhelpers.RandomPhone()
				_ = createIdentityToRecoverPhone(t, reg, recoveryAddress)

				body := submitRecoveryFormInitial(t, c, testCase.ClientType, func(v url.Values) {
					v.Set("recovery_address", recoveryAddress)
				}, http.StatusOK)

				// Submit invalid code.
				body = submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) { v.Set("code", "12312312") }, http.StatusOK)
				flowId := gjson.Get(body, "id").String()
				require.NotEmpty(t, flowId)

				rs, res, err := testhelpers.
					NewSDKCustomClient(public, c).
					FrontendAPI.GetRecoveryFlow(context.Background()).
					Id(flowId).
					Execute()

				require.NoError(t, err)
				getBody := ioutilx.MustReadAll(res.Body)
				require.NotEmpty(t, getBody)

				require.Len(t, rs.Ui.Messages, 1)
				assert.Equal(t, "The recovery code is invalid or has already been used. Please try again.", rs.Ui.Messages[0].Text)

				// Now submit the correct code
				body = extractCodeFromCourierAndSubmit(t, c, testCase.ClientType, recoveryAddress, body, http.StatusOK)

				switch testCase.ClientType {
				case RecoveryClientTypeBrowser:
					assert.Len(t, gjson.Get(body, "ui.messages").Array(), 1)
					assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "You successfully recovered your account.")
				case RecoveryClientTypeSPA:
					require.Len(t, c.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 2)
					cookies := spew.Sdump(c.Jar.Cookies(urlx.ParseOrPanic(public.URL)))
					assert.Contains(t, cookies, "ory_kratos_session")

					require.NotEmpty(t, gjson.Get(body, "continue_with.#(action==show_settings_ui).flow").String(), "%s", body)
				case RecoveryClientTypeAPI:
					require.NotEmpty(t, gjson.Get(body, "continue_with.#(action==show_settings_ui).flow").String(), "%s", body)
					require.NotEmpty(t, gjson.Get(body, "continue_with.#(action==set_ory_session_token).ory_session_token").String(), "%s", body)
				}
			})
		}
	})

	t.Run("description=should not break ui if empty code is submitted", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				recoveryAddress := testhelpers.RandomPhone()
				createIdentityToRecoverPhone(t, reg, recoveryAddress)

				c := testCase.GetClient(t)
				body := submitRecoveryFormInitial(t, c, testCase.ClientType, func(v url.Values) {
					v.Set("recovery_address", recoveryAddress)
				}, http.StatusOK)

				body = submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) {
					v.Set("code", "")
					v.Set("recovery_confirm_address", recoveryAddress)
				}, http.StatusOK)

				// Not an error, just handle it as a code resend.
				testhelpers.AssertMessage(t, []byte(body), text.NewRecoveryCodeRecoverySelectAddressSent(code.MaskAddress(recoveryAddress)).Text)
			})
		}
	})

	t.Run("description=should be able to resend the recovery code", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				recoveryAddress := testhelpers.RandomPhone()
				createIdentityToRecoverPhone(t, reg, recoveryAddress)

				c := testCase.GetClient(t)

				body := submitRecoveryFormInitial(t, c, testCase.ClientType, func(v url.Values) {
					v.Set("recovery_address", recoveryAddress)
				}, http.StatusOK)

				body = submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) {
					v.Set("code", "")
					v.Set("recovery_confirm_address", recoveryAddress) // Trigger resend.
				}, http.StatusOK)

				action := gjson.Get(body, "ui.action").String()
				require.NotEmpty(t, action)
				assert.Equal(t, recoveryAddress, gjson.Get(body, "ui.nodes.#(attributes.name==recovery_confirm_address).attributes.value").String())
				assert.True(t, gjson.Get(body, "ui.nodes.#(attributes.name==code)").Exists())

				body = extractCodeFromCourierAndSubmit(t, c, testCase.ClientType, recoveryAddress, body, http.StatusOK)
				expectRedirectToSettings(t, c, testCase.ClientType, body)
			})
		}
	})

	t.Run("description=should not be able to use first code after re-sending address", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				recoveryAddress := testhelpers.RandomPhone()
				createIdentityToRecoverPhone(t, reg, recoveryAddress)

				c := testCase.GetClient(t)

				body := submitRecoveryFormInitial(t, c, testCase.ClientType, func(v url.Values) {
					v.Set("recovery_address", recoveryAddress)
				}, http.StatusOK)

				message1 := testhelpers.CourierExpectMessage(ctx, t, reg, recoveryAddress, "")
				assert.Contains(t, message1.Body, "Your recovery code is:")
				recoveryCode1 := testhelpers.CourierExpectCodeInMessage(t, message1, 1)
				assert.NotEmpty(t, recoveryCode1)

				body = submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) {
					v.Set("code", "")
					v.Set("recovery_confirm_address", recoveryAddress) // Trigger resend.
				}, http.StatusOK)

				action := gjson.Get(body, "ui.action").String()
				require.NotEmpty(t, action)
				assert.Equal(t, recoveryAddress, gjson.Get(body, "ui.nodes.#(attributes.name==recovery_confirm_address).attributes.value").String())
				assert.True(t, gjson.Get(body, "ui.nodes.#(attributes.name==code)").Exists())

				// Try to submit the old (expired) code.
				body = submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) {
					v.Set("code", recoveryCode1)
				}, http.StatusOK)
				testhelpers.AssertMessage(t, []byte(body), "The recovery code is invalid or has already been used. Please try again.")

				// Send the right code.
				body = extractCodeFromCourierAndSubmit(t, c, testCase.ClientType, recoveryAddress, body, http.StatusOK)
				expectRedirectToSettings(t, c, testCase.ClientType, body)
			})
		}
	})

	t.Run("description=should recover if post recovery hook is successful", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal), []config.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})
				t.Cleanup(func() {
					conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal), nil)
				})

				recoveryAddress := testhelpers.RandomPhone()
				createIdentityToRecoverPhone(t, reg, recoveryAddress)

				cl := testCase.GetClient(t)

				body := recoverHappyPath(t, cl, testCase.ClientType, recoveryAddress)
				expectRedirectToSettings(t, cl, testCase.ClientType, body)
			})
		}
	})

	t.Run("description=should not be able to recover if post recovery hook fails", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal), []config.SelfServiceHook{{Name: "err", Config: []byte(`{"ExecutePostRecoveryHook": "err"}`)}})
				t.Cleanup(func() {
					conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal), nil)
				})

				recoveryAddress := testhelpers.RandomPhone()
				createIdentityToRecoverPhone(t, reg, recoveryAddress)

				cl := testhelpers.NewClientWithCookies(t)

				body := submitRecoveryFormInitial(t, cl, testCase.ClientType, func(v url.Values) {
					v.Set("recovery_address", recoveryAddress)
				}, http.StatusOK)

				assert.Equal(t, recoveryAddress, gjson.Get(body, "ui.nodes.#(attributes.name==recovery_confirm_address).attributes.value").String())

				cl.CheckRedirect = func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				}

				initialFlowId := gjson.Get(body, "id")
				switch testCase.ClientType {
				case RecoveryClientTypeBrowser:
					body = extractCodeFromCourierAndSubmit(t, cl, testCase.ClientType, recoveryAddress, body, http.StatusSeeOther)
					assert.NotEqual(t, gjson.Get(body, "id"), initialFlowId)

					require.Len(t, cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 1)
					cookies := spew.Sdump(cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)))
					assert.NotContains(t, cookies, "ory_kratos_session")
				case RecoveryClientTypeSPA:
					body = extractCodeFromCourierAndSubmit(t, cl, testCase.ClientType, recoveryAddress, body, http.StatusBadRequest)
					assert.NotEqual(t, gjson.Get(body, "id"), initialFlowId)

					require.Len(t, cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 1)
					cookies := spew.Sdump(cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)))
					assert.NotContains(t, cookies, "ory_kratos_session")
				case RecoveryClientTypeAPI:
					body = extractCodeFromCourierAndSubmit(t, cl, testCase.ClientType, recoveryAddress, body, http.StatusBadRequest)
					assert.NotEqual(t, gjson.Get(body, "id"), initialFlowId)
					require.Equal(t, "err", gjson.Get(body, "error.message").String(), "%s", body)
				}
			})
		}
	})

	t.Run("choose different address", func(t *testing.T) {
		recoveryAddress := testhelpers.RandomPhone()
		wrongAddress := testhelpers.RandomPhone()
		createIdentityToRecoverPhone(t, reg, recoveryAddress)

		client := testhelpers.NewClientWithCookies(t)

		recoverySubmissionResponse := submitRecoveryFormInitial(t, client, RecoveryClientTypeBrowser, func(v url.Values) {
			v.Set("recovery_address", wrongAddress)
		}, http.StatusOK)
		assert.True(t, gjson.Get(recoverySubmissionResponse, "ui.nodes.#(attributes.name==code)").Exists(), "%s", recoverySubmissionResponse)

		submitRecoveryFormSubsequent(t, client, recoverySubmissionResponse, RecoveryClientTypeBrowser, func(v url.Values) { v.Set("screen", "previous") }, http.StatusOK)
		recoverHappyPath(t, client, RecoveryClientTypeBrowser, recoveryAddress)
	})
}

func createIdentityToRecoverEmailAndPhone(t *testing.T, reg *driver.RegistryDefault, email string, phone string) *identity.Identity {
	t.Helper()
	id := &identity.Identity{
		Credentials: map[identity.CredentialsType]identity.Credentials{
			"password": {
				Type:        "password",
				Identifiers: []string{phone},
				Config:      sqlxx.JSONRawMessage(`{"hashed_password":"$2a$08$.cOYmAd.vCpDOoiVJrO5B.hjTLKQQ6cAK40u8uB.FnZDyPvVvQ9Q."}`),
			},
		},
		Traits:   identity.Traits(fmt.Sprintf(`{"email":"%s", "phone":"%s"}`, email, phone)),
		SchemaID: config.DefaultIdentityTraitsSchemaID,
		State:    identity.StateActive,
	}
	require.NoError(t, reg.IdentityManager().Create(context.Background(), id, identity.ManagerAllowWriteProtectedTraits))

	return id
}

// Recovery V2 is only tested with `ContinueWith`.
func TestRecovery_V2_WithContinueWith_SeveralAddresses(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.StrategyEnable(t, conf, string(recovery.RecoveryStrategyCode), true)
	testhelpers.StrategyEnable(t, conf, string(recovery.RecoveryStrategyLink), false)
	conf.MustSet(ctx, config.ViperKeyUseContinueWithTransitions, true)
	conf.MustSet(ctx, config.ViperKeyChooseRecoveryAddress, true)

	initViper(t, ctx, conf)

	_ = testhelpers.NewRecoveryUIFlowEchoServer(t, reg)
	_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
	_ = testhelpers.NewSettingsUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)

	public, _, _, _ := testhelpers.NewKratosServerWithCSRFAndRouters(t, reg)

	submitRecoveryFormInitial := func(t *testing.T, client *http.Client, flowType ClientType, values func(url.Values), code int) string {
		isSPA := flowType == RecoveryClientTypeSPA
		isAPI := flowType == RecoveryClientTypeAPI
		if client == nil {
			client = testhelpers.NewDebugClient(t)
			if !isAPI {
				client = testhelpers.NewClientWithCookies(t)
				client.Transport = testhelpers.NewTransportWithLogger(http.DefaultTransport, t).RoundTripper
			}
		}

		expectedUrl := testhelpers.ExpectURL(isAPI || isSPA, public.URL+recovery.RouteSubmitFlow, conf.SelfServiceFlowRecoveryUI(ctx).String())
		return testhelpers.SubmitRecoveryForm(t, isAPI, isSPA, client, public, values, code, expectedUrl)
	}

	submitRecoveryFormSubsequent := func(t *testing.T, client *http.Client, flow string, flowType ClientType, urlValuesFn func(url.Values), statusCode int) string {
		t.Helper()
		action := gjson.Get(flow, "ui.action").String()
		assert.NotEmpty(t, action)

		urlValues := url.Values{}
		urlValuesFn(urlValues)
		values := withCSRFToken(t, flowType, flow, urlValues)

		contentType := "application/json"
		if flowType == RecoveryClientTypeBrowser {
			contentType = "application/x-www-form-urlencoded"
		}

		res, err := client.Post(action, contentType, bytes.NewBufferString(values))
		require.NoError(t, err)
		assert.Equal(t, statusCode, res.StatusCode)

		return string(ioutilx.MustReadAll(res.Body))
	}

	checkRecoveryScreenAskForRecoverySelectAddress := func(t *testing.T, recoverySubmissionResponse string) {
		assert.True(t, gjson.Get(recoverySubmissionResponse, "ui.nodes.#(attributes.name==recovery_select_address)").Exists(), "%s", recoverySubmissionResponse)
		assert.Len(t, gjson.Get(recoverySubmissionResponse, "ui.messages").Array(), 1, "%s", recoverySubmissionResponse)
		assertx.EqualAsJSON(t, text.NewRecoveryAskToChooseAddress(), json.RawMessage(gjson.Get(recoverySubmissionResponse, "ui.messages.0").Raw))
	}

	checkRecoveryScreenAskForRecoveryConfirmAddress := func(t *testing.T, recoverySubmissionResponse string) {
		assert.True(t, gjson.Get(recoverySubmissionResponse, "ui.nodes.#(attributes.name==recovery_confirm_address)").Exists(), "%s", recoverySubmissionResponse)
		assert.Len(t, gjson.Get(recoverySubmissionResponse, "ui.messages").Array(), 1, "%s", recoverySubmissionResponse)
		assertx.EqualAsJSON(t, text.NewRecoveryAskForFullAddress(), json.RawMessage(gjson.Get(recoverySubmissionResponse, "ui.messages.0").Raw))
	}

	checkRecoveryScreenAskForCode := func(t *testing.T, chosenRecoveryConfirmAddress, recoverySubmissionResponse string) {
		assert.True(t, gjson.Get(recoverySubmissionResponse, "ui.nodes.#(attributes.name==code)").Exists(), "%s", recoverySubmissionResponse)
		assert.Len(t, gjson.Get(recoverySubmissionResponse, "ui.messages").Array(), 1, "%s", recoverySubmissionResponse)
		assertx.EqualAsJSON(t, text.NewRecoveryCodeRecoverySelectAddressSent(code.MaskAddress(chosenRecoveryConfirmAddress)), json.RawMessage(gjson.Get(recoverySubmissionResponse, "ui.messages.0").Raw))
	}

	extractCodeFromCourierAndSubmit := func(t *testing.T, client *http.Client, flowType ClientType, chosenRecoveryConfirmAddress string, recoverySubmissionResponse string, expectedCode int) string {
		message := testhelpers.CourierExpectMessage(ctx, t, reg, chosenRecoveryConfirmAddress, "")

		// For some reason the wording is different between email and sms.
		if strings.ContainsRune(chosenRecoveryConfirmAddress, '@') {
			assert.Contains(t, message.Body, "Recover access to your account by entering")
		} else {
			assert.Contains(t, message.Body, "Your recovery code is")
		}

		recoveryCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
		assert.NotEmpty(t, recoveryCode)

		return submitRecoveryFormSubsequent(t, client, recoverySubmissionResponse, flowType, func(v url.Values) { v.Set("code", recoveryCode) }, expectedCode)
	}

	recoverHappyPath := func(t *testing.T, client *http.Client, clientType ClientType, anyAddress string, chosenAddress string) string {
		recoverySubmissionResponse := submitRecoveryFormInitial(t, client, clientType, func(v url.Values) {
			v.Set("recovery_address", anyAddress)
		}, http.StatusOK)

		checkRecoveryScreenAskForRecoverySelectAddress(t, recoverySubmissionResponse)
		recoverySubmissionResponse = submitRecoveryFormSubsequent(t, client, recoverySubmissionResponse, clientType, func(v url.Values) {
			v.Set("recovery_select_address", code.AddressToHashBase64(chosenAddress))
			v.Set("recovery_address", anyAddress)
		}, http.StatusOK)

		// If the first provided address is different from the chosen masked address,
		// the server asks the client to provide the chosen address in full.
		if anyAddress != chosenAddress {
			checkRecoveryScreenAskForRecoveryConfirmAddress(t, recoverySubmissionResponse)
			recoverySubmissionResponse = submitRecoveryFormSubsequent(t, client, recoverySubmissionResponse, clientType, func(v url.Values) {
				v.Set("recovery_confirm_address", chosenAddress)
			}, http.StatusOK)
		}

		checkRecoveryScreenAskForCode(t, chosenAddress, recoverySubmissionResponse)

		body := extractCodeFromCourierAndSubmit(t, client, clientType, chosenAddress, recoverySubmissionResponse, http.StatusOK)
		return body
	}

	expectRedirectToSettings := func(t *testing.T, client *http.Client, clientType ClientType, body string) {
		switch clientType {
		case RecoveryClientTypeBrowser:
			require.Len(t, client.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 2)
			cookies := spew.Sdump(client.Jar.Cookies(urlx.ParseOrPanic(public.URL)))
			assert.Contains(t, cookies, "ory_kratos_session")
			require.Contains(t, body, "You successfully recovered your account. Please change your password or set up an alternative login method (e.g. social sign in) within the next 60.00 minutes.")
		case RecoveryClientTypeSPA:
			require.Len(t, client.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 2)
			cookies := spew.Sdump(client.Jar.Cookies(urlx.ParseOrPanic(public.URL)))
			assert.Contains(t, cookies, "ory_kratos_session")

			require.NotEmpty(t, gjson.Get(body, "continue_with.#(action==show_settings_ui).flow").String(), "%s", body)
		case RecoveryClientTypeAPI:
			require.NotEmpty(t, gjson.Get(body, "continue_with.#(action==show_settings_ui).flow").String(), "%s", body)
			require.NotEmpty(t, gjson.Get(body, "continue_with.#(action==set_ory_session_token).ory_session_token").String(), "%s", body)
		}
	}

	t.Run("description=should recover an account", func(t *testing.T) {
		chosenAddressIdenticalToRecoveryAddressCases := []bool{true, false}

		for _, identical := range chosenAddressIdenticalToRecoveryAddressCases {

			t.Run("type=browser", func(t *testing.T) {
				client := testhelpers.NewClientWithCookies(t)
				address1 := testhelpers.RandomEmail()
				address2 := testhelpers.RandomPhone()
				createIdentityToRecoverEmailAndPhone(t, reg, address1, address2)

				var body string
				if identical {
					body = recoverHappyPath(t, client, RecoveryClientTypeBrowser, address2, address2)
				} else {
					body = recoverHappyPath(t, client, RecoveryClientTypeBrowser, address2, address1)
				}

				assert.Equal(t, text.NewRecoverySuccessful(time.Now().Add(time.Hour)).Text,
					gjson.Get(body, "ui.messages.0.text").String())

				res, err := client.Get(public.URL + session.RouteWhoami)
				require.NoError(t, err)
				body = string(x.MustReadAll(res.Body))
				require.NoError(t, res.Body.Close())
				assert.Equal(t, "code_recovery", gjson.Get(body, "authentication_methods.0.method").String(), "%s", body)
				assert.Equal(t, "aal1", gjson.Get(body, "authenticator_assurance_level").String(), "%s", body)
			})

			t.Run("type=spa", func(t *testing.T) {
				client := testhelpers.NewClientWithCookies(t)
				address1 := testhelpers.RandomEmail()
				address2 := testhelpers.RandomPhone()
				createIdentityToRecoverEmailAndPhone(t, reg, address1, address2)

				var body string
				if identical {
					body = recoverHappyPath(t, client, RecoveryClientTypeSPA, address2, address2)
				} else {
					body = recoverHappyPath(t, client, RecoveryClientTypeSPA, address2, address1)
				}

				assert.Equal(t, "passed_challenge", gjson.Get(body, "state").String())
				assert.Len(t, gjson.Get(body, "continue_with").Array(), 1)
				sfId := gjson.Get(body, "continue_with.#(action==show_settings_ui).flow.id").String()
				assert.NotEmpty(t, uuid.Must(uuid.FromString(sfId)))
			})

			t.Run("type=api", func(t *testing.T) {
				client := &http.Client{}
				address1 := testhelpers.RandomEmail()
				address2 := testhelpers.RandomPhone()
				createIdentityToRecoverEmailAndPhone(t, reg, address1, address2)

				var body string
				if identical {
					body = recoverHappyPath(t, client, RecoveryClientTypeAPI, address2, address2)
				} else {
					body = recoverHappyPath(t, client, RecoveryClientTypeAPI, address2, address1)
				}

				assert.Equal(t, "passed_challenge", gjson.Get(body, "state").String())
				assert.Len(t, gjson.Get(body, "continue_with").Array(), 2)
				assert.NotEmpty(t, gjson.Get(body, "continue_with.#(action==set_ory_session_token).ory_session_token").String())
				sfId := gjson.Get(body, "continue_with.#(action==show_settings_ui).flow.id").String()
				assert.NotEmpty(t, uuid.Must(uuid.FromString(sfId)))
			})
		}

		t.Run("description=should return browser to return url", func(t *testing.T) {
			returnTo := public.URL + "/return-to"
			conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{returnTo})
			for _, tc := range []struct {
				desc        string
				returnTo    string
				f           func(t *testing.T, client *http.Client, identity *identity.Identity) *kratos.RecoveryFlow
				expectedAAL string
			}{
				{
					desc:     "should use return_to from recovery flow",
					returnTo: returnTo,
					f: func(t *testing.T, client *http.Client, identity *identity.Identity) *kratos.RecoveryFlow {
						return testhelpers.InitializeRecoveryFlowViaBrowser(t, client, false, public, url.Values{"return_to": []string{returnTo}})
					},
				},
				{
					desc:     "should use return_to from config",
					returnTo: returnTo,
					f: func(t *testing.T, client *http.Client, identity *identity.Identity) *kratos.RecoveryFlow {
						conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryBrowserDefaultReturnTo, returnTo)
						t.Cleanup(func() {
							conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryBrowserDefaultReturnTo, "")
						})
						return testhelpers.InitializeRecoveryFlowViaBrowser(t, client, false, public, nil)
					},
				},
				{
					desc:     "no return to",
					returnTo: "",
					f: func(t *testing.T, client *http.Client, identity *identity.Identity) *kratos.RecoveryFlow {
						return testhelpers.InitializeRecoveryFlowViaBrowser(t, client, false, public, nil)
					},
				},
				{
					desc:     "should use return_to with an account that has 2fa enabled",
					returnTo: returnTo,
					f: func(t *testing.T, client *http.Client, id *identity.Identity) *kratos.RecoveryFlow {
						conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, config.HighestAvailableAAL)
						conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, config.HighestAvailableAAL)
						conf.MustSet(ctx, config.ViperKeyWebAuthnRPDisplayName, "Kratos")
						conf.MustSet(ctx, config.ViperKeyWebAuthnRPID, "ory.sh")

						t.Cleanup(func() {
							conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, identity.AuthenticatorAssuranceLevel1)
							conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, identity.AuthenticatorAssuranceLevel1)
						})
						testhelpers.StrategyEnable(t, conf, identity.CredentialsTypeWebAuthn.String(), true)

						id.SetCredentials(identity.CredentialsTypeWebAuthn, identity.Credentials{
							Type:        identity.CredentialsTypeWebAuthn,
							Config:      []byte(`{"credentials":[{"is_passwordless":false, "display_name":"test"}]}`),
							Identifiers: []string{testhelpers.RandomPhone()},
						})

						require.NoError(t, reg.IdentityManager().Update(ctx, id, identity.ManagerAllowWriteProtectedTraits))
						return testhelpers.InitializeRecoveryFlowViaBrowser(t, client, false, public, url.Values{"return_to": []string{returnTo}})
					},
					expectedAAL: "aal2",
				},
			} {
				t.Run(tc.desc, func(t *testing.T) {
					client := testhelpers.NewClientWithCookies(t)
					address1 := testhelpers.RandomEmail()
					address2 := testhelpers.RandomPhone()
					i := createIdentityToRecoverEmailAndPhone(t, reg, address1, address2)

					client.Transport = testhelpers.NewTransportWithLogger(http.DefaultTransport, t).RoundTripper
					f := tc.f(t, client, i)

					formPayload := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
					formPayload.Set("recovery_address", address2)

					body, res := testhelpers.RecoveryMakeRequest(t, false, f, client, formPayload.Encode())
					assert.EqualValues(t, http.StatusOK, res.StatusCode, "%s", body)

					// This screen might get skipped in the backend if there is only one possible address to choose.
					if gjson.Get(body, "ui.nodes.#(attributes.name==recovery_select_address)").Exists() {
						checkRecoveryScreenAskForRecoverySelectAddress(t, body)
						body = submitRecoveryFormSubsequent(t, client, body, RecoveryClientTypeBrowser, func(v url.Values) {
							v.Set("recovery_select_address", code.AddressToHashBase64(address1))
							v.Set("recovery_address", address2)
						}, http.StatusOK)
					}

					checkRecoveryScreenAskForRecoveryConfirmAddress(t, body)
					body = submitRecoveryFormSubsequent(t, client, body, RecoveryClientTypeBrowser, func(v url.Values) {
						v.Set("recovery_confirm_address", address2)
					}, http.StatusOK)

					checkRecoveryScreenAskForCode(t, address2, body)

					body = extractCodeFromCourierAndSubmit(t, client, RecoveryClientTypeBrowser, address2, body, http.StatusOK)

					require.Equal(t, text.NewRecoverySuccessful(time.Now().Add(time.Hour)).Text,
						gjson.Get(body, "ui.messages.0.text").String())

					settingsId := gjson.Get(body, "id").String()

					sf, err := reg.SettingsFlowPersister().GetSettingsFlow(ctx, uuid.Must(uuid.FromString(settingsId)))
					require.NoError(t, err)

					u, err := url.Parse(public.URL)
					require.NoError(t, err)
					require.Len(t, client.Jar.Cookies(u), 2)
					found := false
					for _, cookie := range client.Jar.Cookies(u) {
						if cookie.Name == "ory_kratos_session" {
							found = true
						}
					}
					require.True(t, found)

					require.Equal(t, tc.returnTo, sf.ReturnTo)
					res, err = client.Get(public.URL + session.RouteWhoami)
					require.NoError(t, err)
					body = string(x.MustReadAll(res.Body))
					require.NoError(t, res.Body.Close())

					if tc.expectedAAL == "aal2" {
						require.Equal(t, http.StatusForbidden, res.StatusCode)
						require.Equalf(t, session.NewErrAALNotSatisfied("").Reason(), gjson.Get(body, "error.reason").String(), "%s", body)
						require.Equalf(t, "session_aal2_required", gjson.Get(body, "error.id").String(), "%s", body)
					} else {
						assert.Equal(t, "code_recovery", gjson.Get(body, "authentication_methods.0.method").String(), "%s", body)
						assert.Equal(t, "aal1", gjson.Get(body, "authenticator_assurance_level").String(), "%s", body)
					}
				})
			}
		})
	})

	t.Run("description=should set all the correct recovery payloads after submission", func(t *testing.T) {
		fakes := []string{"+491705550166", "+491705550167", "+491705550168"}
		fakeIdx := 0

		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				address2 := fakes[fakeIdx]
				fakeIdx += 1

				address1 := "test_mrecovery_addresses-" + testCase.ClientType.String() + "@ory.sh"
				createIdentityToRecoverEmailAndPhone(t, reg, address1, address2)

				body := submitRecoveryFormInitial(t, testCase.GetClient(t), testCase.ClientType, func(u url.Values) { u.Set("recovery_address", address2) }, http.StatusOK)
				testhelpers.SnapshotTExcept(t, json.RawMessage(gjson.Get(body, "ui.nodes").String()), []string{"0.attributes.value"})
			})
		}
	})

	t.Run("description=should set all the correct recovery payloads", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				c := testCase.GetClient(t)
				rs := testhelpers.GetRecoveryFlowForType(t, c, public, testCase.FlowType)

				testhelpers.SnapshotTExcept(t, rs.Ui.Nodes, []string{"0.attributes.value"})
				assert.EqualValues(t, public.URL+recovery.RouteSubmitFlow+"?flow="+rs.Id, rs.Ui.Action)
				assert.Empty(t, rs.Ui.Messages)
			})
		}
	})

	t.Run("description=should require an address to be sent", func(t *testing.T) {
		for _, flowType := range flowTypes {
			t.Run("type="+flowType.String(), func(t *testing.T) {
				code := testhelpers.ExpectStatusCode(flowType == RecoveryClientTypeAPI || flowType == RecoveryClientTypeSPA, http.StatusBadRequest, http.StatusOK)
				body := submitRecoveryFormInitial(t, nil, flowType, func(v url.Values) {
					v.Del("recovery_address")
				}, code)
				assert.EqualValues(t, node.CodeGroup, gjson.Get(body, "active").String(), "%s", body)
				assert.EqualValues(t, "Property recovery_address is missing.",
					gjson.Get(body, "ui.nodes.#(attributes.name==recovery_address).messages.0.text").String(),
					"%s", body)
			})
		}
	})

	t.Run("description=should pretend the address exists when it does not", func(t *testing.T) {
		for _, flowType := range flowTypes {
			t.Run("type="+flowType.String(), func(t *testing.T) {
				for _, address := range []string{"\\", "asdf", "...", "aiacobelli.sec@gmail.com,alejandro.iacobelli@mercadolibre.com"} {
					body := submitRecoveryFormInitial(t, nil, flowType, func(v url.Values) {
						v.Set("recovery_address", address)
					}, http.StatusOK)

					activeMethod := gjson.Get(body, "active").String()
					assert.EqualValues(t, node.CodeGroup, activeMethod, "expected method to be %s got %s", node.CodeGroup, activeMethod)

					expectedMessage := text.NewRecoveryCodeRecoverySelectAddressSent(code.MaskAddress(address)).Text
					actualMessage := gjson.Get(body, "ui.messages.0.text").String()
					assert.EqualValues(t, expectedMessage, actualMessage, "%s", body)
				}
			})
		}
	})

	t.Run("description=should try to submit the form while authenticated", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				isSPA := testCase.ClientType == "spa"
				isAPI := testCase.ClientType == "api"
				client := testCase.GetClient(t)

				var f *kratos.RecoveryFlow
				if isAPI {
					f = testhelpers.InitializeRecoveryFlowViaAPI(t, client, public)
				} else {
					f = testhelpers.InitializeRecoveryFlowViaBrowser(t, client, isSPA, public, nil)
				}
				req := httptest.NewRequest("GET", "/sessions/whoami", nil).WithContext(contextx.WithConfigValue(ctx, config.ViperKeySessionLifespan, time.Hour))

				session, err := testhelpers.NewActiveSession(
					req,
					reg,
					&identity.Identity{ID: x.NewUUID(), State: identity.StateActive, NID: x.NewUUID()},
					time.Now(),
					identity.CredentialsTypePassword,
					identity.AuthenticatorAssuranceLevel1,
				)

				require.NoError(t, err)

				// Add the authentication to the request
				client.Transport = testhelpers.NewTransportWithLogger(testhelpers.NewAuthorizedTransport(t, ctx, reg, session), t).RoundTripper

				v := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
				v.Set("recovery_address", "some-address@example.org")
				v.Set("method", "code")

				body, res := testhelpers.RecoveryMakeRequest(t, isAPI || isSPA, f, client, testhelpers.EncodeFormAsJSON(t, isAPI || isSPA, v))

				if isAPI || isSPA {
					assert.EqualValues(t, http.StatusBadRequest, res.StatusCode, "%s", body)
					assert.Contains(t, res.Request.URL.String(), recovery.RouteSubmitFlow, "%+v\n\t%s", res.Request, body)
					assertx.EqualAsJSONExcept(t, recovery.ErrAlreadyLoggedIn, json.RawMessage(gjson.Get(body, "error").Raw), nil)
				} else {
					assert.EqualValues(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.Contains(t, res.Request.URL.String(), conf.SelfServiceBrowserDefaultReturnTo(ctx).String(), "%+v\n\t%s", res.Request, body)
				}
			})
		}
	})

	t.Run("description=should not be able to recover account that does not exist", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryNotifyUnknownRecipients, true)

		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryNotifyUnknownRecipients, false)
		})

		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				address := x.NewUUID().String() + "@ory.sh"
				c := testCase.GetClient(t)
				withValues := func(v url.Values) {
					v.Set("recovery_address", address)
				}
				body := submitRecoveryFormInitial(t, c, testCase.ClientType, withValues, http.StatusOK)
				assert.EqualValues(t, node.CodeGroup, gjson.Get(body, "active").String(), "%s", body)
				assert.Empty(t, gjson.Get(body, "ui.nodes.#(attributes.name==code).attributes.value").String(), "%s", body)
				assertx.EqualAsJSON(t, text.NewRecoveryCodeRecoverySelectAddressSent(code.MaskAddress(address)), json.RawMessage(gjson.Get(body, "ui.messages.0").Raw))

				message := testhelpers.CourierExpectMessage(ctx, t, reg, address, "Account access attempted")
				assert.Contains(t, message.Body, "If this was you, check if you signed up using a different address.")
			})
		}
	})

	t.Run("description=should not be able to recover an inactive account", func(t *testing.T) {
		fakes := []string{"+491705550163", "+491705550164", "+491705550165"}
		fakeIdx := 0

		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				address2 := fakes[fakeIdx]
				fakeIdx += 1

				address1 := testhelpers.RandomEmail()
				i := createIdentityToRecoverEmailAndPhone(t, reg, address1, address2)
				values := func(v url.Values) {
					v.Set("recovery_address", address2)
				}
				cl := testhelpers.NewClientWithCookies(t)

				body := submitRecoveryFormInitial(t, cl, testCase.ClientType, values, http.StatusOK)

				// This screen might get skipped in the backend if there is only one possible address to choose.
				if gjson.Get(body, "ui.nodes.#(attributes.name==recovery_select_address)").Exists() {
					checkRecoveryScreenAskForRecoverySelectAddress(t, body)
					body = submitRecoveryFormSubsequent(t, cl, body, testCase.ClientType, func(v url.Values) {
						v.Set("recovery_select_address", code.AddressToHashBase64(address1))
						v.Set("recovery_address", address2)
					}, http.StatusOK)
				}

				checkRecoveryScreenAskForRecoveryConfirmAddress(t, body)
				body = submitRecoveryFormSubsequent(t, cl, body, testCase.ClientType, func(v url.Values) {
					v.Set("recovery_confirm_address", address2)
				}, http.StatusOK)

				checkRecoveryScreenAskForCode(t, address2, body)

				// Deactivate the identity
				require.NoError(t, reg.Persister().GetConnection(context.Background()).RawQuery("UPDATE identities SET state=? WHERE id = ?", identity.StateInactive, i.ID).Exec())

				code := testhelpers.ExpectStatusCode(testCase.ClientType == RecoveryClientTypeAPI || testCase.ClientType == RecoveryClientTypeSPA, http.StatusUnauthorized, http.StatusOK)
				body = extractCodeFromCourierAndSubmit(t, cl, testCase.ClientType, address2, body, code)

				switch testCase.ClientType {
				case RecoveryClientTypeAPI:
					fallthrough
				case RecoveryClientTypeSPA:
					assertx.EqualAsJSON(t, session.ErrIdentityDisabled.WithDetail("identity_id", i.ID), json.RawMessage(gjson.Get(body, "error").Raw), "%s", body)
				default:
					assertx.EqualAsJSON(t, session.ErrIdentityDisabled.WithDetail("identity_id", i.ID), json.RawMessage(body), "%s", body)
				}
			})
		}
	})

	t.Run("description=should see error if invalid recovery address is submitted", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				address2 := testhelpers.RandomPhone()
				address1 := testhelpers.RandomEmail()
				_ = createIdentityToRecoverEmailAndPhone(t, reg, address1, address2)
				values := func(v url.Values) {
					v.Set("recovery_address", address2)
				}
				cl := testhelpers.NewClientWithCookies(t)

				body := submitRecoveryFormInitial(t, cl, testCase.ClientType, values, http.StatusOK)

				checkRecoveryScreenAskForRecoverySelectAddress(t, body)
				sc := http.StatusOK
				if testCase.ClientType != RecoveryClientTypeBrowser {
					sc = http.StatusBadRequest
				}
				body = submitRecoveryFormSubsequent(t, cl, body, testCase.ClientType, func(v url.Values) {
					v.Set("recovery_select_address", code.AddressToHashBase64(address1))
					v.Set("recovery_address", "not-the-correct@email.com")
				}, sc)

				require.Equal(t, 1, len(gjson.Get(body, "ui.messages").Array()), "%s", body)
				assert.Equal(t, "4000001", gjson.Get(body, "ui.messages.0.id").String(), "%s", body)
				assert.Equal(t, "The selected recovery address is not valid.", gjson.Get(body, "ui.messages.0.text").String(), "%s", body)
			})
		}
	})

	t.Run("description=should recover and invalidate all other sessions if hook is set", func(t *testing.T) {
		conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal), []config.SelfServiceHook{{Name: "revoke_active_sessions"}})
		t.Cleanup(func() {
			conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()), nil)
		})

		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				address1 := testhelpers.RandomEmail()
				address2 := testhelpers.RandomPhone()
				id := createIdentityToRecoverEmailAndPhone(t, reg, address1, address2)

				otherSession, err := testhelpers.NewActiveSession(httptest.NewRequest("GET", "/sessions/whoami", nil), reg, id, time.Now(), identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
				require.NoError(t, err)
				require.NoError(t, reg.SessionPersister().UpsertSession(ctx, otherSession))

				refetchedOtherSession, err := reg.SessionPersister().GetSession(ctx, otherSession.ID, session.ExpandNothing)
				require.NoError(t, err)
				assert.True(t, refetchedOtherSession.IsActive())

				cl := testCase.GetClient(t)
				body := recoverHappyPath(t, cl, testCase.ClientType, address2, address1)

				expectRedirectToSettings(t, cl, testCase.ClientType, body)

				refetchedOtherSession, err = reg.SessionPersister().GetSession(ctx, otherSession.ID, session.ExpandNothing)
				require.NoError(t, err)
				assert.False(t, refetchedOtherSession.IsActive())
			})
		}
	})

	t.Run("description=should not be able to use an invalid code more than 5 times", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				address1 := testhelpers.RandomEmail()
				address2 := testhelpers.RandomPhone()
				createIdentityToRecoverEmailAndPhone(t, reg, address1, address2)
				c := testCase.GetClient(t)

				body := submitRecoveryFormInitial(t, c, testCase.ClientType, func(v url.Values) {
					v.Set("recovery_address", address2)
				}, http.StatusOK)

				// This screen might get skipped in the backend if there is only one possible address to choose.
				if gjson.Get(body, "ui.nodes.#(attributes.name==recovery_select_address)").Exists() {
					checkRecoveryScreenAskForRecoverySelectAddress(t, body)
					body = submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) {
						v.Set("recovery_select_address", code.AddressToHashBase64(address1))
						v.Set("recovery_address", address2)
					}, http.StatusOK)
				}

				checkRecoveryScreenAskForRecoveryConfirmAddress(t, body)
				body = submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) {
					v.Set("recovery_confirm_address", address1)
				}, http.StatusOK)

				checkRecoveryScreenAskForCode(t, address1, body)

				initialFlowId := gjson.Get(body, "id")

				for range 5 {
					body := submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) { v.Set("code", "12312312") }, http.StatusOK)

					testhelpers.AssertMessage(t, []byte(body), "The recovery code is invalid or has already been used. Please try again.")
				}

				switch testCase.ClientType {
				case RecoveryClientTypeBrowser:
					// submit an invalid code for the 6th time
					body := submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) { v.Set("code", "12312312") }, http.StatusOK)

					require.Len(t, gjson.Get(body, "ui.messages").Array(), 1, "%s", body)
					assert.Equal(t, "The request was submitted too often. Please request another code.", gjson.Get(body, "ui.messages.0.text").String())

					// check that a new flow has been created
					assert.NotEqual(t, gjson.Get(body, "id"), initialFlowId)

					assert.True(t, gjson.Get(body, "ui.nodes.#(attributes.name==recovery_address)").Exists())
				case RecoveryClientTypeSPA:
					fallthrough
				case RecoveryClientTypeAPI:
					// submit an invalid code for the 6th time
					body := submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) { v.Set("code", "12312312") }, http.StatusBadRequest)

					assert.Equal(t, "Bad Request", gjson.Get(body, "error.status").String(), "%s", body)
					assert.Equal(t, "The request was submitted too often. Please request another code.", gjson.Get(body, "error.reason").String(), "%s", body)
					continueWith := gjson.Get(body, "error.details.continue_with").Array()
					assert.Len(t, continueWith, 1, "%s", body)
					assert.Equal(t, "show_recovery_ui", continueWith[0].Get("action").String(), "%s", body)
					flowId := continueWith[0].Get("flow.id").String()
					assert.NotEmpty(t, flowId, "%s", body)
					require.NotEqual(t, flowId, initialFlowId, "%s", body)

					flow, err := reg.Persister().GetRecoveryFlow(ctx, uuid.Must(uuid.FromString(flowId)))
					require.NoError(t, err)
					assert.Len(t, flow.UI.Messages, 1, "%+v", flow)
					assert.Equal(t, "The request was submitted too often. Please request another code.", flow.UI.Messages[0].Text)
				}
			})
		}
	})

	t.Run("description=should be able to recover after using invalid code", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				c := testCase.GetClient(t)
				address1 := testhelpers.RandomEmail()
				address2 := testhelpers.RandomPhone()
				createIdentityToRecoverEmailAndPhone(t, reg, address1, address2)

				body := submitRecoveryFormInitial(t, c, testCase.ClientType, func(v url.Values) {
					v.Set("recovery_address", address2)
				}, http.StatusOK)

				// This screen might get skipped in the backend if there is only one possible address to choose.
				if gjson.Get(body, "ui.nodes.#(attributes.name==recovery_select_address)").Exists() {
					checkRecoveryScreenAskForRecoverySelectAddress(t, body)
					body = submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) {
						v.Set("recovery_select_address", code.AddressToHashBase64(address1))
						v.Set("recovery_address", address2)
					}, http.StatusOK)
				}

				checkRecoveryScreenAskForRecoveryConfirmAddress(t, body)
				body = submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) {
					v.Set("recovery_confirm_address", address1)
				}, http.StatusOK)

				checkRecoveryScreenAskForCode(t, address1, body)

				// Submit invalid code.
				body = submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) { v.Set("code", "12312312") }, http.StatusOK)
				flowId := gjson.Get(body, "id").String()
				require.NotEmpty(t, flowId)

				rs, res, err := testhelpers.
					NewSDKCustomClient(public, c).
					FrontendAPI.GetRecoveryFlow(context.Background()).
					Id(flowId).
					Execute()

				require.NoError(t, err)
				getBody := ioutilx.MustReadAll(res.Body)
				require.NotEmpty(t, getBody)

				require.Len(t, rs.Ui.Messages, 1)
				assert.Equal(t, "The recovery code is invalid or has already been used. Please try again.", rs.Ui.Messages[0].Text)

				// Now submit the correct code
				body = extractCodeFromCourierAndSubmit(t, c, testCase.ClientType, address1, body, http.StatusOK)

				switch testCase.ClientType {
				case RecoveryClientTypeBrowser:
					assert.Len(t, gjson.Get(body, "ui.messages").Array(), 1)
					assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "You successfully recovered your account.")
				case RecoveryClientTypeSPA:
					require.Len(t, c.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 2)
					cookies := spew.Sdump(c.Jar.Cookies(urlx.ParseOrPanic(public.URL)))
					assert.Contains(t, cookies, "ory_kratos_session")

					require.NotEmpty(t, gjson.Get(body, "continue_with.#(action==show_settings_ui).flow").String(), "%s", body)
				case RecoveryClientTypeAPI:
					require.NotEmpty(t, gjson.Get(body, "continue_with.#(action==show_settings_ui).flow").String(), "%s", body)
					require.NotEmpty(t, gjson.Get(body, "continue_with.#(action==set_ory_session_token).ory_session_token").String(), "%s", body)
				}
			})
		}
	})

	t.Run("description=should not break ui if empty code is submitted", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				address1 := testhelpers.RandomEmail()
				address2 := testhelpers.RandomPhone()
				createIdentityToRecoverEmailAndPhone(t, reg, address1, address2)

				c := testCase.GetClient(t)
				body := submitRecoveryFormInitial(t, c, testCase.ClientType, func(v url.Values) {
					v.Set("recovery_address", address2)
				}, http.StatusOK)

				// This screen might get skipped in the backend if there is only one possible address to choose.
				if gjson.Get(body, "ui.nodes.#(attributes.name==recovery_select_address)").Exists() {
					checkRecoveryScreenAskForRecoverySelectAddress(t, body)
					body = submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) {
						v.Set("recovery_select_address", code.AddressToHashBase64(address1))
						v.Set("recovery_address", address2)
					}, http.StatusOK)
				}

				checkRecoveryScreenAskForRecoveryConfirmAddress(t, body)
				body = submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) {
					v.Set("recovery_confirm_address", address1)
				}, http.StatusOK)

				checkRecoveryScreenAskForCode(t, address1, body)

				body = submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) {
					v.Set("code", "")
					v.Set("recovery_confirm_address", address1)
				}, http.StatusOK)

				// Not an error, just handle it as a code resend.
				testhelpers.AssertMessage(t, []byte(body), text.NewRecoveryCodeRecoverySelectAddressSent(code.MaskAddress(address1)).Text)
			})
		}
	})

	t.Run("description=should be able to resend the recovery code", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				address1 := testhelpers.RandomEmail()
				address2 := testhelpers.RandomPhone()
				createIdentityToRecoverEmailAndPhone(t, reg, address1, address2)

				c := testCase.GetClient(t)

				body := submitRecoveryFormInitial(t, c, testCase.ClientType, func(v url.Values) {
					v.Set("recovery_address", address2)
				}, http.StatusOK)

				// This screen might get skipped in the backend if there is only one possible address to choose.
				if gjson.Get(body, "ui.nodes.#(attributes.name==recovery_select_address)").Exists() {
					checkRecoveryScreenAskForRecoverySelectAddress(t, body)
					body = submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) {
						v.Set("recovery_select_address", code.AddressToHashBase64(address1))
						v.Set("recovery_address", address2)
					}, http.StatusOK)
				}

				checkRecoveryScreenAskForRecoveryConfirmAddress(t, body)
				body = submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) {
					v.Set("recovery_confirm_address", address1)
				}, http.StatusOK)

				checkRecoveryScreenAskForCode(t, address1, body)

				body = submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) {
					v.Set("code", "")
					v.Set("recovery_confirm_address", address1) // Trigger resend.
				}, http.StatusOK)

				action := gjson.Get(body, "ui.action").String()
				require.NotEmpty(t, action)
				assert.Equal(t, address1, gjson.Get(body, "ui.nodes.#(attributes.name==recovery_confirm_address).attributes.value").String())
				assert.True(t, gjson.Get(body, "ui.nodes.#(attributes.name==code)").Exists())

				body = extractCodeFromCourierAndSubmit(t, c, testCase.ClientType, address1, body, http.StatusOK)
				expectRedirectToSettings(t, c, testCase.ClientType, body)
			})
		}
	})

	t.Run("description=should not be able to use first code after re-sending address", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				address1 := testhelpers.RandomEmail()
				address2 := testhelpers.RandomPhone()
				createIdentityToRecoverEmailAndPhone(t, reg, address1, address2)

				c := testCase.GetClient(t)

				body := submitRecoveryFormInitial(t, c, testCase.ClientType, func(v url.Values) {
					v.Set("recovery_address", address2)
				}, http.StatusOK)

				// This screen might get skipped in the backend if there is only one possible address to choose.
				if gjson.Get(body, "ui.nodes.#(attributes.name==recovery_select_address)").Exists() {
					checkRecoveryScreenAskForRecoverySelectAddress(t, body)
					body = submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) {
						v.Set("recovery_select_address", code.AddressToHashBase64(address1))
						v.Set("recovery_address", address2)
					}, http.StatusOK)
				}

				checkRecoveryScreenAskForRecoveryConfirmAddress(t, body)
				body = submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) {
					v.Set("recovery_confirm_address", address1)
				}, http.StatusOK)

				checkRecoveryScreenAskForCode(t, address1, body)

				message1 := testhelpers.CourierExpectMessage(ctx, t, reg, address1, "")
				assert.Contains(t, message1.Body, "Recover access to your account")
				recoveryCode1 := testhelpers.CourierExpectCodeInMessage(t, message1, 1)
				assert.NotEmpty(t, recoveryCode1)

				body = submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) {
					v.Set("code", "")
					v.Set("recovery_confirm_address", address1) // Trigger resend.
				}, http.StatusOK)

				action := gjson.Get(body, "ui.action").String()
				require.NotEmpty(t, action)
				assert.Equal(t, address1, gjson.Get(body, "ui.nodes.#(attributes.name==recovery_confirm_address).attributes.value").String())
				assert.True(t, gjson.Get(body, "ui.nodes.#(attributes.name==code)").Exists())

				// Try to submit the old (expired) code.
				body = submitRecoveryFormSubsequent(t, c, body, testCase.ClientType, func(v url.Values) {
					v.Set("code", recoveryCode1)
				}, http.StatusOK)
				testhelpers.AssertMessage(t, []byte(body), "The recovery code is invalid or has already been used. Please try again.")

				// Send the right code.
				body = extractCodeFromCourierAndSubmit(t, c, testCase.ClientType, address1, body, http.StatusOK)
				expectRedirectToSettings(t, c, testCase.ClientType, body)
			})
		}
	})

	t.Run("description=should recover if post recovery hook is successful", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal), []config.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})
				t.Cleanup(func() {
					conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal), nil)
				})

				address1 := testhelpers.RandomEmail()
				address2 := testhelpers.RandomPhone()
				createIdentityToRecoverEmailAndPhone(t, reg, address1, address2)

				cl := testCase.GetClient(t)

				body := recoverHappyPath(t, cl, testCase.ClientType, address2, address1)
				expectRedirectToSettings(t, cl, testCase.ClientType, body)
			})
		}
	})

	t.Run("description=should not be able to recover if post recovery hook fails", func(t *testing.T) {
		for _, testCase := range flowTypeCases {
			t.Run("type="+testCase.ClientType.String(), func(t *testing.T) {
				conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal), []config.SelfServiceHook{{Name: "err", Config: []byte(`{"ExecutePostRecoveryHook": "err"}`)}})
				t.Cleanup(func() {
					conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal), nil)
				})

				address1 := testhelpers.RandomEmail()
				address2 := testhelpers.RandomPhone()
				createIdentityToRecoverEmailAndPhone(t, reg, address1, address2)

				cl := testhelpers.NewClientWithCookies(t)

				body := submitRecoveryFormInitial(t, cl, testCase.ClientType, func(v url.Values) {
					v.Set("recovery_address", address2)
				}, http.StatusOK)

				// This screen might get skipped in the backend if there is only one possible address to choose.
				if gjson.Get(body, "ui.nodes.#(attributes.name==recovery_select_address)").Exists() {
					checkRecoveryScreenAskForRecoverySelectAddress(t, body)
					body = submitRecoveryFormSubsequent(t, cl, body, testCase.ClientType, func(v url.Values) {
						v.Set("recovery_select_address", code.AddressToHashBase64(address1))
						v.Set("recovery_address", address2)
					}, http.StatusOK)
				}

				checkRecoveryScreenAskForRecoveryConfirmAddress(t, body)
				body = submitRecoveryFormSubsequent(t, cl, body, testCase.ClientType, func(v url.Values) {
					v.Set("recovery_confirm_address", address1)
				}, http.StatusOK)

				checkRecoveryScreenAskForCode(t, address1, body)

				assert.Equal(t, address1, gjson.Get(body, "ui.nodes.#(attributes.name==recovery_confirm_address).attributes.value").String())

				cl.CheckRedirect = func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				}

				initialFlowId := gjson.Get(body, "id")
				switch testCase.ClientType {
				case RecoveryClientTypeBrowser:
					body = extractCodeFromCourierAndSubmit(t, cl, testCase.ClientType, address1, body, http.StatusSeeOther)
					assert.NotEqual(t, gjson.Get(body, "id"), initialFlowId)

					require.Len(t, cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 1)
					cookies := spew.Sdump(cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)))
					assert.NotContains(t, cookies, "ory_kratos_session")
				case RecoveryClientTypeSPA:
					body = extractCodeFromCourierAndSubmit(t, cl, testCase.ClientType, address1, body, http.StatusBadRequest)
					assert.NotEqual(t, gjson.Get(body, "id"), initialFlowId)

					require.Len(t, cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 1)
					cookies := spew.Sdump(cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)))
					assert.NotContains(t, cookies, "ory_kratos_session")
				case RecoveryClientTypeAPI:
					body = extractCodeFromCourierAndSubmit(t, cl, testCase.ClientType, address1, body, http.StatusBadRequest)
					assert.NotEqual(t, gjson.Get(body, "id"), initialFlowId)
					require.Equal(t, "err", gjson.Get(body, "error.message").String(), "%s", body)
				}
			})
		}
	})

	t.Run("choose different address - screens 2->3->2->4", func(t *testing.T) {
		address1 := testhelpers.RandomEmail()
		address2 := testhelpers.RandomPhone()
		createIdentityToRecoverEmailAndPhone(t, reg, address1, address2)

		client := testhelpers.NewClientWithCookies(t)

		body := submitRecoveryFormInitial(t, client, RecoveryClientTypeBrowser, func(v url.Values) {
			v.Set("recovery_address", address1)
		}, http.StatusOK)
		checkRecoveryScreenAskForRecoverySelectAddress(t, body)

		body = submitRecoveryFormSubsequent(t, client, body, RecoveryClientTypeBrowser, func(v url.Values) {
			v.Set("recovery_address", address1)
			v.Set("recovery_select_address", code.AddressToHashBase64(address2))
		}, http.StatusOK)
		checkRecoveryScreenAskForRecoveryConfirmAddress(t, body)

		body = submitRecoveryFormSubsequent(t, client, body, RecoveryClientTypeBrowser, func(v url.Values) {
			v.Set("recovery_address", address1)
			v.Set("screen", "previous")
		}, http.StatusOK)
		checkRecoveryScreenAskForRecoverySelectAddress(t, body)

		// Choose another address
		body = submitRecoveryFormSubsequent(t, client, body, RecoveryClientTypeBrowser, func(v url.Values) {
			v.Set("recovery_address", address1)
			v.Set("recovery_select_address", code.AddressToHashBase64(address1))
		}, http.StatusOK)
		checkRecoveryScreenAskForCode(t, address1, body)

		body = extractCodeFromCourierAndSubmit(t, client, RecoveryClientTypeBrowser, address1, body, http.StatusOK)
		assert.Equal(t, text.NewRecoverySuccessful(time.Now().Add(time.Hour)).Text,
			gjson.Get(body, "ui.messages.0.text").String())
	})

	t.Run("choose different address - screens 2->4->2->3->4", func(t *testing.T) {
		address1 := testhelpers.RandomEmail()
		address2 := testhelpers.RandomPhone()
		createIdentityToRecoverEmailAndPhone(t, reg, address1, address2)

		client := testhelpers.NewClientWithCookies(t)

		body := submitRecoveryFormInitial(t, client, RecoveryClientTypeBrowser, func(v url.Values) {
			v.Set("recovery_address", address1)
		}, http.StatusOK)
		checkRecoveryScreenAskForRecoverySelectAddress(t, body)

		body = submitRecoveryFormSubsequent(t, client, body, RecoveryClientTypeBrowser, func(v url.Values) {
			v.Set("recovery_address", address1)
			v.Set("recovery_select_address", code.AddressToHashBase64(address1))
		}, http.StatusOK)
		checkRecoveryScreenAskForCode(t, address1, body)

		// Choose another address
		body = submitRecoveryFormSubsequent(t, client, body, RecoveryClientTypeBrowser, func(v url.Values) {
			v.Set("recovery_address", address1)
			v.Set("screen", "previous")
		}, http.StatusOK)
		checkRecoveryScreenAskForRecoverySelectAddress(t, body)

		body = submitRecoveryFormSubsequent(t, client, body, RecoveryClientTypeBrowser, func(v url.Values) {
			v.Set("recovery_address", address1)
			v.Set("recovery_select_address", code.AddressToHashBase64(address2))
		}, http.StatusOK)
		checkRecoveryScreenAskForRecoveryConfirmAddress(t, body)

		body = submitRecoveryFormSubsequent(t, client, body, RecoveryClientTypeBrowser, func(v url.Values) {
			v.Set("recovery_address", address1)
			v.Set("recovery_confirm_address", address2)
		}, http.StatusOK)
		checkRecoveryScreenAskForCode(t, address2, body)

		body = extractCodeFromCourierAndSubmit(t, client, RecoveryClientTypeBrowser, address2, body, http.StatusOK)
		assert.Equal(t, text.NewRecoverySuccessful(time.Now().Add(time.Hour)).Text,
			gjson.Get(body, "ui.messages.0.text").String())
	})
}

func TestDisabledStrategy(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	initViper(t, ctx, conf)
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(recovery.RecoveryStrategyLink)+".enabled", false)
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(recovery.RecoveryStrategyCode)+".enabled", false)

	publicTS, adminTS := testhelpers.NewKratosServer(t, reg)
	adminSDK := testhelpers.NewSDKClient(adminTS)
	_ = testhelpers.NewErrorTestServer(t, reg)

	t.Run("role=admin", func(t *testing.T) {
		t.Run("description=can not create recovery link when link method is disabled", func(t *testing.T) {
			id := identity.Identity{Traits: identity.Traits(`{"email":"recovery-endpoint-disabled@ory.sh"}`)}

			require.NoError(t, reg.IdentityManager().Create(context.Background(),
				&id, identity.ManagerAllowWriteProtectedTraits))

			rl, _, err := adminSDK.IdentityAPI.
				CreateRecoveryLinkForIdentity(context.Background()).
				CreateRecoveryLinkForIdentityBody(kratos.CreateRecoveryLinkForIdentityBody{IdentityId: id.ID.String()}).
				Execute()
			assert.Nil(t, rl)
			require.IsType(t, new(kratos.GenericOpenAPIError), err, "%s", err)

			br, _ := err.(*kratos.GenericOpenAPIError)
			assert.Contains(t, string(br.Body()), "This endpoint was disabled by system administrator", "%s", br.Body())
		})
	})

	t.Run("role=public", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)

		t.Run("description=can not recover an account by post request when code method is disabled", func(t *testing.T) {
			f := testhelpers.PersistNewRecoveryFlow(t, code.NewStrategy(reg), conf, reg)
			u := publicTS.URL + recovery.RouteSubmitFlow + "?flow=" + f.ID.String()

			res, err := c.PostForm(u, url.Values{
				"email":      {"email@ory.sh"},
				"method":     {"code"},
				"csrf_token": {f.CSRFToken},
			})
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)

			b := ioutilx.MustReadAll(res.Body)
			assert.Contains(t, string(b), "This endpoint was disabled by system administrator")
		})
	})
}
