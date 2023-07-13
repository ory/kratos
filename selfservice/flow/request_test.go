// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package flow_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/flow"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/x"
)

func TestVerifyRequest(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	require.EqualError(t, flow.EnsureCSRF(reg, &http.Request{}, flow.TypeBrowser, false, x.FakeCSRFTokenGenerator, "not_csrf_token"), x.ErrInvalidCSRFToken.Error())
	require.NoError(t, flow.EnsureCSRF(reg, &http.Request{}, flow.TypeBrowser, false, x.FakeCSRFTokenGenerator, x.FakeCSRFToken), nil)
	require.NoError(t, flow.EnsureCSRF(reg, &http.Request{}, flow.TypeAPI, false, x.FakeCSRFTokenGenerator, ""))
	require.EqualError(t, flow.EnsureCSRF(reg, &http.Request{
		Header: http.Header{"Origin": {"https://www.ory.sh"}},
	}, flow.TypeAPI, false, x.FakeCSRFTokenGenerator, ""), flow.ErrOriginHeaderNeedsBrowserFlow.Error())
	require.EqualError(t, flow.EnsureCSRF(reg, &http.Request{
		Header: http.Header{"Cookie": {"cookie=ory"}},
	}, flow.TypeAPI, false, x.FakeCSRFTokenGenerator, ""), flow.ErrCookieHeaderNeedsBrowserFlow.Error())

	// Cloudflare
	require.NoError(t, flow.EnsureCSRF(reg, &http.Request{
		Header: http.Header{"Cookie": {"__cflb=0pg1RtZzPoPDprTf8gX3TJm8XF5hKZ4pZV74UCe7"}},
	}, flow.TypeAPI, false, x.FakeCSRFTokenGenerator, ""), flow.ErrCookieHeaderNeedsBrowserFlow.Error())
	require.NoError(t, flow.EnsureCSRF(reg, &http.Request{
		Header: http.Header{"Cookie": {"__cflb=0pg1RtZzPoPDprTf8gX3TJm8XF5hKZ4pZV74UCe7; __cfruid=0pg1RtZzPoPDprTf8gX3TJm8XF5hKZ4pZV74UCe7"}},
	}, flow.TypeAPI, false, x.FakeCSRFTokenGenerator, ""), flow.ErrCookieHeaderNeedsBrowserFlow.Error())
	require.Error(t, flow.EnsureCSRF(reg, &http.Request{
		Header: http.Header{"Cookie": {"__cflb=0pg1RtZzPoPDprTf8gX3TJm8XF5hKZ4pZV74UCe7; __cfruid=0pg1RtZzPoPDprTf8gX3TJm8XF5hKZ4pZV74UCe7; some_cookie=some_value"}},
	}, flow.TypeAPI, false, x.FakeCSRFTokenGenerator, ""), flow.ErrCookieHeaderNeedsBrowserFlow.Error())
	require.Error(t, flow.EnsureCSRF(reg, &http.Request{
		Header: http.Header{"Cookie": {"some_cookie=some_value"}},
	}, flow.TypeAPI, false, x.FakeCSRFTokenGenerator, ""), flow.ErrCookieHeaderNeedsBrowserFlow.Error())
	require.NoError(t, flow.EnsureCSRF(reg, &http.Request{}, flow.TypeAPI, false, x.FakeCSRFTokenGenerator, ""), flow.ErrCookieHeaderNeedsBrowserFlow.Error())
}

func TestMethodEnabledAndAllowed(t *testing.T) {
	ctx := context.Background()
	conf, d := internal.NewFastRegistryWithMocks(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := flow.MethodEnabledAndAllowedFromRequest(r, flow.LoginFlow, "password", d); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(ts.Close)

	t.Run("allowed", func(t *testing.T) {
		res, err := ts.Client().PostForm(ts.URL, url.Values{"method": {"password"}})
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		assert.Equal(t, http.StatusNoContent, res.StatusCode)
	})

	t.Run("unknown", func(t *testing.T) {
		res, err := ts.Client().PostForm(ts.URL, url.Values{"method": {"other"}})
		require.NoError(t, err)
		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
		assert.Contains(t, string(body), "is not responsible for this request")
	})

	t.Run("disabled", func(t *testing.T) {
		require.NoError(t, conf.Set(ctx, fmt.Sprintf("%s.%s.enabled", config.ViperKeySelfServiceStrategyConfig, "password"), false))
		res, err := ts.Client().PostForm(ts.URL, url.Values{"method": {"password"}})
		require.NoError(t, err)
		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
		assert.Contains(t, string(body), "The requested resource could not be found")
	})
}

func TestMethodCodeEnabledAndAllowed(t *testing.T) {
	ctx := context.Background()
	conf, d := internal.NewFastRegistryWithMocks(t)

	currentFlow := make(chan flow.FlowName, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f := <-currentFlow
		if err := flow.MethodEnabledAndAllowedFromRequest(r, f, "code", d); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	t.Run("login code allowed", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.login_enabled", true)
		currentFlow <- flow.LoginFlow
		res, err := ts.Client().PostForm(ts.URL, url.Values{"method": {"code"}})
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		assert.Equal(t, http.StatusNoContent, res.StatusCode)
	})

	t.Run("login code not allowed", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.login_enabled", false)
		currentFlow <- flow.LoginFlow
		res, err := ts.Client().PostForm(ts.URL, url.Values{"method": {"code"}})
		require.NoError(t, err)
		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
		assert.Contains(t, string(body), "The requested resource could not be found")
	})

	t.Run("registration code allowed", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.registration_enabled", true)
		currentFlow <- flow.RegistrationFlow
		res, err := ts.Client().PostForm(ts.URL, url.Values{"method": {"code"}})
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		assert.Equal(t, http.StatusNoContent, res.StatusCode)
	})

	t.Run("registration code not allowed", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.registration_enabled", false)
		currentFlow <- flow.RegistrationFlow
		res, err := ts.Client().PostForm(ts.URL, url.Values{"method": {"code"}})
		require.NoError(t, err)
		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
		assert.Contains(t, string(body), "The requested resource could not be found")
	})

	t.Run("recovery and verification should still be allowed if registration and login is disabled", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.registration_enabled", false)
		conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.login_enabled", false)
		conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+".code.enabled", true)

		for _, f := range []flow.FlowName{flow.RecoveryFlow, flow.VerificationFlow} {
			currentFlow <- f
			res, err := ts.Client().PostForm(ts.URL, url.Values{"method": {"code"}})
			require.NoError(t, err)
			require.NoError(t, res.Body.Close())
			assert.Equal(t, http.StatusNoContent, res.StatusCode)
		}
	})
}
