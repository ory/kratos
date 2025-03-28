// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	confighelpers "github.com/ory/kratos/driver/config/testhelpers"
	"github.com/ory/kratos/hydra"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
)

func TestLoginExecutor(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	for _, strategy := range identity.AllCredentialTypes {
		t.Run("strategy="+strategy.String(), func(t *testing.T) {
			t.Parallel()

			conf, reg := internal.NewFastRegistryWithMocks(t)
			reg.WithHydra(hydra.NewFake())
			testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/login.schema.json")
			conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, returnToServer.URL)
			_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)

			newServer := func(t *testing.T, ft flow.Type, useIdentity *identity.Identity, flowCallback ...func(*login.Flow)) *httptest.Server {
				router := httprouter.New()

				router.GET("/login/pre", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					loginFlow, err := login.NewFlow(conf, time.Minute, "", r, ft)
					require.NoError(t, err)
					if testhelpers.SelfServiceHookLoginErrorHandler(t, w, r, reg.LoginHookExecutor().PreLoginHook(w, r, loginFlow)) {
						_, _ = w.Write([]byte("ok"))
					}
				})

				router.GET("/login/post", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					loginFlow, err := login.NewFlow(conf, time.Minute, "", r, ft)
					require.NoError(t, err)
					loginFlow.Active = strategy
					loginFlow.RequestURL = x.RequestURL(r).String()
					for _, cb := range flowCallback {
						cb(loginFlow)
					}

					sess := session.NewInactiveSession()
					sess.CompletedLoginFor(identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
					if useIdentity == nil {
						useIdentity = testhelpers.SelfServiceHookCreateFakeIdentity(t, reg)
					}

					testhelpers.SelfServiceHookLoginErrorHandler(t, w, r,
						reg.LoginHookExecutor().PostLoginHook(w, r, strategy.ToUiNodeGroup(), loginFlow, useIdentity, sess, ""))
				})

				router.GET("/login/post2fa", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					loginFlow, err := login.NewFlow(conf, time.Minute, "", r, ft)
					require.NoError(t, err)
					loginFlow.Active = strategy
					loginFlow.RequestURL = x.RequestURL(r).String()
					for _, cb := range flowCallback {
						cb(loginFlow)
					}

					sess := session.NewInactiveSession()
					sess.CompletedLoginFor(identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
					sess.CompletedLoginFor(identity.CredentialsTypeTOTP, identity.AuthenticatorAssuranceLevel2)
					if useIdentity == nil {
						useIdentity = testhelpers.SelfServiceHookCreateFakeIdentity(t, reg)
					}

					testhelpers.SelfServiceHookLoginErrorHandler(t, w, r,
						reg.LoginHookExecutor().PostLoginHook(w, r, strategy.ToUiNodeGroup(), loginFlow, useIdentity, sess, ""))
				})

				ts := httptest.NewServer(router)
				t.Cleanup(ts.Close)
				conf.MustSet(ctx, config.ViperKeyPublicBaseURL, ts.URL)
				return ts
			}

			makeRequestPost := testhelpers.SelfServiceMakeLoginPostHookRequest
			viperSetPost := testhelpers.SelfServiceHookLoginViperSetPost

			t.Run("method=PostLoginHook", func(t *testing.T) {
				t.Run("case=pass without hooks", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser, nil), false, url.Values{})
					require.EqualValuesf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.EqualValues(t, returnToServer.URL, res.Request.URL.String())
				})

				t.Run("case=pass without hooks if client is ajax", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					ts := newServer(t, flow.TypeBrowser, nil)
					res, body := makeRequestPost(t, ts, true, url.Values{})
					require.Equalf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.Contains(t, res.Request.URL.String(), ts.URL)
					assert.JSONEq(t, fmt.Sprintf(`[{"action":"redirect_browser_to","redirect_browser_to":"%s"}]`, returnToServer.URL), gjson.Get(body, "continue_with").Raw)
				})

				t.Run("case=pass if hooks pass", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					viperSetPost(t, conf, strategy.String(), []config.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})

					res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser, nil), false, url.Values{})
					require.Equalf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.Equal(t, returnToServer.URL, res.Request.URL.String())
				})

				t.Run("case=fail if hooks fail", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					viperSetPost(t, conf, strategy.String(), []config.SelfServiceHook{{Name: "err", Config: []byte(`{"ExecuteLoginPostHook": "abort"}`)}})

					ts := newServer(t, flow.TypeBrowser, nil)
					res, body := makeRequestPost(t, ts, false, url.Values{})
					require.Equalf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.Contains(t, res.Request.URL.String(), ts.URL)
					assert.Empty(t, body)
				})

				t.Run("case=use return_to value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{returnToServer.URL})

					res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser, nil), false, url.Values{"return_to": {returnToServer.URL + "/kratos"}})
					require.Equalf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.Equal(t, returnToServer.URL+"/kratos", res.Request.URL.String())
				})

				t.Run("case=use nested config value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					conf.MustSet(ctx, config.ViperKeySelfServiceLoginAfter+"."+config.DefaultBrowserReturnURL, returnToServer.URL+"/kratos")

					res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser, nil), false, url.Values{})
					require.EqualValuesf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.EqualValues(t, returnToServer.URL+"/kratos", res.Request.URL.String())
				})

				t.Run("case=use nested config value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					testhelpers.SelfServiceHookLoginSetDefaultRedirectTo(t, conf, returnToServer.URL+"/not-kratos")
					testhelpers.SelfServiceHookLoginSetDefaultRedirectToStrategy(t, conf, strategy.String(), returnToServer.URL+"/kratos")

					res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser, nil), false, url.Values{})
					require.EqualValuesf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.EqualValues(t, returnToServer.URL+"/kratos", res.Request.URL.String())
				})

				t.Run("case=pass if hooks pass", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					viperSetPost(t, conf, strategy.String(), []config.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})

					res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser, nil), false, url.Values{})
					require.EqualValuesf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.EqualValues(t, returnToServer.URL, res.Request.URL.String())
				})

				t.Run("case=send a json response for API clients", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					res, body := makeRequestPost(t, newServer(t, flow.TypeAPI, nil), true, url.Values{})
					require.EqualValuesf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.NotEmpty(t, gjson.Get(body, "session.identity.id").String())
				})

				t.Run("suite=handle login challenge with browser and application/json", func(t *testing.T) {
					t.Run("case=includes the return_to address for a valid challenge", func(t *testing.T) {
						t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

						withOAuthChallenge := func(f *login.Flow) {
							f.OAuth2LoginChallenge = hydra.FakeValidLoginChallenge
						}
						res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser, nil, withOAuthChallenge), true, url.Values{})
						require.EqualValuesf(t, http.StatusUnprocessableEntity, res.StatusCode, "%s", body)
						assert.Equal(t, hydra.FakePostLoginURL, gjson.Get(body, "redirect_browser_to").String(), "%s", body)
					})

					t.Run("case=returns an error for an invalid challenge", func(t *testing.T) {
						t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

						withOAuthChallenge := func(f *login.Flow) {
							f.OAuth2LoginChallenge = hydra.FakeInvalidLoginChallenge
						}
						res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser, nil, withOAuthChallenge), true, url.Values{})
						require.EqualValuesf(t, http.StatusInternalServerError, res.StatusCode, "%s", body)
						assert.Equal(t, hydra.ErrFakeAcceptLoginRequestFailed.Error(), body, "%s", body)
					})
				})

				t.Run("case=pass without hooks for browser flow with application/json", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser, nil), true, url.Values{})
					require.EqualValuesf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.NotEmpty(t, gjson.Get(body, "session.identity.id").String())
					assert.Empty(t, gjson.Get(body, "session.token").String())
					assert.Empty(t, gjson.Get(body, "session_token").String())
				})

				t.Run("case=work normally if AAL is satisfied", func(t *testing.T) {
					conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, "aal1")
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					useIdentity := &identity.Identity{Credentials: map[identity.CredentialsType]identity.Credentials{
						identity.CredentialsTypePassword: {Type: identity.CredentialsTypePassword},
						identity.CredentialsTypeTOTP:     {Type: identity.CredentialsTypeTOTP},
					}}
					require.NoError(t, reg.Persister().CreateIdentity(context.Background(), useIdentity))

					t.Run("browser client", func(t *testing.T) {
						res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser, useIdentity), false, url.Values{})
						require.EqualValuesf(t, http.StatusOK, res.StatusCode, "%s", body)
						assert.EqualValues(t, returnToServer.URL, res.Request.URL.String())
					})

					t.Run("api client returns the session with identity and the token", func(t *testing.T) {
						res, body := makeRequestPost(t, newServer(t, flow.TypeAPI, useIdentity), true, url.Values{})
						require.EqualValuesf(t, http.StatusOK, res.StatusCode, "%s", body)
						assert.NotEmpty(t, gjson.Get(body, "session.identity").String())
						assert.NotEmpty(t, gjson.Get(body, "session_token").String())
					})

					t.Run("browser JSON client returns the session with identity but not the token", func(t *testing.T) {
						res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser, useIdentity), true, url.Values{})
						require.EqualValuesf(t, http.StatusOK, res.StatusCode, "%s", body)
						assert.NotEmpty(t, gjson.Get(body, "session.id").String())
						assert.NotEmpty(t, gjson.Get(body, "session.identity").String())
						assert.Empty(t, gjson.Get(body, "session_token").String())
					})
				})

				t.Run("case=redirect to login if AAL is too low", func(t *testing.T) {
					conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, "highest_available")
					t.Cleanup(func() {
						conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, "aal1")
					})
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					conf.MustSet(ctx, config.ViperKeyOAuth2ProviderURL, "https://hydra")

					useIdentity := &identity.Identity{Credentials: map[identity.CredentialsType]identity.Credentials{
						identity.CredentialsTypePassword: {Type: identity.CredentialsTypePassword, Config: []byte(`{"hashed_password": "$argon2id$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRNw"}`), Identifiers: []string{testhelpers.RandomEmail()}},
						identity.CredentialsTypeWebAuthn: {Type: identity.CredentialsTypeWebAuthn, Config: []byte(`{"credentials":[{"is_passwordless":false}]}`), Identifiers: []string{testhelpers.RandomEmail()}},
					}}
					require.NoError(t, reg.Persister().CreateIdentity(context.Background(), useIdentity))

					t.Run("browser client", func(t *testing.T) {
						res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser, useIdentity), false, url.Values{})
						require.EqualValuesf(t, http.StatusNotFound, res.StatusCode, "%s", body)
						assert.Contains(t, res.Request.URL.String(), "/self-service/login/browser?aal=aal2")
					})

					t.Run("browser client with login challenge", func(t *testing.T) {
						res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser, useIdentity), false, url.Values{
							"login_challenge": []string{hydra.FakeValidLoginChallenge},
						})
						require.EqualValuesf(t, http.StatusNotFound, res.StatusCode, "%s", body)

						assert.Equal(t, res.Request.URL.Path, "/self-service/login/browser")
						assert.Equal(t, res.Request.URL.Query().Get("aal"), "aal2")
						assert.Equal(t, res.Request.URL.Query().Get("login_challenge"), hydra.FakeValidLoginChallenge)
					})

					t.Run("api client returns the token and the session without the identity", func(t *testing.T) {
						res, body := makeRequestPost(t, newServer(t, flow.TypeAPI, useIdentity), true, url.Values{})
						require.EqualValuesf(t, http.StatusOK, res.StatusCode, "%s", body)
						assert.Empty(t, gjson.Get(body, "session.identity").String())
						assert.NotEmpty(t, gjson.Get(body, "session_token").String())
					})

					t.Run("browser JSON client", func(t *testing.T) {
						res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser, useIdentity), true, url.Values{})
						require.EqualValuesf(t, http.StatusUnprocessableEntity, res.StatusCode, "%s", body)
						assert.NotEmpty(t, gjson.Get(body, "redirect_browser_to").String())
						assert.Contains(t, gjson.Get(body, "redirect_browser_to").String(), "/self-service/login/browser?aal=aal2", "%s", body)
					})

					t.Run("browser JSON client with login challenge", func(t *testing.T) {
						res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser, useIdentity), true, url.Values{
							"login_challenge": []string{hydra.FakeValidLoginChallenge},
						})
						require.EqualValuesf(t, http.StatusUnprocessableEntity, res.StatusCode, "%s", body)
						assert.NotEmpty(t, gjson.Get(body, "redirect_browser_to").String())

						redirectBrowserTo, err := url.Parse(gjson.Get(body, "redirect_browser_to").String())
						require.NoError(t, err)

						assert.Equal(t, redirectBrowserTo.Path, "/self-service/login/browser")
						assert.Equal(t, redirectBrowserTo.Query().Get("aal"), "aal2")
						assert.Equal(t, redirectBrowserTo.Query().Get("login_challenge"), hydra.FakeValidLoginChallenge)
					})
				})
			})

			t.Run("case=maybe links credential", func(t *testing.T) {
				t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
				conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, config.HighestAvailableAAL)
				conf.MustSet(ctx, "selfservice.methods.totp.enabled", true)

				email1, email2 := testhelpers.RandomEmail(), testhelpers.RandomEmail()
				passwordOnlyIdentity := &identity.Identity{Credentials: map[identity.CredentialsType]identity.Credentials{
					identity.CredentialsTypePassword: {
						Type:        identity.CredentialsTypePassword,
						Config:      []byte(`{"hashed_password": "$argon2id$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRNw"}`),
						Identifiers: []string{email1},
					},
				}}
				twoFAIdentitiy := &identity.Identity{Credentials: map[identity.CredentialsType]identity.Credentials{
					identity.CredentialsTypePassword: {
						Type:        identity.CredentialsTypePassword,
						Config:      []byte(`{"hashed_password": "$argon2id$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRNw"}`),
						Identifiers: []string{email2},
					},
					identity.CredentialsTypeTOTP: {
						Type:        identity.CredentialsTypeTOTP,
						Config:      []byte(`{"totp_url":"otpauth://totp/test"}`),
						Identifiers: []string{email2},
					},
				}}
				require.NoError(t, reg.Persister().CreateIdentity(ctx, passwordOnlyIdentity))
				require.NoError(t, reg.Persister().CreateIdentity(ctx, twoFAIdentitiy))

				credsOIDCPWOnly, err := identity.NewCredentialsOIDC(
					&identity.CredentialsOIDCEncryptedTokens{IDToken: "id-token", AccessToken: "access-token", RefreshToken: "refresh-token"},
					"my-provider",
					email1,
					"",
				)
				require.NoError(t, err)
				credsOIDC2FA, err := identity.NewCredentialsOIDC(
					&identity.CredentialsOIDCEncryptedTokens{IDToken: "id-token", AccessToken: "access-token", RefreshToken: "refresh-token"},
					"my-provider",
					email2,
					"",
				)
				require.NoError(t, err)

				t.Run("sub-case=does not link after first factor when second factor is available", func(t *testing.T) {
					duplicateCredentialsData := flow.DuplicateCredentialsData{
						CredentialsType:     identity.CredentialsTypeOIDC,
						CredentialsConfig:   credsOIDC2FA.Config,
						DuplicateIdentifier: email2,
					}
					ts := newServer(t, flow.TypeBrowser, twoFAIdentitiy, func(l *login.Flow) {
						require.NoError(t, flow.SetDuplicateCredentials(l, duplicateCredentialsData))
					})
					res, _ := makeRequestPost(t, ts, false, url.Values{})

					assert.Equal(t, reg.Config().SelfServiceFlowLoginUI(ctx).Host, res.Request.URL.Host)
					assert.Equal(t, reg.Config().SelfServiceFlowLoginUI(ctx).Path, res.Request.URL.Path)
					newFlowID := res.Request.URL.Query().Get("flow")
					assert.NotEmpty(t, newFlowID)

					newFlow, err := reg.LoginFlowPersister().GetLoginFlow(ctx, uuid.Must(uuid.FromString(newFlowID)))
					require.NoError(t, err)
					newFlowDuplicateCredentialsData, err := flow.DuplicateCredentials(newFlow)
					require.NoError(t, err)

					// Duplicate credentials data should have been copied over
					assert.Equal(t, duplicateCredentialsData.CredentialsType, newFlowDuplicateCredentialsData.CredentialsType)
					assert.Equal(t, duplicateCredentialsData.DuplicateIdentifier, newFlowDuplicateCredentialsData.DuplicateIdentifier)
					assert.JSONEq(t, string(duplicateCredentialsData.CredentialsConfig), string(newFlowDuplicateCredentialsData.CredentialsConfig))

					// AAL should be AAL2
					assert.Equal(t, identity.AuthenticatorAssuranceLevel2, newFlow.RequestedAAL)

					// TOTP nodes should be present
					found := false
					for _, n := range newFlow.UI.Nodes {
						if n.Group == node.TOTPGroup {
							found = true
							break
						}
					}
					assert.True(t, found, "could not find TOTP nodes in %+v", newFlow.UI.Nodes)

					ident, err := reg.Persister().GetIdentity(ctx, twoFAIdentitiy.ID, identity.ExpandCredentials)
					require.NoError(t, err)
					assert.Len(t, ident.Credentials, 2)
				})

				t.Run("sub-case=links after second factor when second factor is available", func(t *testing.T) {
					ts := newServer(t, flow.TypeBrowser, twoFAIdentitiy, func(l *login.Flow) {
						require.NoError(t, flow.SetDuplicateCredentials(l, flow.DuplicateCredentialsData{
							CredentialsType:     identity.CredentialsTypeOIDC,
							CredentialsConfig:   credsOIDC2FA.Config,
							DuplicateIdentifier: email2,
						}))
					})
					res, body := testhelpers.SelfServiceMakeHookRequest(t, ts, "/login/post2fa", false, url.Values{})
					require.Equalf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.Equalf(t, returnToServer.URL, res.Request.URL.String(), "%s", body)

					ident, err := reg.Persister().GetIdentity(ctx, twoFAIdentitiy.ID, identity.ExpandCredentials)
					require.NoError(t, err)
					assert.Len(t, ident.Credentials, 3)
				})

				t.Run("sub-case=links matching identity", func(t *testing.T) {
					res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser, passwordOnlyIdentity, func(l *login.Flow) {
						require.NoError(t, flow.SetDuplicateCredentials(l, flow.DuplicateCredentialsData{
							CredentialsType:     identity.CredentialsTypeOIDC,
							CredentialsConfig:   credsOIDCPWOnly.Config,
							DuplicateIdentifier: email1,
						}))
					}), false, url.Values{})
					require.Equalf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.Equalf(t, returnToServer.URL, res.Request.URL.String(), "%s", body)

					ident, err := reg.Persister().GetIdentity(ctx, passwordOnlyIdentity.ID, identity.ExpandCredentials)
					require.NoError(t, err)
					assert.Len(t, ident.Credentials, 2)
				})

				t.Run("sub-case=errors on non-matching identity", func(t *testing.T) {
					res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser, passwordOnlyIdentity, func(l *login.Flow) {
						require.NoError(t, flow.SetDuplicateCredentials(l, flow.DuplicateCredentialsData{
							CredentialsType:     identity.CredentialsTypeOIDC,
							CredentialsConfig:   credsOIDCPWOnly.Config,
							DuplicateIdentifier: "wrong@example.com",
						}))
					}), false, url.Values{})
					require.EqualValues(t, http.StatusInternalServerError, res.StatusCode)
					assert.Equal(t, schema.NewLinkedCredentialsDoNotMatch().Error(), body, "%s", body)
				})
			})

			t.Run("type=api", func(t *testing.T) {
				t.Run("method=PreLoginHook", testhelpers.TestSelfServicePreHook(
					config.ViperKeySelfServiceLoginBeforeHooks,
					testhelpers.SelfServiceMakeLoginPreHookRequest,
					func(t *testing.T) *httptest.Server {
						return newServer(t, flow.TypeAPI, nil)
					},
					conf,
				))
			})

			t.Run("type=browser", func(t *testing.T) {
				t.Run("method=PreLoginHook", testhelpers.TestSelfServicePreHook(
					config.ViperKeySelfServiceLoginBeforeHooks,
					testhelpers.SelfServiceMakeLoginPreHookRequest,
					func(t *testing.T) *httptest.Server {
						return newServer(t, flow.TypeBrowser, nil)
					},
					conf,
				))
			})
		})
	}

	t.Run("method=checkAAL", func(t *testing.T) {
		ctx := confighelpers.WithConfigValue(ctx, config.ViperKeyPublicBaseURL, returnToServer.URL)

		conf, reg := internal.NewFastRegistryWithMocks(t)
		testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/login.schema.json")
		conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, returnToServer.URL)

		t.Run("returns no error when sufficient", func(t *testing.T) {
			ctx := confighelpers.WithConfigValue(ctx, config.ViperKeySessionWhoAmIAAL, identity.AuthenticatorAssuranceLevel1)
			assert.NoError(t,
				login.CheckAALForTest(ctx, reg.LoginHookExecutor(), &session.Session{
					AMR: session.AuthenticationMethods{{
						Method: identity.CredentialsTypePassword,
						AAL:    identity.AuthenticatorAssuranceLevel1,
					}},
					AuthenticatorAssuranceLevel: identity.AuthenticatorAssuranceLevel1,
				}, nil),
			)

			ctx = confighelpers.WithConfigValue(ctx, config.ViperKeySessionWhoAmIAAL, config.HighestAvailableAAL)
			assert.NoError(t,
				login.CheckAALForTest(ctx, reg.LoginHookExecutor(), &session.Session{
					AMR: session.AuthenticationMethods{{
						Method: identity.CredentialsTypePassword,
						AAL:    identity.AuthenticatorAssuranceLevel1,
					}, {
						Method: identity.CredentialsTypeLookup,
						AAL:    identity.AuthenticatorAssuranceLevel2,
					}},
					AuthenticatorAssuranceLevel: identity.AuthenticatorAssuranceLevel2,
				}, nil),
			)
		})

		t.Run("copies parameters to redirect URL when AAL is not sufficient", func(t *testing.T) {
			ctx := confighelpers.WithConfigValue(ctx, config.ViperKeySessionWhoAmIAAL, config.HighestAvailableAAL)
			aalErr := new(session.ErrAALNotSatisfied)
			require.ErrorAs(t,
				login.CheckAALForTest(ctx, reg.LoginHookExecutor(), &session.Session{
					AMR: session.AuthenticationMethods{{
						Method: identity.CredentialsTypePassword,
						AAL:    identity.AuthenticatorAssuranceLevel1,
					}},
					AuthenticatorAssuranceLevel: identity.AuthenticatorAssuranceLevel1,
					Identity: &identity.Identity{
						InternalAvailableAAL: identity.NullableAuthenticatorAssuranceLevel{sql.NullString{String: string(identity.AuthenticatorAssuranceLevel2), Valid: true}},
					},
				}, &login.Flow{
					RequestURL: "https://www.ory.sh/?return_to=https://www.ory.sh/kratos&login_challenge=challenge",
				}),
				&aalErr,
			)
			assert.Equal(t, returnToServer.URL+"/self-service/login/browser?aal=aal2&login_challenge=challenge&return_to=https%3A%2F%2Fwww.ory.sh%2Fkratos", aalErr.RedirectTo)
		})
	})
}
