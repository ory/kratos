package settings_test

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gobuffalo/httptest"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
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
			viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/identity.schema.json")
			viper.Set(configuration.ViperKeyURLsDefaultReturnTo, "https://www.ory.sh/")

			reg.WithHooks(map[string]func(configuration.SelfServiceHook) interface{}{
				"err": func(c configuration.SelfServiceHook) interface{} {
					return &hook.Error{Config: c.Config}
				},
			})

			newServer := func(t *testing.T) *httptest.Server {
				router := httprouter.New()
				handleErr := testhelpers.SelfServiceHookSettingsErrorHandler
				router.GET("/settings/post", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					i := testhelpers.SelfServiceHookCreateFakeIdentity(t, reg)
					sess := session.NewSession(i, conf, time.Now().UTC())

					a := settings.NewRequest(time.Minute, r, sess)
					a.RequestURL = x.RequestURL(r).String()
					require.NoError(t, reg.SettingsRequestPersister().CreateSettingsRequest(r.Context(), a))
					_ = handleErr(t, w, r, reg.SettingsHookExecutor().
						PostSettingsHook(w, r, strategy, &settings.UpdateContext{Request: a, Session: sess}, i))
				})
				ts := httptest.NewServer(router)
				t.Cleanup(ts.Close)
				viper.Set(configuration.ViperKeyURLsSelfPublic, ts.URL)
				return ts
			}

			makeRequestPost := testhelpers.SelfServiceMakeSettingsPostHookRequest
			viperSetPost := func(strategy string, c []configuration.SelfServiceHook) {
				viper.Set(configuration.HookStrategyKey(configuration.ViperKeySelfServiceSettingsAfter, strategy), c)
			}

			uiTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
			t.Cleanup(uiTS.Close)
			uiURL := uiTS.URL + "/user/settings"
			viper.Set(configuration.ViperKeyURLsSettings, uiURL)

			t.Run("method=PostSettingsHook", func(t *testing.T) {
				t.Run("case=pass without hooks", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset)

					res, _ := makeRequestPost(t, newServer(t), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.Contains(t, res.Request.URL.String(), uiURL)
				})

				t.Run("case=pass if hooks pass", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset)

					viperSetPost(strategy, []configuration.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})
					res, _ := makeRequestPost(t, newServer(t), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.Contains(t, res.Request.URL.String(), uiURL)
				})

				t.Run("case=fail if hooks fail", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset)

					viperSetPost(strategy, []configuration.SelfServiceHook{{Name: "err", Config: []byte(`{"ExecuteSettingsPrePersistHook": "abort"}`)}})
					res, body := makeRequestPost(t, newServer(t), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.Equal(t, "", body)
				})

				t.Run("case=prevent return_to value because domain not whitelisted", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset)

					res, _ := makeRequestPost(t, newServer(t), false, url.Values{"return_to": {"https://www.ory.sh/kratos/"}})
					assert.EqualValues(t, http.StatusInternalServerError, res.StatusCode)
				})

				t.Run("case=use return_to value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset)
					viper.Set(configuration.ViperKeyURLsWhitelistedReturnToDomains, []string{"https://www.ory.sh/"})
					testhelpers.SelfServiceHookSettingsSetDefaultRedirectTo("https://www.ory.sh")

					res, _ := makeRequestPost(t, newServer(t), false, url.Values{"return_to": {"https://www.ory.sh/kratos/"}})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/kratos/", res.Request.URL.String())
				})

				t.Run("case=use nested config value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset)
					testhelpers.SelfServiceHookSettingsSetDefaultRedirectTo("https://www.ory.sh/kratos")

					res, _ := makeRequestPost(t, newServer(t), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/kratos/", res.Request.URL.String())
				})

				t.Run("case=use nested config value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset)
					testhelpers.SelfServiceHookSettingsSetDefaultRedirectTo("https://www.ory.sh/not-kratos")
					testhelpers.SelfServiceHookSettingsSetDefaultRedirectToStrategy(strategy, "https://www.ory.sh/kratos")

					res, _ := makeRequestPost(t, newServer(t), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/kratos/", res.Request.URL.String())
				})

				t.Run("case=pass if hooks pass", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset)
					viperSetPost(strategy, []configuration.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})
					res, _ := makeRequestPost(t, newServer(t), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.Contains(t, res.Request.URL.String(), uiURL)
				})

				t.Run("case=send a json response for API clients", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset)
					viperSetPost(strategy, nil)
					res, body := makeRequestPost(t, newServer(t), true, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.NotEmpty(t, gjson.Get(body, "identity.id"))
				})
			})
		})
	}
}
