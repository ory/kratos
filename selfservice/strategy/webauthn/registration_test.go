package webauthn_test

import (
	"context"
	_ "embed"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	kratos "github.com/ory/kratos-client-go"
	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/registrationhelpers"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/strategy/webauthn"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/assertx"
)

var (
	flows = []string{"spa", "browser"}
	//go:embed fixtures/registration/success/identity.json
	registrationFixtureSuccessIdentity []byte
	//go:embed fixtures/registration/success/response.json
	registrationFixtureSuccessResponse []byte
	//go:embed fixtures/registration/success/internal_context.json
	registrationFixtureSuccessInternalContext []byte
)

func flowToIsSPA(flow string) bool {
	return flow == "spa"
}

func newRegistrationRegistry(t *testing.T) *driver.RegistryDefault {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword)+".enabled", true)
	enableWebAuthn(conf)
	conf.MustSet(config.ViperKeyWebAuthnPasswordless, true)
	return reg
}

func TestRegistration(t *testing.T) {
	reg := newRegistrationRegistry(t)
	conf := reg.Config(context.Background())

	router := x.NewRouterPublic()
	publicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin())

	_ = testhelpers.NewErrorTestServer(t, reg)
	_ = testhelpers.NewRegistrationUIFlowEchoServer(t, reg)
	_ = testhelpers.NewRedirSessionEchoTS(t, reg)

	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/registration.schema.json")
	conf.MustSet(config.ViperKeySecretsDefault, []string{"not-a-secure-session-key"})

	redirTS := testhelpers.NewRedirSessionEchoTS(t, reg)
	redirNoSessionTS := testhelpers.NewRedirNoSessionTS(t, reg)

	// set the "return to" server, which will assert the session state
	// (redirTS: enforce that a session exists, redirNoSessionTS: enforce that no session exists)
	var useReturnToFromTS = func(ts *httptest.Server) {
		conf.MustSet(config.ViperKeySelfServiceBrowserDefaultReturnTo, ts.URL+"/default-return-to")
		conf.MustSet(config.ViperKeySelfServiceRegistrationAfter+"."+config.DefaultBrowserReturnURL, ts.URL+"/registration-return-ts")
	}
	useReturnToFromTS(redirTS)

	//checkURL := func(t *testing.T, shouldRedirect bool, res *http.Response) {
	//	if shouldRedirect {
	//		assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/registration-ts")
	//	} else {
	//		assert.Contains(t, res.Request.URL.String(), publicTS.URL+registration.RouteSubmitFlow)
	//	}
	//}

	t.Run("AssertCommonErrorCases", func(t *testing.T) {
		reg := newRegistrationRegistry(t)
		registrationhelpers.AssertCommonErrorCases(t, reg, flows)
	})

	t.Run("AssertRegistrationRespectsValidation", func(t *testing.T) {
		reg := newRegistrationRegistry(t)
		registrationhelpers.AssertRegistrationRespectsValidation(t, reg, flows, func(v url.Values) {
			v.Del("traits.foobar")
			v.Set(node.WebAuthnRegister, "{}")
			v.Del("method")
		})
	})

	t.Run("AssertCSRFFailures", func(t *testing.T) {
		reg := newRegistrationRegistry(t)
		registrationhelpers.AssertCSRFFailures(t, reg, flows, func(v url.Values) {
			v.Set(node.WebAuthnRegister, "{}")
			v.Del("method")
		})
	})

	t.Run("AssertSchemDoesNotExist", func(t *testing.T) {
		reg := newRegistrationRegistry(t)
		registrationhelpers.AssertSchemDoesNotExist(t, reg, flows, func(v url.Values) {
			v.Set(node.WebAuthnRegister, "{}")
			v.Del("method")
		})
	})

	t.Run("case=webauthn button does not exist when passwordless is disabled", func(t *testing.T) {
		conf.MustSet(config.ViperKeyWebAuthnPasswordless, false)
		t.Cleanup(func() {
			conf.MustSet(config.ViperKeyWebAuthnPasswordless, true)
		})
		for _, f := range flows {
			t.Run(f, func(t *testing.T) {
				client := testhelpers.NewClientWithCookies(t)
				f := testhelpers.InitializeRegistrationFlowViaBrowser(t, client, publicTS, flowToIsSPA(f))
				testhelpers.SnapshotTExcept(t, f.Ui.Nodes, []string{
					"0.attributes.value",
				})
			})
		}
	})

	t.Run("case=webauthn button exists", func(t *testing.T) {
		for _, f := range flows {
			t.Run(f, func(t *testing.T) {
				client := testhelpers.NewClientWithCookies(t)
				f := testhelpers.InitializeRegistrationFlowViaBrowser(t, client, publicTS, flowToIsSPA(f))
				testhelpers.SnapshotTExcept(t, f.Ui.Nodes, []string{
					"2.attributes.value",
					"5.attributes.onclick",
					"6.attributes.nonce",
					"6.attributes.src",
				})
			})
		}
	})

	t.Run("case=should return an error because not passing validation", func(t *testing.T) {
		email := testhelpers.RandomEmail()

		var values = func(v url.Values) {
			v.Set("traits.username", email)
			v.Del("traits.foobar")
			v.Set(node.WebAuthnRegister, "{}")
			v.Del("method")
		}

		for _, f := range flows {
			t.Run("type="+f, func(t *testing.T) {
				actual := registrationhelpers.ExpectValidationError(t, publicTS, conf, f, values)

				assert.NotEmpty(t, gjson.Get(actual, "id").String(), "%s", actual)
				assert.Contains(t, gjson.Get(actual, "ui.action").String(), publicTS.URL+registration.RouteSubmitFlow, "%s", actual)
				registrationhelpers.CheckFormContent(t, []byte(actual), node.WebAuthnRegisterTrigger, "csrf_token", "traits.username", "traits.foobar")
				assert.Contains(t, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.foobar).messages.0").String(), `Property foobar is missing`, "%s", actual)
				assert.Equal(t, email, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.username).attributes.value").String(), "%s", actual)
			})
		}
	})

	t.Run("case=should return an error because webauthn response is invalid", func(t *testing.T) {
		email := testhelpers.RandomEmail()
		var values = func(v url.Values) {
			v.Set("traits.username", email)
			v.Set("traits.foobar", "bazbar")
			v.Set(node.WebAuthnRegister, "{}")
			v.Del("method")
		}

		for _, f := range flows {
			t.Run("type="+f, func(t *testing.T) {
				actual := registrationhelpers.ExpectValidationError(t, publicTS, conf, f, values)
				assert.NotEmpty(t, gjson.Get(actual, "id").String(), "%s", actual)
				assert.Contains(t, gjson.Get(actual, "ui.action").String(), publicTS.URL+registration.RouteSubmitFlow, "%s", actual)
				registrationhelpers.CheckFormContent(t, []byte(actual), node.WebAuthnRegisterTrigger, "csrf_token", "traits.username", "traits.foobar")
				assert.Equal(t, "bazbar", gjson.Get(actual, "ui.nodes.#(attributes.name==traits.foobar).attributes.value").String(), "%s", actual)
				assert.Equal(t, email, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.username).attributes.value").String(), "%s", actual)
				assert.Contains(t, gjson.Get(actual, "ui.messages.0").String(), `Unable to parse WebAuthn response: Parse error for Registration`, "%s", actual)
			})
		}
	})

	submitWebAuthnRegistrationWithClient := func(t *testing.T, flow string, contextFixture []byte, client *http.Client, cb func(values url.Values), opts ...testhelpers.InitFlowWithOption) (string, *http.Response, *kratos.SelfServiceRegistrationFlow) {
		isSPA := flow == "spa"
		f := testhelpers.InitializeRegistrationFlowViaBrowser(t, client, publicTS, isSPA, opts...)

		// We inject the session to replay
		interim, err := reg.RegistrationFlowPersister().GetRegistrationFlow(context.Background(), uuid.FromStringOrNil(f.Id))
		require.NoError(t, err)
		interim.InternalContext = contextFixture
		require.NoError(t, reg.RegistrationFlowPersister().UpdateRegistrationFlow(context.Background(), interim))

		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		cb(values)

		// We use the response replay
		body, res := testhelpers.RegistrationMakeRequest(t, false, isSPA, f, client, values.Encode())
		return body, res, f
	}

	t.Run("case=should fail to create identity if schema is missing the identifier", func(t *testing.T) {
		testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/noid.schema.json")
		t.Cleanup(func() {
			testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/registration.schema.json")
		})

		email := testhelpers.RandomEmail()

		var values = func(v url.Values) {
			v.Set("traits.email", email)
			v.Set(node.WebAuthnRegister, string(registrationFixtureSuccessResponse))
			v.Del("method")
		}

		for _, f := range flows {
			t.Run("type="+f, func(t *testing.T) {
				actual, _, _ := submitWebAuthnRegistrationWithClient(t, f, registrationFixtureSuccessInternalContext, testhelpers.NewClientWithCookies(t), values)

				assert.Contains(t, gjson.Get(actual, "ui.action").String(), publicTS.URL+registration.RouteSubmitFlow, "%s", actual)
				registrationhelpers.CheckFormContent(t, []byte(actual), node.WebAuthnRegisterTrigger, "csrf_token", "traits.email")
				assert.Equal(t, text.NewErrorValidationIdentifierMissing().Text, gjson.Get(actual, "ui.messages.0.text").String(), "%s", actual)
			})
		}
	})

	makeRegistration := func(t *testing.T, f string, values func(v url.Values)) (actual string, res *http.Response, fetchedFlow *registration.Flow) {
		actual, res, actualFlow := submitWebAuthnRegistrationWithClient(t, f, registrationFixtureSuccessInternalContext, testhelpers.NewClientWithCookies(t), values)
		fetchedFlow, err := reg.RegistrationFlowPersister().GetRegistrationFlow(context.Background(), uuid.FromStringOrNil(actualFlow.Id))
		require.NoError(t, err)

		return actual, res, fetchedFlow
	}

	makeSuccessfulRegistration := func(t *testing.T, f string, expectReturnTo string, values func(v url.Values)) (actual string) {
		actual, res, fetchedFlow := makeRegistration(t, f, values)
		assert.Empty(t, gjson.GetBytes(fetchedFlow.InternalContext, flow.PrefixInternalContextKey(identity.CredentialsTypeWebAuthn, webauthn.InternalContextKeySessionData)), "has cleaned up the internal context after success")
		if f == "spa" {
			expectReturnTo = publicTS.URL
		}
		assert.Contains(t, res.Request.URL.String(), expectReturnTo, "%+v\n\t%s", res.Request, assertx.PrettifyJSONPayload(t, actual))
		return actual
	}

	getPrefix := func(f string) (prefix string) {
		if f == "spa" {
			prefix = "session."
		}
		return
	}

	t.Run("successful registration", func(t *testing.T) {
		t.Cleanup(func() {
			conf.MustSet(config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypeWebAuthn.String()), nil)
		})

		var values = func(email string) func(v url.Values) {
			return func(v url.Values) {
				v.Set("traits.username", email)
				v.Set("traits.foobar", "bazbar")
				v.Set(node.WebAuthnRegister, string(registrationFixtureSuccessResponse))
				v.Del("method")
			}
		}

		t.Run("case=should create the identity but not a session", func(t *testing.T) {
			useReturnToFromTS(redirNoSessionTS)
			t.Cleanup(func() {
				useReturnToFromTS(redirTS)
			})
			conf.MustSet(config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypePassword.String()), nil)

			for _, f := range flows {
				t.Run("type="+f, func(t *testing.T) {
					email := testhelpers.RandomEmail()
					actual := makeSuccessfulRegistration(t, f, redirNoSessionTS.URL+"/registration-return-ts", values(email))

					if f == "spa" {
						assert.Equal(t, email, gjson.Get(actual, "identity.traits.username").String(), "%s", actual)
						assert.False(t, gjson.Get(actual, "session").Exists(), "because the registration yielded no session, the user is not expected to be signed in: %s", actual)
					} else {
						assert.Equal(t, "null\n", actual, "because the registration yielded no session, the user is not expected to be signed in: %s", actual)
					}

					i, _, err := reg.PrivilegedIdentityPool().FindByCredentialsIdentifier(context.Background(), identity.CredentialsTypeWebAuthn, email)
					require.NoError(t, err)
					assert.Equal(t, email, gjson.GetBytes(i.Traits, "username").String(), "%s", actual)
				})
			}
		})

		t.Run("case=should create the identity and a session and use the correct schema", func(t *testing.T) {
			conf.MustSet(config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypeWebAuthn.String()), []config.SelfServiceHook{{Name: "session"}})
			conf.MustSet(config.ViperKeyDefaultIdentitySchemaID, "advanced-user")
			conf.MustSet(config.ViperKeyIdentitySchemas, config.Schemas{
				{ID: "does-not-exist", URL: "file://./stub/profile.schema.json"},
				{ID: "advanced-user", URL: "file://./stub/registration.schema.json"},
			})

			for _, f := range flows {
				t.Run("type="+f, func(t *testing.T) {
					email := testhelpers.RandomEmail()
					actual := makeSuccessfulRegistration(t, f, redirTS.URL+"/registration-return-ts", values(email))

					prefix := getPrefix(f)

					assert.Equal(t, email, gjson.Get(actual, prefix+"identity.traits.username").String(), "%s", actual)
					assert.True(t, gjson.Get(actual, prefix+"active").Bool(), "%s", actual)

					i, _, err := reg.PrivilegedIdentityPool().FindByCredentialsIdentifier(context.Background(), identity.CredentialsTypeWebAuthn, email)
					require.NoError(t, err)
					assert.Equal(t, email, gjson.GetBytes(i.Traits, "username").String(), "%s", actual)
				})
			}
		})

		t.Run("case=not able to create the same account twice", func(t *testing.T) {
			conf.MustSet(config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypeWebAuthn.String()), []config.SelfServiceHook{{Name: "session"}})
			testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/registration.schema.json")

			for _, f := range flows {
				t.Run("type="+f, func(t *testing.T) {
					email := testhelpers.RandomEmail()
					actual := makeSuccessfulRegistration(t, f, redirTS.URL+"/registration-return-ts", values(email))
					assert.True(t, gjson.Get(actual, getPrefix(f)+"active").Bool(), "%s", actual)

					actual, _, _ = makeRegistration(t, f, values(email))
					assert.Contains(t, gjson.Get(actual, "ui.action").String(), publicTS.URL+registration.RouteSubmitFlow, "%s", actual)
					registrationhelpers.CheckFormContent(t, []byte(actual), node.WebAuthnRegisterTrigger, "csrf_token", "traits.username")
					assert.Equal(t, text.NewErrorValidationDuplicateCredentials().Text, gjson.Get(actual, "ui.messages.0.text").String(), "%s", actual)
				})
			}
		})

		t.Run("case=reset previous form errors", func(t *testing.T) {
			conf.MustSet(config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypeWebAuthn.String()), []config.SelfServiceHook{{Name: "session"}})
			testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/registration.schema.json")

			for _, f := range flows {
				t.Run("type="+f, func(t *testing.T) {
					email := testhelpers.RandomEmail()
					actual, _, _ := makeRegistration(t, f, func(v url.Values) {
						v.Del("traits.username")
						v.Set("traits.foobar", "bazbar")
						v.Set(node.WebAuthnRegister, string(registrationFixtureSuccessResponse))
						v.Del("method")
					})
					registrationhelpers.CheckFormContent(t, []byte(actual), node.WebAuthnRegisterTrigger, "csrf_token", "traits.username", "traits.foobar")
					assert.Contains(t, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.username).messages.0").String(), `Property username is missing`, "%s", actual)

					actual, _, _ = makeRegistration(t, f, func(v url.Values) {
						v.Set("traits.username", email)
						v.Del("traits.foobar")
						v.Set(node.WebAuthnRegister, string(registrationFixtureSuccessResponse))
						v.Del("method")
					})
					registrationhelpers.CheckFormContent(t, []byte(actual), node.WebAuthnRegisterTrigger, "csrf_token", "traits.username", "traits.foobar")
					assert.Contains(t, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.foobar).messages.0").String(), `Property foobar is missing`, "%s", actual)
					assert.Empty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.username).messages").Array())
					assert.Empty(t, gjson.Get(actual, "ui.nodes.messages").Array())
				})
			}
		})
	})

	t.Run("case=should fail if no identifier was set in the schema", func(t *testing.T) {
		testhelpers.SetDefaultIdentitySchema(conf, "file://stub/missing-identifier.schema.json")

		for _, f := range []string{"spa", "api", "browser"} {
			t.Run("type="+f, func(t *testing.T) {
				actual, _, _ := makeRegistration(t, f, func(v url.Values) {
					v.Set("traits.email", testhelpers.RandomEmail())
					v.Set(node.WebAuthnRegister, string(registrationFixtureSuccessResponse))
					v.Del("method")
				})
				assert.Equal(t, text.NewErrorValidationIdentifierMissing().Text, gjson.Get(actual, "ui.messages.0.text").String(), "%s", actual)
			})
		}
	})
}
