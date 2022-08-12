package code_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/session"

	"github.com/ory/kratos/ui/node"

	kratos "github.com/ory/kratos-client-go"

	"github.com/ory/kratos/corpx"

	"github.com/ory/x/ioutilx"
	"github.com/ory/x/urlx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/sqlxx"

	"github.com/ory/x/assertx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

func init() {
	corpx.RegisterFakes()
}

// func TestAdminStrategy(t *testing.T) {
// 	conf, reg := internal.NewFastRegistryWithMocks(t)
// 	initViper(t, conf)

// 	_ = testhelpers.NewRecoveryUIFlowEchoServer(t, reg)
// 	_ = testhelpers.NewSettingsUIFlowEchoServer(t, reg)
// 	_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
// 	_ = testhelpers.NewErrorTestServer(t, reg)

// 	publicTS, adminTS := testhelpers.NewKratosServer(t, reg)
// 	adminSDK := testhelpers.NewSDKClient(adminTS)

// 	createCode := func(id string, expiresIn *string) (*kratos.SelfServiceRecoveryCode, *http.Response, error) {
// 		return adminSDK.V0alpha2Api.
// 			AdminCreateSelfServiceRecoveryCode(context.Background()).
// 			AdminCreateSelfServiceRecoveryCodeBody(
// 				kratos.AdminCreateSelfServiceRecoveryCodeBody{
// 					IdentityId: id,
// 					ExpiresIn:  expiresIn,
// 				}).
// 			Execute()
// 	}

// 	t.Run("no panic on empty body #1384", func(t *testing.T) {
// 		ctx := context.Background()
// 		s, err := reg.RecoveryStrategies(ctx).Strategy("code")
// 		require.NoError(t, err)
// 		w := httptest.NewRecorder()
// 		r := &http.Request{URL: new(url.URL)}
// 		f, err := recovery.NewFlow(reg.Config(ctx), time.Minute, "", r, reg.RecoveryStrategies(ctx), flow.TypeBrowser)
// 		require.NoError(t, err)
// 		require.NotPanics(t, func() {
// 			require.Error(t, s.(*code.Strategy).HandleRecoveryError(w, r, f, nil, errors.New("test")))
// 		})
// 	})

// 	t.Run("description=should not be able to recover an account that does not exist", func(t *testing.T) {
// 		_, _, err := createCode(x.NewUUID().String(), nil)

// 		require.IsType(t, err, new(kratos.GenericOpenAPIError), "%T", err)
// 		assert.EqualError(t, err.(*kratos.GenericOpenAPIError), "400 Bad Request")
// 	})

// 	t.Run("description=should create code without email", func(t *testing.T) {
// 		id := identity.Identity{Traits: identity.Traits(`{}`)}

// 		require.NoError(t, reg.IdentityManager().Create(context.Background(),
// 			&id, identity.ManagerAllowWriteProtectedTraits))

// 		code, _, err := createCode(id.ID.String(), pointerx.String("100ms"))
// 		require.NoError(t, err)

// 		require.NotEmpty(t, code.RecoveryLink)
// 		require.True(t, code.ExpiresAt.Before(time.Now().Add(conf.SelfServiceFlowRecoveryRequestLifespan())))

// 		res, err := publicTS.Client().Get(code.RecoveryLink)
// 		body := ioutilx.MustReadAll(res.Body)

// 		require.Len(t, gjson.GetBytes(body, "ui.messages").Array(), 1)
// 		require.Contains(t, gjson.GetBytes(body, "ui.messages.0.text").Str, "The recovery flow expired")

// 		require.True(t, false)
// 	})

// 	t.Run("description=should not be able to recover with expired code", func(t *testing.T) {
// 		recoveryEmail := "recover.expired@ory.sh"
// 		id := identity.Identity{Traits: identity.Traits(fmt.Sprintf(`{"email":"%s"}`, recoveryEmail))}

// 		require.NoError(t, reg.IdentityManager().Create(context.Background(),
// 			&id, identity.ManagerAllowWriteProtectedTraits))

// 		code, _, err := createCode(id.ID.String(), nil)
// 		require.NoError(t, err)

// 		time.Sleep(time.Millisecond * 100)
// 		require.NotEmpty(t, code.RecoveryLink)
// 		require.True(t, code.ExpiresAt.Before(time.Now().Add(conf.SelfServiceFlowRecoveryRequestLifespan())))

// 		res, err := publicTS.Client().Get(code.RecoveryLink)
// 		body := ioutilx.MustReadAll(res.Body)

// 		require.Len(t, gjson.GetBytes(body, "ui.messages").Array(), 1)
// 		require.Contains(t, gjson.GetBytes(body, "ui.messages.0.text").Str, "The recovery flow expired")

// 		// TODO: what does this do?
// 		addr, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, recoveryEmail)
// 		assert.NoError(t, err)
// 		assert.False(t, addr.Verified)
// 		assert.Nil(t, addr.VerifiedAt)
// 		assert.Equal(t, identity.VerifiableAddressStatusPending, addr.Status)
// 	})

// 	t.Run("description=should create a valid recovery link and set the expiry time as well and recover the account", func(t *testing.T) {
// 		recoveryEmail := "recoverme@ory.sh"
// 		id := identity.Identity{Traits: identity.Traits(fmt.Sprintf(`{"email":"%s"}`, recoveryEmail))}

// 		require.NoError(t, reg.IdentityManager().Create(context.Background(),
// 			&id, identity.ManagerAllowWriteProtectedTraits))

// 		code, _, err := createCode(id.ID.String(), nil)
// 		require.NoError(t, err)

// 		require.NotEmpty(t, code.RecoveryLink)
// 		require.True(t, code.ExpiresAt.Before(time.Now().Add(conf.SelfServiceFlowRecoveryRequestLifespan()+time.Second)))

// 		// TODO: Fix this
// 		// f, err := reg.SettingsFlowPersister().GetSettingsFlow(context.Background(), uuid.FromStringOrNil(res.Request.URL.Query().Get("flow")))
// 		// require.NoError(t, err, "%s", res.Request.URL.String())

// 		// require.Len(t, f.UI.Messages, 1)
// 		// assert.Equal(t, "You successfully recovered your account. Please change your password or set up an alternative login method (e.g. social sign in) within the next 60.00 minutes.", f.UI.Messages[0].Text)

// 		// addr, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, recoveryEmail)
// 		// assert.NoError(t, err)
// 		// assert.False(t, addr.Verified)
// 		// assert.Nil(t, addr.VerifiedAt)
// 		// assert.Equal(t, identity.VerifiableAddressStatusPending, addr.Status)
// 	})
// }

const (
	RecoveryFlowTypeBrowser string = "browser"
	RecoveryFlowTypeSPA     string = "spa"
	RecoveryFlowTypeAPI     string = "api"
)

var flowTypes = []string{RecoveryFlowTypeBrowser, RecoveryFlowTypeAPI, RecoveryFlowTypeSPA}

func createIdentityToRecover(t *testing.T, reg *driver.RegistryDefault, email string) *identity.Identity {
	t.Helper()
	var id = &identity.Identity{
		Credentials: map[identity.CredentialsType]identity.Credentials{
			"password": {
				Type:        "password",
				Identifiers: []string{email},
				Config:      sqlxx.JSONRawMessage(`{"hashed_password":"foo"}`),
			},
		},
		Traits:   identity.Traits(fmt.Sprintf(`{"email":"%s"}`, email)),
		SchemaID: config.DefaultIdentityTraitsSchemaID,
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
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+recovery.StrategyRecoveryLinkName+".enabled", false)
	initViper(t, conf)

	_ = testhelpers.NewRecoveryUIFlowEchoServer(t, reg)
	_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
	_ = testhelpers.NewSettingsUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)

	public, _, publicRouter, _ := testhelpers.NewKratosServerWithCSRFAndRouters(t, reg)

	var submitRecovery = func(t *testing.T, client *http.Client, flowType string, values func(url.Values), code int) string {
		isSPA := flowType == RecoveryFlowTypeSPA
		isAPI := isSPA || flowType == RecoveryFlowTypeAPI
		if client == nil {
			client = testhelpers.NewDebugClient(t)
			if !isAPI {
				client = testhelpers.NewClientWithCookies(t)
				client.Transport = testhelpers.NewTransportWithLogger(http.DefaultTransport, t).RoundTripper
			}
		}

		return testhelpers.SubmitRecoveryForm(t, isAPI, isSPA, client, public, values, code,
			testhelpers.ExpectURL(isAPI || isSPA, public.URL+recovery.RouteSubmitFlow, conf.SelfServiceFlowRecoveryUI().String()))
	}

	var submitAndExpectValidationError = func(t *testing.T, hc *http.Client, flowType string, values func(url.Values)) string {
		code := testhelpers.ExpectStatusCode(flowType == RecoveryFlowTypeAPI || flowType == RecoveryFlowTypeSPA, http.StatusBadRequest, http.StatusOK)
		return submitRecovery(t, hc, flowType, values, code)
	}

	var submitAndExpectSuccess = func(t *testing.T, hc *http.Client, flowType string, values func(url.Values)) string {
		return submitRecovery(t, hc, flowType, values, http.StatusOK)
	}

	t.Run("description=should recover an account", func(t *testing.T) {
		var check = func(t *testing.T, recoverySubmissionResponse, recoveryEmail, returnTo string) {
			addr, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, recoveryEmail)
			assert.NoError(t, err)
			assert.False(t, addr.Verified)
			assert.Nil(t, addr.VerifiedAt)
			assert.Equal(t, identity.VerifiableAddressStatusPending, addr.Status)

			assert.EqualValues(t, node.CodeGroup, gjson.Get(recoverySubmissionResponse, "active").String(), "%s", recoverySubmissionResponse)
			assert.True(t, gjson.Get(recoverySubmissionResponse, "ui.nodes.#(attributes.name==code)").Exists(), "%s", recoverySubmissionResponse)
			assert.Len(t, gjson.Get(recoverySubmissionResponse, "ui.messages").Array(), 1, "%s", recoverySubmissionResponse)
			assertx.EqualAsJSON(t, text.NewRecoveryEmailSent(), json.RawMessage(gjson.Get(recoverySubmissionResponse, "ui.messages.0").Raw))

			message := testhelpers.CourierExpectMessage(t, reg, recoveryEmail, "Recover access to your account")
			assert.Contains(t, message.Body, "please recover access to your account by entering the following code")

			recoveryCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)
			assert.NotEmpty(t, recoveryCode)

			cl := testhelpers.NewClientWithCookies(t)

			action := gjson.Get(recoverySubmissionResponse, "ui.action").String()
			assert.NotEmpty(t, action)

			res, err := cl.PostForm(action, url.Values{
				"code": {recoveryCode},
			})
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowSettingsUI().String())

			body := ioutilx.MustReadAll(res.Body)
			assert.Equal(t, text.NewRecoverySuccessful(time.Now().Add(time.Hour)).Text,
				gjson.GetBytes(body, "ui.messages.0.text").String())
			assert.Equal(t, returnTo, gjson.GetBytes(body, "return_to").String())

			addr, err = reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, recoveryEmail)
			assert.NoError(t, err)
			assert.True(t, addr.Verified)
			assert.NotEqual(t, sqlxx.NullTime{}, addr.VerifiedAt)
			assert.Equal(t, identity.VerifiableAddressStatusCompleted, addr.Status)

			res, err = cl.Get(public.URL + session.RouteWhoami)
			require.NoError(t, err)
			body = x.MustReadAll(res.Body)
			require.NoError(t, res.Body.Close())
			assert.Equal(t, "link_recovery", gjson.GetBytes(body, "authentication_methods.0.method").String(), "%s", body)
			assert.Equal(t, "aal1", gjson.GetBytes(body, "authenticator_assurance_level").String(), "%s", body)
		}

		t.Run("type=browser", func(t *testing.T) {
			email := "recoverme1@ory.sh"
			createIdentityToRecover(t, reg, email)
			check(t, submitAndExpectSuccess(t, nil, RecoveryFlowTypeBrowser, func(v url.Values) {
				v.Set("email", email)
			}), email, "")
		})

		t.Run("type=browser set return_to", func(t *testing.T) {
			email := "recoverme2@ory.sh"
			returnTo := "https://www.ory.sh"
			createIdentityToRecover(t, reg, email)

			hc := testhelpers.NewClientWithCookies(t)
			hc.Transport = testhelpers.NewTransportWithLogger(http.DefaultTransport, t).RoundTripper

			f := testhelpers.InitializeRecoveryFlowViaBrowser(t, hc, false, public, url.Values{"return_to": []string{returnTo}})

			time.Sleep(time.Millisecond) // add a bit of delay to allow `1ns` to time out.

			formPayload := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
			formPayload.Set("email", email)

			b, res := testhelpers.RecoveryMakeRequest(t, false, f, hc, testhelpers.EncodeFormAsJSON(t, false, formPayload))
			assert.EqualValues(t, http.StatusOK, res.StatusCode, "%s", b)
			expectedURL := testhelpers.ExpectURL(false, public.URL+recovery.RouteSubmitFlow, conf.SelfServiceFlowRecoveryUI().String())
			assert.Contains(t, res.Request.URL.String(), expectedURL, "%+v\n\t%s", res.Request, b)

			check(t, b, email, returnTo)
		})

		t.Run("type=spa", func(t *testing.T) {
			email := "recoverme3@ory.sh"
			createIdentityToRecover(t, reg, email)
			check(t, submitAndExpectSuccess(t, nil, RecoveryFlowTypeSPA, func(v url.Values) {
				v.Set("email", email)
			}), email, "")
		})

		t.Run("type=api", func(t *testing.T) {
			email := "recoverme4@ory.sh"
			createIdentityToRecover(t, reg, email)
			check(t, submitAndExpectSuccess(t, nil, RecoveryFlowTypeAPI, func(v url.Values) {
				v.Set("email", email)
			}), email, "")
		})
	})

	t.Run("description=should set all the correct recovery payloads after submission", func(t *testing.T) {
		body := submitAndExpectSuccess(t, nil, RecoveryFlowTypeBrowser, func(v url.Values) {
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
			t.Run("type="+flowType, func(t *testing.T) {
				body := submitAndExpectValidationError(t, nil, flowType, func(v url.Values) {
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
				t.Run("type="+flowType, func(t *testing.T) {
					responseJSON := submitAndExpectValidationError(t, nil, flowType, func(v url.Values) {
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
			t.Run("type="+flowType, func(t *testing.T) {
				isSPA := flowType == "spa"
				isAPI := isSPA || flowType == "api"
				hc := testhelpers.NewDebugClient(t)
				if !isAPI {
					hc = testhelpers.NewClientWithCookies(t)
					hc.Transport = testhelpers.NewTransportWithLogger(http.DefaultTransport, t).RoundTripper
				}

				var f *kratos.SelfServiceRecoveryFlow
				if isAPI {
					f = testhelpers.InitializeRecoveryFlowViaAPI(t, hc, public)
				} else {
					f = testhelpers.InitializeRecoveryFlowViaBrowser(t, hc, isSPA, public, nil)
				}

				authClient := testhelpers.NewHTTPClientWithArbitrarySessionToken(t, reg)
				if isAPI {
					s, err := session.NewActiveSession(
						&identity.Identity{ID: x.NewUUID(), State: identity.StateActive},
						testhelpers.NewSessionLifespanProvider(time.Hour),
						time.Now(),
						identity.CredentialsTypePassword,
						identity.AuthenticatorAssuranceLevel1,
					)
					require.NoError(t, err)
					authClient = testhelpers.NewHTTPClientWithSessionCookieLocalhost(t, reg, s)
				}

				v := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
				v.Set("email", "some-email@example.org")
				v.Set("method", "code")

				body, res := testhelpers.RecoveryMakeRequest(t, isAPI || isSPA, f, authClient, testhelpers.EncodeFormAsJSON(t, isAPI || isSPA, v))

				if isAPI || isSPA {
					assert.EqualValues(t, http.StatusBadRequest, res.StatusCode, "%s", body)
					assert.Contains(t, res.Request.URL.String(), recovery.RouteSubmitFlow, "%+v\n\t%s", res.Request, body)
					assertx.EqualAsJSONExcept(t, recovery.ErrAlreadyLoggedIn, json.RawMessage(gjson.Get(body, "error").Raw), nil)
				} else {
					assert.EqualValues(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.Contains(t, res.Request.URL.String(), conf.SelfServiceBrowserDefaultReturnTo().String(), "%+v\n\t%s", res.Request, body)
				}
			})
		}
	})

	t.Run("description=should not be able to recover account that does not exist", func(t *testing.T) {
		for _, flowType := range flowTypes {
			t.Run("type="+flowType, func(t *testing.T) {
				email := x.NewUUID().String() + "@ory.sh"
				var values = func(v url.Values) {
					v.Set("email", email)
				}

				actual := submitAndExpectSuccess(t, nil, flowType, values)

				assert.EqualValues(t, node.CodeGroup, gjson.Get(actual, "active").String(), "%s", actual)
				assert.Empty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==code).attributes.value").String(), "%s", actual)
				assertx.EqualAsJSON(t, text.NewRecoveryEmailSent(), json.RawMessage(gjson.Get(actual, "ui.messages.0").Raw))

				message := testhelpers.CourierExpectMessage(t, reg, email, "Account access attempted")
				assert.Contains(t, message.Body, "If this was you, check if you signed up using a different address.")
			})
		}
	})

	t.Run("description=should not be able to recover an inactive account", func(t *testing.T) {
		for _, flowType := range flowTypes {
			t.Run("type="+flowType, func(t *testing.T) {
				email := "recoverinactive_" + flowType + "@ory.sh"
				createIdentityToRecover(t, reg, email)
				values := func(v url.Values) {
					v.Set("email", email)
				}

				body := submitAndExpectSuccess(t, nil, flowType, values)
				addr, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, email)
				assert.NoError(t, err)

				emailText := testhelpers.CourierExpectMessage(t, reg, email, "Recover access to your account")
				recoveryCode := testhelpers.CourierExpectCodeInMessage(t, emailText, 1)
				cl := testhelpers.NewClientWithCookies(t)

				// Deactivate the identity
				require.NoError(t, reg.Persister().GetConnection(context.Background()).RawQuery("UPDATE identities SET state=? WHERE id = ?", identity.StateInactive, addr.IdentityID).Exec())

				action := gjson.Get(body, "ui.action").String()
				assert.NotEmpty(t, action)

				res, err := cl.PostForm(action, url.Values{
					"code": {recoveryCode},
				})
				require.NoError(t, err)

				body = string(ioutilx.MustReadAll(res.Body))
				if flowType == RecoveryFlowTypeAPI || flowType == RecoveryFlowTypeSPA {
					assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
					assert.Contains(t, res.Request.URL.String(), public.URL+recovery.RouteSubmitFlow)
					assertx.EqualAsJSON(t, session.ErrIdentityDisabled.WithDetail("identity_id", addr.IdentityID), json.RawMessage(gjson.Get(body, "error").Raw), "%s", body)
				} else {
					assert.Equal(t, http.StatusOK, res.StatusCode)
					assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowErrorURL().String())
					assertx.EqualAsJSON(t, session.ErrIdentityDisabled.WithDetail("identity_id", addr.IdentityID), json.RawMessage(body), "%s", body)
				}
			})
		}
	})

	t.Run("description=should recover an account and set the csrf cookies", func(t *testing.T) {
		var check = func(t *testing.T, responseBody, recoveryEmail string, cl *http.Client, do func(*http.Client, *http.Request) (*http.Response, error)) {
			message := testhelpers.CourierExpectMessage(t, reg, recoveryEmail, "Recover access to your account")
			recoveryCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)

			cl.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}

			form := url.Values{
				"code": {recoveryCode},
			}
			action := gjson.Get(responseBody, "ui.action").String()
			req := x.NewTestHTTPRequest(t, "POST", action, strings.NewReader(form.Encode()))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			res, err := do(cl, req)
			require.NoError(t, err)
			require.NoError(t, res.Body.Close())
			assert.Equal(t, http.StatusSeeOther, res.StatusCode)
			body := ioutilx.MustReadAll(res.Body)
			loc, err := res.Location()
			require.NoError(t, err)
			t.Logf("body: %s - loc: %s", body, loc)

			cookies := cl.Jar.Cookies(urlx.ParseOrPanic(public.URL))
			cookieStr := spew.Sdump(cookies)
			require.Len(t, cookies, 2, "expected 2 cookies, got %d: %s", len(cookies), cookieStr)
			assert.Contains(t, cookieStr, x.CSRFTokenName)
			assert.Contains(t, cookieStr, "ory_kratos_session")
			returnTo, err := res.Location()
			require.NoError(t, err)
			assert.Contains(t, returnTo.String(), conf.SelfServiceFlowSettingsUI().String(), "we end up at the settings screen")

			rl := urlx.ParseOrPanic(action)
			actualRes, err := cl.Get(public.URL + recovery.RouteGetFlow + "?id=" + rl.Query().Get("flow"))
			require.NoError(t, err)
			body = x.MustReadAll(actualRes.Body)
			require.NoError(t, actualRes.Body.Close())
			assert.Equal(t, http.StatusOK, actualRes.StatusCode, "%s", body)
			assert.Equal(t, string(recovery.StatePassedChallenge), gjson.GetBytes(body, "state").String(), "%s", body)
		}

		email := x.NewUUID().String() + "@ory.sh"
		id := createIdentityToRecover(t, reg, email)

		t.Run("case=unauthenticated", func(t *testing.T) {
			var values = func(v url.Values) {
				v.Set("email", email)
			}
			check(t, submitAndExpectSuccess(t, nil, RecoveryFlowTypeBrowser, values), email, testhelpers.NewClientWithCookies(t), (*http.Client).Do)
		})

		t.Run("case=already logged into another account", func(t *testing.T) {
			var values = func(v url.Values) {
				v.Set("email", email)
			}

			check(t, submitAndExpectSuccess(t, nil, RecoveryFlowTypeBrowser, values), email, testhelpers.NewClientWithCookies(t), func(cl *http.Client, req *http.Request) (*http.Response, error) {
				_, res := testhelpers.MockMakeAuthenticatedRequestWithClient(t, reg, conf, publicRouter.Router, req, cl)
				return res, nil
			})
		})

		t.Run("case=already logged into the account", func(t *testing.T) {
			var values = func(v url.Values) {
				v.Set("email", email)
			}

			cl := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id)
			check(t, submitAndExpectSuccess(t, nil, RecoveryFlowTypeBrowser, values), email, cl, func(_ *http.Client, req *http.Request) (*http.Response, error) {
				_, res := testhelpers.MockMakeAuthenticatedRequestWithClientAndID(t, reg, conf, publicRouter.Router, req, cl, id)
				return res, nil
			})
		})
	})

	t.Run("description=should recover and invalidate all other sessions if hook is set", func(t *testing.T) {
		conf.MustSet(config.HookStrategyKey(config.ViperKeySelfServiceRecoveryAfter, config.HookGlobal), []config.SelfServiceHook{{Name: "revoke_active_sessions"}})
		t.Cleanup(func() {
			conf.MustSet(config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()), nil)
		})

		email := strings.ToLower(testhelpers.RandomEmail())
		id := createIdentityToRecover(t, reg, email)

		sess, err := session.NewActiveSession(id, conf, time.Now(), identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
		require.NoError(t, err)
		require.NoError(t, reg.SessionPersister().UpsertSession(context.Background(), sess))

		actualSession, err := reg.SessionPersister().GetSession(context.Background(), sess.ID)
		require.NoError(t, err)
		assert.True(t, actualSession.IsActive())

		var values = func(v url.Values) {
			v.Set("email", email)
		}

		actual := submitAndExpectSuccess(t, nil, RecoveryFlowTypeBrowser, values)
		message := testhelpers.CourierExpectMessage(t, reg, email, "Recover access to your account")
		recoveryCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)

		cl := testhelpers.NewClientWithCookies(t)
		cl.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}

		action := gjson.Get(actual, "ui.action").String()

		res, err := cl.PostForm(action, url.Values{
			"code": {recoveryCode},
		})
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		assert.Equal(t, http.StatusSeeOther, res.StatusCode)
		require.Len(t, cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)), 2)
		cookies := spew.Sdump(cl.Jar.Cookies(urlx.ParseOrPanic(public.URL)))
		assert.Contains(t, cookies, "ory_kratos_session")

		actualSession, err = reg.SessionPersister().GetSession(context.Background(), sess.ID)
		require.NoError(t, err)
		assert.False(t, actualSession.IsActive())
	})

	t.Run("description=should not be able to use an invalid code", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)
		f := testhelpers.InitializeRecoveryFlowViaBrowser(t, c, false, public, nil)

		res, err := c.PostForm(f.Ui.Action, url.Values{
			"code": {"123123"},
		})

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowRecoveryUI().String()+"?flow=")

		rs, _, err := testhelpers.NewSDKCustomClient(public, c).V0alpha2Api.GetSelfServiceRecoveryFlow(context.Background()).Id(res.Request.URL.Query().Get("flow")).Execute()
		require.NoError(t, err)

		require.Len(t, rs.Ui.Messages, 1)
		// TODO: Change message "token -> code"
		assert.Equal(t, "The recovery token is invalid or has already been used. Please retry the flow.", rs.Ui.Messages[0].Text)
	})

	t.Run("description=should not be able to use an outdated flow", func(t *testing.T) {
		recoveryEmail := "recoverme5@ory.sh"
		createIdentityToRecover(t, reg, recoveryEmail)
		conf.MustSet(config.ViperKeySelfServiceRecoveryRequestLifespan, time.Millisecond*200)
		t.Cleanup(func() {
			conf.MustSet(config.ViperKeySelfServiceRecoveryRequestLifespan, time.Minute)
		})

		c := testhelpers.NewClientWithCookies(t)
		rs := testhelpers.GetRecoveryFlow(t, c, public)

		time.Sleep(time.Millisecond * 201)

		res, err := c.PostForm(rs.Ui.Action, url.Values{"email": {recoveryEmail}})
		require.NoError(t, err)
		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.NotContains(t, res.Request.URL.String(), "flow="+rs.Id)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowRecoveryUI().String())

		addr, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, recoveryEmail)
		assert.NoError(t, err)
		assert.False(t, addr.Verified)
		assert.Nil(t, addr.VerifiedAt)
		assert.Equal(t, identity.VerifiableAddressStatusPending, addr.Status)
	})

	t.Run("description=should not be able to use an outdated flow", func(t *testing.T) {
		recoveryEmail := "recoverme6@ory.sh"
		createIdentityToRecover(t, reg, recoveryEmail)
		conf.MustSet(config.ViperKeySelfServiceRecoveryRequestLifespan, time.Millisecond*200)
		t.Cleanup(func() {
			conf.MustSet(config.ViperKeySelfServiceRecoveryRequestLifespan, time.Minute)
		})

		c := testhelpers.NewClientWithCookies(t)
		body := submitAndExpectSuccess(t, c, RecoveryFlowTypeBrowser, func(v url.Values) {
			v.Set("email", recoveryEmail)
		})

		message := testhelpers.CourierExpectMessage(t, reg, recoveryEmail, "Recover access to your account")
		assert.Contains(t, message.Body, "please recover access to your account by entering the following code")

		recoveryCode := testhelpers.CourierExpectCodeInMessage(t, message, 1)

		time.Sleep(time.Millisecond * 201)

		action := gjson.Get(body, "ui.action").String()
		require.NotEmpty(t, action)

		res, err := c.PostForm(action, url.Values{
			"code": {recoveryCode},
		})
		require.NoError(t, err)

		require.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, res.Request.URL.String(), conf.SelfServiceFlowRecoveryUI().String())
		assert.NotContains(t, res.Request.URL.String(), gjson.Get(body, "id").String())

		rs, _, err := testhelpers.NewSDKCustomClient(public, c).V0alpha2Api.GetSelfServiceRecoveryFlow(context.Background()).Id(res.Request.URL.Query().Get("flow")).Execute()
		require.NoError(t, err)

		require.Len(t, rs.Ui.Messages, 1)
		assert.Contains(t, rs.Ui.Messages[0].Text, "The recovery flow expired")

		addr, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, recoveryEmail)
		require.NoError(t, err)
		assert.False(t, addr.Verified)
		assert.Nil(t, addr.VerifiedAt)
		assert.Equal(t, identity.VerifiableAddressStatusPending, addr.Status)
	})
}

func TestDisabledEndpoint(t *testing.T) {
	t.Skip()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	initViper(t, conf)
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+recovery.StrategyRecoveryCodeName+".enabled", false)

	publicTS, adminTS := testhelpers.NewKratosServer(t, reg)
	adminSDK := testhelpers.NewSDKClient(adminTS)
	_ = testhelpers.NewErrorTestServer(t, reg)

	t.Run("role=admin", func(t *testing.T) {
		t.Run("description=can not create recovery link when link method is disabled", func(t *testing.T) {
			id := identity.Identity{Traits: identity.Traits(`{"email":"recovery-endpoint-disabled@ory.sh"}`)}

			require.NoError(t, reg.IdentityManager().Create(context.Background(),
				&id, identity.ManagerAllowWriteProtectedTraits))

			rl, _, err := adminSDK.V0alpha2Api.AdminCreateSelfServiceRecoveryLink(context.Background()).AdminCreateSelfServiceRecoveryLinkBody(kratos.AdminCreateSelfServiceRecoveryLinkBody{
				IdentityId: id.ID.String(),
			}).Execute()
			assert.Nil(t, rl)
			require.IsType(t, new(kratos.GenericOpenAPIError), err, "%s", err)

			br, _ := err.(*kratos.GenericOpenAPIError)
			assert.Contains(t, string(br.Body()), "This endpoint was disabled by system administrator", "%s", br.Body())
		})
	})

	t.Run("role=public", func(t *testing.T) {
		c := testhelpers.NewClientWithCookies(t)

		t.Run("description=can not recover an account by get request when link method is disabled", func(t *testing.T) {
			f := testhelpers.InitializeRecoveryFlowViaBrowser(t, c, false, publicTS, nil)
			u := publicTS.URL + recovery.RouteSubmitFlow + "?flow=" + f.Id + "&token=endpoint-disabled"
			res, err := c.Get(u)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)

			b := ioutilx.MustReadAll(res.Body)
			assert.Contains(t, string(b), "This endpoint was disabled by system administrator")
		})

		t.Run("description=can not recover an account by post request when link method is disabled", func(t *testing.T) {
			f := testhelpers.InitializeRecoveryFlowViaBrowser(t, c, false, publicTS, nil)
			u := publicTS.URL + recovery.RouteSubmitFlow + "?flow=" + f.Id
			res, err := c.PostForm(u, url.Values{"email": {"email@ory.sh"}, "method": {"link"}})
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)

			b := ioutilx.MustReadAll(res.Body)
			assert.Contains(t, string(b), "This endpoint was disabled by system administrator")
		})
	})
}