// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/selfservice/strategy/oidc"
)

func TestConfig(t *testing.T) {
	ctx := context.Background()
	conf, reg := pkg.NewFastRegistryWithMocks(t)

	var c map[string]interface{}
	require.NoError(t, json.NewDecoder(
		bytes.NewBufferString(`{"config":{"providers": [{"provider": "generic"}]}}`)).Decode(&c))
	conf.MustSet(ctx, config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeOIDC), c)

	s := oidc.NewStrategy(reg)
	collection, err := s.Config(ctx)
	require.NoError(t, err)

	require.Len(t, collection.Providers, 1)
	assert.Equal(t, "generic", collection.Providers[0].Provider)
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
