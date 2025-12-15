// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package passkey_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/ory/kratos/x/nosurfx"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/strategy/passkey"

	"github.com/ory/x/snapshotx"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"

	"github.com/stretchr/testify/require"

	"github.com/ory/x/assertx"
	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/x"
)

//go:embed fixtures/settings/success/identity.json
var settingsFixtureSuccessIdentity []byte

//go:embed fixtures/settings/success/response.json
var settingsFixtureSuccessResponse []byte

//go:embed fixtures/settings/success/internal_context.json
var settingsFixtureSuccessInternalContext []byte

var ctx = context.Background()

func TestCompleteSettings(t *testing.T) {
	fix := newSettingsFixture(t)

	t.Run("case=invalid passkey config", func(t *testing.T) {
		fix := newSettingsFixture(t)
		fix.conf.MustSet(ctx, config.ViperKeyPasskeyRPID, "")
		id := fix.createIdentity(t)
		apiClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, ctx, fix.reg, id)

		req, err := http.NewRequest("GET", fix.publicTS.URL+settings.RouteInitBrowserFlow, nil)
		require.NoError(t, err)
		req.Header.Set("Accept", "application/json")
		res, err := apiClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	})

	t.Run("case=a device is shown which can be unlinked", func(t *testing.T) {
		id := fix.createIdentity(t)

		apiClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, ctx, fix.reg, id)
		f := testhelpers.InitializeSettingsFlowViaBrowser(t, apiClient, true, fix.publicTS)

		testhelpers.SnapshotTExcept(t, f.Ui.Nodes, []string{
			"0.attributes.value", // CSRF
			"1.attributes.nonce", // script
			"1.attributes.src",   // script
			"6.attributes.value", // passkey_settings_register
		})
	})

	t.Run("case=one activation element is shown", func(t *testing.T) {
		id := fix.createIdentityWithoutPasskey(t)
		require.NoError(t, fix.reg.PrivilegedIdentityPool().UpdateIdentity(fix.ctx, id))

		apiClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, ctx, fix.reg, id)
		f := testhelpers.InitializeSettingsFlowViaBrowser(t, apiClient, true, fix.publicTS)

		testhelpers.SnapshotTExcept(t, f.Ui.Nodes, []string{
			"0.attributes.value", // CSRF
			"1.attributes.nonce", // script
			"1.attributes.src",   // script
			"4.attributes.value", // passkey_create_data
		})
	})

	t.Run("case=passkeys only work for browsers", func(t *testing.T) {
		id := fix.createIdentityWithoutPasskey(t)
		require.NoError(t, fix.reg.PrivilegedIdentityPool().UpdateIdentity(fix.ctx, id))

		apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, ctx, fix.reg, id)
		f := testhelpers.InitializeSettingsFlowViaAPI(t, apiClient, fix.publicTS)
		for _, n := range f.Ui.Nodes {
			assert.NotEqual(t, n.Group, "passkey", "unexpected group: %s", n.Group)
		}
	})

	doAPIFlow := func(t *testing.T, v func(url.Values), id *identity.Identity) (string, *http.Response) {
		apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, ctx, fix.reg, id)
		f := testhelpers.InitializeSettingsFlowViaAPI(t, apiClient, fix.publicTS)
		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		v(values)
		payload := testhelpers.EncodeFormAsJSON(t, true, values)
		return testhelpers.SettingsMakeRequest(t, true, false, f, apiClient, payload)
	}

	doBrowserFlow := func(t *testing.T, spa bool, v func(url.Values), id *identity.Identity) (string, *http.Response) {
		browserClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, ctx, fix.reg, id)
		f := testhelpers.InitializeSettingsFlowViaBrowser(t, browserClient, spa, fix.publicTS)
		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		v(values)
		return testhelpers.SettingsMakeRequest(t, false, spa, f, browserClient, testhelpers.EncodeFormAsJSON(t, spa, values))
	}

	t.Run("case=fails with api submit because only browsers are supported", func(t *testing.T) {
		id := fix.createIdentityWithoutPasskey(t)
		body, res := doAPIFlow(t, func(v url.Values) {
			v.Set(node.PasskeySettingsRegister, "{}")
			v.Set("method", "passkey")
		}, id)

		assert.Contains(t, res.Request.URL.String(), fix.publicTS.URL+settings.RouteSubmitFlow)
		assert.Equal(t, text.NewErrorValidationSettingsNoStrategyFound().Text, gjson.Get(body, "ui.messages.0.text").String(), "%s", body)
	})

	t.Run("case=fails with browser submit because csrf token is missing", func(t *testing.T) {
		run := func(t *testing.T, spa bool) {
			id := fix.createIdentityWithoutPasskey(t)
			body, res := doBrowserFlow(t, spa, func(v url.Values) {
				v.Del("csrf_token")
				v.Set(node.PasskeySettingsRegister, "{}")
				v.Set("method", "passkey")
			}, id)
			if spa {
				assert.Contains(t, res.Request.URL.String(), fix.publicTS.URL+settings.RouteSubmitFlow)
				assert.Equal(t, nosurfx.ErrInvalidCSRFToken.Reason(), gjson.Get(body, "error.reason").String(), body)
			} else {
				assert.Contains(t, res.Request.URL.String(), fix.errTS.URL)
				assert.Equal(t, nosurfx.ErrInvalidCSRFToken.Reason(), gjson.Get(body, "reason").String(), body)
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
			id := fix.createIdentityWithoutPasskey(t)
			body, res := doBrowserFlow(t, spa, func(v url.Values) {
				v.Set(node.PasskeySettingsRegister, "{}")
				v.Set("method", "passkey")
			}, id)
			if spa {
				assert.Contains(t, res.Request.URL.String(), fix.publicTS.URL+settings.RouteSubmitFlow)
			} else {
				assert.Contains(t, res.Request.URL.String(), fix.uiTS.URL)
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
		fix.conf.MustSet(ctx, config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1ns")
		t.Cleanup(func() {
			fix.conf.MustSet(ctx, config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "5m")
		})

		run := func(t *testing.T, spa bool) {
			id := fix.createIdentityWithoutPasskey(t)

			body, res := doBrowserFlow(t, spa, func(v url.Values) {
				v.Set(node.PasskeySettingsRegister, "{}")
				v.Set("method", "passkey")
			}, id)

			if spa {
				assert.NotEmpty(t, gjson.Get(body, "redirect_browser_to").String())
				assert.Equal(t, http.StatusForbidden, res.StatusCode)
				assertx.EqualAsJSONExcept(t, settings.NewFlowNeedsReAuth(), json.RawMessage(body), []string{"redirect_browser_to"})
			} else {
				assert.Contains(t, res.Request.URL.String(), fix.loginTS.URL+"/login-ts")
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

	t.Run("case=add a passkey", func(t *testing.T) {
		run := func(t *testing.T, spa bool) {
			// We load our identity which we will use to replay the webauth session
			var id identity.Identity
			require.NoError(t, json.Unmarshal(settingsFixtureSuccessIdentity, &id))
			id.NID = x.NewUUID()
			_ = fix.reg.PrivilegedIdentityPool().DeleteIdentity(fix.ctx, id.ID)
			browserClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, ctx, fix.reg, &id)
			f := testhelpers.InitializeSettingsFlowViaBrowser(t, browserClient, spa, fix.publicTS)

			// We inject the session to replay
			interim, err := fix.reg.SettingsFlowPersister().GetSettingsFlow(fix.ctx, uuid.FromStringOrNil(f.Id))
			require.NoError(t, err)
			interim.InternalContext = settingsFixtureSuccessInternalContext
			require.NoError(t, fix.reg.SettingsFlowPersister().UpdateSettingsFlow(fix.ctx, interim))

			values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)

			// We use the response replay
			values.Set("method", "passkey")
			values.Set(node.PasskeySettingsRegister, string(settingsFixtureSuccessResponse))
			body, res := testhelpers.SettingsMakeRequest(t, false, spa, f, browserClient, testhelpers.EncodeFormAsJSON(t, spa, values))
			require.Equal(t, http.StatusOK, res.StatusCode, "%s", body)

			if spa {
				assert.Contains(t, res.Request.URL.String(), fix.publicTS.URL+settings.RouteSubmitFlow)
			} else {
				assert.Contains(t, res.Request.URL.String(), fix.uiTS.URL)
			}
			assert.EqualValues(t, flow.StateSuccess, gjson.Get(body, "state").String(), body)

			actual, err := fix.reg.Persister().GetIdentityConfidential(fix.ctx, id.ID)
			require.NoError(t, err)
			cred, ok := actual.GetCredentials(identity.CredentialsTypePasskey)
			require.True(t, ok)
			assert.Len(t, gjson.GetBytes(cred.Config, "credentials").Array(), 1)

			actualFlow, err := fix.reg.SettingsFlowPersister().GetSettingsFlow(fix.ctx, uuid.FromStringOrNil(f.Id))
			require.NoError(t, err)
			// new session data has been generated
			assert.NotEqual(t, settingsFixtureSuccessInternalContext,
				gjson.GetBytes(actualFlow.InternalContext,
					flow.PrefixInternalContextKey(identity.CredentialsTypePasskey, passkey.InternalContextKeySessionData)))

			testhelpers.EnsureAAL(t, browserClient, fix.publicTS, "aal1", string(identity.CredentialsTypePasskey))

			if spa {
				assert.EqualValues(t, flow.ContinueWithActionRedirectBrowserToString, gjson.Get(body, "continue_with.0.action").String(), "%s", body)
				assert.Contains(t, gjson.Get(body, "continue_with.0.redirect_browser_to").String(), fix.uiTS.URL, "%s", body)
			} else {
				assert.Empty(t, gjson.Get(body, "continue_with").Array(), "%s", body)
			}
		}

		t.Run("type=browser", func(t *testing.T) {
			run(t, false)
		})

		t.Run("type=spa", func(t *testing.T) {
			run(t, true)
		})
	})

	t.Run("case=fails to remove passkey if it is the last credential available", func(t *testing.T) {
		run := func(t *testing.T, spa bool) {
			id := fix.createIdentity(t)
			id.DeleteCredentialsType(identity.CredentialsTypePassword)
			conf := sqlxx.JSONRawMessage(`{"credentials":[{"id":"Zm9vZm9v","display_name":"foo","is_passwordless":true}]}`)
			id.UpsertCredentialsConfig(identity.CredentialsTypePasskey, conf, 0)
			require.NoError(t, fix.reg.IdentityManager().Update(ctx, id, identity.ManagerAllowWriteProtectedTraits))

			body, res := doBrowserFlow(t, spa, func(v url.Values) {
				// The remove key should be empty
				snapshotx.SnapshotT(t, v, snapshotx.ExceptPaths("csrf_token", "passkey_create_data"))
				v.Set(node.PasskeyRemove, "666f6f666f6f")
			}, id)

			if spa {
				assert.Contains(t, res.Request.URL.String(), fix.publicTS.URL+settings.RouteSubmitFlow)
			} else {
				assert.Contains(t, res.Request.URL.String(), fix.uiTS.URL)
			}

			t.Run("response", func(t *testing.T) {
				assert.EqualValues(t, flow.StateShowForm, gjson.Get(body, "state").String(), body)
				snapshotx.SnapshotTExcept(t, json.RawMessage(gjson.Get(body, "ui.nodes.#(attributes.name==passkey_remove)").String()), nil)

				actual, err := fix.reg.Persister().GetIdentityConfidential(fix.ctx, id.ID)
				require.NoError(t, err)
				cred, ok := actual.GetCredentials(identity.CredentialsTypePasskey)
				assert.True(t, ok)
				assert.Len(t, gjson.GetBytes(cred.Config, "credentials").Array(), 1)
			})
		}

		t.Run("type=browser", func(t *testing.T) {
			run(t, false)
		})

		t.Run("type=spa", func(t *testing.T) {
			run(t, true)
		})
	})

	t.Run("case=remove all passkeys", func(t *testing.T) {
		run := func(t *testing.T, spa bool) {
			id := fix.createIdentity(t)
			allCred, ok := id.GetCredentials(identity.CredentialsTypePasskey)
			assert.True(t, ok)

			var cc identity.CredentialsWebAuthnConfig
			require.NoError(t, json.Unmarshal(allCred.Config, &cc))
			require.Len(t, cc.Credentials, 2)

			for _, cred := range cc.Credentials {
				body, res := doBrowserFlow(t, spa, func(v url.Values) {
					v.Set(node.PasskeyRemove, fmt.Sprintf("%x", cred.ID))
				}, id)

				if spa {
					assert.Contains(t, res.Request.URL.String(), fix.publicTS.URL+settings.RouteSubmitFlow)
				} else {
					assert.Contains(t, res.Request.URL.String(), fix.uiTS.URL)
				}
				assert.EqualValues(t, flow.StateSuccess, gjson.Get(body, "state").String(), body)
			}

			actual, err := fix.reg.Persister().GetIdentityConfidential(fix.ctx, id.ID)
			require.NoError(t, err)
			_, ok = actual.GetCredentials(identity.CredentialsTypePasskey)
			assert.False(t, ok)
			// Check not to remove other credentials with webauthn
			_, ok = actual.GetCredentials(identity.CredentialsTypePassword)
			assert.True(t, ok)
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
			id := fix.createIdentity(t)
			body, res := doBrowserFlow(t, spa, func(v url.Values) {
				v.Set(node.PasskeyRemove, fmt.Sprintf("%x", []byte("foofoo")))
			}, id)

			if spa {
				assert.Contains(t, res.Request.URL.String(), fix.publicTS.URL+settings.RouteSubmitFlow)
			} else {
				assert.Contains(t, res.Request.URL.String(), fix.uiTS.URL)
			}
			assert.EqualValues(t, flow.StateSuccess, json.RawMessage(gjson.Get(body, "state").String()))

			actual, err := fix.reg.Persister().GetIdentityConfidential(fix.ctx, id.ID)
			require.NoError(t, err)
			cred, ok := actual.GetCredentials(identity.CredentialsTypePasskey)
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

	t.Run("case=is not responsible if neither remove or register is set", func(t *testing.T) {
		run := func(t *testing.T, spa bool) {
			id := fix.createIdentity(t)
			body, res := doBrowserFlow(t, spa, func(v url.Values) {
				v.Set(node.PasskeyRemove, "")
				v.Set(node.PasskeyRegister, "")
			}, id)

			if spa {
				assert.Contains(t, res.Request.URL.String(), fix.publicTS.URL+settings.RouteSubmitFlow)
			} else {
				assert.Contains(t, res.Request.URL.String(), fix.uiTS.URL)
			}

			assert.Equal(t, text.NewErrorValidationSettingsNoStrategyFound().Text, gjson.Get(body, "ui.messages.0.text").String(), "%s", body)
		}

		t.Run("type=browser", func(t *testing.T) {
			run(t, false)
		})

		t.Run("type=spa", func(t *testing.T) {
			run(t, true)
		})
	})

	t.Run("case=should fail if no identifier was set in the schema", func(t *testing.T) {
		testhelpers.SetDefaultIdentitySchema(fix.conf, "file://stub/missing-identifier.schema.json")

		for _, f := range []string{"spa", "browser"} {
			t.Run("type="+f, func(t *testing.T) {
				isSPA := f == "spa"

				var id identity.Identity
				require.NoError(t, json.Unmarshal(settingsFixtureSuccessIdentity, &id))
				_ = fix.reg.PrivilegedIdentityPool().DeleteIdentity(fix.ctx, id.ID)
				id.NID = x.NewUUID()
				browserClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, ctx, fix.reg, &id)

				req, err := http.NewRequest("GET", fix.publicTS.URL+settings.RouteInitBrowserFlow, nil)
				require.NoError(t, err)
				if isSPA {
					req.Header.Set("Accept", "application/json")
				}
				res, err := browserClient.Do(req)
				require.NoError(t, err)

				actual := x.MustReadAll(res.Body)
				defer func() { _ = res.Body.Close() }()

				assert.Equal(t, text.NewErrorValidationIdentifierMissing().Text, gjson.GetBytes(actual, "ui.messages.0.text").String(), "%s", actual)
			})
		}
	})
}
