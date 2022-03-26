// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package registration_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gobuffalo/httptest"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/x"
)

func TestRegistrationExecutor(t *testing.T) {
	ctx := context.Background()
	for _, strategy := range []string{
		identity.CredentialsTypePassword.String(),
		identity.CredentialsTypeOIDC.String(),
		identity.CredentialsTypeTOTP.String(),
		identity.CredentialsTypeWebAuthn.String(),
	} {
		t.Run("strategy="+strategy, func(t *testing.T) {
			conf, reg := internal.NewFastRegistryWithMocks(t)
			testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/registration.schema.json")
			conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh/")

			newServer := func(t *testing.T, i *identity.Identity, ft flow.Type) *httptest.Server {
				router := httprouter.New()
				handleErr := testhelpers.SelfServiceHookRegistrationErrorHandler
				router.GET("/registration/pre", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					f, err := registration.NewFlow(conf, time.Minute, x.FakeCSRFToken, r, ft)
					require.NoError(t, err)
					if handleErr(t, w, r, reg.RegistrationHookExecutor().PreRegistrationHook(w, r, f)) {
						_, _ = w.Write([]byte("ok"))
					}
				})

				router.GET("/registration/pre-persist", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					if i == nil {
						i = testhelpers.SelfServiceHookFakeIdentity(t)
					}
					a, err := registration.NewFlow(conf, time.Minute, x.FakeCSRFToken, r, ft)
					require.NoError(t, err)
					a.RequestURL = x.RequestURL(r).String()
					_ = handleErr(t, w, r, reg.RegistrationHookExecutor().PostRegistrationHook(w, r, identity.CredentialsType(strategy), a, i))
				})

				router.GET("/registration/post", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					if i == nil {
						i = testhelpers.SelfServiceHookFakeIdentity(t)
					}
					a, err := registration.NewFlow(conf, time.Minute, x.FakeCSRFToken, r, ft)
					require.NoError(t, err)
					a.RequestURL = x.RequestURL(r).String()
					_ = handleErr(t, w, r, reg.RegistrationHookExecutor().PostRegistrationHook(w, r, identity.CredentialsType(strategy), a, i))
				})

				ts := httptest.NewServer(router)
				t.Cleanup(ts.Close)
				conf.MustSet(ctx, config.ViperKeyPublicBaseURL, ts.URL)
				return ts
			}

			makeRequestPost := testhelpers.SelfServiceMakeRegistrationPostHookRequest
			viperSetPost := testhelpers.SelfServiceHookRegistrationViperSetPost
			viperSetPostPre := testhelpers.SelfServiceHookRegistrationViperSetPostPre
			t.Run("method=PostRegistrationHook", func(t *testing.T) {
				t.Run("case=pass without hooks", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					i := testhelpers.SelfServiceHookFakeIdentity(t)

					ts := newServer(t, i, flow.TypeBrowser)
					res, _ := makeRequestPost(t, ts, false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/", res.Request.URL.String())

					actual, err := reg.IdentityPool().GetIdentity(context.Background(), i.ID)
					require.NoError(t, err)
					assert.Equal(t, actual.Traits, i.Traits)
				})

				t.Run("case=pass if hooks pass", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					viperSetPost(t, conf, strategy, []config.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})

					res, _ := makeRequestPost(t, newServer(t, nil, flow.TypeBrowser), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/", res.Request.URL.String())
				})

				t.Run("case=fail if hooks fail", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					viperSetPostPre(t, conf, strategy, []config.SelfServiceHook{{Name: "err", Config: []byte(`{"ExecutePostRegistrationPrePersistHook": "abort"}`)}})
					i := testhelpers.SelfServiceHookFakeIdentity(t)

					res, body := makeRequestPost(t, newServer(t, i, flow.TypeBrowser), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.Equal(t, "", body)

					_, err := reg.IdentityPool().GetIdentity(context.Background(), i.ID)
					require.Error(t, err)
				})

				t.Run("case=fail if hooks fail, but still creates if post", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					viperSetPost(t, conf, strategy, []config.SelfServiceHook{{Name: "err", Config: []byte(`{"ExecutePostRegistrationPostPersistHook": "abort"}`)}})
					i := testhelpers.SelfServiceHookFakeIdentity(t)

					res, body := makeRequestPost(t, newServer(t, i, flow.TypeBrowser), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.Equal(t, "", body)

					actual, err := reg.IdentityPool().GetIdentity(context.Background(), i.ID)
					require.NoError(t, err)
					assert.Equal(t, actual.Traits, i.Traits)
				})

				t.Run("case=use return_to value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{"https://www.ory.sh/"})

					res, _ := makeRequestPost(t, newServer(t, nil, flow.TypeBrowser), false, url.Values{"return_to": {"https://www.ory.sh/kratos/"}})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/kratos/", res.Request.URL.String())
				})

				t.Run("case=use nested config value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					testhelpers.SelfServiceHookRegistrationSetDefaultRedirectToStrategy(t, conf, strategy, "https://www.ory.sh/kratos")

					res, _ := makeRequestPost(t, newServer(t, nil, flow.TypeBrowser), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/kratos/", res.Request.URL.String())
				})

				t.Run("case=use nested config value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{"https://www.ory.sh/kratos"})
					testhelpers.SelfServiceHookRegistrationSetDefaultRedirectTo(t, conf, "https://www.ory.sh/not-kratos")
					testhelpers.SelfServiceHookRegistrationSetDefaultRedirectToStrategy(t, conf, strategy, "https://www.ory.sh/kratos")

					res, _ := makeRequestPost(t, newServer(t, nil, flow.TypeBrowser), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/kratos/", res.Request.URL.String())
				})

				t.Run("case=pass if hooks pass", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					viperSetPost(t, conf, strategy, []config.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})

					res, _ := makeRequestPost(t, newServer(t, nil, flow.TypeBrowser), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/", res.Request.URL.String())
				})

				t.Run("case=send a json response for API clients", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					res, body := makeRequestPost(t, newServer(t, nil, flow.TypeAPI), true, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.NotEmpty(t, gjson.Get(body, "identity.id"))
				})

				t.Run("case=pass without hooks for browser flow with application/json", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					res, body := makeRequestPost(t, newServer(t, nil, flow.TypeBrowser), true, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.NotEmpty(t, gjson.Get(body, "identity.id"))
					assert.Empty(t, gjson.Get(body, "session.token"))
					assert.Empty(t, gjson.Get(body, "session_token"))
				})
			})

			for _, kind := range []flow.Type{flow.TypeBrowser, flow.TypeAPI} {
				t.Run("type="+string(kind)+"/method=PreRegistrationHook", testhelpers.TestSelfServicePreHook(
					config.ViperKeySelfServiceRegistrationBeforeHooks,
					testhelpers.SelfServiceMakeRegistrationPreHookRequest,
					func(t *testing.T) *httptest.Server {
						return newServer(t, nil, kind)
					},
					conf,
				))
			}
		})
	}
}
