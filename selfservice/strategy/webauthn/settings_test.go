package webauthn_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/strategy/webauthn"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	kratos "github.com/ory/kratos-client-go"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"

	"github.com/stretchr/testify/require"

	"github.com/ory/x/assertx"
	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/x"
)

//go:embed fixtures/settings/has_webauth.json
var settingsFixtureHasWebAuthn []byte

//go:embed fixtures/settings/no_webauth.json
var settingsFixtureNoWebauthn []byte

//go:embed fixtures/settings/success/identity.json
var settingsFixtureSuccessIdentity []byte

//go:embed fixtures/settings/success/response.json
var settingsFixtureSuccessResponse []byte

//go:embed fixtures/settings/success/internal_context.json
var settingsFixtureSuccessInternalContext []byte

const registerDisplayNameGJSONQuery = "ui.nodes.#(attributes.name==" + node.WebAuthnRegisterDisplayName + ")"

func createIdentityWithoutWebAuthn(t *testing.T, reg driver.Registry) *identity.Identity {
	id := createIdentity(t, reg)
	delete(id.Credentials, identity.CredentialsTypeWebAuthn)
	require.NoError(t, reg.PrivilegedIdentityPool().UpdateIdentity(context.Background(), id))
	return id
}

func createIdentity(t *testing.T, reg driver.Registry) *identity.Identity {
	identifier := x.NewUUID().String() + "@ory.sh"
	password := x.NewUUID().String()
	p, err := reg.Hasher().Generate(context.Background(), []byte(password))
	require.NoError(t, err)
	i := &identity.Identity{
		Traits: identity.Traits(fmt.Sprintf(`{"subject":"%s"}`, identifier)),
		VerifiableAddresses: []identity.VerifiableAddress{
			{
				Value:     identifier,
				Verified:  false,
				CreatedAt: time.Now(),
			},
		},
	}
	require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))
	i.Credentials = map[identity.CredentialsType]identity.Credentials{
		identity.CredentialsTypePassword: {
			Type:        identity.CredentialsTypePassword,
			Identifiers: []string{identifier},
			Config:      sqlxx.JSONRawMessage(`{"hashed_password":"` + string(p) + `"}`),
		},
		identity.CredentialsTypeWebAuthn: {
			Type:        identity.CredentialsTypeWebAuthn,
			Identifiers: []string{i.ID.String()},
			Config:      sqlxx.JSONRawMessage(`{"credentials":[{"id":"Zm9vZm9v","display_name":"foo"},{"id":"YmFyYmFy","display_name":"bar"}]}`),
		},
	}
	require.NoError(t, reg.PrivilegedIdentityPool().UpdateIdentity(context.Background(), i))
	return i
}

func enableWebAuthn(conf *config.Config) {
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeWebAuthn)+".enabled", true)
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeWebAuthn)+".config.rp.display_name", "Ory Corp")
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeWebAuthn)+".config.rp.id", "localhost")
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeWebAuthn)+".config.rp.origin", "http://localhost:4455")
}

func ensureReplacement(t *testing.T, index string, ui kratos.UiContainer, expected string) {
	actual, err := json.Marshal(ui.Nodes)
	require.NoError(t, err)
	assert.Contains(t, gjson.GetBytes(actual, index+".attributes.onclick").String(), expected, "ensure that the replacement works")
}

func TestCompleteSettings(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword)+".enabled", false)
	enableWebAuthn(conf)
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+".profile.enabled", false)
	conf.MustSet(config.ViperKeySelfServiceSettingsRequiredAAL, "aal1")

	router := x.NewRouterPublic()
	publicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin())

	errTS := testhelpers.NewErrorTestServer(t, reg)
	uiTS := testhelpers.NewSettingsUIFlowEchoServer(t, reg)
	_ = testhelpers.NewRedirSessionEchoTS(t, reg)
	loginTS := testhelpers.NewLoginUIFlowEchoServer(t, reg)

	conf.MustSet(config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1m")

	conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/login.schema.json")
	conf.MustSet(config.ViperKeySecretsDefault, []string{"not-a-secure-session-key"})

	t.Run("case=a device is shown which can be unlinked", func(t *testing.T) {
		id := createIdentity(t, reg)

		apiClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id)
		f := testhelpers.InitializeSettingsFlowViaBrowser(t, apiClient, true, publicTS)

		testhelpers.SnapshotTExcept(t, f.Ui.Nodes, []string{
			"0.attributes.value",
			"4.attributes.onclick",
			"6.attributes.src",
			"6.attributes.nonce",
		})
		ensureReplacement(t, "4", f.Ui, "Ory Corp")
	})

	t.Run("case=one activation element is shown", func(t *testing.T) {
		id := createIdentityWithoutWebAuthn(t, reg)
		require.NoError(t, reg.PrivilegedIdentityPool().UpdateIdentity(context.Background(), id))

		apiClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id)
		f := testhelpers.InitializeSettingsFlowViaBrowser(t, apiClient, true, publicTS)

		testhelpers.SnapshotTExcept(t, f.Ui.Nodes, []string{
			"0.attributes.value",
			"2.attributes.onload",
			"2.attributes.onclick",
			"4.attributes.src",
			"4.attributes.nonce",
		})
		ensureReplacement(t, "2", f.Ui, "Ory Corp")
	})

	t.Run("case=webauthn only works for browsers", func(t *testing.T) {
		id := createIdentityWithoutWebAuthn(t, reg)
		require.NoError(t, reg.PrivilegedIdentityPool().UpdateIdentity(context.Background(), id))

		apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)
		f := testhelpers.InitializeSettingsFlowViaAPI(t, apiClient, publicTS)
		assert.Empty(t, f.Ui.Nodes)
	})

	doAPIFlow := func(t *testing.T, v func(url.Values), id *identity.Identity) (string, *http.Response) {
		apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)
		f := testhelpers.InitializeSettingsFlowViaAPI(t, apiClient, publicTS)
		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		v(values)
		payload := testhelpers.EncodeFormAsJSON(t, true, values)
		return testhelpers.SettingsMakeRequest(t, true, false, f, apiClient, payload)
	}

	doBrowserFlow := func(t *testing.T, spa bool, v func(url.Values), id *identity.Identity) (string, *http.Response) {
		browserClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id)
		f := testhelpers.InitializeSettingsFlowViaBrowser(t, browserClient, spa, publicTS)
		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		v(values)
		return testhelpers.SettingsMakeRequest(t, false, spa, f, browserClient, testhelpers.EncodeFormAsJSON(t, spa, values))
	}

	t.Run("case=fails with api submit because only browsers are supported", func(t *testing.T) {
		id := createIdentityWithoutWebAuthn(t, reg)
		body, res := doAPIFlow(t, func(v url.Values) {
			v.Set(node.WebAuthnRegister, "{}")
		}, id)

		assert.Contains(t, res.Request.URL.String(), publicTS.URL+settings.RouteSubmitFlow)
		assert.Equal(t, text.NewErrorValidationSettingsNoStrategyFound().Text, gjson.Get(body, "ui.messages.0.text").String(), "%s", body)
	})

	t.Run("case=fails with browser submit if name is missing", func(t *testing.T) {
		run := func(t *testing.T, spa bool) {
			id := createIdentityWithoutWebAuthn(t, reg)
			body, res := doBrowserFlow(t, spa, func(v url.Values) {
				v.Set(node.WebAuthnRegister, "{}")
			}, id)
			if spa {
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+settings.RouteSubmitFlow)
			} else {
				assert.Contains(t, res.Request.URL.String(), uiTS.URL)
			}
			assert.Contains(t, gjson.Get(body, registerDisplayNameGJSONQuery+".messages.0.text").String(), "Property webauthn_register_displayname is missing.", "%s", body)
		}

		t.Run("type=browser", func(t *testing.T) {
			run(t, false)
		})

		t.Run("type=spa", func(t *testing.T) {
			run(t, true)
		})
	})

	t.Run("case=fails with browser submit because csrf token is missing", func(t *testing.T) {
		run := func(t *testing.T, spa bool) {
			id := createIdentityWithoutWebAuthn(t, reg)
			body, res := doBrowserFlow(t, spa, func(v url.Values) {
				v.Del("csrf_token")
				v.Set(node.WebAuthnRegister, "{}")
				v.Set(node.WebAuthnRegisterDisplayName, "foobar")
			}, id)
			if spa {
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+settings.RouteSubmitFlow)
				assert.Equal(t, x.ErrInvalidCSRFToken.Reason(), gjson.Get(body, "error.reason").String(), body)
			} else {
				assert.Contains(t, res.Request.URL.String(), errTS.URL)
				assert.Equal(t, x.ErrInvalidCSRFToken.Reason(), gjson.Get(body, "reason").String(), body)
			}
		}

		t.Run("type=browser", func(t *testing.T) {
			run(t, false)
		})

		t.Run("type=spa", func(t *testing.T) {
			run(t, true)
		})
	})

	t.Run("case=fails with browser submit register payload is invalid", func(t *testing.T) {
		run := func(t *testing.T, spa bool) {
			id := createIdentityWithoutWebAuthn(t, reg)
			body, res := doBrowserFlow(t, spa, func(v url.Values) {
				v.Set(node.WebAuthnRegister, "{}")
				v.Set(node.WebAuthnRegisterDisplayName, "foobar")
			}, id)
			if spa {
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+settings.RouteSubmitFlow)
			} else {
				assert.Contains(t, res.Request.URL.String(), uiTS.URL)
			}
			assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "Parse error for Registration", "%s", body)
		}

		t.Run("type=browser", func(t *testing.T) {
			run(t, false)
		})

		t.Run("type=spa", func(t *testing.T) {
			run(t, true)
		})
	})

	t.Run("case=requires privileged session for register", func(t *testing.T) {
		conf.MustSet(config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1ns")
		t.Cleanup(func() {
			conf.MustSet(config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "5m")
		})

		run := func(t *testing.T, spa bool) {
			id := createIdentityWithoutWebAuthn(t, reg)
			body, res := doBrowserFlow(t, spa, func(v url.Values) {
				v.Set(node.WebAuthnRegister, "{}")
				v.Set(node.WebAuthnRegisterDisplayName, "foobar")
			}, id)

			if spa {
				assert.NotEmpty(t, gjson.Get(body, "redirect_browser_to").String())
				assert.Equal(t, http.StatusForbidden, res.StatusCode)
				assertx.EqualAsJSONExcept(t, settings.NewFlowNeedsReAuth(), json.RawMessage(body), []string{"redirect_browser_to"})
			} else {
				assert.Contains(t, res.Request.URL.String(), loginTS.URL+"/login-ts")
				assertx.EqualAsJSON(t, text.NewInfoLoginReAuth(), json.RawMessage(gjson.Get(body, "ui.messages.0").Raw))
			}
		}

		t.Run("type=browser", func(t *testing.T) {
			run(t, false)
		})

		t.Run("type=spa", func(t *testing.T) {
			run(t, true)
		})
	})

	t.Run("case=add a security key", func(t *testing.T) {
		run := func(t *testing.T, spa bool) {
			// We load our identity which we will use to replay the webauth session
			var id identity.Identity
			require.NoError(t, json.Unmarshal(settingsFixtureSuccessIdentity, &id))
			_ = reg.PrivilegedIdentityPool().DeleteIdentity(context.Background(), id.ID)
			browserClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, &id)
			f := testhelpers.InitializeSettingsFlowViaBrowser(t, browserClient, spa, publicTS)

			// We inject the session to replay
			interim, err := reg.SettingsFlowPersister().GetSettingsFlow(context.Background(), uuid.FromStringOrNil(f.Id))
			require.NoError(t, err)
			interim.InternalContext = settingsFixtureSuccessInternalContext
			require.NoError(t, reg.SettingsFlowPersister().UpdateSettingsFlow(context.Background(), interim))

			values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)

			// We use the response replay
			values.Set(node.WebAuthnRegister, string(settingsFixtureSuccessResponse))
			values.Set(node.WebAuthnRegisterDisplayName, "foobar")
			body, res := testhelpers.SettingsMakeRequest(t, false, spa, f, browserClient, testhelpers.EncodeFormAsJSON(t, spa, values))

			if spa {
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+settings.RouteSubmitFlow)
			} else {
				assert.Contains(t, res.Request.URL.String(), uiTS.URL)
			}
			assert.EqualValues(t, settings.StateSuccess, gjson.Get(body, "state").String(), body)

			actual, err := reg.Persister().GetIdentityConfidential(context.Background(), id.ID)
			require.NoError(t, err)
			cred, ok := actual.GetCredentials(identity.CredentialsTypeWebAuthn)
			assert.True(t, ok)
			assert.Len(t, gjson.GetBytes(cred.Config, "credentials").Array(), 1)

			actualFlow, err := reg.SettingsFlowPersister().GetSettingsFlow(context.Background(), uuid.FromStringOrNil(f.Id))
			require.NoError(t, err)
			assert.Empty(t, gjson.GetBytes(actualFlow.InternalContext, flow.PrefixInternalContextKey(identity.CredentialsTypeWebAuthn, webauthn.InternalContextKeySessionData)))

			testhelpers.EnsureAAL(t, browserClient, publicTS, "aal2", string(identity.CredentialsTypeWebAuthn))
		}

		t.Run("type=browser", func(t *testing.T) {
			run(t, false)
		})

		t.Run("type=spa", func(t *testing.T) {
			run(t, true)
		})
	})

	t.Run("case=remove all security keys", func(t *testing.T) {
		run := func(t *testing.T, spa bool) {
			id := createIdentity(t, reg)
			allCred, ok := id.GetCredentials(identity.CredentialsTypeWebAuthn)
			assert.True(t, ok)

			var cc webauthn.CredentialsConfig
			require.NoError(t, json.Unmarshal(allCred.Config, &cc))
			require.Len(t, cc.Credentials, 2)

			for _, cred := range cc.Credentials {
				body, res := doBrowserFlow(t, spa, func(v url.Values) {
					v.Set(node.WebAuthnRemove, fmt.Sprintf("%x", cred.ID))
				}, id)

				if spa {
					assert.Contains(t, res.Request.URL.String(), publicTS.URL+settings.RouteSubmitFlow)
				} else {
					assert.Contains(t, res.Request.URL.String(), uiTS.URL)
				}
				assert.EqualValues(t, settings.StateSuccess, gjson.Get(body, "state").String(), body)
			}

			actual, err := reg.Persister().GetIdentityConfidential(context.Background(), id.ID)
			require.NoError(t, err)
			_, ok = actual.GetCredentials(identity.CredentialsTypeWebAuthn)
			assert.False(t, ok)
		}

		t.Run("type=browser", func(t *testing.T) {
			run(t, false)
		})

		t.Run("type=spa", func(t *testing.T) {
			run(t, true)
		})
	})

	t.Run("case=fails with browser submit register payload is invalid", func(t *testing.T) {
		run := func(t *testing.T, spa bool) {
			id := createIdentity(t, reg)
			body, res := doBrowserFlow(t, spa, func(v url.Values) {
				v.Set(node.WebAuthnRemove, fmt.Sprintf("%x", []byte("foofoo")))
			}, id)

			if spa {
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+settings.RouteSubmitFlow)
			} else {
				assert.Contains(t, res.Request.URL.String(), uiTS.URL)
			}
			assert.EqualValues(t, settings.StateSuccess, json.RawMessage(gjson.Get(body, "state").String()))

			actual, err := reg.Persister().GetIdentityConfidential(context.Background(), id.ID)
			require.NoError(t, err)
			cred, ok := actual.GetCredentials(identity.CredentialsTypeWebAuthn)
			assert.True(t, ok)
			assert.Len(t, gjson.GetBytes(cred.Config, "credentials").Array(), 1)
			assert.Equal(t, "bar", gjson.GetBytes(cred.Config, "credentials.0.display_name").String())
		}

		t.Run("type=browser", func(t *testing.T) {
			run(t, false)
		})

		t.Run("type=spa", func(t *testing.T) {
			run(t, true)
		})
	})
}
