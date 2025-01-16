// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package password_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ory/kratos/selfservice/strategy/idfirst"

	configtesthelpers "github.com/ory/kratos/driver/config/testhelpers"

	"github.com/ory/x/randx"
	"github.com/ory/x/snapshotx"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/internal/registrationhelpers"

	"github.com/ory/kratos/selfservice/flow"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hash"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	kratos "github.com/ory/kratos/internal/httpclient"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
	"github.com/ory/x/assertx"
	"github.com/ory/x/errorsx"
	"github.com/ory/x/ioutilx"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"
)

//go:embed stub/login.schema.json
var loginSchema []byte

func createIdentity(ctx context.Context, reg *driver.RegistryDefault, t *testing.T, identifier, password string) *identity.Identity {
	p, _ := reg.Hasher(ctx).Generate(context.Background(), []byte(password))
	iId := x.NewUUID()
	id := &identity.Identity{
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
	}
	require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(ctx, id))
	return id
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

	testhelpers.SetIdentitySchemas(t, conf, map[string]string{
		"migration": "file://./stub/migration.schema.json",
		"default":   "file://./stub/login.schema.json",
	})

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
		check := func(t *testing.T, actual string) {
			assert.Equal(t, int64(http.StatusNotFound), gjson.Get(actual, "code").Int(), "%s", actual)
			assert.Equal(t, "Not Found", gjson.Get(actual, "status").String(), "%s", actual)
			assert.Contains(t, gjson.Get(actual, "message").String(), "Unable to locate the resource", "%s", actual)
		}

		fakeFlow := &kratos.LoginFlow{
			Ui: kratos.UiContainer{
				Action: publicTS.URL + login.RouteSubmitFlow + "?flow=" + x.NewUUID().String(),
			},
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
		values := url.Values{
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

	expectValidationError := func(t *testing.T, isAPI, refresh, isSPA bool, values func(url.Values)) string {
		return testhelpers.SubmitLoginForm(t, isAPI, nil, publicTS, values,
			isSPA, refresh,
			testhelpers.ExpectStatusCode(isAPI || isSPA, http.StatusBadRequest, http.StatusOK),
			testhelpers.ExpectURL(isAPI || isSPA, publicTS.URL+login.RouteSubmitFlow, conf.SelfServiceFlowLoginUI(ctx).String()))
	}

	t.Run("should return an error because the credentials are invalid (user does not exist)", func(t *testing.T) {
		check := func(t *testing.T, body string, start time.Time) {
			delay := time.Since(start)
			minConfiguredDelay := conf.HasherArgon2(ctx).ExpectedDuration - conf.HasherArgon2(ctx).ExpectedDeviation
			assert.GreaterOrEqual(t, delay, minConfiguredDelay)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.action").String(), publicTS.URL+login.RouteSubmitFlow, "%s", body)
			assert.Equal(t, text.NewErrorValidationInvalidCredentials().Text, gjson.Get(body, "ui.messages.0.text").String(), body)
		}

		values := func(v url.Values) {
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
		check := func(t *testing.T, body string) {
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.action").String(), publicTS.URL+login.RouteSubmitFlow, "%s", body)

			ensureFieldsExist(t, []byte(body))
			assert.Equal(t, "Property identifier is missing.", gjson.Get(body, "ui.nodes.#(attributes.name==identifier).messages.0.text").String(), "%s", body)
			assert.Len(t, gjson.Get(body, "ui.nodes").Array(), 4)

			// The password value should not be returned!
			assert.Empty(t, gjson.Get(body, "ui.nodes.#(attributes.name==password).attributes.value").String())
		}

		values := func(v url.Values) {
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
		check := func(t *testing.T, body string) {
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.action").String(), publicTS.URL+login.RouteSubmitFlow, "%s", body)

			ensureFieldsExist(t, []byte(body))
			assert.Equal(t, "Property password is missing.", gjson.Get(body, "ui.nodes.#(attributes.name==password).messages.0.text").String(), "%s", body)
			assert.Equal(t, "identifier", gjson.Get(body, "ui.nodes.#(attributes.name==identifier).attributes.value").String(), "%s", body)
			assert.Len(t, gjson.Get(body, "ui.nodes").Array(), 4)

			// This must not include the password!
			assert.Empty(t, gjson.Get(body, "ui.nodes.#(attributes.name==password).attributes.value").String())
		}

		values := func(v url.Values) {
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
		check := func(t *testing.T, body string) {
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.action").String(), publicTS.URL+login.RouteSubmitFlow, "%s", body)

			ensureFieldsExist(t, []byte(body))
			assert.Equal(t, "Property password is missing.", gjson.Get(body, "ui.nodes.#(attributes.name==password).messages.0.text").String(), "%s", body)
			assert.Equal(t, "Property identifier is missing.", gjson.Get(body, "ui.nodes.#(attributes.name==identifier).messages.0.text").String(), "%s", body)
			assert.Len(t, gjson.Get(body, "ui.nodes").Array(), 4)

			// This must not include the password!
			assert.Empty(t, gjson.Get(body, "ui.nodes.#(attributes.name==password).attributes.value").String())
		}

		values := func(v url.Values) {
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
		check := func(t *testing.T, body string) {
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

		values := func(v url.Values) {
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

		values := func(v url.Values) {
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
				id := identity.NewIdentity("default")
				id.NID = x.NewUUID()
				browserClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, ctx, reg, id)

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
				id := identity.NewIdentity("default")
				id.NID = x.NewUUID()
				hc := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, ctx, reg, id)

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
				c := &http.Client{Transport: testhelpers.NewTransportWithHeader(t, http.Header{"Authorization": {"Bearer " + st}})}

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
				id := identity.NewIdentity("default")
				id.NID = x.NewUUID()
				hc := testhelpers.NewHTTPClientWithIdentitySessionToken(t, ctx, reg, id)

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
		check := func(t *testing.T, actual string) {
			assert.NotEmpty(t, gjson.Get(actual, "id").String(), "%s", actual)
			assert.Contains(t, gjson.Get(actual, "ui.action").String(), publicTS.URL+login.RouteSubmitFlow, "%s", actual)
		}

		checkFirst := func(t *testing.T, actual string) {
			check(t, actual)
			assert.Contains(t, gjson.Get(actual, "ui.nodes.#(attributes.name==identifier).messages.0").String(), "Property identifier is missing.", "%s", actual)
		}

		checkSecond := func(t *testing.T, actual string) {
			check(t, actual)

			assert.Empty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==identifier).attributes.error"))
			assert.EqualValues(t, "identifier", gjson.Get(actual, "ui.nodes.#(attributes.name==identifier).attributes.value").String(), actual)
			assert.EqualValues(t, "password", gjson.Get(actual, "ui.nodes.#(attributes.name==method).attributes.value").String(), actual)
			assert.Empty(t, gjson.Get(actual, "ui.error"))
			assert.Contains(t, gjson.Get(actual, "ui.nodes.#(attributes.name==password).messages.0").String(), "Property password is missing.", "%s", actual)
		}

		valuesFirst := func(v url.Values) url.Values {
			v.Del("identifier")
			v.Set("password", x.NewUUID().String())
			return v
		}

		valuesSecond := func(v url.Values) url.Values {
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

		values := url.Values{
			"method": {"password"}, "identifier": {identifier},
			"password": {pwd}, "csrf_token": {x.FakeCSRFToken},
		}.Encode()

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

	t.Run("should succeed and include redirect continue_with in SPA flow", func(t *testing.T) {
		identifier, pwd := x.NewUUID().String(), "password"
		createIdentity(ctx, reg, t, identifier, pwd)

		browserClient := testhelpers.NewClientWithCookies(t)
		f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, true, false, false)
		values := url.Values{"method": {"password"}, "identifier": {strings.ToUpper(identifier)}, "password": {pwd}, "csrf_token": {x.FakeCSRFToken}}.Encode()
		body, res := testhelpers.LoginMakeRequest(t, false, true, f, browserClient, values)

		assert.EqualValues(t, http.StatusOK, res.StatusCode)
		assert.EqualValues(t, flow.ContinueWithActionRedirectBrowserToString, gjson.Get(body, "continue_with.0.action").String(), "%s", body)
		assert.EqualValues(t, conf.SelfServiceBrowserDefaultReturnTo(ctx).String(), gjson.Get(body, "continue_with.0.redirect_browser_to").String(), "%s", body)
	})

	t.Run("should succeed and not have redirect continue_with in api flow", func(t *testing.T) {
		identifier, pwd := x.NewUUID().String(), "password"
		createIdentity(ctx, reg, t, identifier, pwd)
		browserClient := testhelpers.NewClientWithCookies(t)
		f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false)

		body, res := testhelpers.LoginMakeRequest(t, true, true, f, browserClient, fmt.Sprintf(`{"method":"password","identifier":"%s","password":"%s"}`, strings.ToUpper(identifier), pwd))

		assert.EqualValues(t, http.StatusOK, res.StatusCode, body)
		assert.Empty(t, gjson.Get(body, "continue_with").Array(), "%s", body)
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

		values := func(v url.Values) {
			v.Set("method", "password")
			v.Set("identifier", identifier)
			v.Set("password", pwd)
		}

		check := func(t *testing.T, body string) {
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
		identifier, pwd := x.NewUUID().String()+"@google.com", "password"
		h := &hash.Pbkdf2{
			Algorithm:  "sha256",
			Iterations: 100000,
			SaltLength: 32,
			KeyLength:  32,
		}
		p, _ := h.Generate(context.Background(), []byte(pwd))

		iId := x.NewUUID()
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &identity.Identity{
			ID:       iId,
			SchemaID: "migration",
			Traits:   identity.Traits(fmt.Sprintf(`{"email":"%s"}`, identifier)),
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

		values := func(v url.Values) {
			v.Set("identifier", identifier)
			v.Set("method", identity.CredentialsTypePassword.String())
			v.Set("password", pwd)
		}

		browserClient := testhelpers.NewClientWithCookies(t)

		body := testhelpers.SubmitLoginForm(t, false, browserClient, publicTS, values,
			false, false, http.StatusOK, redirTS.URL)

		assert.Equal(t, identifier, gjson.Get(body, "identity.traits.email").String(), "%s", body)

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
		assert.Equal(t, identifier, gjson.Get(body, "identity.traits.email").String(), "%s", body)
	})

	t.Run("suite=password rehashing degrades gracefully during login", func(t *testing.T) {
		identifier := x.NewUUID().String() + "@google.com"
		// pwd := "Kd9hUV4Xkcq87VSca6A4fq1iBijrMScBFhkpIPEwBtvTDsBwfqJCqXPPr4TkhOhsd9wFGeB3MzS4bJuesLCAjJc5s1GKJ51zW7F"
		pwd := randx.MustString(100, randx.AlphaNum) // longer than bcrypt max length
		require.Greater(t, len(pwd), 72)             // bcrypt max length
		salt := randx.MustString(32, randx.AlphaNum)
		sha := sha256.Sum256([]byte(pwd + salt))
		hashed := "{SSHA256}" + base64.StdEncoding.EncodeToString(slices.Concat(sha[:], []byte(salt)))
		iId := x.NewUUID()
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &identity.Identity{
			ID:       iId,
			SchemaID: "migration",
			Traits:   identity.Traits(fmt.Sprintf(`{"email":%q}`, identifier)),
			Credentials: map[identity.CredentialsType]identity.Credentials{
				identity.CredentialsTypePassword: {
					Type:        identity.CredentialsTypePassword,
					Identifiers: []string{identifier},
					Config:      sqlxx.JSONRawMessage(`{"hashed_password":"` + hashed + `"}`),
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

		values := func(v url.Values) {
			v.Set("identifier", identifier)
			v.Set("method", identity.CredentialsTypePassword.String())
			v.Set("password", pwd)
		}

		browserClient := testhelpers.NewClientWithCookies(t)

		body := testhelpers.SubmitLoginForm(t, false, browserClient, publicTS, values,
			false, false, http.StatusOK, redirTS.URL)

		assert.Equal(t, identifier, gjson.Get(body, "identity.traits.email").String(), "%s", body)

		// check that the password hash algorithm is unchanged
		_, c, err := reg.PrivilegedIdentityPool().FindByCredentialsIdentifier(context.Background(), identity.CredentialsTypePassword, identifier)
		require.NoError(t, err)
		var o identity.CredentialsPassword
		require.NoError(t, json.NewDecoder(bytes.NewBuffer(c.Config)).Decode(&o))
		assert.Equal(t, hashed, o.HashedPassword)

		// login still works
		body = testhelpers.SubmitLoginForm(t, false, browserClient, publicTS, values,
			false, true, http.StatusOK, redirTS.URL)
		assert.Equal(t, identifier, gjson.Get(body, "identity.traits.email").String(), "%s", body)
	})

	t.Run("suite=password migration hook", func(t *testing.T) {
		ctx := context.Background()

		type (
			hookPayload = struct {
				Identifier string `json:"identifier"`
				Password   string `json:"password"`
			}
			tsRequestHandler = func(hookPayload) (status int, body string)
		)
		returnStatus := func(status int) func(string, string) tsRequestHandler {
			return func(string, string) tsRequestHandler {
				return func(hookPayload) (int, string) { return status, "" }
			}
		}
		returnStatic := func(status int, body string) func(string, string) tsRequestHandler {
			return func(string, string) tsRequestHandler {
				return func(hookPayload) (int, string) { return status, body }
			}
		}

		// each test case sends (number of expected calls) handlers to the channel, at a max of 3
		tsChan := make(chan tsRequestHandler, 3)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			_ = r.Body.Close()
			var payload hookPayload
			require.NoError(t, json.Unmarshal(b, &payload))

			select {
			case handlerFn := <-tsChan:
				status, body := handlerFn(payload)
				w.WriteHeader(status)
				_, _ = io.WriteString(w, body)

			default:
				t.Fatal("unexpected call to the password migration hook")
			}
		}))
		t.Cleanup(ts.Close)

		require.NoError(t, reg.Config().Set(ctx, config.ViperKeyPasswordMigrationHook, map[string]any{
			"config":  map[string]any{"url": ts.URL},
			"enabled": true,
		}))

		for _, tc := range []struct {
			name              string
			hookHandler       func(identifier, password string) tsRequestHandler
			expectHookCalls   int
			setupFn           func() func()
			credentialsConfig string
			expectSuccess     bool
		}{{
			name:              "should call migration hook",
			credentialsConfig: `{"use_password_migration_hook": true}`,
			hookHandler: func(identifier, password string) tsRequestHandler {
				return func(payload hookPayload) (status int, body string) {
					if payload.Identifier == identifier && payload.Password == password {
						return http.StatusOK, `{"status":"password_match"}`
					} else {
						return http.StatusOK, `{"status":"no_match"}`
					}
				}
			},
			expectHookCalls: 1,
			expectSuccess:   true,
		}, {
			name:              "should not update identity when the password is wrong",
			credentialsConfig: `{"use_password_migration_hook": true}`,
			hookHandler:       returnStatus(http.StatusForbidden),
			expectHookCalls:   1,
			expectSuccess:     false,
		}, {
			name:              "should inspect response",
			credentialsConfig: `{"use_password_migration_hook": true}`,
			hookHandler:       returnStatic(http.StatusOK, `{"status":"password_no_match"}`),
			expectHookCalls:   1,
			expectSuccess:     false,
		}, {
			name:              "should not update identity when the migration hook returns 200 without JSON",
			credentialsConfig: `{"use_password_migration_hook": true}`,
			hookHandler:       returnStatus(http.StatusOK),
			expectHookCalls:   1,
			expectSuccess:     false,
		}, {
			name:              "should not update identity when the migration hook returns 500",
			credentialsConfig: `{"use_password_migration_hook": true}`,
			hookHandler:       returnStatus(http.StatusInternalServerError),
			expectHookCalls:   3, // expect retries on 500
			expectSuccess:     false,
		}, {
			name:              "should not update identity when the migration hook returns 201",
			credentialsConfig: `{"use_password_migration_hook": true}`,
			hookHandler:       returnStatic(http.StatusCreated, `{"status":"password_match"}`),
			expectHookCalls:   1,
			expectSuccess:     false,
		}, {
			name:              "should not update identity and not call hook when hash is set",
			credentialsConfig: `{"use_password_migration_hook": true, "hashed_password":"hash"}`,
			expectSuccess:     false,
		}, {
			name:              "should not update identity and not call hook when use_password_migration_hook is not set",
			credentialsConfig: `{"hashed_password":"hash"}`,
			expectSuccess:     false,
		}, {
			name:              "should not update identity and not call hook when credential is empty",
			credentialsConfig: `{}`,
			expectSuccess:     false,
		}, {
			name:              "should not call migration hook if disabled",
			credentialsConfig: `{"use_password_migration_hook": true}`,
			setupFn: func() func() {
				require.NoError(t, reg.Config().Set(ctx, config.ViperKeyPasswordMigrationHook+".enabled", false))
				return func() {
					require.NoError(t, reg.Config().Set(ctx, config.ViperKeyPasswordMigrationHook+".enabled", true))
				}
			},
			expectSuccess: false,
		}} {
			t.Run("case="+tc.name, func(t *testing.T) {
				if tc.setupFn != nil {
					cleanup := tc.setupFn()
					t.Cleanup(cleanup)
				}

				identifier := x.NewUUID().String() + "@google.com"
				password := x.NewUUID().String()
				iId := x.NewUUID()
				require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(ctx, &identity.Identity{
					ID:       iId,
					SchemaID: "migration",
					Traits:   identity.Traits(fmt.Sprintf(`{"email":"%s"}`, identifier)),
					Credentials: map[identity.CredentialsType]identity.Credentials{
						identity.CredentialsTypePassword: {
							Type:        identity.CredentialsTypePassword,
							Identifiers: []string{identifier},
							Config:      sqlxx.JSONRawMessage(tc.credentialsConfig),
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

				values := func(v url.Values) {
					v.Set("identifier", identifier)
					v.Set("method", identity.CredentialsTypePassword.String())
					v.Set("password", password)
				}

				for range tc.expectHookCalls {
					tsChan <- tc.hookHandler(identifier, password)
				}

				browserClient := testhelpers.NewClientWithCookies(t)

				if tc.expectSuccess {
					body := testhelpers.SubmitLoginForm(t, false, browserClient, publicTS, values,
						false, false, http.StatusOK, redirTS.URL)
					assert.Equal(t, identifier, gjson.Get(body, "identity.traits.email").String(), "%s", body)

					// check if password hash algorithm is upgraded
					_, c, err := reg.PrivilegedIdentityPool().FindByCredentialsIdentifier(ctx, identity.CredentialsTypePassword, identifier)
					require.NoError(t, err)
					var o identity.CredentialsPassword
					require.NoError(t, json.NewDecoder(bytes.NewBuffer(c.Config)).Decode(&o))
					assert.True(t, reg.Hasher(ctx).Understands([]byte(o.HashedPassword)), "%s", o.HashedPassword)
					assert.True(t, hash.IsBcryptHash([]byte(o.HashedPassword)), "%s", o.HashedPassword)

					// retry after upgraded
					body = testhelpers.SubmitLoginForm(t, false, browserClient, publicTS, values,
						false, true, http.StatusOK, redirTS.URL)
					assert.Equal(t, identifier, gjson.Get(body, "identity.traits.email").String(), "%s", body)
				} else {
					body := testhelpers.SubmitLoginForm(t, false, browserClient, publicTS, values,
						false, false, http.StatusOK, "")
					assert.Empty(t, gjson.Get(body, "identity.traits.subject").String(), "%s", body)
					// Check that the config did not change
					_, c, err := reg.PrivilegedIdentityPool().FindByCredentialsIdentifier(context.Background(), identity.CredentialsTypePassword, identifier)
					require.NoError(t, err)
					assert.JSONEq(t, tc.credentialsConfig, string(c.Config))
				}

				// expect all hook calls to be done
				select {
				case <-tsChan:
					t.Fatal("the test unexpectedly did too few calls to the password hook")
				default:
					// pass
				}
			})
		}
	})
}

func TestFormHydration(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	ctx = configtesthelpers.WithConfigValue(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword), map[string]interface{}{"enabled": true})
	ctx = testhelpers.WithDefaultIdentitySchemaFromRaw(ctx, loginSchema)

	s, err := reg.AllLoginStrategies().Strategy(identity.CredentialsTypePassword)
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

	t.Run("method=PopulateLoginMethodSecondFactor", func(t *testing.T) {
		r, f := newFlow(ctx, t)
		f.RequestedAAL = identity.AuthenticatorAssuranceLevel2
		require.NoError(t, fh.PopulateLoginMethodSecondFactor(r, f))
		toSnapshot(t, f)
	})

	t.Run("method=PopulateLoginMethodFirstFactor", func(t *testing.T) {
		r, f := newFlow(ctx, t)
		require.NoError(t, fh.PopulateLoginMethodFirstFactor(r, f))
		toSnapshot(t, f)
	})

	t.Run("method=PopulateLoginMethodFirstFactorRefresh", func(t *testing.T) {
		r, f := newFlow(ctx, t)
		id := createIdentity(ctx, reg, t, "some@user.com", "password")
		r.Header = testhelpers.NewHTTPClientWithIdentitySessionToken(t, ctx, reg, id).Transport.(*testhelpers.TransportWithHeader).GetHeader()
		f.Refresh = true
		require.NoError(t, fh.PopulateLoginMethodFirstFactorRefresh(r, f))
		toSnapshot(t, f)
	})

	t.Run("method=PopulateLoginMethodSecondFactorRefresh", func(t *testing.T) {
		r, f := newFlow(ctx, t)
		require.NoError(t, fh.PopulateLoginMethodSecondFactorRefresh(r, f))
		toSnapshot(t, f)
	})

	t.Run("method=PopulateLoginMethodIdentifierFirstCredentials", func(t *testing.T) {
		t.Run("case=no options", func(t *testing.T) {
			t.Run("case=account enumeration mitigation disabled", func(t *testing.T) {
				ctx := configtesthelpers.WithConfigValue(ctx, config.ViperKeySecurityAccountEnumerationMitigate, false)
				r, f := newFlow(ctx, t)
				require.ErrorIs(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f), idfirst.ErrNoCredentialsFound)
				toSnapshot(t, f)
			})

			t.Run("case=account enumeration mitigation enabled", func(t *testing.T) {
				ctx := configtesthelpers.WithConfigValue(ctx, config.ViperKeySecurityAccountEnumerationMitigate, true)
				r, f := newFlow(ctx, t)
				require.ErrorIs(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f), idfirst.ErrNoCredentialsFound)
				toSnapshot(t, f)
			})
		})

		t.Run("case=WithIdentifier", func(t *testing.T) {
			t.Run("case=account enumeration mitigation disabled", func(t *testing.T) {
				ctx := configtesthelpers.WithConfigValue(ctx, config.ViperKeySecurityAccountEnumerationMitigate, false)
				r, f := newFlow(ctx, t)
				require.ErrorIs(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f, login.WithIdentifier("foo@bar.com")), idfirst.ErrNoCredentialsFound)
				toSnapshot(t, f)
			})

			t.Run("case=account enumeration mitigation enabled", func(t *testing.T) {
				ctx := configtesthelpers.WithConfigValue(ctx, config.ViperKeySecurityAccountEnumerationMitigate, true)
				r, f := newFlow(ctx, t)
				require.ErrorIs(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f, login.WithIdentifier("foo@bar.com")), idfirst.ErrNoCredentialsFound)
				toSnapshot(t, f)
			})
		})

		t.Run("case=WithIdentityHint", func(t *testing.T) {
			t.Run("case=account enumeration mitigation enabled and identity has no password", func(t *testing.T) {
				ctx := configtesthelpers.WithConfigValue(ctx, config.ViperKeySecurityAccountEnumerationMitigate, true)

				id := identity.NewIdentity("default")
				r, f := newFlow(ctx, t)
				require.ErrorIs(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f, login.WithIdentityHint(id)), idfirst.ErrNoCredentialsFound)
				toSnapshot(t, f)
			})

			t.Run("case=account enumeration mitigation disabled", func(t *testing.T) {
				ctx := configtesthelpers.WithConfigValue(ctx, config.ViperKeySecurityAccountEnumerationMitigate, false)

				t.Run("case=identity has password", func(t *testing.T) {
					identifier, pwd := x.NewUUID().String(), "password"
					id := createIdentity(ctx, reg, t, identifier, pwd)

					r, f := newFlow(ctx, t)
					require.NoError(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f, login.WithIdentityHint(id)))
					toSnapshot(t, f)
				})

				t.Run("case=identity does not have a password", func(t *testing.T) {
					id := identity.NewIdentity("default")
					r, f := newFlow(ctx, t)
					require.ErrorIs(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f, login.WithIdentityHint(id)), idfirst.ErrNoCredentialsFound)
					toSnapshot(t, f)
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
