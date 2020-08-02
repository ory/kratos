package testhelpers

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/gobuffalo/httptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
)

func TestSelfServicePreHook(
	configKey string,
	makeRequestPre func(t *testing.T, ts *httptest.Server) (*http.Response, string),
	newServer func(t *testing.T) *httptest.Server,
) func(t *testing.T) {
	return func(t *testing.T) {
		t.Run("case=pass without hooks", func(t *testing.T) {
			t.Cleanup(SelfServiceHookConfigReset)

			res, _ := makeRequestPre(t, newServer(t))
			assert.EqualValues(t, http.StatusOK, res.StatusCode)
		})

		t.Run("case=pass if hooks pass", func(t *testing.T) {
			t.Cleanup(SelfServiceHookConfigReset)
			viper.Set(configKey, []configuration.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})

			res, _ := makeRequestPre(t, newServer(t))
			assert.EqualValues(t, http.StatusOK, res.StatusCode)
		})

		t.Run("case=err if hooks err", func(t *testing.T) {
			t.Cleanup(SelfServiceHookConfigReset)
			viper.Set(configKey, []configuration.SelfServiceHook{{Name: "err", Config: []byte(`{"ExecuteLoginPreHook": "err","ExecuteRegistrationPreHook": "err"}`)}})

			res, body := makeRequestPre(t, newServer(t))
			assert.EqualValues(t, http.StatusInternalServerError, res.StatusCode, "%s", body)
			assert.EqualValues(t, "err", body)
		})

		t.Run("case=abort if hooks aborts", func(t *testing.T) {
			t.Cleanup(SelfServiceHookConfigReset)
			viper.Set(configKey, []configuration.SelfServiceHook{{Name: "err", Config: []byte(`{"ExecuteLoginPreHook": "abort","ExecuteRegistrationPreHook": "abort"}`)}})

			res, body := makeRequestPre(t, newServer(t))
			assert.EqualValues(t, http.StatusOK, res.StatusCode)
			assert.Empty(t, body)
		})

		t.Run("case=redirect", func(t *testing.T) {
			t.Skipf("Skipped because pre-redirect is no longer supported")

			t.Cleanup(SelfServiceHookConfigReset)
			viper.Set(configKey, []configuration.SelfServiceHook{{Name: "redirect", Config: []byte(`{"to": "https://www.ory.sh/"}`)}})

			res, _ := makeRequestPre(t, newServer(t))
			assert.EqualValues(t, http.StatusOK, res.StatusCode)
			assert.EqualValues(t, "https://www.ory.sh/", res.Request.URL.String())
		})
	}
}

func SelfServiceHookCreateFakeIdentity(t *testing.T, reg driver.Registry) *identity.Identity {
	i := SelfServiceHookFakeIdentity(t)
	require.NoError(t, reg.IdentityManager().Create(context.Background(), i))
	return i
}

func SelfServiceHookFakeIdentity(t *testing.T) *identity.Identity {
	var i identity.Identity
	require.NoError(t, faker.FakeData(&i))
	i.Traits = identity.Traits(`{}`)
	return &i
}

func SelfServiceHookConfigReset() {
	viper.Set(configuration.ViperKeySelfServiceLoginAfter, nil)
	viper.Set(configuration.ViperKeySelfServiceLoginBeforeHooks, nil)
	viper.Set(configuration.ViperKeySelfServiceRegistrationAfter, nil)
	viper.Set(configuration.ViperKeySelfServiceRegistrationBeforeHooks, nil)
	viper.Set(configuration.ViperKeySelfServiceSettingsAfter, nil)
}

func SelfServiceHookSettingsSetDefaultRedirectTo(value string) {
	viper.Set(configuration.ViperKeySelfServiceSettingsAfter+"."+configuration.DefaultBrowserReturnURL, value)
}

func SelfServiceHookSettingsSetDefaultRedirectToStrategy(strategy, value string) {
	viper.Set(configuration.ViperKeySelfServiceSettingsAfter+"."+strategy+"."+configuration.DefaultBrowserReturnURL, value)
}
func SelfServiceHookLoginSetDefaultRedirectTo(value string) {
	viper.Set(configuration.ViperKeySelfServiceLoginAfter+"."+configuration.DefaultBrowserReturnURL, value)
}

func SelfServiceHookLoginSetDefaultRedirectToStrategy(strategy, value string) {
	viper.Set(configuration.ViperKeySelfServiceLoginAfter+"."+strategy+"."+configuration.DefaultBrowserReturnURL, value)
}

func SelfServiceHookRegistrationSetDefaultRedirectTo(value string) {
	viper.Set(configuration.ViperKeySelfServiceRegistrationAfter+"."+configuration.DefaultBrowserReturnURL, value)
}

func SelfServiceHookRegistrationSetDefaultRedirectToStrategy(strategy, value string) {
	viper.Set(configuration.ViperKeySelfServiceRegistrationAfter+"."+strategy+"."+configuration.DefaultBrowserReturnURL, value)
}

func SelfServiceHookLoginViperSetPost(strategy string, c []configuration.SelfServiceHook) {
	viper.Set(configuration.HookStrategyKey(configuration.ViperKeySelfServiceLoginAfter, strategy), c)
}

func SelfServiceHookRegistrationViperSetPost(strategy string, c []configuration.SelfServiceHook) {
	viper.Set(configuration.HookStrategyKey(configuration.ViperKeySelfServiceRegistrationAfter, strategy), c)
}

func SelfServiceHookLoginErrorHandler(t *testing.T, w http.ResponseWriter, r *http.Request, err error) bool {
	return selfServiceHookErrorHandler(t, w, r, login.ErrHookAbortRequest, err)
}

func SelfServiceHookRegistrationErrorHandler(t *testing.T, w http.ResponseWriter, r *http.Request, err error) bool {
	return selfServiceHookErrorHandler(t, w, r, registration.ErrHookAbortRequest, err)
}

func SelfServiceHookSettingsErrorHandler(t *testing.T, w http.ResponseWriter, r *http.Request, err error) bool {
	return selfServiceHookErrorHandler(t, w, r, settings.ErrHookAbortRequest, err)
}

func selfServiceHookErrorHandler(t *testing.T, w http.ResponseWriter, r *http.Request, abortErr error, actualErr error) bool {
	if actualErr != nil {
		t.Logf("received error: %+v", actualErr)
		if errors.Is(actualErr, abortErr) {
			return false
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(actualErr.Error()))
		return false
	}
	return true
}

func SelfServiceMakeLoginPreHookRequest(t *testing.T, ts *httptest.Server) (*http.Response, string) {
	return selfServiceMakeHookRequest(t, ts, "/login/pre", false, url.Values{})
}

func SelfServiceMakeLoginPostHookRequest(t *testing.T, ts *httptest.Server, asAPI bool, query url.Values) (*http.Response, string) {
	return selfServiceMakeHookRequest(t, ts, "/login/post", asAPI, query)
}

func SelfServiceMakeRegistrationPreHookRequest(t *testing.T, ts *httptest.Server) (*http.Response, string) {
	return selfServiceMakeHookRequest(t, ts, "/registration/pre", false, url.Values{})
}

func SelfServiceMakeRegistrationPostHookRequest(t *testing.T, ts *httptest.Server, asAPI bool, query url.Values) (*http.Response, string) {
	return selfServiceMakeHookRequest(t, ts, "/registration/post", asAPI, query)
}

func SelfServiceMakeSettingsPostHookRequest(t *testing.T, ts *httptest.Server, asAPI bool, query url.Values) (*http.Response, string) {
	return selfServiceMakeHookRequest(t, ts, "/settings/post", asAPI, query)
}

func selfServiceMakeHookRequest(t *testing.T, ts *httptest.Server, suffix string, asAPI bool, query url.Values) (*http.Response, string) {
	if len(query) > 0 {
		suffix += "?" + query.Encode()
	}
	req, err := http.NewRequest("GET", ts.URL+suffix, nil)
	require.NoError(t, err)
	if asAPI {
		req.Header.Set("Accept", "application/json")
	}
	res, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	return res, string(body)
}
