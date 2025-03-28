// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package registration_test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gobuffalo/httptest"
	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hydra"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/hook"
	"github.com/ory/kratos/x"
)

func TestRegistrationExecutor(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	for _, strategy := range identity.AllCredentialTypes {
		strategy := strategy.String()

		t.Run("strategy="+strategy, func(t *testing.T) {
			t.Parallel()

			conf, reg := internal.NewFastRegistryWithMocks(t)
			reg.WithHydra(hydra.NewFake())
			testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/registration.schema.json")
			conf.MustSet(ctx, config.ViperKeySelfServiceBrowserDefaultReturnTo, returnToServer.URL)

			newServer := func(t *testing.T, i *identity.Identity, ft flow.Type, flowCallbacks ...func(*registration.Flow)) *httptest.Server {
				router := httprouter.New()

				handleErr := testhelpers.SelfServiceHookRegistrationErrorHandler
				router.GET("/registration/pre", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					f, err := registration.NewFlow(conf, time.Minute, x.FakeCSRFToken, r, ft)
					require.NoError(t, err)
					if handleErr(t, w, r, reg.RegistrationHookExecutor().PreRegistrationHook(w, r, f)) {
						_, _ = w.Write([]byte("ok"))
					}
				})

				router.GET("/registration/post", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					if i == nil {
						i = testhelpers.SelfServiceHookFakeIdentity(t)
					}
					regFlow, err := registration.NewFlow(conf, time.Minute, x.FakeCSRFToken, r, ft)
					require.NoError(t, err)
					regFlow.RequestURL = x.RequestURL(r).String()
					for _, callback := range flowCallbacks {
						callback(regFlow)
					}
					_ = handleErr(t, w, r, reg.RegistrationHookExecutor().PostRegistrationHook(w, r, identity.CredentialsType(strategy), "", "", regFlow, i))
				})

				ts := httptest.NewServer(router)
				t.Cleanup(ts.Close)
				conf.MustSet(ctx, config.ViperKeyPublicBaseURL, ts.URL)
				return ts
			}

			makeRequestPost := testhelpers.SelfServiceMakeRegistrationPostHookRequest
			viperSetPost := testhelpers.SelfServiceHookRegistrationViperSetPost
			t.Run("method=PostRegistrationHook", func(t *testing.T) {
				t.Run("case=pass without hooks", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					i := testhelpers.SelfServiceHookFakeIdentity(t)

					ts := newServer(t, i, flow.TypeBrowser)
					res, body := makeRequestPost(t, ts, false, url.Values{})
					require.EqualValuesf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.EqualValues(t, returnToServer.URL, res.Request.URL.String())

					actual, err := reg.IdentityPool().GetIdentity(context.Background(), i.ID, identity.ExpandNothing)
					require.NoError(t, err)
					assert.Equal(t, actual.Traits, i.Traits)
				})

				t.Run("case=pass without hooks if ajax client", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					i := testhelpers.SelfServiceHookFakeIdentity(t)

					ts := newServer(t, i, flow.TypeBrowser)
					res, body := makeRequestPost(t, ts, true, url.Values{})
					require.EqualValuesf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.Contains(t, res.Request.URL.String(), ts.URL)
					assert.JSONEq(t, fmt.Sprintf(`[{"action":"redirect_browser_to","redirect_browser_to":"%s"}]`, returnToServer.URL), gjson.Get(body, "continue_with").Raw)

					actual, err := reg.IdentityPool().GetIdentity(context.Background(), i.ID, identity.ExpandNothing)
					require.NoError(t, err)
					assert.Equal(t, actual.Traits, i.Traits)
				})

				t.Run("case=pass if hooks pass", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					viperSetPost(t, conf, strategy, []config.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})

					res, body := makeRequestPost(t, newServer(t, nil, flow.TypeBrowser), false, url.Values{})
					require.EqualValuesf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.EqualValues(t, returnToServer.URL, res.Request.URL.String())
				})

				t.Run("case=fail if hooks fail", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					viperSetPost(t, conf, strategy, []config.SelfServiceHook{{Name: "err", Config: []byte(`{"ExecutePostRegistrationPrePersistHook": "abort"}`)}})
					i := testhelpers.SelfServiceHookFakeIdentity(t)

					res, body := makeRequestPost(t, newServer(t, i, flow.TypeBrowser), false, url.Values{})
					require.EqualValuesf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.Equal(t, "", body)

					_, err := reg.IdentityPool().GetIdentity(context.Background(), i.ID, identity.ExpandNothing)
					require.Error(t, err)
				})

				t.Run("case=use return_to value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{returnToServer.URL})

					res, body := makeRequestPost(t, newServer(t, nil, flow.TypeBrowser), false, url.Values{"return_to": {returnToServer.URL + "/kratos"}})
					require.EqualValuesf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.EqualValues(t, returnToServer.URL+"/kratos", res.Request.URL.String())
				})

				t.Run("case=use nested config value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					testhelpers.SelfServiceHookRegistrationSetDefaultRedirectToStrategy(t, conf, strategy, returnToServer.URL+"/kratos")

					res, body := makeRequestPost(t, newServer(t, nil, flow.TypeBrowser), false, url.Values{})
					require.EqualValuesf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.EqualValues(t, returnToServer.URL+"/kratos", res.Request.URL.String())
				})

				t.Run("case=use nested config value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					conf.MustSet(ctx, config.ViperKeyURLsAllowedReturnToDomains, []string{returnToServer.URL + "/kratos"})
					testhelpers.SelfServiceHookRegistrationSetDefaultRedirectTo(t, conf, returnToServer.URL+"/not-kratos")
					testhelpers.SelfServiceHookRegistrationSetDefaultRedirectToStrategy(t, conf, strategy, returnToServer.URL+"/kratos")

					res, body := makeRequestPost(t, newServer(t, nil, flow.TypeBrowser), false, url.Values{})
					require.EqualValuesf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.EqualValues(t, returnToServer.URL+"/kratos", res.Request.URL.String())
				})

				t.Run("case=pass if hooks pass", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					viperSetPost(t, conf, strategy, []config.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})

					res, body := makeRequestPost(t, newServer(t, nil, flow.TypeBrowser), false, url.Values{})
					require.EqualValuesf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.EqualValues(t, returnToServer.URL, res.Request.URL.String())
				})

				t.Run("case=send a json response for API clients", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					res, body := makeRequestPost(t, newServer(t, nil, flow.TypeAPI), true, url.Values{})
					require.EqualValuesf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.NotEmpty(t, gjson.Get(body, "identity.id"))
				})

				t.Run("case=pass without hooks for browser flow with application/json", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))

					res, body := makeRequestPost(t, newServer(t, nil, flow.TypeBrowser), true, url.Values{})
					require.EqualValuesf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.NotEmpty(t, gjson.Get(body, "identity.id"))
					assert.Empty(t, gjson.Get(body, "session.token"))
					assert.Empty(t, gjson.Get(body, "session_token"))
				})

				t.Run("case=should redirect to verification UI if show_verification_ui hook is set", func(t *testing.T) {
					verificationTS := testhelpers.NewVerificationUIFlowEchoServer(t, reg)
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					conf.MustSet(ctx, config.ViperKeySelfServiceVerificationEnabled, true)
					conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationAfter+".hooks", []map[string]interface{}{
						{
							"hook": hook.KeyVerificationUI,
						},
					})
					i := testhelpers.SelfServiceHookFakeIdentity(t)
					i.Traits = identity.Traits(`{"email": "verifiable@ory.sh"}`)

					res, body := makeRequestPost(t, newServer(t, i, flow.TypeBrowser), false, url.Values{})
					require.EqualValuesf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.Contains(t, res.Request.URL.String(), verificationTS.URL)
					assert.NotEmpty(t, res.Request.URL.Query().Get("flow"))
				})

				t.Run("case=should redirect to verification UI if there is a login_challenge", func(t *testing.T) {
					verificationTS := testhelpers.NewVerificationUIFlowEchoServer(t, reg)
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					conf.MustSet(ctx, config.ViperKeySelfServiceVerificationEnabled, true)
					conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationAfter+".hooks", []map[string]interface{}{{
						"hook": hook.KeyVerificationUI,
					}})
					i := testhelpers.SelfServiceHookFakeIdentity(t)
					i.Traits = identity.Traits(`{"email": "verifiable-valid-login_challenge@ory.sh"}`)

					withOAuthChallenge := func(f *registration.Flow) {
						f.OAuth2LoginChallenge = hydra.FakeValidLoginChallenge
					}
					res, body := makeRequestPost(t, newServer(t, i, flow.TypeBrowser, withOAuthChallenge), false, url.Values{})
					require.EqualValuesf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.Contains(t, res.Request.URL.String(), verificationTS.URL)
					flowID := res.Request.URL.Query().Get("flow")
					require.NotEmpty(t, flowID)
					flow, err := reg.VerificationFlowPersister().GetVerificationFlow(ctx, uuid.Must(uuid.FromString(flowID)))
					require.NoError(t, err)
					assert.Equal(t, hydra.FakeValidLoginChallenge, flow.OAuth2LoginChallenge.String())
				})

				t.Run("case=should redirect to first verification UI if show_verification_ui hook is set and multiple verifiable addresses", func(t *testing.T) {
					verificationTS := testhelpers.NewVerificationUIFlowEchoServer(t, reg)
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					conf.MustSet(ctx, config.ViperKeySelfServiceVerificationEnabled, true)
					conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationAfter+".hooks", []map[string]interface{}{
						{
							"hook": hook.KeyVerificationUI,
						},
					})

					i := testhelpers.SelfServiceHookFakeIdentity(t)
					i.SchemaID = testhelpers.UseIdentitySchema(t, conf, "file://./stub/registration-multi-email.schema.json")
					i.Traits = identity.Traits(`{"emails": ["one@ory.sh", "two@ory.sh"]}`)

					res, body := makeRequestPost(t, newServer(t, i, flow.TypeBrowser), false, url.Values{})
					require.EqualValuesf(t, http.StatusOK, res.StatusCode, "%s", body)
					assert.Contains(t, res.Request.URL.String(), verificationTS.URL)
					assert.NotEmpty(t, res.Request.URL.Query().Get("flow"))
				})

				t.Run("case=should still sent session if show_verification_ui is set after session hook", func(t *testing.T) {
					verificationTS := testhelpers.NewVerificationUIFlowEchoServer(t, reg)
					t.Cleanup(testhelpers.SelfServiceHookConfigReset(t, conf))
					conf.MustSet(ctx, config.ViperKeySelfServiceVerificationEnabled, true)
					conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationAfter+".hooks", []map[string]interface{}{
						{
							"hook": hook.KeyVerificationUI,
						},
						{
							"hook": hook.KeySessionIssuer,
						},
					})

					i := testhelpers.SelfServiceHookFakeIdentity(t)
					i.Traits = identity.Traits(`{"email": "verifiable4@ory.sh"}`)

					jar := testhelpers.EasyCookieJar(t, nil)
					s := newServer(t, i, flow.TypeBrowser)
					s.Client().Jar = jar
					res, _ := makeRequestPost(t, s, false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.Contains(t, res.Request.URL.String(), verificationTS.URL)
					assert.NotEmpty(t, res.Request.URL.Query().Get("flow"))
					u, err := url.Parse(s.URL)
					require.NoError(t, err)
					cookies := jar.Cookies(u)
					require.Len(t, cookies, 1)
					assert.Equal(t, "ory_kratos_session", cookies[0].Name)
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
