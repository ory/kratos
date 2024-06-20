// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package passkey_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/ory/kratos/selfservice/strategy/idfirst"

	configtesthelpers "github.com/ory/kratos/driver/config/testhelpers"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/strategy/passkey"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/snapshotx"
)

var (
	//go:embed fixtures/login/success/identity.json
	loginSuccessIdentity []byte
	//go:embed fixtures/login/success/credentials.json
	loginPasswordlessCredentials []byte
	//go:embed fixtures/login/success/internal_context.json
	loginPasswordlessContext []byte
	//go:embed fixtures/login/success/response.json
	loginPasswordlessResponse []byte
)

func TestPopulateLoginMethod(t *testing.T) {
	t.Parallel()
	fix := newLoginFixture(t)
	s := passkey.NewStrategy(fix.reg)

	t.Run("case=should not handle AAL2", func(t *testing.T) {
		loginFlow := &login.Flow{Type: flow.TypeBrowser}
		assert.Nil(t, s.PopulateLoginMethodSecondFactor(nil, loginFlow))
	})

	t.Run("case=should not handle API flows", func(t *testing.T) {
		loginFlow := &login.Flow{Type: flow.TypeAPI}
		assert.Nil(t, s.PopulateLoginMethodFirstFactor(nil, loginFlow))
	})
}

func TestCompleteLogin(t *testing.T) {
	t.Parallel()
	fix := newLoginFixture(t)

	t.Run("case=should return webauthn.js", func(t *testing.T) {
		res, err := fix.publicTS.Client().Get(fix.publicTS.URL + "/.well-known/ory/webauthn.js")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, "text/javascript; charset=UTF-8", res.Header.Get("Content-Type"))
	})

	t.Run("flow=passwordless", func(t *testing.T) {
		t.Run("case=passkey button exists", func(t *testing.T) {
			client := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, client, fix.publicTS, false, true, false, false)
			testhelpers.SnapshotTExcept(t, f.Ui.Nodes, []string{
				"0.attributes.value",
				"2.attributes.nonce",
				"2.attributes.src",
				"5.attributes.value",
			})
		})

		t.Run("case=passkey shows error if user tries to sign in but no such user exists", func(t *testing.T) {
			payload := func(v url.Values) {
				v.Set("method", "passkey")
				v.Set(node.PasskeyLogin, string(loginPasswordlessResponse))
			}

			check := func(t *testing.T, shouldRedirect bool, body string, res *http.Response) {
				fix.checkURL(t, shouldRedirect, res)
				assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
				assert.Equal(t, text.NewErrorValidationSuchNoWebAuthnUser().Text, gjson.Get(body, "ui.messages.0.text").String(), "%s", body)
			}

			t.Run("type=browser", func(t *testing.T) {
				body, res := fix.loginViaBrowser(t, false, payload, testhelpers.NewClientWithCookies(t))
				check(t, true, body, res)
			})

			t.Run("type=spa", func(t *testing.T) {
				body, res := fix.loginViaBrowser(t, true, payload, testhelpers.NewClientWithCookies(t))
				check(t, false, body, res)
			})
		})

		t.Run("case=should fail if passkey login is invalid", func(t *testing.T) {
			payload := func(v url.Values) {
				v.Set("method", "passkey")
				v.Set("passkey_login", "invalid passkey data")
			}

			check := func(t *testing.T, shouldRedirect bool, body string, res *http.Response) {
				fix.checkURL(t, shouldRedirect, res)
				assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
				assert.Equal(t, "Unable to parse WebAuthn response.", gjson.Get(body, "ui.messages.0.text").String(), "%s", body)
			}

			t.Run("type=browser", func(t *testing.T) {
				body, res := fix.loginViaBrowser(t, false, payload, testhelpers.NewClientWithCookies(t))
				check(t, true, body, res)
			})

			t.Run("type=spa", func(t *testing.T) {
				body, res := fix.loginViaBrowser(t, true, payload, testhelpers.NewClientWithCookies(t))
				check(t, false, body, res)
			})
		})

		t.Run("case=should fail if passkey login is empty", func(t *testing.T) {
			payload := func(v url.Values) {
				v.Set("method", "passkey")
			}

			t.Run("type=browser", func(t *testing.T) {
				_, res := fix.loginViaBrowser(t, false, payload, testhelpers.NewClientWithCookies(t))
				fix.checkURL(t, true, res)
			})

			t.Run("type=spa", func(t *testing.T) {
				body, res := fix.loginViaBrowser(t, true, payload, testhelpers.NewClientWithCookies(t))
				fix.checkURL(t, false, res)
				assert.Equal(t, "browser_location_change_required", gjson.Get(body, "error.id").String(), "%s", body)
			})
		})

		t.Run("case=fails with invalid internal state", func(t *testing.T) {
			run := func(t *testing.T, spa bool) {
				fix.conf.MustSet(fix.ctx, config.ViperKeySessionWhoAmIAAL, "aal1")
				// We load our identity which we will use to replay the webauth session
				fix.createIdentityWithPasskey(t, identity.Credentials{
					Config:  loginPasswordlessCredentials,
					Version: 1,
				})

				browserClient := testhelpers.NewClientWithCookies(t)
				body, _, _ := fix.submitWebAuthnLoginWithClient(t, spa, []byte("invalid context"), browserClient, func(values url.Values) {
					values.Set(node.PasskeyLogin, string(loginPasswordlessResponse))
				}, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel1))

				if spa {
					assert.Equal(
						t,
						"Expected WebAuthN in internal context to be an object but got: unexpected end of JSON input",
						gjson.Get(body, "error.reason").String(),
						"%s", body,
					)
				} else {
					assert.Equal(
						t,
						"Expected WebAuthN in internal context to be an object but got: unexpected end of JSON input",
						gjson.Get(body, "reason").String(),
						"%s", body,
					)
				}
			}

			t.Run("type=browser", func(t *testing.T) {
				run(t, false)
			})

			t.Run("type=spa", func(t *testing.T) {
				run(t, true)
			})
		})

		t.Run("case=succeeds with passwordless login", func(t *testing.T) {
			run := func(t *testing.T, spa bool) {
				fix.conf.MustSet(fix.ctx, config.ViperKeySessionWhoAmIAAL, "aal1")
				// We load our identity which we will use to replay the webauth session
				id := fix.createIdentityWithPasskey(t, identity.Credentials{
					Config:  loginPasswordlessCredentials,
					Version: 1,
				})

				browserClient := testhelpers.NewClientWithCookies(t)
				body, res, f := fix.submitWebAuthnLoginWithClient(t, spa, loginPasswordlessContext, browserClient, func(values url.Values) {
					values.Set(node.PasskeyLogin, string(loginPasswordlessResponse))
				}, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel1))

				prefix := ""
				if spa {
					assert.Contains(t, res.Request.URL.String(), fix.publicTS.URL+login.RouteSubmitFlow)
					prefix = "session."
				} else {
					assert.Contains(t, res.Request.URL.String(), fix.redirTS.URL)
				}

				assert.True(t, gjson.Get(body, prefix+"active").Bool(), "%s", body)
				assert.EqualValues(t, identity.AuthenticatorAssuranceLevel1, gjson.Get(body, prefix+"authenticator_assurance_level").String(), "%s", body)
				assert.EqualValues(t, identity.CredentialsTypePasskey, gjson.Get(body, prefix+"authentication_methods.#(method==passkey).method").String(), "%s", body)
				assert.EqualValues(t, id.ID.String(), gjson.Get(body, prefix+"identity.id").String(), "%s", body)

				actualFlow, err := fix.reg.LoginFlowPersister().GetLoginFlow(context.Background(), uuid.FromStringOrNil(f.Id))
				require.NoError(t, err)

				assert.Empty(t, gjson.GetBytes(actualFlow.InternalContext, flow.PrefixInternalContextKey(identity.CredentialsTypePasskey, passkey.InternalContextKeySessionData)))
				if spa {
					assert.EqualValues(t, flow.ContinueWithActionRedirectBrowserToString, gjson.Get(body, "continue_with.0.action").String(), "%s", body)
					assert.Contains(t, gjson.Get(body, "continue_with.0.redirect_browser_to").String(), fix.conf.SelfServiceBrowserDefaultReturnTo(ctx).String(), "%s", body)
				} else {
					assert.Empty(t, gjson.Get(body, "continue_with").Array(), "%s", body)
				}
			}

			// We test here that login works even if the identity schema contains
			// { webauthn: { identifier: true } } instead of
			// { passkey: { display_name: true } }
			t.Run("webauthn_identifier", func(t *testing.T) {
				testhelpers.SetDefaultIdentitySchema(fix.conf, "file://./stub/login_webauthn.schema.json")
				t.Run("type=browser", func(t *testing.T) { run(t, false) })
				t.Run("type=spa", func(t *testing.T) { run(t, true) })
			})
			t.Run("passkey_display_name", func(t *testing.T) {
				testhelpers.SetDefaultIdentitySchema(fix.conf, "file://./stub/login.schema.json")
				t.Run("type=browser", func(t *testing.T) { run(t, false) })
				t.Run("type=spa", func(t *testing.T) { run(t, true) })
			})
		})
	})

	t.Run("flow=refresh", func(t *testing.T) {
		fix := newLoginFixture(t)
		fix.conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, "aal1")
		loginFixtureSuccessEmail := gjson.GetBytes(loginSuccessIdentity, "traits.email").String()

		run := func(t *testing.T, ctx context.Context, id *identity.Identity, context, response []byte, isSPA bool, expectedAAL identity.AuthenticatorAssuranceLevel) {
			body, res, f := fix.submitWebAuthnLogin(t, ctx, isSPA, id, context, func(values url.Values) {
				values.Set("identifier", loginFixtureSuccessEmail)
				values.Set(node.PasskeyLogin, string(response))
			}, testhelpers.InitFlowWithRefresh())
			snapshotx.SnapshotTExcept(t, f.Ui.Nodes, []string{
				"0.attributes.value",
				"2.attributes.nonce",
				"2.attributes.src",
				"5.attributes.value",
			})
			nodes, err := json.Marshal(f.Ui.Nodes)
			require.NoError(t, err)
			assert.Equal(t, loginFixtureSuccessEmail, gjson.GetBytes(nodes, "#(attributes.name==identifier).attributes.value").String(), "%s", nodes)

			prefix := ""
			if isSPA {
				assert.Contains(t, res.Request.URL.String(), fix.publicTS.URL+login.RouteSubmitFlow, "%s", body)
				prefix = "session."
			} else {
				assert.Contains(t, res.Request.URL.String(), fix.redirTS.URL, "%s", body)
			}

			assert.True(t, gjson.Get(body, prefix+"active").Bool(), "%s", body)

			assert.EqualValues(t, expectedAAL, gjson.Get(body, prefix+"authenticator_assurance_level").String(), "%s", body)
			assert.EqualValues(t, identity.CredentialsTypePasskey, gjson.Get(body, prefix+"authentication_methods.#(method==passkey).method").String(), "%s", body)
			assert.Len(t, gjson.Get(body, prefix+"authentication_methods").Array(), 2, "%s", body)
			assert.EqualValues(t, id.ID.String(), gjson.Get(body, prefix+"identity.id").String(), "%s", body)
		}

		expectedAAL := identity.AuthenticatorAssuranceLevel1

		for _, tc := range []struct {
			creds    identity.Credentials
			response []byte
			context  []byte
			descript string
		}{
			{
				creds: identity.Credentials{
					Config:  loginPasswordlessCredentials,
					Version: 1,
				},
				context:  loginPasswordlessContext,
				response: loginPasswordlessResponse,
				descript: "passwordless credentials",
			},
		} {
			t.Run("case=refresh "+tc.descript, func(t *testing.T) {
				id := fix.createIdentityWithPasskey(t, tc.creds)

				for _, f := range []string{
					"browser",
					"spa",
				} {
					t.Run(f, func(t *testing.T) {
						run(t, ctx, id, tc.context, tc.response, f == "spa", expectedAAL)
					})
				}
			})
		}
	})
}

func createIdentity(t *testing.T, ctx context.Context, reg driver.Registry, id uuid.UUID) *identity.Identity {
	i := identity.NewIdentity("default")
	i.SetCredentials(identity.CredentialsTypePasskey, identity.Credentials{
		Identifiers: []string{id.String()},
		Config:      loginPasswordlessCredentials,
		Type:        identity.CredentialsTypePasskey,
		Version:     1,
	})

	require.NoError(t, reg.IdentityManager().Create(ctx, i))
	return i
}

func TestFormHydration(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)

	ctx = configtesthelpers.WithConfigValue(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePasskey)+".enabled", true)
	ctx = configtesthelpers.WithConfigValue(
		ctx,
		config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePasskey)+".config",
		map[string]interface{}{
			"rp": map[string]interface{}{
				"display_name": "foo",
				"id":           "localhost",
				"origins":      []string{"http://localhost"},
			},
		},
	)
	ctx = testhelpers.WithDefaultIdentitySchema(ctx, "file://stub/login.schema.json")

	s, err := reg.AllLoginStrategies().Strategy(identity.CredentialsTypePasskey)
	require.NoError(t, err)
	fh, ok := s.(login.FormHydrator)
	require.True(t, ok)

	toSnapshot := func(t *testing.T, f *login.Flow) {
		t.Helper()
		// The CSRF token has a unique value that messes with the snapshot - ignore it.
		f.UI.Nodes.ResetNodes("csrf_token")
		f.UI.Nodes.ResetNodes("passkey_challenge")
		snapshotx.SnapshotT(t, f.UI.Nodes, snapshotx.ExceptNestedKeys("nonce", "src"))
	}

	newFlow := func(ctx context.Context, t *testing.T) (*http.Request, *login.Flow) {
		r := httptest.NewRequest("GET", "/self-service/login/browser", nil)
		r = r.WithContext(ctx)
		t.Helper()
		f, err := login.NewFlow(conf, time.Minute, "csrf_token", r, flow.TypeBrowser)
		f.UI.Nodes = make(node.Nodes, 0)
		require.NoError(t, err)
		return r, f
	}

	t.Run("method=PopulateLoginMethodSecondFactor", func(t *testing.T) {
		r, f := newFlow(ctx, t)
		f.RequestedAAL = identity.AuthenticatorAssuranceLevel2
		require.NoError(t, fh.PopulateLoginMethodSecondFactor(r, f))
		toSnapshot(t, f)
	})

	t.Run("method=PopulateLoginMethodFirstFactor", func(t *testing.T) {
		r, f := newFlow(ctx, t)
		require.NoError(t, fh.PopulateLoginMethodFirstFactor(r, f))
		toSnapshot(t, f)
	})

	t.Run("method=PopulateLoginMethodFirstFactorRefresh", func(t *testing.T) {
		r, f := newFlow(ctx, t)

		id := createIdentity(t, ctx, reg, x.NewUUID())
		r.Header = testhelpers.NewHTTPClientWithIdentitySessionToken(t, ctx, reg, id).Transport.(*testhelpers.TransportWithHeader).GetHeader()
		f.Refresh = true

		require.NoError(t, fh.PopulateLoginMethodFirstFactorRefresh(r, f))
		toSnapshot(t, f)
	})

	t.Run("method=PopulateLoginMethodSecondFactorRefresh", func(t *testing.T) {
		r, f := newFlow(ctx, t)
		require.NoError(t, fh.PopulateLoginMethodSecondFactorRefresh(r, f))
		toSnapshot(t, f)
	})

	t.Run("method=PopulateLoginMethodIdentifierFirstCredentials", func(t *testing.T) {
		t.Run("case=no options", func(t *testing.T) {
			t.Run("case=account enumeration mitigation disabled", func(t *testing.T) {
				ctx := configtesthelpers.WithConfigValue(ctx, config.ViperKeySecurityAccountEnumerationMitigate, false)
				r, f := newFlow(ctx, t)
				require.ErrorIs(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f), idfirst.ErrNoCredentialsFound)
				toSnapshot(t, f)
			})

			t.Run("case=account enumeration mitigation enabled", func(t *testing.T) {
				ctx := configtesthelpers.WithConfigValue(ctx, config.ViperKeySecurityAccountEnumerationMitigate, true)
				r, f := newFlow(ctx, t)
				require.ErrorIs(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f), idfirst.ErrNoCredentialsFound)
				toSnapshot(t, f)
			})
		})

		t.Run("case=WithIdentifier", func(t *testing.T) {
			t.Run("case=account enumeration mitigation disabled", func(t *testing.T) {
				ctx := configtesthelpers.WithConfigValue(ctx, config.ViperKeySecurityAccountEnumerationMitigate, false)
				r, f := newFlow(ctx, t)
				require.ErrorIs(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f, login.WithIdentifier("foo@bar.com")), idfirst.ErrNoCredentialsFound)
				toSnapshot(t, f)
			})

			t.Run("case=account enumeration mitigation enabled", func(t *testing.T) {
				ctx := configtesthelpers.WithConfigValue(ctx, config.ViperKeySecurityAccountEnumerationMitigate, true)
				r, f := newFlow(ctx, t)
				require.ErrorIs(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f, login.WithIdentifier("foo@bar.com")), idfirst.ErrNoCredentialsFound)
				toSnapshot(t, f)
			})
		})

		t.Run("case=WithIdentityHint", func(t *testing.T) {
			t.Run("case=account enumeration mitigation enabled", func(t *testing.T) {
				ctx := configtesthelpers.WithConfigValue(ctx, config.ViperKeySecurityAccountEnumerationMitigate, true)

				id := identity.NewIdentity("test-provider")
				r, f := newFlow(ctx, t)
				require.ErrorIs(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f, login.WithIdentityHint(id)), idfirst.ErrNoCredentialsFound)
				toSnapshot(t, f)
			})

			t.Run("case=account enumeration mitigation disabled", func(t *testing.T) {
				ctx := configtesthelpers.WithConfigValue(ctx, config.ViperKeySecurityAccountEnumerationMitigate, false)

				t.Run("case=identity has passkey", func(t *testing.T) {
					identifier := x.NewUUID()
					id := createIdentity(t, ctx, reg, identifier)

					r, f := newFlow(ctx, t)
					require.NoError(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f, login.WithIdentityHint(id)))
					toSnapshot(t, f)
				})

				t.Run("case=identity does not have a passkey", func(t *testing.T) {
					id := identity.NewIdentity("default")
					r, f := newFlow(ctx, t)
					require.ErrorIs(t, fh.PopulateLoginMethodIdentifierFirstCredentials(r, f, login.WithIdentityHint(id)), idfirst.ErrNoCredentialsFound)
					toSnapshot(t, f)
				})
			})
		})
	})

	t.Run("method=PopulateLoginMethodIdentifierFirstIdentification", func(t *testing.T) {
		r, f := newFlow(ctx, t)
		require.NoError(t, fh.PopulateLoginMethodIdentifierFirstIdentification(r, f))
		toSnapshot(t, f)
	})
}
