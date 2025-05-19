// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/snapshotx"
)

func TestImportCredentialsOidcSAML(t *testing.T) {
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
			// Set  up a fresh identity for each test
			i := tc.setupIdentity()
			var err error

			// Perform the import based on a credential type
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

func TestImportLookupSecretCredentials(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Setup handler with minimal mock requirements
	h := &Handler{}

	testCases := []struct {
		name          string
		setupIdentity func() *Identity
		credentials   *AdminIdentityImportCredentialsLookupSecret
		expectedCodes int
	}{
		{
			name: "new lookup secret credential",
			setupIdentity: func() *Identity {
				return &Identity{}
			},
			credentials: &AdminIdentityImportCredentialsLookupSecret{
				Config: AdminIdentityImportCredentialsLookupSecretConfig{
					Codes: []RecoveryCode{
						{Code: "code1"},
						{Code: "code2"},
					},
				},
			},
			expectedCodes: 2,
		},
		{
			name: "update existing lookup secret credential",
			setupIdentity: func() *Identity {
				i := &Identity{}
				_ = i.SetCredentialsWithConfig(
					CredentialsTypeLookup,
					Credentials{},
					CredentialsLookupConfig{
						RecoveryCodes: []RecoveryCode{
							{Code: "existing-code"},
						},
					},
				)
				return i
			},
			credentials: &AdminIdentityImportCredentialsLookupSecret{
				Config: AdminIdentityImportCredentialsLookupSecretConfig{
					Codes: []RecoveryCode{
						{Code: "new-code1"},
						{Code: "new-code2"},
					},
				},
			},
			expectedCodes: 3, // 1 existing + 2 new
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up a fresh identity for each test
			i := tc.setupIdentity()

			// Perform the import
			err := h.importLookupSecretCredentials(ctx, i, tc.credentials)
			require.NoError(t, err)

			// Verify credential was set correctly
			creds, ok := i.GetCredentials(CredentialsTypeLookup)
			require.True(t, ok, "credentials should be set")

			// Parse the config to check the recovery codes
			var config CredentialsLookupConfig
			require.NoError(t, json.Unmarshal(creds.Config, &config))

			// Verify the expected number of codes
			assert.Len(t, config.RecoveryCodes, tc.expectedCodes)

			// Take a snapshot of the credentials
			snapshotx.SnapshotT(t, creds)

			// Verify specific codes based on a test case
			if tc.name == "new lookup secret credential" {
				assert.Contains(t, getCodeValues(config.RecoveryCodes), "code1")
				assert.Contains(t, getCodeValues(config.RecoveryCodes), "code2")
			} else {
				assert.Contains(t, getCodeValues(config.RecoveryCodes), "existing-code")
				assert.Contains(t, getCodeValues(config.RecoveryCodes), "new-code1")
				assert.Contains(t, getCodeValues(config.RecoveryCodes), "new-code2")
			}
		})
	}
}

func TestImportPasskeyCredentials(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Setup handler
	h := &Handler{}

	// Create test credentials
	cred1 := CredentialWebAuthn{ID: []byte("cred1"), PublicKey: []byte("pk1")}
	cred2 := CredentialWebAuthn{ID: []byte("cred2"), PublicKey: []byte("pk2")}
	cred1Updated := CredentialWebAuthn{ID: []byte("cred1"), PublicKey: []byte("pk1-updated")}
	cred3 := CredentialWebAuthn{ID: []byte("cred3"), PublicKey: []byte("pk3")}

	testCases := []struct {
		name          string
		setupIdentity func() *Identity
		credentials   CredentialsWebAuthn
		userHandle    []byte
		verify        func(t *testing.T, i *Identity)
	}{
		{
			name: "new passkey credentials",
			setupIdentity: func() *Identity {
				return &Identity{}
			},
			credentials: CredentialsWebAuthn{cred1, cred2},
			userHandle:  []byte("user1"),
			verify: func(t *testing.T, i *Identity) {
				creds, ok := i.GetCredentials(CredentialsTypePasskey)
				require.True(t, ok)

				var config CredentialsWebAuthnConfig
				require.NoError(t, json.Unmarshal(creds.Config, &config))

				assert.Len(t, config.Credentials, 2)
				assert.Equal(t, []byte("user1"), config.UserHandle)
				assert.Contains(t, creds.Identifiers, "user1")
			},
		},
		{
			name: "update existing credentials - add new, update existing",
			setupIdentity: func() *Identity {
				i := &Identity{}
				err := i.SetCredentialsWithConfig(
					CredentialsTypePasskey,
					Credentials{
						Identifiers: []string{"existingUser"},
					},
					CredentialsWebAuthnConfig{
						Credentials: CredentialsWebAuthn{cred1, cred2},
						UserHandle:  []byte("existingUser"),
					},
				)
				require.NoError(t, err)
				return i
			},
			credentials: CredentialsWebAuthn{cred1Updated, cred3},
			userHandle:  []byte{}, // Empty to test reusing existing
			verify: func(t *testing.T, i *Identity) {
				creds, ok := i.GetCredentials(CredentialsTypePasskey)
				require.True(t, ok)

				var config CredentialsWebAuthnConfig
				require.NoError(t, json.Unmarshal(creds.Config, &config))

				// Should have updated cred1, kept cred2, and added cred3
				assert.Len(t, config.Credentials, 3)
				assert.Equal(t, []byte("existingUser"), config.UserHandle)

				// Find the updated credential
				var found bool
				for _, c := range config.Credentials {
					if bytes.Equal(c.ID, []byte("cred1")) {
						assert.Equal(t, []byte("pk1-updated"), c.PublicKey)
						found = true
						break
					}
				}
				assert.True(t, found, "Updated credential should be present")

				// Verify cred3 was added
				var foundNew bool
				for _, c := range config.Credentials {
					if bytes.Equal(c.ID, []byte("cred3")) {
						foundNew = true
						break
					}
				}
				assert.True(t, foundNew, "New credential should be added")
			},
		},
		{
			name: "new user handle added to identifiers",
			setupIdentity: func() *Identity {
				i := &Identity{}
				err := i.SetCredentialsWithConfig(
					CredentialsTypePasskey,
					Credentials{
						Identifiers: []string{"existingUser"},
					},
					CredentialsWebAuthnConfig{
						Credentials: CredentialsWebAuthn{cred1},
						UserHandle:  []byte("existingUser"),
					},
				)
				require.NoError(t, err)
				return i
			},
			credentials: CredentialsWebAuthn{cred2},
			userHandle:  []byte("newUser"),
			verify: func(t *testing.T, i *Identity) {
				creds, ok := i.GetCredentials(CredentialsTypePasskey)
				require.True(t, ok)

				var config CredentialsWebAuthnConfig
				require.NoError(t, json.Unmarshal(creds.Config, &config))

				assert.Equal(t, []byte("newUser"), config.UserHandle)
				assert.Contains(t, creds.Identifiers, "existingUser")
				assert.Contains(t, creds.Identifiers, "newUser")
				assert.Len(t, creds.Identifiers, 2)
				assert.Len(t, config.Credentials, 2)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			i := tc.setupIdentity()
			err := h.importPasskeyCredentials(ctx, i, tc.credentials, tc.userHandle)
			require.NoError(t, err)
			tc.verify(t, i)
			snapshotx.SnapshotT(t, i.Credentials)
		})
	}
}

func TestImportWebAuthnCredentials(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Setup handler
	h := &Handler{}

	// Create test credentials
	cred1 := CredentialWebAuthn{ID: []byte("cred1"), PublicKey: []byte("pk1")}
	cred2 := CredentialWebAuthn{ID: []byte("cred2"), PublicKey: []byte("pk2")}
	cred1Updated := CredentialWebAuthn{ID: []byte("cred1"), PublicKey: []byte("pk1-updated")}
	cred3 := CredentialWebAuthn{ID: []byte("cred3"), PublicKey: []byte("pk3")}

	testCases := []struct {
		name          string
		setupIdentity func() *Identity
		credentials   CredentialsWebAuthn
		userHandle    []byte
		verify        func(t *testing.T, i *Identity)
	}{
		{
			name: "new webauthn credentials",
			setupIdentity: func() *Identity {
				return &Identity{}
			},
			credentials: CredentialsWebAuthn{cred1, cred2},
			userHandle:  []byte("user1"),
			verify: func(t *testing.T, i *Identity) {
				creds, ok := i.GetCredentials(CredentialsTypeWebAuthn)
				require.True(t, ok)

				var config CredentialsWebAuthnConfig
				require.NoError(t, json.Unmarshal(creds.Config, &config))

				assert.Len(t, config.Credentials, 2)
				assert.Equal(t, []byte("user1"), config.UserHandle)
			},
		},
		{
			name: "update existing credentials - add new, update existing",
			setupIdentity: func() *Identity {
				i := &Identity{}
				err := i.SetCredentialsWithConfig(
					CredentialsTypeWebAuthn,
					Credentials{},
					CredentialsWebAuthnConfig{
						Credentials: CredentialsWebAuthn{cred1, cred2},
						UserHandle:  []byte("existingUser"),
					},
				)
				require.NoError(t, err)
				return i
			},
			credentials: CredentialsWebAuthn{cred1Updated, cred3},
			userHandle:  []byte{}, // Empty to test reusing existing
			verify: func(t *testing.T, i *Identity) {
				creds, ok := i.GetCredentials(CredentialsTypeWebAuthn)
				require.True(t, ok)

				var config CredentialsWebAuthnConfig
				require.NoError(t, json.Unmarshal(creds.Config, &config))

				// Should have updated cred1, kept cred2, and added cred3
				assert.Len(t, config.Credentials, 3)
				assert.Equal(t, []byte("existingUser"), config.UserHandle)

				// Find the updated credential
				var found bool
				for _, c := range config.Credentials {
					if bytes.Equal(c.ID, []byte("cred1")) {
						assert.Equal(t, []byte("pk1-updated"), c.PublicKey)
						found = true
						break
					}
				}
				assert.True(t, found, "Updated credential should be present")

				// Verify cred3 was added
				var foundNew bool
				for _, c := range config.Credentials {
					if bytes.Equal(c.ID, []byte("cred3")) {
						foundNew = true
						break
					}
				}
				assert.True(t, foundNew, "New credential should be added")
			},
		},
		{
			name: "override existing user handle",
			setupIdentity: func() *Identity {
				i := &Identity{}
				err := i.SetCredentialsWithConfig(
					CredentialsTypeWebAuthn,
					Credentials{},
					CredentialsWebAuthnConfig{
						Credentials: CredentialsWebAuthn{cred1},
						UserHandle:  []byte("existingUser"),
					},
				)
				require.NoError(t, err)
				return i
			},
			credentials: CredentialsWebAuthn{cred2},
			userHandle:  []byte("newUser"),
			verify: func(t *testing.T, i *Identity) {
				creds, ok := i.GetCredentials(CredentialsTypeWebAuthn)
				require.True(t, ok)

				var config CredentialsWebAuthnConfig
				require.NoError(t, json.Unmarshal(creds.Config, &config))

				assert.Equal(t, []byte("newUser"), config.UserHandle)
				assert.Len(t, config.Credentials, 2)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			i := tc.setupIdentity()
			err := h.importWebAuthnCredentials(ctx, i, tc.credentials, tc.userHandle)
			require.NoError(t, err)
			tc.verify(t, i)
			snapshotx.SnapshotT(t, i.Credentials)
		})
	}
}

func TestImportTOTPCredentials(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Setup handler
	h := &Handler{}

	testCases := []struct {
		name          string
		setupIdentity func() *Identity
		credentials   *AdminIdentityImportCredentialsTOTP
		verify        func(t *testing.T, i *Identity)
	}{
		{
			name: "new totp credential",
			setupIdentity: func() *Identity {
				return &Identity{}
			},
			credentials: &AdminIdentityImportCredentialsTOTP{
				Config: AdminIdentityImportCredentialsTOTPConfig{
					TOTPURL: "otpauth://totp/Example:alice@example.com?secret=JBSWY3DPEHPK3PXP&issuer=Example",
				},
			},
			verify: func(t *testing.T, i *Identity) {
				creds, ok := i.GetCredentials(CredentialsTypeTOTP)
				require.True(t, ok)

				var config CredentialsTOTPConfig
				require.NoError(t, json.Unmarshal(creds.Config, &config))

				assert.Equal(t, "otpauth://totp/Example:alice@example.com?secret=JBSWY3DPEHPK3PXP&issuer=Example", config.TOTPURL)
			},
		},
		{
			name: "update existing totp credential",
			setupIdentity: func() *Identity {
				i := &Identity{}
				err := i.SetCredentialsWithConfig(
					CredentialsTypeTOTP,
					Credentials{},
					CredentialsTOTPConfig{
						TOTPURL: "otpauth://totp/Example:alice@example.com?secret=OLDSECRET&issuer=Example",
					},
				)
				require.NoError(t, err)
				return i
			},
			credentials: &AdminIdentityImportCredentialsTOTP{
				Config: AdminIdentityImportCredentialsTOTPConfig{
					TOTPURL: "otpauth://totp/Example:alice@example.com?secret=NEWSECRET&issuer=Example",
				},
			},
			verify: func(t *testing.T, i *Identity) {
				creds, ok := i.GetCredentials(CredentialsTypeTOTP)
				require.True(t, ok)

				var config CredentialsTOTPConfig
				require.NoError(t, json.Unmarshal(creds.Config, &config))

				// Should have the new TOTP URL
				assert.Equal(t, "otpauth://totp/Example:alice@example.com?secret=NEWSECRET&issuer=Example", config.TOTPURL)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			i := tc.setupIdentity()
			err := h.importTOTPCredentials(ctx, i, tc.credentials)
			require.NoError(t, err)
			tc.verify(t, i)
			snapshotx.SnapshotT(t, i.Credentials)
		})
	}
}

// Helper function to extract code values for easier assertions
func getCodeValues(codes []RecoveryCode) []string {
	var values []string
	for _, code := range codes {
		values = append(values, code.Code)
	}
	return values
}
