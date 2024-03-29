// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package passkey_test

import (
	_ "embed"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal/registrationhelpers"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/x/randx"
	"github.com/ory/x/sqlxx"
)

var (
	flows = []string{"spa", "browser"}

	//go:embed fixtures/registration/success/response.json
	registrationFixtureSuccessResponse []byte
	//go:embed fixtures/registration/success/internal_context.json
	registrationFixtureSuccessInternalContext []byte
	//go:embed fixtures/registration/failure/internal_context_missing_user_id.json
	registrationFixtureFailureInternalContextMissingUserID []byte
	//go:embed fixtures/registration/failure/internal_context_wrong_user_id.json
	registrationFixtureFailureInternalContextWrongUserID []byte
)

func flowIsSPA(flow string) bool {
	return flow == "spa"
}

func TestRegistration(t *testing.T) {
	t.Parallel()

	t.Run("AssertCommonErrorCases", func(t *testing.T) {
		registrationhelpers.AssertCommonErrorCases(t, flows)
	})

	t.Run("AssertRegistrationRespectsValidation", func(t *testing.T) {
		t.Parallel()
		reg := newRegistrationRegistry(t)
		registrationhelpers.AssertRegistrationRespectsValidation(t, reg, flows, func(v url.Values) {
			v.Del("traits.foobar")
			v.Set(node.PasskeyRegister, "{}")
			v.Del("method")
		})
	})

	t.Run("AssertCSRFFailures", func(t *testing.T) {
		t.Parallel()
		reg := newRegistrationRegistry(t)
		registrationhelpers.AssertCSRFFailures(t, reg, flows, func(v url.Values) {
			v.Set(node.PasskeyRegister, "{}")
			v.Del("method")
		})
	})

	t.Run("AssertSchemaDoesNotExist", func(t *testing.T) {
		t.Parallel()
		reg := newRegistrationRegistry(t)
		registrationhelpers.AssertSchemDoesNotExist(t, reg, flows, func(v url.Values) {
			v.Set(node.PasskeyRegister, "{}")
			v.Del("method")
		})
	})

	t.Run("case=passkey button does not exist when passwordless is disabled", func(t *testing.T) {
		t.Parallel()
		fix := newRegistrationFixture(t)
		fix.conf.MustSet(fix.ctx, config.ViperKeyPasskeyEnabled, false)
		t.Cleanup(func() { fix.conf.MustSet(fix.ctx, config.ViperKeyPasskeyEnabled, true) })
		for _, flowType := range flows {
			flowType := flowType
			t.Run(flowType, func(t *testing.T) {
				t.Parallel()
				client := testhelpers.NewClientWithCookies(t)
				flo := testhelpers.InitializeRegistrationFlowViaBrowser(t, client, fix.publicTS, flowIsSPA(flowType), false, false)
				testhelpers.SnapshotTExcept(t, flo.Ui.Nodes, []string{
					"0.attributes.value",
				})
			})
		}
	})

	t.Run("case=passkey button exists", func(t *testing.T) {
		t.Parallel()
		fix := newRegistrationFixture(t)
		for _, flowType := range flows {
			flowType := flowType
			t.Run(flowType, func(t *testing.T) {
				t.Parallel()
				client := testhelpers.NewClientWithCookies(t)
				f := testhelpers.InitializeRegistrationFlowViaBrowser(t, client, fix.publicTS, flowIsSPA(flowType), false, false)
				testhelpers.SnapshotTExcept(t, f.Ui.Nodes, []string{
					"2.attributes.value",
					"3.attributes.src",
					"3.attributes.nonce",
					"6.attributes.value",
				})
			})
		}
	})

	t.Run("case=should return an error because not passing validation", func(t *testing.T) {
		t.Parallel()
		fix := newRegistrationFixture(t)
		email := testhelpers.RandomEmail()

		var values = func(v url.Values) {
			v.Set("traits.username", email)
			v.Del("traits.foobar")
			v.Set(node.PasskeyRegister, "{}")
			v.Del("method")
		}

		for _, f := range flows {
			t.Run("type="+f, func(t *testing.T) {
				actual := registrationhelpers.ExpectValidationError(t, fix.publicTS, fix.conf, f, values)

				assert.NotEmpty(t, gjson.Get(actual, "id").String(), "%s", actual)
				assert.Contains(t, gjson.Get(actual, "ui.action").String(), fix.publicTS.URL+registration.RouteSubmitFlow, "%s", actual)
				registrationhelpers.CheckFormContent(t, []byte(actual), "csrf_token", "traits.username", "traits.foobar")
				assert.Contains(t, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.foobar).messages.0").String(), `Property foobar is missing`, "%s", actual)
				assert.Equal(t, email, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.username).attributes.value").String(), "%s", actual)
			})
		}
	})

	t.Run("case=should reject invalid transient payload", func(t *testing.T) {
		t.Parallel()
		fix := newRegistrationFixture(t)
		email := testhelpers.RandomEmail()

		var values = func(v url.Values) {
			v.Set("traits.username", email)
			v.Set("traits.foobar", "bar")
			v.Set("transient_payload", "42")
			v.Set(node.PasskeyRegister, "{}")
			v.Del("method")
		}

		for _, f := range flows {
			t.Run("type="+f, func(t *testing.T) {
				actual := registrationhelpers.ExpectValidationError(t, fix.publicTS, fix.conf, f, values)

				assert.NotEmpty(t, gjson.Get(actual, "id").String(), "%s", actual)
				assert.Contains(t, gjson.Get(actual, "ui.action").String(), fix.publicTS.URL+registration.RouteSubmitFlow, "%s", actual)
				registrationhelpers.CheckFormContent(t, []byte(actual), "csrf_token", "traits.username", "traits.foobar")
				assert.Equal(t, "bar", gjson.Get(actual, "ui.nodes.#(attributes.name==traits.foobar).attributes.value").String(), "%s", actual)
				assert.Equal(t, email, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.username).attributes.value").String(), "%s", actual)
				assert.Equal(t, int64(4000026), gjson.Get(actual, "ui.nodes.#(attributes.name==transient_payload).messages.0.id").Int(), "%s", actual)
			})
		}
	})

	t.Run("case=should return an error because passkey response is invalid", func(t *testing.T) {
		t.Parallel()
		fix := newRegistrationFixture(t)
		email := testhelpers.RandomEmail()

		var values = func(v url.Values) {
			v.Set("traits.username", email)
			v.Set("traits.foobar", "bazbar")
			v.Set(node.PasskeyRegister, "invalid")
			v.Set("method", "passkey")
		}

		for _, f := range flows {
			t.Run("type="+f, func(t *testing.T) {
				actual, _, _ := fix.submitPasskeyRegistration(t, f, testhelpers.NewClientWithCookies(t), values)
				assert.NotEmpty(t, gjson.Get(actual, "id").String(), "%s", actual)
				assert.Contains(t, gjson.Get(actual, "ui.action").String(), fix.publicTS.URL+registration.RouteSubmitFlow, "%s", actual)
				registrationhelpers.CheckFormContent(t, []byte(actual), node.PasskeyRegister, "csrf_token", "traits.username", "traits.foobar")
				assert.Equal(t, "bazbar", gjson.Get(actual, "ui.nodes.#(attributes.name==traits.foobar).attributes.value").String(), "%s", actual)
				assert.Equal(t, email, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.username).attributes.value").String(), "%s", actual)
				assert.Contains(t, gjson.Get(actual, "ui.messages.0").String(), `Unable to parse WebAuthn response: Parse error for Registration`, "%s", actual)
			})
		}
	})

	t.Run("case=should return an error because internal context is invalid", func(t *testing.T) {
		t.Parallel()
		fix := newRegistrationFixture(t)
		email := testhelpers.RandomEmail()

		for _, tc := range []struct {
			name            string
			internalContext string
		}{{
			name:            "invalid json",
			internalContext: "invalid",
		}, {
			name:            "missing user ID",
			internalContext: string(registrationFixtureFailureInternalContextMissingUserID),
		}, {
			name:            "wrong user ID",
			internalContext: string(registrationFixtureFailureInternalContextWrongUserID),
		}} {
			tc := tc
			t.Run("context="+tc.name, func(t *testing.T) {
				var values = func(v url.Values) {
					v.Set("traits.username", email)
					v.Set("traits.foobar", "bazbar")
					v.Set(node.PasskeyRegister, string(registrationFixtureSuccessResponse))
					v.Del("method")
				}

				for _, f := range flows {
					t.Run("type="+f, func(t *testing.T) {
						actual, _, _ := fix.submitPasskeyRegistration(t, f, testhelpers.NewClientWithCookies(t), values,
							withInternalContext(sqlxx.JSONRawMessage(tc.internalContext)))
						if flowIsSPA(f) {
							assert.Equal(t, "Internal Server Error", gjson.Get(actual, "error.status").String(), "%s", actual)
						} else {
							assert.Equal(t, "Internal Server Error", gjson.Get(actual, "status").String(), "%s", actual)
						}
					})
				}
			})
		}
	})

	t.Run("case=should fail to create identity if schema is missing the identifier", func(t *testing.T) {
		t.Parallel()
		fix := newRegistrationFixture(t)
		testhelpers.SetDefaultIdentitySchema(fix.conf, "file://./stub/noid.schema.json")
		email := testhelpers.RandomEmail()

		for _, f := range flows {
			t.Run("type="+f, func(t *testing.T) {
				client := testhelpers.NewClientWithCookies(t)
				isSPA := f == "spa"
				regFlow := testhelpers.InitializeRegistrationFlowViaBrowser(t, client, fix.publicTS, isSPA, false, false)

				// fill out traits and click on "sign up with passkey"
				urlValues := testhelpers.SDKFormFieldsToURLValues(regFlow.Ui.Nodes)
				urlValues.Set("traits.email", email)
				urlValues.Set("method", "passkey")
				actual, _ := testhelpers.RegistrationMakeRequest(t, false, isSPA, regFlow, client, urlValues.Encode())

				assert.Contains(t, gjson.Get(actual, "ui.action").String(), fix.publicTS.URL+registration.RouteSubmitFlow, "%s", actual)
				registrationhelpers.CheckFormContent(t, []byte(actual), "csrf_token", "traits.email")
				assert.Equal(t, text.NewErrorValidationRegistrationNoStrategyFound().Text, gjson.Get(actual, "ui.messages.0.text").String(), "%s", actual)
			})
		}
	})

	getPrefix := func(f string) (prefix string) {
		if f == "spa" {
			prefix = "session."
		}
		return
	}

	t.Run("successful registration", func(t *testing.T) {
		t.Parallel()
		fix := newRegistrationFixture(t)
		t.Cleanup(fix.disableSessionAfterRegistration)

		var values = func(email string) func(v url.Values) {
			return func(v url.Values) {
				v.Set("traits.username", email)
				v.Set("traits.foobar", "bazbar")
				v.Set(node.PasskeyRegister, string(registrationFixtureSuccessResponse))
				v.Del("method")
			}
		}

		t.Run("case=should create the identity but not a session", func(t *testing.T) {
			fix.useRedirNoSessionTS()
			t.Cleanup(fix.useRedirTS)
			fix.disableSessionAfterRegistration()

			for _, f := range flows {
				t.Run("type="+f, func(t *testing.T) {
					email := f + "-" + testhelpers.RandomEmail()
					userID := f + "-user-" + randx.MustString(8, randx.AlphaNum)
					actual := fix.makeSuccessfulRegistration(t, f, fix.redirNoSessionTS.URL+"/registration-return-ts", values(email), withUserID(userID))

					if f == "spa" {
						assert.Equal(t, email, gjson.Get(actual, "identity.traits.username").String(), "%s", actual)
						assert.False(t, gjson.Get(actual, "session").Exists(), "because the registration yielded no session, the user is not expected to be signed in: %s", actual)
					} else {
						assert.Equal(t, "null\n", actual, "because the registration yielded no session, the user is not expected to be signed in: %s", actual)
					}

					i, _, err := fix.reg.PrivilegedIdentityPool().FindByCredentialsIdentifier(fix.ctx, identity.CredentialsTypePasskey, userID)
					require.NoError(t, err)
					assert.Equal(t, "aal1", i.AvailableAAL.String)
					assert.Equal(t, email, gjson.GetBytes(i.Traits, "username").String(), "%s", actual)
				})
			}
		})

		t.Run("case=should accept valid transient payload", func(t *testing.T) {
			fix.useRedirNoSessionTS()
			t.Cleanup(fix.useRedirTS)
			fix.disableSessionAfterRegistration()

			for _, f := range flows {
				t.Run("type="+f, func(t *testing.T) {
					email := testhelpers.RandomEmail()
					userID := f + "-user-" + randx.MustString(8, randx.AlphaNum)
					actual := fix.makeSuccessfulRegistration(t, f, fix.redirNoSessionTS.URL+"/registration-return-ts", func(v url.Values) {
						values(email)(v)
						v.Set("transient_payload.stuff", "42")
					}, withUserID(userID))

					if f == "spa" {
						assert.Equal(t, email, gjson.Get(actual, "identity.traits.username").String(), "%s", actual)
						assert.False(t, gjson.Get(actual, "session").Exists(), "because the registration yielded no session, the user is not expected to be signed in: %s", actual)
					} else {
						assert.Equal(t, "null\n", actual, "because the registration yielded no session, the user is not expected to be signed in: %s", actual)
					}

					i, _, err := fix.reg.PrivilegedIdentityPool().FindByCredentialsIdentifier(fix.ctx, identity.CredentialsTypePasskey, userID)
					require.NoError(t, err)
					assert.Equal(t, email, gjson.GetBytes(i.Traits, "username").String(), "%s", actual)
				})
			}
		})

		t.Run("case=should create the identity and a session and use the correct schema", func(t *testing.T) {
			fix.enableSessionAfterRegistration()
			fix.conf.MustSet(fix.ctx, config.ViperKeyDefaultIdentitySchemaID, "advanced-user")
			fix.conf.MustSet(fix.ctx, config.ViperKeyIdentitySchemas, config.Schemas{
				{ID: "does-not-exist", URL: "file://./stub/profile.schema.json"},
				{ID: "advanced-user", URL: "file://./stub/registration.schema.json"},
			})

			for _, f := range flows {
				t.Run("type="+f, func(t *testing.T) {
					email := testhelpers.RandomEmail()
					userID := f + "-user-" + randx.MustString(8, randx.AlphaNum)
					actual := fix.makeSuccessfulRegistration(t, f, fix.redirTS.URL+"/registration-return-ts", values(email), withUserID(userID))

					prefix := getPrefix(f)

					assert.Equal(t, email, gjson.Get(actual, prefix+"identity.traits.username").String(), "%s", actual)
					assert.True(t, gjson.Get(actual, prefix+"active").Bool(), "%s", actual)

					i, _, err := fix.reg.PrivilegedIdentityPool().FindByCredentialsIdentifier(fix.ctx, identity.CredentialsTypePasskey, userID)
					require.NoError(t, err)
					assert.Equal(t, email, gjson.GetBytes(i.Traits, "username").String(), "%s", actual)
				})
			}
		})

		t.Run("case=not able to create the same account twice", func(t *testing.T) {
			fix.enableSessionAfterRegistration()
			testhelpers.SetDefaultIdentitySchema(fix.conf, "file://./stub/registration.schema.json")

			for _, f := range flows {
				t.Run("type="+f, func(t *testing.T) {
					email := testhelpers.RandomEmail()
					userID := f + "-user-" + randx.MustString(8, randx.AlphaNum)
					actual := fix.makeSuccessfulRegistration(t, f, fix.redirTS.URL+"/registration-return-ts", values(email), withUserID(userID))
					assert.True(t, gjson.Get(actual, getPrefix(f)+"active").Bool(), "%s", actual)

					actual, _, _ = fix.makeRegistration(t, f, values(email))
					assert.Contains(t, gjson.Get(actual, "ui.action").String(), fix.publicTS.URL+registration.RouteSubmitFlow, "%s", actual)
					registrationhelpers.CheckFormContent(t, []byte(actual), "csrf_token", "traits.username")
					assert.Equal(t,
						"You tried signing in with "+email+" which is already in use by another account. You can sign in using your password.",
						gjson.Get(actual, "ui.messages.0.text").String(), "%s", actual)
				})
			}
		})

		t.Run("case=reset previous form errors", func(t *testing.T) {
			fix.enableSessionAfterRegistration()
			testhelpers.SetDefaultIdentitySchema(fix.conf, "file://./stub/registration.schema.json")

			for _, f := range flows {
				t.Run("type="+f, func(t *testing.T) {
					email := testhelpers.RandomEmail()
					actual, _, _ := fix.makeRegistration(t, f, func(v url.Values) {
						v.Del("traits.username")
						v.Set("traits.foobar", "bazbar")
						v.Set(node.PasskeyRegister, string(registrationFixtureSuccessResponse))
					})
					registrationhelpers.CheckFormContent(t, []byte(actual), "csrf_token", "traits.username", "traits.foobar")
					assert.Contains(t, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.username).messages.0").String(), `Property username is missing`, "%s", actual)

					actual, _, _ = fix.makeRegistration(t, f, func(v url.Values) {
						v.Set("traits.username", email)
						v.Del("traits.foobar")
						v.Set(node.PasskeyRegister, string(registrationFixtureSuccessResponse))
						v.Del("method")
					})
					registrationhelpers.CheckFormContent(t, []byte(actual), "csrf_token", "traits.username", "traits.foobar")
					assert.Contains(t, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.foobar).messages.0").String(), `Property foobar is missing`, "%s", actual)
					assert.Empty(t, gjson.Get(actual, "ui.nodes.#(attributes.name==traits.username).messages").Array())
					assert.Empty(t, gjson.Get(actual, "ui.nodes.messages").Array())
				})
			}
		})
	})
}
