package password_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/assertx"

	"github.com/ory/viper"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/selfservice/strategy/password"
	"github.com/ory/kratos/text"
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
		_, reg := internal.NewFastRegistryWithMocks(t)

		router := x.NewRouterPublic()
		admin := x.NewRouterAdmin()
		viper.Set(configuration.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword), map[string]interface{}{
			"enabled": true})

		publicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, admin)
		errTS := testhelpers.NewErrorTestServer(t, reg)
		uiTS := testhelpers.NewRegistrationUIFlowEchoServer(t, reg)
		redirTS := testhelpers.NewRedirSessionEchoTS(t, reg)

		// Overwrite these two to ensure that they run
		viper.Set(configuration.ViperKeySelfServiceBrowserDefaultReturnTo, redirTS.URL+"/default-return-to")
		viper.Set(configuration.ViperKeySelfServiceRegistrationAfter+"."+configuration.DefaultBrowserReturnURL, redirTS.URL+"/registration-return-ts")

		viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")

		var newRegistrationRequest = func(t *testing.T, exp time.Duration, isAPI bool) *registration.Flow {
			ft := flow.TypeBrowser
			if isAPI {
				ft = flow.TypeAPI
			}
			rr := &registration.Flow{
				ID:       x.NewUUID(),
				Type:     ft,
				IssuedAt: time.Now().UTC(), ExpiresAt: time.Now().UTC().Add(exp), RequestURL: publicTS.URL,
				Methods: map[identity.CredentialsType]*registration.FlowMethod{
					identity.CredentialsTypePassword: {
						Method: identity.CredentialsTypePassword,
						Config: &registration.FlowMethodConfig{
							FlowMethodConfigurator: password.RequestMethod{
								HTMLForm: &form.HTMLForm{
									Method: "POST",
									Action: "/action",
									Fields: form.Fields{
										{Name: "password", Type: "password", Required: true},
										{Name: "csrf_token", Type: "hidden", Required: true, Value: "csrf-token"},
									},
								},
							},
						},
					},
				},
			}
			require.NoError(t, reg.RegistrationFlowPersister().CreateRegistrationFlow(context.Background(), rr))
			return rr
		}

		var makeHttpRequest = func(t *testing.T, client *http.Client, req *http.Request, isAPI bool, expectedStatusCode int) ([]byte, *http.Response) {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Accept", "text/html")
			if isAPI {
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Accept", "application/json")
			}
			res, err := client.Do(req)
			require.NoError(t, err)

			result, err := ioutil.ReadAll(res.Body)
			require.NoError(t, res.Body.Close())
			require.NoError(t, err)

			require.EqualValues(t, expectedStatusCode, res.StatusCode, "Flow: %+v\n\t\tResponse: %+v\n\t\tResponse Headers: %+v\n\t\tBody: %s", res.Request, res, res.Header, result)
			return result, res
		}

		var makeRequestWithCookieJar = func(t *testing.T, rid uuid.UUID, isAPI bool, body string, expectedStatusCode int, jar *cookiejar.Jar) ([]byte, *http.Response) {
			client := &http.Client{Jar: jar}
			req, err := http.NewRequest("POST", publicTS.URL+password.RouteRegistration+"?flow="+rid.String(), strings.NewReader(body))
			require.NoError(t, err)
			return makeHttpRequest(t, client, req, isAPI, expectedStatusCode)
		}

		var makeRequest = func(t *testing.T, rid uuid.UUID, isAPI bool, body string, expectedStatusCode int) ([]byte, *http.Response) {
			jar, _ := cookiejar.New(&cookiejar.Options{})
			return makeRequestWithCookieJar(t, rid, isAPI, body, expectedStatusCode, jar)
		}

		t.Run("case=should show the error ui because the request payload is malformed", func(t *testing.T) {
			run := func(t *testing.T, isAPI bool) (*registration.Flow, []byte, *http.Response) {
				rr := newRegistrationRequest(t, time.Minute, isAPI)
				body, res := makeRequest(t, rr.ID, isAPI, "14=)=!(%)$/ZP()GHIÖ", expectStatusCode(isAPI, http.StatusBadRequest))
				return rr, body, res
			}

			t.Run("type=api", func(t *testing.T) {
				rr, body, res := run(t, true)
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+password.RouteRegistration)
				assert.Equal(t, rr.ID.String(), gjson.GetBytes(body, "id").String(), "%s", body)
				assert.Contains(t, string(body), `Expected JSON sent in request body to be an object but got: Number`)
			})

			t.Run("type=browser", func(t *testing.T) {
				rr, body, res := run(t, false)
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/registration-ts")
				assert.Equal(t, rr.ID.String(), gjson.GetBytes(body, "id").String(), "%s", body)
				assert.Contains(t, gjson.GetBytes(body, "methods.password.config.messages.0.text").String(), "invalid URL escape", "%s", body)
			})
		})

		t.Run("case=should show the error ui because the request id is missing", func(t *testing.T) {
			run := func(t *testing.T, isAPI bool) ([]byte, *http.Response) {
				_ = newRegistrationRequest(t, time.Minute, isAPI)
				uuidDesNotExistInStore := x.NewUUID()
				return makeRequest(t, uuidDesNotExistInStore, isAPI, "", expectStatusCode(isAPI, http.StatusNotFound))
			}

			t.Run("type=api", func(t *testing.T) {
				body, res := run(t, true)
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+password.RouteRegistration)
				assert.Equal(t, int64(http.StatusNotFound), gjson.GetBytes(body, "error.code").Int(), "%s", body)
				assert.Equal(t, "Not Found", gjson.GetBytes(body, "error.status").String(), "%s", body)
				assert.Contains(t, gjson.GetBytes(body, "error.message").String(), "Unable to locate the resource", "%s", body)
			})

			t.Run("type=browser", func(t *testing.T) {
				body, res := run(t, false)
				assert.Contains(t, res.Request.URL.String(), errTS.URL)
				assert.Equal(t, int64(http.StatusNotFound), gjson.GetBytes(body, "0.code").Int(), "%s", body)
				assert.Equal(t, "Not Found", gjson.GetBytes(body, "0.status").String(), "%s", body)
				assert.Contains(t, gjson.GetBytes(body, "0.message").String(), "Unable to locate the resource", "%s", body)
			})
		})

		t.Run("case=should return an error because the request is expired", func(t *testing.T) {
			run := func(t *testing.T, isAPI bool) *http.Response {
				rr := newRegistrationRequest(t, -time.Minute, isAPI)
				body, res := makeRequest(t, rr.ID, isAPI, "", http.StatusOK)
				assert.NotEqual(t, rr.ID.String(), gjson.GetBytes(body, "id").String(), "%s", body)
				assert.Contains(t, gjson.GetBytes(body, "messages.0.text").String(), "expired", "%s", body)
				return res
			}

			t.Run("type=api", func(t *testing.T) {
				res := run(t, true)
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+registration.RouteGetFlow)
			})

			t.Run("type=browser", func(t *testing.T) {
				res := run(t, false)
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/registration-ts")
			})
		})

		t.Run("case=should return an error because the password failed validation", func(t *testing.T) {
			run := func(t *testing.T, isAPI bool, payload string) *http.Response {
				rr := newRegistrationRequest(t, time.Minute, isAPI)
				body, res := makeRequest(t, rr.ID, isAPI, payload, expectStatusCode(isAPI, http.StatusBadRequest))
				assert.Equal(t, rr.ID.String(), gjson.GetBytes(body, "id").String(), "%s", body)
				assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String(), "%s", body)
				checkFormContent(t, body, "password", "csrf_token", "traits.username", "traits.foobar")
				assert.Contains(t, gjson.GetBytes(body, "methods.password.config.fields.#(name==password).messages.0").String(), "data breaches and must no longer be used.", "%s", body)
				return res
			}

			t.Run("type=api", func(t *testing.T) {
				res := run(t, true, `{"password":"password","traits.foobar":"bar","traits.username":"registration-identifier-4"}`)
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+password.RouteRegistration)
			})

			t.Run("type=browser", func(t *testing.T) {
				res := run(t, false, url.Values{
					"traits.username": {"registration-identifier-4"},
					"password":        {"password"},
					"traits.foobar":   {"bar"},
				}.Encode())
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/registration-ts")
			})
		})

		t.Run("case=should return an error because not passing validation", func(t *testing.T) {
			run := func(t *testing.T, isAPI bool, payload string) *http.Response {
				rr := newRegistrationRequest(t, time.Minute, isAPI)
				body, res := makeRequest(t, rr.ID, isAPI, payload, expectStatusCode(isAPI, http.StatusBadRequest))
				assert.Equal(t, rr.ID.String(), gjson.GetBytes(body, "id").String(), "%s", body)
				assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String(), "%s", body)
				checkFormContent(t, body, "password", "csrf_token", "traits.username", "traits.foobar")
				assert.Contains(t, gjson.GetBytes(body, "methods.password.config.fields.#(name==traits.foobar).messages.0").String(), `Property foobar is missing`, "%s", body)
				return res
			}

			t.Run("type=api", func(t *testing.T) {
				res := run(t, true, `{"password":"c0a5af7a-fa32-4fe1-85b9-3beb4a127164","traits.username":"registration-identifier-5"}`)
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+password.RouteRegistration)
			})

			t.Run("type=browser", func(t *testing.T) {
				res := run(t, false, url.Values{
					"traits.username": {"registration-identifier-5"},
					"password":        {x.NewUUID().String()},
				}.Encode())
				assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/registration-ts")
			})
		})

		t.Run("case=should fail because schema did not specify an identifier", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/missing-identifier.schema.json")
			run := func(t *testing.T, isAPI bool, payload string) ([]byte, *http.Response) {
				rr := newRegistrationRequest(t, time.Minute, isAPI)
				return makeRequest(t, rr.ID, isAPI, payload, expectStatusCode(isAPI, http.StatusInternalServerError))
			}

			t.Run("type=api", func(t *testing.T) {
				body, res := run(t, true, `{"password":"c0a5af7a-fa32-4fe1-85b9-3beb4a127164","traits.username":"registration-identifier-6","traits.foobar":"bar"}`)
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+password.RouteRegistration)
				assert.Equal(t, int64(http.StatusInternalServerError), gjson.GetBytes(body, "error.code").Int(), "%s", body)
				assert.Equal(t, "Internal Server Error", gjson.GetBytes(body, "error.status").String(), "%s", body)
				assert.Contains(t, gjson.GetBytes(body, "error.reason").String(), "No login identifiers", "%s", body)
			})

			t.Run("type=browser", func(t *testing.T) {
				body, res := run(t, false, url.Values{
					"traits.username": {"registration-identifier-6"},
					"password":        {x.NewUUID().String()},
					"traits.foobar":   {"bar"},
				}.Encode())
				assert.Contains(t, res.Request.URL.String(), errTS.URL)
				assert.Equal(t, int64(http.StatusInternalServerError), gjson.GetBytes(body, "0.code").Int(), "%s", body)
				assert.Equal(t, "Internal Server Error", gjson.GetBytes(body, "0.status").String(), "%s", body)
				assert.Contains(t, gjson.GetBytes(body, "0.reason").String(), "No login identifiers", "%s", body)
			})
		})

		t.Run("case=should fail because schema does not exist", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/i-do-not-exist.schema.json")

			run := func(t *testing.T, isAPI bool, payload string) ([]byte, *http.Response) {
				rr := newRegistrationRequest(t, time.Minute, isAPI)
				body, res := makeRequest(t, rr.ID, isAPI, payload, expectStatusCode(isAPI, http.StatusInternalServerError))
				return body, res
			}

			t.Run("type=api", func(t *testing.T) {
				body, res := run(t, true, `{"password":"c0a5af7a-fa32-4fe1-85b9-3beb4a127164","traits.username":"registration-identifier-7","traits.foobar":"bar"}`)
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+password.RouteRegistration)
				assert.Equal(t, int64(http.StatusInternalServerError), gjson.GetBytes(body, "error.code").Int(), "%s", body)
				assert.Equal(t, "Internal Server Error", gjson.GetBytes(body, "error.status").String(), "%s", body)
				assert.Contains(t, gjson.GetBytes(body, "error.message").String(), "no such file or directory", "%s", body)
			})

			t.Run("type=browser", func(t *testing.T) {
				body, res := run(t, false, url.Values{
					"traits.username": {"registration-identifier-7"},
					"password":        {x.NewUUID().String()},
					"traits.foobar":   {"bar"},
				}.Encode())
				assert.Contains(t, res.Request.URL.String(), errTS.URL)
				assert.Equal(t, int64(http.StatusInternalServerError), gjson.GetBytes(body, "0.code").Int(), "%s", body)
				assert.Equal(t, "Internal Server Error", gjson.GetBytes(body, "0.status").String(), "%s", body)
				assert.Contains(t, gjson.GetBytes(body, "0.reason").String(), "no such file or directory", "%s", body)
			})
		})

		t.Run("case=should pass and set up a session", func(t *testing.T) {
			run := func(t *testing.T, isAPI bool, payload string) ([]byte, *http.Response) {
				viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")
				viper.Set(
					configuration.HookStrategyKey(configuration.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()),
					[]configuration.SelfServiceHook{{Name: "session"}})

				rr := newRegistrationRequest(t, time.Minute, isAPI)
				body, res := makeRequest(t, rr.ID, isAPI, payload, http.StatusOK)
				return body, res
			}

			t.Run("type=api", func(t *testing.T) {
				body, res := run(t, true, `{"password":"c0a5af7a-fa32-4fe1-85b9-3beb4a127164","traits.username":"registration-identifier-8-api","traits.foobar":"bar"}`)
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+password.RouteRegistration)
				assert.Equal(t, `registration-identifier-8-api`, gjson.GetBytes(body, "identity.traits.username").String(), "%s", body)
				assert.NotEmpty(t, gjson.GetBytes(body, "session_token").String(), "%s", body)
				assert.NotEmpty(t, gjson.GetBytes(body, "session.id").String(), "%s", body)
			})

			t.Run("type=browser", func(t *testing.T) {
				body, res := run(t, false, url.Values{
					"traits.username": {"registration-identifier-8-browser"},
					"password":        {x.NewUUID().String()},
					"traits.foobar":   {"bar"},
				}.Encode())
				assert.Contains(t, res.Request.URL.String(), redirTS.URL+"/registration-return-ts")
				assert.Equal(t, `registration-identifier-8-browser`, gjson.GetBytes(body, "identity.traits.username").String(), "%s", body)
			})
		})

		t.Run("case=should fail to register the same user again", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")
			run := func(t *testing.T, isAPI bool, payload string, sc int) ([]byte, *http.Response) {
				rr := newRegistrationRequest(t, time.Minute, isAPI)
				body, res := makeRequest(t, rr.ID, isAPI, payload, sc)
				return body, res
			}

			t.Run("type=api", func(t *testing.T) {
				_, _ = run(t, true, `{"password":"c0a5af7a-fa32-4fe1-85b9-3beb4a127164","traits.username":"registration-identifier-8-api-duplicate","traits.foobar":"bar"}`,http.StatusOK)

				body, res := run(t, true, `{"password":"c0a5af7a-fa32-4fe1-85b9-3beb4a127164","traits.username":"registration-identifier-8-api-duplicate","traits.foobar":"bar"}`,http.StatusBadRequest)
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+password.RouteRegistration)
				assert.Contains(t, gjson.GetBytes(body, "methods.password.config.messages.0.text").String(), "An account with the same identifier (email, phone, username, ...) exists already.", "%s", body)
			})

			t.Run("type=browser", func(t *testing.T) {
				_, _ = run(t, false, url.Values{
					"traits.username": {"registration-identifier-8-browser-duplicate"},
					"password":        {x.NewUUID().String()},
					"traits.foobar":   {"bar"},
				}.Encode(), http.StatusOK)

				body, res := run(t, false, url.Values{
					"traits.username": {"registration-identifier-8-browser-duplicate"},
					"password":        {x.NewUUID().String()},
					"traits.foobar":   {"bar"},
				}.Encode(), http.StatusOK)
				assert.Contains(t, res.Request.URL.Path, "registration-ts")
				assert.Contains(t, gjson.GetBytes(body, "methods.password.config.messages.0.text").String(), "An account with the same identifier (email, phone, username, ...) exists already.", "%s", body)
			})
		})

		t.Run("case=should return an error because not passing validation and reset previous errors and values", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")
			run := func(t *testing.T, isAPI bool, payload string) *http.Response {
				ft := flow.TypeBrowser
				if isAPI {
					ft = flow.TypeAPI
				}
				rr := &registration.Flow{
					ID:        x.NewUUID(),
					ExpiresAt: time.Now().Add(time.Minute),
					Type:      ft,
					Methods: map[identity.CredentialsType]*registration.FlowMethod{
						identity.CredentialsTypePassword: {
							Method: identity.CredentialsTypePassword,
							Config: &registration.FlowMethodConfig{
								FlowMethodConfigurator: &password.RequestMethod{
									HTMLForm: &form.HTMLForm{
										Method:   "POST",
										Action:   "/action",
										Messages: text.Messages{{Text: "some error"}},
										Fields: form.Fields{
											{
												Name: "traits.foo", Value: "bar", Type: "text",
												Messages: text.Messages{{Text: "bar"}},
											},
											{Name: "password"},
										},
									},
								},
							},
						},
					},
				}

				require.NoError(t, reg.RegistrationFlowPersister().CreateRegistrationFlow(context.Background(), rr))
				body, res := makeRequest(t, rr.ID, isAPI, payload, expectStatusCode(isAPI, http.StatusBadRequest))

				assert.Equal(t, rr.ID.String(), gjson.GetBytes(body, "id").String(), "%s", body)
				assert.Equal(t, "/action", gjson.GetBytes(body, "methods.password.config.action").String(), "%s", body)
				checkFormContent(t, body, "password", "csrf_token", "traits.username")

				assert.Empty(t, gjson.GetBytes(body, "methods.password.config.fields.#(name==traits.foo).value"), "%s", body)
				assert.Empty(t, gjson.GetBytes(body, "methods.password.config.fields.#(name==traits.foo).error"))
				assert.Empty(t, gjson.GetBytes(body, "methods.password.config.error"))
				assert.Contains(t, gjson.GetBytes(body, "methods.password.config.fields.#(name==traits.foobar).messages.0").String(), `Property foobar is missing`, "%s", body)

				return res
			}

			t.Run("type=api", func(t *testing.T) {
				res := run(t, true, `{"password":"c0a5af7a-fa32-4fe1-85b9-3beb4a127164","traits.username":"registration-identifier-9"}`)
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+password.RouteRegistration)
			})

			t.Run("type=browser", func(t *testing.T) {
				res := run(t, false, url.Values{
					"traits.username": {"registration-identifier-9"},
					"password":        {x.NewUUID().String()},
				}.Encode())
				assert.Contains(t, res.Request.URL.Path, "registration-ts")
			})
		})

		t.Run("case=should work even if password is just numbers", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://stub/registration.schema.json")
			viper.Set(
				configuration.HookStrategyKey(configuration.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()),
				[]configuration.SelfServiceHook{{Name: "session"}})

			run := func(t *testing.T, isAPI bool, payload string) ([]byte, *http.Response) {
				rr := newRegistrationRequest(t, time.Minute, isAPI)
				return makeRequest(t, rr.ID, isAPI, payload, http.StatusOK)
			}

			t.Run("type=api", func(t *testing.T) {
				body, res := run(t, true, `{"password":"93172388957812344432","traits.username":"registration-identifier-10-api","traits.foobar":"bar"}`)
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+password.RouteRegistration)
				assert.Equal(t, `registration-identifier-10-api`, gjson.GetBytes(body, "identity.traits.username").String(), "%s", body)
			})

			t.Run("type=browser", func(t *testing.T) {
				body, res := run(t, false, url.Values{
					"traits.username": {"registration-identifier-10-browser"},
					"password":        {"93172388957812344432"},
					"traits.foobar":   {"bar"},
				}.Encode())
				assert.Equal(t, res.Request.URL.String(), redirTS.URL+"/registration-return-ts")
				assert.Equal(t, `registration-identifier-10-browser`, gjson.GetBytes(body, "identity.traits.username").String(), "%s", body)
			})
		})

		t.Run("case=register and then send same request", func(t *testing.T) {
			viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://stub/registration.schema.json")
			viper.Set(
				configuration.HookStrategyKey(configuration.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()),
				[]configuration.SelfServiceHook{{Name: "session"}})

			t.Run("type=api", func(t *testing.T) {
				payload := `{"password":"O(lf<ys87LÖ:(h<dsjfl","traits.username":"registration-identifier-11","traits.foobar":"bar"}`

				body1, res1 := makeRequest(t, newRegistrationRequest(t, time.Minute, true).ID,
					true, payload, http.StatusOK)
				sessionToken := gjson.GetBytes(body1, "session_token").String()
				assert.NotEmpty(t, sessionToken)

				rid := newRegistrationRequest(t, time.Minute, true).ID
				req, err := http.NewRequest("POST", publicTS.URL+password.RouteRegistration+"?flow="+rid.String(),
					strings.NewReader(payload))
				require.NoError(t, err)
				req.Header.Set("Authorization", "Bearer "+sessionToken)
				body2, res2 := makeHttpRequest(t, new(http.Client), req, true, http.StatusBadRequest)

				assert.Contains(t, res1.Request.URL.String(), publicTS.URL+password.RouteRegistration)
				assert.Contains(t, res2.Request.URL.String(), publicTS.URL+password.RouteRegistration)
				assert.NotEmpty(t, gjson.GetBytes(body1, "session_token").String())
				assertx.EqualAsJSON(t, registration.ErrAlreadyLoggedIn, json.RawMessage(gjson.GetBytes(body2, "error").Raw))
			})

			t.Run("type=browser", func(t *testing.T) {
				jar, _ := cookiejar.New(&cookiejar.Options{})
				payload := url.Values{
					"traits.username": {"registration-identifier-11-browser"},
					"password":        {"O(lf<ys87LÖ:(h<dsjfl"},
					"traits.foobar":   {"bar"},
				}.Encode()

				body1, res1 := makeRequestWithCookieJar(t, newRegistrationRequest(t, time.Minute, false).ID,
					false, payload, http.StatusOK, jar)

				body2, res2 := makeRequestWithCookieJar(t, newRegistrationRequest(t, time.Minute, false).ID,
					false, payload, expectStatusCode(false, http.StatusBadRequest), jar)

				assert.Contains(t, res1.Request.URL.String(), redirTS.URL+"/registration-return-ts")
				assert.Contains(t, res2.Request.URL.String(), redirTS.URL+"/default-return-to")
				assert.Equal(t, body1, body2)
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
				FlowMethodConfigurator: &password.RequestMethod{
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
		assert.EqualValues(t, expected.Config.FlowMethodConfigurator.(*password.RequestMethod).HTMLForm, actual.Config.FlowMethodConfigurator.(*password.RequestMethod).HTMLForm)
	})
}
