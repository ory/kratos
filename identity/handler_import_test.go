// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/snapshotx"
)

func TestImportCredentials(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Setup handler with minimal mock requirements
	h := &Handler{}

	testCases := []struct {
		name          string
		setupIdentity func() *Identity
		credentials   interface{}
		credType      CredentialsType
	}{
		{
			name: "OIDC new credential without organization",
			setupIdentity: func() *Identity {
				return &Identity{}
			},
			credentials: &AdminIdentityImportCredentialsOIDC{
				Config: AdminIdentityImportCredentialsOIDCConfig{
					Providers: []AdminCreateIdentityImportCredentialsOIDCProvider{
						{
							Provider: "github",
							Subject:  "12345",
						},
					},
				},
			},
			credType: CredentialsTypeOIDC,
		},
		{
			name: "OIDC new credential with organization",
			setupIdentity: func() *Identity {
				return &Identity{}
			},
			credentials: &AdminIdentityImportCredentialsOIDC{
				Config: AdminIdentityImportCredentialsOIDCConfig{
					Providers: []AdminCreateIdentityImportCredentialsOIDCProvider{
						{
							Provider:     "github",
							Subject:      "12345",
							Organization: uuid.NullUUID{UUID: uuid.FromStringOrNil("e7e3cbae-04cc-45f3-ae52-ea749a2ffaff"), Valid: true},
						},
					},
				},
			},
			credType: CredentialsTypeOIDC,
		},
		{
			name: "OIDC update credential without organization",
			setupIdentity: func() *Identity {
				i := &Identity{}
				_ = i.SetCredentialsWithConfig(
					CredentialsTypeOIDC,
					Credentials{
						Identifiers: []string{OIDCUniqueID("google", "67890")},
					},
					CredentialsOIDC{
						Providers: []CredentialsOIDCProvider{
							{
								Provider: "google",
								Subject:  "67890",
							},
						},
					},
				)
				return i
			},
			credentials: &AdminIdentityImportCredentialsOIDC{
				Config: AdminIdentityImportCredentialsOIDCConfig{
					Providers: []AdminCreateIdentityImportCredentialsOIDCProvider{
						{
							Provider: "github",
							Subject:  "12345",
						},
					},
				},
			},
			credType: CredentialsTypeOIDC,
		},
		{
			name: "OIDC update credential with organization",
			setupIdentity: func() *Identity {
				i := &Identity{}
				_ = i.SetCredentialsWithConfig(
					CredentialsTypeOIDC,
					Credentials{
						Identifiers: []string{OIDCUniqueID("google", "67890")},
					},
					CredentialsOIDC{
						Providers: []CredentialsOIDCProvider{
							{
								Provider: "google",
								Subject:  "67890",
							},
						},
					},
				)
				return i
			},
			credentials: &AdminIdentityImportCredentialsOIDC{
				Config: AdminIdentityImportCredentialsOIDCConfig{
					Providers: []AdminCreateIdentityImportCredentialsOIDCProvider{
						{
							Provider:     "github",
							Subject:      "12345",
							Organization: uuid.NullUUID{UUID: uuid.FromStringOrNil("e7e3cbae-04cc-45f3-ae52-ea749a2ffaff"), Valid: true},
						},
					},
				},
			},
			credType: CredentialsTypeOIDC,
		},
		{
			name: "OIDC update with multiple providers",
			setupIdentity: func() *Identity {
				i := &Identity{}
				_ = i.SetCredentialsWithConfig(
					CredentialsTypeOIDC,
					Credentials{
						Identifiers: []string{OIDCUniqueID("google", "67890")},
					},
					CredentialsOIDC{
						Providers: []CredentialsOIDCProvider{
							{
								Provider: "google",
								Subject:  "67890",
							},
						},
					},
				)
				return i
			},
			credentials: &AdminIdentityImportCredentialsOIDC{
				Config: AdminIdentityImportCredentialsOIDCConfig{
					Providers: []AdminCreateIdentityImportCredentialsOIDCProvider{
						{
							Provider:     "github",
							Subject:      "12345",
							Organization: uuid.NullUUID{UUID: uuid.FromStringOrNil("e7e3cbae-04cc-45f3-ae52-ea749a2ffaff"), Valid: true},
						},
						{
							Provider: "gitlab",
							Subject:  "abcdef",
						},
					},
				},
			},
			credType: CredentialsTypeOIDC,
		},
		{
			name: "SAML new credential without organization",
			setupIdentity: func() *Identity {
				return &Identity{}
			},
			credentials: &AdminIdentityImportCredentialsSAML{
				Config: AdminIdentityImportCredentialsSAMLConfig{
					Providers: []AdminCreateIdentityImportCredentialsSAMLProvider{
						{
							Provider: "okta",
							Subject:  "user123",
						},
					},
				},
			},
			credType: CredentialsTypeSAML,
		},
		{
			name: "SAML new credential with organization",
			setupIdentity: func() *Identity {
				return &Identity{}
			},
			credentials: &AdminIdentityImportCredentialsSAML{
				Config: AdminIdentityImportCredentialsSAMLConfig{
					Providers: []AdminCreateIdentityImportCredentialsSAMLProvider{
						{
							Provider:     "okta",
							Subject:      "user123",
							Organization: uuid.NullUUID{UUID: uuid.FromStringOrNil("e7e3cbae-04cc-45f3-ae52-ea749a2ffaff"), Valid: true},
						},
					},
				},
			},
			credType: CredentialsTypeSAML,
		},
		{
			name: "SAML update credential without organization",
			setupIdentity: func() *Identity {
				i := &Identity{}
				_ = i.SetCredentialsWithConfig(
					CredentialsTypeSAML,
					Credentials{
						Identifiers: []string{OIDCUniqueID("onelogin", "user456")},
					},
					CredentialsOIDC{
						Providers: []CredentialsOIDCProvider{
							{
								Provider: "onelogin",
								Subject:  "user456",
							},
						},
					},
				)
				return i
			},
			credentials: &AdminIdentityImportCredentialsSAML{
				Config: AdminIdentityImportCredentialsSAMLConfig{
					Providers: []AdminCreateIdentityImportCredentialsSAMLProvider{
						{
							Provider: "okta",
							Subject:  "user123",
						},
					},
				},
			},
			credType: CredentialsTypeSAML,
		},
		{
			name: "SAML update credential with organization",
			setupIdentity: func() *Identity {
				i := &Identity{}
				_ = i.SetCredentialsWithConfig(
					CredentialsTypeSAML,
					Credentials{
						Identifiers: []string{OIDCUniqueID("onelogin", "user456")},
					},
					CredentialsOIDC{
						Providers: []CredentialsOIDCProvider{
							{
								Provider: "onelogin",
								Subject:  "user456",
							},
						},
					},
				)
				return i
			},
			credentials: &AdminIdentityImportCredentialsSAML{
				Config: AdminIdentityImportCredentialsSAMLConfig{
					Providers: []AdminCreateIdentityImportCredentialsSAMLProvider{
						{
							Provider:     "okta",
							Subject:      "user123",
							Organization: uuid.NullUUID{UUID: uuid.FromStringOrNil("e7e3cbae-04cc-45f3-ae52-ea749a2ffaff"), Valid: true},
						},
					},
				},
			},
			credType: CredentialsTypeSAML,
		},
		{
			name: "SAML update with multiple providers",
			setupIdentity: func() *Identity {
				i := &Identity{}
				_ = i.SetCredentialsWithConfig(
					CredentialsTypeSAML,
					Credentials{
						Identifiers: []string{OIDCUniqueID("onelogin", "user456")},
					},
					CredentialsOIDC{
						Providers: []CredentialsOIDCProvider{
							{
								Provider: "onelogin",
								Subject:  "user456",
							},
						},
					},
				)
				return i
			},
			credentials: &AdminIdentityImportCredentialsSAML{
				Config: AdminIdentityImportCredentialsSAMLConfig{
					Providers: []AdminCreateIdentityImportCredentialsSAMLProvider{
						{
							Provider:     "okta",
							Subject:      "user123",
							Organization: uuid.NullUUID{UUID: uuid.FromStringOrNil("e7e3cbae-04cc-45f3-ae52-ea749a2ffaff"), Valid: true},
						},
						{
							Provider: "auth0",
							Subject:  "user789",
						},
					},
				},
			},
			credType: CredentialsTypeSAML,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup a fresh identity for each test
			i := tc.setupIdentity()
			var err error

			// Perform the import based on credential type
			switch tc.credType {
			case CredentialsTypeOIDC:
				err = h.importOIDCCredentials(ctx, i, tc.credentials.(*AdminIdentityImportCredentialsOIDC))
			case CredentialsTypeSAML:
				err = h.importSAMLCredentials(ctx, i, tc.credentials.(*AdminIdentityImportCredentialsSAML))
			}

			require.NoError(t, err)

			// Verify credential was set correctly
			creds, ok := i.GetCredentials(tc.credType)
			require.True(t, ok, "credentials should be set")

			// Verify the credentials contain proper identifiers and config
			assert.NotEmpty(t, creds.Identifiers)
			assert.NotEmpty(t, creds.Config)

			// Take a snapshot of the credentials
			snapshotx.SnapshotT(t, creds)

			// Additional checks based on credential type
			switch tc.credType {
			case CredentialsTypeOIDC:
				oidcCreds := tc.credentials.(*AdminIdentityImportCredentialsOIDC)
				for _, p := range oidcCreds.Config.Providers {
					id := OIDCUniqueID(p.Provider, p.Subject)
					assert.Contains(t, creds.Identifiers, id)
				}
			case CredentialsTypeSAML:
				samlCreds := tc.credentials.(*AdminIdentityImportCredentialsSAML)
				for _, p := range samlCreds.Config.Providers {
					id := OIDCUniqueID(p.Provider, p.Subject)
					assert.Contains(t, creds.Identifiers, id)
				}
			}
		})
	}
}
