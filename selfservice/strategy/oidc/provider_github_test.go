// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"testing"
	"time"

	ghapi "github.com/google/go-github/v89/github"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/x/httpx"
)

type githubMockTransport struct {
	targetURL *url.URL
}

func (t *githubMockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqCopy := req.Clone(req.Context())
	reqCopy.URL.Scheme = t.targetURL.Scheme
	reqCopy.URL.Host = t.targetURL.Host
	return http.DefaultTransport.RoundTrip(reqCopy)
}

type mockGitHubRegistry struct {
	oidc.Dependencies
	client *retryablehttp.Client
}

func (m *mockGitHubRegistry) HTTPClient(ctx context.Context, opts ...httpx.ResilientOptions) *retryablehttp.Client {
	return m.client
}

// TestGitHubPublicRawClaimsAllowlist verifies that gitHubPublicRawClaims maps only
// the approved public GitHub profile fields, strictly excludes private/security data,
// preserves explicit zero/false/empty-string values, formats timestamps as UTC RFC3339,
// and omits absent or nil fields.
func TestGitHubPublicRawClaimsAllowlist(t *testing.T) {
	t.Parallel()

	intPtr := func(v int) *int { return &v }
	int64Ptr := func(v int64) *int64 { return &v }
	strPtr := func(v string) *string { return &v }
	boolPtr := func(v bool) *bool { return &v }

	createdAtTime, err := time.Parse(time.RFC3339, "2008-01-14T06:33:35+02:00")
	require.NoError(t, err)
	updatedAtTime, err := time.Parse(time.RFC3339, "2026-02-01T12:00:00Z")
	require.NoError(t, err)
	suspendedAtTime, err := time.Parse(time.RFC3339, "2026-01-01T00:00:00Z")
	require.NoError(t, err)

	// Fixture contains all SDK-modeled approved fields AND known forbidden fields
	fullUser := &ghapi.User{
		// Approved fields
		ID:                int64Ptr(1),
		Login:             strPtr("octocat"),
		NodeID:            strPtr("MDQ6VXNlcjE="),
		AvatarURL:         strPtr("https://github.example.com/images/error/octocat_happy.gif"),
		GravatarID:        strPtr(""),
		URL:               strPtr("https://api.github.example.com/users/octocat"),
		HTMLURL:           strPtr("https://github.example.com/octocat"),
		FollowersURL:      strPtr("https://api.github.example.com/users/octocat/followers"),
		FollowingURL:      strPtr("https://api.github.example.com/users/octocat/following{/other_user}"),
		GistsURL:          strPtr("https://api.github.example.com/users/octocat/gists{/gist_id}"),
		StarredURL:        strPtr("https://api.github.example.com/users/octocat/starred{/owner}{/repo}"),
		SubscriptionsURL:  strPtr("https://api.github.example.com/users/octocat/subscriptions"),
		OrganizationsURL:  strPtr("https://api.github.example.com/users/octocat/orgs"),
		ReposURL:          strPtr("https://api.github.example.com/users/octocat/repos"),
		EventsURL:         strPtr("https://api.github.example.com/users/octocat/events{/privacy}"),
		ReceivedEventsURL: strPtr("https://api.github.example.com/users/octocat/received_events"),
		Type:              strPtr("User"),
		UserViewType:      strPtr("public"),
		SiteAdmin:         boolPtr(false),
		Name:              strPtr("The Octocat"),
		Company:           strPtr("GitHub"),
		Blog:              strPtr("https://github.blog"),
		Location:          strPtr("San Francisco"),
		Hireable:          boolPtr(false),
		Bio:               strPtr("There once was..."),
		TwitterUsername:   strPtr("monatheoctocat"),
		PublicRepos:       intPtr(2),
		PublicGists:       intPtr(1),
		Followers:         intPtr(20),
		Following:         intPtr(0),
		CreatedAt:         &ghapi.Timestamp{Time: createdAtTime},
		UpdatedAt:         &ghapi.Timestamp{Time: updatedAtTime},

		// Forbidden / private / security fields that must NOT appear in RawClaims
		Email:                   strPtr("octocat@github.example.com"),
		NotificationEmail:       strPtr("octocat-notify@github.example.com"),
		TotalPrivateRepos:       int64Ptr(100),
		OwnedPrivateRepos:       int64Ptr(50),
		PrivateGists:            intPtr(8),
		DiskUsage:               intPtr(10000),
		Collaborators:           intPtr(8),
		TwoFactorAuthentication: boolPtr(true),
		Plan:                    &ghapi.Plan{Name: strPtr("enterprise")},
		BusinessPlus:            boolPtr(true),
		LdapDn:                  strPtr("cn=octocat,dc=example,dc=com"),
		SuspendedAt:             &ghapi.Timestamp{Time: suspendedAtTime},
	}

	raw := oidc.ExportGitHubPublicRawClaims(fullUser)
	require.NotNil(t, raw)

	expectedKeys := []string{
		"id", "login", "node_id", "avatar_url", "gravatar_id", "url", "html_url",
		"followers_url", "following_url", "gists_url", "starred_url", "subscriptions_url",
		"organizations_url", "repos_url", "events_url", "received_events_url", "type",
		"user_view_type", "site_admin", "name", "company", "blog", "location", "hireable",
		"bio", "twitter_username", "public_repos", "public_gists", "followers", "following",
		"created_at", "updated_at",
	}

	actualKeys := make([]string, 0, len(raw))
	for k := range raw {
		actualKeys = append(actualKeys, k)
	}
	slices.Sort(expectedKeys)
	slices.Sort(actualKeys)
	assert.Equal(t, expectedKeys, actualKeys, "exact allowlist key set match")

	// Representative values, explicit false, zero, empty string, RFC3339 timestamps
	assert.Equal(t, int64(1), raw["id"])
	assert.Equal(t, "octocat", raw["login"])
	assert.Equal(t, "", raw["gravatar_id"])
	assert.Equal(t, false, raw["site_admin"])
	assert.Equal(t, false, raw["hireable"])
	assert.Equal(t, 0, raw["following"])
	assert.Equal(t, "2008-01-14T04:33:35Z", raw["created_at"])
	assert.Equal(t, "2026-02-01T12:00:00Z", raw["updated_at"])

	// Assert forbidden keys explicitly absent
	forbiddenKeys := []string{
		"email", "notification_email", "total_private_repos", "owned_private_repos",
		"private_gists", "disk_usage", "collaborators", "two_factor_authentication",
		"plan", "business_plus", "ldap_dn", "suspended_at",
	}
	for _, fk := range forbiddenKeys {
		assert.NotContains(t, raw, fk)
	}

	// Nil created_at must not invent a zero date, and nil user must return nil
	assert.Nil(t, oidc.ExportGitHubPublicRawClaims(nil))

	partialUser := &ghapi.User{
		Login: strPtr("mona"),
	}
	partialRaw := oidc.ExportGitHubPublicRawClaims(partialUser)
	require.NotNil(t, partialRaw)
	assert.Equal(t, "mona", partialRaw["login"])
	assert.NotContains(t, partialRaw, "created_at")
	assert.NotContains(t, partialRaw, "id")
}

func TestGitHubClaimsRegression(t *testing.T) {
	t.Parallel()

	createdAtStr := "2008-01-14T04:33:35Z"
	updatedAtStr := "2026-02-01T12:00:00Z"

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/user":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":                        123456,
				"login":                     "test-octocat",
				"name":                      "Octo Cat",
				"avatar_url":                "https://avatars.example.com/u/123456",
				"html_url":                  "https://github.example.com/test-octocat",
				"blog":                      "https://example.com/blog",
				"created_at":                createdAtStr,
				"updated_at":                updatedAtStr,
				"bio":                       "fixture bio",
				"public_repos":              42,
				"total_private_repos":       99,
				"private_gists":             5,
				"two_factor_authentication": true,
			})
		case "/user/emails":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{
					"email":    "primary@example.com",
					"primary":  true,
					"verified": true,
				},
				{
					"email":    "secondary@example.com",
					"primary":  false,
					"verified": false,
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(mockServer.Close)

	serverURL, err := url.Parse(mockServer.URL)
	require.NoError(t, err)

	transport := &githubMockTransport{targetURL: serverURL}

	t.Run("provider=github", func(t *testing.T) {
		_, baseReg := pkg.NewFastRegistryWithMocks(t)
		client := retryablehttp.NewClient()
		client.HTTPClient = &http.Client{Transport: transport}
		reg := &mockGitHubRegistry{Dependencies: baseReg, client: client}

		p := oidc.NewProviderGitHub(&oidc.Configuration{
			ClientID:     "fake-github-client-id",
			ClientSecret: "fake-github-client-secret",
			Scope:        []string{"user:email"},
		}, reg).(*oidc.ProviderGitHub)

		token := &oauth2.Token{
			AccessToken: "fake-github-access-token",
		}
		token = token.WithExtra(map[string]any{
			"scope": "user:email",
		})

		claims, err := p.Claims(t.Context(), token, nil)
		require.NoError(t, err)
		require.NotNil(t, claims)

		// Assert normalized fields remain exact
		assert.Equal(t, "123456", claims.Subject)
		assert.Equal(t, "test-octocat", claims.Nickname)
		assert.Equal(t, "Octo Cat", claims.Name)
		assert.Equal(t, "https://example.com/blog", claims.Website)
		assert.Equal(t, "https://avatars.example.com/u/123456", claims.Picture)
		assert.Equal(t, "https://github.example.com/test-octocat", claims.Profile)
		assert.Equal(t, "primary@example.com", claims.Email)
		assert.True(t, bool(claims.EmailVerified))

		// Assert RawClaims populated with created_at and public fields, excluding private
		require.NotNil(t, claims.RawClaims)
		assert.Equal(t, "2008-01-14T04:33:35Z", claims.RawClaims["created_at"])
		assert.Equal(t, "fixture bio", claims.RawClaims["bio"])
		assert.Equal(t, 42, claims.RawClaims["public_repos"])
		assert.NotContains(t, claims.RawClaims, "total_private_repos")
		assert.NotContains(t, claims.RawClaims, "private_gists")
		assert.NotContains(t, claims.RawClaims, "two_factor_authentication")
		assert.NotContains(t, claims.RawClaims, "email")
	})

	t.Run("provider=github-app", func(t *testing.T) {
		_, baseReg := pkg.NewFastRegistryWithMocks(t)
		client := retryablehttp.NewClient()
		client.HTTPClient = &http.Client{Transport: transport}
		reg := &mockGitHubRegistry{Dependencies: baseReg, client: client}

		p := oidc.NewProviderGitHubApp(&oidc.Configuration{
			ClientID:     "fake-github-app-client-id",
			ClientSecret: "fake-github-app-client-secret",
		}, reg).(*oidc.ProviderGitHubApp)

		token := &oauth2.Token{
			AccessToken: "fake-github-app-access-token",
		}

		claims, err := p.Claims(t.Context(), token, nil)
		require.NoError(t, err)
		require.NotNil(t, claims)

		// Assert normalized fields remain exact
		assert.Equal(t, "123456", claims.Subject)
		assert.Equal(t, "test-octocat", claims.Nickname)
		assert.Equal(t, "Octo Cat", claims.Name)
		assert.Equal(t, "https://example.com/blog", claims.Website)
		assert.Equal(t, "https://avatars.example.com/u/123456", claims.Picture)
		assert.Equal(t, "https://github.example.com/test-octocat", claims.Profile)
		assert.Equal(t, "primary@example.com", claims.Email)
		assert.True(t, bool(claims.EmailVerified))

		// Assert RawClaims populated with created_at and public fields, excluding private
		require.NotNil(t, claims.RawClaims)
		assert.Equal(t, "2008-01-14T04:33:35Z", claims.RawClaims["created_at"])
		assert.Equal(t, "fixture bio", claims.RawClaims["bio"])
		assert.Equal(t, 42, claims.RawClaims["public_repos"])
		assert.NotContains(t, claims.RawClaims, "total_private_repos")
		assert.NotContains(t, claims.RawClaims, "private_gists")
		assert.NotContains(t, claims.RawClaims, "two_factor_authentication")
		assert.NotContains(t, claims.RawClaims, "email")
	})
}
