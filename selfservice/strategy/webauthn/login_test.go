package webauthn_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/ory/x/jsonx"

	"github.com/duo-labs/webauthn/protocol"

	kratos "github.com/ory/kratos-client-go"
	"github.com/ory/kratos/text"
	"github.com/ory/x/snapshotx"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/strategy/webauthn"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/assertx"
)

var (
	//go:embed fixtures/login/success/mfa/response.invalid.json
	loginFixtureSuccessResponseInvalid []byte
	//go:embed fixtures/login/success/mfa/identity.json
	loginFixtureSuccessIdentity []byte
	//go:embed fixtures/login/success/mfa/v0/credentials.json
	loginFixtureSuccessV0Credentials []byte
	//go:embed fixtures/login/success/mfa/v0/internal_context.json
	loginFixtureSuccessV0Context []byte
	//go:embed fixtures/login/success/mfa/v0/response.json
	loginFixtureSuccessV0Response []byte
	//go:embed fixtures/login/success/mfa/v1/credentials.json
	loginFixtureSuccessV1Credentials []byte
	//go:embed fixtures/login/success/mfa/v1/internal_context.json
	loginFixtureSuccessV1Context []byte
	//go:embed fixtures/login/success/mfa/v1/response.json
	loginFixtureSuccessV1Response []byte
	//go:embed fixtures/login/success/mfa/v1_handle/credentials.json
	loginFixtureSuccessV1WithHandleCredentials []byte
	//go:embed fixtures/login/success/mfa/v1_handle/internal_context.json
	loginFixtureSuccessV1WithHandleContext []byte
	//go:embed fixtures/login/success/mfa/v1_handle/response.json
	loginFixtureSuccessV1WithHandleResponse []byte
	//go:embed fixtures/login/success/mfa/v1_passwordless/credentials.json
	loginFixtureSuccessV1PasswordlessCredentials []byte
	//go:embed fixtures/login/success/mfa/v1_passwordless/internal_context.json
	loginFixtureSuccessV1PasswordlessContext []byte
	//go:embed fixtures/login/success/mfa/v1_passwordless/response.json
	loginFixtureSuccessV1PasswordlessResponse []byte
)

var loginFixtureSuccessEmail = gjson.GetBytes(loginFixtureSuccessIdentity, "traits.email").String()

func TestCompleteLogin(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypePassword)+".enabled", false)
	enableWebAuthn(conf)

	router := x.NewRouterPublic()
	publicTS, _ := testhelpers.NewKratosServerWithRouters(t, reg, router, x.NewRouterAdmin())

	errTS := testhelpers.NewErrorTestServer(t, reg)
	uiTS := testhelpers.NewLoginUIFlowEchoServer(t, reg)
	redirTS := testhelpers.NewRedirSessionEchoTS(t, reg)

	// Overwrite these two to make it more explicit when tests fail
	conf.MustSet(config.ViperKeySelfServiceErrorUI, errTS.URL+"/error-ts")
	conf.MustSet(config.ViperKeySelfServiceLoginUI, uiTS.URL+"/login-ts")

	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/login.schema.json")
	conf.MustSet(config.ViperKeySecretsDefault, []string{"not-a-secure-session-key"})

	checkURL := func(t *testing.T, shouldRedirect bool, res *http.Response) {
		if shouldRedirect {
			assert.Contains(t, res.Request.URL.String(), uiTS.URL+"/login-ts")
		} else {
			assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
		}
	}

	doAPIFlow := func(t *testing.T, v func(url.Values), apiClient *http.Client, opts ...testhelpers.InitFlowWithOption) (string, *http.Response) {
		f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false, opts...)
		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		v(values)
		payload := testhelpers.EncodeFormAsJSON(t, true, values)
		return testhelpers.LoginMakeRequest(t, true, false, f, apiClient, payload)
	}

	doBrowserFlow := func(t *testing.T, spa bool, v func(url.Values), browserClient *http.Client, opts ...testhelpers.InitFlowWithOption) (string, *http.Response) {
		f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, spa, opts...)
		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		v(values)
		return testhelpers.LoginMakeRequest(t, false, spa, f, browserClient, values.Encode())
	}

	createIdentityWithWebAuthn := func(t *testing.T, c identity.Credentials) *identity.Identity {
		var id identity.Identity
		require.NoError(t, json.Unmarshal(loginFixtureSuccessIdentity, &id))

		id.SetCredentials(identity.CredentialsTypeWebAuthn, identity.Credentials{
			Identifiers: []string{loginFixtureSuccessEmail},
			Config:      c.Config,
			Type:        identity.CredentialsTypeWebAuthn,
			Version:     c.Version,
		})

		// We clean up the identity in case it has been created before
		_ = reg.PrivilegedIdentityPool().DeleteIdentity(context.Background(), id.ID)

		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &id))
		return &id
	}

	submitWebAuthnLoginFlowWithClient := func(t *testing.T, isSPA bool, f *kratos.SelfServiceLoginFlow, contextFixture []byte, client *http.Client, cb func(values url.Values)) (string, *http.Response, *kratos.SelfServiceLoginFlow) {
		// We inject the session to replay
		interim, err := reg.LoginFlowPersister().GetLoginFlow(context.Background(), uuid.FromStringOrNil(f.Id))
		require.NoError(t, err)
		interim.InternalContext = contextFixture
		require.NoError(t, reg.LoginFlowPersister().UpdateLoginFlow(context.Background(), interim))

		values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)
		cb(values)

		// We use the response replay
		body, res := testhelpers.LoginMakeRequest(t, false, isSPA, f, client, values.Encode())
		return body, res, f
	}

	submitWebAuthnLoginWithClient := func(t *testing.T, isSPA bool, id *identity.Identity, contextFixture []byte, client *http.Client, cb func(values url.Values), opts ...testhelpers.InitFlowWithOption) (string, *http.Response, *kratos.SelfServiceLoginFlow) {
		f := testhelpers.InitializeLoginFlowViaBrowser(t, client, publicTS, false, isSPA, opts...)
		return submitWebAuthnLoginFlowWithClient(t, isSPA, f, contextFixture, client, cb)
	}

	submitWebAuthnLogin := func(t *testing.T, isSPA bool, id *identity.Identity, contextFixture []byte, cb func(values url.Values), opts ...testhelpers.InitFlowWithOption) (string, *http.Response, *kratos.SelfServiceLoginFlow) {
		browserClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id)
		return submitWebAuthnLoginWithClient(t, isSPA, id, contextFixture, browserClient, cb, opts...)
	}

	t.Run("flow=refresh", func(t *testing.T) {
		conf.MustSet(config.ViperKeySessionWhoAmIAAL, "aal1")
		t.Cleanup(func() {
			conf.MustSet(config.ViperKeySessionWhoAmIAAL, nil)
		})

		run := func(t *testing.T, id *identity.Identity, context, response []byte, isSPA bool, expectedAAL identity.AuthenticatorAssuranceLevel) {
			body, res, f := submitWebAuthnLogin(t, isSPA, id, context, func(values url.Values) {
				values.Set("identifier", loginFixtureSuccessEmail)
				values.Set(node.WebAuthnLogin, string(response))
			}, testhelpers.InitFlowWithRefresh())
			snapshotx.SnapshotTExcept(t, f.Ui.Nodes, []string{
				"0.attributes.value",
				"2.attributes.onclick",
				"4.attributes.nonce",
				"4.attributes.src",
			})
			nodes, err := json.Marshal(f.Ui.Nodes)
			require.NoError(t, err)
			assert.Equal(t, loginFixtureSuccessEmail, gjson.GetBytes(nodes, "#(attributes.name==identifier).attributes.value").String(), "%s", nodes)

			prefix := ""
			if isSPA {
				assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow, "%s", body)
				prefix = "session."
			} else {
				assert.Contains(t, res.Request.URL.String(), redirTS.URL, "%s", body)
			}

			assert.True(t, gjson.Get(body, prefix+"active").Bool(), "%s", body)

			assert.EqualValues(t, expectedAAL, gjson.Get(body, prefix+"authenticator_assurance_level").String(), "%s", body)
			assert.EqualValues(t, identity.CredentialsTypeWebAuthn, gjson.Get(body, prefix+"authentication_methods.#(method==webauthn).method").String(), "%s", body)
			assert.Len(t, gjson.Get(body, prefix+"authentication_methods").Array(), 2, "%s", body)
			assert.EqualValues(t, id.ID.String(), gjson.Get(body, prefix+"identity.id").String(), "%s", body)
		}

		t.Run("case=passwordless", func(t *testing.T) {
			for _, e := range []bool{
				true,
				false,
			} {
				conf.MustSet(config.ViperKeyWebAuthnPasswordless, e)
				expectedAAL := identity.AuthenticatorAssuranceLevel1
				if !e {
					// If passwordless is disabled, using WebAuthn means that we have a second factor enabled.
					// Thus, AAL2 :)
					expectedAAL = identity.AuthenticatorAssuranceLevel2
				}

				for _, tc := range []struct {
					creds    identity.Credentials
					response []byte
					context  []byte
					descript string
				}{
					{
						creds: identity.Credentials{
							Config:  loginFixtureSuccessV0Credentials,
							Version: 0,
						},
						context:  loginFixtureSuccessV0Context,
						response: loginFixtureSuccessV0Response,
						descript: "mfa v0 credentials",
					},
					{
						creds: identity.Credentials{
							Config:  loginFixtureSuccessV1Credentials,
							Version: 1,
						},
						context:  loginFixtureSuccessV1Context,
						response: loginFixtureSuccessV1Response,
						descript: "mfa v1 credentials",
					},
					{
						creds: identity.Credentials{
							Config:  loginFixtureSuccessV1PasswordlessCredentials,
							Version: 1,
						},
						context:  loginFixtureSuccessV1PasswordlessContext,
						response: loginFixtureSuccessV1PasswordlessResponse,
						descript: "passwordless credentials",
					},
				} {
					t.Run(fmt.Sprintf("case=mfa v0 credentials/passwordless enabled=%v", e), func(t *testing.T) {
						id := createIdentityWithWebAuthn(t, tc.creds)

						for _, f := range []string{
							"browser",
							"spa",
						} {
							t.Run(f, func(t *testing.T) {
								run(t, id, tc.context, tc.response, f == "spa", expectedAAL)
							})
						}
					})
				}
			}
		})

		t.Run("case=no webauth credentials", func(t *testing.T) {
			for _, e := range []bool{true, false} {
				conf.MustSet(config.ViperKeyWebAuthnPasswordless, e)
				t.Run(fmt.Sprintf("passwordless=%v", e), func(t *testing.T) {
					for _, f := range []string{"browser", "spa"} {
						t.Run(f, func(t *testing.T) {
							id := identity.NewIdentity("")
							client := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id)

							f := testhelpers.InitializeLoginFlowViaBrowser(t, client, publicTS, true, f == "spa")
							snapshotx.SnapshotTExcept(t, f.Ui.Nodes, []string{
								"0.attributes.value",
							})
							nodes, err := json.Marshal(f.Ui.Nodes)
							require.NoError(t, err)
							assert.False(t, gjson.GetBytes(nodes, "#(attributes.name==identifier).attributes.value").Bool(), "%s", nodes)
							assert.False(t, gjson.GetBytes(nodes, "#(attributes.name=="+node.WebAuthnLoginTrigger+").attributes.value").Bool(), "%s", nodes)
						})
					}
				})
			}
		})
	})

	t.Run("flow=passwordless", func(t *testing.T) {
		conf.MustSet(config.ViperKeyWebAuthnPasswordless, true)
		t.Cleanup(func() {
			conf.MustSet(config.ViperKeyWebAuthnPasswordless, false)
		})

		t.Run("case=webauthn button exists", func(t *testing.T) {
			client := testhelpers.NewClientWithCookies(t)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, client, publicTS, false, true)
			testhelpers.SnapshotTExcept(t, f.Ui.Nodes, []string{"0.attributes.value"})
		})

		t.Run("case=webauthn shows error if user tries to sign in but no such user exists", func(t *testing.T) {
			payload := func(v url.Values) {
				v.Set("method", identity.CredentialsTypeWebAuthn.String())
				v.Set("identifier", "doesnotexist")
			}

			check := func(t *testing.T, shouldRedirect bool, body string, res *http.Response) {
				checkURL(t, shouldRedirect, res)
				assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
				assert.Equal(t, text.NewErrorValidationSuchNoWebAuthnUser().Text, gjson.Get(body, "ui.messages.0.text").String(), "%s", body)
			}

			t.Run("type=browser", func(t *testing.T) {
				body, res := doBrowserFlow(t, false, payload, testhelpers.NewClientWithCookies(t))
				check(t, true, body, res)
			})

			t.Run("type=spa", func(t *testing.T) {
				body, res := doBrowserFlow(t, true, payload, testhelpers.NewClientWithCookies(t))
				check(t, false, body, res)
			})
		})

		t.Run("case=webauthn shows error if user tries to sign in but user has no webauth credentials set up", func(t *testing.T) {
			id, subject := createIdentityAndReturnIdentifier(t, reg, nil)
			id.DeleteCredentialsType(identity.CredentialsTypeWebAuthn)
			require.NoError(t, reg.IdentityManager().Update(ctx, id, identity.ManagerAllowWriteProtectedTraits))

			payload := func(v url.Values) {
				v.Set("method", identity.CredentialsTypeWebAuthn.String())
				v.Set("identifier", subject)
			}

			check := func(t *testing.T, shouldRedirect bool, body string, res *http.Response) {
				checkURL(t, shouldRedirect, res)
				assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
				assert.Equal(t, text.NewErrorValidationSuchNoWebAuthnUser().Text, gjson.Get(body, "ui.messages.0.text").String(), "%s", body)
			}

			t.Run("type=browser", func(t *testing.T) {
				body, res := doBrowserFlow(t, false, payload, testhelpers.NewClientWithCookies(t))
				check(t, true, body, res)
			})

			t.Run("type=spa", func(t *testing.T) {
				body, res := doBrowserFlow(t, true, payload, testhelpers.NewClientWithCookies(t))
				check(t, false, body, res)
			})
		})

		t.Run("case=webauthn MFA credentials can not be used for passwordless login", func(t *testing.T) {
			_, subject := createIdentityAndReturnIdentifier(t, reg, []byte(`{"credentials":[{"id":"Zm9vZm9v","is_passwordless":false}]}`))

			payload := func(v url.Values) {
				v.Set("method", identity.CredentialsTypeWebAuthn.String())
				v.Set("identifier", subject)
			}

			check := func(t *testing.T, shouldRedirect bool, body string, res *http.Response) {
				checkURL(t, shouldRedirect, res)
				assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
				assert.Equal(t, text.NewErrorValidationSuchNoWebAuthnUser().Text, gjson.Get(body, "ui.messages.0.text").String(), "%s", body)
			}

			t.Run("type=browser", func(t *testing.T) {
				body, res := doBrowserFlow(t, false, payload, testhelpers.NewClientWithCookies(t))
				check(t, true, body, res)
			})

			t.Run("type=spa", func(t *testing.T) {
				body, res := doBrowserFlow(t, true, payload, testhelpers.NewClientWithCookies(t))
				check(t, false, body, res)
			})
		})

		t.Run("case=should fail if webauthn login is invalid", func(t *testing.T) {
			_, subject := createIdentityAndReturnIdentifier(t, reg, []byte(`{"credentials":[{"id":"Zm9vZm9v","display_name":"foo","is_passwordless":true}]}`))

			doBrowserFlow := func(t *testing.T, spa bool, browserClient *http.Client, opts ...testhelpers.InitFlowWithOption) {
				f := testhelpers.InitializeLoginFlowViaBrowser(t, browserClient, publicTS, false, spa, opts...)
				values := testhelpers.SDKFormFieldsToURLValues(f.Ui.Nodes)

				values.Set("method", identity.CredentialsTypeWebAuthn.String())
				values.Set("identifier", subject)
				body, res := testhelpers.LoginMakeRequest(t, false, spa, f, browserClient, values.Encode())
				if spa {
					assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
					redir := gjson.Get(body, "redirect_browser_to").String()
					assert.NotEmpty(t, redir)

					res, err := browserClient.Get(redir)
					require.NoError(t, err)

					defer res.Body.Close()
					raw, err := io.ReadAll(res.Body)
					require.NoError(t, err)
					body = string(raw)
				} else {
					checkURL(t, !spa, res)
				}

				assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
				snapshotx.SnapshotTExceptMatchingKeys(t, json.RawMessage(body), []string{"value", "src", "nonce", "action", "request_url", "issued_at", "expires_at", "created_at", "updated_at", "id", "onclick"})
				assert.Equal(t, text.NewInfoLoginWebAuthnPasswordless().Text, gjson.Get(body, "ui.messages.0.text").String(), "%s", body)

				values.Set(node.WebAuthnLogin, string(loginFixtureSuccessResponseInvalid))
				values.Set("identifier", subject)
				body, res = testhelpers.LoginMakeRequest(t, false, spa, f, browserClient, values.Encode())
				checkURL(t, !spa, res)
				assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
				assert.Equal(t, "The provided authentication code is invalid, please try again.", gjson.Get(body, "ui.messages.0.text").String(), "%s", body)
			}

			t.Run("type=browser", func(t *testing.T) {
				doBrowserFlow(t, false, testhelpers.NewClientWithCookies(t))
			})

			t.Run("type=spa", func(t *testing.T) {
				doBrowserFlow(t, true, testhelpers.NewClientWithCookies(t))
			})
		})

		t.Run("case=succeeds with passwordless login", func(t *testing.T) {
			run := func(t *testing.T, spa bool) {
				// We load our identity which we will use to replay the webauth session
				id := createIdentityWithWebAuthn(t, identity.Credentials{
					Config:  loginFixtureSuccessV1PasswordlessCredentials,
					Version: 1,
				})

				browserClient := testhelpers.NewClientWithCookies(t)
				body, res, f := submitWebAuthnLoginWithClient(t, spa, id, loginFixtureSuccessV1PasswordlessContext, browserClient, func(values url.Values) {
					values.Set("identifier", loginFixtureSuccessEmail)
					values.Set(node.WebAuthnLogin, string(loginFixtureSuccessV1PasswordlessResponse))
				}, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel1))

				prefix := ""
				if spa {
					assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
					prefix = "session."
				} else {
					assert.Contains(t, res.Request.URL.String(), redirTS.URL)
				}

				assert.True(t, gjson.Get(body, prefix+"active").Bool(), "%s", body)
				assert.EqualValues(t, identity.AuthenticatorAssuranceLevel1, gjson.Get(body, prefix+"authenticator_assurance_level").String(), "%s", body)
				assert.EqualValues(t, identity.CredentialsTypeWebAuthn, gjson.Get(body, prefix+"authentication_methods.#(method==webauthn).method").String(), "%s", body)
				assert.EqualValues(t, id.ID.String(), gjson.Get(body, prefix+"identity.id").String(), "%s", body)

				actualFlow, err := reg.LoginFlowPersister().GetLoginFlow(context.Background(), uuid.FromStringOrNil(f.Id))
				require.NoError(t, err)
				assert.Empty(t, gjson.GetBytes(actualFlow.InternalContext, flow.PrefixInternalContextKey(identity.CredentialsTypeWebAuthn, webauthn.InternalContextKeySessionData)))
			}

			t.Run("type=browser", func(t *testing.T) {
				run(t, false)
			})

			t.Run("type=spa", func(t *testing.T) {
				run(t, true)
			})
		})
	})

	t.Run("flow=mfa", func(t *testing.T) {
		t.Run("case=webauthn payload is set when identity has webauthn", func(t *testing.T) {
			id := createIdentity(t, reg)

			apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, apiClient, publicTS, false, true, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))
			assert.Equal(t, gjson.GetBytes(id.Traits, "subject").String(), f.Ui.Nodes[1].Attributes.UiNodeInputAttributes.Value, jsonx.TestMarshalJSONString(t, f.Ui))
			testhelpers.SnapshotTExcept(t, f.Ui.Nodes, []string{
				"0.attributes.value",
				"1.attributes.value",
				"2.attributes.onclick",
				"2.attributes.onload",
				"4.attributes.src",
				"4.attributes.nonce",
			})
			ensureReplacement(t, "2", f.Ui, "allowCredentials")
		})

		t.Run("case=webauthn payload is not set when identity has no webauthn", func(t *testing.T) {
			id := createIdentityWithoutWebAuthn(t, reg)
			apiClient := testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id)
			f := testhelpers.InitializeLoginFlowViaBrowser(t, apiClient, publicTS, false, true, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))

			testhelpers.SnapshotTExcept(t, f.Ui.Nodes, []string{
				"0.attributes.value",
			})
		})

		t.Run("case=webauthn payload is not set for API clients", func(t *testing.T) {
			id := createIdentity(t, reg)

			apiClient := testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id)
			f := testhelpers.InitializeLoginFlowViaAPI(t, apiClient, publicTS, false, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))
			assertx.EqualAsJSON(t, nil, f.Ui.Nodes)
		})

		doAPIFlowSignedIn := func(t *testing.T, v func(url.Values), id *identity.Identity) (string, *http.Response) {
			return doAPIFlow(t, v, testhelpers.NewHTTPClientWithIdentitySessionToken(t, reg, id), testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))
		}

		doBrowserFlowSignIn := func(t *testing.T, spa bool, v func(url.Values), id *identity.Identity) (string, *http.Response) {
			return doBrowserFlow(t, spa, v, testhelpers.NewHTTPClientWithIdentitySessionCookie(t, reg, id), testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))
		}

		t.Run("case=should refuse to execute api flow", func(t *testing.T) {
			id := createIdentity(t, reg)
			payload := func(v url.Values) {
				v.Set(node.WebAuthnLogin, "{}")
			}

			body, res := doAPIFlowSignedIn(t, payload, id)
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
			assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
			assert.Equal(t, "Could not find a strategy to log you in with. Did you fill out the form correctly?", gjson.Get(body, "ui.messages.0.text").String(), "%s", body)
		})

		t.Run("case=should fail if webauthn login is invalid", func(t *testing.T) {
			id, sub := createIdentityAndReturnIdentifier(t, reg, nil)
			payload := func(v url.Values) {
				v.Set("identifier", sub)
				v.Set(node.WebAuthnLogin, string(loginFixtureSuccessResponseInvalid))
			}

			check := func(t *testing.T, shouldRedirect bool, body string, res *http.Response) {
				checkURL(t, shouldRedirect, res)
				assert.NotEmpty(t, gjson.Get(body, "id").String(), "%s", body)
				assert.Equal(t, "The provided authentication code is invalid, please try again.", gjson.Get(body, "ui.messages.0.text").String(), "%s", body)
			}

			t.Run("type=browser", func(t *testing.T) {
				body, res := doBrowserFlowSignIn(t, false, payload, id)
				check(t, true, body, res)
			})

			t.Run("type=spa", func(t *testing.T) {
				body, res := doBrowserFlowSignIn(t, true, payload, id)
				check(t, false, body, res)
			})
		})

		t.Run("case=can not use security key for passwordless in mfa flow", func(t *testing.T) {
			id := createIdentityWithWebAuthn(t, identity.Credentials{
				Config:  loginFixtureSuccessV1PasswordlessCredentials,
				Version: 1,
			})

			body, res, _ := submitWebAuthnLogin(t, true, id, loginFixtureSuccessV1PasswordlessContext, func(values url.Values) {
				values.Set("identifier", loginFixtureSuccessEmail)
				values.Set(node.WebAuthnLogin, string(loginFixtureSuccessV1PasswordlessResponse))
			}, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
			snapshotx.SnapshotTExcept(t, json.RawMessage(gjson.Get(body, "ui.messages").Raw), []string{})
		})

		t.Run("case=login with a security key using", func(t *testing.T) {
			idd := uuid.FromStringOrNil("44fc22c9-abae-4c3e-a56b-37c7b38d973e")
			out, err := json.Marshal(webauthn.CredentialsConfig{UserHandle: idd[:]})
			require.NoError(t, err)
			t.Logf("json: %s", out)
			out, err = json.Marshal(protocol.AuthenticatorAssertionResponse{UserHandle: idd[:]})
			require.NoError(t, err)
			t.Logf("wa: %s", out)

			for _, tc := range []struct {
				d  string
				v  int
				cf []byte
				sf []byte
				ix []byte
			}{
				{d: "v0 without userhandle", v: 0, cf: loginFixtureSuccessV0Credentials, sf: loginFixtureSuccessV0Response, ix: loginFixtureSuccessV0Context},
				{d: "v1 without userhandle", v: 1, cf: loginFixtureSuccessV1Credentials, sf: loginFixtureSuccessV1Response, ix: loginFixtureSuccessV1Context},
				{d: "v1 with differing userhandle", v: 1, cf: loginFixtureSuccessV1WithHandleCredentials, sf: loginFixtureSuccessV1WithHandleResponse, ix: loginFixtureSuccessV1WithHandleContext},
			} {
				t.Run(tc.d, func(t *testing.T) {
					run := func(t *testing.T, spa bool) {
						// We load our identity which we will use to replay the webauth session
						id := createIdentityWithWebAuthn(t, identity.Credentials{
							Config:  tc.cf,
							Version: tc.v,
						})

						body, res, f := submitWebAuthnLogin(t, spa, id, tc.ix, func(values url.Values) {
							values.Set("identifier", loginFixtureSuccessEmail)
							values.Set(node.WebAuthnLogin, string(tc.sf))
						}, testhelpers.InitFlowWithAAL(identity.AuthenticatorAssuranceLevel2))

						prefix := ""
						if spa {
							assert.Contains(t, res.Request.URL.String(), publicTS.URL+login.RouteSubmitFlow)
							prefix = "session."
						} else {
							assert.Contains(t, res.Request.URL.String(), redirTS.URL)
						}

						assert.True(t, gjson.Get(body, prefix+"active").Bool(), "%s", body)
						assert.EqualValues(t, identity.AuthenticatorAssuranceLevel2, gjson.Get(body, prefix+"authenticator_assurance_level").String(), "%s", body)
						assert.EqualValues(t, identity.CredentialsTypeWebAuthn, gjson.Get(body, prefix+"authentication_methods.#(method==webauthn).method").String(), "%s", body)
						assert.EqualValues(t, id.ID.String(), gjson.Get(body, prefix+"identity.id").String(), "%s", body)

						actualFlow, err := reg.LoginFlowPersister().GetLoginFlow(context.Background(), uuid.FromStringOrNil(f.Id))
						require.NoError(t, err)
						assert.Empty(t, gjson.GetBytes(actualFlow.InternalContext, flow.PrefixInternalContextKey(identity.CredentialsTypeWebAuthn, webauthn.InternalContextKeySessionData)))
					}

					t.Run("type=browser", func(t *testing.T) {
						run(t, false)
					})

					t.Run("type=spa", func(t *testing.T) {
						run(t, true)
					})
				})
			}
		})
	})
}
