// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package lookup_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/ory/x/sqlcon"

	"github.com/gofrs/uuid"

	kratos "github.com/ory/kratos/internal/httpclient"
	"github.com/ory/kratos/selfservice/flow"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/strategy/lookup"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/assertx"
	"github.com/ory/x/sqlxx"
)

func createIdentityWithoutLookup(t *testing.T, reg driver.Registry) *identity.Identity {
	id, _ := createIdentity(t, reg)
	delete(id.Credentials, identity.CredentialsTypeLookup)
	require.NoError(t, reg.PrivilegedIdentityPool().UpdateIdentity(context.Background(), id))
	return id
}

func createIdentity(t *testing.T, reg driver.Registry) (*identity.Identity, []identity.RecoveryCode) {
	codes := make([]identity.RecoveryCode, 12)
	for k := range codes {
		var usedAt sqlxx.NullTime
		if k%3 == 1 {
			usedAt = sqlxx.NullTime(time.Unix(int64(1629199958+k), 0))
		}
		codes[k] = identity.RecoveryCode{Code: fmt.Sprintf("key-%d", k), UsedAt: usedAt}
	}
	identifier := x.NewUUID().String() + "@ory.sh"
	password := x.NewUUID().String()
	p, err := reg.Hasher(context.Background()).Generate(context.Background(), []byte(password))
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

	rc, err := json.Marshal(&identity.CredentialsLookupConfig{RecoveryCodes: codes})
	require.NoError(t, err)
	require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))
	i.Credentials = map[identity.CredentialsType]identity.Credentials{
		identity.CredentialsTypePassword: {
			Type:        identity.CredentialsTypePassword,
			Identifiers: []string{identifier},
			Config:      sqlxx.JSONRawMessage(`{"hashed_password":"` + string(p) + `"}`),
		},
		identity.CredentialsTypeLookup: {
			Type:        identity.CredentialsTypeLookup,
			Identifiers: []string{i.ID.String()},
			Config:      rc,
		},
	}

	require.NoError(t, reg.PrivilegedIdentityPool().UpdateIdentity(context.Background(), i))
	return i, codes
}

func TestCompleteSettings(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword)+".enabled", false)
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".profile.enabled", false)
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeLookup)+".enabled", true)
	conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, "aal1")

	router := x.NewRouterPublic()
	publicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin())

	errTS := testhelpers.NewErrorTestServer(t, reg)
	uiTS := testhelpers.NewSettingsUIFlowEchoServer(t, reg)
	_ = testhelpers.NewRedirSessionEchoTS(t, reg)
	loginTS := testhelpers.NewLoginUIFlowEchoServer(t, reg)

	conf.MustSet(ctx, config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1m")

	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/login.schema.json")
	conf.MustSet(ctx, config.ViperKeySecretsDefault, []string{"not-a-secure-session-key"})

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

	t.Run("case=hide recovery codes behind reveal button and show disable button", func(t *testing.T) {
		id, _ := createIdentity(t, reg)
		browserClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)

		t.Run("case=spa", func(t *testing.T) {
			f := testhelpers.InitializeSettingsFlowViaBrowser(t, browserClient, true, publicTS)
			testhelpers.SnapshotTExcept(t, f.Ui.Nodes, []string{"0.attributes.value"})
		})

		t.Run("case=browser", func(t *testing.T) {
			f := testhelpers.InitializeSettingsFlowViaBrowser(t, browserClient, false, publicTS)
			testhelpers.SnapshotTExcept(t, f.Ui.Nodes, []string{"0.attributes.value"})
		})

		t.Run("case=api", func(t *testing.T) {
			apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)
			f := testhelpers.InitializeSettingsFlowViaAPI(t, apiClient, publicTS)
			testhelpers.SnapshotTExcept(t, f.Ui.Nodes, []string{"0.attributes.value"})
		})
	})

	t.Run("case=button for regeneration is displayed when identity has no recovery codes yet", func(t *testing.T) {
		id := createIdentityWithoutLookup(t, reg)
		browserClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)

		t.Run("case=spa", func(t *testing.T) {
			f := testhelpers.InitializeSettingsFlowViaBrowser(t, browserClient, true, publicTS)
			testhelpers.SnapshotTExcept(t, f.Ui.Nodes, []string{"0.attributes.value"})
		})

		t.Run("case=browser", func(t *testing.T) {
			f := testhelpers.InitializeSettingsFlowViaBrowser(t, browserClient, false, publicTS)
			testhelpers.SnapshotTExcept(t, f.Ui.Nodes, []string{"0.attributes.value"})
		})

		t.Run("case=api", func(t *testing.T) {
			apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)
			f := testhelpers.InitializeSettingsFlowViaAPI(t, apiClient, publicTS)
			testhelpers.SnapshotTExcept(t, f.Ui.Nodes, []string{"0.attributes.value"})
		})
	})

	t.Run("case=should pass without csrf if API flow", func(t *testing.T) {
		id, _ := createIdentity(t, reg)

		body, res := doAPIFlow(t, func(v url.Values) {
			v.Del("csrf_token")
			v.Set(node.LookupReveal, "true")
		}, id)

		assert.Contains(t, res.Request.URL.String(), publicTS.URL+settings.RouteSubmitFlow)
		testhelpers.SnapshotTExcept(t, json.RawMessage(gjson.Get(body, "ui.nodes").Raw), []string{"0.attributes.value"})
	})

	t.Run("case=should fail if CSRF token is invalid", func(t *testing.T) {
		id := createIdentityWithoutLookup(t, reg)

		t.Run("type=browser", func(t *testing.T) {
			body, res := doBrowserFlow(t, false, func(v url.Values) {
				v.Del("csrf_token")
				v.Set(node.LookupReveal, "true")
			}, id)

			assert.Contains(t, res.Request.URL.String(), errTS.URL)
			assert.Equal(t, x.ErrInvalidCSRFToken.Reason(), gjson.Get(body, "reason").String(), body)
		})

		t.Run("type=spa", func(t *testing.T) {
			body, res := doBrowserFlow(t, true, func(v url.Values) {
				v.Del("csrf_token")
				v.Set(node.LookupReveal, "true")
			}, id)

			assert.Contains(t, res.Request.URL.String(), publicTS.URL+settings.RouteSubmitFlow)
			assert.Equal(t, x.ErrInvalidCSRFToken.Reason(), gjson.Get(body, "error.reason").String(), body)
		})
	})

	t.Run("type=can not reveal or regenerate or remove without privileged session", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "1ns")
		t.Cleanup(func() {
			conf.MustSet(ctx, config.ViperKeySelfServiceSettingsPrivilegedAuthenticationAfter, "5m")
		})

		id, codes := createIdentity(t, reg)

		checkIdentity := func(t *testing.T) {
			_, cred, err := reg.PrivilegedIdentityPool().FindByCredentialsIdentifier(context.Background(), identity.CredentialsTypeLookup, id.ID.String())
			require.NoError(t, err)
			assertx.EqualAsJSON(t, codes, json.RawMessage(gjson.GetBytes(cred.Config, "recovery_codes").Raw))
		}

		for _, tc := range []struct {
			v func(v url.Values)
			d string
		}{
			{
				d: "reveal",
				v: func(v url.Values) {
					v.Set(node.LookupReveal, "true")
				},
			},
			{
				d: "regenerate",
				v: func(v url.Values) {
					v.Set(node.LookupRegenerate, "true")
				},
			},
			{
				d: "disable",
				v: func(v url.Values) {
					v.Set(node.LookupDisable, "true")
				},
			},
		} {
			t.Run("case="+tc.d, func(t *testing.T) {
				payload := tc.v
				t.Run("type=api", func(t *testing.T) {
					actual, res := doAPIFlow(t, payload, id)
					assert.Equal(t, http.StatusForbidden, res.StatusCode)
					assert.Contains(t, gjson.Get(actual, "redirect_browser_to").String(), publicTS.URL+"/self-service/login/browser?refresh=true&return_to=")
					assertx.EqualAsJSONExcept(t, settings.NewFlowNeedsReAuth(), json.RawMessage(actual), []string{"redirect_browser_to"})
					checkIdentity(t)
				})

				t.Run("type=spa", func(t *testing.T) {
					actual, res := doBrowserFlow(t, true, payload, id)
					assert.Equal(t, http.StatusForbidden, res.StatusCode)
					assert.Contains(t, gjson.Get(actual, "redirect_browser_to").String(), publicTS.URL+"/self-service/login/browser?refresh=true&return_to=")
					assertx.EqualAsJSONExcept(t, settings.NewFlowNeedsReAuth(), json.RawMessage(actual), []string{"redirect_browser_to"})
					checkIdentity(t)
				})

				t.Run("type=browser", func(t *testing.T) {
					actual, res := doBrowserFlow(t, false, payload, id)
					assert.Equal(t, http.StatusOK, res.StatusCode)
					assert.Contains(t, res.Request.URL.String(), loginTS.URL+"/login-ts")
					assertx.EqualAsJSON(t, text.NewInfoLoginReAuth().Text, json.RawMessage(gjson.Get(actual, "ui.messages.0.text").Raw), actual)
					checkIdentity(t)
				})
			})
		}
	})

	t.Run("type=can not confirm without regenerate", func(t *testing.T) {
		id, codes := createIdentity(t, reg)
		payload := func(v url.Values) {
			v.Set(node.LookupConfirm, "true")
		}

		const reason = "You must (re-)generate recovery backup codes before you can save them."

		checkIdentity := func(t *testing.T) {
			_, cred, err := reg.PrivilegedIdentityPool().FindByCredentialsIdentifier(context.Background(), identity.CredentialsTypeLookup, id.ID.String())
			require.NoError(t, err)
			assertx.EqualAsJSON(t, codes, json.RawMessage(gjson.GetBytes(cred.Config, "recovery_codes").Raw))
		}

		t.Run("type=api", func(t *testing.T) {
			actual, res := doAPIFlow(t, payload, id)
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
			assert.EqualValues(t, reason, gjson.Get(actual, "ui.messages.0.text").String(), "%s", actual)
			checkIdentity(t)
		})

		t.Run("type=spa", func(t *testing.T) {
			actual, res := doBrowserFlow(t, true, payload, id)
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
			assert.EqualValues(t, reason, gjson.Get(actual, "ui.messages.0.text").String(), "%s", actual)
			checkIdentity(t)
		})

		t.Run("type=browser", func(t *testing.T) {
			actual, res := doBrowserFlow(t, false, payload, id)
			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Contains(t, res.Request.URL.String(), uiTS.URL)
			assert.EqualValues(t, reason, gjson.Get(actual, "ui.messages.0.text").String(), "%s", actual)
			checkIdentity(t)
		})
	})

	t.Run("type=regenerate but no confirmation", func(t *testing.T) {
		id, codes := createIdentity(t, reg)
		payload := func(v url.Values) {
			v.Set(node.LookupRegenerate, "true")
		}

		checkIdentity := func(t *testing.T) {
			_, cred, err := reg.PrivilegedIdentityPool().FindByCredentialsIdentifier(context.Background(), identity.CredentialsTypeLookup, id.ID.String())
			require.NoError(t, err)
			assertx.EqualAsJSON(t, codes, json.RawMessage(gjson.GetBytes(cred.Config, "recovery_codes").Raw))
		}

		t.Run("type=api", func(t *testing.T) {
			actual, res := doAPIFlow(t, payload, id)
			assert.Equal(t, http.StatusOK, res.StatusCode)
			testhelpers.SnapshotTExcept(t, json.RawMessage(gjson.Get(actual, "ui.nodes").Raw), []string{"0.attributes.value"})
			checkIdentity(t)
		})

		t.Run("type=spa", func(t *testing.T) {
			actual, res := doBrowserFlow(t, true, payload, id)
			assert.Equal(t, http.StatusOK, res.StatusCode)
			testhelpers.SnapshotTExcept(t, json.RawMessage(gjson.Get(actual, "ui.nodes").Raw), []string{"0.attributes.value"})
			checkIdentity(t)
		})

		t.Run("type=browser", func(t *testing.T) {
			actual, res := doBrowserFlow(t, false, payload, id)
			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Contains(t, res.Request.URL.String(), uiTS.URL)
			testhelpers.SnapshotTExcept(t, json.RawMessage(gjson.Get(actual, "ui.nodes").Raw), []string{"0.attributes.value"})
			checkIdentity(t)
		})
	})

	t.Run("type=regenerate with confirmation", func(t *testing.T) {
		for _, tc := range []struct {
			d string
			c func(t *testing.T) *identity.Identity
		}{
			{
				d: "with",
				c: func(t *testing.T) *identity.Identity {
					i, _ := createIdentity(t, reg)
					return i
				},
			},
			{
				d: "without",
				c: func(t *testing.T) *identity.Identity {
					return createIdentityWithoutLookup(t, reg)
				},
			},
		} {
			t.Run("credentials="+tc.d, func(t *testing.T) {
				payload := func(v url.Values) {
					v.Del(node.LookupReveal)
					v.Del(node.LookupDisable)
					v.Set(node.LookupRegenerate, "true")
				}

				payloadConfirm := func(v url.Values) {
					v.Del(node.LookupRegenerate)
					v.Del(node.LookupDisable)
					v.Del(node.LookupReveal)
					v.Set(node.LookupConfirm, "true")
				}

				checkIdentity := func(t *testing.T, id *identity.Identity, f *kratos.SettingsFlow) {
					_, cred, err := reg.PrivilegedIdentityPool().FindByCredentialsIdentifier(context.Background(), identity.CredentialsTypeLookup, id.ID.String())
					require.NoError(t, err)
					assert.NotContains(t, gjson.GetBytes(cred.Config, "recovery_codes").Raw, "key-1")
					assert.NotContains(t, gjson.GetBytes(cred.Config, "recovery_codes").Raw, "key-0")

					actualFlow, err := reg.SettingsFlowPersister().GetSettingsFlow(context.Background(), uuid.FromStringOrNil(f.Id))
					require.NoError(t, err)
					assert.Empty(t, gjson.GetBytes(actualFlow.InternalContext, flow.PrefixInternalContextKey(identity.CredentialsTypeLookup, lookup.InternalContextKeyRegenerated)))
				}

				t.Run("type=api", func(t *testing.T) {
					id, _ := createIdentity(t, reg)
					apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)
					f := testhelpers.InitializeSettingsFlowViaAPI(t, apiClient, publicTS)
					values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)

					payload(values)
					_, _ = testhelpers.SettingsMakeRequest(t, true, false, f, apiClient, testhelpers.EncodeFormAsJSON(t, true, values))

					payloadConfirm(values)
					actual, res := testhelpers.SettingsMakeRequest(t, true, false, f, apiClient, testhelpers.EncodeFormAsJSON(t, true, values))
					assert.Equal(t, http.StatusOK, res.StatusCode)

					assert.Contains(t, res.Request.URL.String(), publicTS.URL+settings.RouteSubmitFlow)
					assert.EqualValues(t, flow.StateSuccess, json.RawMessage(gjson.Get(actual, "state").String()))

					checkIdentity(t, id, f)
					testhelpers.EnsureAAL(t, apiClient, publicTS, "aal2", string(identity.CredentialsTypeLookup))
				})

				runBrowser := func(t *testing.T, spa bool) {
					id, _ := createIdentity(t, reg)

					browserClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id)
					f := testhelpers.InitializeSettingsFlowViaBrowser(t, browserClient, spa, publicTS)
					values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)

					payload(values)
					_, _ = testhelpers.SettingsMakeRequest(t, false, spa, f, browserClient, testhelpers.EncodeFormAsJSON(t, spa, values))

					payloadConfirm(values)
					actual, res := testhelpers.SettingsMakeRequest(t, false, spa, f, browserClient, testhelpers.EncodeFormAsJSON(t, spa, values))
					assert.Equal(t, http.StatusOK, res.StatusCode)

					if spa {
						assert.Contains(t, res.Request.URL.String(), publicTS.URL+settings.RouteSubmitFlow)
					} else {
						assert.Contains(t, res.Request.URL.String(), uiTS.URL)
					}

					assert.EqualValues(t, flow.StateSuccess, json.RawMessage(gjson.Get(actual, "state").String()))
					checkIdentity(t, id, f)
					testhelpers.EnsureAAL(t, browserClient, publicTS, "aal2", string(identity.CredentialsTypeLookup))
				}

				t.Run("type=browser", func(t *testing.T) {
					runBrowser(t, false)
				})

				t.Run("type=spa", func(t *testing.T) {
					runBrowser(t, true)
				})
			})
		}
	})

	t.Run("type=remove lookup codes", func(t *testing.T) {
		for _, tc := range []struct {
			d string
			c func(t *testing.T) *identity.Identity
		}{
			{
				d: "with",
				c: func(t *testing.T) *identity.Identity {
					i, _ := createIdentity(t, reg)
					return i
				},
			},
			{
				d: "without",
				c: func(t *testing.T) *identity.Identity {
					return createIdentityWithoutLookup(t, reg)
				},
			},
		} {
			t.Run("credentials="+tc.d, func(t *testing.T) {
				payloadConfirm := func(v url.Values) {
					v.Del(node.LookupRegenerate)
					v.Del(node.LookupReveal)
					v.Set(node.LookupDisable, "true")
				}

				checkIdentity := func(t *testing.T, id *identity.Identity, f *kratos.SettingsFlow) {
					_, _, err := reg.PrivilegedIdentityPool().FindByCredentialsIdentifier(context.Background(), identity.CredentialsTypeLookup, id.ID.String())
					require.ErrorIs(t, err, sqlcon.ErrNoRows)

					actualFlow, err := reg.SettingsFlowPersister().GetSettingsFlow(context.Background(), uuid.FromStringOrNil(f.Id))
					require.NoError(t, err)
					assert.Empty(t, gjson.GetBytes(actualFlow.InternalContext, flow.PrefixInternalContextKey(identity.CredentialsTypeLookup, lookup.InternalContextKeyRegenerated)).Raw)
				}

				t.Run("type=api", func(t *testing.T) {
					id, _ := createIdentity(t, reg)
					apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)
					f := testhelpers.InitializeSettingsFlowViaAPI(t, apiClient, publicTS)
					values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)

					payloadConfirm(values)
					actual, res := testhelpers.SettingsMakeRequest(t, true, false, f, apiClient, testhelpers.EncodeFormAsJSON(t, true, values))
					assert.Equal(t, http.StatusOK, res.StatusCode)

					assert.Contains(t, res.Request.URL.String(), publicTS.URL+settings.RouteSubmitFlow)
					assert.EqualValues(t, flow.StateSuccess, json.RawMessage(gjson.Get(actual, "state").String()))

					checkIdentity(t, id, f)
					testhelpers.EnsureAAL(t, apiClient, publicTS, "aal1")
				})

				runBrowser := func(t *testing.T, spa bool) {
					id, _ := createIdentity(t, reg)

					browserClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id)
					f := testhelpers.InitializeSettingsFlowViaBrowser(t, browserClient, spa, publicTS)
					values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)

					payloadConfirm(values)
					actual, res := testhelpers.SettingsMakeRequest(t, false, spa, f, browserClient, testhelpers.EncodeFormAsJSON(t, spa, values))
					assert.Equal(t, http.StatusOK, res.StatusCode)

					if spa {
						assert.Contains(t, res.Request.URL.String(), publicTS.URL+settings.RouteSubmitFlow)
					} else {
						assert.Contains(t, res.Request.URL.String(), uiTS.URL)
					}

					assert.EqualValues(t, flow.StateSuccess, json.RawMessage(gjson.Get(actual, "state").String()))
					checkIdentity(t, id, f)
					testhelpers.EnsureAAL(t, browserClient, publicTS, "aal1")
				}

				t.Run("type=browser", func(t *testing.T) {
					runBrowser(t, false)
				})

				t.Run("type=spa", func(t *testing.T) {
					runBrowser(t, true)
				})
			})
		}
	})
}
