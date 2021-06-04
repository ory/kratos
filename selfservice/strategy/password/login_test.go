package password_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	kratos "github.com/ory/kratos-client-go"

	"github.com/ory/x/ioutilx"

	"github.com/ory/x/assertx"
	"github.com/ory/x/errorsx"
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

func TestCompleteLogin(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword),
		map[string]interface{}{"enabled": true})
	publicTS, _ := testhelpers.NewKratosServer(t, reg)

	errTS := testhelpers.NewErrorTestServer(t, reg)
	uiTS := testhelpers.NewLoginUIFlowEchoServer(t, reg)
	redirTS := newReturnTs(t, reg)

	// Overwrite these two:
	conf.MustSet(config.ViperKeySelfServiceErrorUI, errTS.URL+"/error-ts")
	conf.MustSet(config.ViperKeySelfServiceLoginUI, uiTS.URL+"/login-ts")

	conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/login.schema.json")
	conf.MustSet(config.ViperKeySecretsDefault, []string{"not-a-secure-session-key"})

	ensureFieldsExist := func(t *testing.T, body []byte) {
		checkFormContent(t, body, "password_identifier",
			"password",
			"csrf_token")
	}

	createIdentity := func(identifier, password string) {
		p, _ := reg.Hasher().Generate(context.Background(), []byte(password))
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &identity.Identity{
			ID:     x.NewUUID(),
			Traits: identity.Traits(fmt.Sprintf(`{"subject":"%s"}`, identifier)),
			Credentials: map[identity.CredentialsType]identity.Credentials{
				identity.CredentialsTypePassword: {
					Type:        identity.CredentialsTypePassword,
					Identifiers: []string{identifier},
					Config:      sqlxx.JSONRawMessage(`{"hashed_password":"` + string(p) + `"}`),
				},
			},
		}))
	}

	apiClient := testhelpers.NewDebugClient(t)

	t.Run("case=should show the error ui because the request payload is malformed", func(t *testing.T) {
		t.Run("type=api", func(t *testing.T) {
			f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false)

			body, res := testhelpers.LoginMakeRequest(t, true, f, apiClient, "14=)=!(%)$/ZP()GHIÖ")
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, body, `Expected JSON sent in request body to be an object but got: Number`)
		})

		t.Run("type=browser", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false)

			body, res := testhelpers.LoginMakeRequest(t, false, f, browserClient, "14=)=!(%)$/ZP()GHIÖ")
			assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/login-ts")
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "invalid URL escape", "%s", body)
		})
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
			actual, res := testhelpers.LoginMakeRequest(t, true, fakeFlow, apiClient, "{}")
			assert.Len(t, res.Cookies(), 0)
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
			check(t, gjson.Get(actual, "error").Raw)
		})

		t.Run("type=browser", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			actual, res := testhelpers.LoginMakeRequest(t, false, fakeFlow, browserClient, "")
			assert.Contains(t, res.Request.URL.String(), errTS.URL)
			check(t, gjson.Get(actual, "0").Raw)
		})
	})

	t.Run("case=should return an error because the request is expired", func(t *testing.T) {
		conf.MustSet(config.ViperKeySelfServiceLoginRequestLifespan, "50ms")
		t.Cleanup(func() {
			conf.MustSet(config.ViperKeySelfServiceLoginRequestLifespan, "10m")
		})
		values := url.Values{
			"csrf_token":          {x.FakeCSRFToken},
			"password_identifier": {"identifier"},
			"password":            {"password"},
		}

		t.Run("type=api", func(t *testing.T) {
			f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false)

			time.Sleep(time.Millisecond * 60)
			actual, res := testhelpers.LoginMakeRequest(t, true, f, apiClient, testhelpers.EncodeFormAsJSON(t, true, values))
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteGetFlow)
			assert.NotEqual(t, f.Id, gjson.Get(actual, "id").String(), "%s", actual)
			assert.Contains(t, gjson.Get(actual, "ui.messages.0.text").String(), "expired", "%s", actual)
		})

		t.Run("type=browser", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false)

			time.Sleep(time.Millisecond * 60)
			actual, res := testhelpers.LoginMakeRequest(t, false, f, browserClient, values.Encode())
			assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/login-ts")
			assert.NotEqual(t, f.Id, gjson.Get(actual, "id").String(), "%s", actual)
			assert.Contains(t, gjson.Get(actual, "ui.messages.0.text").String(), "expired", "%s", actual)
		})
	})

	t.Run("case=should have correct CSRF behavior", func(t *testing.T) {
		var values = url.Values{
			"method":              {"password"},
			"csrf_token":          {"invalid_token"},
			"password_identifier": {"login-identifier-csrf-browser"},
			"password":            {x.NewUUID().String()},
		}

		t.Run("case=should fail because of missing CSRF token/type=browser", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false)

			actual, res := testhelpers.LoginMakeRequest(t, false, f, browserClient, values.Encode())
			assert.EqualValues(t, http.StatusOK, res.StatusCode)
			assertx.EqualAsJSON(t, x.ErrInvalidCSRFToken,
				json.RawMessage(gjson.Get(actual, "0").Raw), "%s", actual)
		})

		t.Run("case=should pass even without CSRF token/type=api", func(t *testing.T) {
			f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false)

			actual, res := testhelpers.LoginMakeRequest(t, true, f, apiClient, testhelpers.EncodeFormAsJSON(t, true, values))
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

	var expectValidationError = func(t *testing.T, isAPI, forced bool, values func(url.Values)) string {
		return testhelpers.SubmitLoginForm(t, isAPI, nil, publicTS, values,
			identity.CredentialsTypePassword, forced,
			testhelpers.ExpectStatusCode(isAPI, http.StatusBadRequest, http.StatusOK),
			testhelpers.ExpectURL(isAPI, publicTS.URL+login.RouteSubmitFlow, conf.SelfServiceFlowLoginUI().String()))
	}

	t.Run("should return an error because the credentials are invalid (user does not exist)", func(t *testing.T) {
		var check = func(t *testing.T, body string, start time.Time) {
			delay := time.Since(start)
			minConfiguredDelay := conf.HasherArgon2().ExpectedDuration - conf.HasherArgon2().ExpectedDeviation
			assert.GreaterOrEqual(t, delay, minConfiguredDelay)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.action").String(), publicTS.URL+login.RouteSubmitFlow, "%s", body)
			assert.Equal(t, text.NewErrorValidationInvalidCredentials().Text, gjson.Get(body, "ui.messages.0.text").String(), body)
		}

		var values = func(v url.Values) {
			v.Set("password_identifier", "identifier")
			v.Set("password", "password")
		}

		t.Run("type=browser", func(t *testing.T) {
			start := time.Now()
			check(t, expectValidationError(t, false, false, values), start)
		})

		t.Run("type=api", func(t *testing.T) {
			start := time.Now()
			check(t, expectValidationError(t, true, false, values), start)
		})
	})

	t.Run("should return an error because no identifier is set", func(t *testing.T) {
		var check = func(t *testing.T, body string) {
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.action").String(), publicTS.URL+login.RouteSubmitFlow, "%s", body)

			ensureFieldsExist(t, []byte(body))
			assert.Equal(t, "Property password_identifier is missing.", gjson.Get(body, "ui.nodes.#(attributes.name==password_identifier).messages.0.text").String(), "%s", body)
			assert.Len(t, gjson.Get(body, "ui.nodes").Array(), 4)

			// The password value should not be returned!
			assert.Empty(t, gjson.Get(body, "ui.nodes.#(attributes.name==password).attributes.value").String())
		}

		var values = func(v url.Values) {
			v.Del("password_identifier")
			v.Set("password", "password")
		}

		t.Run("type=browser", func(t *testing.T) {
			check(t, expectValidationError(t, false, false, values))
		})

		t.Run("type=api", func(t *testing.T) {
			check(t, expectValidationError(t, true, false, values))
		})
	})

	t.Run("should return an error because no password is set", func(t *testing.T) {
		var check = func(t *testing.T, body string) {
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.action").String(), publicTS.URL+login.RouteSubmitFlow, "%s", body)

			ensureFieldsExist(t, []byte(body))
			assert.Equal(t, "Property password is missing.", gjson.Get(body, "ui.nodes.#(attributes.name==password).messages.0.text").String(), "%s", body)
			assert.Equal(t, "identifier", gjson.Get(body, "ui.nodes.#(attributes.name==password_identifier).attributes.value").String(), "%s", body)
			assert.Len(t, gjson.Get(body, "ui.nodes").Array(), 4)

			// This must not include the password!
			assert.Empty(t, gjson.Get(body, "ui.nodes.#(attributes.name==password).attributes.value").String())
		}

		var values = func(v url.Values) {
			v.Set("password_identifier", "identifier")
			v.Del("password")
		}

		t.Run("type=browser", func(t *testing.T) {
			check(t, expectValidationError(t, false, false, values))
		})

		t.Run("type=api", func(t *testing.T) {
			check(t, expectValidationError(t, true, false, values))
		})
	})

	t.Run("should return an error both identifier and password are missing", func(t *testing.T) {
		var check = func(t *testing.T, body string) {
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Contains(t, gjson.Get(body, "ui.action").String(), publicTS.URL+login.RouteSubmitFlow, "%s", body)

			ensureFieldsExist(t, []byte(body))
			assert.Equal(t, "length must be >= 1, but got 0", gjson.Get(body, "ui.nodes.#(attributes.name==password).messages.0.text").String(), "%s", body)
			assert.Equal(t, "length must be >= 1, but got 0", gjson.Get(body, "ui.nodes.#(attributes.name==password_identifier).messages.0.text").String(), "%s", body)
			assert.Len(t, gjson.Get(body, "ui.nodes").Array(), 4)

			// This must not include the password!
			assert.Empty(t, gjson.Get(body, "ui.nodes.#(attributes.name==password).attributes.value").String())
		}

		var values = func(v url.Values) {
			v.Set("password", "")
			v.Set("password_identifier", "")
		}

		t.Run("type=browser", func(t *testing.T) {
			check(t, expectValidationError(t, false, false, values))
		})

		t.Run("type=api", func(t *testing.T) {
			check(t, expectValidationError(t, true, false, values))
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
		createIdentity(identifier, pwd)

		var values = func(v url.Values) {
			v.Set("password_identifier", identifier)
			v.Set("password", "not-password")
		}

		t.Run("type=browser", func(t *testing.T) {
			check(t, expectValidationError(t, false, false, values))
		})

		t.Run("type=api", func(t *testing.T) {
			check(t, expectValidationError(t, true, false, values))
		})
	})

	t.Run("should pass with real request", func(t *testing.T) {
		identifier, pwd := x.NewUUID().String(), "password"
		createIdentity(identifier, pwd)

		var values = func(v url.Values) {
			v.Set("password_identifier", identifier)
			v.Set("password", pwd)
		}

		t.Run("type=browser", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)

			body := testhelpers.SubmitLoginForm(t, false, browserClient, publicTS, values,
				identity.CredentialsTypePassword, false, http.StatusOK, redirTS.URL)

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

					body, err := ioutil.ReadAll(res.Body)
					require.NoError(t, err)
					assert.True(t, gjson.GetBytes(body, "forced").Bool())
					assert.Equal(t, identifier, gjson.GetBytes(body, "ui.nodes.#(attributes.name==password_identifier).attributes.value").String(), "%s", body)
					assert.Empty(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==password).attributes.value").String(), "%s", body)
				})
			})
		})

		t.Run("type=api", func(t *testing.T) {
			body := testhelpers.SubmitLoginForm(t, true, nil, publicTS, values,
				identity.CredentialsTypePassword, false, http.StatusOK, publicTS.URL+login.RouteSubmitFlow)

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

					assert.True(t, gjson.GetBytes(body, "forced").Bool())
					assert.Equal(t, identifier, gjson.GetBytes(body, "ui.nodes.#(attributes.name==password_identifier).attributes.value").String(), "%s", body)
					assert.Empty(t, gjson.GetBytes(body, "ui.nodes.#(attributes.name==password).attributes.value").String(), "%s", body)
				})
			})
		})
	})

	t.Run("case=should return an error because not passing validation and reset previous errors and values", func(t *testing.T) {
		conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/login.schema.json")

		var check = func(t *testing.T, actual string) {
			assert.NotEmpty(t, gjson.Get(actual, "id").String(), "%s", actual)
			assert.Contains(t, gjson.Get(actual, "ui.action").String(), publicTS.URL+login.RouteSubmitFlow, "%s", actual)
		}

		var checkFirst = func(t *testing.T, actual string) {
			check(t, actual)
			assert.Contains(t, gjson.Get(actual, "ui.nodes.#(attributes.name==password_identifier).messages.0").String(), "Property password_identifier is missing.", "%s", actual)
		}

		var checkSecond = func(t *testing.T, actual string) {
			check(t, actual)

			assert.Empty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==password_identifier).attributes.error"))
			assert.EqualValues(t, "identifier", gjson.Get(actual, "ui.nodes.#(attributes.name==password_identifier).attributes.value").String(), actual)
			assert.EqualValues(t, "password", gjson.Get(actual, "ui.nodes.#(attributes.name==method).attributes.value").String(), actual)
			assert.Empty(t, gjson.Get(actual, "ui.error"))
			assert.Contains(t, gjson.Get(actual, "ui.nodes.#(attributes.name==password).messages.0").String(), "Property password is missing.", "%s", actual)
		}

		var valuesFirst = func(v url.Values) url.Values {
			v.Del("password_identifier")
			v.Set("password", x.NewUUID().String())
			return v
		}

		var valuesSecond = func(v url.Values) url.Values {
			v.Set("password_identifier", "identifier")
			v.Del("password")
			return v
		}

		t.Run("type=api", func(t *testing.T) {
			f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false)

			actual, _ := testhelpers.LoginMakeRequest(t, true, f, apiClient, testhelpers.EncodeFormAsJSON(t, true, valuesFirst(testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes))))
			checkFirst(t, actual)
			actual, _ = testhelpers.LoginMakeRequest(t, true, f, apiClient, testhelpers.EncodeFormAsJSON(t, true, valuesSecond(testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes))))
			checkSecond(t, actual)
		})

		t.Run("type=browser", func(t *testing.T) {
			browserClient := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false)

			actual, _ := testhelpers.LoginMakeRequest(t, false, f, browserClient, valuesFirst(testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)).Encode())
			checkFirst(t, actual)
			actual, _ = testhelpers.LoginMakeRequest(t, false, f, browserClient, valuesSecond(testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)).Encode())
			checkSecond(t, actual)
		})
	})

	t.Run("should be a new session with forced flag", func(t *testing.T) {
		identifier, pwd := x.NewUUID().String(), "password"
		createIdentity(identifier, pwd)

		browserClient := testhelpers.NewClientWithCookies(t)
		f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false)

		values := url.Values{"method": {"password"}, "password_identifier": {identifier},
			"password": {pwd}, "csrf_token": {x.FakeCSRFToken}}.Encode()

		body1, res := testhelpers.LoginMakeRequest(t, false, f, browserClient, values)
		assert.EqualValues(t, http.StatusOK, res.StatusCode)

		f = testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, true)
		body2, res := testhelpers.LoginMakeRequest(t, false, f, browserClient, values)

		require.Contains(t, res.Request.URL.Path, "return-ts", "%s", res.Request.URL.String())
		assert.Equal(t, identifier, gjson.Get(body2, "identity.traits.subject").String(), "%s", body2)
		assert.NotEqual(t, gjson.Get(body1, "id").String(), gjson.Get(body2, "id").String(), "%s\n\n%s\n", body1, body2)
	})

	t.Run("should login same identity regardless of identifier capitalization", func(t *testing.T) {
		identifier, pwd := x.NewUUID().String(), "password"
		createIdentity(identifier, pwd)

		browserClient := testhelpers.NewClientWithCookies(t)
		f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false)

		values := url.Values{"method": {"password"}, "password_identifier": {strings.ToUpper(identifier)}, "password": {pwd}, "csrf_token": {x.FakeCSRFToken}}.Encode()

		_, res := testhelpers.LoginMakeRequest(t, false, f, browserClient, values)
		assert.EqualValues(t, http.StatusOK, res.StatusCode)

		f = testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, true)
		body2, res := testhelpers.LoginMakeRequest(t, false, f, browserClient, values)

		assert.Equal(t, identifier, gjson.Get(body2, "identity.traits.subject").String(), "%s", body2)
	})
}
