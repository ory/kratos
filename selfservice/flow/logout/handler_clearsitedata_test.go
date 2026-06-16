// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package logout_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/kratos/x"
)

func TestLogoutClearSiteData(t *testing.T) {
	ctx := context.Background()
	conf, reg := pkg.NewFastRegistryWithMocks(t)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/identity.schema.json")
	public, _, publicRouter, _ := testhelpers.NewKratosServerWithCSRFAndRouters(t, reg)
	publicRouter.GET("/session/browser/set", func(w http.ResponseWriter, r *http.Request) {
		testhelpers.MockSetSession(t, reg, conf)(w, r)
	})
	conf.MustSet(ctx, config.ViperKeySelfServiceLogoutBrowserDefaultReturnTo, public.URL+"/")

	// Resolve the submit URL and token for a fresh session, then issue the logout
	// without following the redirect so we can inspect the 303 response headers.
	logout := func(t *testing.T, forwardedProto string) *http.Response {
		hc := testhelpers.NewSessionClient(t, public.URL+"/session/browser/set")

		u, err := url.Parse(public.URL + "/self-service/logout/browser")
		require.NoError(t, err)
		body, res := testhelpers.HTTPRequestJSON(t, hc, "GET", u.String(), nil)
		require.EqualValues(t, http.StatusOK, res.StatusCode)
		logoutURL := gjson.GetBytes(body, "logout_url").String()
		require.NotEmpty(t, logoutURL)

		hc.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
		req, err := http.NewRequest("GET", logoutURL, nil)
		require.NoError(t, err)
		if forwardedProto != "" {
			req.Header.Set("X-Forwarded-Proto", forwardedProto)
		}
		out, err := hc.Do(req)
		require.NoError(t, err)
		t.Cleanup(func() { _ = out.Body.Close() })
		require.Equal(t, http.StatusSeeOther, out.StatusCode, "%s", x.MustReadAll(out.Body))
		return out
	}

	t.Run("case=flag disabled does not set header", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceLogoutClearBrowserData, false)
		res := logout(t, "https")
		assert.Empty(t, res.Header.Get("Clear-Site-Data"))
	})

	t.Run("case=flag enabled over https sets header", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceLogoutClearBrowserData, true)
		res := logout(t, "https")
		assert.Equal(t, `"cookies", "storage", "cache"`, res.Header.Get("Clear-Site-Data"))
	})

	t.Run("case=flag enabled over http does not set header", func(t *testing.T) {
		conf.MustSet(ctx, config.ViperKeySelfServiceLogoutClearBrowserData, true)
		res := logout(t, "http")
		assert.Empty(t, res.Header.Get("Clear-Site-Data"))
	})
}
