package login_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

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
	for _, strategy := range []string{
		identity.CredentialsTypePassword.String(),
		identity.CredentialsTypeOIDC.String(),
	} {
		t.Run("strategy="+strategy, func(t *testing.T) {
			conf, reg := internal.NewFastRegistryWithMocks(t)
			conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/login.schema.json")
			conf.MustSet(config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh/")

			newServer := func(t *testing.T, ft flow.Type, useIdentity *identity.Identity) *httptest.Server {
				router := httprouter.New()

				router.GET("/login/pre", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					f, err := login.NewFlow(conf, time.Minute, "", r, ft)
					require.NoError(t, err)
					if testhelpers.SelfServiceHookLoginErrorHandler(t, w, r, reg.LoginHookExecutor().PreLoginHook(w, r, f)) {
						_, _ = w.Write([]byte("ok"))
					}
				})

				router.GET("/login/post", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					a, err := login.NewFlow(conf, time.Minute, "", r, ft)
					require.NoError(t, err)
					a.Active = identity.CredentialsType(strategy)
					a.RequestURL = x.RequestURL(r).String()
					sess := session.NewInactiveSession()
					sess.CompletedLoginFor(identity.CredentialsTypePassword)
					if useIdentity == nil {
						useIdentity = testhelpers.SelfServiceHookCreateFakeIdentity(t, reg)
					}

					testhelpers.SelfServiceHookLoginErrorHandler(t, w, r,
						reg.LoginHookExecutor().PostLoginHook(w, r, a, useIdentity, sess))
				})

				ts := httptest.NewServer(router)
				t.Cleanup(ts.Close)
				conf.MustSet(config.ViperKeyPublicBaseURL, ts.URL)
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
					viperSetPost(t, conf, strategy, []config.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})

					res, _ := makeRequestPost(t, newServer(t, flow.TypeBrowser, nil), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/", res.Request.URL.String())
				})

				t.Run("case=fail if hooks fail", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					viperSetPost(t, conf, strategy, []config.SelfServiceHook{{Name: "err", Config: []byte(`{"ExecuteLoginPostHook": "abort"}`)}})

					res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser, nil), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.Equal(t, "", body)
				})

				t.Run("case=use return_to value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					conf.MustSet(config.ViperKeyURLsWhitelistedReturnToDomains, []string{"https://www.ory.sh/"})

					res, _ := makeRequestPost(t, newServer(t, flow.TypeBrowser, nil), false, url.Values{"return_to": {"https://www.ory.sh/kratos/"}})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/kratos/", res.Request.URL.String())
				})

				t.Run("case=use nested config value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					conf.MustSet(config.ViperKeySelfServiceLoginAfter+"."+config.DefaultBrowserReturnURL, "https://www.ory.sh/kratos")

					res, _ := makeRequestPost(t, newServer(t, flow.TypeBrowser, nil), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/kratos/", res.Request.URL.String())
				})

				t.Run("case=use nested config value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					testhelpers.SelfServiceHookLoginSetDefaultRedirectTo(t, conf, "https://www.ory.sh/not-kratos")
					testhelpers.SelfServiceHookLoginSetDefaultRedirectToStrategy(t, conf, strategy, "https://www.ory.sh/kratos")

					res, _ := makeRequestPost(t, newServer(t, flow.TypeBrowser, nil), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/kratos/", res.Request.URL.String())
				})

				t.Run("case=pass if hooks pass", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					viperSetPost(t, conf, strategy, []config.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})

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

				t.Run("case=pass without hooks for browser flow with application/json", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser, nil), true, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.NotEmpty(t, gjson.Get(body, "session.identity.id").String())
					assert.Empty(t, gjson.Get(body, "session.token").String())
					assert.Empty(t, gjson.Get(body, "session_token").String())
				})

				t.Run("case=work normally if AAL is satisfied", func(t *testing.T) {
					conf.MustSet(config.ViperKeySessionWhoAmIAAL, "aal1")
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
					conf.MustSet(config.ViperKeySessionWhoAmIAAL, "highest_available")
					_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
					t.Cleanup(func() {
						conf.MustSet(config.ViperKeySessionWhoAmIAAL, "aal1")
					})
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					useIdentity := &identity.Identity{Credentials: map[identity.CredentialsType]identity.Credentials{
						identity.CredentialsTypePassword: {Type: identity.CredentialsTypePassword},
						identity.CredentialsTypeTOTP:     {Type: identity.CredentialsTypeTOTP},
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
		})
	}
}
