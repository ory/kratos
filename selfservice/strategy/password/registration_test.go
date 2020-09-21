package password_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/httpx"
	"github.com/ory/x/pointerx"

	"github.com/ory/x/assertx"

	"github.com/ory/viper"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/httpclient/models"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/selfservice/strategy/password"
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
		assert.Equal(t, f, gjson.GetBytes(body, fmt.Sprintf("methods.password.config.fields.#(name==%s).name", f)).String(), "%s", body)
	}
}

// checks if some keys are not set, this should be used to catch regression issues
func outdatedFieldsDoNotExist(t *testing.T, body []byte) {
	for _, k := range []string{"request"} {
		assert.Equal(t, false, gjson.GetBytes(body, fmt.Sprintf("methods.password.config.fields.#(name==%s)", k)).Exists())
	}
}

func formMethodIsPOST(t *testing.T, body []byte) {
	assert.Equal(t, "POST", gjson.GetBytes(body, "methods.password.config.method").String())
}

func TestRegistration(t *testing.T) {
	t.Run("case=registration", func(t *testing.T) {
		conf, reg := internal.NewFastRegistryWithMocks(t)

		router := x.NewRouterPublic()
		admin := x.NewRouterAdmin()
		viper.Set(configuration.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword), map[string]interface{}{"enabled": true})

		publicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, admin)
		errTS := testhelpers.NewErrorTestServer(t, reg)
		uiTS := testhelpers.NewRegistrationUIFlowEchoServer(t, reg)
		redirTS := testhelpers.NewRedirSessionEchoTS(t, reg)

		// Overwrite these two to ensure that they run
		viper.Set(configuration.ViperKeySelfServiceBrowserDefaultReturnTo, redirTS.URL+"/default-return-to")
		viper.Set(configuration.ViperKeySelfServiceRegistrationAfter+"."+configuration.DefaultBrowserReturnURL, redirTS.URL+"/registration-return-ts")
		viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")

		apiClient := testhelpers.NewDebugClient(t)

		t.Run("description=can call endpoints only without session", func(t *testing.T) {
			// Needed to set up the mock IDs...
			viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/profile.schema.json")
			defer viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")

			values := url.Values{}

			t.Run("type=browser", func(t *testing.T) {
				res, err := testhelpers.NewHTTPClientWithArbitrarySessionCookie(t, reg).
					Do(httpx.MustNewRequest("POST", publicTS.URL+password.RouteRegistration, strings.NewReader(values.Encode()), "application/x-www-form-urlencoded"))
				require.NoError(t, err)
				defer res.Body.Close()
				assert.EqualValues(t, http.StatusOK, res.StatusCode, "%+v", res.Request)
				assert.Contains(t, res.Request.URL.String(), viper.GetString(configuration.ViperKeySelfServiceBrowserDefaultReturnTo))
			})

			t.Run("type=api", func(t *testing.T) {
				res, err := testhelpers.NewHTTPClientWithArbitrarySessionToken(t, reg).
					Do(httpx.MustNewRequest("POST", publicTS.URL+password.RouteRegistration, strings.NewReader(testhelpers.EncodeFormAsJSON(t, true, values)), "application/json"))
				require.NoError(t, err)
				defer res.Body.Close()
				assertx.EqualAsJSON(t, registration.ErrAlreadyLoggedIn, json.RawMessage(gjson.GetBytes(x.MustReadAll(res.Body), "error").Raw))
			})
		})

		t.Run("case=should show the error ui because the request payload is malformed", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/profile.schema.json")
			defer viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")

			t.Run("type=api", func(t *testing.T) {
				f := testhelpers.InitializeRegistrationFlowViaAPI(t, apiClient, publicTS)
				c := testhelpers.GetRegistrationFlowMethodConfig(t, f.Payload, identity.CredentialsTypePassword.String())
				body, res := testhelpers.RegistrationMakeRequest(t, true, c, apiClient, "14=)=!(%)$/ZP()GHIÖ")
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+password.RouteRegistration)
				assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
				assert.Contains(t, body, `Expected JSON sent in request body to be an object but got: Number`)
			})

			t.Run("type=browser", func(t *testing.T) {
				browserClient := testhelpers.NewClientWithCookies(t)
				f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS)
				c := testhelpers.GetRegistrationFlowMethodConfig(t, f.Payload, identity.CredentialsTypePassword.String())
				body, res := testhelpers.RegistrationMakeRequest(t, false, c, browserClient, "14=)=!(%)$/ZP()GHIÖ")
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/registration-ts")
				assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
				assert.Contains(t, gjson.Get(body, "methods.password.config.messages.0.text").String(), "invalid URL escape", "%s", body)
				assert.Equal(t, "email", gjson.Get(body, "methods.password.config.fields.#(name==\"traits.email\").type").String(), "%s", body)
			})
		})

		t.Run("case=should show the error ui because the request id is missing", func(t *testing.T) {
			var check = func(t *testing.T, actual string) {
				assert.Equal(t, int64(http.StatusNotFound), gjson.Get(actual, "code").Int(), "%s", actual)
				assert.Equal(t, "Not Found", gjson.Get(actual, "status").String(), "%s", actual)
				assert.Contains(t, gjson.Get(actual, "message").String(), "Unable to locate the resource", "%s", actual)
			}

			fakeFlow := &models.RegistrationFlowMethodConfig{
				Action: pointerx.String(publicTS.URL + password.RouteRegistration + "?flow=" + x.NewUUID().String())}

			t.Run("type=api", func(t *testing.T) {
				actual, res := testhelpers.RegistrationMakeRequest(t, true, fakeFlow, apiClient, "{}")
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+password.RouteRegistration)
				check(t, gjson.Get(actual, "error").Raw)
			})

			t.Run("type=browser", func(t *testing.T) {
				browserClient := testhelpers.NewClientWithCookies(t)
				actual, res := testhelpers.RegistrationMakeRequest(t, false, fakeFlow, browserClient, "")
				assert.Contains(t, res.Request.URL.String(), errTS.URL)
				check(t, gjson.Get(actual, "0").Raw)
			})
		})

		t.Run("case=should return an error because the request is expired", func(t *testing.T) {
			viper.Set(configuration.ViperKeySelfServiceRegistrationRequestLifespan, "100ms")
			defer viper.Set(configuration.ViperKeySelfServiceRegistrationRequestLifespan, "10m")

			t.Run("type=api", func(t *testing.T) {
				f := testhelpers.InitializeRegistrationFlowViaAPI(t, apiClient, publicTS)
				c := testhelpers.GetRegistrationFlowMethodConfig(t, f.Payload, identity.CredentialsTypePassword.String())

				time.Sleep(time.Millisecond * 200)
				actual, res := testhelpers.RegistrationMakeRequest(t, true, c, apiClient, "{}")
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+registration.RouteGetFlow)
				assert.NotEqual(t, f.Payload.ID, gjson.Get(actual, "id").String(), "%s", actual)
				assert.Contains(t, gjson.Get(actual, "messages.0.text").String(), "expired", "%s", actual)
			})

			t.Run("type=browser", func(t *testing.T) {
				browserClient := testhelpers.NewClientWithCookies(t)
				f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS)
				c := testhelpers.GetRegistrationFlowMethodConfig(t, f.Payload, identity.CredentialsTypePassword.String())

				time.Sleep(time.Millisecond * 200)
				actual, res := testhelpers.RegistrationMakeRequest(t, false, c, browserClient, "")
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/registration-ts")
				assert.NotEqual(t, f.Payload.ID, gjson.Get(actual, "id").String(), "%s", actual)
				assert.Contains(t, gjson.Get(actual, "messages.0.text").String(), "expired", "%s", actual)
			})
		})

		var expectValidationError = func(t *testing.T, isAPI bool, values func(url.Values)) string {
			return testhelpers.SubmitRegistrationForm(t, isAPI, nil, publicTS, values,
				identity.CredentialsTypePassword,
				testhelpers.ExpectStatusCode(isAPI, http.StatusBadRequest, http.StatusOK),
				testhelpers.ExpectURL(isAPI, publicTS.URL+password.RouteRegistration, conf.SelfServiceFlowRegistrationUI().String()))
		}

		t.Run("case=should return an error because the password failed validation", func(t *testing.T) {
			var check = func(t *testing.T, actual string) {
				assert.NotEmpty(t, gjson.Get(actual, "id").String(), "%s", actual)
				assert.Contains(t, gjson.Get(actual, "methods.password.config.action").String(), publicTS.URL+password.RouteRegistration, "%s", actual)
				checkFormContent(t, []byte(actual), "password", "csrf_token", "traits.username", "traits.foobar")
				assert.Contains(t, gjson.Get(actual, "methods.password.config.fields.#(name==password).messages.0").String(), "data breaches and must no longer be used.", "%s", actual)
			}

			var values = func(v url.Values) {
				v.Set("traits.username", "registration-identifier-4")
				v.Set("password", "password")
				v.Set("traits.foobar", "bar")
			}

			t.Run("type=api", func(t *testing.T) {
				check(t, expectValidationError(t, true, values))
			})

			t.Run("type=browser", func(t *testing.T) {
				check(t, expectValidationError(t, false, values))
			})
		})

		t.Run("case=should return an error because not passing validation", func(t *testing.T) {
			var check = func(t *testing.T, actual string) {
				assert.NotEmpty(t, gjson.Get(actual, "id").String(), "%s", actual)
				assert.Contains(t, gjson.Get(actual, "methods.password.config.action").String(), publicTS.URL+password.RouteRegistration, "%s", actual)
				checkFormContent(t, []byte(actual), "password", "csrf_token", "traits.username", "traits.foobar")
				assert.Contains(t, gjson.Get(actual, "methods.password.config.fields.#(name==traits.foobar).messages.0").String(), `Property foobar is missing`, "%s", actual)
			}

			var values = func(v url.Values) {
				v.Set("traits.username", "registration-identifier-5")
				v.Set("password", x.NewUUID().String())
				v.Del("traits.foobar")
			}

			t.Run("type=api", func(t *testing.T) {
				check(t, expectValidationError(t, true, values))
			})

			t.Run("type=browser", func(t *testing.T) {
				check(t, expectValidationError(t, false, values))
			})
		})

		t.Run("case=should have correct CSRF behavior", func(t *testing.T) {
			var values = url.Values{
				"csrf_token":      {"invalid_token"},
				"traits.username": {"registration-identifier-csrf-browser"},
				"password":        {x.NewUUID().String()},
				"traits.foobar":   {"bar"},
			}
			t.Run("case=should fail because of missing CSRF token/type=browser", func(t *testing.T) {
				browserClient := testhelpers.NewClientWithCookies(t)
				f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS)
				c := testhelpers.GetRegistrationFlowMethodConfig(t, f.Payload, identity.CredentialsTypePassword.String())

				actual, res := testhelpers.RegistrationMakeRequest(t, false, c, browserClient, values.Encode())
				assert.EqualValues(t, http.StatusOK, res.StatusCode)
				assertx.EqualAsJSON(t, x.ErrInvalidCSRFToken,
					json.RawMessage(gjson.Get(actual, "0").Raw), "%s", actual)
			})

			t.Run("case=should pass even without CSRF token/type=api", func(t *testing.T) {
				f := testhelpers.InitializeRegistrationFlowViaAPI(t, apiClient, publicTS)
				c := testhelpers.GetRegistrationFlowMethodConfig(t, f.Payload, identity.CredentialsTypePassword.String())

				actual, res := testhelpers.RegistrationMakeRequest(t, true, c, apiClient, testhelpers.EncodeFormAsJSON(t, true, values))
				assert.EqualValues(t, http.StatusOK, res.StatusCode)
				assert.NotEmpty(t, gjson.Get(actual, "identity.id").Raw, "%s", actual) // registration successful
			})
		})

		t.Run("case=should fail because schema did not specify an identifier", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/missing-identifier.schema.json")

			var check = func(t *testing.T, actual string) {
				assert.Equal(t, int64(http.StatusInternalServerError), gjson.Get(actual, "code").Int(), "%s", actual)
				assert.Equal(t, "Internal Server Error", gjson.Get(actual, "status").String(), "%s", actual)
				assert.Contains(t, gjson.Get(actual, "reason").String(), "No login identifiers", "%s", actual)
			}

			values := url.Values{
				"traits.username": {"registration-identifier-6"},
				"password":        {x.NewUUID().String()},
				"traits.foobar":   {"bar"},
				"csrf_token":      {x.FakeCSRFToken},
			}

			t.Run("type=api", func(t *testing.T) {
				f := testhelpers.InitializeRegistrationFlowViaAPI(t, apiClient, publicTS)
				c := testhelpers.GetRegistrationFlowMethodConfig(t, f.Payload, identity.CredentialsTypePassword.String())

				body, res := testhelpers.RegistrationMakeRequest(t, false, c, apiClient, values.Encode())
				assert.Contains(t, res.Request.URL.String(), publicTS.URL)
				check(t, gjson.Get(body, "error").Raw)
			})

			t.Run("type=browser", func(t *testing.T) {
				browserClient := testhelpers.NewClientWithCookies(t)
				f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS)
				c := testhelpers.GetRegistrationFlowMethodConfig(t, f.Payload, identity.CredentialsTypePassword.String())

				body, res := testhelpers.RegistrationMakeRequest(t, false, c, apiClient, values.Encode())
				assert.Contains(t, res.Request.URL.String(), errTS.URL)
				check(t, gjson.Get(body, "0").Raw)
			})
		})

		t.Run("case=should fail because schema does not exist", func(t *testing.T) {
			var check = func(t *testing.T, actual string) {
				assert.Equal(t, int64(http.StatusInternalServerError), gjson.Get(actual, "code").Int(), "%s", actual)
				assert.Equal(t, "Internal Server Error", gjson.Get(actual, "status").String(), "%s", actual)
				assert.Contains(t, gjson.Get(actual, "reason").String(), "no such file or directory", "%s", actual)
			}

			values := url.Values{
				"traits.username": {"registration-identifier-7"},
				"password":        {x.NewUUID().String()},
				"traits.foobar":   {"bar"},
				"csrf_token":      {x.FakeCSRFToken},
			}

			t.Run("type=api", func(t *testing.T) {
				f := testhelpers.InitializeRegistrationFlowViaAPI(t, apiClient, publicTS)
				c := testhelpers.GetRegistrationFlowMethodConfig(t, f.Payload, identity.CredentialsTypePassword.String())

				viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/i-do-not-exist.schema.json")
				defer viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")

				body, res := testhelpers.RegistrationMakeRequest(t, false, c, apiClient, values.Encode())
				assert.Contains(t, res.Request.URL.String(), publicTS.URL)
				check(t, gjson.Get(body, "error").Raw)
			})

			t.Run("type=browser", func(t *testing.T) {
				browserClient := testhelpers.NewClientWithCookies(t)
				f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS)
				c := testhelpers.GetRegistrationFlowMethodConfig(t, f.Payload, identity.CredentialsTypePassword.String())

				viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/i-do-not-exist.schema.json")
				defer viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")

				body, res := testhelpers.RegistrationMakeRequest(t, false, c, apiClient, values.Encode())
				assert.Contains(t, res.Request.URL.String(), errTS.URL)
				check(t, gjson.Get(body, "0").Raw)
			})
		})

		var expectSuccessfulLogin = func(t *testing.T, isAPI bool, hc *http.Client, values func(url.Values)) string {
			if isAPI {
				return testhelpers.SubmitRegistrationForm(t, isAPI, hc, publicTS, values,
					identity.CredentialsTypePassword, http.StatusOK,
					publicTS.URL+password.RouteRegistration)
			}

			if hc == nil {
				hc = testhelpers.NewClientWithCookies(t)
			}

			return testhelpers.SubmitRegistrationForm(t, isAPI, hc, publicTS, values,
				identity.CredentialsTypePassword, http.StatusOK,
				redirTS.URL+"/registration-return-ts")
		}

		t.Run("case=should pass and set up a session", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")
			viper.Set(configuration.HookStrategyKey(configuration.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()), []configuration.SelfServiceHook{{Name: "session"}})
			defer viper.Set(configuration.HookStrategyKey(configuration.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()), nil)

			var values = func(isAPI bool) func(v url.Values) {
				return func(v url.Values) {
					v.Set("traits.username", "registration-identifier-8-browser")
					if isAPI {
						v.Set("traits.username", "registration-identifier-8-api")
					}
					v.Set("password", x.NewUUID().String())
					v.Set("traits.foobar", "bar")
				}
			}

			t.Run("type=api", func(t *testing.T) {
				body := expectSuccessfulLogin(t, true, nil, values(true))
				assert.Equal(t, `registration-identifier-8-api`, gjson.Get(body, "identity.traits.username").String(), "%s", body)
				assert.NotEmpty(t, gjson.Get(body, "session_token").String(), "%s", body)
				assert.NotEmpty(t, gjson.Get(body, "session.id").String(), "%s", body)
			})

			t.Run("type=browser", func(t *testing.T) {
				body := expectSuccessfulLogin(t, false, nil, values(false))
				assert.Equal(t, `registration-identifier-8-browser`, gjson.Get(body, "identity.traits.username").String(), "%s", body)
			})
		})

		t.Run("case=should fail to register the same user again", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")
			viper.Set(configuration.HookStrategyKey(configuration.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()), []configuration.SelfServiceHook{{Name: "session"}})
			defer viper.Set(configuration.HookStrategyKey(configuration.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()), nil)

			var values = func(isAPI bool) func(v url.Values) {
				return func(v url.Values) {
					v.Set("traits.username", "registration-identifier-8-browser-duplicate")
					if isAPI {
						v.Set("traits.username", "registration-identifier-8-api-duplicate")
					}
					v.Set("password", x.NewUUID().String())
					v.Set("traits.foobar", "bar")
				}
			}

			t.Run("type=api", func(t *testing.T) {
				_ = expectSuccessfulLogin(t, true, apiClient, values(true))
				body := testhelpers.SubmitRegistrationForm(t, true, apiClient, publicTS, values(true),
					identity.CredentialsTypePassword, http.StatusBadRequest,
					publicTS.URL+password.RouteRegistration)

				assert.Contains(t, gjson.Get(body, "methods.password.config.messages.0.text").String(), "An account with the same identifier (email, phone, username, ...) exists already.", "%s", body)
			})

			t.Run("type=browser", func(t *testing.T) {
				_ = expectSuccessfulLogin(t, false, nil, values(false))
				body := expectValidationError(t, false, values(false))
				assert.Contains(t, gjson.Get(body, "methods.password.config.messages.0.text").String(), "An account with the same identifier (email, phone, username, ...) exists already.", "%s", body)
			})
		})

		t.Run("case=should return an error because not passing validation and reset previous errors and values", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")

			var check = func(t *testing.T, actual string) {
				assert.NotEmpty(t, gjson.Get(actual, "id").String(), "%s", actual)
				assert.Contains(t, gjson.Get(actual, "methods.password.config.action").String(), publicTS.URL+password.RouteRegistration, "%s", actual)
				checkFormContent(t, []byte(actual), "password", "csrf_token", "traits.username")
			}

			var checkFirst = func(t *testing.T, actual string) {
				check(t, actual)
				assert.Contains(t, gjson.Get(actual, "methods.password.config.fields.#(name==traits.username).messages.0").String(), `Property username is missing`, "%s", actual)
			}

			var checkSecond = func(t *testing.T, actual string) {
				check(t, actual)
				assert.EqualValues(t, "registration-identifier-9", gjson.Get(actual, "methods.password.config.fields.#(name==traits.username).value").String(), "%s", actual)
				assert.Empty(t, gjson.Get(actual, "methods.password.config.fields.#(name==traits.username).error").Raw)
				assert.Empty(t, gjson.Get(actual, "methods.password.config.error").Raw)
				assert.Contains(t, gjson.Get(actual, "methods.password.config.fields.#(name==traits.foobar).messages.0").String(), `Property foobar is missing`, "%s", actual)
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
				c := testhelpers.GetRegistrationFlowMethodConfig(t, f.Payload, identity.CredentialsTypePassword.String())

				actual, _ := testhelpers.RegistrationMakeRequest(t, true, c, apiClient, testhelpers.EncodeFormAsJSON(t, true, valuesFirst(testhelpers.SDKFormFieldsToURLValues(c.Fields))))
				checkFirst(t, actual)
				actual, _ = testhelpers.RegistrationMakeRequest(t, true, c, apiClient, testhelpers.EncodeFormAsJSON(t, true, valuesSecond(testhelpers.SDKFormFieldsToURLValues(c.Fields))))
				checkSecond(t, actual)
			})

			t.Run("type=browser", func(t *testing.T) {
				browserClient := testhelpers.NewClientWithCookies(t)
				f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS)
				c := testhelpers.GetRegistrationFlowMethodConfig(t, f.Payload, identity.CredentialsTypePassword.String())

				actual, _ := testhelpers.RegistrationMakeRequest(t, false, c, browserClient, valuesFirst(testhelpers.SDKFormFieldsToURLValues(c.Fields)).Encode())
				checkFirst(t, actual)
				actual, _ = testhelpers.RegistrationMakeRequest(t, false, c, browserClient, valuesSecond(testhelpers.SDKFormFieldsToURLValues(c.Fields)).Encode())
				checkSecond(t, actual)
			})
		})

		t.Run("case=should work even if password is just numbers", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://stub/registration.schema.json")
			viper.Set(configuration.HookStrategyKey(configuration.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()), []configuration.SelfServiceHook{{Name: "session"}})
			defer viper.Set(configuration.HookStrategyKey(configuration.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()), nil)

			var values = func(isAPI bool) func(v url.Values) {
				return func(v url.Values) {
					v.Set("traits.username", "registration-identifier-10-browser")
					if isAPI {
						v.Set("traits.username", "registration-identifier-10-api")
					}
					v.Set("password", x.NewUUID().String())
					v.Set("traits.foobar", "bar")
				}
			}

			t.Run("type=api", func(t *testing.T) {
				actual := expectSuccessfulLogin(t, true, nil, values(true))
				assert.Equal(t, `registration-identifier-10-api`, gjson.Get(actual, "identity.traits.username").String(), "%s", actual)
			})

			t.Run("type=browser", func(t *testing.T) {
				actual := expectSuccessfulLogin(t, false, nil, values(false))
				assert.Equal(t, `registration-identifier-10-browser`, gjson.Get(actual, "identity.traits.username").String(), "%s", actual)
			})
		})
	})

	t.Run("method=PopulateSignUpMethod", func(t *testing.T) {
		_, reg := internal.NewFastRegistryWithMocks(t)

		viper.Set(configuration.ViperKeyPublicBaseURL, urlx.ParseOrPanic("https://foo/"))
		viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://stub/registration.schema.json")
		viper.Set(configuration.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword), map[string]interface{}{
			"enabled": true})

		sr := registration.NewFlow(time.Minute, "nosurf", &http.Request{URL: urlx.ParseOrPanic("/")}, flow.TypeBrowser)
		require.NoError(t, reg.RegistrationStrategies().MustStrategy(identity.CredentialsTypePassword).(*password.Strategy).PopulateRegistrationMethod(&http.Request{}, sr))

		expected := &registration.FlowMethod{
			Method: identity.CredentialsTypePassword,
			Config: &registration.FlowMethodConfig{
				FlowMethodConfigurator: &password.FlowMethod{
					HTMLForm: &form.HTMLForm{
						Action: "https://foo" + password.RouteRegistration + "?flow=" + sr.ID.String(),
						Method: "POST",
						Fields: form.Fields{
							{
								Name:     "csrf_token",
								Type:     "hidden",
								Required: true,
								Value:    x.FakeCSRFToken,
							},
							{
								Name:     "password",
								Type:     "password",
								Required: true,
							},
							{
								Name: "traits.foobar",
								Type: "text",
							},
							{
								Name: "traits.username",
								Type: "text",
							},
						},
					},
				},
			},
		}

		actual := sr.Methods[identity.CredentialsTypePassword]
		assert.EqualValues(t, expected.Config.FlowMethodConfigurator.(*password.FlowMethod).HTMLForm, actual.Config.FlowMethodConfigurator.(*password.FlowMethod).HTMLForm)
	})
}
