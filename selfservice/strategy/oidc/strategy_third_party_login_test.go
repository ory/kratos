// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/x/configx"
	"github.com/ory/x/httprouterx"
)

func TestThirdPartyLoginInit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Mock OIDC discovery server over plain HTTP. Dev mode is enabled so
	// the handler accepts HTTP issuers.
	mockOIDC := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			issuer := "http://" + r.Host
			fmt.Fprintf(w, `{
				"issuer": %q,
				"authorization_endpoint": %q,
				"token_endpoint": %q,
				"jwks_uri": %q,
				"code_challenge_methods_supported": ["S256"]
			}`,
				issuer,
				issuer+"/authorize",
				issuer+"/token",
				issuer+"/jwks",
			)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(mockOIDC.Close)
	mockOIDCIssuer := mockOIDC.URL // http://127.0.0.1:<port>

	conf, reg := pkg.NewFastRegistryWithMocks(t,
		configx.WithValues(map[string]any{
			"dev": true,
			config.ViperKeyIdentitySchemas: config.Schemas{
				{ID: "default", URL: "file://./stub/registration.schema.json"},
			},
			config.ViperKeyDefaultIdentitySchemaID: "default",
			config.HookStrategyKey(config.ViperKeySelfServiceRegistrationAfter, identity.CredentialsTypeOIDC.String()): []config.SelfServiceHook{{Name: "session"}},
		}),
	)

	returnTS := newReturnTS(t, reg)
	_ = newUI(t, reg)
	errTS := testhelpers.NewErrorTestServer(t, reg)

	routerP := httprouterx.NewTestRouterPublic(t)
	routerA := httprouterx.NewTestRouterAdminWithPrefix(t)
	ts, _ := testhelpers.NewKratosServerWithRouters(t, reg, routerP, routerA)

	viperSetProviderConfig(t, conf, oidc.Configuration{
		ID:           "test-provider",
		Provider:     "generic",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		IssuerURL:    mockOIDCIssuer,
		Mapper:       "file://./stub/oidc.hydra.jsonnet",
	})

	thirdPartyURL := ts.URL + oidc.RouteThirdPartyLoginInit

	// noRedirectClient follows all redirects within the test infrastructure.
	noRedirectClient := func(t *testing.T) *http.Client {
		return &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				t.Logf("Redirect: %s", req.URL.String())
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		}
	}

	// stopOnExternalRedirect stops following redirects once the target leaves
	// the test infrastructure servers.
	stopOnExternalRedirect := func(t *testing.T) *http.Client {
		return &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				for _, allowed := range []string{ts.URL, errTS.URL, returnTS.URL} {
					u, _ := url.Parse(allowed)
					if req.URL.Host == u.Host {
						return nil
					}
				}
				return http.ErrUseLastResponse
			},
		}
	}

	t.Run("case=should fail when iss is missing", func(t *testing.T) {
		res, err := noRedirectClient(t).Get(thirdPartyURL)
		require.NoError(t, err)
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		require.Contains(t, res.Request.URL.String(), errTS.URL, "%s", body)
		assert.Equal(t, int64(http.StatusBadRequest), gjson.GetBytes(body, "code").Int(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "reason").String(), "`iss` parameter is required", "%s", body)
	})

	t.Run("case=should fail when iss is not HTTPS in production mode", func(t *testing.T) {
		// Temporarily disable dev mode to test HTTPS enforcement.
		conf.MustSet(ctx, "dev", false)
		t.Cleanup(func() { conf.MustSet(ctx, "dev", true) })

		res, err := noRedirectClient(t).Get(thirdPartyURL + "?iss=http://example.com")
		require.NoError(t, err)
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		require.Contains(t, res.Request.URL.String(), errTS.URL, "%s", body)
		assert.Equal(t, int64(http.StatusBadRequest), gjson.GetBytes(body, "code").Int(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "reason").String(), "HTTPS", "%s", body)
	})

	t.Run("case=should fail when iss is not a valid URL", func(t *testing.T) {
		res, err := noRedirectClient(t).Get(thirdPartyURL + "?iss=not-a-url")
		require.NoError(t, err)
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		require.Contains(t, res.Request.URL.String(), errTS.URL, "%s", body)
		assert.Equal(t, int64(http.StatusBadRequest), gjson.GetBytes(body, "code").Int(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "reason").String(), "HTTPS", "%s", body)
	})

	t.Run("case=should fail when issuer is unknown", func(t *testing.T) {
		res, err := noRedirectClient(t).Get(thirdPartyURL + "?iss=https://unknown-issuer.example.com")
		require.NoError(t, err)
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		require.Contains(t, res.Request.URL.String(), errTS.URL, "%s", body)
		assert.Equal(t, int64(http.StatusNotFound), gjson.GetBytes(body, "code").Int(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "reason").String(), "No configured OpenID Connect provider", "%s", body)
	})

	t.Run("case=should fail when target_link_uri is not allowed", func(t *testing.T) {
		issuer := url.QueryEscape(mockOIDCIssuer)
		target := url.QueryEscape("https://evil.example.com/steal-session")
		res, err := noRedirectClient(t).Get(thirdPartyURL + "?iss=" + issuer + "&target_link_uri=" + target)
		require.NoError(t, err)
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		require.Contains(t, res.Request.URL.String(), errTS.URL, "%s", body)
		assert.Equal(t, int64(http.StatusBadRequest), gjson.GetBytes(body, "code").Int(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "reason").String(), "target_link_uri", "%s", body)
	})

	t.Run("case=should redirect to OIDC provider on valid request", func(t *testing.T) {
		issuer := url.QueryEscape(mockOIDCIssuer)
		client := stopOnExternalRedirect(t)
		res, err := client.Get(thirdPartyURL + "?iss=" + issuer)
		require.NoError(t, err)
		defer res.Body.Close()

		require.Equal(t, http.StatusSeeOther, res.StatusCode)
		location := res.Header.Get("Location")
		require.NotEmpty(t, location, "Expected Location header with auth redirect URL")

		locURL, err := url.Parse(location)
		require.NoError(t, err)

		// The auth redirect should point to the mock OIDC server's
		// authorization_endpoint (which is its own URL + /authorize).
		mockURL, _ := url.Parse(mockOIDCIssuer)
		assert.Equal(t, mockURL.Host, locURL.Host)
		assert.Equal(t, "/authorize", locURL.Path)
		assert.Equal(t, "test-client-id", locURL.Query().Get("client_id"))
		assert.Equal(t, "code", locURL.Query().Get("response_type"))
		assert.NotEmpty(t, locURL.Query().Get("state"))
	})

	t.Run("case=should pass login_hint to provider", func(t *testing.T) {
		issuer := url.QueryEscape(mockOIDCIssuer)
		client := stopOnExternalRedirect(t)
		res, err := client.Get(thirdPartyURL + "?iss=" + issuer + "&login_hint=user@example.com")
		require.NoError(t, err)
		defer res.Body.Close()

		require.Equal(t, http.StatusSeeOther, res.StatusCode)
		location := res.Header.Get("Location")
		require.NotEmpty(t, location)

		locURL, err := url.Parse(location)
		require.NoError(t, err)
		assert.Equal(t, "user@example.com", locURL.Query().Get("login_hint"))
	})

	t.Run("case=should accept valid target_link_uri", func(t *testing.T) {
		issuer := url.QueryEscape(mockOIDCIssuer)
		target := url.QueryEscape(returnTS.URL + "/after-login")
		client := stopOnExternalRedirect(t)
		res, err := client.Get(thirdPartyURL + "?iss=" + issuer + "&target_link_uri=" + target)
		require.NoError(t, err)
		defer res.Body.Close()

		require.Equal(t, http.StatusSeeOther, res.StatusCode)
		location := res.Header.Get("Location")
		require.NotEmpty(t, location)

		locURL, err := url.Parse(location)
		require.NoError(t, err)
		mockURL, _ := url.Parse(mockOIDCIssuer)
		assert.Equal(t, mockURL.Host, locURL.Host)
	})

	t.Run("case=should handle trailing slash in issuer", func(t *testing.T) {
		issuer := url.QueryEscape(mockOIDCIssuer + "/")
		client := stopOnExternalRedirect(t)
		res, err := client.Get(thirdPartyURL + "?iss=" + issuer)
		require.NoError(t, err)
		defer res.Body.Close()

		require.Equal(t, http.StatusSeeOther, res.StatusCode)
		location := res.Header.Get("Location")
		require.NotEmpty(t, location)

		locURL, err := url.Parse(location)
		require.NoError(t, err)
		mockURL, _ := url.Parse(mockOIDCIssuer)
		assert.Equal(t, mockURL.Host, locURL.Host)
	})

	t.Run("case=should work with POST method", func(t *testing.T) {
		client := stopOnExternalRedirect(t)
		res, err := client.PostForm(thirdPartyURL, url.Values{
			"iss":        {mockOIDCIssuer},
			"login_hint": {"user@example.com"},
		})
		require.NoError(t, err)
		defer res.Body.Close()

		require.Equal(t, http.StatusSeeOther, res.StatusCode)
		location := res.Header.Get("Location")
		require.NotEmpty(t, location)

		locURL, err := url.Parse(location)
		require.NoError(t, err)
		mockURL, _ := url.Parse(mockOIDCIssuer)
		assert.Equal(t, mockURL.Host, locURL.Host)
		assert.Equal(t, "user@example.com", locURL.Query().Get("login_hint"))
	})

	t.Run("case=should return 404 when OIDC strategy is disabled", func(t *testing.T) {
		baseKey := fmt.Sprintf("%s.%s", config.ViperKeySelfServiceStrategyConfig, identity.CredentialsTypeOIDC)
		conf.MustSet(ctx, baseKey+".enabled", false)
		t.Cleanup(func() {
			conf.MustSet(ctx, baseKey+".enabled", true)
		})

		res, err := noRedirectClient(t).Get(thirdPartyURL + "?iss=https://example.com")
		require.NoError(t, err)
		defer res.Body.Close()

		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})
}
