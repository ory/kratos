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

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/x"
)

func TestRegistrationExecutor(t *testing.T) {
	for _, strategy := range []string{
		identity.CredentialsTypePassword.String(),
		identity.CredentialsTypeOIDC.String(),
	} {
		t.Run("strategy="+strategy, func(t *testing.T) {
			_, reg := internal.NewFastRegistryWithMocks(t)
			viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/registration.schema.json")
			viper.Set(configuration.ViperKeySelfServiceBrowserDefaultReturnTo, "https://www.ory.sh/")

			newServer := func(t *testing.T, i *identity.Identity, ft flow.Type) *httptest.Server {
				router := httprouter.New()
				handleErr := testhelpers.SelfServiceHookRegistrationErrorHandler
				router.GET("/registration/pre", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					if handleErr(t, w, r, reg.RegistrationHookExecutor().PreRegistrationHook(w, r, registration.NewFlow(time.Minute, x.FakeCSRFToken, r, ft))) {
						_, _ = w.Write([]byte("ok"))
					}
				})

				router.GET("/registration/post", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					if i == nil {
						i = testhelpers.SelfServiceHookFakeIdentity(t)
					}
					a := registration.NewFlow(time.Minute, x.FakeCSRFToken, r, ft)
					a.RequestURL = x.RequestURL(r).String()
					_ = handleErr(t, w, r, reg.RegistrationHookExecutor().PostRegistrationHook(w, r, identity.CredentialsType(strategy), a, i))
				})

				ts := httptest.NewServer(router)
				t.Cleanup(ts.Close)
				viper.Set(configuration.ViperKeyPublicBaseURL, ts.URL)
				return ts
			}

			makeRequestPost := testhelpers.SelfServiceMakeRegistrationPostHookRequest
			viperSetPost := testhelpers.SelfServiceHookRegistrationViperSetPost
			t.Run("method=PostRegistrationHook", func(t *testing.T) {
				t.Run("case=pass without hooks", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset)
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
					t.Cleanup(testhelpers.SelfServiceHookConfigReset)
					viperSetPost(strategy, []configuration.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})

					res, _ := makeRequestPost(t, newServer(t, nil, flow.TypeBrowser), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/", res.Request.URL.String())
				})

				t.Run("case=fail if hooks fail", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset)
					viperSetPost(strategy, []configuration.SelfServiceHook{{Name: "err", Config: []byte(`{"ExecutePostRegistrationPrePersistHook": "abort"}`)}})
					i := testhelpers.SelfServiceHookFakeIdentity(t)

					res, body := makeRequestPost(t, newServer(t, i, flow.TypeBrowser), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.Equal(t, "", body)

					_, err := reg.IdentityPool().GetIdentity(context.Background(), i.ID)
					require.Error(t, err)
				})

				t.Run("case=prevent return_to value because domain not whitelisted", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset)
					i := testhelpers.SelfServiceHookFakeIdentity(t)

					res, body := makeRequestPost(t, newServer(t, i, flow.TypeBrowser), false, url.Values{"return_to": {"https://www.ory.sh/kratos/"}})
					assert.EqualValues(t, http.StatusInternalServerError, res.StatusCode)
					assert.Contains(t, body, "malformed or contained invalid")

					actual, err := reg.IdentityPool().GetIdentity(context.Background(), i.ID)
					require.NoError(t, err)
					assert.Equal(t, actual.Traits, i.Traits)
				})

				t.Run("case=use return_to value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset)
					viper.Set(configuration.ViperKeyURLsWhitelistedReturnToDomains, []string{"https://www.ory.sh/"})

					res, _ := makeRequestPost(t, newServer(t, nil, flow.TypeBrowser), false, url.Values{"return_to": {"https://www.ory.sh/kratos/"}})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/kratos/", res.Request.URL.String())
				})

				t.Run("case=use nested config value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset)
					testhelpers.SelfServiceHookRegistrationSetDefaultRedirectToStrategy(strategy, "https://www.ory.sh/kratos")

					res, _ := makeRequestPost(t, newServer(t, nil, flow.TypeBrowser), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/kratos/", res.Request.URL.String())
				})

				t.Run("case=use nested config value", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset)
					testhelpers.SelfServiceHookRegistrationSetDefaultRedirectTo("https://www.ory.sh/not-kratos")
					testhelpers.SelfServiceHookRegistrationSetDefaultRedirectToStrategy(strategy, "https://www.ory.sh/kratos")

					res, _ := makeRequestPost(t, newServer(t, nil, flow.TypeBrowser), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/kratos/", res.Request.URL.String())
				})

				t.Run("case=pass if hooks pass", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset)
					viperSetPost(strategy, []configuration.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})

					res, _ := makeRequestPost(t, newServer(t, nil, flow.TypeBrowser), false, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.EqualValues(t, "https://www.ory.sh/", res.Request.URL.String())
				})

				t.Run("case=send a json response for API clients", func(t *testing.T) {
					t.Cleanup(testhelpers.SelfServiceHookConfigReset)

					res, body := makeRequestPost(t, newServer(t, nil, flow.TypeAPI), true, url.Values{})
					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.NotEmpty(t, gjson.Get(body, "identity.id"))
				})
			})

			t.Run("type=browser/method=PreRegistrationHook", testhelpers.TestSelfServicePreHook(
				configuration.ViperKeySelfServiceRegistrationBeforeHooks,
				testhelpers.SelfServiceMakeRegistrationPreHookRequest,
				func(t *testing.T) *httptest.Server {
					return newServer(t, nil, flow.TypeBrowser)
				},
			))

			t.Run("type=api/method=PreRegistrationHook", testhelpers.TestSelfServicePreHook(
				configuration.ViperKeySelfServiceRegistrationBeforeHooks,
				testhelpers.SelfServiceMakeRegistrationPreHookRequest,
				func(t *testing.T) *httptest.Server {
					return newServer(t, nil, flow.TypeAPI)
				},
			))
		})
	}
}
