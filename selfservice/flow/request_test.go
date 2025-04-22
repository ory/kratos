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

	"github.com/ory/kratos/x/nosurfx"

	"github.com/stretchr/testify/assert"

	"github.com/ory/herodot"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/flow"

	"github.com/stretchr/testify/require"
)

func TestVerifyRequest(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	require.EqualError(t, flow.EnsureCSRF(reg, &http.Request{}, flow.TypeBrowser, false, nosurfx.FakeCSRFTokenGenerator, "not_csrf_token"), nosurfx.ErrInvalidCSRFToken.Error())
	require.NoError(t, flow.EnsureCSRF(reg, &http.Request{}, flow.TypeBrowser, false, nosurfx.FakeCSRFTokenGenerator, nosurfx.FakeCSRFToken), nil)
	require.NoError(t, flow.EnsureCSRF(reg, &http.Request{}, flow.TypeAPI, false, nosurfx.FakeCSRFTokenGenerator, ""))
	require.EqualError(t, flow.EnsureCSRF(reg, &http.Request{
		Header: http.Header{"Origin": {"https://www.ory.sh"}},
	}, flow.TypeAPI, false, nosurfx.FakeCSRFTokenGenerator, ""), flow.ErrOriginHeaderNeedsBrowserFlow.Error())
	require.EqualError(t, flow.EnsureCSRF(reg, &http.Request{
		Header: http.Header{"Cookie": {"cookie=ory"}},
	}, flow.TypeAPI, false, nosurfx.FakeCSRFTokenGenerator, ""), flow.ErrCookieHeaderNeedsBrowserFlow.Error(), "should error because of cookie=ory")

	err := flow.EnsureCSRF(reg, &http.Request{
		Header: http.Header{"Cookie": {"cookie1=cookievalue", "cookie2=cookievalue"}},
	}, flow.TypeAPI, false, nosurfx.FakeCSRFTokenGenerator, "")
	var he herodot.DetailsCarrier
	require.ErrorAs(t, err, &he)
	cs, ok := he.Details()["found cookies"].([]string)
	require.True(t, ok)
	require.ElementsMatch(t, cs, []string{"cookie1", "cookie2"})

	// Cloudflare
	require.NoError(t, flow.EnsureCSRF(reg, &http.Request{
		Header: http.Header{"Cookie": {"__cflb=0pg1RtZzPoPDprTf8gX3TJm8XF5hKZ4pZV74UCe7", "_cfuvid=blub", "cf_clearance=bla"}},
	}, flow.TypeAPI, false, nosurfx.FakeCSRFTokenGenerator, ""), "should ignore Cloudflare cookies")
	require.NoError(t, flow.EnsureCSRF(reg, &http.Request{
		Header: http.Header{"Cookie": {"__cflb=0pg1RtZzPoPDprTf8gX3TJm8XF5hKZ4pZV74UCe7; __cfruid=0pg1RtZzPoPDprTf8gX3TJm8XF5hKZ4pZV74UCe7"}},
	}, flow.TypeAPI, false, nosurfx.FakeCSRFTokenGenerator, ""), "should ignore Cloudflare cookies")
	require.EqualError(t, flow.EnsureCSRF(reg, &http.Request{
		Header: http.Header{"Cookie": {"__cflb=0pg1RtZzPoPDprTf8gX3TJm8XF5hKZ4pZV74UCe7; __cfruid=0pg1RtZzPoPDprTf8gX3TJm8XF5hKZ4pZV74UCe7; some_cookie=some_value"}},
	}, flow.TypeAPI, false, nosurfx.FakeCSRFTokenGenerator, ""), flow.ErrCookieHeaderNeedsBrowserFlow.Error(), "should error because of some_cookie")
	require.EqualError(t, flow.EnsureCSRF(reg, &http.Request{
		Header: http.Header{"Cookie": {"some_cookie=some_value"}},
	}, flow.TypeAPI, false, nosurfx.FakeCSRFTokenGenerator, ""), flow.ErrCookieHeaderNeedsBrowserFlow.Error(), "should error because of some_cookie")
	require.NoError(t, flow.EnsureCSRF(reg, &http.Request{}, flow.TypeAPI, false, nosurfx.FakeCSRFTokenGenerator, ""), "no cookie, no error")
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
