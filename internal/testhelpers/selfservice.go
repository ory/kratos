// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/gobuffalo/httptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
)

func TestSelfServicePreHook(
	configKey string,
	makeRequestPre func(t *testing.T, ts *httptest.Server) (*http.Response, string),
	newServer func(t *testing.T) *httptest.Server,
	conf *config.Config,
) func(t *testing.T) {
	ctx := context.Background()
	return func(t *testing.T) {
		t.Run("case=pass without hooks", func(t *testing.T) {
			t.Cleanup(SelfServiceHookConfigReset(t, conf))

			res, _ := makeRequestPre(t, newServer(t))
			assert.EqualValues(t, http.StatusOK, res.StatusCode)
		})

		t.Run("case=pass if hooks pass", func(t *testing.T) {
			t.Cleanup(SelfServiceHookConfigReset(t, conf))
			conf.MustSet(ctx, configKey, []config.SelfServiceHook{{Name: "err", Config: []byte(`{}`)}})

			res, _ := makeRequestPre(t, newServer(t))
			assert.EqualValues(t, http.StatusOK, res.StatusCode)
		})

		t.Run("case=err if hooks err", func(t *testing.T) {
			t.Cleanup(SelfServiceHookConfigReset(t, conf))
			conf.MustSet(ctx, configKey, []config.SelfServiceHook{{Name: "err", Config: []byte(`{"ExecuteLoginPreHook": "err","ExecuteRegistrationPreHook": "err","ExecuteSettingsPreHook": "err","ExecuteVerificationPreHook": "err","ExecuteRecoveryPreHook": "err"}`)}})

			res, body := makeRequestPre(t, newServer(t))
			assert.EqualValues(t, http.StatusInternalServerError, res.StatusCode, "%s", body)
			assert.EqualValues(t, "err", body)
		})

		t.Run("case=abort if hooks aborts", func(t *testing.T) {
			t.Cleanup(SelfServiceHookConfigReset(t, conf))
			conf.MustSet(ctx, configKey, []config.SelfServiceHook{{Name: "err", Config: []byte(`{"ExecuteLoginPreHook": "abort","ExecuteRegistrationPreHook": "abort","ExecuteSettingsPreHook": "abort","ExecuteVerificationPreHook": "abort","ExecuteRecoveryPreHook": "abort"}`)}})

			res, body := makeRequestPre(t, newServer(t))
			assert.EqualValues(t, http.StatusOK, res.StatusCode)
			assert.Empty(t, body)
		})

		t.Run("case=redirect", func(t *testing.T) {
			t.Skipf("Skipped because pre-redirect is no longer supported")

			t.Cleanup(SelfServiceHookConfigReset(t, conf))
			conf.MustSet(ctx, configKey, []config.SelfServiceHook{{Name: "redirect", Config: []byte(`{"to": "https://www.ory.sh/"}`)}})

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
	i.State = identity.StateActive
	return &i
}

func SelfServiceHookConfigReset(t *testing.T, conf *config.Config) func() {
	ctx := context.Background()
	return func() {
		conf.MustSet(ctx, config.ViperKeySelfServiceLoginAfter, nil)
		conf.MustSet(ctx, config.ViperKeySelfServiceLoginAfter+".hooks", nil)
		conf.MustSet(ctx, config.ViperKeySelfServiceLoginBeforeHooks, nil)
		conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryAfter, nil)
		conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryAfter+".hooks", nil)
		conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationAfter, nil)
		conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationAfter+".hooks", nil)
		conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationPrePersist, nil)
		conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationPrePersist+".hooks", nil)
		conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationPostPersist, nil)
		conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationPostPersist+".hooks", nil)
		conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationBeforeHooks, nil)
		conf.MustSet(ctx, config.ViperKeySelfServiceSettingsAfter, nil)
		conf.MustSet(ctx, config.ViperKeySelfServiceSettingsAfter+".hooks", nil)
	}
}

func SelfServiceHookSettingsSetDefaultRedirectTo(t *testing.T, conf *config.Config, value string) {
	ctx := context.Background()
	conf.MustSet(ctx, config.ViperKeySelfServiceSettingsAfter+"."+config.DefaultBrowserReturnURL, value)
}

func SelfServiceHookSettingsSetDefaultRedirectToStrategy(t *testing.T, conf *config.Config, strategy, value string) {
	ctx := context.Background()
	conf.MustSet(ctx, config.ViperKeySelfServiceSettingsAfter+"."+strategy+"."+config.DefaultBrowserReturnURL, value)
}

func SelfServiceHookLoginSetDefaultRedirectTo(t *testing.T, conf *config.Config, value string) {
	ctx := context.Background()
	conf.MustSet(ctx, config.ViperKeySelfServiceLoginAfter+"."+config.DefaultBrowserReturnURL, value)
}

func SelfServiceHookLoginSetDefaultRedirectToStrategy(t *testing.T, conf *config.Config, strategy, value string) {
	ctx := context.Background()
	conf.MustSet(ctx, config.ViperKeySelfServiceLoginAfter+"."+strategy+"."+config.DefaultBrowserReturnURL, value)
}

func SelfServiceHookRegistrationSetDefaultRedirectTo(t *testing.T, conf *config.Config, value string) {
	ctx := context.Background()
	conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationAfter+"."+config.DefaultBrowserReturnURL, value)
}

func SelfServiceHookRegistrationSetDefaultRedirectToStrategy(t *testing.T, conf *config.Config, strategy, value string) {
	ctx := context.Background()
	conf.MustSet(ctx, config.ViperKeySelfServiceRegistrationAfter+"."+strategy+"."+config.DefaultBrowserReturnURL, value)
}

func SelfServiceHookLoginViperSetPost(t *testing.T, conf *config.Config, strategy string, c []config.SelfServiceHook) {
	ctx := context.Background()
	conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceLoginAfter, strategy), c)
}

func SelfServiceHookRegistrationViperSetPost(t *testing.T, conf *config.Config, strategy string, c []config.SelfServiceHook) {
	ctx := context.Background()
	conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRegistrationPostPersist, strategy), c)
}

func SelfServiceHookRegistrationViperSetPostPre(t *testing.T, conf *config.Config, strategy string, c []config.SelfServiceHook) {
	ctx := context.Background()
	conf.MustSet(ctx, config.HookStrategyKey(config.ViperKeySelfServiceRegistrationPrePersist, strategy), c)
}

func SelfServiceHookLoginErrorHandler(t *testing.T, w http.ResponseWriter, r *http.Request, err error) bool {
	return SelfServiceHookErrorHandler(t, w, r, login.ErrHookAbortFlow, err)
}

func SelfServiceHookRegistrationErrorHandler(t *testing.T, w http.ResponseWriter, r *http.Request, err error) bool {
	return SelfServiceHookErrorHandler(t, w, r, registration.ErrHookAbortFlow, err)
}

func SelfServiceHookSettingsErrorHandler(t *testing.T, w http.ResponseWriter, r *http.Request, err error) bool {
	return SelfServiceHookErrorHandler(t, w, r, settings.ErrHookAbortFlow, err)
}

func SelfServiceHookErrorHandler(t *testing.T, w http.ResponseWriter, r *http.Request, abortErr error, actualErr error) bool {
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
	return SelfServiceMakeHookRequest(t, ts, "/login/pre", false, url.Values{})
}

func SelfServiceMakeLoginPostHookRequest(t *testing.T, ts *httptest.Server, asAPI bool, query url.Values) (*http.Response, string) {
	return SelfServiceMakeHookRequest(t, ts, "/login/post", asAPI, query)
}

func SelfServiceMakeRegistrationPreHookRequest(t *testing.T, ts *httptest.Server) (*http.Response, string) {
	return SelfServiceMakeHookRequest(t, ts, "/registration/pre", false, url.Values{})
}

func SelfServiceMakeSettingsPreHookRequest(t *testing.T, ts *httptest.Server) (*http.Response, string) {
	return SelfServiceMakeHookRequest(t, ts, "/settings/pre", false, url.Values{})
}

func SelfServiceMakeRecoveryPreHookRequest(t *testing.T, ts *httptest.Server) (*http.Response, string) {
	return SelfServiceMakeHookRequest(t, ts, "/recovery/pre", false, url.Values{})
}

func SelfServiceMakeVerificationPreHookRequest(t *testing.T, ts *httptest.Server) (*http.Response, string) {
	return SelfServiceMakeHookRequest(t, ts, "/verification/pre", false, url.Values{})
}

func SelfServiceMakeRegistrationPrePersistHookRequest(t *testing.T, ts *httptest.Server) (*http.Response, string) {
	return SelfServiceMakeHookRequest(t, ts, "/registration/pre-persist", false, url.Values{})
}

func SelfServiceMakeRegistrationPostHookRequest(t *testing.T, ts *httptest.Server, asAPI bool, query url.Values) (*http.Response, string) {
	return SelfServiceMakeHookRequest(t, ts, "/registration/post", asAPI, query)
}

func SelfServiceMakeSettingsPostHookRequest(t *testing.T, ts *httptest.Server, asAPI bool, query url.Values) (*http.Response, string) {
	return SelfServiceMakeHookRequest(t, ts, "/settings/post", asAPI, query)
}

func SelfServiceMakeHookRequest(t *testing.T, ts *httptest.Server, suffix string, asAPI bool, query url.Values) (*http.Response, string) {
	if len(query) > 0 {
		suffix += "?" + query.Encode()
	}
	req, err := http.NewRequest("GET", ts.URL+suffix, nil)
	require.NoError(t, err)
	req.Header.Set("Accept", "text/html")
	if asAPI {
		req.Header.Set("Accept", "application/json")
	}
	res, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	return res, string(body)
}

func GetSelfServiceRedirectLocation(t *testing.T, url string) string {
	c := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)
	res, err := c.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	return res.Header.Get("Location")
}

func AssertMessage(t *testing.T, body []byte, message string) {
	t.Helper()
	assert.Len(t, gjson.GetBytes(body, "ui.messages").Array(), 1)
	assert.Equal(t, message, gjson.GetBytes(body, "ui.messages.0.text").String(), "%v", string(body))
}

func AssertFieldMessage(t *testing.T, body []byte, fieldName string, message string) {
	t.Helper()
	messages := gjson.GetBytes(body, "ui.nodes.#(attributes.name=="+fieldName+").messages")
	assert.Len(t, messages.Array(), 1, "expected field %s to have one message, got %s", fieldName, messages)
	assert.Equal(t, message, messages.Get("0.text").String(), "%v", string(body))
}
