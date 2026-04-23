// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package passkey_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/ory/kratos/x/nosurfx"
	"github.com/ory/x/configx"

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
	"github.com/ory/x/contextx"
	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg"
	kratos "github.com/ory/kratos/pkg/httpclient"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/kratos/x"
)

//go:embed fixtures/settings/success/identity.json
var settingsFixtureSuccessIdentity []byte

//go:embed fixtures/settings/success/response.json
var settingsFixtureSuccessResponse []byte

//go:embed fixtures/settings/success/internal_context.json
var settingsFixtureSuccessInternalContext []byte

func TestCompleteSettings(t *testing.T) {
	fix := newSettingsFixture(t)

	t.Run("case=invalid passkey config", func(t *testing.T) {
		fix := newSettingsFixture(t, configx.WithValue(config.ViperKeyPasskeyRPID, ""))
		id := fix.createIdentity(t)
		apiClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t.Context(), t, fix.reg, id)

		req, err := http.NewRequest("GET", fix.publicTS.URL+settings.RouteInitBrowserFlow, nil)
		require.NoError(t, err)
		req.Header.Set("Accept", "application/json")
		res, err := apiClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	})

	t.Run("case=a device is shown which can be unlinked", func(t *testing.T) {
		id := fix.createIdentity(t)

		apiClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t.Context(), t, fix.reg, id)
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
		require.NoError(t, fix.reg.PrivilegedIdentityPool().UpdateIdentity(t.Context(), id))

		apiClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t.Context(), t, fix.reg, id)
		f := testhelpers.InitializeSettingsFlowViaBrowser(t, apiClient, true, fix.publicTS)

		testhelpers.SnapshotTExcept(t, f.Ui.Nodes, []string{
			"0.attributes.value", // CSRF
			"1.attributes.nonce", // script
			"1.attributes.src",   // script
			"4.attributes.value", // passkey_create_data
		})
	})

	t.Run("case=passkey nodes exist for API but without browser script", func(t *testing.T) {
		id := fix.createIdentityWithoutPasskey(t)
		require.NoError(t, fix.reg.PrivilegedIdentityPool().UpdateIdentity(t.Context(), id))

		apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t.Context(), t, fix.reg, id)
		f := testhelpers.InitializeSettingsFlowViaAPI(t, apiClient, fix.publicTS)

		// Passkey nodes should exist for API
		hasPasskeyGroup := false
		hasWebAuthnScript := false
		for _, n := range f.Ui.Nodes {
			if n.Group == "passkey" {
				hasPasskeyGroup = true
			}
			if n.Type == "script" {
				hasWebAuthnScript = true
			}
		}
		assert.True(t, hasPasskeyGroup, "passkey group should be present for API flows")
		assert.False(t, hasWebAuthnScript, "WebAuthn script should not be present for API flows")
	})

	doAPIFlow := func(t *testing.T, v func(url.Values), id *identity.Identity) (string, *http.Response) {
		apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t.Context(), t, fix.reg, id)
		f := testhelpers.InitializeSettingsFlowViaAPI(t, apiClient, fix.publicTS)
		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		v(values)
		payload := testhelpers.EncodeFormAsJSON(t, true, values)
		return testhelpers.SettingsMakeRequest(t, true, false, f, apiClient, payload)
	}

	doBrowserFlow := func(t *testing.T, spa bool, v func(url.Values), id *identity.Identity) (string, *http.Response) {
		browserClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t.Context(), t, fix.reg, id)
		f := testhelpers.InitializeSettingsFlowViaBrowser(t, browserClient, spa, fix.publicTS)
		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		v(values)
		return testhelpers.SettingsMakeRequest(t, false, spa, f, browserClient, testhelpers.EncodeFormAsJSON(t, spa, values))
	}

	t.Run("case=fails with invalid passkey payload for API", func(t *testing.T) {
		id := fix.createIdentityWithoutPasskey(t)
		body, res := doAPIFlow(t, func(v url.Values) {
			v.Set(node.PasskeySettingsRegister, "{}")
			v.Set("method", "passkey")
		}, id)

		assert.Contains(t, res.Request.URL.String(), fix.publicTS.URL+settings.RouteSubmitFlow)
		// API should process the request but fail on invalid WebAuthn payload
		assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "Parse error for Registration", "%s", body)
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
				assert.Equal(t, nosurfx.ErrInvalidCSRFToken().Reason(), gjson.Get(body, "error.reason").String(), body)
			} else {
				assert.Contains(t, res.Request.URL.String(), fix.errTS.URL)
				assert.Equal(t, nosurfx.ErrInvalidCSRFToken().Reason(), gjson.Get(body, "reason").String(), body)
			}
		}

		t.Run("type=browser", func(t *testing.T) {
			run(t, false)
		})

		t.Run("type=spa", func(t *testing.T) {
			run(t, true)
		})

		t.Run("type=api", func(t *testing.T) {
			// API should work without CSRF token
			id := fix.createIdentityWithoutPasskey(t)
			body, res := doAPIFlow(t, func(v url.Values) {
				v.Del("csrf_token")
				v.Set(node.PasskeySettingsRegister, "{}")
				v.Set("method", "passkey")
			}, id)
			assert.Contains(t, res.Request.URL.String(), fix.publicTS.URL+settings.RouteSubmitFlow)
			// API should process and fail on invalid payload, not CSRF
			assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "Parse error for Registration", "%s", body)
		})
	})

	t.Run("case=fails when register payload is invalid", func(t *testing.T) {
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

		t.Run("type=api", func(t *testing.T) {
			id := fix.createIdentityWithoutPasskey(t)
			body, res := doAPIFlow(t, func(v url.Values) {
				v.Set(node.PasskeySettingsRegister, "{}")
				v.Set("method", "passkey")
			}, id)
			assert.Contains(t, res.Request.URL.String(), fix.publicTS.URL+settings.RouteSubmitFlow)
			assert.Contains(t, gjson.Get(body, "ui.messages.0.text").String(), "Parse error for Registration", "%s", body)
		})
	})

	t.Run("case=requires privileged session for register", func(t *testing.T) {
		fix.conf.MustSet(t.Context(), config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1ns")
		t.Cleanup(func() {
			fix.conf.MustSet(t.Context(), config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "5m")
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

		t.Run("type=api", func(t *testing.T) {
			id := fix.createIdentityWithoutPasskey(t)
			body, res := doAPIFlow(t, func(v url.Values) {
				v.Set(node.PasskeySettingsRegister, "{}")
				v.Set("method", "passkey")
			}, id)

			assert.Contains(t, res.Request.URL.String(), fix.publicTS.URL+settings.RouteSubmitFlow)
			// API flow should fail with 403 Forbidden due to insufficient privileges
			assert.Equal(t, http.StatusForbidden, res.StatusCode)
			assertx.EqualAsJSON(t, settings.NewFlowNeedsReAuth().DefaultError, json.RawMessage(gjson.Get(body, "error").Raw))
		})
	})

	t.Run("case=add a passkey", func(t *testing.T) {
		run := func(t *testing.T, flowType string) {
			// We load our identity which we will use to replay the webauth session
			var id identity.Identity
			require.NoError(t, json.Unmarshal(settingsFixtureSuccessIdentity, &id))
			id.NID = x.NewUUID()
			_ = fix.reg.PrivilegedIdentityPool().DeleteIdentity(t.Context(), id.ID)

			var f *kratos.SettingsFlow
			var client *http.Client
			var body string
			var res *http.Response

			if flowType == "api" {
				client = testhelpers.NewHTTPClientWithIdentitySessionToken(t.Context(), t, fix.reg, &id)
				f = testhelpers.InitializeSettingsFlowViaAPI(t, client, fix.publicTS)
			} else {
				client = testhelpers.NewHTTPClientWithIdentitySessionCookie(t.Context(), t, fix.reg, &id)
				f = testhelpers.InitializeSettingsFlowViaBrowser(t, client, flowType == "spa", fix.publicTS)
			}

			// We inject the session to replay
			interim, err := fix.reg.SettingsFlowPersister().GetSettingsFlow(t.Context(), uuid.FromStringOrNil(f.Id))
			require.NoError(t, err)
			interim.InternalContext = settingsFixtureSuccessInternalContext
			require.NoError(t, fix.reg.SettingsFlowPersister().UpdateSettingsFlow(t.Context(), interim))

			values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)

			// We use the response replay
			values.Set("method", "passkey")
			values.Set(node.PasskeySettingsRegister, string(settingsFixtureSuccessResponse))

			if flowType == "api" {
				body, res = testhelpers.SettingsMakeRequest(t, true, false, f, client, testhelpers.EncodeFormAsJSON(t, true, values))
			} else {
				body, res = testhelpers.SettingsMakeRequest(t, false, flowType == "spa", f, client, testhelpers.EncodeFormAsJSON(t, flowType == "spa", values))
			}
			require.Equal(t, http.StatusOK, res.StatusCode, "%s", body)

			switch flowType {
			case "spa":
				assert.Contains(t, res.Request.URL.String(), fix.publicTS.URL+settings.RouteSubmitFlow)
			case "browser":
				assert.Contains(t, res.Request.URL.String(), fix.uiTS.URL)
			}
			// For API, no redirect check needed

			assert.EqualValues(t, flow.StateSuccess, gjson.Get(body, "state").String(), body)

			actual, err := fix.reg.Persister().GetIdentityConfidential(t.Context(), id.ID)
			require.NoError(t, err)
			cred, ok := actual.GetCredentials(identity.CredentialsTypePasskey)
			require.True(t, ok)
			assert.Len(t, gjson.GetBytes(cred.Config, "credentials").Array(), 1)

			actualFlow, err := fix.reg.SettingsFlowPersister().GetSettingsFlow(t.Context(), uuid.FromStringOrNil(f.Id))
			require.NoError(t, err)
			// new session data has been generated
			assert.NotEqual(t, settingsFixtureSuccessInternalContext,
				gjson.GetBytes(actualFlow.InternalContext,
					flow.PrefixInternalContextKey(identity.CredentialsTypePasskey, passkey.InternalContextKeySessionData)))

			testhelpers.EnsureAAL(t, client, fix.publicTS, "aal1", string(identity.CredentialsTypePasskey))

			if flowType == "spa" {
				assert.EqualValues(t, flow.ContinueWithActionRedirectBrowserToString, gjson.Get(body, "continue_with.0.action").String(), "%s", body)
				assert.Contains(t, gjson.Get(body, "continue_with.0.redirect_browser_to").String(), fix.uiTS.URL, "%s", body)
			} else {
				assert.Empty(t, gjson.Get(body, "continue_with").Array(), "%s", body)
			}
		}

		t.Run("type=browser", func(t *testing.T) {
			run(t, "browser")
		})

		t.Run("type=spa", func(t *testing.T) {
			run(t, "spa")
		})

		t.Run("type=api", func(t *testing.T) {
			run(t, "api")
		})
	})

	t.Run("case=fails to remove passkey if it is the last credential available", func(t *testing.T) {
		run := func(t *testing.T, flowType string) {
			id := fix.createIdentity(t)
			id.DeleteCredentialsType(identity.CredentialsTypePassword)
			conf := sqlxx.JSONRawMessage(`{"credentials":[{"id":"Zm9vZm9v","display_name":"foo","is_passwordless":true}]}`)
			id.UpsertCredentialsConfig(identity.CredentialsTypePasskey, conf, 0)
			require.NoError(t, fix.reg.IdentityManager().Update(t.Context(), id, identity.ManagerAllowWriteProtectedTraits))

			var body string
			var res *http.Response
			if flowType == "api" {
				body, res = doAPIFlow(t, func(v url.Values) {
					// The remove key should be empty
					snapshotx.SnapshotT(t, v, snapshotx.ExceptPaths("csrf_token", "passkey_create_data"))
					v.Set(node.PasskeyRemove, "666f6f666f6f")
				}, id)
			} else {
				body, res = doBrowserFlow(t, flowType == "spa", func(v url.Values) {
					// The remove key should be empty
					snapshotx.SnapshotT(t, v, snapshotx.ExceptPaths("csrf_token", "passkey_create_data"))
					v.Set(node.PasskeyRemove, "666f6f666f6f")
				}, id)
			}

			switch flowType {
			case "spa":
				assert.Contains(t, res.Request.URL.String(), fix.publicTS.URL+settings.RouteSubmitFlow)
			case "browser":
				assert.Contains(t, res.Request.URL.String(), fix.uiTS.URL)
			}
			// For API, no redirect check needed

			t.Run("response", func(t *testing.T) {
				assert.EqualValues(t, flow.StateShowForm, gjson.Get(body, "state").String(), body)
				snapshotx.SnapshotT(t, json.RawMessage(gjson.Get(body, "ui.nodes.#(attributes.name==passkey_remove)").String()))

				actual, err := fix.reg.Persister().GetIdentityConfidential(t.Context(), id.ID)
				require.NoError(t, err)
				cred, ok := actual.GetCredentials(identity.CredentialsTypePasskey)
				assert.True(t, ok)
				assert.Len(t, gjson.GetBytes(cred.Config, "credentials").Array(), 1)
			})
		}

		t.Run("type=browser", func(t *testing.T) {
			run(t, "browser")
		})

		t.Run("type=spa", func(t *testing.T) {
			run(t, "spa")
		})

		t.Run("type=api", func(t *testing.T) {
			run(t, "api")
		})
	})

	t.Run("case=remove all passkeys", func(t *testing.T) {
		run := func(t *testing.T, flowType string) {
			id := fix.createIdentity(t)
			allCred, ok := id.GetCredentials(identity.CredentialsTypePasskey)
			assert.True(t, ok)

			var cc identity.CredentialsWebAuthnConfig
			require.NoError(t, json.Unmarshal(allCred.Config, &cc))
			require.Len(t, cc.Credentials, 2)

			for _, cred := range cc.Credentials {
				var body string
				var res *http.Response
				if flowType == "api" {
					body, res = doAPIFlow(t, func(v url.Values) {
						v.Set(node.PasskeyRemove, fmt.Sprintf("%x", cred.ID))
					}, id)
				} else {
					body, res = doBrowserFlow(t, flowType == "spa", func(v url.Values) {
						v.Set(node.PasskeyRemove, fmt.Sprintf("%x", cred.ID))
					}, id)
				}

				switch flowType {
				case "spa":
					assert.Contains(t, res.Request.URL.String(), fix.publicTS.URL+settings.RouteSubmitFlow)
				case "browser":
					assert.Contains(t, res.Request.URL.String(), fix.uiTS.URL)
				}
				// For API, no redirect check needed
				assert.EqualValues(t, flow.StateSuccess, gjson.Get(body, "state").String(), body)
			}

			actual, err := fix.reg.Persister().GetIdentityConfidential(t.Context(), id.ID)
			require.NoError(t, err)
			_, ok = actual.GetCredentials(identity.CredentialsTypePasskey)
			assert.False(t, ok)
			// Check not to remove other credentials with webauthn
			_, ok = actual.GetCredentials(identity.CredentialsTypePassword)
			assert.True(t, ok)
		}

		t.Run("type=browser", func(t *testing.T) {
			run(t, "browser")
		})

		t.Run("type=spa", func(t *testing.T) {
			run(t, "spa")
		})

		t.Run("type=api", func(t *testing.T) {
			run(t, "api")
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

			actual, err := fix.reg.Persister().GetIdentityConfidential(t.Context(), id.ID)
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
				_ = fix.reg.PrivilegedIdentityPool().DeleteIdentity(t.Context(), id.ID)
				id.NID = x.NewUUID()
				browserClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t.Context(), t, fix.reg, &id)

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

func TestPopulateSettingsMethod(t *testing.T) {
	ctx := context.Background()
	conf, reg := pkg.NewFastRegistryWithMocks(t)

	ctx = testhelpers.WithDefaultIdentitySchema(ctx, "file://stub/settings.schema.json")
	ctx = contextx.WithConfigValue(ctx, config.ViperKeyPasskeyRPDisplayName, "localhost")
	ctx = contextx.WithConfigValue(ctx, config.ViperKeyPasskeyRPID, "localhost")

	s, err := reg.AllSettingsStrategies().Strategy(string(identity.CredentialsTypePasskey))
	require.NoError(t, err)

	toSnapshot := func(t *testing.T, f node.Nodes, except ...snapshotx.Opt) {
		t.Helper()
		// The CSRF token has a unique value that messes with the snapshot - ignore it.
		f.ResetNodes("csrf_token")
		snapshotx.SnapshotT(t, f, append(except, snapshotx.ExceptNestedKeys("nonce", "src"))...)
	}

	newFlowWithType := func(ctx context.Context, t *testing.T, flowType flow.Type) (*http.Request, *settings.Flow, *identity.Identity) {
		r := httptest.NewRequest("GET", "/self-service/settings/browser", nil)
		r = r.WithContext(ctx)
		t.Helper()
		id := identity.NewIdentity("default")
		id.Traits = identity.Traits(`{"email":"testuser@ory.sh"}`)
		f, err := settings.NewFlow(conf, time.Minute, r, id, flowType)
		f.UI.Nodes = make(node.Nodes, 0)
		require.NoError(t, err)
		return r, f, id
	}

	t.Run("method=PopulateSettingsMethod", func(t *testing.T) {
		t.Run("type=browser", func(t *testing.T) {
			r, f, id := newFlowWithType(ctx, t, flow.TypeBrowser)
			require.NoError(t, s.PopulateSettingsMethod(ctx, r, id, f))
			toSnapshot(t, f.UI.Nodes, snapshotx.ExceptPaths("4.attributes.value"))
		})
		t.Run("type=api", func(t *testing.T) {
			r, f, id := newFlowWithType(ctx, t, flow.TypeAPI)
			require.NoError(t, s.PopulateSettingsMethod(ctx, r, id, f))
			toSnapshot(t, f.UI.Nodes, snapshotx.ExceptPaths("3.attributes.value"))
		})
	})
}
