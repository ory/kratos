// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/configx"
	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/selfservice/strategy/oidc"
)

type testUpdateProvider struct {
	config *oidc.Configuration
}

func (p *testUpdateProvider) Config() *oidc.Configuration {
	return p.config
}

func TestEvaluateClaimsMapper(t *testing.T) {
	_, reg := pkg.NewFastRegistryWithMocks(t)
	s := oidc.NewStrategy(reg)
	ctx := context.Background()

	provider := &testUpdateProvider{config: &oidc.Configuration{
		ID:       "test",
		Provider: "generic",
		Mapper:   "file://./stub/oidc.hydra.jsonnet",
	}}

	claims := &oidc.Claims{
		Subject: "alice@example.com",
		Email:   "alice@example.com",
		Website: "https://example.com",
		Picture: "https://example.com/alice.png",
		RawClaims: map[string]any{
			"groups": []any{"admin", "user"},
		},
	}

	t.Run("case=without current identity", func(t *testing.T) {
		t.Parallel()
		evaluated, claimsJSON, err := s.EvaluateClaimsMapper(ctx, claims, provider, nil)
		require.NoError(t, err)
		assert.NotEmpty(t, evaluated)
		assert.NotEmpty(t, claimsJSON)
		assert.Contains(t, evaluated, "alice@example.com")
		assert.Contains(t, evaluated, "https://example.com")
	})

	t.Run("case=with current identity", func(t *testing.T) {
		t.Parallel()
		currentIdentity := &identity.Identity{
			Traits:         identity.Traits(`{"subject":"alice@example.com","website":"https://old.example.com"}`),
			MetadataPublic: []byte(`{"picture":"https://example.com/old.png"}`),
			MetadataAdmin:  []byte(`{}`),
		}

		providerWithIdentity := &testUpdateProvider{config: &oidc.Configuration{
			ID:       "test",
			Provider: "generic",
			Mapper:   "file://./stub/oidc.update-identity.jsonnet",
		}}

		evaluated, _, err := s.EvaluateClaimsMapper(ctx, claims, providerWithIdentity, currentIdentity)
		require.NoError(t, err)
		assert.NotEmpty(t, evaluated)
		assert.Contains(t, evaluated, "alice@example.com")
	})
}

func TestUpdateIdentityFromClaims(t *testing.T) {
	_, reg := pkg.NewFastRegistryWithMocks(t, configx.WithValues(map[string]any{
		config.ViperKeyIdentitySchemas: config.Schemas{
			{ID: "default", URL: "file://./stub/registration.schema.json"},
		},
		config.ViperKeyDefaultIdentitySchemaID: "default",
	}))
	s := oidc.NewStrategy(reg)
	ctx := context.Background()

	provider := &testUpdateProvider{config: &oidc.Configuration{
		ID:       "test",
		Provider: "generic",
		Mapper:   "file://./stub/oidc.hydra.jsonnet",
	}}

	t.Run("case=detects changed traits", func(t *testing.T) {
		t.Parallel()
		i := &identity.Identity{
			SchemaID:       "default",
			Traits:         identity.Traits(`{"subject":"alice@example.com","website":"https://old.example.com"}`),
			MetadataPublic: []byte(`{}`),
			MetadataAdmin:  []byte(`{}`),
		}

		claims := &oidc.Claims{
			Subject: "alice@example.com",
			Website: "https://new.example.com",
			RawClaims: map[string]any{
				"groups": []any{"admin"},
			},
		}

		changed, err := s.UpdateIdentityFromClaims(ctx, claims, provider, i)
		require.NoError(t, err)
		assert.True(t, changed, "identity should be marked as changed")

		// The mapper keeps the existing website, so only the new group is added.
		assert.Contains(t, string(i.Traits), "https://old.example.com", "existing website should be preserved")
		assert.Contains(t, string(i.Traits), "admin")
	})

	t.Run("case=detects changed metadata", func(t *testing.T) {
		t.Parallel()
		i := &identity.Identity{
			SchemaID:       "default",
			Traits:         identity.Traits(`{"subject":"alice@example.com"}`),
			MetadataPublic: []byte(`{}`),
			MetadataAdmin:  []byte(`{}`),
		}

		claims := &oidc.Claims{
			Subject:     "alice@example.com",
			Picture:     "https://example.com/new-pic.png",
			PhoneNumber: "+1234567890",
			RawClaims:   map[string]any{"_placeholder": true},
		}

		changed, err := s.UpdateIdentityFromClaims(ctx, claims, provider, i)
		require.NoError(t, err)
		assert.True(t, changed, "identity should be marked as changed due to metadata")

		assert.Contains(t, string(i.MetadataPublic), "new-pic.png")
		assert.Contains(t, string(i.MetadataAdmin), "+1234567890")
	})

	t.Run("case=no change when claims match existing identity", func(t *testing.T) {
		t.Parallel()
		i := &identity.Identity{
			SchemaID:       "default",
			Traits:         identity.Traits(`{"subject":"alice@example.com","website":"https://example.com"}`),
			MetadataPublic: []byte(`{"picture":"https://example.com/alice.png"}`),
			MetadataAdmin:  []byte(`{"phone_number":"+1234567890"}`),
		}

		claims := &oidc.Claims{
			Subject:     "alice@example.com",
			Website:     "https://example.com",
			Picture:     "https://example.com/alice.png",
			PhoneNumber: "+1234567890",
			RawClaims:   map[string]any{"_placeholder": true},
		}

		changed, err := s.UpdateIdentityFromClaims(ctx, claims, provider, i)
		require.NoError(t, err)
		assert.False(t, changed, "identity should not be marked as changed when claims match")
	})

	t.Run("case=does not touch credentials", func(t *testing.T) {
		t.Parallel()
		originalCreds := map[identity.CredentialsType]identity.Credentials{
			identity.CredentialsTypeOIDC: {
				Type:   identity.CredentialsTypeOIDC,
				Config: []byte(`{"providers":[{"subject":"alice","provider":"test"}]}`),
			},
		}
		i := &identity.Identity{
			Traits:         identity.Traits(`{"subject":"alice@example.com"}`),
			MetadataPublic: []byte(`{}`),
			MetadataAdmin:  []byte(`{}`),
			Credentials:    originalCreds,
		}

		claims := &oidc.Claims{
			Subject:   "alice@example.com",
			Website:   "https://new.example.com",
			RawClaims: map[string]any{"_placeholder": true},
		}

		_, err := s.UpdateIdentityFromClaims(ctx, claims, provider, i)
		require.NoError(t, err)

		// Credentials must remain unchanged.
		assert.Equal(t, originalCreds, i.Credentials)
	})

	t.Run("case=identity context available in jsonnet", func(t *testing.T) {
		t.Parallel()
		providerWithIdentityMapper := &testUpdateProvider{config: &oidc.Configuration{
			ID:       "test",
			Provider: "generic",
			Mapper:   "file://./stub/oidc.update-identity.jsonnet",
		}}

		i := &identity.Identity{
			SchemaID:       "default",
			Traits:         identity.Traits(`{"subject":"alice@example.com","website":"https://old.example.com"}`),
			MetadataPublic: []byte(`{"existing_field":"keep-me","picture":"https://example.com/old.png"}`),
			MetadataAdmin:  []byte(`{}`),
		}

		claims := &oidc.Claims{
			Subject:   "alice@example.com",
			Website:   "https://new.example.com",
			Picture:   "https://example.com/new.png",
			RawClaims: map[string]any{"_placeholder": true},
		}

		changed, err := s.UpdateIdentityFromClaims(ctx, claims, providerWithIdentityMapper, i)
		require.NoError(t, err)
		assert.True(t, changed)

		// The mapper should preserve existing_field from the current identity's metadata.
		assert.Contains(t, string(i.MetadataPublic), "keep-me")
		// The mapper should update picture from the new claims.
		assert.Contains(t, string(i.MetadataPublic), "new.png")
	})

	t.Run("case=mapper that omits traits preserves existing via merge", func(t *testing.T) {
		t.Parallel()
		providerMissingSubject := &testUpdateProvider{config: &oidc.Configuration{
			ID:       "test",
			Provider: "generic",
			// This mapper outputs website but omits the required "subject" trait.
			Mapper: "file://./stub/oidc.missing-subject.jsonnet",
		}}

		i := &identity.Identity{
			SchemaID:       "default",
			Traits:         identity.Traits(`{"subject":"alice@example.com","website":"https://example.com"}`),
			MetadataPublic: []byte(`{}`),
			MetadataAdmin:  []byte(`{}`),
		}

		claims := &oidc.Claims{
			Subject:   "alice@example.com",
			Website:   "https://new.example.com",
			RawClaims: map[string]any{"_placeholder": true},
		}

		// The mapper omits "subject", but merge preserves it from the existing identity.
		changed, err := s.UpdateIdentityFromClaims(ctx, claims, providerMissingSubject, i)
		require.NoError(t, err)
		assert.True(t, changed, "website changed")
		assert.Contains(t, string(i.Traits), "alice@example.com", "subject should be preserved")
		assert.Contains(t, string(i.Traits), "https://new.example.com", "website should be updated")
	})

	t.Run("case=nil metadata fields are coalesced to empty objects", func(t *testing.T) {
		t.Parallel()
		// Identity with nil metadata — simulates an identity that was created
		// without metadata (e.g., password registration).
		i := &identity.Identity{
			SchemaID:       "default",
			Traits:         identity.Traits(`{"subject":"alice@example.com"}`),
			MetadataPublic: nil,
			MetadataAdmin:  nil,
		}

		providerWithIdentity := &testUpdateProvider{config: &oidc.Configuration{
			ID:       "test",
			Provider: "generic",
			Mapper:   "file://./stub/oidc.update-identity.jsonnet",
		}}

		claims := &oidc.Claims{
			Subject:   "alice@example.com",
			Website:   "https://example.com",
			RawClaims: map[string]any{"_placeholder": true},
		}

		// Should not crash even though metadata fields are nil — the coalesce
		// function converts nil to {} before passing to Jsonnet.
		_, err := s.UpdateIdentityFromClaims(ctx, claims, providerWithIdentity, i)
		require.NoError(t, err)
	})

	t.Run("case=verified addresses are applied from mapper output", func(t *testing.T) {
		t.Parallel()
		// Use a schema with verification enabled so Validate populates VerifiableAddresses.
		_, verifyReg := pkg.NewFastRegistryWithMocks(t, configx.WithValues(map[string]any{
			config.ViperKeyIdentitySchemas: config.Schemas{
				{ID: "with-verification", URL: "file://./stub/registration-with-verification.schema.json"},
			},
			config.ViperKeyDefaultIdentitySchemaID: "with-verification",
		}))
		verifyStrategy := oidc.NewStrategy(verifyReg)

		i := &identity.Identity{
			SchemaID:       "with-verification",
			Traits:         identity.Traits(`{"subject":"alice@example.com"}`),
			MetadataPublic: []byte(`{}`),
			MetadataAdmin:  []byte(`{}`),
		}

		claims := &oidc.Claims{
			Subject: "alice@example.com",
			RawClaims: map[string]any{
				"groups": []any{"admin"},
			},
		}

		// The hydra mapper outputs verified_addresses: [{via: "email", value: claims.sub}].
		changed, err := verifyStrategy.UpdateIdentityFromClaims(ctx, claims, provider, i)
		require.NoError(t, err)
		assert.True(t, changed)

		// The address should now be verified.
		require.Len(t, i.VerifiableAddresses, 1)
		assert.True(t, i.VerifiableAddresses[0].Verified)
		assert.Equal(t, identity.VerifiableAddressStatusCompleted, i.VerifiableAddresses[0].Status)
		assert.NotNil(t, i.VerifiableAddresses[0].VerifiedAt)
	})

	t.Run("case=already verified addresses are not re-verified", func(t *testing.T) {
		t.Parallel()
		_, verifyReg := pkg.NewFastRegistryWithMocks(t, configx.WithValues(map[string]any{
			config.ViperKeyIdentitySchemas: config.Schemas{
				{ID: "with-verification", URL: "file://./stub/registration-with-verification.schema.json"},
			},
			config.ViperKeyDefaultIdentitySchemaID: "with-verification",
		}))
		verifyStrategy := oidc.NewStrategy(verifyReg)

		// Use subject-only mapper so traits don't change, isolating the
		// verified address behavior.
		providerSubjectOnly := &testUpdateProvider{config: &oidc.Configuration{
			ID:       "test",
			Provider: "generic",
			Mapper:   "file://./stub/oidc.subject-only.jsonnet",
		}}

		originalVerifiedAt := sqlxx.NullTime(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
		i := &identity.Identity{
			SchemaID:       "with-verification",
			Traits:         identity.Traits(`{"subject":"alice@example.com"}`),
			MetadataPublic: []byte(`{}`),
			MetadataAdmin:  []byte(`{}`),
			VerifiableAddresses: []identity.VerifiableAddress{
				{
					Value:      "alice@example.com",
					Via:        identity.AddressTypeEmail,
					Status:     identity.VerifiableAddressStatusCompleted,
					Verified:   true,
					VerifiedAt: &originalVerifiedAt,
				},
			},
			RecoveryAddresses: []identity.RecoveryAddress{
				{
					Value: "alice@example.com",
					Via:   identity.AddressTypeEmail,
				},
			},
		}

		claims := &oidc.Claims{
			Subject:   "alice@example.com",
			RawClaims: map[string]any{"_placeholder": true},
		}

		changed, err := verifyStrategy.UpdateIdentityFromClaims(ctx, claims, providerSubjectOnly, i)
		require.NoError(t, err)
		// No change — traits are the same, address was already verified.
		assert.False(t, changed)

		// VerifiedAt should not be updated.
		require.Len(t, i.VerifiableAddresses, 1)
		assert.True(t, i.VerifiableAddresses[0].Verified)
		assert.Equal(t, &originalVerifiedAt, i.VerifiableAddresses[0].VerifiedAt)
	})

	t.Run("case=mapper that omits metadata preserves existing values", func(t *testing.T) {
		t.Parallel()
		// This mapper only outputs traits, not metadata_public or metadata_admin.
		providerTraitsOnly := &testUpdateProvider{config: &oidc.Configuration{
			ID:       "test",
			Provider: "generic",
			Mapper:   "file://./stub/oidc.traits-only.jsonnet",
		}}

		i := &identity.Identity{
			SchemaID:       "default",
			Traits:         identity.Traits(`{"subject":"alice@example.com"}`),
			MetadataPublic: []byte(`{"preserved":"yes"}`),
			MetadataAdmin:  []byte(`{"secret":"admin-data"}`),
		}

		claims := &oidc.Claims{
			Subject:   "alice@example.com",
			Website:   "https://new.example.com",
			RawClaims: map[string]any{"_placeholder": true},
		}

		changed, err := s.UpdateIdentityFromClaims(ctx, claims, providerTraitsOnly, i)
		require.NoError(t, err)
		assert.True(t, changed, "traits changed (website added)")

		// Metadata should be preserved since the mapper doesn't output metadata keys.
		assert.Equal(t, `{"preserved":"yes"}`, string(i.MetadataPublic))
		assert.Equal(t, `{"secret":"admin-data"}`, string(i.MetadataAdmin))
	})

	t.Run("case=recovery addresses survive UpdateIdentityFromClaims", func(t *testing.T) {
		t.Parallel()
		_, verifyReg := pkg.NewFastRegistryWithMocks(t, configx.WithValues(map[string]any{
			config.ViperKeyIdentitySchemas: config.Schemas{
				{ID: "with-verification", URL: "file://./stub/registration-with-verification.schema.json"},
			},
			config.ViperKeyDefaultIdentitySchemaID: "with-verification",
		}))
		verifyStrategy := oidc.NewStrategy(verifyReg)

		recoveryID := uuid.Must(uuid.NewV4())
		i := &identity.Identity{
			SchemaID:       "with-verification",
			Traits:         identity.Traits(`{"subject":"alice@example.com"}`),
			MetadataPublic: []byte(`{}`),
			MetadataAdmin:  []byte(`{}`),
			RecoveryAddresses: []identity.RecoveryAddress{
				{
					ID:    recoveryID,
					Value: "alice@example.com",
					Via:   identity.AddressTypeEmail,
				},
			},
		}

		claims := &oidc.Claims{
			Subject:   "alice@example.com",
			RawClaims: map[string]any{"_placeholder": true},
		}

		_, err := verifyStrategy.UpdateIdentityFromClaims(ctx, claims, provider, i)
		require.NoError(t, err)

		// Recovery address should survive the Validate call.
		require.Len(t, i.RecoveryAddresses, 1)
		assert.Equal(t, "alice@example.com", i.RecoveryAddresses[0].Value)
		assert.Equal(t, identity.AddressTypeEmail, i.RecoveryAddresses[0].Via)
		// The ID should be preserved from the original.
		assert.Equal(t, recoveryID, i.RecoveryAddresses[0].ID)
	})

	t.Run("case=verified status is sticky when mapper stops asserting it", func(t *testing.T) {
		t.Parallel()
		_, verifyReg := pkg.NewFastRegistryWithMocks(t, configx.WithValues(map[string]any{
			config.ViperKeyIdentitySchemas: config.Schemas{
				{ID: "with-verification", URL: "file://./stub/registration-with-verification.schema.json"},
			},
			config.ViperKeyDefaultIdentitySchemaID: "with-verification",
		}))
		verifyStrategy := oidc.NewStrategy(verifyReg)

		// Use a mapper that does NOT output verified_addresses.
		providerNoVerify := &testUpdateProvider{config: &oidc.Configuration{
			ID:       "test",
			Provider: "generic",
			Mapper:   "file://./stub/oidc.subject-only.jsonnet",
		}}

		originalVerifiedAt := sqlxx.NullTime(time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC))
		i := &identity.Identity{
			SchemaID:       "with-verification",
			Traits:         identity.Traits(`{"subject":"alice@example.com"}`),
			MetadataPublic: []byte(`{}`),
			MetadataAdmin:  []byte(`{}`),
			VerifiableAddresses: []identity.VerifiableAddress{
				{
					Value:      "alice@example.com",
					Via:        identity.AddressTypeEmail,
					Status:     identity.VerifiableAddressStatusCompleted,
					Verified:   true,
					VerifiedAt: &originalVerifiedAt,
				},
			},
			RecoveryAddresses: []identity.RecoveryAddress{
				{
					Value: "alice@example.com",
					Via:   identity.AddressTypeEmail,
				},
			},
		}

		claims := &oidc.Claims{
			Subject:   "alice@example.com",
			RawClaims: map[string]any{"_placeholder": true},
		}

		changed, err := verifyStrategy.UpdateIdentityFromClaims(ctx, claims, providerNoVerify, i)
		require.NoError(t, err)
		assert.False(t, changed, "nothing should change")

		// Verified status should be sticky — the mapper not asserting it
		// should not revoke verification.
		require.Len(t, i.VerifiableAddresses, 1)
		assert.True(t, i.VerifiableAddresses[0].Verified)
		assert.Equal(t, identity.VerifiableAddressStatusCompleted, i.VerifiableAddresses[0].Status)
		assert.Equal(t, &originalVerifiedAt, i.VerifiableAddresses[0].VerifiedAt)
	})
}
