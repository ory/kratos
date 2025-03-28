// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package settings_test

import (
	"context"
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
	"github.com/ory/kratos/x"
)

func TestSettingsExecutor(t *testing.T) {
	ctx := context.Background()
	for _, strategy := range []string{
		identity.CredentialsTypePassword.String(),
		settings.StrategyProfile,
	} {
		t.Run("strategy="+strategy, func(t *testing.T) {

			conf, reg := internal.NewFastRegistryWithMocks(t)
			testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/identity.schema.json")
			conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, returnToServer.URL)

			reg.WithHooks(map[string]func(config.SelfServiceHook) interface{}{
				"err": func(c config.SelfServiceHook) interface{} {
					return &hook.Error{Config: c.Config}
				},
			})

			newServer := func(t *testing.T, i *identity.Identity, ft flow.Type) *httptest.Server {
				t.Helper()
				router := httprouter.New()
				handleErr := testhelpers.SelfServiceHookSettingsErrorHandler
				router.GET("/settings/pre", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					if i == nil {
						i = testhelpers.SelfServiceHookCreateFakeIdentity(t, reg)
					}
					sess, _ := testhelpers.NewActiveSession(r, reg, i, time.Now().UTC(), identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)

					f, err := settings.NewFlow(conf, time.Minute, r, sess.Identity, ft)
					require.NoError(t, err)
					if handleErr(t, w, r, reg.SettingsHookExecutor().PreSettingsHook(r.Context(), w, r, f)) {
						_, _ = w.Write([]byte("ok"))
					}
				})

				router.GET("/settings/post", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					if i == nil {
						i = testhelpers.SelfServiceHookCreateFakeIdentity(t, reg)
					}
					sess, _ := testhelpers.NewActiveSession(r, reg, i, time.Now().UTC(), identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)

					a, err := settings.NewFlow(conf, time.Minute, r, sess.Identity, ft)
					require.NoError(t, err)
					a.RequestURL = x.RequestURL(r).String()
					require.NoError(t, reg.SettingsFlowPersister().CreateSettingsFlow(r.Context(), a))
					_ = handleErr(t, w, r, reg.SettingsHookExecutor().
						PostSettingsHook(ctx, w, r, strategy, &settings.UpdateContext{Flow: a, Session: sess}, i))
				})
				ts := httptest.NewServer(router)
				t.Cleanup(ts.Close)
				conf.MustSet(ctx, config.ViperKeyPublicBaseURL, ts.URL)
				return ts
			}

			makeRequestPost := testhelpers.SelfServiceMakeSettingsPostHookRequest
			viperSetPost := func(strategy string, c []config.SelfServiceHook) {
				conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceSettingsAfter, strategy), c)
			}

			uiTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
			t.Cleanup(uiTS.Close)
			uiURL := uiTS.URL + "/user/settings"
			conf.MustSet(ctx, config.ViperKeySelfServiceSettingsURL, uiURL)

			t.Run("method=PostSettingsHook", func(t *testing.T) {
				t.Run("case=pass without hooks", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					res, body := makeRequestPost(t, newServer(t, nil, flow.TypeBrowser), false, url.Values{})
					require.Equalf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.Contains(t, res.Request.URL.String(), uiURL)
				})

				t.Run("case=pass without hooks if ajax client", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					ts := newServer(t, nil, flow.TypeBrowser)
					res, body := makeRequestPost(t, ts, true, url.Values{})
					require.Equalf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.Contains(t, res.Request.URL.String(), ts.URL)
					assert.EqualValues(t, gjson.Get(body, "continue_with.0.action").String(), "redirect_browser_to")
				})

				t.Run("case=pass if hooks pass", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					viperSetPost(strategy, []config.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})
					res, body := makeRequestPost(t, newServer(t, nil, flow.TypeBrowser), false, url.Values{})
					require.Equalf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.Contains(t, res.Request.URL.String(), uiURL)
				})

				t.Run("case=fail if hooks fail", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					viperSetPost(strategy, []config.SelfServiceHook{{Name: "err", Config: []byte(`{"ExecuteSettingsPrePersistHook": "abort"}`)}})
					res, body := makeRequestPost(t, newServer(t, nil, flow.TypeBrowser), false, url.Values{})
					require.Equalf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.Equal(t, "", body)
				})

				t.Run("case=use return_to value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{returnToServer.URL})
					testhelpers.SelfServiceHookSettingsSetDefaultRedirectTo(t, conf, "https://www.ory.sh")

					res, body := makeRequestPost(t, newServer(t, nil, flow.TypeBrowser), false, url.Values{"return_to": {returnToServer.URL + "/kratos"}})
					require.Equalf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.EqualValues(t, returnToServer.URL+"/kratos", res.Request.URL.String())
				})

				t.Run("case=use nested config value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					testhelpers.SelfServiceHookSettingsSetDefaultRedirectTo(t, conf, returnToServer.URL+"/kratos")

					res, body := makeRequestPost(t, newServer(t, nil, flow.TypeBrowser), false, url.Values{})
					require.Equalf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.EqualValues(t, returnToServer.URL+"/kratos", res.Request.URL.String())
				})

				t.Run("case=use nested config value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					testhelpers.SelfServiceHookSettingsSetDefaultRedirectTo(t, conf, "https://www.ory.sh/not-kratos")
					testhelpers.SelfServiceHookSettingsSetDefaultRedirectToStrategy(t, conf, strategy, returnToServer.URL+"/kratos")

					res, body := makeRequestPost(t, newServer(t, nil, flow.TypeBrowser), false, url.Values{})
					require.Equalf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.EqualValues(t, returnToServer.URL+"/kratos", res.Request.URL.String())
				})

				t.Run("case=pass if hooks pass", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					viperSetPost(strategy, []config.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})
					res, body := makeRequestPost(t, newServer(t, nil, flow.TypeBrowser), false, url.Values{})
					require.Equalf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.Contains(t, res.Request.URL.String(), uiURL)
				})

				t.Run("case=send a json response for API clients", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					viperSetPost(strategy, nil)
					res, body := makeRequestPost(t, newServer(t, nil, flow.TypeAPI), true, url.Values{})
					require.Equalf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.NotEmpty(t, gjson.Get(body, "identity.id"))
				})

				t.Run("case=pass without hooks for browser flow with application/json", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					res, body := makeRequestPost(t, newServer(t, nil, flow.TypeBrowser), true, url.Values{})
					require.Equalf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.NotEmpty(t, gjson.Get(body, "identity.id"))
				})
			})

			for _, kind := range []flow.Type{flow.TypeBrowser, flow.TypeAPI} {
				t.Run("type="+string(kind)+"/method=PreSettingsHook", testhelpers.TestSelfServicePreHook(
					config.ViperKeySelfServiceSettingsBeforeHooks,
					testhelpers.SelfServiceMakeSettingsPreHookRequest,
					func(t *testing.T) *httptest.Server {
						return newServer(t, nil, kind)
					},
					conf,
				))
			}
		})
	}
}
