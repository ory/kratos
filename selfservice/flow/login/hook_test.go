package login_test

import (
	"net/http"
	"net/url"
	"testing"
	"time"

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

			newServer := func(t *testing.T, ft flow.Type) *httptest.Server {
				router := httprouter.New()

				router.GET("/login/pre", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					if testhelpers.SelfServiceHookLoginErrorHandler(t, w, r, reg.LoginHookExecutor().PreLoginHook(w, r, login.NewFlow(time.Minute, "", r, ft))) {
						_, _ = w.Write([]byte("ok"))
					}
				})

				router.GET("/login/post", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					a := login.NewFlow(time.Minute, "", r, ft)
					a.RequestURL = x.RequestURL(r).String()
					testhelpers.SelfServiceHookLoginErrorHandler(t, w, r,
						reg.LoginHookExecutor().PostLoginHook(w, r, identity.CredentialsType(strategy), a, testhelpers.SelfServiceHookCreateFakeIdentity(t, reg)))
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

					res, _ := makeRequestPost(t, newServer(t, flow.TypeBrowser), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/", res.Request.URL.String())
				})

				t.Run("case=pass if hooks pass", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					viperSetPost(t, conf, strategy, []config.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})

					res, _ := makeRequestPost(t, newServer(t, flow.TypeBrowser), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/", res.Request.URL.String())
				})

				t.Run("case=fail if hooks fail", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					viperSetPost(t, conf, strategy, []config.SelfServiceHook{{Name: "err", Config: []byte(`{"ExecuteLoginPostHook": "abort"}`)}})

					res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.Equal(t, "", body)
				})

				t.Run("case=prevent return_to value because domain not whitelisted", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					res, _ := makeRequestPost(t, newServer(t, flow.TypeBrowser), false, url.Values{"return_to": {"https://www.ory.sh/kratos/"}})
					assert.EqualValues(t, http.StatusInternalServerError, res.StatusCode)
				})

				t.Run("case=use return_to value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					conf.MustSet(config.ViperKeyURLsWhitelistedReturnToDomains, []string{"https://www.ory.sh/"})

					res, _ := makeRequestPost(t, newServer(t, flow.TypeBrowser), false, url.Values{"return_to": {"https://www.ory.sh/kratos/"}})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/kratos/", res.Request.URL.String())
				})

				t.Run("case=use nested config value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					conf.MustSet(config.ViperKeySelfServiceLoginAfter+"."+config.DefaultBrowserReturnURL, "https://www.ory.sh/kratos")

					res, _ := makeRequestPost(t, newServer(t, flow.TypeBrowser), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/kratos/", res.Request.URL.String())
				})

				t.Run("case=use nested config value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					testhelpers.SelfServiceHookLoginSetDefaultRedirectTo(t, conf, "https://www.ory.sh/not-kratos")
					testhelpers.SelfServiceHookLoginSetDefaultRedirectToStrategy(t, conf, strategy, "https://www.ory.sh/kratos")

					res, _ := makeRequestPost(t, newServer(t, flow.TypeBrowser), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/kratos/", res.Request.URL.String())
				})

				t.Run("case=pass if hooks pass", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					viperSetPost(t, conf, strategy, []config.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})

					res, _ := makeRequestPost(t, newServer(t, flow.TypeBrowser), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/", res.Request.URL.String())
				})

				t.Run("case=send a json response for API clients", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					res, body := makeRequestPost(t, newServer(t, flow.TypeAPI), true, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.NotEmpty(t, gjson.Get(body, "session.identity.id"))
				})
			})

			t.Run("type=api", func(t *testing.T) {
				t.Run("method=PreLoginHook", testhelpers.TestSelfServicePreHook(
					config.ViperKeySelfServiceLoginBeforeHooks,
					testhelpers.SelfServiceMakeLoginPreHookRequest,
					func(t *testing.T) *httptest.Server {
						return newServer(t, flow.TypeAPI)
					},
					conf,
				))
			})

			t.Run("type=browser", func(t *testing.T) {
				t.Run("method=PreLoginHook", testhelpers.TestSelfServicePreHook(
					config.ViperKeySelfServiceLoginBeforeHooks,
					testhelpers.SelfServiceMakeLoginPreHookRequest,
					func(t *testing.T) *httptest.Server {
						return newServer(t, flow.TypeBrowser)
					},
					conf,
				))
			})
		})
	}
}
