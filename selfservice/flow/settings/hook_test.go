package settings_test

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/tidwall/gjson"

	"github.com/gobuffalo/httptest"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/hook"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

func TestSettingsExecutor(t *testing.T) {
	for _, strategy := range []string{
		identity.CredentialsTypePassword.String(),
		settings.StrategyProfile,
	} {
		t.Run("strategy="+strategy, func(t *testing.T) {

			conf, reg := internal.NewFastRegistryWithMocks(t)
			conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/identity.schema.json")
			conf.MustSet(config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh/")

			reg.WithHooks(map[string]func(config.SelfServiceHook) interface{}{
				"err": func(c config.SelfServiceHook) interface{} {
					return &hook.Error{Config: c.Config}
				},
			})

			newServer := func(t *testing.T, ft flow.Type) *httptest.Server {
				router := httprouter.New()
				handleErr := testhelpers.SelfServiceHookSettingsErrorHandler
				router.GET("/settings/post", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					i := testhelpers.SelfServiceHookCreateFakeIdentity(t, reg)
					sess, _ := session.NewActiveSession(i, conf, time.Now().UTC())

					a := settings.NewFlow(conf, time.Minute, r, sess.Identity, ft)
					a.RequestURL = x.RequestURL(r).String()
					require.NoError(t, reg.SettingsFlowPersister().CreateSettingsFlow(r.Context(), a))
					_ = handleErr(t, w, r, reg.SettingsHookExecutor().
						PostSettingsHook(w, r, strategy, &settings.UpdateContext{Flow: a, Session: sess}, i))
				})
				ts := httptest.NewServer(router)
				t.Cleanup(ts.Close)
				conf.MustSet(config.ViperKeyPublicBaseURL, ts.URL)
				return ts
			}

			makeRequestPost := testhelpers.SelfServiceMakeSettingsPostHookRequest
			viperSetPost := func(strategy string, c []config.SelfServiceHook) {
				conf.MustSet(config.HookStrategyKey(config.ViperKeySelfServiceSettingsAfter, strategy), c)
			}

			uiTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
			t.Cleanup(uiTS.Close)
			uiURL := uiTS.URL + "/user/settings"
			conf.MustSet(config.ViperKeySelfServiceSettingsURL, uiURL)

			t.Run("method=PostSettingsHook", func(t *testing.T) {

				t.Run("case=pass without hooks", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					res, _ := makeRequestPost(t, newServer(t, flow.TypeBrowser), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.Contains(t, res.Request.URL.String(), uiURL)
				})

				t.Run("case=pass if hooks pass", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					viperSetPost(strategy, []config.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})
					res, _ := makeRequestPost(t, newServer(t, flow.TypeBrowser), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.Contains(t, res.Request.URL.String(), uiURL)
				})

				t.Run("case=fail if hooks fail", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					viperSetPost(strategy, []config.SelfServiceHook{{Name: "err", Config: []byte(`{"ExecuteSettingsPrePersistHook": "abort"}`)}})
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
					testhelpers.SelfServiceHookSettingsSetDefaultRedirectTo(t, conf, "https://www.ory.sh")

					res, _ := makeRequestPost(t, newServer(t, flow.TypeBrowser), false, url.Values{"return_to": {"https://www.ory.sh/kratos/"}})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/kratos/", res.Request.URL.String())
				})

				t.Run("case=use nested config value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					testhelpers.SelfServiceHookSettingsSetDefaultRedirectTo(t, conf, "https://www.ory.sh/kratos")

					res, _ := makeRequestPost(t, newServer(t, flow.TypeBrowser), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/kratos/", res.Request.URL.String())
				})

				t.Run("case=use nested config value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					testhelpers.SelfServiceHookSettingsSetDefaultRedirectTo(t, conf, "https://www.ory.sh/not-kratos")
					testhelpers.SelfServiceHookSettingsSetDefaultRedirectToStrategy(t, conf, strategy, "https://www.ory.sh/kratos")

					res, _ := makeRequestPost(t, newServer(t, flow.TypeBrowser), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/kratos/", res.Request.URL.String())
				})

				t.Run("case=pass if hooks pass", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					viperSetPost(strategy, []config.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})
					res, _ := makeRequestPost(t, newServer(t, flow.TypeBrowser), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.Contains(t, res.Request.URL.String(), uiURL)
				})

				t.Run("case=send a json response for API clients", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					viperSetPost(strategy, nil)
					res, body := makeRequestPost(t, newServer(t, flow.TypeAPI), true, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.NotEmpty(t, gjson.Get(body, "identity.id"))
				})

				t.Run("case=pass without hooks for browser flow with application/json", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					res, body := makeRequestPost(t, newServer(t, flow.TypeBrowser), true, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.NotEmpty(t, gjson.Get(body, "identity.id"))
				})
			})
		})
	}
}
