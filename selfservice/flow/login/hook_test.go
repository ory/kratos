// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/hydra"
	"github.com/ory/kratos/session"

	"github.com/gobuffalo/httptest"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/x"
)

func TestLoginExecutor(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	for _, strategy := range []identity.CredentialsType{
		identity.CredentialsTypePassword,
		identity.CredentialsTypeOIDC,
		identity.CredentialsTypeTOTP,
		identity.CredentialsTypeWebAuthn,
		identity.CredentialsTypeLookup,
	} {
		strategy := strategy

		t.Run("strategy="+strategy.String(), func(t *testing.T) {
			t.Parallel()

			conf, reg := internal.NewFastRegistryWithMocks(t)
			reg.WithHydra(hydra.NewFake())
			testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/login.schema.json")
			conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh/")

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

					res, _ := makeRequestPost(t, newServer(t, flow.TypeBrowser, nil), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/", res.Request.URL.String())
				})

				t.Run("case=pass if hooks pass", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					viperSetPost(t, conf, strategy.String(), []config.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})

					res, _ := makeRequestPost(t, newServer(t, flow.TypeBrowser, nil), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/", res.Request.URL.String())
				})

				t.Run("case=fail if hooks fail", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					viperSetPost(t, conf, strategy.String(), []config.SelfServiceHook{{Name: "err", Config: []byte(`{"ExecuteLoginPostHook": "abort"}`)}})

					res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser, nil), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.Equal(t, "", body)
				})

				t.Run("case=use return_to value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{"https://www.ory.sh/"})

					res, _ := makeRequestPost(t, newServer(t, flow.TypeBrowser, nil), false, url.Values{"return_to": {"https://www.ory.sh/kratos/"}})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/kratos/", res.Request.URL.String())
				})

				t.Run("case=use nested config value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					conf.MustSet(ctx, config.ViperKeySelfServiceLoginAfter+"."+config.DefaultBrowserReturnURL, "https://www.ory.sh/kratos")

					res, _ := makeRequestPost(t, newServer(t, flow.TypeBrowser, nil), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/kratos/", res.Request.URL.String())
				})

				t.Run("case=use nested config value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					testhelpers.SelfServiceHookLoginSetDefaultRedirectTo(t, conf, "https://www.ory.sh/not-kratos")
					testhelpers.SelfServiceHookLoginSetDefaultRedirectToStrategy(t, conf, strategy.String(), "https://www.ory.sh/kratos")

					res, _ := makeRequestPost(t, newServer(t, flow.TypeBrowser, nil), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/kratos/", res.Request.URL.String())
				})

				t.Run("case=pass if hooks pass", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					viperSetPost(t, conf, strategy.String(), []config.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})

					res, _ := makeRequestPost(t, newServer(t, flow.TypeBrowser, nil), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/", res.Request.URL.String())
				})

				t.Run("case=send a json response for API clients", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					res, body := makeRequestPost(t, newServer(t, flow.TypeAPI, nil), true, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.NotEmpty(t, gjson.Get(body, "session.identity.id").String())
				})

				t.Run("suite=handle login challenge with browser and application/json", func(t *testing.T) {
					t.Run("case=includes the return_to address for a valid challenge", func(t *testing.T) {
						t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

						withOAuthChallenge := func(f *login.Flow) {
							f.OAuth2LoginChallenge = hydra.FakeValidLoginChallenge
						}
						res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser, nil, withOAuthChallenge), true, url.Values{})
						assert.EqualValues(t, http.StatusUnprocessableEntity, res.StatusCode)
						assert.Equal(t, hydra.FakePostLoginURL, gjson.Get(body, "redirect_browser_to").String(), "%s", body)
					})

					t.Run("case=returns an error for an invalid challenge", func(t *testing.T) {
						t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

						withOAuthChallenge := func(f *login.Flow) {
							f.OAuth2LoginChallenge = hydra.FakeInvalidLoginChallenge
						}
						res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser, nil, withOAuthChallenge), true, url.Values{})
						assert.EqualValues(t, http.StatusInternalServerError, res.StatusCode)
						assert.Equal(t, hydra.ErrFakeAcceptLoginRequestFailed.Error(), body, "%s", body)
					})
				})

				t.Run("case=pass without hooks for browser flow with application/json", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser, nil), true, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.NotEmpty(t, gjson.Get(body, "session.identity.id").String())
					assert.Empty(t, gjson.Get(body, "session.token").String())
					assert.Empty(t, gjson.Get(body, "session_token").String())
				})

				t.Run("case=work normally if AAL is satisfied", func(t *testing.T) {
					conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, "aal1")
					_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					useIdentity := &identity.Identity{Credentials: map[identity.CredentialsType]identity.Credentials{
						identity.CredentialsTypePassword: {Type: identity.CredentialsTypePassword},
						identity.CredentialsTypeTOTP:     {Type: identity.CredentialsTypeTOTP},
					}}
					require.NoError(t, reg.Persister().CreateIdentity(context.Background(), useIdentity))

					t.Run("browser client", func(t *testing.T) {
						res, _ := makeRequestPost(t, newServer(t, flow.TypeBrowser, useIdentity), false, url.Values{})
						assert.EqualValues(t, http.StatusOK, res.StatusCode)
						assert.EqualValues(t, "https://www.ory.sh/", res.Request.URL.String())
					})

					t.Run("api client returns the token but not the identity", func(t *testing.T) {
						res, body := makeRequestPost(t, newServer(t, flow.TypeAPI, useIdentity), true, url.Values{})
						assert.EqualValues(t, http.StatusOK, res.StatusCode)
						assert.NotEmpty(t, gjson.Get(body, "session.identity").String())
						assert.NotEmpty(t, gjson.Get(body, "session_token").String())
					})

					t.Run("browser JSON client returns the token but not the identity", func(t *testing.T) {
						res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser, useIdentity), true, url.Values{})
						assert.EqualValues(t, http.StatusOK, res.StatusCode)
						assert.NotEmpty(t, gjson.Get(body, "session.id").String())
						assert.NotEmpty(t, gjson.Get(body, "session.identity").String())
						assert.Empty(t, gjson.Get(body, "session_token").String())
					})
				})

				t.Run("case=redirect to login if AAL is too low", func(t *testing.T) {
					conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, "highest_available")
					_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
					t.Cleanup(func() {
						conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, "aal1")
					})
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					useIdentity := &identity.Identity{Credentials: map[identity.CredentialsType]identity.Credentials{
						identity.CredentialsTypePassword: {Type: identity.CredentialsTypePassword, Config: []byte(`{"hashed_password": "$argon2id$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRNw"}`), Identifiers: []string{testhelpers.RandomEmail()}},
						identity.CredentialsTypeWebAuthn: {Type: identity.CredentialsTypeWebAuthn, Config: []byte(`{"credentials":[{"is_passwordless":false}]}`), Identifiers: []string{testhelpers.RandomEmail()}},
					}}
					require.NoError(t, reg.Persister().CreateIdentity(context.Background(), useIdentity))

					t.Run("browser client", func(t *testing.T) {
						res, _ := makeRequestPost(t, newServer(t, flow.TypeBrowser, useIdentity), false, url.Values{})
						assert.EqualValues(t, http.StatusNotFound, res.StatusCode)
						assert.Contains(t, res.Request.URL.String(), "/self-service/login/browser?aal=aal2")
					})

					t.Run("api client returns the token but not the identity", func(t *testing.T) {
						res, body := makeRequestPost(t, newServer(t, flow.TypeAPI, useIdentity), true, url.Values{})
						assert.EqualValues(t, http.StatusOK, res.StatusCode)
						assert.Empty(t, gjson.Get(body, "session.identity").String())
						assert.NotEmpty(t, gjson.Get(body, "session_token").String())
					})

					t.Run("browser JSON client returns the token but not the identity", func(t *testing.T) {
						res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser, useIdentity), true, url.Values{})
						assert.EqualValues(t, http.StatusOK, res.StatusCode)
						assert.NotEmpty(t, gjson.Get(body, "session.id").String())
						assert.Empty(t, gjson.Get(body, "session.identity").String())
						assert.Empty(t, gjson.Get(body, "session_token").String())
					})
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

			t.Run("requiresAAL2 should return true if there's an error", func(t *testing.T) {
				requiresAAL2, err := login.RequiresAAL2ForTest(*reg.LoginHookExecutor(), &http.Request{}, &session.Session{})
				require.NotNil(t, err)
				require.True(t, requiresAAL2)
			})
		})
	}
}
