// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package password_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/internal/registrationhelpers"

	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/ui/container"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/x/assertx"

	_ "embed"

	"github.com/ory/kratos/x"
)

var flows = []string{"spa", "api", "browser"}

//go:embed stub/registration.schema.json
var registrationSchema []byte

func newRegistrationRegistry(t *testing.T) *driver.RegistryDefault {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword), map[string]interface{}{"enabled": true})
	return reg
}

func TestRegistration(t *testing.T) {
	ctx := context.Background()
	t.Run("case=registration", func(t *testing.T) {
		reg := newRegistrationRegistry(t)
		conf := reg.Config()

		router := x.NewRouterPublic()
		admin := x.NewRouterAdmin()

		publicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, admin)
		_ = testhelpers.NewErrorTestServer(t, reg)
		_ = testhelpers.NewRegistrationUIFlowEchoServer(t, reg)
		redirTS := testhelpers.NewRedirSessionEchoTS(t, reg)
		redirNoSessionTS := testhelpers.NewRedirNoSessionTS(t, reg)

		// set the "return to" server, which will assert the session state
		// (redirTS: enforce that a session exists, redirNoSessionTS: enforce that no session exists)
		var useReturnToFromTS = func(ts *httptest.Server) {
			conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, ts.URL+"/default-return-to")
			conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationAfter+"."+config.DefaultBrowserReturnURL, ts.URL+"/registration-return-ts")
		}

		useReturnToFromTS(redirTS)
		testhelpers.SetDefaultIdentitySchemaFromRaw(conf, registrationSchema)

		apiClient := testhelpers.NewDebugClient(t)

		t.Run("AssertCommonErrorCases", func(t *testing.T) {
			reg := newRegistrationRegistry(t)
			registrationhelpers.AssertCommonErrorCases(t, reg, flows)
		})

		t.Run("AssertRegistrationRespectsValidation", func(t *testing.T) {
			reg := newRegistrationRegistry(t)
			registrationhelpers.AssertRegistrationRespectsValidation(t, reg, flows, func(v url.Values) {
				v.Del("traits.foobar")
			})
		})

		t.Run("AssertCSRFFailures", func(t *testing.T) {
			reg := newRegistrationRegistry(t)
			registrationhelpers.AssertCSRFFailures(t, reg, flows, func(v url.Values) {
				v.Set("password", x.NewUUID().String())
				v.Set("method", identity.CredentialsTypePassword.String())
			})
		})

		t.Run("AssertSchemDoesNotExist", func(t *testing.T) {
			reg := newRegistrationRegistry(t)
			registrationhelpers.AssertSchemDoesNotExist(t, reg, flows, func(v url.Values) {
				v.Set("password", x.NewUUID().String())
				v.Set("method", identity.CredentialsTypePassword.String())
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
			testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/registration.schema.json")
			conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRegistrationPostPersist, identity.CredentialsTypePassword.String()), []config.SelfServiceHook{{Name: "session"}})
			t.Cleanup(func() {
				conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRegistrationPostPersist, identity.CredentialsTypePassword.String()), nil)
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
			testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/registration.schema.json")
			conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRegistrationPostPersist, identity.CredentialsTypePassword.String()), nil)

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
			testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/registration.schema.json")
			conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRegistrationPostPersist, identity.CredentialsTypePassword.String()), []config.SelfServiceHook{{Name: "session"}})
			t.Cleanup(func() {
				conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRegistrationPostPersist, identity.CredentialsTypePassword.String()), nil)
			})

			var applyTransform = func(values, transform func(v url.Values)) func(v url.Values) {
				return func(v url.Values) {
					values(v)
					transform(v)
				}
			}

			// test duplicate registration on all client types, where the values can be transformed before
			// they are sent the second time
			var testWithTransform = func(t *testing.T, suffix string, transform func(v url.Values)) {
				t.Run("type=api", func(t *testing.T) {
					values := func(v url.Values) {
						v.Set("traits.username", "registration-identifier-8-api-duplicate-"+suffix)
						v.Set("password", x.NewUUID().String())
						v.Set("traits.foobar", "bar")
					}

					_ = expectSuccessfulLogin(t, true, false, apiClient, values)
					body := testhelpers.SubmitRegistrationForm(t, true, apiClient, publicTS,
						applyTransform(values, transform), false, http.StatusBadRequest,
						publicTS.URL+registration.RouteSubmitFlow)
					assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "An account with the same identifier (email, phone, username, ...) exists already.", "%s", body)
				})

				t.Run("type=spa", func(t *testing.T) {
					values := func(v url.Values) {
						v.Set("traits.username", "registration-identifier-8-spa-duplicate-"+suffix)
						v.Set("password", x.NewUUID().String())
						v.Set("traits.foobar", "bar")
					}

					_ = expectSuccessfulLogin(t, false, true, nil, values)
					body := registrationhelpers.ExpectValidationError(t, publicTS, conf, "spa", applyTransform(values, transform))
					assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "An account with the same identifier (email, phone, username, ...) exists already.", "%s", body)
				})

				t.Run("type=browser", func(t *testing.T) {
					values := func(v url.Values) {
						v.Set("traits.username", "registration-identifier-8-browser-duplicate-"+suffix)
						v.Set("password", x.NewUUID().String())
						v.Set("traits.foobar", "bar")
					}

					_ = expectSuccessfulLogin(t, false, false, nil, values)
					body := registrationhelpers.ExpectValidationError(t, publicTS, conf, "browser", applyTransform(values, transform))
					assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "An account with the same identifier (email, phone, username, ...) exists already.", "%s", body)
				})
			}

			t.Run("case=identical input", func(t *testing.T) {
				testWithTransform(t, "identical", func(v url.Values) {
					// base case
				})
			})

			t.Run("case=different capitalization", func(t *testing.T) {
				testWithTransform(t, "caps", func(v url.Values) {
					v.Set("traits.username", strings.ToUpper(v.Get("traits.username")))
				})
			})

			t.Run("case=leading whitespace", func(t *testing.T) {
				testWithTransform(t, "leading", func(v url.Values) {
					v.Set("traits.username", "  "+v.Get("traits.username"))
				})
			})

			t.Run("case=trailing whitespace", func(t *testing.T) {
				testWithTransform(t, "trailing", func(v url.Values) {
					v.Set("traits.username", v.Get("traits.username")+"  ")
				})
			})
		})

		t.Run("case=should return an error because not passing validation and reset previous errors and values", func(t *testing.T) {
			testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/registration.schema.json")

			var check = func(t *testing.T, actual string) {
				assert.NotEmpty(t, gjson.Get(actual, "id").String(), "%s", actual)
				assert.Contains(t, gjson.Get(actual, "ui.action").String(), publicTS.URL+registration.RouteSubmitFlow, "%s", actual)
				registrationhelpers.CheckFormContent(t, []byte(actual), "password", "csrf_token", "traits.username")
			}

			var checkFirst = func(t *testing.T, actual string) {
				check(t, actual)
				assert.Contains(t, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.username).messages.0").String(), `Property username is missing`, "%s", actual)
			}

			var checkSecond = func(t *testing.T, actual string) {
				check(t, actual)
				assert.EqualValues(t, "registration-identifier-9", gjson.Get(actual, "ui.nodes.#(attributes.name==traits.username).attributes.value").String(), "%s", actual)
				assert.Empty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.username).messages").Array())
				assert.Empty(t, gjson.Get(actual, "ui.nodes.messages").Array())
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
				f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS, true, false, false)
				c := f.Ui

				actual, _ := testhelpers.RegistrationMakeRequest(t, false, true, f, browserClient, testhelpers.EncodeFormAsJSON(t, false, valuesFirst(testhelpers.SDKFormFieldsToURLValues(c.Nodes))))
				checkFirst(t, actual)
				actual, _ = testhelpers.RegistrationMakeRequest(t, false, true, f, browserClient, testhelpers.EncodeFormAsJSON(t, false, valuesSecond(testhelpers.SDKFormFieldsToURLValues(c.Nodes))))
				checkSecond(t, actual)
			})

			t.Run("type=browser", func(t *testing.T) {
				browserClient := testhelpers.NewClientWithCookies(t)
				f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS, false, false, false)
				c := f.Ui

				actual, _ := testhelpers.RegistrationMakeRequest(t, false, false, f, browserClient, valuesFirst(testhelpers.SDKFormFieldsToURLValues(c.Nodes)).Encode())
				checkFirst(t, actual)
				actual, _ = testhelpers.RegistrationMakeRequest(t, false, false, f, browserClient, valuesSecond(testhelpers.SDKFormFieldsToURLValues(c.Nodes)).Encode())
				checkSecond(t, actual)
			})
		})

		t.Run("case=should work even if password is just numbers", func(t *testing.T) {
			testhelpers.SetDefaultIdentitySchema(conf, "file://stub/registration.schema.json")
			conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRegistrationPostPersist, identity.CredentialsTypePassword.String()), []config.SelfServiceHook{{Name: "session"}})
			t.Cleanup(func() {
				conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRegistrationPostPersist, identity.CredentialsTypePassword.String()), nil)
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

		t.Run("case=should fail if no identifier was set in the schema", func(t *testing.T) {
			testhelpers.SetDefaultIdentitySchema(conf, "file://stub/missing-identifier.schema.json")

			for _, f := range []string{"spa", "api", "browser"} {
				t.Run("type="+f, func(t *testing.T) {
					actual := registrationhelpers.ExpectValidationError(t, publicTS, conf, f, func(v url.Values) {
						v.Set("traits.email", testhelpers.RandomEmail())
						v.Set("password", x.NewUUID().String())
						v.Set("method", "password")
					})
					assert.Equal(t, text.NewErrorValidationIdentifierMissing().Text, gjson.Get(actual, "ui.messages.0.text").String(), "%s", actual)
				})
			}
		})

		t.Run("case=should work with regular JSON", func(t *testing.T) {
			testhelpers.SetDefaultIdentitySchema(conf, "file://stub/registration.schema.json")
			conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRegistrationPostPersist, identity.CredentialsTypePassword.String()), []config.SelfServiceHook{{Name: "session"}})
			t.Cleanup(func() {
				conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRegistrationPostPersist, identity.CredentialsTypePassword.String()), nil)
			})

			hc := testhelpers.NewClientWithCookies(t)
			hc.Transport = testhelpers.NewTransportWithLogger(hc.Transport, t)
			payload := testhelpers.InitializeRegistrationFlowViaBrowser(t, hc, publicTS, false, false, false)
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

		t.Run("case=should choose the correct identity schema", func(t *testing.T) {
			conf.MustSet(ctx, config.ViperKeyDefaultIdentitySchemaID, "advanced-user")
			conf.MustSet(ctx, config.ViperKeyIdentitySchemas, config.Schemas{
				{ID: "does-not-exist", URL: "file://./stub/not-exists.schema.json"},
				{ID: "advanced-user", URL: "file://./stub/registration.secondary.schema.json"},
			})
			conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRegistrationPostPersist, identity.CredentialsTypePassword.String()), nil)

			username := "registration-custom-schema"
			t.Run("type=api", func(t *testing.T) {
				body := expectNoLogin(t, true, false, nil, func(v url.Values) {
					v.Set("traits.username", username+"-api")
					v.Set("password", x.NewUUID().String())
					v.Set("traits.baz", "bar")
				})
				assert.Equal(t, username+"-api", gjson.Get(body, "identity.traits.username").String(), "%s", body)
				assert.Empty(t, gjson.Get(body, "session_token").String(), "%s", body)
				assert.Empty(t, gjson.Get(body, "session.id").String(), "%s", body)
			})

			t.Run("type=spa", func(t *testing.T) {
				expectNoLogin(t, false, true, nil, func(v url.Values) {
					v.Set("traits.username", username+"-spa")
					v.Set("password", x.NewUUID().String())
					v.Set("traits.baz", "bar")
				})
			})

			t.Run("type=browser", func(t *testing.T) {
				expectNoLogin(t, false, false, nil, func(v url.Values) {
					v.Set("traits.username", username+"-browser")
					v.Set("password", x.NewUUID().String())
					v.Set("traits.baz", "bar")
				})
			})
		})
	})

	t.Run("method=PopulateSignUpMethod", func(t *testing.T) {
		conf, reg := internal.NewFastRegistryWithMocks(t)

		conf.MustSet(ctx, config.ViperKeyPublicBaseURL, "https://foo/")
		testhelpers.SetDefaultIdentitySchema(conf, "file://stub/sort.schema.json")
		conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword)+".enabled", true)

		router := x.NewRouterPublic()
		publicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin())
		_ = testhelpers.NewRegistrationUIFlowEchoServer(t, reg)

		browserClient := testhelpers.NewClientWithCookies(t)
		f := testhelpers.InitializeRegistrationFlowViaBrowser(t, browserClient, publicTS, false, false, false)

		assertx.EqualAsJSON(t, container.Container{
			Action: conf.SelfPublicURL(ctx).String() + registration.RouteSubmitFlow + "?flow=" + f.Id,
			Method: "POST",
			Nodes: node.Nodes{
				node.NewCSRFNode(x.FakeCSRFToken),
				node.NewInputField("traits.username", nil, node.PasswordGroup, node.InputAttributeTypeText),
				node.NewInputField("password", nil, node.PasswordGroup, node.InputAttributeTypePassword, node.WithRequiredInputAttribute, node.WithInputAttributes(func(a *node.InputAttributes) {
					a.Autocomplete = node.InputAttributeAutocompleteNewPassword
				})).WithMetaLabel(text.NewInfoNodeInputPassword()),
				node.NewInputField("traits.bar", nil, node.PasswordGroup, node.InputAttributeTypeText),
				node.NewInputField("method", "password", node.PasswordGroup, node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoRegistration()),
			},
		}, f.Ui)
	})
}
