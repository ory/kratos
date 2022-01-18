package password_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"

	"github.com/ory/kratos/selfservice/flow"

	kratos "github.com/ory/kratos-client-go"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/x/ioutilx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/assertx"
	"github.com/ory/x/httpx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/registration"

	"github.com/ory/kratos/x"
)

func checkFormContent(t *testing.T, body []byte, requiredFields ...string) {
	fieldNameSet(t, body, requiredFields)
	outdatedFieldsDoNotExist(t, body)
	formMethodIsPOST(t, body)
}

// fieldNameSet checks if the fields have the right "name" set.
func fieldNameSet(t *testing.T, body []byte, fields []string) {
	for _, f := range fields {
		assert.Equal(t, f, gjson.GetBytes(body, fmt.Sprintf("ui.nodes.#(attributes.name==%s).attributes.name", f)).String(), "%s", body)
	}
}

// checks if some keys are not set, this should be used to catch regression issues
func outdatedFieldsDoNotExist(t *testing.T, body []byte) {
	for _, k := range []string{"request"} {
		assert.Equal(t, false, gjson.GetBytes(body, fmt.Sprintf("ui.nodes.fields.#(name==%s)", k)).Exists())
	}
}

func formMethodIsPOST(t *testing.T, body []byte) {
	assert.Equal(t, "POST", gjson.GetBytes(body, "ui.method").String())
}

func TestRegistration(t *testing.T) {
	t.Run("case=registration", func(t *testing.T) {
		conf, reg := internal.NewFastRegistryWithMocks(t)

		router := x.NewRouterPublic()
		admin := x.NewRouterAdmin()
		conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword), map[string]interface{}{"enabled": true})

		publicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, admin)
		errTS := testhelpers.NewErrorTestServer(t, reg)
		uiTS := testhelpers.NewRegistrationUIFlowEchoServer(t, reg)
		redirTS := testhelpers.NewRedirSessionEchoTS(t, reg)
		redirNoSessionTS := testhelpers.NewRedirNoSessionTS(t, reg)

		// set the "return to" server, which will assert the session state
		// (redirTS: enforce that a session exists, redirNoSessionTS: enforce that no session exists)
		var useReturnToFromTS = func(ts *httptest.Server) {
			conf.MustSet(config.ViperKeySelfServiceBrowserDefaultReturnTo, ts.URL+"/default-return-to")
			conf.MustSet(config.ViperKeySelfServiceRegistrationAfter+"."+config.DefaultBrowserReturnURL, ts.URL+"/registration-return-ts")
		}

		useReturnToFromTS(redirTS)
		conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")

		apiClient := testhelpers.NewDebugClient(t)

		t.Run("description=can call endpoints only without session", func(t *testing.T) {
			// Needed to set up the mock IDs...
			conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/profile.schema.json")
			t.Cleanup(func() {
				conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")
			})

			values := url.Values{}

			t.Run("type=browser", func(t *testing.T) {
				res, err := testhelpers.NewHTTPClientWithArbitrarySessionCookie(t, reg).
					Do(httpx.MustNewRequest("POST", publicTS.URL+registration.RouteSubmitFlow, strings.NewReader(values.Encode()), "application/x-www-form-urlencoded"))
				require.NoError(t, err)
				defer res.Body.Close()
				assert.EqualValues(t, http.StatusOK, res.StatusCode, "%+v", res.Request)
				assert.Contains(t, res.Request.URL.String(), conf.Source().String(config.ViperKeySelfServiceBrowserDefaultReturnTo))
			})

			t.Run("type=api", func(t *testing.T) {
				res, err := testhelpers.NewHTTPClientWithArbitrarySessionToken(t, reg).
					Do(httpx.MustNewRequest("POST", publicTS.URL+registration.RouteSubmitFlow, strings.NewReader(testhelpers.EncodeFormAsJSON(t, true, values)), "application/json"))
				require.NoError(t, err)
				assert.Len(t, res.Cookies(), 0)
				defer res.Body.Close()
				assertx.EqualAsJSON(t, registration.ErrAlreadyLoggedIn, json.RawMessage(gjson.GetBytes(ioutilx.MustReadAll(res.Body), "error").Raw))
			})
		})

		t.Run("case=should show the error ui because the request payload is malformed", func(t *testing.T) {
			conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/profile.schema.json")
			t.Cleanup(func() {
				conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")
			})

			t.Run("type=api", func(t *testing.T) {
				f := testhelpers.InitializeRegistrationFlowViaAPI(t, apiClient, publicTS)
				body, res := testhelpers.RegistrationMakeRequest(t, true, false, f, apiClient, "14=)=!(%)$/ZP()GHIÖ")
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+registration.RouteSubmitFlow)
				assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
				assert.Contains(t, body, `Expected JSON sent in request body to be an object but got: Number`)
			})

			t.Run("type=spa", func(t *testing.T) {
				f := testhelpers.InitializeRegistrationFlowViaBrowser(t, apiClient, publicTS, true)
				body, res := testhelpers.RegistrationMakeRequest(t, true, false, f, apiClient, "14=)=!(%)$/ZP()GHIÖ")
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+registration.RouteSubmitFlow)
				assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
				assert.Contains(t, body, `Expected JSON sent in request body to be an object but got: Number`)
			})

			t.Run("type=browser", func(t *testing.T) {
				browserClient := testhelpers.NewClientWithCookies(t)
				f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS, false)
				body, res := testhelpers.RegistrationMakeRequest(t, false, false, f, browserClient, "14=)=!(%)$/ZP()GHIÖ")
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/registration-ts")
				assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
				assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "invalid URL escape", "%s", body)
				assert.Equal(t, "email", gjson.Get(body, "ui.nodes.#(attributes.name==\"traits.email\").attributes.type").String(), "%s", body)
			})
		})

		t.Run("case=should show the error ui because the method is missing in payload", func(t *testing.T) {
			conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/profile.schema.json")
			t.Cleanup(func() {
				conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")
			})

			t.Run("type=api", func(t *testing.T) {
				f := testhelpers.InitializeRegistrationFlowViaAPI(t, apiClient, publicTS)
				body, res := testhelpers.RegistrationMakeRequest(t, true, false, f, apiClient, "{}}")
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+registration.RouteSubmitFlow)
				assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
				assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "Could not find a strategy to sign you up with. Did you fill out the form correctly?", "%s", body)
			})

			t.Run("type=spa", func(t *testing.T) {
				browserClient := testhelpers.NewClientWithCookies(t)
				f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS, true)
				body, res := testhelpers.RegistrationMakeRequest(t, false, true, f, browserClient, "{}}")
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+registration.RouteSubmitFlow)
				assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
				assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "Could not find a strategy to sign you up with. Did you fill out the form correctly?", "%s", body)
			})

			t.Run("type=browser", func(t *testing.T) {
				browserClient := testhelpers.NewClientWithCookies(t)
				f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS, false)
				body, res := testhelpers.RegistrationMakeRequest(t, false, false, f, browserClient, "foo=bar")
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/registration-ts")
				assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
				assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "Could not find a strategy to sign you up with. Did you fill out the form correctly?", "%s", body)
			})
		})

		t.Run("case=should show the error ui because the request id is missing", func(t *testing.T) {
			var check = func(t *testing.T, actual string) {
				assert.Equal(t, int64(http.StatusNotFound), gjson.Get(actual, "code").Int(), "%s", actual)
				assert.Equal(t, "Not Found", gjson.Get(actual, "status").String(), "%s", actual)
				assert.Contains(t, gjson.Get(actual, "message").String(), "Unable to locate the resource", "%s", actual)
			}

			fakeFlow := &kratos.SelfServiceRegistrationFlow{Ui: kratos.UiContainer{
				Action: publicTS.URL + registration.RouteSubmitFlow + "?flow=" + x.NewUUID().String(),
			}}

			t.Run("type=api", func(t *testing.T) {
				actual, res := testhelpers.RegistrationMakeRequest(t, true, false, fakeFlow, apiClient, "{}")
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+registration.RouteSubmitFlow)
				check(t, gjson.Get(actual, "error").Raw)
			})

			t.Run("type=api", func(t *testing.T) {
				browserClient := testhelpers.NewClientWithCookies(t)
				actual, res := testhelpers.RegistrationMakeRequest(t, false, true, fakeFlow, browserClient, "{}")
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+registration.RouteSubmitFlow)
				check(t, gjson.Get(actual, "error").Raw)
			})

			t.Run("type=browser", func(t *testing.T) {
				browserClient := testhelpers.NewClientWithCookies(t)
				actual, res := testhelpers.RegistrationMakeRequest(t, false, false, fakeFlow, browserClient, "")
				assert.Contains(t, res.Request.URL.String(), errTS.URL)
				check(t, actual)
			})
		})

		t.Run("case=should return an error because the request is expired", func(t *testing.T) {
			conf.MustSet(config.ViperKeySelfServiceRegistrationRequestLifespan, "500ms")
			t.Cleanup(func() {
				conf.MustSet(config.ViperKeySelfServiceRegistrationRequestLifespan, "10m")
			})

			t.Run("type=api", func(t *testing.T) {
				f := testhelpers.InitializeRegistrationFlowViaAPI(t, apiClient, publicTS)

				time.Sleep(time.Millisecond * 600)
				actual, res := testhelpers.RegistrationMakeRequest(t, true, false, f, apiClient, "{}")
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+registration.RouteSubmitFlow)
				assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", gjson.Get(actual, "use_flow_id").String())
				assertx.EqualAsJSONExcept(t, flow.NewFlowExpiredError(time.Now()), json.RawMessage(actual), []string{"use_flow_id", "since"}, "expired", "%s", actual)
			})

			t.Run("type=spa", func(t *testing.T) {
				browserClient := testhelpers.NewClientWithCookies(t)
				f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS, true)

				time.Sleep(time.Millisecond * 600)
				actual, res := testhelpers.RegistrationMakeRequest(t, false, true, f, browserClient, "{}")
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+registration.RouteSubmitFlow)
				assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", gjson.Get(actual, "use_flow_id").String())
				assertx.EqualAsJSONExcept(t, flow.NewFlowExpiredError(time.Now()), json.RawMessage(actual), []string{"use_flow_id", "since"}, "expired", "%s", actual)
			})

			t.Run("type=browser", func(t *testing.T) {
				browserClient := testhelpers.NewClientWithCookies(t)
				f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS, false)

				time.Sleep(time.Millisecond * 600)
				actual, res := testhelpers.RegistrationMakeRequest(t, false, false, f, browserClient, "")
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/registration-ts")
				assert.NotEqual(t, f.Id, gjson.Get(actual, "id").String(), "%s", actual)
				assert.Contains(t, gjson.Get(actual, "ui.messages.0.text").String(), "expired", "%s", actual)
			})
		})

		var expectValidationError = func(t *testing.T, isAPI, isSPA bool, values func(url.Values)) string {
			return testhelpers.SubmitRegistrationForm(t, isAPI, nil, publicTS, values,
				isSPA,
				testhelpers.ExpectStatusCode(isAPI || isSPA, http.StatusBadRequest, http.StatusOK),
				testhelpers.ExpectURL(isAPI || isSPA, publicTS.URL+registration.RouteSubmitFlow, conf.SelfServiceFlowRegistrationUI().String()))
		}

		t.Run("case=should fail because the return_to url is not allowed", func(t *testing.T) {
			var check = func(t *testing.T, actual string) {
				assert.NotEmpty(t, gjson.Get(actual, "id").String(), "%s", actual)
				assert.Contains(t, gjson.Get(actual, "ui.action").String(), publicTS.URL+registration.RouteSubmitFlow, "%s", actual)
				checkFormContent(t, []byte(actual), "password", "csrf_token", "traits.username", "traits.foobar")
				assert.Contains(t, gjson.Get(actual, "ui.nodes.#(attributes.name==password).messages.0").String(), "data breaches and must no longer be used.", "%s", actual)

				// but the method is still set
				assert.Equal(t, "password", gjson.Get(actual, "ui.nodes.#(attributes.name==method).attributes.value").String(), "%s", actual)
			}

			var values = func(v url.Values) {
				v.Set("traits.username", "registration-identifier-4")
				v.Set("password", "password")
				v.Set("traits.foobar", "bar")
			}

			t.Run("type=api", func(t *testing.T) {
				check(t, expectValidationError(t, true, false, values))
			})

			t.Run("type=spa", func(t *testing.T) {
				check(t, expectValidationError(t, false, false, values))
			})

			t.Run("type=browser", func(t *testing.T) {
				check(t, expectValidationError(t, false, false, values))
			})
		})

		t.Run("case=should return an error because not passing validation", func(t *testing.T) {
			var check = func(t *testing.T, actual string) {
				assert.NotEmpty(t, gjson.Get(actual, "id").String(), "%s", actual)
				assert.Contains(t, gjson.Get(actual, "ui.action").String(), publicTS.URL+registration.RouteSubmitFlow, "%s", actual)
				checkFormContent(t, []byte(actual), "password", "csrf_token", "traits.username", "traits.foobar")
				assert.Contains(t, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.foobar).messages.0").String(), `Property foobar is missing`, "%s", actual)
			}

			var values = func(v url.Values) {
				v.Set("traits.username", "registration-identifier-5")
				v.Set("password", x.NewUUID().String())
				v.Del("traits.foobar")
			}

			t.Run("type=api", func(t *testing.T) {
				check(t, expectValidationError(t, true, false, values))
			})

			t.Run("type=spa", func(t *testing.T) {
				check(t, expectValidationError(t, false, true, values))
			})

			t.Run("type=browser", func(t *testing.T) {
				check(t, expectValidationError(t, false, false, values))
			})
		})

		t.Run("case=should have correct CSRF behavior", func(t *testing.T) {
			var values = url.Values{
				"method":          {"password"},
				"csrf_token":      {"invalid_token"},
				"traits.username": {"registration-identifier-csrf-browser"},
				"password":        {x.NewUUID().String()},
				"traits.foobar":   {"bar"},
			}

			t.Run("case=should fail because of missing CSRF token/type=browser", func(t *testing.T) {
				browserClient := testhelpers.NewClientWithCookies(t)
				f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS, false)

				actual, res := testhelpers.RegistrationMakeRequest(t, false, false, f, browserClient, values.Encode())
				assert.EqualValues(t, http.StatusOK, res.StatusCode)
				assertx.EqualAsJSON(t, x.ErrInvalidCSRFToken,
					json.RawMessage(actual), "%s", actual)
			})

			t.Run("case=should fail because of missing CSRF token/type=spa", func(t *testing.T) {
				browserClient := testhelpers.NewClientWithCookies(t)
				f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS, true)

				actual, res := testhelpers.RegistrationMakeRequest(t, false, true, f, browserClient, values.Encode())
				assert.EqualValues(t, http.StatusForbidden, res.StatusCode)
				assertx.EqualAsJSON(t, x.ErrInvalidCSRFToken,
					json.RawMessage(gjson.Get(actual, "error").Raw), "%s", actual)
			})

			t.Run("case=should pass even without CSRF token/type=api", func(t *testing.T) {
				f := testhelpers.InitializeRegistrationFlowViaAPI(t, apiClient, publicTS)

				actual, res := testhelpers.RegistrationMakeRequest(t, true, false, f, apiClient, testhelpers.EncodeFormAsJSON(t, true, values))
				assert.EqualValues(t, http.StatusOK, res.StatusCode)
				assert.NotEmpty(t, gjson.Get(actual, "identity.id").Raw, "%s", actual) // registration successful
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
						f := testhelpers.InitializeRegistrationFlowViaAPI(t, apiClient, publicTS)
						c := f.Ui

						req := testhelpers.NewRequest(t, true, "POST", c.Action, bytes.NewBufferString(testhelpers.EncodeFormAsJSON(t, true, values)))
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

		t.Run("case=should fail because schema did not specify an identifier", func(t *testing.T) {
			conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/missing-identifier.schema.json")

			values := func(values url.Values) {
				values.Set("method", "password")
				values.Set("traits.username", "registration-identifier-6")
				values.Set("password", x.NewUUID().String())
				values.Set("traits.foobar", "bar")
			}

			t.Run("type=api", func(t *testing.T) {
				body := testhelpers.SubmitRegistrationForm(t, true, apiClient, publicTS, values,
					false, http.StatusBadRequest,
					publicTS.URL+registration.RouteSubmitFlow)
				assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "Could not find any login identifiers. Did you forget to set them?", "%s", body)
			})

			t.Run("type=spa", func(t *testing.T) {
				browserClient := testhelpers.NewClientWithCookies(t)
				body := testhelpers.SubmitRegistrationForm(t, false, browserClient, publicTS, values,
					true, http.StatusBadRequest,
					publicTS.URL+registration.RouteSubmitFlow)
				assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "Could not find any login identifiers. Did you forget to set them?", "%s", body)
			})

			t.Run("type=browser", func(t *testing.T) {
				body := expectValidationError(t, false, false, values)
				assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "Could not find any login identifiers. Did you forget to set them?", "%s", body)
			})
		})

		t.Run("case=should fail because schema does not exist", func(t *testing.T) {
			var check = func(t *testing.T, actual string) {
				assert.Equal(t, int64(http.StatusInternalServerError), gjson.Get(actual, "code").Int(), "%s", actual)
				assert.Equal(t, "Internal Server Error", gjson.Get(actual, "status").String(), "%s", actual)
				assert.Contains(t, gjson.Get(actual, "reason").String(), "no such file or directory", "%s", actual)
			}

			values := url.Values{
				"method":          {"password"},
				"traits.username": {"registration-identifier-7"},
				"password":        {x.NewUUID().String()},
				"traits.foobar":   {"bar"},
				"csrf_token":      {x.FakeCSRFToken},
			}

			t.Run("type=api", func(t *testing.T) {
				f := testhelpers.InitializeRegistrationFlowViaAPI(t, apiClient, publicTS)

				conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/i-do-not-exist.schema.json")
				t.Cleanup(func() {
					conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")
				})

				body, res := testhelpers.RegistrationMakeRequest(t, false, false, f, apiClient, values.Encode())
				assert.Contains(t, res.Request.URL.String(), publicTS.URL)
				check(t, gjson.Get(body, "error").Raw)
			})

			t.Run("type=spa", func(t *testing.T) {
				browserClient := testhelpers.NewClientWithCookies(t)
				f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS, true)

				conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/i-do-not-exist.schema.json")
				t.Cleanup(func() {
					conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")
				})

				body, res := testhelpers.RegistrationMakeRequest(t, false, true, f, apiClient, values.Encode())
				assert.Contains(t, res.Request.URL.String(), publicTS.URL)
				check(t, gjson.Get(body, "error").Raw)
			})

			t.Run("type=browser", func(t *testing.T) {
				browserClient := testhelpers.NewClientWithCookies(t)
				f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS, false)

				conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/i-do-not-exist.schema.json")
				t.Cleanup(func() {
					conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")
				})
				body, res := testhelpers.RegistrationMakeRequest(t, false, false, f, apiClient, values.Encode())
				assert.Contains(t, res.Request.URL.String(), errTS.URL)
				check(t, body)
			})
		})

		var expectLoginBody = func(t *testing.T, browserRedirTS *httptest.Server, isAPI, isSPA bool, hc *http.Client, values func(url.Values)) string {
			if isAPI {
				return testhelpers.SubmitRegistrationForm(t, isAPI, hc, publicTS, values,
					isSPA, http.StatusOK,
					publicTS.URL+registration.RouteSubmitFlow)
			}

			if hc == nil {
				hc = testhelpers.NewClientWithCookies(t)
			}

			expectReturnTo := browserRedirTS.URL + "/registration-return-ts"
			if isSPA {
				expectReturnTo = publicTS.URL
			}

			return testhelpers.SubmitRegistrationForm(t, isAPI, hc, publicTS, values,
				isSPA, http.StatusOK, expectReturnTo)
		}

		var expectSuccessfulLogin = func(t *testing.T, isAPI, isSPA bool, hc *http.Client, values func(url.Values)) string {
			useReturnToFromTS(redirTS)
			return expectLoginBody(t, redirTS, isAPI, isSPA, hc, values)
		}

		var expectNoLogin = func(t *testing.T, isAPI, isSPA bool, hc *http.Client, values func(url.Values)) string {
			useReturnToFromTS(redirNoSessionTS)
			t.Cleanup(func() {
				useReturnToFromTS(redirTS)
			})
			return expectLoginBody(t, redirNoSessionTS, isAPI, isSPA, hc, values)
		}

		t.Run("case=should pass and set up a session", func(t *testing.T) {
			conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")
			conf.MustSet(config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()), []config.SelfServiceHook{{Name: "session"}})
			t.Cleanup(func() {
				conf.MustSet(config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()), nil)
			})

			t.Run("type=api", func(t *testing.T) {
				body := expectSuccessfulLogin(t, true, false, nil, func(v url.Values) {
					v.Set("traits.username", "registration-identifier-8-api")
					v.Set("password", x.NewUUID().String())
					v.Set("traits.foobar", "bar")
				})
				assert.Equal(t, `registration-identifier-8-api`, gjson.Get(body, "identity.traits.username").String(), "%s", body)
				assert.NotEmpty(t, gjson.Get(body, "session_token").String(), "%s", body)
				assert.NotEmpty(t, gjson.Get(body, "session.id").String(), "%s", body)
			})

			t.Run("type=spa", func(t *testing.T) {
				body := expectSuccessfulLogin(t, false, true, nil, func(v url.Values) {
					v.Set("traits.username", "registration-identifier-8-spa")
					v.Set("password", x.NewUUID().String())
					v.Set("traits.foobar", "bar")
				})
				assert.Equal(t, `registration-identifier-8-spa`, gjson.Get(body, "identity.traits.username").String(), "%s", body)
				assert.Empty(t, gjson.Get(body, "session_token").String(), "%s", body)
				assert.NotEmpty(t, gjson.Get(body, "session.id").String(), "%s", body)
			})

			t.Run("type=browser", func(t *testing.T) {
				body := expectSuccessfulLogin(t, false, false, nil, func(v url.Values) {
					v.Set("traits.username", "registration-identifier-8-browser")
					v.Set("password", x.NewUUID().String())
					v.Set("traits.foobar", "bar")
				})
				assert.Equal(t, `registration-identifier-8-browser`, gjson.Get(body, "identity.traits.username").String(), "%s", body)
			})
		})

		t.Run("case=should not set up a session if hook is not configured", func(t *testing.T) {
			conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")
			conf.MustSet(config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()), nil)

			t.Run("type=api", func(t *testing.T) {
				body := expectNoLogin(t, true, false, nil, func(v url.Values) {
					v.Set("traits.username", "registration-identifier-8-api-nosession")
					v.Set("password", x.NewUUID().String())
					v.Set("traits.foobar", "bar")
				})
				assert.Equal(t, `registration-identifier-8-api-nosession`, gjson.Get(body, "identity.traits.username").String(), "%s", body)
				assert.Empty(t, gjson.Get(body, "session_token").String(), "%s", body)
				assert.Empty(t, gjson.Get(body, "session.id").String(), "%s", body)
			})

			t.Run("type=spa", func(t *testing.T) {
				expectNoLogin(t, false, true, nil, func(v url.Values) {
					v.Set("traits.username", "registration-identifier-8-spa-nosession")
					v.Set("password", x.NewUUID().String())
					v.Set("traits.foobar", "bar")
				})
			})

			t.Run("type=browser", func(t *testing.T) {
				expectNoLogin(t, false, false, nil, func(v url.Values) {
					v.Set("traits.username", "registration-identifier-8-browser-nosession")
					v.Set("password", x.NewUUID().String())
					v.Set("traits.foobar", "bar")
				})
			})
		})

		t.Run("case=should fail to register the same user again", func(t *testing.T) {
			conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")
			conf.MustSet(config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()), []config.SelfServiceHook{{Name: "session"}})
			t.Cleanup(func() {
				conf.MustSet(config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()), nil)
			})

			t.Run("type=api", func(t *testing.T) {
				values := func(v url.Values) {
					v.Set("traits.username", "registration-identifier-8-api-duplicate")
					v.Set("password", x.NewUUID().String())
					v.Set("traits.foobar", "bar")
				}

				_ = expectSuccessfulLogin(t, true, false, apiClient, values)
				body := testhelpers.SubmitRegistrationForm(t, true, apiClient, publicTS, values,
					false, http.StatusBadRequest,
					publicTS.URL+registration.RouteSubmitFlow)

				assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "An account with the same identifier (email, phone, username, ...) exists already.", "%s", body)
			})

			t.Run("type=spa", func(t *testing.T) {
				values := func(v url.Values) {
					v.Set("traits.username", "registration-identifier-8-spa-duplicate")
					v.Set("password", x.NewUUID().String())
					v.Set("traits.foobar", "bar")
				}

				_ = expectSuccessfulLogin(t, false, true, nil, values)
				body := expectValidationError(t, false, true, values)
				assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "An account with the same identifier (email, phone, username, ...) exists already.", "%s", body)
			})

			t.Run("type=browser", func(t *testing.T) {
				values := func(v url.Values) {
					v.Set("traits.username", "registration-identifier-8-browser-duplicate")
					v.Set("password", x.NewUUID().String())
					v.Set("traits.foobar", "bar")
				}

				_ = expectSuccessfulLogin(t, false, false, nil, values)
				body := expectValidationError(t, false, false, values)
				assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "An account with the same identifier (email, phone, username, ...) exists already.", "%s", body)
			})
		})

		t.Run("case=should return an error because not passing validation and reset previous errors and values", func(t *testing.T) {
			conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")

			var check = func(t *testing.T, actual string) {
				assert.NotEmpty(t, gjson.Get(actual, "id").String(), "%s", actual)
				assert.Contains(t, gjson.Get(actual, "ui.action").String(), publicTS.URL+registration.RouteSubmitFlow, "%s", actual)
				checkFormContent(t, []byte(actual), "password", "csrf_token", "traits.username")
			}

			var checkFirst = func(t *testing.T, actual string) {
				check(t, actual)
				assert.Contains(t, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.username).messages.0").String(), `Property username is missing`, "%s", actual)
			}

			var checkSecond = func(t *testing.T, actual string) {
				check(t, actual)
				assert.EqualValues(t, "registration-identifier-9", gjson.Get(actual, "ui.nodes.#(attributes.name==traits.username).attributes.value").String(), "%s", actual)
				assert.Empty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.username).attributes.error").Raw)
				assert.Empty(t, gjson.Get(actual, "ui.nodes.error").Raw)
				assert.Contains(t, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.foobar).messages.0").String(), `Property foobar is missing`, "%s", actual)
			}

			var valuesFirst = func(v url.Values) url.Values {
				v.Del("traits.username")
				v.Set("password", x.NewUUID().String())
				v.Set("traits.foobar", "bar")
				return v
			}

			var valuesSecond = func(v url.Values) url.Values {
				v.Set("traits.username", "registration-identifier-9")
				v.Set("password", x.NewUUID().String())
				v.Del("traits.foobar")
				return v
			}

			t.Run("type=api", func(t *testing.T) {
				f := testhelpers.InitializeRegistrationFlowViaAPI(t, apiClient, publicTS)
				c := f.Ui

				actual, _ := testhelpers.RegistrationMakeRequest(t, true, false, f, apiClient, testhelpers.EncodeFormAsJSON(t, true, valuesFirst(testhelpers.SDKFormFieldsToURLValues(c.Nodes))))
				checkFirst(t, actual)
				actual, _ = testhelpers.RegistrationMakeRequest(t, true, false, f, apiClient, testhelpers.EncodeFormAsJSON(t, true, valuesSecond(testhelpers.SDKFormFieldsToURLValues(c.Nodes))))
				checkSecond(t, actual)
			})

			t.Run("type=spa", func(t *testing.T) {
				browserClient := testhelpers.NewClientWithCookies(t)
				f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS, true)
				c := f.Ui

				actual, _ := testhelpers.RegistrationMakeRequest(t, false, true, f, browserClient, testhelpers.EncodeFormAsJSON(t, false, valuesFirst(testhelpers.SDKFormFieldsToURLValues(c.Nodes))))
				checkFirst(t, actual)
				actual, _ = testhelpers.RegistrationMakeRequest(t, false, true, f, browserClient, testhelpers.EncodeFormAsJSON(t, false, valuesSecond(testhelpers.SDKFormFieldsToURLValues(c.Nodes))))
				checkSecond(t, actual)
			})

			t.Run("type=browser", func(t *testing.T) {
				browserClient := testhelpers.NewClientWithCookies(t)
				f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS, false)
				c := f.Ui

				actual, _ := testhelpers.RegistrationMakeRequest(t, false, false, f, browserClient, valuesFirst(testhelpers.SDKFormFieldsToURLValues(c.Nodes)).Encode())
				checkFirst(t, actual)
				actual, _ = testhelpers.RegistrationMakeRequest(t, false, false, f, browserClient, valuesSecond(testhelpers.SDKFormFieldsToURLValues(c.Nodes)).Encode())
				checkSecond(t, actual)
			})
		})

		t.Run("case=should work even if password is just numbers", func(t *testing.T) {
			conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://stub/registration.schema.json")
			conf.MustSet(config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()), []config.SelfServiceHook{{Name: "session"}})
			t.Cleanup(func() {
				conf.MustSet(config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()), nil)
			})

			t.Run("type=api", func(t *testing.T) {
				actual := expectSuccessfulLogin(t, true, false, nil, func(v url.Values) {
					v.Set("traits.username", "registration-identifier-10-api")
					v.Set("password", x.NewUUID().String())
					v.Set("traits.foobar", "bar")
				})
				assert.Equal(t, `registration-identifier-10-api`, gjson.Get(actual, "identity.traits.username").String(), "%s", actual)
			})

			t.Run("type=spa", func(t *testing.T) {
				actual := expectSuccessfulLogin(t, false, false, nil, func(v url.Values) {
					v.Set("traits.username", "registration-identifier-10-spa")
					v.Set("password", x.NewUUID().String())
					v.Set("traits.foobar", "bar")
				})
				assert.Equal(t, `registration-identifier-10-spa`, gjson.Get(actual, "identity.traits.username").String(), "%s", actual)
			})

			t.Run("type=browser", func(t *testing.T) {
				actual := expectSuccessfulLogin(t, false, false, nil, func(v url.Values) {
					v.Set("traits.username", "registration-identifier-10-browser")
					v.Set("password", x.NewUUID().String())
					v.Set("traits.foobar", "bar")
				})
				assert.Equal(t, `registration-identifier-10-browser`, gjson.Get(actual, "identity.traits.username").String(), "%s", actual)
			})
		})

		t.Run("case=should work with regular JSON", func(t *testing.T) {
			conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://stub/registration.schema.json")
			conf.MustSet(config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()), []config.SelfServiceHook{{Name: "session"}})
			t.Cleanup(func() {
				conf.MustSet(config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()), nil)
			})

			hc := testhelpers.NewClientWithCookies(t)
			hc.Transport = testhelpers.NewTransportWithLogger(hc.Transport, t)
			payload := testhelpers.InitializeRegistrationFlowViaBrowser(t, hc, publicTS, false)
			values := testhelpers.SDKFormFieldsToURLValues(payload.Ui.Nodes)
			time.Sleep(time.Millisecond) // add a bit of delay to allow `1ns` to time out.

			username := x.NewUUID()
			actual, res := testhelpers.RegistrationMakeRequest(t, true, false, payload, hc, fmt.Sprintf(`{
  "method": "password",
  "csrf_token": "%s",
  "password": "%s",
  "traits": {
    "foobar": "bar",
    "username": "%s"
  }
}`, values.Get("csrf_token"), x.NewUUID(), username))
			assert.EqualValues(t, http.StatusOK, res.StatusCode, assertx.PrettifyJSONPayload(t, actual))
			assert.Equal(t, username.String(), gjson.Get(actual, "identity.traits.username").String(), "%s", actual)
		})
	})

	t.Run("method=PopulateSignUpMethod", func(t *testing.T) {
		conf, reg := internal.NewFastRegistryWithMocks(t)

		conf.MustSet(config.ViperKeyPublicBaseURL, "https://foo/")
		conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://stub/sort.schema.json")
		conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword)+".enabled", true)

		router := x.NewRouterPublic()
		publicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin())
		_ = testhelpers.NewRegistrationUIFlowEchoServer(t, reg)

		browserClient := testhelpers.NewClientWithCookies(t)
		f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS, false)

		assertx.EqualAsJSON(t, container.Container{
			Action: conf.SelfPublicURL().String() + registration.RouteSubmitFlow + "?flow=" + f.Id,
			Method: "POST",
			Nodes: node.Nodes{
				node.NewCSRFNode(x.FakeCSRFToken),
				node.NewInputField("traits.username", nil, node.PasswordGroup, node.InputAttributeTypeText),
				node.NewInputField("password", nil, node.PasswordGroup, node.InputAttributeTypePassword, node.WithRequiredInputAttribute).WithMetaLabel(text.NewInfoNodeInputPassword()),
				node.NewInputField("traits.bar", nil, node.PasswordGroup, node.InputAttributeTypeText),
				node.NewInputField("method", "password", node.PasswordGroup, node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoRegistration()),
			},
		}, f.Ui)
	})
}
