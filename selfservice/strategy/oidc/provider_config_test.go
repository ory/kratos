// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/x/configx"
)

func TestConfig(t *testing.T) {
	_, reg := pkg.NewFastRegistryWithMocks(t, configx.WithValue(
		config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeOIDC)+".config",
		map[string]any{"providers": []map[string]any{{"provider": "generic"}}}))

	s := oidc.NewStrategy(reg)
	collection, err := s.Config(t.Context())
	require.NoError(t, err)

	require.Len(t, collection.Providers, 1)
	assert.Equal(t, "generic", collection.Providers[0].Provider)
}

func TestConfigSecretResolution(t *testing.T) {
	t.Parallel()

	providersKey := config.ViperKeySelfServiceStrategyConfig + "." + string(identity.CredentialsTypeOIDC) + ".config"

	writeSecret := func(t *testing.T, contents string) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), "client_secret")
		require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))
		return path
	}

	configFor := func(secret string) map[string]any {
		return map[string]any{"providers": []map[string]any{
			{"provider": "generic", "id": "acme", "client_secret": secret},
		}}
	}

	newStrategy := func(t *testing.T, providers map[string]any, allowSecretURIs bool) *oidc.Strategy {
		t.Helper()
		_, reg := pkg.NewFastRegistryWithMocks(t,
			configx.WithValue(providersKey, providers),
			configx.WithValue(config.ViperKeySecurityAllowSecretURIsInOIDCConfig, allowSecretURIs))
		return oidc.NewStrategy(reg)
	}

	clientSecretFor := func(t *testing.T, providers map[string]any) string {
		t.Helper()
		p, err := newStrategy(t, providers, true).Provider(t.Context(), "acme")
		require.NoError(t, err)
		return p.Config().ClientSecret
	}

	t.Run("resolves file:// to the file contents", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "s3cr3t", clientSecretFor(t, configFor("file://"+writeSecret(t, "s3cr3t"))))
	})

	t.Run("trims exactly one trailing newline from a file secret", func(t *testing.T) {
		t.Parallel()
		for contents, want := range map[string]string{
			"s3cr3t\n":   "s3cr3t",
			"s3cr3t\r\n": "s3cr3t",
			"s3cr3t\n\n": "s3cr3t\n",
			"s3cr3t \n":  "s3cr3t ",
			"  s3cr3t\n": "  s3cr3t",
		} {
			assert.Equal(t, want, clientSecretFor(t, configFor("file://"+writeSecret(t, contents))), "%q", contents)
		}
	})

	t.Run("resolves base64:// to the exact decoded bytes", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "s3cr3t\n", clientSecretFor(t, configFor(
			"base64://"+base64.StdEncoding.EncodeToString([]byte("s3cr3t\n")))))
	})

	t.Run("leaves an inline secret untouched", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "s3cr3t", clientSecretFor(t, configFor("s3cr3t")))
	})

	t.Run("leaves an unsupported scheme untouched", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "https://example.com/secret", clientSecretFor(t, configFor("https://example.com/secret")))
	})

	t.Run("resolves apple_private_key", func(t *testing.T) {
		t.Parallel()

		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)
		der, err := x509.MarshalPKCS8PrivateKey(key)
		require.NoError(t, err)
		pemKey := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))

		s := newStrategy(t, map[string]any{"providers": []map[string]any{{
			"provider":             "apple",
			"id":                   "apple",
			"client_id":            "client",
			"apple_team_id":        "team",
			"apple_private_key_id": "key-id",
			"apple_private_key":    "file://" + writeSecret(t, pemKey),
		}}}, true)

		p, err := s.Provider(t.Context(), "apple")
		require.NoError(t, err)
		assert.Equal(t, strings.TrimSuffix(pemKey, "\n"), p.Config().PrivateKey)
	})

	t.Run("passes URIs through verbatim by default", func(t *testing.T) {
		t.Parallel()

		secret := "file://" + writeSecret(t, "s3cr3t")
		_, reg := pkg.NewFastRegistryWithMocks(t, configx.WithValue(providersKey, configFor(secret)))

		p, err := oidc.NewStrategy(reg).Provider(t.Context(), "acme")
		require.NoError(t, err)
		assert.Equal(t, secret, p.Config().ClientSecret)
	})

	t.Run("does not mutate the shared configuration collection", func(t *testing.T) {
		t.Parallel()

		secret := "file://" + writeSecret(t, "s3cr3t")
		s := newStrategy(t, configFor(secret), true)

		p, err := s.Provider(t.Context(), "acme")
		require.NoError(t, err)
		assert.Equal(t, "s3cr3t", p.Config().ClientSecret)

		collection, err := s.Config(t.Context())
		require.NoError(t, err)
		assert.Equal(t, secret, collection.Providers[0].ClientSecret)
	})

	t.Run("a broken provider does not affect enumeration or other providers", func(t *testing.T) {
		t.Parallel()

		s := newStrategy(t, map[string]any{"providers": []map[string]any{
			{"provider": "generic", "id": "acme", "client_secret": "s3cr3t"},
			{"provider": "generic", "id": "broken", "client_secret": "file://" + filepath.Join(t.TempDir(), "does-not-exist")},
		}}, true)

		collection, err := s.Config(t.Context())
		require.NoError(t, err)
		assert.Len(t, collection.Providers, 2)

		p, err := s.Provider(t.Context(), "acme")
		require.NoError(t, err)
		assert.Equal(t, "s3cr3t", p.Config().ClientSecret)

		_, err = s.Provider(t.Context(), "broken")
		require.Error(t, err)
	})

	t.Run("errors when the file is missing without leaking the path", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "does-not-exist")
		s := newStrategy(t, configFor("file://"+path), true)

		_, err := s.Provider(t.Context(), "acme")
		var he *herodot.DefaultError
		require.ErrorAs(t, err, &he)
		assert.Contains(t, he.Reason(), "acme", "%+v", err)
		assert.Contains(t, he.Reason(), "client_secret", "%+v", err)
		assert.NotContains(t, he.Reason(), path, "%+v", err)
	})

	t.Run("errors when the resolved secret is empty", func(t *testing.T) {
		t.Parallel()

		s := newStrategy(t, configFor("file://"+writeSecret(t, "\n")), true)

		_, err := s.Provider(t.Context(), "acme")
		var he *herodot.DefaultError
		require.ErrorAs(t, err, &he)
		assert.Contains(t, he.Reason(), "acme", "%+v", err)
	})

	t.Run("does not leak the secret into the error", func(t *testing.T) {
		t.Parallel()

		s := newStrategy(t, configFor("base64://not-valid-base64!!"), true)

		_, err := s.Provider(t.Context(), "acme")
		var he *herodot.DefaultError
		require.ErrorAs(t, err, &he)
		assert.NotContains(t, he.Reason(), "not-valid-base64", "%+v", err)
		assert.NotContains(t, he.DebugField, "not-valid-base64", "%+v", err)
	})
}

func TestConfiguration_AALForClaims(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		config oidc.Configuration
		claims *oidc.Claims
		want   identity.AuthenticatorAssuranceLevel
	}{
		{
			name:   "nil claims default to aal1",
			config: oidc.Configuration{AAL2ACRValues: []string{"urn:mfa"}, AAL2AMRValues: []string{"mfa"}},
			claims: nil,
			want:   identity.AuthenticatorAssuranceLevel1,
		},
		{
			name:   "empty config with claims stays aal1",
			config: oidc.Configuration{},
			claims: &oidc.Claims{ACR: "urn:mfa", AMR: []string{"mfa", "pwd"}},
			want:   identity.AuthenticatorAssuranceLevel1,
		},
		{
			name:   "acr match elevates to aal2",
			config: oidc.Configuration{AAL2ACRValues: []string{"urn:mfa", "urn:strong"}},
			claims: &oidc.Claims{ACR: "urn:mfa"},
			want:   identity.AuthenticatorAssuranceLevel2,
		},
		{
			name:   "acr mismatch stays aal1",
			config: oidc.Configuration{AAL2ACRValues: []string{"urn:mfa"}},
			claims: &oidc.Claims{ACR: "urn:basic"},
			want:   identity.AuthenticatorAssuranceLevel1,
		},
		{
			name:   "empty acr does not accidentally match empty configured value",
			config: oidc.Configuration{AAL2ACRValues: []string{""}},
			claims: &oidc.Claims{ACR: ""},
			want:   identity.AuthenticatorAssuranceLevel1,
		},
		{
			name:   "amr match elevates to aal2",
			config: oidc.Configuration{AAL2AMRValues: []string{"mfa"}},
			claims: &oidc.Claims{AMR: []string{"pwd", "mfa"}},
			want:   identity.AuthenticatorAssuranceLevel2,
		},
		{
			name:   "amr mismatch stays aal1",
			config: oidc.Configuration{AAL2AMRValues: []string{"mfa"}},
			claims: &oidc.Claims{AMR: []string{"pwd"}},
			want:   identity.AuthenticatorAssuranceLevel1,
		},
		{
			name:   "any configured amr value is sufficient",
			config: oidc.Configuration{AAL2AMRValues: []string{"otp", "hwk", "mfa"}},
			claims: &oidc.Claims{AMR: []string{"pwd", "hwk"}},
			want:   identity.AuthenticatorAssuranceLevel2,
		},
		{
			name:   "both acr and amr configured, acr matches",
			config: oidc.Configuration{AAL2ACRValues: []string{"urn:mfa"}, AAL2AMRValues: []string{"mfa"}},
			claims: &oidc.Claims{ACR: "urn:mfa", AMR: []string{"pwd"}},
			want:   identity.AuthenticatorAssuranceLevel2,
		},
		{
			name:   "both acr and amr configured, amr matches",
			config: oidc.Configuration{AAL2ACRValues: []string{"urn:mfa"}, AAL2AMRValues: []string{"mfa"}},
			claims: &oidc.Claims{ACR: "urn:basic", AMR: []string{"pwd", "mfa"}},
			want:   identity.AuthenticatorAssuranceLevel2,
		},
		{
			name:   "both configured, neither matches",
			config: oidc.Configuration{AAL2ACRValues: []string{"urn:mfa"}, AAL2AMRValues: []string{"mfa"}},
			claims: &oidc.Claims{ACR: "urn:basic", AMR: []string{"pwd"}},
			want:   identity.AuthenticatorAssuranceLevel1,
		},
	} {
		t.Run("case="+tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, tc.config.AALForClaims(tc.claims))
		})
	}
}
