// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package password_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/internal/registrationhelpers"

	"github.com/ory/kratos/selfservice/flow"

	"github.com/gofrs/uuid"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/hash"
	kratos "github.com/ory/kratos/internal/httpclient"
	"github.com/ory/x/assertx"
	"github.com/ory/x/errorsx"
	"github.com/ory/x/ioutilx"
	"github.com/ory/x/sqlxx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

//go:embed stub/login.schema.json
var loginSchema []byte

func createIdentity(ctx context.Context, reg *driver.RegistryDefault, t *testing.T, identifier, password string) {
	p, _ := reg.Hasher(ctx).Generate(context.Background(), []byte(password))
	iId := x.NewUUID()
	require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &identity.Identity{
		ID:     iId,
		Traits: identity.Traits(fmt.Sprintf(`{"subject":"%s"}`, identifier)),
		Credentials: map[identity.CredentialsType]identity.Credentials{
			identity.CredentialsTypePassword: {
				Type:        identity.CredentialsTypePassword,
				Identifiers: []string{identifier},
				Config:      sqlxx.JSONRawMessage(`{"hashed_password":"` + string(p) + `"}`),
			},
		},
		VerifiableAddresses: []identity.VerifiableAddress{
			{
				ID:         x.NewUUID(),
				Value:      identifier,
				Verified:   false,
				CreatedAt:  time.Now(),
				IdentityID: iId,
			},
		},
	}))
}

func TestCompleteLogin(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword),
		map[string]interface{}{"enabled": true})
	router := x.NewRouterPublic()
	publicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin())

	errTS := testhelpers.NewErrorTestServer(t, reg)
	uiTS := testhelpers.NewLoginUIFlowEchoServer(t, reg)
	redirTS := testhelpers.NewRedirSessionEchoTS(t, reg)

	// Overwrite these two:
	conf.MustSet(ctx, config.ViperKeySelfServiceErrorUI, errTS.URL+"/error-ts")
	conf.MustSet(ctx, config.ViperKeySelfServiceLoginUI, uiTS.URL+"/login-ts")

	testhelpers.SetDefaultIdentitySchemaFromRaw(conf, loginSchema)
	conf.MustSet(ctx, config.ViperKeySecretsDefault, []string{"not-a-secure-session-key"})

	ensureFieldsExist := func(t *testing.T, body []byte) {
		registrationhelpers.CheckFormContent(t, body, "identifier",
			"password",
			"csrf_token")
	}

	apiClient := testhelpers.NewDebugClient(t)

	t.Run("case=should show the error ui because the request payload is malformed", func(t *testing.T) {
		t.Run("type=api", func(t *testing.T) {
			f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false)

			body, res := testhelpers.LoginMakeRequest(t, true, false, f, apiClient, "14=)=!(%)$/ZP()GHIÖ")
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, body, `Expected JSON sent in request body to be an object but got: Number`)
		})

		t.Run("type=browser", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, false, false, false)

			body, res := testhelpers.LoginMakeRequest(t, false, false, f, browserClient, "14=)=!(%)$/ZP()GHIÖ")
			assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/login-ts")
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "invalid URL escape", "%s", body)
		})

		t.Run("type=spa", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, true, false, false)

			body, res := testhelpers.LoginMakeRequest(t, false, true, f, browserClient, "14=)=!(%)$/ZP()GHIÖ")
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "invalid URL escape", "%s", body)
		})
	})

	t.Run("case=should fail because password can not handle AAL2", func(t *testing.T) {
		f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false)

		update, err := reg.LoginFlowPersister().GetLoginFlow(context.Background(), uuid.FromStringOrNil(f.Id))
		require.NoError(t, err)
		update.RequestedAAL = identity.AuthenticatorAssuranceLevel2
		require.NoError(t, reg.LoginFlowPersister().UpdateLoginFlow(context.Background(), update))

		req, err := http.NewRequest("POST", f.Ui.Action, bytes.NewBufferString(`{"method":"password"}`))
		require.NoError(t, err)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		actual, res := testhelpers.MockMakeAuthenticatedRequest(t, reg, conf, router.Router, req)
		assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
		assert.Equal(t, text.NewErrorValidationLoginNoStrategyFound().Text, gjson.GetBytes(actual, "ui.messages.0.text").String())
	})

	t.Run("should return an error because the request does not exist", func(t *testing.T) {
		var check = func(t *testing.T, actual string) {
			assert.Equal(t, int64(http.StatusNotFound), gjson.Get(actual, "code").Int(), "%s", actual)
			assert.Equal(t, "Not Found", gjson.Get(actual, "status").String(), "%s", actual)
			assert.Contains(t, gjson.Get(actual, "message").String(), "Unable to locate the resource", "%s", actual)
		}

		fakeFlow := &kratos.LoginFlow{
			Ui: kratos.UiContainer{
				Action: publicTS.URL + login.RouteSubmitFlow + "?flow=" + x.NewUUID().String()},
		}

		t.Run("type=api", func(t *testing.T) {
			actual, res := testhelpers.LoginMakeRequest(t, true, false, fakeFlow, apiClient, "{}")
			assert.Len(t, res.Cookies(), 0)
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
			check(t, gjson.Get(actual, "error").Raw)
		})

		t.Run("type=browser", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			actual, res := testhelpers.LoginMakeRequest(t, false, false, fakeFlow, browserClient, "")
			assert.Contains(t, res.Request.URL.String(), errTS.URL)
			check(t, actual)
		})

		t.Run("type=api", func(t *testing.T) {
			actual, res := testhelpers.LoginMakeRequest(t, false, true, fakeFlow, apiClient, "{}")
			assert.Len(t, res.Cookies(), 0)
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
			check(t, gjson.Get(actual, "error").Raw)
		})
	})

	t.Run("case=should return an error because the request is expired", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceLoginRequestLifespan, "50ms")
		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceLoginRequestLifespan, "10m")
		})
		values := url.Values{
			"csrf_token": {x.FakeCSRFToken},
			"identifier": {"identifier"},
			"password":   {"password"},
		}

		t.Run("type=api", func(t *testing.T) {
			f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false)

			time.Sleep(time.Millisecond * 60)
			actual, res := testhelpers.LoginMakeRequest(t, true, false, f, apiClient, testhelpers.EncodeFormAsJSON(t, true, values))
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
			assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", gjson.Get(actual, "use_flow_id").String())
			assertx.EqualAsJSONExcept(t, flow.NewFlowExpiredError(time.Now()), json.RawMessage(actual), []string{"use_flow_id", "since", "expired_at"}, "expired", "%s", actual)
		})

		t.Run("type=browser", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, false, false, false)

			time.Sleep(time.Millisecond * 60)
			actual, res := testhelpers.LoginMakeRequest(t, false, false, f, browserClient, values.Encode())
			assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/login-ts")
			assert.NotEqual(t, f.Id, gjson.Get(actual, "id").String(), "%s", actual)
			assert.Contains(t, gjson.Get(actual, "ui.messages.0.text").String(), "expired", "%s", actual)
		})

		t.Run("type=SPA", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, true, false, false)

			time.Sleep(time.Millisecond * 60)
			actual, res := testhelpers.LoginMakeRequest(t, false, true, f, apiClient, testhelpers.EncodeFormAsJSON(t, true, values))
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
			assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", gjson.Get(actual, "use_flow_id").String())
			assertx.EqualAsJSONExcept(t, flow.NewFlowExpiredError(time.Now()), json.RawMessage(actual), []string{"use_flow_id", "since", "expired_at"}, "expired", "%s", actual)
		})
	})

	t.Run("case=should have correct CSRF behavior", func(t *testing.T) {
		var values = url.Values{
			"method":     {"password"},
			"csrf_token": {"invalid_token"},
			"identifier": {"login-identifier-csrf-browser"},
			"password":   {x.NewUUID().String()},
		}

		t.Run("case=should fail because of missing CSRF token/type=browser", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, false, false, false)

			actual, res := testhelpers.LoginMakeRequest(t, false, false, f, browserClient, values.Encode())
			assert.EqualValues(t, http.StatusOK, res.StatusCode)
			assertx.EqualAsJSON(t, x.ErrInvalidCSRFToken,
				json.RawMessage(actual), "%s", actual)
		})

		t.Run("case=should fail because of missing CSRF token/type=spa", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, true, false, false)

			actual, res := testhelpers.LoginMakeRequest(t, false, true, f, browserClient, values.Encode())
			assert.EqualValues(t, http.StatusForbidden, res.StatusCode)
			assertx.EqualAsJSON(t, x.ErrInvalidCSRFToken,
				json.RawMessage(gjson.Get(actual, "error").Raw), "%s", actual)
		})

		t.Run("case=should pass even without CSRF token/type=api", func(t *testing.T) {
			f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false)

			actual, res := testhelpers.LoginMakeRequest(t, true, false, f, apiClient, testhelpers.EncodeFormAsJSON(t, true, values))
			assert.EqualValues(t, http.StatusBadRequest, res.StatusCode)
			assert.Contains(t, actual, "provided credentials are invalid")
		})

		t.Run("case=should fail with correct CSRF error cause/type=api", func(t *testing.T) {
			for k, tc := range []struct {
				mod func(http.Header)
				exp string
			}{
				{
					mod: func(h http.Header) {
						h.Add("Cookie", "name=bar")
					},
					exp: "The HTTP Request Header included the \\\"Cookie\\\" key",
				},
				{
					mod: func(h http.Header) {
						h.Add("Origin", "www.bar.com")
					},
					exp: "The HTTP Request Header included the \\\"Origin\\\" key",
				},
			} {
				t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
					f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false)

					req := testhelpers.NewRequest(t, true, "POST", f.Ui.Action, bytes.NewBufferString(testhelpers.EncodeFormAsJSON(t, true, values)))
					tc.mod(req.Header)

					res, err := apiClient.Do(req)
					require.NoError(t, err)
					defer res.Body.Close()

					actual := string(ioutilx.MustReadAll(res.Body))
					assert.EqualValues(t, http.StatusBadRequest, res.StatusCode)
					assert.Contains(t, actual, tc.exp)
				})
			}
		})
	})

	var expectValidationError = func(t *testing.T, isAPI, refresh, isSPA bool, values func(url.Values)) string {
		return testhelpers.SubmitLoginForm(t, isAPI, nil, publicTS, values,
			isSPA, refresh,
			testhelpers.ExpectStatusCode(isAPI || isSPA, http.StatusBadRequest, http.StatusOK),
			testhelpers.ExpectURL(isAPI || isSPA, publicTS.URL+login.RouteSubmitFlow, conf.SelfServiceFlowLoginUI(ctx).String()))
	}

	t.Run("should return an error because the credentials are invalid (user does not exist)", func(t *testing.T) {
		var check = func(t *testing.T, body string, start time.Time) {
			delay := time.Since(start)
			minConfiguredDelay := conf.HasherArgon2(ctx).ExpectedDuration - conf.HasherArgon2(ctx).ExpectedDeviation
			assert.GreaterOrEqual(t, delay, minConfiguredDelay)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.action").String(), publicTS.URL+login.RouteSubmitFlow, "%s", body)
			assert.Equal(t, text.NewErrorValidationInvalidCredentials().Text, gjson.Get(body, "ui.messages.0.text").String(), body)
		}

		var values = func(v url.Values) {
			v.Set("identifier", "identifier")
			v.Set("password", "password")
		}

		t.Run("type=browser", func(t *testing.T) {
			start := time.Now()
			check(t, expectValidationError(t, false, false, false, values), start)
		})

		t.Run("type=SPA", func(t *testing.T) {
			start := time.Now()
			check(t, expectValidationError(t, false, false, true, values), start)
		})

		t.Run("type=api", func(t *testing.T) {
			start := time.Now()
			check(t, expectValidationError(t, true, false, false, values), start)
		})
	})

	t.Run("should return an error because no identifier is set", func(t *testing.T) {
		var check = func(t *testing.T, body string) {
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.action").String(), publicTS.URL+login.RouteSubmitFlow, "%s", body)

			ensureFieldsExist(t, []byte(body))
			assert.Equal(t, "Property identifier is missing.", gjson.Get(body, "ui.nodes.#(attributes.name==identifier).messages.0.text").String(), "%s", body)
			assert.Len(t, gjson.Get(body, "ui.nodes").Array(), 4)

			// The password value should not be returned!
			assert.Empty(t, gjson.Get(body, "ui.nodes.#(attributes.name==password).attributes.value").String())
		}

		var values = func(v url.Values) {
			v.Del("identifier")
			v.Set("method", identity.CredentialsTypePassword.String())
			v.Set("password", "password")
		}

		t.Run("type=browser", func(t *testing.T) {
			check(t, expectValidationError(t, false, false, false, values))
		})

		t.Run("type=spa", func(t *testing.T) {
			check(t, expectValidationError(t, false, false, true, values))
		})

		t.Run("type=api", func(t *testing.T) {
			check(t, expectValidationError(t, true, false, false, values))
		})
	})

	t.Run("should return an error because no password is set", func(t *testing.T) {
		var check = func(t *testing.T, body string) {
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.action").String(), publicTS.URL+login.RouteSubmitFlow, "%s", body)

			ensureFieldsExist(t, []byte(body))
			assert.Equal(t, "Property password is missing.", gjson.Get(body, "ui.nodes.#(attributes.name==password).messages.0.text").String(), "%s", body)
			assert.Equal(t, "identifier", gjson.Get(body, "ui.nodes.#(attributes.name==identifier).attributes.value").String(), "%s", body)
			assert.Len(t, gjson.Get(body, "ui.nodes").Array(), 4)

			// This must not include the password!
			assert.Empty(t, gjson.Get(body, "ui.nodes.#(attributes.name==password).attributes.value").String())
		}

		var values = func(v url.Values) {
			v.Set("identifier", "identifier")
			v.Del("password")
		}

		t.Run("type=browser", func(t *testing.T) {
			check(t, expectValidationError(t, false, false, false, values))
		})

		t.Run("type=api", func(t *testing.T) {
			check(t, expectValidationError(t, true, false, false, values))
		})
	})

	t.Run("should return an error both identifier and password are missing", func(t *testing.T) {
		var check = func(t *testing.T, body string) {
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.action").String(), publicTS.URL+login.RouteSubmitFlow, "%s", body)

			ensureFieldsExist(t, []byte(body))
			assert.Equal(t, "Property password is missing.", gjson.Get(body, "ui.nodes.#(attributes.name==password).messages.0.text").String(), "%s", body)
			assert.Equal(t, "Property identifier is missing.", gjson.Get(body, "ui.nodes.#(attributes.name==identifier).messages.0.text").String(), "%s", body)
			assert.Len(t, gjson.Get(body, "ui.nodes").Array(), 4)

			// This must not include the password!
			assert.Empty(t, gjson.Get(body, "ui.nodes.#(attributes.name==password).attributes.value").String())
		}

		var values = func(v url.Values) {
			v.Set("password", "")
			v.Set("identifier", "")
		}

		t.Run("type=browser", func(t *testing.T) {
			check(t, expectValidationError(t, false, false, false, values))
		})

		t.Run("type=spa", func(t *testing.T) {
			check(t, expectValidationError(t, true, false, true, values))
		})

		t.Run("type=api", func(t *testing.T) {
			check(t, expectValidationError(t, true, false, false, values))
		})
	})

	t.Run("should return an error because the credentials are invalid (password not correct)", func(t *testing.T) {
		var check = func(t *testing.T, body string) {
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.action").String(), publicTS.URL+login.RouteSubmitFlow, "%s", body)

			ensureFieldsExist(t, []byte(body))
			assert.Equal(t,
				errorsx.Cause(schema.NewInvalidCredentialsError()).(*schema.ValidationError).Messages[0].Text,
				gjson.Get(body, "ui.messages.0.text").String(),
				"%s", body,
			)

			// This must not include the password!
			assert.Empty(t, gjson.Get(body, "ui.nodes.#(attributes.name==password).attributes.value").String())
		}

		identifier, pwd := x.NewUUID().String(), "password"
		createIdentity(ctx, reg, t, identifier, pwd)

		var values = func(v url.Values) {
			v.Set("identifier", identifier)
			v.Set("password", "not-password")
		}

		t.Run("type=browser", func(t *testing.T) {
			check(t, expectValidationError(t, false, false, false, values))
		})

		t.Run("type=api", func(t *testing.T) {
			check(t, expectValidationError(t, true, false, false, values))
		})
		t.Run("type=spa", func(t *testing.T) {
			check(t, expectValidationError(t, true, false, true, values))
		})
	})

	t.Run("should pass with real request", func(t *testing.T) {
		identifier, pwd := x.NewUUID().String(), "password"
		createIdentity(ctx, reg, t, identifier, pwd)

		var values = func(v url.Values) {
			v.Set("identifier", identifier)
			v.Set("password", pwd)
		}

		t.Run("type=browser", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)

			body := testhelpers.SubmitLoginForm(t, false, browserClient, publicTS, values,
				false, false, http.StatusOK, redirTS.URL)

			assert.Equal(t, identifier, gjson.Get(body, "identity.traits.subject").String(), "%s", body)

			t.Run("retry with different refresh", func(t *testing.T) {
				t.Run("redirect to returnTS if refresh is missing", func(t *testing.T) {
					res, err := browserClient.Get(publicTS.URL + login.RouteInitBrowserFlow)
					require.NoError(t, err)
					require.EqualValues(t, http.StatusOK, res.StatusCode)
				})

				t.Run("show UI and hint at username", func(t *testing.T) {
					res, err := browserClient.Get(publicTS.URL + login.RouteInitBrowserFlow + "?refresh=true")
					require.NoError(t, err)
					require.EqualValues(t, http.StatusOK, res.StatusCode)

					rid := res.Request.URL.Query().Get("flow")
					assert.NotEmpty(t, rid, "%s", res.Request.URL)

					res, err = browserClient.Get(publicTS.URL + login.RouteGetFlow + "?id=" + rid)
					require.NoError(t, err)
					require.EqualValues(t, http.StatusOK, res.StatusCode)

					body, err := io.ReadAll(res.Body)
					require.NoError(t, err)
					assert.True(t, gjson.GetBytes(body, "refresh").Bool())
					assert.Equal(t, identifier, gjson.GetBytes(body, "ui.nodes.#(attributes.name==identifier).attributes.value").String(), "%s", body)
					assert.Empty(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==password).attributes.value").String(), "%s", body)
					assert.True(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==password)").Exists(), "%s", body)
				})
			})

			t.Run("do not show password method if identity has no password set", func(t *testing.T) {
				id := identity.NewIdentity("")
				browserClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id)

				res, err := browserClient.Get(publicTS.URL + login.RouteInitBrowserFlow + "?refresh=true")
				require.NoError(t, err)
				require.EqualValues(t, http.StatusOK, res.StatusCode)

				rid := res.Request.URL.Query().Get("flow")
				assert.NotEmpty(t, rid, "%s", res.Request.URL)

				res, err = browserClient.Get(publicTS.URL + login.RouteGetFlow + "?id=" + rid)
				require.NoError(t, err)
				require.EqualValues(t, http.StatusOK, res.StatusCode)

				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.True(t, gjson.GetBytes(body, "refresh").Bool())
				assert.False(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==identifier)").Exists(), "%s", body)
				assert.False(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==identifier)").Exists(), "%s", body)
				assert.False(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==password)").Exists(), "%s", body)
			})
		})

		t.Run("type=spa", func(t *testing.T) {
			hc := testhelpers.NewClientWithCookies(t)

			body := testhelpers.SubmitLoginForm(t, false, hc, publicTS, values,
				true, false, http.StatusOK, publicTS.URL+login.RouteSubmitFlow)

			assert.Equal(t, identifier, gjson.Get(body, "session.identity.traits.subject").String(), "%s", body)
			assert.Empty(t, gjson.Get(body, "session_token").String(), "%s", body)
			assert.Empty(t, gjson.Get(body, "session.token").String(), "%s", body)

			// Was the session cookie set?
			require.NotEmpty(t, hc.Jar.Cookies(urlx.ParseOrPanic(publicTS.URL)), "%+v", hc.Jar)

			t.Run("retry with different refresh", func(t *testing.T) {
				t.Run("redirect to returnTS if refresh is missing", func(t *testing.T) {
					res, err := hc.Do(testhelpers.NewHTTPGetAJAXRequest(t, publicTS.URL+login.RouteInitBrowserFlow))
					require.NoError(t, err)
					defer res.Body.Close()
					body := ioutilx.MustReadAll(res.Body)

					assert.EqualValues(t, http.StatusBadRequest, res.StatusCode, "%s", body)
					assertx.EqualAsJSON(t, login.ErrAlreadyLoggedIn, json.RawMessage(gjson.GetBytes(body, "error").Raw), "%s", body)
				})

				t.Run("show UI and hint at username", func(t *testing.T) {
					res, err := hc.Do(testhelpers.NewHTTPGetAJAXRequest(t, publicTS.URL+login.RouteInitBrowserFlow+"?refresh=true"))
					require.NoError(t, err)
					defer res.Body.Close()
					body := ioutilx.MustReadAll(res.Body)

					assert.True(t, gjson.GetBytes(body, "refresh").Bool())
					assert.Equal(t, identifier, gjson.GetBytes(body, "ui.nodes.#(attributes.name==identifier).attributes.value").String(), "%s", body)
					assert.Empty(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==password).attributes.value").String(), "%s", body)
				})
			})

			t.Run("do not show password method if identity has no password set", func(t *testing.T) {
				id := identity.NewIdentity("")
				hc := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id)

				res, err := hc.Do(testhelpers.NewHTTPGetAJAXRequest(t, publicTS.URL+login.RouteInitBrowserFlow+"?refresh=true"))
				require.NoError(t, err)
				defer res.Body.Close()
				body := ioutilx.MustReadAll(res.Body)

				assert.True(t, gjson.GetBytes(body, "refresh").Bool())
				assert.False(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==identifier)").Exists(), "%s", body)
				assert.False(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==identifier)").Exists(), "%s", body)
				assert.False(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==password)").Exists(), "%s", body)
			})
		})

		t.Run("type=api", func(t *testing.T) {
			body := testhelpers.SubmitLoginForm(t, true, nil, publicTS, values,
				false, false, http.StatusOK, publicTS.URL+login.RouteSubmitFlow)

			assert.Equal(t, identifier, gjson.Get(body, "session.identity.traits.subject").String(), "%s", body)
			st := gjson.Get(body, "session_token").String()
			assert.NotEmpty(t, st, "%s", body)

			t.Run("retry with different refresh", func(t *testing.T) {
				c := &http.Client{Transport: x.NewTransportWithHeader(http.Header{"Authorization": {"Bearer " + st}})}

				t.Run("redirect to returnTS if refresh is missing", func(t *testing.T) {
					res, err := c.Do(testhelpers.NewHTTPGetJSONRequest(t, publicTS.URL+login.RouteInitAPIFlow))
					require.NoError(t, err)
					defer res.Body.Close()
					body := ioutilx.MustReadAll(res.Body)

					require.EqualValues(t, http.StatusBadRequest, res.StatusCode)
					assertx.EqualAsJSON(t, login.ErrAlreadyLoggedIn, json.RawMessage(gjson.GetBytes(body, "error").Raw), "%s", body)
				})

				t.Run("show UI and hint at username", func(t *testing.T) {
					res, err := c.Do(testhelpers.NewHTTPGetJSONRequest(t, publicTS.URL+login.RouteInitAPIFlow+"?refresh=true"))
					require.NoError(t, err)
					defer res.Body.Close()
					body := ioutilx.MustReadAll(res.Body)

					assert.True(t, gjson.GetBytes(body, "refresh").Bool())
					assert.Equal(t, identifier, gjson.GetBytes(body, "ui.nodes.#(attributes.name==identifier).attributes.value").String(), "%s", body)
					assert.Empty(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==password).attributes.value").String(), "%s", body)
				})

				t.Run("show verification confirmation when refresh is set to true", func(t *testing.T) {
					res, err := c.Do(testhelpers.NewHTTPGetJSONRequest(t, publicTS.URL+login.RouteInitAPIFlow+"?refresh=true"))
					require.NoError(t, err)
					defer res.Body.Close()
					body := ioutilx.MustReadAll(res.Body)

					assert.True(t, gjson.GetBytes(body, "refresh").Bool())
					assert.Contains(t, gjson.GetBytes(body, "ui.messages.0.text").String(), "verifying that", "%s", body)
				})
			})

			t.Run("do not show password method if identity has no password set", func(t *testing.T) {
				id := identity.NewIdentity("")
				hc := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)

				res, err := hc.Do(testhelpers.NewHTTPGetAJAXRequest(t, publicTS.URL+login.RouteInitAPIFlow+"?refresh=true"))
				require.NoError(t, err)
				defer res.Body.Close()
				body := ioutilx.MustReadAll(res.Body)

				assert.True(t, gjson.GetBytes(body, "refresh").Bool())
				assert.False(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==identifier)").Exists(), "%s", body)
				assert.False(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==identifier)").Exists(), "%s", body)
				assert.False(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==password)").Exists(), "%s", body)
			})
		})
	})

	t.Run("case=should return an error because not passing validation and reset previous errors and values", func(t *testing.T) {
		testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/login.schema.json")

		var check = func(t *testing.T, actual string) {
			assert.NotEmpty(t, gjson.Get(actual, "id").String(), "%s", actual)
			assert.Contains(t, gjson.Get(actual, "ui.action").String(), publicTS.URL+login.RouteSubmitFlow, "%s", actual)
		}

		var checkFirst = func(t *testing.T, actual string) {
			check(t, actual)
			assert.Contains(t, gjson.Get(actual, "ui.nodes.#(attributes.name==identifier).messages.0").String(), "Property identifier is missing.", "%s", actual)
		}

		var checkSecond = func(t *testing.T, actual string) {
			check(t, actual)

			assert.Empty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==identifier).attributes.error"))
			assert.EqualValues(t, "identifier", gjson.Get(actual, "ui.nodes.#(attributes.name==identifier).attributes.value").String(), actual)
			assert.EqualValues(t, "password", gjson.Get(actual, "ui.nodes.#(attributes.name==method).attributes.value").String(), actual)
			assert.Empty(t, gjson.Get(actual, "ui.error"))
			assert.Contains(t, gjson.Get(actual, "ui.nodes.#(attributes.name==password).messages.0").String(), "Property password is missing.", "%s", actual)
		}

		var valuesFirst = func(v url.Values) url.Values {
			v.Del("identifier")
			v.Set("password", x.NewUUID().String())
			return v
		}

		var valuesSecond = func(v url.Values) url.Values {
			v.Set("identifier", "identifier")
			v.Del("password")
			return v
		}

		t.Run("type=api", func(t *testing.T) {
			f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false)

			actual, _ := testhelpers.LoginMakeRequest(t, true, false, f, apiClient, testhelpers.EncodeFormAsJSON(t, true, valuesFirst(testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes))))
			checkFirst(t, actual)
			actual, _ = testhelpers.LoginMakeRequest(t, true, false, f, apiClient, testhelpers.EncodeFormAsJSON(t, true, valuesSecond(testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes))))
			checkSecond(t, actual)
		})

		t.Run("type=browser", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, false, false, false)

			actual, _ := testhelpers.LoginMakeRequest(t, false, false, f, browserClient, valuesFirst(testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)).Encode())
			checkFirst(t, actual)
			actual, _ = testhelpers.LoginMakeRequest(t, false, false, f, browserClient, valuesSecond(testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)).Encode())
			checkSecond(t, actual)
		})
	})

	t.Run("should be a new session with refresh flag", func(t *testing.T) {
		identifier, pwd := x.NewUUID().String(), "password"
		createIdentity(ctx, reg, t, identifier, pwd)

		browserClient := testhelpers.NewClientWithCookies(t)
		f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, false, false, false)

		values := url.Values{"method": {"password"}, "identifier": {identifier},
			"password": {pwd}, "csrf_token": {x.FakeCSRFToken}}.Encode()

		body1, res := testhelpers.LoginMakeRequest(t, false, false, f, browserClient, values)
		assert.EqualValues(t, http.StatusOK, res.StatusCode)

		f = testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, true, false, false, false)
		body2, res := testhelpers.LoginMakeRequest(t, false, false, f, browserClient, values)

		require.Contains(t, res.Request.URL.Path, "return-ts", "%s", res.Request.URL.String())
		assert.Equal(t, identifier, gjson.Get(body2, "identity.traits.subject").String(), "%s", body2)
		assert.Equal(t, gjson.Get(body1, "id").String(), gjson.Get(body2, "id").String(), "%s\n\n%s\n", body1, body2)
	})

	t.Run("should login same identity regardless of identifier capitalization", func(t *testing.T) {
		identifier, pwd := x.NewUUID().String(), "password"
		createIdentity(ctx, reg, t, identifier, pwd)

		browserClient := testhelpers.NewClientWithCookies(t)
		f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, false, false, false)

		values := url.Values{"method": {"password"}, "identifier": {strings.ToUpper(identifier)}, "password": {pwd}, "csrf_token": {x.FakeCSRFToken}}.Encode()

		body, res := testhelpers.LoginMakeRequest(t, false, false, f, browserClient, values)

		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, identifier, gjson.Get(body, "identity.traits.subject").String(), "%s", body)
	})

	t.Run("should login even if old form field name is used", func(t *testing.T) {
		identifier, pwd := x.NewUUID().String(), "password"
		createIdentity(ctx, reg, t, identifier, pwd)

		browserClient := testhelpers.NewClientWithCookies(t)
		f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, false, false, false)

		values := url.Values{"method": {"password"}, "password_identifier": {strings.ToUpper(identifier)}, "password": {pwd}, "csrf_token": {x.FakeCSRFToken}}.Encode()

		body, res := testhelpers.LoginMakeRequest(t, false, false, f, browserClient, values)

		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, identifier, gjson.Get(body, "identity.traits.subject").String(), "%s", body)
	})

	t.Run("should login same identity regardless of leading or trailing whitespace", func(t *testing.T) {
		identifier, pwd := x.NewUUID().String(), "password"
		createIdentity(ctx, reg, t, identifier, pwd)

		browserClient := testhelpers.NewClientWithCookies(t)
		f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, false, false, false)

		values := url.Values{"method": {"password"}, "identifier": {"  " + identifier + "  "}, "password": {pwd}, "csrf_token": {x.FakeCSRFToken}}.Encode()

		body, res := testhelpers.LoginMakeRequest(t, false, false, f, browserClient, values)

		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, identifier, gjson.Get(body, "identity.traits.subject").String(), "%s", body)
	})

	t.Run("should fail as email is not yet verified", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceLoginAfter+".password.hooks", []map[string]interface{}{
			{"hook": "require_verified_address"},
		})

		identifier, pwd := x.NewUUID().String(), "password"
		createIdentity(ctx, reg, t, identifier, pwd)

		var values = func(v url.Values) {
			v.Set("method", "password")
			v.Set("identifier", identifier)
			v.Set("password", pwd)
		}

		var check = func(t *testing.T, body string) {
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.action").String(), publicTS.URL+login.RouteSubmitFlow, "%s", body)

			ensureFieldsExist(t, []byte(body))
			assert.Equal(t,
				errorsx.Cause(schema.NewAddressNotVerifiedError()).(*schema.ValidationError).Messages[0].Text,
				gjson.Get(body, "ui.messages.0.text").String(),
				"%s", body,
			)
		}

		t.Run("type=browser", func(t *testing.T) {
			check(t, expectValidationError(t, false, false, false, values))
		})

		t.Run("type=api", func(t *testing.T) {
			check(t, expectValidationError(t, true, false, false, values))
		})

	})

	t.Run("should upgrade password not primary hashing algorithm", func(t *testing.T) {
		identifier, pwd := x.NewUUID().String(), "password"
		h := &hash.Pbkdf2{
			Algorithm:  "sha256",
			Iterations: 100000,
			SaltLength: 32,
			KeyLength:  32,
		}
		p, _ := h.Generate(context.Background(), []byte(pwd))

		iId := x.NewUUID()
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &identity.Identity{
			ID:     iId,
			Traits: identity.Traits(fmt.Sprintf(`{"subject":"%s"}`, identifier)),
			Credentials: map[identity.CredentialsType]identity.Credentials{
				identity.CredentialsTypePassword: {
					Type:        identity.CredentialsTypePassword,
					Identifiers: []string{identifier},
					Config:      sqlxx.JSONRawMessage(`{"hashed_password":"` + string(p) + `"}`),
				},
			},
			VerifiableAddresses: []identity.VerifiableAddress{
				{
					ID:         x.NewUUID(),
					Value:      identifier,
					Verified:   true,
					CreatedAt:  time.Now(),
					IdentityID: iId,
				},
			},
		}))

		var values = func(v url.Values) {
			v.Set("identifier", identifier)
			v.Set("method", identity.CredentialsTypePassword.String())
			v.Set("password", pwd)
		}

		browserClient := testhelpers.NewClientWithCookies(t)

		body := testhelpers.SubmitLoginForm(t, false, browserClient, publicTS, values,
			false, false, http.StatusOK, redirTS.URL)

		assert.Equal(t, identifier, gjson.Get(body, "identity.traits.subject").String(), "%s", body)

		// check if password hash algorithm is upgraded
		_, c, err := reg.PrivilegedIdentityPool().FindByCredentialsIdentifier(context.Background(), identity.CredentialsTypePassword, identifier)
		require.NoError(t, err)
		var o identity.CredentialsPassword
		require.NoError(t, json.NewDecoder(bytes.NewBuffer(c.Config)).Decode(&o))
		assert.True(t, reg.Hasher(ctx).Understands([]byte(o.HashedPassword)), "%s", o.HashedPassword)
		assert.True(t, hash.IsBcryptHash([]byte(o.HashedPassword)), "%s", o.HashedPassword)

		// retry after upgraded
		body = testhelpers.SubmitLoginForm(t, false, browserClient, publicTS, values,
			false, true, http.StatusOK, redirTS.URL)
		assert.Equal(t, identifier, gjson.Get(body, "identity.traits.subject").String(), "%s", body)
	})
}
