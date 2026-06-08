// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login_test

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gofrs/uuid"

	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/x/clock"
	"github.com/ory/x/jsonx"
	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/pkg"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/x"
)

func TestFakeFlow(t *testing.T) {
	t.Parallel()
	var r login.Flow
	require.NoError(t, faker.FakeData(&r))

	assert.Equal(t, uuid.Nil, r.ID)
	assert.NotEmpty(t, r.IssuedAt)
	assert.NotEmpty(t, r.ExpiresAt)
	assert.NotEmpty(t, r.RequestURL)
	assert.NotEmpty(t, r.Active)
	assert.NotNil(t, r.UI)
}

func TestNewFlow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	conf, reg := pkg.NewFastRegistryWithMocks(t)

	t.Run("captures request base URL from context at init", func(t *testing.T) {
		// A proxy-aware middleware (cloud CaptureOriginalBaseURLMiddleware)
		// stashes the validated customer-facing base URL on the request
		// context at flow init. NewFlow must persist it on the flow so the
		// OIDC/SAML provider submit can read it back.
		req := (&http.Request{URL: &url.URL{Path: "/"}, Host: "slug.projects.oryapis.com"}).
			WithContext(x.WithBaseURL(ctx, urlx.ParseOrPanic("http://localhost:4000")))
		r, err := login.NewFlow(reg, req, flow.TypeBrowser)
		require.NoError(t, err)
		assert.Equal(t, "http://localhost:4000", flow.GetRequestBaseURL(r))

		// No captured base URL → nothing persisted (plain oryapis traffic).
		r, err = login.NewFlow(reg, &http.Request{URL: &url.URL{Path: "/"}, Host: "slug.projects.oryapis.com"}, flow.TypeBrowser)
		require.NoError(t, err)
		assert.Empty(t, flow.GetRequestBaseURL(r))
	})

	t.Run("type=aal", func(t *testing.T) {
		r, err := login.NewFlow(reg, &http.Request{URL: &url.URL{Path: "/", RawQuery: "aal=aal2&refresh=true"}, Host: "ory.sh"}, flow.TypeBrowser)
		require.NoError(t, err)
		assert.True(t, r.Refresh)
		assert.Equal(t, identity.AuthenticatorAssuranceLevel2, r.RequestedAAL)

		r, err = login.NewFlow(reg, &http.Request{URL: &url.URL{Path: "/", RawQuery: "refresh=true"}, Host: "ory.sh"}, flow.TypeBrowser)
		require.NoError(t, err)
		assert.True(t, r.Refresh)
		assert.Equal(t, identity.AuthenticatorAssuranceLevel1, r.RequestedAAL)
	})

	t.Run("type=refresh", func(t *testing.T) {
		t.Run("case=refresh accepts any truthy value", func(t *testing.T) {
			parameters := []string{"true", "True", "1"}

			for _, refresh := range parameters {
				r, err := login.NewFlow(reg, &http.Request{URL: &url.URL{Path: "/", RawQuery: fmt.Sprintf("refresh=%v", refresh)}, Host: "ory.sh"}, flow.TypeBrowser)
				require.NoError(t, err)
				assert.True(t, r.Refresh)
			}
		})

		t.Run("case=refresh silently ignores invalid values", func(t *testing.T) {
			r, err := login.NewFlow(reg, &http.Request{URL: &url.URL{Path: "/", RawQuery: "refresh=foo"}, Host: "ory.sh"}, flow.TypeBrowser)
			require.NoError(t, err)
			assert.False(t, r.Refresh)
		})
	})

	t.Run("type=return_to", func(t *testing.T) {
		_, err := login.NewFlow(reg, &http.Request{URL: &url.URL{Path: "/", RawQuery: "return_to=https://not-allowed/foobar"}, Host: "ory.sh"}, flow.TypeBrowser)
		require.Error(t, err)

		_, err = login.NewFlow(reg, &http.Request{URL: &url.URL{Path: "/", RawQuery: "return_to=" + urlx.AppendPaths(conf.SelfPublicURL(ctx), "/self-service/login/browser").String()}, Host: "ory.sh"}, flow.TypeBrowser)
		require.NoError(t, err)
	})

	t.Run("type=browser", func(t *testing.T) {
		t.Run("case=regular flow creation without a session", func(t *testing.T) {
			r, err := login.NewFlow(reg, &http.Request{
				URL:  urlx.ParseOrPanic("/"),
				Host: "ory.sh", TLS: &tls.ConnectionState{},
			}, flow.TypeBrowser)

			require.NoError(t, err)
			assert.True(t, r.ExpiresAt.After(r.IssuedAt))
			assert.Equal(t, flow.TypeBrowser, r.Type)
			assert.False(t, r.Refresh)
			assert.Equal(t, "https://ory.sh/", r.RequestURL)
		})

		t.Run("case=regular flow creation", func(t *testing.T) {
			r, err := login.NewFlow(reg, &http.Request{
				URL:  urlx.ParseOrPanic("https://ory.sh/"),
				Host: "ory.sh",
			}, flow.TypeBrowser)

			require.NoError(t, err)
			assert.Equal(t, "https://ory.sh/", r.RequestURL)
		})
	})

	t.Run("type=api", func(t *testing.T) {
		t.Run("case=flow with refresh", func(t *testing.T) {
			r, err := login.NewFlow(reg, &http.Request{
				URL:  urlx.ParseOrPanic("/?refresh=true"),
				Host: "ory.sh",
			}, flow.TypeAPI)

			require.NoError(t, err)
			assert.True(t, r.ExpiresAt.After(r.IssuedAt))
			assert.Equal(t, flow.TypeAPI, r.Type)
			assert.True(t, r.Refresh)
			assert.Equal(t, "http://ory.sh/?refresh=true", r.RequestURL)
		})

		t.Run("case=flow without refresh", func(t *testing.T) {
			r, err := login.NewFlow(reg, &http.Request{
				URL:  urlx.ParseOrPanic("/"),
				Host: "ory.sh",
			}, flow.TypeAPI)

			require.NoError(t, err)
			assert.True(t, r.ExpiresAt.After(r.IssuedAt))
			assert.Equal(t, flow.TypeAPI, r.Type)
			assert.False(t, r.Refresh)
			assert.Equal(t, "http://ory.sh/", r.RequestURL)
		})
	})

	t.Run("should parse login_challenge when Hydra is configured", func(t *testing.T) {
		_, err := login.NewFlow(reg, &http.Request{URL: urlx.ParseOrPanic("https://ory.sh/?login_challenge=badee1"), Host: "ory.sh"}, flow.TypeBrowser)
		require.Error(t, err)

		conf.MustSet(ctx, config.ViperKeyOAuth2ProviderURL, "https://hydra")

		r, err := login.NewFlow(reg, &http.Request{URL: urlx.ParseOrPanic("https://ory.sh/?login_challenge=8aadcb8fc1334186a84c4da9813356d9"), Host: "ory.sh"}, flow.TypeBrowser)
		require.NoError(t, err)
		assert.Equal(t, "8aadcb8fc1334186a84c4da9813356d9", string(r.OAuth2LoginChallenge))
	})
}

func TestFlow(t *testing.T) {
	t.Parallel()
	r := &login.Flow{ID: x.NewUUID()}
	assert.Equal(t, r.ID, r.GetID())

	t.Run("case=expired", func(t *testing.T) {
		for _, tc := range []struct {
			r     *login.Flow
			valid bool
		}{
			{
				r:     &login.Flow{ExpiresAt: time.Now().Add(time.Hour), IssuedAt: time.Now().Add(-time.Minute)},
				valid: true,
			},
			{r: &login.Flow{ExpiresAt: time.Now().Add(-time.Hour), IssuedAt: time.Now().Add(-time.Minute)}},
		} {
			if tc.valid {
				require.NoError(t, tc.r.Valid(clock.New()))
			} else {
				require.Error(t, tc.r.Valid(clock.New()))
			}
		}
	})
}

func TestGetType(t *testing.T) {
	t.Parallel()
	for _, ft := range []flow.Type{
		flow.TypeAPI,
		flow.TypeBrowser,
	} {
		t.Run(fmt.Sprintf("case=%s", ft), func(t *testing.T) {
			r := &login.Flow{Type: ft}
			assert.Equal(t, ft, r.GetType())
		})
	}
}

func TestGetRequestURL(t *testing.T) {
	t.Parallel()
	expectedURL := "http://foo/bar/baz"
	f := &login.Flow{RequestURL: expectedURL}
	assert.Equal(t, expectedURL, f.GetRequestURL())
}

func TestFlowEncodeJSON(t *testing.T) {
	t.Parallel()
	assert.EqualValues(t, "", gjson.Get(jsonx.TestMarshalJSONString(t, &login.Flow{RequestURL: "https://foo.bar?foo=bar"}), "return_to").String())
	assert.EqualValues(t, "/bar", gjson.Get(jsonx.TestMarshalJSONString(t, &login.Flow{RequestURL: "https://foo.bar?return_to=/bar"}), "return_to").String())
	assert.EqualValues(t, "/bar", gjson.Get(jsonx.TestMarshalJSONString(t, login.Flow{RequestURL: "https://foo.bar?return_to=/bar"}), "return_to").String())
}

func TestFlowTestContextNotUnmarshalable(t *testing.T) {
	t.Parallel()

	// test_context is a derived field populated from internal_context during
	// MarshalJSON. It must not be settable via incoming JSON, otherwise an
	// attacker who can submit a Flow blob could forge a debug payload.
	raw := []byte(`{
		"id": "00000000-0000-0000-0000-000000000000",
		"test_context": {
			"provider_id": "forged",
			"debug_payload": {"id_token_claims": {"sub": "forged"}}
		}
	}`)

	var f login.Flow
	require.NoError(t, json.Unmarshal(raw, &f))
	assert.Nil(t, f.TestContext, "test_context must not be populated from JSON input")
}

func TestFlowDontOverrideReturnTo(t *testing.T) {
	t.Parallel()
	f := &login.Flow{ReturnTo: "/foo"}
	f.SetReturnTo()
	assert.Equal(t, "/foo", f.ReturnTo)

	f = &login.Flow{RequestURL: "https://foo.bar?return_to=/bar"}
	f.SetReturnTo()
	assert.Equal(t, "/bar", f.ReturnTo)
}

func TestDuplicateCredentials(t *testing.T) {
	t.Parallel()
	t.Run("case=returns previous data", func(t *testing.T) {
		t.Parallel()
		f := new(login.Flow)
		dc := flow.DuplicateCredentialsData{
			CredentialsType:     "foo",
			CredentialsConfig:   sqlxx.JSONRawMessage(`{"bar":"baz"}`),
			DuplicateIdentifier: "bar",
		}

		require.NoError(t, flow.SetDuplicateCredentials(f, dc))
		actual, err := flow.DuplicateCredentials(f)
		require.NoError(t, err)
		assert.Equal(t, dc, *actual)
	})

	t.Run("case=returns nil data", func(t *testing.T) {
		t.Parallel()
		f := new(login.Flow)
		actual, err := flow.DuplicateCredentials(f)
		require.NoError(t, err)
		assert.Nil(t, actual)
	})
}
