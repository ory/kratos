// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/herodot"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/persistence"
	idpersistence "github.com/ory/kratos/persistence/sql/identity"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/x"
	"github.com/ory/x/assertx"
	"github.com/ory/x/contextx"
	"github.com/ory/x/crdbx"
	"github.com/ory/x/pagination/keysetpagination"
	"github.com/ory/x/randx"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"
)

// assertContainsValues is a test helper that checks if a slice contains expected values and doesn't contain unexpected values.
func assertContainsValues(t *testing.T, actual []string, shouldContain, shouldNotContain []string) {
	t.Helper()
	for _, expected := range shouldContain {
		assert.Contains(t, actual, expected)
	}
	for _, notExpected := range shouldNotContain {
		assert.NotContains(t, actual, notExpected)
	}
}

func TestPool(ctx context.Context, p persistence.Persister, m *identity.Manager, dbname string) func(t *testing.T) {
	return func(t *testing.T) {
		nid, p := testhelpers.NewNetworkUnlessExisting(t, ctx, p)

		exampleServerURL := urlx.ParseOrPanic("http://example.com")
		expandSchema := schema.Schema{
			ID:     "expandSchema",
			URL:    urlx.ParseOrPanic("file://./stub/expand.schema.json"),
			RawURL: "file://./stub/expand.schema.json",
		}
		defaultSchema := schema.Schema{
			ID:     config.DefaultIdentityTraitsSchemaID,
			URL:    urlx.ParseOrPanic("file://./stub/identity.schema.json"),
			RawURL: "file://./stub/identity.schema.json",
		}
		altSchema := schema.Schema{
			ID:     "altSchema",
			URL:    urlx.ParseOrPanic("file://./stub/identity-2.schema.json"),
			RawURL: "file://./stub/identity-2.schema.json",
		}
		multipleEmailsSchema := schema.Schema{
			ID:     "multiple_emails",
			URL:    urlx.ParseOrPanic("file://./stub/handler/multiple_emails.schema.json"),
			RawURL: "file://./stub/identity-2.schema.json",
		}
		phoneEmailSchema := schema.Schema{
			ID:     "phoneIdentifier",
			URL:    urlx.ParseOrPanic("file://./stub/phone.schema.json"),
			RawURL: "file://./stub/phone.schema.json",
		}
		ctx := contextx.WithConfigValues(ctx, map[string]any{
			config.ViperKeyPublicBaseURL: exampleServerURL.String(),
			config.ViperKeyIdentitySchemas: []config.Schema{
				{
					ID:  altSchema.ID,
					URL: altSchema.RawURL,
				},
				{
					ID:  defaultSchema.ID,
					URL: defaultSchema.RawURL,
				},
				{
					ID:  expandSchema.ID,
					URL: expandSchema.RawURL,
				},
				{
					ID:  multipleEmailsSchema.ID,
					URL: multipleEmailsSchema.RawURL,
				},
				{
					ID:  phoneEmailSchema.ID,
					URL: phoneEmailSchema.RawURL,
				},
			},
		})

		t.Run("case=expand", func(t *testing.T) {
			require.NoError(t, p.GetConnection(ctx).RawQuery("DELETE FROM identities WHERE nid = ?", nid).Exec())
			t.Cleanup(func() {
				require.NoError(t, p.GetConnection(ctx).RawQuery("DELETE FROM identities WHERE nid = ?", nid).Exec())
			})

			expected := identity.NewIdentity(expandSchema.ID)
			expected.Traits = identity.Traits(`{"email":"` + uuid.Must(uuid.NewV4()).String() + "@ory.sh" + `","name":"john doe"}`)
			require.NoError(t, m.ValidateIdentity(ctx, expected, new(identity.ManagerOptions)))
			require.NoError(t, p.CreateIdentity(ctx, expected))
			require.NoError(t, identity.UpgradeCredentials(expected))

			assert.NotEmpty(t, expected.RecoveryAddresses)
			assert.NotEmpty(t, expected.VerifiableAddresses)
			assert.NotEmpty(t, expected.Credentials)
			assert.NotEqual(t, uuid.Nil, expected.RecoveryAddresses[0].ID)
			assert.NotEqual(t, uuid.Nil, expected.VerifiableAddresses[0].ID)

			runner := func(t *testing.T, expand sqlxx.Expandables, cb func(*testing.T, *identity.Identity)) {
				assertion := func(t *testing.T, actual *identity.Identity) {
					assertx.EqualAsJSONExcept(t, expected, actual, []string{
						"verifiable_addresses", "recovery_addresses", "updated_at", "created_at", "credentials", "state_changed_at",
					})
					cb(t, actual)
				}

				t.Run("find", func(t *testing.T) {
					actual, err := p.GetIdentity(ctx, expected.ID, expand)
					require.NoError(t, err)
					assertion(t, actual)
				})

				t.Run("list/page-pagination", func(t *testing.T) {
					actual, _, err := p.ListIdentities(ctx, identity.ListIdentityParameters{Expand: expand, PagePagination: &x.Page{Page: 0, ItemsPerPage: 10}})
					require.NoError(t, err)
					require.Len(t, actual, 1)
					assertion(t, &actual[0])
				})

				t.Run("list/token-pagination", func(t *testing.T) {
					actual, next, err := p.ListIdentities(ctx, identity.ListIdentityParameters{Expand: expand, KeySetPagination: []keysetpagination.Option{keysetpagination.WithSize(10)}})
					require.NoError(t, err)
					require.Len(t, actual, 1)
					require.True(t, next.IsLast())
					assertion(t, &actual[0])
				})
			}

			t.Run("expand=nothing", func(t *testing.T) {
				runner(t, identity.ExpandNothing, func(t *testing.T, actual *identity.Identity) {
					assert.Empty(t, actual.RecoveryAddresses)
					assert.Empty(t, actual.VerifiableAddresses)
					assert.Empty(t, actual.Credentials)
				})
			})

			t.Run("expand=credentials", func(t *testing.T) {
				runner(t, identity.ExpandCredentials, func(t *testing.T, actual *identity.Identity) {
					assert.Empty(t, actual.RecoveryAddresses)
					assert.Empty(t, actual.VerifiableAddresses)

					require.Len(t, actual.Credentials, 2)

					assertx.EqualAsJSONExcept(t, expected.Credentials[identity.CredentialsTypePassword], actual.Credentials[identity.CredentialsTypePassword], []string{"updated_at", "created_at"})
					assertx.EqualAsJSONExcept(t, expected.Credentials[identity.CredentialsTypeWebAuthn], actual.Credentials[identity.CredentialsTypeWebAuthn], []string{"updated_at", "created_at"})
				})
			})

			t.Run("expand=recovery address", func(t *testing.T) {
				runner(t, sqlxx.Expandables{identity.ExpandFieldRecoveryAddresses}, func(t *testing.T, actual *identity.Identity) {
					assert.Empty(t, actual.Credentials)
					assert.Empty(t, actual.VerifiableAddresses)

					require.Len(t, actual.RecoveryAddresses, 1)
					assertx.EqualAsJSONExcept(t, expected.RecoveryAddresses, actual.RecoveryAddresses, []string{"0.updated_at", "0.created_at"})
				})
			})

			t.Run("expand=verification address", func(t *testing.T) {
				runner(t, sqlxx.Expandables{identity.ExpandFieldVerifiableAddresses}, func(t *testing.T, actual *identity.Identity) {
					assert.Empty(t, actual.Credentials)
					assert.Empty(t, actual.RecoveryAddresses)

					require.Len(t, actual.VerifiableAddresses, 1)
					assertx.EqualAsJSONExcept(t, expected.VerifiableAddresses, actual.VerifiableAddresses, []string{"0.updated_at", "0.created_at"})
				})
			})

			t.Run("expand=default", func(t *testing.T) {
				runner(t, identity.ExpandDefault, func(t *testing.T, actual *identity.Identity) {
					assert.Empty(t, actual.Credentials)

					require.Len(t, actual.RecoveryAddresses, 1)
					assertx.EqualAsJSONExcept(t, expected.RecoveryAddresses, actual.RecoveryAddresses, []string{"0.updated_at", "0.created_at"})

					require.Len(t, actual.VerifiableAddresses, 1)
					assertx.EqualAsJSONExcept(t, expected.VerifiableAddresses, actual.VerifiableAddresses, []string{"0.updated_at", "0.created_at"})
				})
			})

			t.Run("expand=everything", func(t *testing.T) {
				runner(t, identity.ExpandEverything, func(t *testing.T, actual *identity.Identity) {
					require.Len(t, actual.Credentials, 2)

					assertx.EqualAsJSONExcept(t, expected.Credentials[identity.CredentialsTypePassword], actual.Credentials[identity.CredentialsTypePassword], []string{"updated_at", "created_at"})
					assertx.EqualAsJSONExcept(t, expected.Credentials[identity.CredentialsTypeWebAuthn], actual.Credentials[identity.CredentialsTypeWebAuthn], []string{"updated_at", "created_at"})

					require.Len(t, actual.RecoveryAddresses, 1)
					assertx.EqualAsJSONExcept(t, expected.RecoveryAddresses, actual.RecoveryAddresses, []string{"0.updated_at", "0.created_at"})

					require.Len(t, actual.VerifiableAddresses, 1)
					assertx.EqualAsJSONExcept(t, expected.VerifiableAddresses, actual.VerifiableAddresses, []string{"0.updated_at", "0.created_at"})
				})
			})

			t.Run("expand=load", func(t *testing.T) {
				runner(t, identity.ExpandNothing, func(t *testing.T, actual *identity.Identity) {
					require.NoError(t, p.HydrateIdentityAssociations(ctx, actual, identity.ExpandEverything))

					require.Len(t, actual.Credentials, 2)

					assertx.EqualAsJSONExcept(t, expected.Credentials[identity.CredentialsTypePassword], actual.Credentials[identity.CredentialsTypePassword], []string{"updated_at", "created_at"})
					assertx.EqualAsJSONExcept(t, expected.Credentials[identity.CredentialsTypeWebAuthn], actual.Credentials[identity.CredentialsTypeWebAuthn], []string{"updated_at", "created_at"})

					require.Len(t, actual.RecoveryAddresses, 1)
					assertx.EqualAsJSONExcept(t, expected.RecoveryAddresses, actual.RecoveryAddresses, []string{"0.updated_at", "0.created_at"})

					require.Len(t, actual.VerifiableAddresses, 1)
					assertx.EqualAsJSONExcept(t, expected.VerifiableAddresses, actual.VerifiableAddresses, []string{"0.updated_at", "0.created_at"})
				})
			})

			t.Run("confidential", func(t *testing.T) {
				// confidential is like expand=all
				actual, err := p.GetIdentityConfidential(ctx, expected.ID)
				require.NoError(t, err)
				assertx.EqualAsJSONExcept(t, expected, actual, []string{
					"verifiable_addresses", "recovery_addresses", "updated_at", "created_at", "credentials", "state_changed_at",
				})
				require.Len(t, actual.Credentials, 2)

				assertx.EqualAsJSONExcept(t, expected.Credentials[identity.CredentialsTypePassword], actual.Credentials[identity.CredentialsTypePassword], []string{"updated_at", "created_at"})
				assertx.EqualAsJSONExcept(t, expected.Credentials[identity.CredentialsTypeWebAuthn], actual.Credentials[identity.CredentialsTypeWebAuthn], []string{"updated_at", "created_at"})

				require.Len(t, actual.RecoveryAddresses, 1)
				assertx.EqualAsJSONExcept(t, expected.RecoveryAddresses, actual.RecoveryAddresses, []string{"0.updated_at", "0.created_at"})

				require.Len(t, actual.VerifiableAddresses, 1)
				assertx.EqualAsJSONExcept(t, expected.VerifiableAddresses, actual.VerifiableAddresses, []string{"0.updated_at", "0.created_at"})
			})
		})

		t.Run("case=phone number backward compatibility", func(t *testing.T) {
			// Clean up identities after test to prevent pollution of subsequent tests
			t.Cleanup(func() {
				require.NoError(t, p.GetConnection(ctx).RawQuery("DELETE FROM identities WHERE nid = ? AND schema_id = ?", nid, phoneEmailSchema.ID).Exec())
			})

			// Test backward compatibility: simulating the upgrade path for legacy phone credentials
			nonNormalizedPhone := "+49 176 671 11 638"
			normalizedPhone := "+4917667111638"

			t.Run("new identities get normalized credentials from the start", func(t *testing.T) {
				// Create identity through validation (simulates user registration)
				newIdentity := identity.NewIdentity(phoneEmailSchema.ID)
				newIdentity.Traits = identity.Traits(`{"phone":"` + nonNormalizedPhone + `","name":"john doe"}`)
				require.NoError(t, m.ValidateIdentity(ctx, newIdentity, new(identity.ManagerOptions)))
				require.NoError(t, p.CreateIdentity(ctx, newIdentity))

				// Verify it was created with normalized credentials from the start
				created, err := p.GetIdentityConfidential(ctx, newIdentity.ID)
				require.NoError(t, err)

				// Should have normalized credentials
				if cred, ok := created.Credentials[identity.CredentialsTypePassword]; ok {
					require.Contains(t, cred.Identifiers, normalizedPhone, "New identity should have normalized phone")
				}

				// Should be findable by normalized phone
				found, cred, err := p.FindByCredentialsIdentifier(ctx, identity.CredentialsTypePassword, normalizedPhone)
				require.NoError(t, err)
				require.NotNil(t, found)
				require.Equal(t, newIdentity.ID, found.ID)
				require.NotNil(t, cred)
			})

			t.Run("legacy identities are upgraded in memory when loaded", func(t *testing.T) {
				// Use different phone numbers to avoid conflict with first sub-test
				legacyNonNormalizedPhone := "+49 30 1234 5678"
				legacyNormalizedPhone := "+493012345678"

				// Simulate legacy data: create identity with version 0 credential using raw SQL
				// This is necessary because CreateIdentity now normalizes identifiers at persistence layer
				var credTypes []identity.CredentialsTypeTable
				require.NoError(t, p.GetConnection(ctx).All(&credTypes))
				var passwordCredType identity.CredentialsTypeTable
				for _, ct := range credTypes {
					if ct.Name == identity.CredentialsTypePassword {
						passwordCredType = ct
						break
					}
				}

				legacyID := x.NewUUID()
				credID := x.NewUUID()
				identifierID := x.NewUUID()
				now := time.Now()
				traits := sqlxx.JSONRawMessage(`{"phone":"` + legacyNonNormalizedPhone + `","name":"jane doe"}`)

				// Cleanup
				t.Cleanup(func() {
					_ = p.GetConnection(ctx).RawQuery("DELETE FROM identity_credential_identifiers WHERE identity_credential_id = ?", credID).Exec()
					_ = p.GetConnection(ctx).RawQuery("DELETE FROM identity_credentials WHERE id = ?", credID).Exec()
					_ = p.GetConnection(ctx).RawQuery("DELETE FROM identities WHERE id = ?", legacyID).Exec()
				})

				// Insert legacy identity with raw SQL
				require.NoError(t, p.GetConnection(ctx).RawQuery(
					"INSERT INTO identities (id, nid, schema_id, traits, created_at, updated_at, state) VALUES (?, ?, ?, ?, ?, ?, ?)",
					legacyID, nid, phoneEmailSchema.ID, traits, now, now, identity.StateActive,
				).Exec())

				// Insert version 0 credential
				require.NoError(t, p.GetConnection(ctx).RawQuery(
					"INSERT INTO identity_credentials (id, identity_id, nid, identity_credential_type_id, created_at, updated_at, config, version) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
					credID, legacyID, nid, passwordCredType.ID, now, now, sqlxx.JSONRawMessage(`{"hashed_password":"$2a$10$..."}`), 0,
				).Exec())

				// Insert NON-NORMALIZED identifier only
				require.NoError(t, p.GetConnection(ctx).RawQuery(
					"INSERT INTO identity_credential_identifiers (id, identity_id, identity_credential_id, nid, identifier, created_at, updated_at, identity_credential_type_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
					identifierID, legacyID, credID, nid, legacyNonNormalizedPhone, now, now, passwordCredType.ID,
				).Exec())

				// Load the identity - this triggers UpgradeCredentials in memory
				loaded, err := p.GetIdentityConfidential(ctx, legacyID)
				require.NoError(t, err)
				require.Equal(t, 1, loaded.Credentials[identity.CredentialsTypePassword].Version, "Should be upgraded in memory")

				// In-memory upgrade adds normalized identifier
				identifiers := loaded.Credentials[identity.CredentialsTypePassword].Identifiers
				require.Len(t, identifiers, 1, "Should have single identifier in memory")
				require.Contains(t, identifiers, legacyNormalizedPhone)

				// Can find by non-normalized phone (what's in DB)
				found, cred, err := p.FindByCredentialsIdentifier(ctx, identity.CredentialsTypePassword, legacyNonNormalizedPhone)
				require.NoError(t, err, "Should find by non-normalized phone")
				require.NotNil(t, found)
				require.Equal(t, legacyID, found.ID)
				require.NotNil(t, cred)
			})

			t.Run("post-migration: legacy data normalized in DB", func(t *testing.T) {
				// Simulates the state after running `kratos migrate normalize-phone-numbers`:
				// Legacy data has been normalized in the DB, and all lookups should work with any input format.
				legacyNonNormalizedPhone := "+49 176 12345678"
				legacyNormalizedPhone := "+4917612345678"

				var credTypes []identity.CredentialsTypeTable
				require.NoError(t, p.GetConnection(ctx).All(&credTypes))
				var passwordCredType identity.CredentialsTypeTable
				for _, ct := range credTypes {
					if ct.Name == identity.CredentialsTypePassword {
						passwordCredType = ct
						break
					}
				}

				legacyID := x.NewUUID()
				credID := x.NewUUID()
				identifierID := x.NewUUID()
				now := time.Now()
				traits := sqlxx.JSONRawMessage(`{"phone":"` + legacyNonNormalizedPhone + `","name":"migrated user"}`)

				t.Cleanup(func() {
					_ = p.GetConnection(ctx).RawQuery("DELETE FROM identity_credential_identifiers WHERE identity_credential_id = ?", credID).Exec()
					_ = p.GetConnection(ctx).RawQuery("DELETE FROM identity_credentials WHERE id = ?", credID).Exec()
					_ = p.GetConnection(ctx).RawQuery("DELETE FROM identities WHERE id = ?", legacyID).Exec()
				})

				// Step 1: Insert legacy identity with non-normalized phone (pre-migration state)
				require.NoError(t, p.GetConnection(ctx).RawQuery(
					"INSERT INTO identities (id, nid, schema_id, traits, created_at, updated_at, state) VALUES (?, ?, ?, ?, ?, ?, ?)",
					legacyID, nid, phoneEmailSchema.ID, traits, now, now, identity.StateActive,
				).Exec())
				require.NoError(t, p.GetConnection(ctx).RawQuery(
					"INSERT INTO identity_credentials (id, identity_id, nid, identity_credential_type_id, created_at, updated_at, config, version) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
					credID, legacyID, nid, passwordCredType.ID, now, now, sqlxx.JSONRawMessage(`{"hashed_password":"$2a$10$..."}`), 0,
				).Exec())
				require.NoError(t, p.GetConnection(ctx).RawQuery(
					"INSERT INTO identity_credential_identifiers (id, identity_id, identity_credential_id, nid, identifier, created_at, updated_at, identity_credential_type_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
					identifierID, legacyID, credID, nid, legacyNonNormalizedPhone, now, now, passwordCredType.ID,
				).Exec())

				// Step 2: Simulate migration — normalize the identifier in DB (what the CLI tool does)
				require.NoError(t, p.GetConnection(ctx).RawQuery(
					"UPDATE identity_credential_identifiers SET identifier = ? WHERE id = ?",
					legacyNormalizedPhone, identifierID,
				).Exec())

				// Step 3: Verify lookups work with both formats after migration
				// Find by E.164 (normalized) — should match the migrated DB value
				found, cred, err := p.FindByCredentialsIdentifier(ctx, identity.CredentialsTypePassword, legacyNormalizedPhone)
				require.NoError(t, err, "should find migrated identity by E.164 phone")
				require.Equal(t, legacyID, found.ID)
				require.NotNil(t, cred)

				// Find by original non-normalized format — IN query normalizes input, matches migrated DB value
				found2, cred2, err := p.FindByCredentialsIdentifier(ctx, identity.CredentialsTypePassword, legacyNonNormalizedPhone)
				require.NoError(t, err, "should find migrated identity by non-normalized phone (IN query normalizes)")
				require.Equal(t, legacyID, found2.ID)
				require.NotNil(t, cred2)

				// Find by yet another format variant — also normalizes to same E.164
				found3, _, err := p.FindByCredentialsIdentifier(ctx, identity.CredentialsTypePassword, "+49 17612345678")
				require.NoError(t, err, "should find migrated identity by different format variant")
				require.Equal(t, legacyID, found3.ID)

				// Also test FindIdentityByCredentialIdentifier
				found4, err := p.FindIdentityByCredentialIdentifier(ctx, legacyNormalizedPhone, false, identity.ExpandDefault)
				require.NoError(t, err, "should find migrated identity by E.164 via FindIdentityByCredentialIdentifier")
				require.Equal(t, legacyID, found4.ID)

				found5, err := p.FindIdentityByCredentialIdentifier(ctx, legacyNonNormalizedPhone, false, identity.ExpandDefault)
				require.NoError(t, err, "should find migrated identity by non-normalized via FindIdentityByCredentialIdentifier")
				require.Equal(t, legacyID, found5.ID)
			})
		})

		var createdIDs []uuid.UUID
		passwordIdentity := func(schemaID string, credentialsID string) *identity.Identity {
			i := identity.NewIdentity(schemaID)
			i.SetCredentials(identity.CredentialsTypePassword, identity.Credentials{
				Type: identity.CredentialsTypePassword, Identifiers: []string{credentialsID},
				Config: sqlxx.JSONRawMessage(`{"foo":"bar"}`),
			})
			return i
		}

		oidcIdentity := func(schemaID string, credentialsID string) *identity.Identity {
			i := identity.NewIdentity(schemaID)
			i.SetCredentials(identity.CredentialsTypeOIDC, identity.Credentials{
				Type: identity.CredentialsTypeOIDC, Identifiers: []string{credentialsID},
				Config: sqlxx.JSONRawMessage(`{}`),
			})
			return i
		}

		assertEqual := func(t *testing.T, expected, actual *identity.Identity) {
			assert.Empty(t, actual.Credentials)
			require.Equal(t, expected.Traits, actual.Traits)
			require.Equal(t, expected.ID, actual.ID)
		}

		t.Run("case=should create and set missing ID", func(t *testing.T) {
			i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			i.SetCredentials(identity.CredentialsTypeOIDC, identity.Credentials{
				Type: identity.CredentialsTypeOIDC, Identifiers: []string{x.NewUUID().String()},
				Config: sqlxx.JSONRawMessage(`{}`),
			})
			i.ID = uuid.Nil
			require.NoError(t, p.CreateIdentity(ctx, i))
			assert.NotEqual(t, uuid.Nil, i.ID)
			assert.Equal(t, nid, i.NID)
			createdIDs = append(createdIDs, i.ID)

			count, err := p.CountIdentities(ctx)
			require.NoError(t, err)
			assert.EqualValues(t, int64(1), count)

			t.Run("different network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, ctx, p)
				count, err := p.CountIdentities(ctx)
				require.NoError(t, err)
				assert.EqualValues(t, int64(0), count)
			})
		})

		t.Run("case=create with default values", func(t *testing.T) {
			expected := passwordIdentity("", x.NewUUID().String())
			require.NoError(t, p.CreateIdentity(ctx, expected))
			createdIDs = append(createdIDs, expected.ID)

			actual, err := p.GetIdentity(ctx, expected.ID, identity.ExpandDefault)
			require.NoError(t, err)

			assert.Equal(t, expected.ID, actual.ID)
			assert.Equal(t, config.DefaultIdentityTraitsSchemaID, actual.SchemaID)
			assert.Equal(t, defaultSchema.SchemaURL(exampleServerURL).String(), actual.SchemaURL)
			assertEqual(t, expected, actual)

			count, err := p.CountIdentities(ctx)
			require.NoError(t, err)
			assert.EqualValues(t, 2, count)

			t.Run("different network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, ctx, p)
				_, err := p.GetIdentity(ctx, expected.ID, identity.ExpandDefault)
				require.ErrorIs(t, err, sqlcon.ErrNoRows())

				count, err := p.CountIdentities(ctx)
				require.NoError(t, err)
				assert.EqualValues(t, int64(0), count)
			})
		})

		t.Run("case=should set external ID", func(t *testing.T) {
			i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			i.SetCredentials(identity.CredentialsTypeOIDC, identity.Credentials{
				Type: identity.CredentialsTypeOIDC, Identifiers: []string{x.NewUUID().String()},
				Config: sqlxx.JSONRawMessage(`{}`),
			})
			i.ID = uuid.Nil
			externalID := sqlxx.NullString("external-id-" + randx.MustString(10, randx.AlphaNum))
			i.ExternalID = externalID
			require.NoError(t, p.CreateIdentity(ctx, i))
			assert.NotEqual(t, uuid.Nil, i.ID)
			assert.Equal(t, nid, i.NID)
			assert.Equal(t, externalID, i.ExternalID)
			createdIDs = append(createdIDs, i.ID)

			t.Run("find by external ID", func(t *testing.T) {
				i2, err := p.FindIdentityByExternalID(ctx, externalID.String(), identity.ExpandEverything)
				require.NoError(t, err)
				assert.Equal(t, i.ID, i2.ID)
			})

			t.Run("must be unique", func(t *testing.T) {
				i2 := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
				i2.SetCredentials(identity.CredentialsTypeOIDC, identity.Credentials{
					Type: identity.CredentialsTypeOIDC, Identifiers: []string{x.NewUUID().String()},
					Config: sqlxx.JSONRawMessage(`{}`),
				})
				i2.ExternalID = externalID

				err := new(herodot.DefaultError)
				require.ErrorAs(t, p.CreateIdentity(ctx, i2), &err)
				assert.Equal(t, http.StatusConflict, err.CodeField)
			})
		})

		t.Run("case=create with null AAL", func(t *testing.T) {
			expected := passwordIdentity("", "id-"+uuid.Must(uuid.NewV4()).String())
			expected.InternalAvailableAAL.Valid = false
			require.NoError(t, p.CreateIdentity(ctx, expected))
			createdIDs = append(createdIDs, expected.ID)

			actual, err := p.GetIdentity(ctx, expected.ID, identity.ExpandDefault)
			require.NoError(t, err)

			assert.False(t, actual.InternalAvailableAAL.Valid)
		})

		t.Run("suite=create multiple identities", func(t *testing.T) {
			t.Run("create multiple identities", func(t *testing.T) {
				identities := make([]*identity.Identity, 100)
				for i := range identities {
					identities[i] = NewTestIdentity(4, "persister-create-multiple", i)
				}
				require.NoError(t, p.CreateIdentities(ctx, identities...))
				createdAt := time.Now().UTC()

				for _, id := range identities {
					idFromDB, err := p.GetIdentity(ctx, id.ID, identity.ExpandEverything)
					require.NoError(t, err)

					credFromDB := idFromDB.Credentials[identity.CredentialsTypePassword]
					assert.Equal(t, id.ID, idFromDB.ID)
					assert.Equal(t, id.SchemaID, idFromDB.SchemaID)
					assert.Equal(t, id.SchemaURL, idFromDB.SchemaURL)
					assert.Equal(t, id.State, idFromDB.State)

					// We test that the values are plausible in the handler test already.
					assert.Equal(t, len(id.VerifiableAddresses), len(idFromDB.VerifiableAddresses))
					assert.Equal(t, len(id.RecoveryAddresses), len(idFromDB.RecoveryAddresses))

					assert.Equal(t, id.Credentials["password"].Identifiers, credFromDB.Identifiers)
					assert.WithinDuration(t, createdAt, credFromDB.CreatedAt, time.Minute)
					assert.WithinDuration(t, createdAt, credFromDB.UpdatedAt, time.Minute)
					// because of mysql precision
					assert.WithinDuration(t, id.CreatedAt, idFromDB.CreatedAt, time.Second)
					assert.WithinDuration(t, id.UpdatedAt, idFromDB.UpdatedAt, time.Second)

					require.NoError(t, p.DeleteIdentity(ctx, id.ID))
				}
			})

			t.Run("create exactly the non-conflicting ones", func(t *testing.T) {
				identities := make([]*identity.Identity, 100)
				for i := range identities {
					identities[i] = NewTestIdentity(4, "persister-create-multiple-2", i%60)
				}
				err := p.CreateIdentities(ctx, identities...)
				if dbname == "mysql" {
					// partial inserts are not supported on mysql
					assert.ErrorIs(t, err, sqlcon.ErrUniqueViolation())
					return
				}
				createdAt := time.Now().UTC()

				errWithCtx := new(identity.CreateIdentitiesError)
				require.ErrorAsf(t, err, &errWithCtx, "%#v", err)

				for _, id := range identities[:60] {
					require.NotZero(t, id.ID)

					idFromDB, err := p.GetIdentity(ctx, id.ID, identity.ExpandEverything)
					require.NoError(t, err)

					credFromDB := idFromDB.Credentials[identity.CredentialsTypePassword]
					assert.Equal(t, id.ID, idFromDB.ID)
					assert.Equal(t, id.SchemaID, idFromDB.SchemaID)
					assert.Equal(t, id.SchemaURL, idFromDB.SchemaURL)
					assert.Equal(t, id.State, idFromDB.State)

					// We test that the values are plausible in the handler test already.
					assert.Equal(t, len(id.VerifiableAddresses), len(idFromDB.VerifiableAddresses))
					assert.Equal(t, len(id.RecoveryAddresses), len(idFromDB.RecoveryAddresses))

					assert.Equal(t, id.Credentials["password"].Identifiers, credFromDB.Identifiers)
					assert.WithinDuration(t, createdAt, credFromDB.CreatedAt, time.Minute)
					assert.WithinDuration(t, createdAt, credFromDB.UpdatedAt, time.Minute)
					// because of mysql precision
					assert.WithinDuration(t, id.CreatedAt, idFromDB.CreatedAt, time.Second)
					assert.WithinDuration(t, id.UpdatedAt, idFromDB.UpdatedAt, time.Second)

					require.NoError(t, p.DeleteIdentity(ctx, id.ID))
				}

				for _, id := range identities[60:] {
					failed := errWithCtx.Find(id)
					assert.NotNil(t, failed)
				}
			})
		})

		t.Run("case=external_id conflict in batch create returns partial error", func(t *testing.T) {
			// Regression test for https://github.com/ory-corp/cloud/issues/10580
			// When all identities in a batch conflict on external_id, CreateIdentities
			// should return a CreateIdentitiesError, not a raw SQL error.
			// On Postgres, the bug causes: ERROR: syntax error at or near ")" (SQLSTATE 42601)
			// because DeleteIdentities is called with an empty ID slice, generating IN ().

			// Step 1: Create two identities with external_ids.
			first := make([]*identity.Identity, 2)
			for i := range first {
				first[i] = NewTestIdentity(1, "ext-id-conflict-first", i)
				first[i].ExternalID = sqlxx.NullString(fmt.Sprintf("ext-conflict-pool-%d", i))
			}
			require.NoError(t, p.CreateIdentities(ctx, first...))
			for _, id := range first {
				createdIDs = append(createdIDs, id.ID)
			}

			// Step 2: Create new identities with different traits but same external_ids.
			second := make([]*identity.Identity, 2)
			for i := range second {
				second[i] = NewTestIdentity(1, "ext-id-conflict-second", i)
				second[i].ExternalID = sqlxx.NullString(fmt.Sprintf("ext-conflict-pool-%d", i))
			}
			err := p.CreateIdentities(ctx, second...)
			if dbname == "mysql" {
				assert.ErrorIs(t, err, sqlcon.ErrUniqueViolation())
				return
			}
			errWithCtx := new(identity.CreateIdentitiesError)
			require.ErrorAsf(t, err, &errWithCtx, "expected CreateIdentitiesError, got: %#v", err)
			for _, id := range second {
				assert.NotNil(t, errWithCtx.Find(id), "expected identity %s to be in the error", id.ID)
			}
		})

		t.Run("case=should error when the identity ID does not exist", func(t *testing.T) {
			_, err := p.GetIdentity(ctx, uuid.UUID{}, identity.ExpandNothing)
			require.Error(t, err)

			_, err = p.GetIdentity(ctx, x.NewUUID(), identity.ExpandNothing)
			require.Error(t, err)

			_, err = p.GetIdentityConfidential(ctx, x.NewUUID())
			require.Error(t, err)
		})

		t.Run("case=run migrations when fetching credentials", func(t *testing.T) {
			expected := func(schemaID string, credentialsID string) *identity.Identity {
				i := identity.NewIdentity(schemaID)
				i.SetCredentials(identity.CredentialsTypeWebAuthn, identity.Credentials{
					Type: identity.CredentialsTypeWebAuthn, Identifiers: []string{credentialsID},
					Config: sqlxx.JSONRawMessage(`{"credentials":[{}]}`),
				})
				return i
			}(altSchema.ID, "webauthn")
			require.NoError(t, p.CreateIdentity(ctx, expected))
			createdIDs = append(createdIDs, expected.ID)

			actual, err := p.GetIdentityConfidential(ctx, expected.ID)
			require.NoError(t, err)
			c := actual.GetCredentialsOr(identity.CredentialsTypeWebAuthn, &identity.Credentials{})
			assert.True(t, gjson.GetBytes(c.Config, "credentials.0.is_passwordless").Exists())
			assert.Equal(t, base64.StdEncoding.EncodeToString(expected.ID[:]), gjson.GetBytes(c.Config, "user_handle").String())
		})

		t.Run("case=create and keep set values", func(t *testing.T) {
			expected := passwordIdentity(altSchema.ID, x.NewUUID().String())
			require.NoError(t, p.CreateIdentity(ctx, expected))
			createdIDs = append(createdIDs, expected.ID)

			actual, err := p.GetIdentity(ctx, expected.ID, identity.ExpandDefault)
			require.NoError(t, err)
			assert.Equal(t, altSchema.ID, actual.SchemaID)
			assert.Equal(t, altSchema.SchemaURL(exampleServerURL).String(), actual.SchemaURL)
			assertEqual(t, expected, actual)

			actual, err = p.GetIdentityConfidential(ctx, expected.ID)
			require.NoError(t, err)
			require.Equal(t, expected.Traits, actual.Traits)
			require.Equal(t, expected.ID, actual.ID)

			assert.NotEmpty(t, actual.Credentials)
			assert.NotEmpty(t, expected.Credentials)

			for m, expected := range expected.Credentials {
				assert.Equal(t, expected.ID, actual.Credentials[m].ID)
				assert.JSONEq(t, string(expected.Config), string(actual.Credentials[m].Config))
				assert.Equal(t, expected.Identifiers, actual.Credentials[m].Identifiers)
				assert.Equal(t, expected.Type, actual.Credentials[m].Type)
			}

			t.Run("different network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, ctx, p)
				_, err := p.GetIdentity(ctx, expected.ID, identity.ExpandNothing)
				require.ErrorIs(t, err, sqlcon.ErrNoRows())

				_, err = p.GetIdentityConfidential(ctx, expected.ID)
				require.ErrorIs(t, err, sqlcon.ErrNoRows())
			})
		})

		t.Run("case=fail on duplicate credential identifiers if type is password", func(t *testing.T) {
			email := randx.MustString(16, randx.AlphaLowerNum) + "@bar.com"
			initial := passwordIdentity("", email)
			require.NoError(t, p.CreateIdentity(ctx, initial))
			createdIDs = append(createdIDs, initial.ID)

			for _, transform := range []func(string) string{
				strings.ToLower,
				func(s string) string { return s[:1] + strings.ToUpper(s[1:2]) + s[2:] },
				strings.ToUpper,
				func(s string) string { left, right, _ := strings.Cut(s, "@"); return left + "@" + strings.Title(right) },
			} {
				ids := transform(email)
				expected := passwordIdentity("", ids)
				err := p.CreateIdentity(ctx, expected)
				require.ErrorIs(t, err, sqlcon.ErrUniqueViolation(), "%+v", err)

				_, err = p.GetIdentity(ctx, expected.ID, identity.ExpandNothing)
				require.Error(t, err)

				t.Run("succeeds on different network/id="+ids, func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					expected := passwordIdentity("", ids)
					err := p.CreateIdentity(ctx, expected)
					require.NoError(t, err)

					_, err = p.GetIdentity(ctx, expected.ID, identity.ExpandNothing)
					require.NoError(t, err)
				})
			}
		})

		t.Run("case=fail on duplicate credential identifiers if type is oidc", func(t *testing.T) {
			oidcID := randx.MustString(16, randx.AlphaLowerNum)
			initial := oidcIdentity("", oidcID)
			require.NoError(t, p.CreateIdentity(ctx, initial))
			createdIDs = append(createdIDs, initial.ID)

			expected := oidcIdentity("", oidcID)
			require.Error(t, p.CreateIdentity(ctx, expected))

			_, err := p.GetIdentity(ctx, expected.ID, identity.ExpandNothing)
			require.Error(t, err)

			second := oidcIdentity("", strings.ToUpper(oidcID))
			require.NoError(t, p.CreateIdentity(ctx, second), "should work because oidc is not case-sensitive")
			createdIDs = append(createdIDs, second.ID)

			t.Run("succeeds on different network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, ctx, p)
				expected := oidcIdentity("", oidcID)
				require.NoError(t, p.CreateIdentity(ctx, expected))

				_, err = p.GetIdentity(ctx, expected.ID, identity.ExpandNothing)
				require.NoError(t, err)
			})
		})

		t.Run("case=create with invalid traits data", func(t *testing.T) {
			expected := oidcIdentity("", x.NewUUID().String())
			expected.Traits = identity.Traits(`{"bar":123}`) // bar should be a string
			err := p.CreateIdentity(ctx, expected)
			require.Error(t, err)
			assert.Contains(t, fmt.Sprintf("%+v", err.Error()), "malformed")
		})

		t.Run("case=get classified credentials", func(t *testing.T) {
			initial := oidcIdentity("", x.NewUUID().String())
			initial.SetCredentials(identity.CredentialsTypeOIDC, identity.Credentials{
				Type: identity.CredentialsTypeOIDC, Identifiers: []string{"aylmao-oidc"},
				Config: sqlxx.JSONRawMessage(`{"ay":"lmao"}`),
			})
			require.NoError(t, p.CreateIdentity(ctx, initial))
			createdIDs = append(createdIDs, initial.ID)

			initial, err := p.GetIdentityConfidential(ctx, initial.ID)
			require.NoError(t, err)
			require.NotEqual(t, uuid.Nil, initial.ID)
			require.NotEmpty(t, initial.Credentials)

			t.Run("fails on different network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, ctx, p)
				_, err := p.GetIdentityConfidential(ctx, initial.ID)
				require.ErrorIs(t, err, sqlcon.ErrNoRows())
			})
		})

		t.Run("case=update an identity and set credentials", func(t *testing.T) {
			initial := oidcIdentity("", x.NewUUID().String())
			require.NoError(t, p.CreateIdentity(ctx, initial))
			createdIDs = append(createdIDs, initial.ID)

			assert.Equal(t, config.DefaultIdentityTraitsSchemaID, initial.SchemaID)
			assert.Equal(t, defaultSchema.SchemaURL(exampleServerURL).String(), initial.SchemaURL)

			expected := initial.CopyWithoutCredentials()
			expected.SetCredentials(identity.CredentialsTypePassword, identity.Credentials{
				Type:        identity.CredentialsTypePassword,
				Identifiers: []string{"ignore-me"},
				Config:      sqlxx.JSONRawMessage(`{"oh":"nono"}`),
			})
			expected.Traits = identity.Traits(`{"update":"me"}`)
			expected.SchemaID = altSchema.ID
			require.NoError(t, p.UpdateIdentity(ctx, expected))

			actual, err := p.GetIdentityConfidential(ctx, expected.ID)
			require.NoError(t, err)
			assert.Equal(t, altSchema.ID, actual.SchemaID)
			assert.Equal(t, altSchema.SchemaURL(exampleServerURL).String(), actual.SchemaURL)
			assert.NotEmpty(t, actual.Credentials[identity.CredentialsTypePassword])
			assert.Empty(t, actual.Credentials[identity.CredentialsTypeOIDC])

			assert.Equal(t, expected.Credentials[identity.CredentialsTypeOIDC], actual.Credentials[identity.CredentialsTypeOIDC])

			t.Run("fails on different network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, ctx, p)
				require.ErrorIs(t, p.UpdateIdentity(ctx, expected), sqlcon.ErrNoRows())
			})
		})

		t.Run("case=fail to update because validation fails", func(t *testing.T) {
			initial := oidcIdentity("", x.NewUUID().String())

			require.NoError(t, p.CreateIdentity(ctx, initial))
			createdIDs = append(createdIDs, initial.ID)

			initial.Traits = identity.Traits(`{"bar":123}`)
			err := p.UpdateIdentity(ctx, initial)
			require.Error(t, err)
			require.Contains(t, err.Error(), "malformed")
		})

		t.Run("case=update an identity column", func(t *testing.T) {
			initial := oidcIdentity("", x.NewUUID().String())
			initial.InternalAvailableAAL = identity.NewNullableAuthenticatorAssuranceLevel(identity.NoAuthenticatorAssuranceLevel)
			require.NoError(t, p.CreateIdentity(ctx, initial))
			createdIDs = append(createdIDs, initial.ID)

			initial.InternalAvailableAAL = identity.NewNullableAuthenticatorAssuranceLevel(identity.AuthenticatorAssuranceLevel1)
			initial.State = identity.StateInactive
			require.NoError(t, p.UpdateIdentityColumns(ctx, initial, "available_aal"))

			actual, err := p.GetIdentity(ctx, initial.ID, identity.ExpandDefault)
			require.NoError(t, err)
			assert.Equal(t, string(identity.AuthenticatorAssuranceLevel1), actual.InternalAvailableAAL.String)
			assert.Equal(t, identity.StateActive, actual.State, "the state remains unchanged")
		})

		t.Run("case=should fail to insert identity because credentials from traits exist", func(t *testing.T) {
			email := randx.MustString(16, randx.AlphaLowerNum) + "@ory.sh"
			first := passwordIdentity("", email)
			first.Traits = identity.Traits(`{}`)
			require.NoError(t, p.CreateIdentity(ctx, first))
			createdIDs = append(createdIDs, first.ID)

			second := passwordIdentity("", email)
			require.Error(t, p.CreateIdentity(ctx, second))

			t.Run("passes on different network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, ctx, p)
				require.NoError(t, p.CreateIdentity(ctx, second))
			})

			t.Run("case=should fail to update identity because credentials exist", func(t *testing.T) {
				first := passwordIdentity("", x.NewUUID().String())
				first.Traits = identity.Traits(`{}`)
				require.NoError(t, p.CreateIdentity(ctx, first))
				createdIDs = append(createdIDs, first.ID)

				c := first.Credentials[identity.CredentialsTypePassword]
				c.Identifiers = []string{email}
				first.Credentials[identity.CredentialsTypePassword] = c
				require.Error(t, p.UpdateIdentity(ctx, first))

				t.Run("passes on different network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					first := passwordIdentity("", x.NewUUID().String())
					first.Traits = identity.Traits(`{}`)
					require.NoError(t, p.CreateIdentity(ctx, first))

					c := first.Credentials[identity.CredentialsTypePassword]
					c.Identifiers = []string{"test-identity@ory.sh"}
					first.Credentials[identity.CredentialsTypePassword] = c
					require.NoError(t, p.UpdateIdentity(ctx, first))
				})
			})
		})

		t.Run("case=should succeed to update credentials from traits", func(t *testing.T) {
			expected := passwordIdentity("", x.NewUUID().String())
			require.NoError(t, p.CreateIdentity(ctx, expected))
			createdIDs = append(createdIDs, expected.ID)

			expected.Traits = identity.Traits(`{"email":"update-test-identity@ory.sh"}`)
			require.NoError(t, p.UpdateIdentity(ctx, expected))

			actual, err := p.GetIdentityConfidential(ctx, expected.ID)
			require.NoError(t, err)

			assert.Equal(t, expected.Credentials[identity.CredentialsTypePassword].Identifiers, actual.Credentials[identity.CredentialsTypePassword].Identifiers)
		})

		t.Run("case=delete an identity", func(t *testing.T) {
			expected := passwordIdentity("", x.NewUUID().String())
			require.NoError(t, p.CreateIdentity(ctx, expected))

			t.Run("fails on different network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, ctx, p)
				require.ErrorIs(t, p.DeleteIdentity(ctx, expected.ID), sqlcon.ErrNoRows())

				p = testhelpers.ExistingNetwork(t, p, nid)
				_, err := p.GetIdentity(ctx, expected.ID, identity.ExpandNothing)
				require.NoError(t, err)
			})

			require.NoError(t, p.DeleteIdentity(ctx, expected.ID))

			_, err := p.GetIdentity(ctx, expected.ID, identity.ExpandNothing)
			require.Error(t, err)
		})

		t.Run("case=create with empty credentials config", func(t *testing.T) {
			// This test covers a case where the config value of a credentials setting is empty. This causes
			// issues with postgres' json field.
			expected := passwordIdentity("", x.NewUUID().String())
			expected.SetCredentials(identity.CredentialsTypePassword, identity.Credentials{
				Type:        identity.CredentialsTypePassword,
				Identifiers: []string{"id-missing-creds-config"},
				Config:      sqlxx.JSONRawMessage(``),
			})
			require.NoError(t, p.CreateIdentity(ctx, expected))
			createdIDs = append(createdIDs, expected.ID)
		})

		t.Run("case=list", func(t *testing.T) {
			is, _, err := p.ListIdentities(ctx, identity.ListIdentityParameters{Expand: identity.ExpandDefault})
			require.NoError(t, err)
			require.NotEmpty(t, is)
			require.Len(t, is, len(createdIDs))
			for _, id := range createdIDs {
				var found bool
				for _, i := range is {
					if i.ID == id {
						expected, err := p.GetIdentity(ctx, id, identity.ExpandDefault)
						require.NoError(t, err)
						assertx.EqualAsJSON(t, expected, i)
						found = true
					}
				}
				require.True(t, found, id)
			}

			t.Run("no results on other network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, ctx, p)
				is, _, err := p.ListIdentities(ctx, identity.ListIdentityParameters{Expand: identity.ExpandDefault})
				require.NoError(t, err)
				assert.Len(t, is, 0)
			})

			t.Run("list some using ids filter", func(t *testing.T) {
				filterIds := createdIDs[:2]

				is, _, err := p.ListIdentities(ctx, identity.ListIdentityParameters{Expand: identity.ExpandDefault, IdsFilter: filterIds})
				require.NoError(t, err)
				assert.Len(t, is, len(filterIds))
			})

			t.Run("eventually consistent", func(t *testing.T) {
				if dbname != "cockroach" {
					t.Skipf("Test only works with cockroachdb")
					return
				}

				id := x.NewUUID().String()
				another := oidcIdentity("", id)
				require.NoError(t, p.CreateIdentity(ctx, another))
				createdIDs = append(createdIDs, another.ID)

				is, _, err := p.ListIdentities(ctx, identity.ListIdentityParameters{
					Expand:           identity.ExpandDefault,
					KeySetPagination: []keysetpagination.Option{keysetpagination.WithSize(25)},
					ConsistencyLevel: crdbx.ConsistencyLevelStrong,
				})
				require.NoError(t, err)
				require.Len(t, is, len(createdIDs))

				var results []identity.Identity
				// It takes about 4.8 seconds to replicate the data.
				for i := 0; i < 8; i++ {
					time.Sleep(time.Second)

					// The error here is explicitly ignored because the table / schema might not yet be replicated.
					// This can lead to "ERROR: cached plan must not change result type (SQLSTATE 0A000)" whih is caused
					// because the prepared query exist but the schema is not yet replicated.
					is, _, _ := p.ListIdentities(ctx, identity.ListIdentityParameters{
						Expand:           identity.ExpandEverything,
						KeySetPagination: []keysetpagination.Option{keysetpagination.WithSize(25)},
						ConsistencyLevel: crdbx.ConsistencyLevelEventual,
					})

					if len(is) == len(createdIDs) {
						results = is
					}
				}
				require.NotZero(t, len(results))
				require.Len(t, results, len(createdIDs), "Could not find all identities after 8 seconds")

				var found bool
				for _, i := range results {
					if i.ID == another.ID {
						found = true
					}
				}
				require.True(t, found, id, "Unable to find created identity in eventually consistent results.")
			})
		})

		t.Run("case=find identity by its credentials identifier", func(t *testing.T) {
			var expectedIdentifiers []string
			var expectedIdentities []*identity.Identity

			for _, c := range []identity.CredentialsType{
				identity.CredentialsTypePassword,
				identity.CredentialsTypeWebAuthn,
				identity.CredentialsTypeOIDC,
			} {
				identityIdentifier := fmt.Sprintf("find-identity-by-identifier-%s@ory.sh", c)
				expected := identity.NewIdentity("")
				expected.SetCredentials(c, identity.Credentials{Type: c, Identifiers: []string{identityIdentifier}, Config: sqlxx.JSONRawMessage(`{}`)})

				require.NoError(t, p.CreateIdentity(ctx, expected))
				createdIDs = append(createdIDs, expected.ID)
				expectedIdentifiers = append(expectedIdentifiers, identityIdentifier)
				expectedIdentities = append(expectedIdentities, expected)
			}

			create := identity.NewIdentity("")
			create.SetCredentials(identity.CredentialsTypePassword, identity.Credentials{Type: identity.CredentialsTypePassword, Identifiers: []string{"find-identity-by-identifier-common@ory.sh"}, Config: sqlxx.JSONRawMessage(`{}`)})
			create.SetCredentials(identity.CredentialsTypeWebAuthn, identity.Credentials{Type: identity.CredentialsTypeWebAuthn, Identifiers: []string{"find-identity-by-identifier-common@ory.sh"}, Config: sqlxx.JSONRawMessage(`{}`)})
			require.NoError(t, p.CreateIdentity(ctx, create))

			actual, _, err := p.ListIdentities(ctx, identity.ListIdentityParameters{
				Expand: identity.ExpandEverything,
			})
			require.NoError(t, err)
			require.Greater(t, len(actual), 0)

			for c, ct := range []identity.CredentialsType{
				identity.CredentialsTypePassword,
				identity.CredentialsTypeWebAuthn,
			} {
				t.Run(ct.String(), func(t *testing.T) {
					actual, _, err := p.ListIdentities(ctx, identity.ListIdentityParameters{
						// Match is normalized
						CredentialsIdentifier: expectedIdentifiers[c],
					})
					require.NoError(t, err)

					expected := expectedIdentities[c]
					require.Len(t, actual, 1)
					assertx.EqualAsJSONExcept(t, expected, actual[0], []string{"credentials.config", "created_at", "updated_at", "state_changed_at"})
				})
			}

			t.Run("similarity search", func(t *testing.T) {
				actual, _, err := p.ListIdentities(ctx, identity.ListIdentityParameters{
					CredentialsIdentifierSimilar: "find-identity-by-identifier",
					Expand:                       identity.ExpandCredentials,
				})
				require.NoError(t, err)
				assert.Len(t, actual, 4) // webauthn, common, password, oidc

			outer:
				for _, e := range append(expectedIdentities[:2], create) {
					for _, a := range actual {
						if e.ID == a.ID {
							assertx.EqualAsJSONExcept(t, e, a, []string{"credentials.config", "created_at", "updated_at", "state_changed_at"})
							continue outer
						}
					}
					actualCredentials := make([]map[identity.CredentialsType]identity.Credentials, len(actual))
					for k, a := range actual {
						actualCredentials[k] = a.Credentials
					}
					t.Fatalf("expected identity %+v not found in actual result set %+v", e.Credentials, actualCredentials)
				}
			})

			t.Run("find by OIDC identifier", func(t *testing.T) {
				actual, next, err := p.ListIdentities(ctx, identity.ListIdentityParameters{
					CredentialsIdentifier: "find-identity-by-identifier-oidc@ory.sh",
					Expand:                identity.ExpandEverything,
				})
				require.NoError(t, err)
				assert.Len(t, actual, 1)
				assert.True(t, next.IsLast())
			})

			t.Run("one result set even if multiple matches", func(t *testing.T) {
				actual, next, err := p.ListIdentities(ctx, identity.ListIdentityParameters{
					CredentialsIdentifier: "find-identity-by-identifier-common@ory.sh",
					Expand:                identity.ExpandEverything,
				})
				require.NoError(t, err)
				assert.Len(t, actual, 1)
				assert.True(t, next.IsLast())
			})

			t.Run("non existing identifier", func(t *testing.T) {
				actual, next, err := p.ListIdentities(ctx, identity.ListIdentityParameters{
					CredentialsIdentifier: "find-identity-by-identifier-non-existing@ory.sh",
					Expand:                identity.ExpandEverything,
				})
				require.NoError(t, err)
				assert.Len(t, actual, 0)
				assert.True(t, next.IsLast())
			})

			t.Run("not if on another network", func(t *testing.T) {
				_, on := testhelpers.NewNetwork(t, ctx, p)
				actual, next, err := on.ListIdentities(ctx, identity.ListIdentityParameters{
					CredentialsIdentifier: expectedIdentifiers[0],
					Expand:                identity.ExpandEverything,
				})
				require.NoError(t, err)
				assert.Len(t, actual, 0)
				assert.True(t, next.IsLast())
			})
		})

		t.Run("case=find identity by its credentials type and identifier", func(t *testing.T) {
			email := randx.MustString(16, randx.AlphaLowerNum) + "@ory.sh"
			expected := passwordIdentity("", email)
			expected.Traits = identity.Traits(`{}`)

			require.NoError(t, p.CreateIdentity(ctx, expected))
			createdIDs = append(createdIDs, expected.ID)

			actual, creds, err := p.FindByCredentialsIdentifier(ctx, identity.CredentialsTypePassword, email)
			require.NoError(t, err)

			assert.EqualValues(t, expected.Credentials[identity.CredentialsTypePassword].ID, creds.ID)
			assert.EqualValues(t, expected.Credentials[identity.CredentialsTypePassword].Identifiers, creds.Identifiers)
			assert.JSONEq(t, string(expected.Credentials[identity.CredentialsTypePassword].Config), string(creds.Config))
			// assert.EqualValues(t, expected.Credentials[CredentialsTypePassword].CreatedAt.Unix(), creds.CreatedAt.Unix())
			// assert.EqualValues(t, expected.Credentials[CredentialsTypePassword].UpdatedAt.Unix(), creds.UpdatedAt.Unix())

			require.Equal(t, expected.Traits, actual.Traits)
			require.Equal(t, expected.ID, actual.ID)
			require.NotNil(t, actual.Credentials[identity.CredentialsTypePassword])
			assert.EqualValues(t, expected.Credentials[identity.CredentialsTypePassword].ID, actual.Credentials[identity.CredentialsTypePassword].ID)
			assert.EqualValues(t, expected.Credentials[identity.CredentialsTypePassword].Identifiers, actual.Credentials[identity.CredentialsTypePassword].Identifiers)
			assert.JSONEq(t, string(expected.Credentials[identity.CredentialsTypePassword].Config), string(actual.Credentials[identity.CredentialsTypePassword].Config))

			t.Run("not if on another network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, ctx, p)
				_, _, err := p.FindByCredentialsIdentifier(ctx, identity.CredentialsTypePassword, email)
				require.ErrorIs(t, err, sqlcon.ErrNoRows())
			})
		})

		t.Run("case=find identity by its webauthn credential user handle", func(t *testing.T) {
			expected := identity.NewIdentity("")
			expected.SetCredentials(identity.CredentialsTypeWebAuthn, identity.Credentials{
				Type:        identity.CredentialsTypeWebAuthn,
				Identifiers: []string{"find-webauth-user-handle-identifier@ory.sh"},
				Config: sqlxx.JSONRawMessage(`{
  "credentials": [
    {
      "added_at": "2024-02-13T10:36:16Z",
      "attestation_type": "none",
      "authenticator": {
        "aaguid": "+/wwBxVOTsyMC24CBVfXvQ==",
        "clone_warning": false,
        "sign_count": 0
      },
      "display_name": "Yubikey",
      "id": "f2uGd/Bg1rGcGXtYp4MT4WcN+eA=",
      "is_passwordless": true,
      "public_key": "pQECAyYgASFYIBkNvUxvjdhuA36FworTmS/rxZR1I+NyRWBpoTYY/R+CIlggw+gFFrFoEi+rS82zq7+tDHAukBUJcFpQ7Z3NLBZH5vk="
    }
  ],
  "user_handle": "51z80nYJTSGmr6UBe1VGLg=="
}`),
			})
			expected.Traits = identity.Traits(`{}`)
			userHandle := x.Must(base64.StdEncoding.DecodeString("51z80nYJTSGmr6UBe1VGLg=="))

			require.NoError(t, p.CreateIdentity(ctx, expected))
			createdIDs = append(createdIDs, expected.ID)

			actual, err := p.FindIdentityByWebauthnUserHandle(ctx, userHandle)
			require.NoError(t, err)

			expected.Credentials = nil
			assertEqual(t, expected, actual)

			t.Run("not if on another network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, ctx, p)
				_, err = p.FindIdentityByWebauthnUserHandle(ctx, userHandle)
				require.ErrorIs(t, err, sqlcon.ErrNoRows())
			})
		})

		t.Run("case=find identity only by credentials identifier", func(t *testing.T) {
			email := randx.MustString(16, randx.AlphaLowerNum) + "@ory.sh"
			expected := passwordIdentity("", email)
			expected.Traits = identity.Traits(`{}`)

			require.NoError(t, p.CreateIdentity(ctx, expected))
			createdIDs = append(createdIDs, expected.ID)

			actual, err := p.FindIdentityByCredentialIdentifier(ctx, strings.ToUpper(email), false, identity.ExpandDefault)
			require.NoError(t, err)

			expected.Credentials = nil
			assertEqual(t, expected, actual)

			t.Run("not if on another network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, ctx, p)
				_, err := p.FindIdentityByCredentialIdentifier(ctx, strings.ToUpper(email), false, identity.ExpandDefault)
				require.ErrorIs(t, err, sqlcon.ErrNoRows())
			})
		})

		t.Run("case=find identity only by credentials identifier case sensitive", func(t *testing.T) {
			email := randx.MustString(16, randx.AlphaLowerNum) + "@ory.sh"
			expected := passwordIdentity("", email)
			expected.Traits = identity.Traits(`{}`)

			require.NoError(t, p.CreateIdentity(ctx, expected))
			createdIDs = append(createdIDs, expected.ID)

			_, err := p.FindIdentityByCredentialIdentifier(ctx, strings.ToUpper(email), true, identity.ExpandDefault)
			require.ErrorIs(t, err, sqlcon.ErrNoRows())

			actual, err := p.FindIdentityByCredentialIdentifier(ctx, email, true, identity.ExpandDefault)
			require.NoError(t, err)

			expected.Credentials = nil
			assertEqual(t, expected, actual)

			t.Run("not if on another network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, ctx, p)
				_, err := p.FindIdentityByCredentialIdentifier(ctx, email, true, identity.ExpandDefault)
				require.ErrorIs(t, err, sqlcon.ErrNoRows())
			})
		})

		t.Run("case=find identity by phone number credential identifier with backward compatibility", func(t *testing.T) {
			// Use unique phone numbers to avoid conflicts with other tests
			nonNormalizedPhone := "+44 20 7946 0958"
			normalizedPhone := "+442079460958"

			t.Run("upgraded identity with both identifiers via raw SQL", func(t *testing.T) {
				// This test simulates a legacy identity that has been upgraded and persisted,
				// so it has BOTH non-normalized and normalized identifiers in the database.
				// This is the expected state after a user with legacy data logs in and updates settings.

				// Get credential type metadata
				var credTypes []identity.CredentialsTypeTable
				require.NoError(t, p.GetConnection(ctx).All(&credTypes))
				var passwordCredType identity.CredentialsTypeTable
				for _, ct := range credTypes {
					if ct.Name == identity.CredentialsTypePassword {
						passwordCredType = ct
						break
					}
				}

				legacyID := x.NewUUID()
				credID := x.NewUUID()
				identifierID1 := x.NewUUID()
				identifierID2 := x.NewUUID()
				now := time.Now()

				// Cleanup after test
				t.Cleanup(func() {
					_ = p.GetConnection(ctx).RawQuery("DELETE FROM identity_credential_identifiers WHERE identity_credential_id = ?", credID).Exec()
					_ = p.GetConnection(ctx).RawQuery("DELETE FROM identity_credentials WHERE id = ?", credID).Exec()
					_ = p.GetConnection(ctx).RawQuery("DELETE FROM identities WHERE id = ?", legacyID).Exec()
				})

				// Insert identity with raw SQL
				require.NoError(t, p.GetConnection(ctx).RawQuery(
					"INSERT INTO identities (id, nid, schema_id, traits, created_at, updated_at, state) VALUES (?, ?, ?, ?, ?, ?, ?)",
					legacyID, nid, config.DefaultIdentityTraitsSchemaID, sqlxx.JSONRawMessage(`{}`), now, now, identity.StateActive,
				).Exec())

				// Insert credential (version 1 after upgrade)
				require.NoError(t, p.GetConnection(ctx).RawQuery(
					"INSERT INTO identity_credentials (id, identity_id, nid, identity_credential_type_id, created_at, updated_at, config, version) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
					credID, legacyID, nid, passwordCredType.ID, now, now, sqlxx.JSONRawMessage(`{"foo":"bar"}`), 1,
				).Exec())

				// Insert BOTH identifiers (simulating post-upgrade state)
				require.NoError(t, p.GetConnection(ctx).RawQuery(
					"INSERT INTO identity_credential_identifiers (id, identity_id, identity_credential_id, nid, identifier, created_at, updated_at, identity_credential_type_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
					identifierID1, legacyID, credID, nid, nonNormalizedPhone, now, now, passwordCredType.ID,
				).Exec())
				require.NoError(t, p.GetConnection(ctx).RawQuery(
					"INSERT INTO identity_credential_identifiers (id, identity_id, identity_credential_id, nid, identifier, created_at, updated_at, identity_credential_type_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
					identifierID2, legacyID, credID, nid, normalizedPhone, now, now, passwordCredType.ID,
				).Exec())

				// Test 1: Find with non-normalized phone
				actualNonNorm, err := p.FindIdentityByCredentialIdentifier(ctx, nonNormalizedPhone, false, identity.ExpandDefault)
				require.NoError(t, err, "should find upgraded identity with non-normalized phone")
				require.Equal(t, legacyID, actualNonNorm.ID)

				// Test 2: Find with normalized phone
				actualNorm, err := p.FindIdentityByCredentialIdentifier(ctx, normalizedPhone, false, identity.ExpandDefault)
				require.NoError(t, err, "should find upgraded identity with normalized phone")
				require.Equal(t, legacyID, actualNorm.ID)

				// Test 3: Case sensitive should only match exact
				actualCaseSens, err := p.FindIdentityByCredentialIdentifier(ctx, nonNormalizedPhone, true, identity.ExpandDefault)
				require.NoError(t, err)
				require.Equal(t, legacyID, actualCaseSens.ID)

				// Test 4: Case sensitive with normalized should also match exact
				actualCaseSensNorm, err := p.FindIdentityByCredentialIdentifier(ctx, normalizedPhone, true, identity.ExpandDefault)
				require.NoError(t, err)
				require.Equal(t, legacyID, actualCaseSensNorm.ID)
			})

			t.Run("new identity with normalized phone can be found with both formats", func(t *testing.T) {
				// Create a new identity with normalized phone (simulating post-normalization data)
				newIdentity := passwordIdentity("", normalizedPhone)
				newIdentity.Traits = identity.Traits(`{}`)
				newIdentity.Credentials[identity.CredentialsTypePassword] = identity.Credentials{
					Type:        identity.CredentialsTypePassword,
					Identifiers: []string{normalizedPhone}, // Only normalized identifier
					Config:      newIdentity.Credentials[identity.CredentialsTypePassword].Config,
					Version:     1,
				}

				require.NoError(t, p.CreateIdentity(ctx, newIdentity))
				createdIDs = append(createdIDs, newIdentity.ID)

				// Test 1: Find with normalized phone (exact match in DB)
				actualNorm, err := p.FindIdentityByCredentialIdentifier(ctx, normalizedPhone, false, identity.ExpandDefault)
				require.NoError(t, err, "should find new identity with normalized phone")
				require.Equal(t, newIdentity.ID, actualNorm.ID)

				// Test 2: Find with non-normalized phone (gets normalized, then matches)
				actualNonNorm, err := p.FindIdentityByCredentialIdentifier(ctx, nonNormalizedPhone, false, identity.ExpandDefault)
				require.NoError(t, err, "should find new identity with non-normalized phone (gets normalized)")
				require.Equal(t, newIdentity.ID, actualNonNorm.ID)
			})
		})

		t.Run("case=find identity by credential identifier prefers exact match when multiple records exist", func(t *testing.T) {
			// This test verifies that when multiple identities match (one with normalized,
			// another with non-normalized), the function returns the one with exact match.
			// Use unique phone numbers to avoid conflicts with other tests
			nonNormalizedPhone := "+1 555-111-2222"
			normalizedPhone := "+15551112222"

			// Get credential type metadata
			var credTypes []identity.CredentialsTypeTable
			require.NoError(t, p.GetConnection(ctx).All(&credTypes))
			var passwordCredType identity.CredentialsTypeTable
			for _, ct := range credTypes {
				if ct.Name == identity.CredentialsTypePassword {
					passwordCredType = ct
					break
				}
			}

			// Create first identity with normalized phone number
			normalizedID := x.NewUUID()
			normalizedCredID := x.NewUUID()
			normalizedIdentifierID := x.NewUUID()
			now := time.Now()

			// Cleanup after test - will be executed in reverse order
			t.Cleanup(func() {
				_ = p.GetConnection(ctx).RawQuery("DELETE FROM identity_credential_identifiers WHERE identity_credential_id = ?", normalizedCredID).Exec()
				_ = p.GetConnection(ctx).RawQuery("DELETE FROM identity_credentials WHERE id = ?", normalizedCredID).Exec()
				_ = p.GetConnection(ctx).RawQuery("DELETE FROM identities WHERE id = ?", normalizedID).Exec()
			})

			// Insert first identity with normalized phone
			require.NoError(t, p.GetConnection(ctx).RawQuery(
				"INSERT INTO identities (id, nid, schema_id, traits, created_at, updated_at, state) VALUES (?, ?, ?, ?, ?, ?, ?)",
				normalizedID, nid, config.DefaultIdentityTraitsSchemaID, sqlxx.JSONRawMessage(`{}`), now, now, identity.StateActive,
			).Exec())

			require.NoError(t, p.GetConnection(ctx).RawQuery(
				"INSERT INTO identity_credentials (id, identity_id, nid, identity_credential_type_id, created_at, updated_at, config, version) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
				normalizedCredID, normalizedID, nid, passwordCredType.ID, now, now, sqlxx.JSONRawMessage(`{"foo":"bar"}`), 1,
			).Exec())

			require.NoError(t, p.GetConnection(ctx).RawQuery(
				"INSERT INTO identity_credential_identifiers (id, identity_id, identity_credential_id, nid, identifier, created_at, updated_at, identity_credential_type_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
				normalizedIdentifierID, normalizedID, normalizedCredID, nid, normalizedPhone, now, now, passwordCredType.ID,
			).Exec())

			// Create second identity with non-normalized phone number
			nonNormalizedID := x.NewUUID()
			nonNormalizedCredID := x.NewUUID()
			nonNormalizedIdentifierID := x.NewUUID()

			// Update cleanup to include second identity
			t.Cleanup(func() {
				_ = p.GetConnection(ctx).RawQuery("DELETE FROM identity_credential_identifiers WHERE identity_credential_id = ?", nonNormalizedCredID).Exec()
				_ = p.GetConnection(ctx).RawQuery("DELETE FROM identity_credentials WHERE id = ?", nonNormalizedCredID).Exec()
				_ = p.GetConnection(ctx).RawQuery("DELETE FROM identities WHERE id = ?", nonNormalizedID).Exec()
			})

			// Insert second identity with non-normalized phone
			require.NoError(t, p.GetConnection(ctx).RawQuery(
				"INSERT INTO identities (id, nid, schema_id, traits, created_at, updated_at, state) VALUES (?, ?, ?, ?, ?, ?, ?)",
				nonNormalizedID, nid, config.DefaultIdentityTraitsSchemaID, sqlxx.JSONRawMessage(`{}`), now, now, identity.StateActive,
			).Exec())

			require.NoError(t, p.GetConnection(ctx).RawQuery(
				"INSERT INTO identity_credentials (id, identity_id, nid, identity_credential_type_id, created_at, updated_at, config, version) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
				nonNormalizedCredID, nonNormalizedID, nid, passwordCredType.ID, now, now, sqlxx.JSONRawMessage(`{"foo":"bar"}`), 1,
			).Exec())

			require.NoError(t, p.GetConnection(ctx).RawQuery(
				"INSERT INTO identity_credential_identifiers (id, identity_id, identity_credential_id, nid, identifier, created_at, updated_at, identity_credential_type_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
				nonNormalizedIdentifierID, nonNormalizedID, nonNormalizedCredID, nid, nonNormalizedPhone, now, now, passwordCredType.ID,
			).Exec())

			// Test: Query with non-normalized phone should return the identity with exact match
			// Even though both identities match (one has normalized, other has non-normalized),
			// the function should prefer the exact match
			actual, err := p.FindIdentityByCredentialIdentifier(ctx, nonNormalizedPhone, false, identity.ExpandDefault)
			require.NoError(t, err, "should find identity with exact non-normalized phone match")
			require.Equal(t, nonNormalizedID, actual.ID, "should return identity with exact match (non-normalized phone)")
			require.NotEqual(t, normalizedID, actual.ID, "should not return identity with only normalized match")
		})

		t.Run("case=FindByCredentialsIdentifier with phone number backward compatibility", func(t *testing.T) {
			t.Run("upgraded identity with both identifiers via raw SQL", func(t *testing.T) {
				// Use unique phone numbers for this test
				nonNormalizedPhone := "+1 555-987-6543"
				normalizedPhone := "+15559876543"
				// This test simulates a legacy identity that has been upgraded and persisted
				// Get credential type metadata
				var credTypes []identity.CredentialsTypeTable
				require.NoError(t, p.GetConnection(ctx).All(&credTypes))
				var passwordCredType identity.CredentialsTypeTable
				for _, ct := range credTypes {
					if ct.Name == identity.CredentialsTypePassword {
						passwordCredType = ct
						break
					}
				}

				legacyID := x.NewUUID()
				credID := x.NewUUID()
				identifierID1 := x.NewUUID()
				identifierID2 := x.NewUUID()
				now := time.Now()

				// Cleanup after test
				t.Cleanup(func() {
					_ = p.GetConnection(ctx).RawQuery("DELETE FROM identity_credential_identifiers WHERE identity_credential_id = ?", credID).Exec()
					_ = p.GetConnection(ctx).RawQuery("DELETE FROM identity_credentials WHERE id = ?", credID).Exec()
					_ = p.GetConnection(ctx).RawQuery("DELETE FROM identities WHERE id = ?", legacyID).Exec()
				})

				// Insert identity with raw SQL
				require.NoError(t, p.GetConnection(ctx).RawQuery(
					"INSERT INTO identities (id, nid, schema_id, traits, created_at, updated_at, state) VALUES (?, ?, ?, ?, ?, ?, ?)",
					legacyID, nid, config.DefaultIdentityTraitsSchemaID, sqlxx.JSONRawMessage(`{}`), now, now, identity.StateActive,
				).Exec())

				// Insert credential (version 1 after upgrade)
				require.NoError(t, p.GetConnection(ctx).RawQuery(
					"INSERT INTO identity_credentials (id, identity_id, nid, identity_credential_type_id, created_at, updated_at, config, version) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
					credID, legacyID, nid, passwordCredType.ID, now, now, sqlxx.JSONRawMessage(`{"foo":"bar"}`), 1,
				).Exec())

				// Insert BOTH identifiers (simulating post-upgrade state)
				require.NoError(t, p.GetConnection(ctx).RawQuery(
					"INSERT INTO identity_credential_identifiers (id, identity_id, identity_credential_id, nid, identifier, created_at, updated_at, identity_credential_type_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
					identifierID1, legacyID, credID, nid, nonNormalizedPhone, now, now, passwordCredType.ID,
				).Exec())
				require.NoError(t, p.GetConnection(ctx).RawQuery(
					"INSERT INTO identity_credential_identifiers (id, identity_id, identity_credential_id, nid, identifier, created_at, updated_at, identity_credential_type_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
					identifierID2, legacyID, credID, nid, normalizedPhone, now, now, passwordCredType.ID,
				).Exec())

				// Test 1: Find with non-normalized phone
				actualNonNorm, cred, err := p.FindByCredentialsIdentifier(ctx, identity.CredentialsTypePassword, nonNormalizedPhone)
				require.NoError(t, err, "should find upgraded identity with non-normalized phone")
				require.Equal(t, legacyID, actualNonNorm.ID)
				require.NotNil(t, cred)

				// Test 2: Find with normalized phone
				actualNorm, cred, err := p.FindByCredentialsIdentifier(ctx, identity.CredentialsTypePassword, normalizedPhone)
				require.NoError(t, err, "should find upgraded identity with normalized phone")
				require.Equal(t, legacyID, actualNorm.ID)
				require.NotNil(t, cred)
			})

			t.Run("new identity with normalized phone", func(t *testing.T) {
				// Use different phone numbers to avoid conflicts with previous test
				// Using UK format which normalizes reliably
				nonNormalizedPhone := "+44 161 496 0123"
				normalizedPhone := "+441614960123"

				// Create new identity with normalized phone
				newIdentity := passwordIdentity("", normalizedPhone)
				newIdentity.Traits = identity.Traits(`{}`)
				newIdentity.Credentials[identity.CredentialsTypePassword] = identity.Credentials{
					Type:        identity.CredentialsTypePassword,
					Identifiers: []string{normalizedPhone},
					Config:      newIdentity.Credentials[identity.CredentialsTypePassword].Config,
					Version:     1,
				}

				require.NoError(t, p.CreateIdentity(ctx, newIdentity))
				createdIDs = append(createdIDs, newIdentity.ID)

				// Test 1: Find with normalized phone (exact match)
				actualNorm, cred, err := p.FindByCredentialsIdentifier(ctx, identity.CredentialsTypePassword, normalizedPhone)
				require.NoError(t, err)
				require.Equal(t, newIdentity.ID, actualNorm.ID)
				require.NotNil(t, cred)

				// Test 2: Find with non-normalized phone (gets normalized)
				actualNonNorm, cred, err := p.FindByCredentialsIdentifier(ctx, identity.CredentialsTypePassword, nonNormalizedPhone)
				require.NoError(t, err)
				require.Equal(t, newIdentity.ID, actualNonNorm.ID)
				require.NotNil(t, cred)
			})
		})

		t.Run("case=find identity by its credentials respects cases", func(t *testing.T) {
			baseEmail := randx.MustString(16, randx.AlphaLowerNum)
			caseSensitive := baseEmail + "@ory.sh"
			caseInsensitiveWithSpaces := " " + strings.ToUpper(baseEmail) + "@ORY.sh "

			expected := identity.NewIdentity("")
			for _, c := range []identity.CredentialsType{
				identity.CredentialsTypePassword,
				identity.CredentialsTypeOIDC,
				identity.CredentialsTypeTOTP,
				identity.CredentialsTypeLookup,
				identity.CredentialsTypeWebAuthn,
			} {
				expected.SetCredentials(c, identity.Credentials{Type: c, Identifiers: []string{caseSensitive}, Config: sqlxx.JSONRawMessage(`{}`)})
			}
			require.NoError(t, p.CreateIdentity(ctx, expected))
			createdIDs = append(createdIDs, expected.ID)

			t.Run("case sensitive", func(t *testing.T) {
				for _, ct := range []identity.CredentialsType{
					identity.CredentialsTypeOIDC,
					identity.CredentialsTypeTOTP,
					identity.CredentialsTypeLookup,
				} {
					t.Run(ct.String(), func(t *testing.T) {
						_, _, err := p.FindByCredentialsIdentifier(ctx, ct, caseInsensitiveWithSpaces)
						require.Error(t, err)

						actual, creds, err := p.FindByCredentialsIdentifier(ctx, ct, caseSensitive)
						require.NoError(t, err)
						assertx.EqualAsJSONExcept(t, expected.Credentials[ct], creds, []string{"created_at", "updated_at", "id"})
						assertx.EqualAsJSONExcept(t, expected, actual, []string{"created_at", "state_changed_at", "updated_at", "id"})
					})
				}
			})

			t.Run("case insensitive", func(t *testing.T) {
				for _, ct := range []identity.CredentialsType{
					identity.CredentialsTypePassword,
					identity.CredentialsTypeWebAuthn,
				} {
					t.Run(ct.String(), func(t *testing.T) {
						for _, cs := range []string{caseSensitive, caseInsensitiveWithSpaces} {
							actual, creds, err := p.FindByCredentialsIdentifier(ctx, ct, cs)
							require.NoError(t, err)
							ec := expected.Credentials[ct]
							ec.Identifiers = []string{strings.ToLower(caseSensitive)}
							assertx.EqualAsJSONExcept(t, ec, creds, []string{"created_at", "updated_at", "id", "config.user_handle", "config.credentials", "version"})
							assertx.EqualAsJSONExcept(t, expected, actual, []string{"created_at", "state_changed_at", "updated_at", "id"})
						}
					})
				}
			})
		})

		t.Run("case=find identity by its credentials case insensitive", func(t *testing.T) {
			identifier := strings.ToUpper(x.NewUUID().String())
			expected := passwordIdentity("", identifier)
			expected.Traits = identity.Traits(`{}`)

			require.NoError(t, p.CreateIdentity(ctx, expected))
			createdIDs = append(createdIDs, expected.ID)

			actual, creds, err := p.FindByCredentialsIdentifier(ctx, identity.CredentialsTypePassword, identifier)
			require.NoError(t, err)

			assert.EqualValues(t, expected.Credentials[identity.CredentialsTypePassword].ID, creds.ID)
			assert.EqualValues(t, []string{strings.ToLower(identifier)}, creds.Identifiers)
			assert.JSONEq(t, string(expected.Credentials[identity.CredentialsTypePassword].Config), string(creds.Config))

			require.Equal(t, expected.Traits, actual.Traits)
			require.Equal(t, expected.ID, actual.ID)
			require.NotNil(t, actual.Credentials[identity.CredentialsTypePassword])
			assert.EqualValues(t, expected.Credentials[identity.CredentialsTypePassword].ID, actual.Credentials[identity.CredentialsTypePassword].ID)
			assert.EqualValues(t, []string{strings.ToLower(identifier)}, actual.Credentials[identity.CredentialsTypePassword].Identifiers)
			assert.JSONEq(t, string(expected.Credentials[identity.CredentialsTypePassword].Config), string(actual.Credentials[identity.CredentialsTypePassword].Config))

			t.Run("not if on another network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, ctx, p)
				_, _, err := p.FindByCredentialsIdentifier(ctx, identity.CredentialsTypePassword, identifier)
				require.ErrorIs(t, err, sqlcon.ErrNoRows())
			})
		})

		t.Run("suite=verifiable-address", func(t *testing.T) {
			createIdentityWithAddresses := func(t *testing.T, email string) identity.VerifiableAddress {
				var i identity.Identity
				require.NoError(t, faker.FakeData(&i))

				address := identity.NewVerifiableEmailAddress(email, i.ID)
				i.VerifiableAddresses = append(i.VerifiableAddresses, *address)

				require.NoError(t, p.CreateIdentity(ctx, &i))
				return i.VerifiableAddresses[0]
			}

			t.Run("case=not found", func(t *testing.T) {
				_, err := p.FindVerifiableAddressByValue(ctx, identity.AddressTypeEmail, "does-not-exist")
				require.ErrorIs(t, err, sqlcon.ErrNoRows())
			})

			transform := func(k int, value string) string {
				switch k % 5 {
				case 0:
					value = strings.ToLower(value)
				case 1:
					value = strings.ToUpper(value)
				case 2:
					value = " " + value
				case 3:
					value = value + " "
				}
				return value
			}

			t.Run("case=create and find", func(t *testing.T) {
				addresses := make([]identity.VerifiableAddress, 15)
				for k := range addresses {
					value := randx.MustString(16, randx.AlphaLowerNum) + "@ory.sh"
					addresses[k] = createIdentityWithAddresses(t, transform(k, value))
					require.NotEmpty(t, addresses[k].ID)
				}

				compare := func(t *testing.T, expected, actual identity.VerifiableAddress) {
					actual.CreatedAt = actual.CreatedAt.UTC().Truncate(time.Hour * 24)
					actual.UpdatedAt = actual.UpdatedAt.UTC().Truncate(time.Hour * 24)
					expected.CreatedAt = expected.CreatedAt.UTC().Truncate(time.Hour * 24)
					expected.UpdatedAt = expected.UpdatedAt.UTC().Truncate(time.Hour * 24)
					expected.Value = strings.TrimSpace(strings.ToLower(expected.Value))
					assert.EqualValues(t, expected, actual)
				}

				for k, expected := range addresses {
					t.Run("method=FindVerifiableAddressByValue", func(t *testing.T) {
						t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
							actual, err := p.FindVerifiableAddressByValue(ctx, expected.Via, transform(k+1, expected.Value))
							require.NoError(t, err)
							compare(t, expected, *actual)

							t.Run("not if on another network", func(t *testing.T) {
								_, p := testhelpers.NewNetwork(t, ctx, p)
								_, err := p.FindVerifiableAddressByValue(ctx, expected.Via, transform(k+1, expected.Value))
								require.ErrorIs(t, err, sqlcon.ErrNoRows())
							})
						})
					})
				}
			})

			t.Run("case=update", func(t *testing.T) {
				address := createIdentityWithAddresses(t, "verification.TestPersister.Update@ory.sh ")

				address.Value = "new-codE "
				require.NoError(t, p.UpdateVerifiableAddress(ctx, &address))

				t.Run("not if on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					require.ErrorIs(t, p.UpdateVerifiableAddress(ctx, &address), sqlcon.ErrNoRows())
				})

				actual, err := p.FindVerifiableAddressByValue(ctx, address.Via, address.Value)
				require.NoError(t, err)
				assert.Equal(t, "new-code", actual.Value)
			})

			t.Run("case=create and update and find", func(t *testing.T) {
				var i identity.Identity
				require.NoError(t, faker.FakeData(&i))

				address := identity.NewVerifiableEmailAddress("verification.TestPersister.Update-Identity@ory.sh", i.ID)
				i.VerifiableAddresses = append(i.VerifiableAddresses, *address)
				require.NoError(t, p.CreateIdentity(ctx, &i))

				_, err := p.FindVerifiableAddressByValue(ctx, identity.AddressTypeEmail, "verification.TestPersister.Update-Identity@ory.sh")
				require.NoError(t, err)

				t.Run("can not find if on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.FindVerifiableAddressByValue(ctx, identity.AddressTypeEmail, "verification.TestPersister.Update-Identity@ory.sh")
					require.ErrorIs(t, err, sqlcon.ErrNoRows())
				})

				address = identity.NewVerifiableEmailAddress("verification.TestPersister.Update-Identity-next@ory.sh", i.ID)
				i.VerifiableAddresses = []identity.VerifiableAddress{*address}
				require.NoError(t, p.UpdateIdentity(ctx, &i))

				_, err = p.FindVerifiableAddressByValue(ctx, identity.AddressTypeEmail, "verification.TestPersister.Update-Identity@ory.sh")
				require.EqualError(t, err, sqlcon.ErrNoRows().Error())

				t.Run("can not find if on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.FindVerifiableAddressByValue(ctx, identity.AddressTypeEmail, "verification.TestPersister.Update-Identity@ory.sh")
					require.ErrorIs(t, err, sqlcon.ErrNoRows())
				})

				actual, err := p.FindVerifiableAddressByValue(ctx, identity.AddressTypeEmail, "verification.TestPersister.Update-Identity-next@ory.sh")
				require.NoError(t, err)
				assert.Equal(t, identity.AddressTypeEmail, actual.Via)
				assert.Equal(t, "verification.testpersister.update-identity-next@ory.sh", actual.Value)

				t.Run("can not find if on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.FindVerifiableAddressByValue(ctx, identity.AddressTypeEmail, "verification.TestPersister.Update-Identity-next@ory.sh")
					require.ErrorIs(t, err, sqlcon.ErrNoRows())
				})
			})

			t.Run("case=create and update and find case insensitive", func(t *testing.T) {
				var i identity.Identity
				require.NoError(t, faker.FakeData(&i))

				address := identity.NewVerifiableEmailAddress("verification.TestPersister.Update-Identity-case-insensitive@ory.sh", i.ID)
				i.VerifiableAddresses = append(i.VerifiableAddresses, *address)
				require.NoError(t, p.CreateIdentity(ctx, &i))

				_, err := p.FindVerifiableAddressByValue(ctx, identity.AddressTypeEmail, strings.ToUpper("verification.TestPersister.Update-Identity-case-insensitive@ory.sh"))
				require.NoError(t, err)

				t.Run("can not find if on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.FindVerifiableAddressByValue(ctx, identity.AddressTypeEmail, strings.ToUpper("verification.TestPersister.Update-Identity-case-insensitive@ory.sh"))
					require.ErrorIs(t, err, sqlcon.ErrNoRows())
				})

				address = identity.NewVerifiableEmailAddress("verification.TestPersister.Update-Identity-case-insensitive-next@ory.sh", i.ID)
				i.VerifiableAddresses = []identity.VerifiableAddress{*address}
				require.NoError(t, p.UpdateIdentity(ctx, &i))

				_, err = p.FindVerifiableAddressByValue(ctx, identity.AddressTypeEmail, strings.ToUpper("verification.TestPersister.Update-Identity-case-insensitive@ory.sh"))
				require.EqualError(t, err, sqlcon.ErrNoRows().Error())

				t.Run("can not find if on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.FindVerifiableAddressByValue(ctx, identity.AddressTypeEmail, strings.ToUpper("verification.TestPersister.Update-Identity-case-insensitive@ory.sh"))
					require.ErrorIs(t, err, sqlcon.ErrNoRows())
				})

				actual, err := p.FindVerifiableAddressByValue(ctx, identity.AddressTypeEmail, strings.ToUpper("verification.TestPersister.Update-Identity-case-insensitive-next@ory.sh"))
				require.NoError(t, err)
				assert.Equal(t, identity.AddressTypeEmail, actual.Via)
				assert.Equal(t, "verification.testpersister.update-identity-case-insensitive-next@ory.sh", actual.Value)

				t.Run("can not find if on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.FindVerifiableAddressByValue(ctx, identity.AddressTypeEmail, "verification.TestPersister.Update-Identity-case-insensitive-next@ory.sh")
					require.ErrorIs(t, err, sqlcon.ErrNoRows())
				})
			})

			t.Run("case=find verifiable address by phone number with normalization", func(t *testing.T) {
				// Use unique phone numbers to avoid conflicts with other tests
				nonNormalizedPhone := "+1 650-253-0000"
				normalizedPhone := "+16502530000"

				t.Run("create with normalized phone and find with both formats", func(t *testing.T) {
					var i identity.Identity
					require.NoError(t, faker.FakeData(&i))

					// Create identity with normalized phone
					address := identity.NewVerifiableAddress(normalizedPhone, i.ID, identity.AddressTypeSMS)
					i.VerifiableAddresses = append(i.VerifiableAddresses, *address)

					require.NoError(t, p.CreateIdentity(ctx, &i))
					createdIDs = append(createdIDs, i.ID)

					// Test 1: Find with normalized phone (exact match in DB)
					actualNorm, err := p.FindVerifiableAddressByValue(ctx, identity.AddressTypeSMS, normalizedPhone)
					require.NoError(t, err, "should find verifiable address with normalized phone")
					assert.Equal(t, normalizedPhone, actualNorm.Value)
					assert.Equal(t, identity.AddressTypeSMS, actualNorm.Via)

					// Test 2: Find with non-normalized phone (gets normalized, then matches)
					actualNonNorm, err := p.FindVerifiableAddressByValue(ctx, identity.AddressTypeSMS, nonNormalizedPhone)
					require.NoError(t, err, "should find verifiable address with non-normalized phone (gets normalized)")
					assert.Equal(t, normalizedPhone, actualNonNorm.Value)
					assert.Equal(t, actualNorm.ID, actualNonNorm.ID, "should return same address")
				})

				t.Run("create with non-normalized phone and find with both formats", func(t *testing.T) {
					// Use different phone number for this test
					nonNormPhone2 := "+44 20 7946 0958"
					normPhone2 := "+442079460958"

					var i identity.Identity
					require.NoError(t, faker.FakeData(&i))

					// Create identity with non-normalized phone
					// Note: CreateIdentity will normalize the phone during creation
					address := identity.NewVerifiableAddress(nonNormPhone2, i.ID, identity.AddressTypeSMS)
					i.VerifiableAddresses = append(i.VerifiableAddresses, *address)

					require.NoError(t, p.CreateIdentity(ctx, &i))
					createdIDs = append(createdIDs, i.ID)

					// Test 1: Find with non-normalized phone (gets normalized for search)
					actualNonNorm, err := p.FindVerifiableAddressByValue(ctx, identity.AddressTypeSMS, nonNormPhone2)
					require.NoError(t, err, "should find verifiable address with non-normalized phone")
					assert.Equal(t, normPhone2, actualNonNorm.Value, "stored value should be normalized")
					assert.Equal(t, identity.AddressTypeSMS, actualNonNorm.Via)

					// Test 2: Find with normalized phone (direct match)
					actualNorm, err := p.FindVerifiableAddressByValue(ctx, identity.AddressTypeSMS, normPhone2)
					require.NoError(t, err, "should find verifiable address with normalized phone")
					assert.Equal(t, normPhone2, actualNorm.Value)
					assert.Equal(t, actualNonNorm.ID, actualNorm.ID, "should return same address")
				})

				t.Run("backward compatibility with legacy non-normalized phone data", func(t *testing.T) {
					// This test simulates legacy data that was stored before phone normalization was implemented
					// The database contains a non-normalized phone number
					legacyPhone := "+1 650-253-0500"

					identityID := x.NewUUID()
					addressID := x.NewUUID()
					now := time.Now()

					// Cleanup after test
					t.Cleanup(func() {
						_ = p.GetConnection(ctx).RawQuery("DELETE FROM identity_verifiable_addresses WHERE id = ?", addressID).Exec()
						_ = p.GetConnection(ctx).RawQuery("DELETE FROM identities WHERE id = ?", identityID).Exec()
					})

					// Insert identity with raw SQL
					require.NoError(t, p.GetConnection(ctx).RawQuery(
						"INSERT INTO identities (id, nid, schema_id, traits, created_at, updated_at, state) VALUES (?, ?, ?, ?, ?, ?, ?)",
						identityID, nid, config.DefaultIdentityTraitsSchemaID, sqlxx.JSONRawMessage(`{}`), now, now, identity.StateActive,
					).Exec())

					// Insert verifiable address with NON-NORMALIZED phone (legacy data)
					require.NoError(t, p.GetConnection(ctx).RawQuery(
						"INSERT INTO identity_verifiable_addresses (id, identity_id, nid, via, value, verified, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
						addressID, identityID, nid, identity.AddressTypeSMS, legacyPhone, false, identity.VerifiableAddressStatusPending, now, now,
					).Exec())

					// Test: Find with exact non-normalized phone (what user would enter)
					// Should find the legacy non-normalized phone data
					actual, err := p.FindVerifiableAddressByValue(ctx, identity.AddressTypeSMS, legacyPhone)
					require.NoError(t, err, "should find legacy non-normalized phone data with non-normalized search")
					assert.Equal(t, legacyPhone, actual.Value)
					assert.Equal(t, identity.AddressTypeSMS, actual.Via)
					assert.Equal(t, addressID, actual.ID)
				})

				t.Run("post-migration: legacy verifiable address normalized in DB", func(t *testing.T) {
					legacyPhone := "+1 650-253-0600"
					normalizedPhone := "+16502530600"

					identityID := x.NewUUID()
					addressID := x.NewUUID()
					now := time.Now()

					t.Cleanup(func() {
						_ = p.GetConnection(ctx).RawQuery("DELETE FROM identity_verifiable_addresses WHERE id = ?", addressID).Exec()
						_ = p.GetConnection(ctx).RawQuery("DELETE FROM identities WHERE id = ?", identityID).Exec()
					})

					require.NoError(t, p.GetConnection(ctx).RawQuery(
						"INSERT INTO identities (id, nid, schema_id, traits, created_at, updated_at, state) VALUES (?, ?, ?, ?, ?, ?, ?)",
						identityID, nid, config.DefaultIdentityTraitsSchemaID, sqlxx.JSONRawMessage(`{}`), now, now, identity.StateActive,
					).Exec())

					// Insert with non-normalized phone, then simulate migration
					require.NoError(t, p.GetConnection(ctx).RawQuery(
						"INSERT INTO identity_verifiable_addresses (id, identity_id, nid, via, value, verified, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
						addressID, identityID, nid, identity.AddressTypeSMS, legacyPhone, false, identity.VerifiableAddressStatusPending, now, now,
					).Exec())

					// Simulate migration: normalize the value in DB
					require.NoError(t, p.GetConnection(ctx).RawQuery(
						"UPDATE identity_verifiable_addresses SET value = ? WHERE id = ?",
						normalizedPhone, addressID,
					).Exec())

					// Find by E.164
					actual, err := p.FindVerifiableAddressByValue(ctx, identity.AddressTypeSMS, normalizedPhone)
					require.NoError(t, err, "should find migrated verifiable address by E.164")
					assert.Equal(t, normalizedPhone, actual.Value)

					// Find by original non-normalized format (IN query normalizes input)
					actual2, err := p.FindVerifiableAddressByValue(ctx, identity.AddressTypeSMS, legacyPhone)
					require.NoError(t, err, "should find migrated verifiable address by non-normalized phone")
					assert.Equal(t, normalizedPhone, actual2.Value)
					assert.Equal(t, actual.ID, actual2.ID)
				})
			})
		})

		t.Run("suite=credential-types", func(t *testing.T) {
			for _, ct := range identity.AllCredentialTypes {
				t.Run("type="+ct.String(), func(t *testing.T) {
					id, err := idpersistence.FindIdentityCredentialsTypeByName(p.GetConnection(ctx), ct)
					require.NoError(t, err)

					require.NotEqual(t, uuid.Nil, id)
					name, err := idpersistence.FindIdentityCredentialsTypeByID(p.GetConnection(ctx), id)
					require.NoError(t, err)

					assert.Equal(t, ct, name)
				})
			}

			_, err := idpersistence.FindIdentityCredentialsTypeByName(p.GetConnection(ctx), "unknown")
			require.Error(t, err)

			_, err = idpersistence.FindIdentityCredentialsTypeByID(p.GetConnection(ctx), x.NewUUID())
			require.Error(t, err)
		})

		t.Run("suite=recovery-address", func(t *testing.T) {
			sortAddresses := func(addresses []identity.RecoveryAddress) {
				slices.SortFunc(addresses, func(a, b identity.RecoveryAddress) int {
					return strings.Compare(a.Value, b.Value)
				})
			}

			createIdentityWithAddresses := func(t *testing.T, email string) *identity.Identity {
				var i identity.Identity
				require.NoError(t, faker.FakeData(&i))
				i.Traits = []byte(`{"email":"` + email + `"}`)
				address := identity.NewRecoveryEmailAddress(email, i.ID)
				i.RecoveryAddresses = append(i.RecoveryAddresses, *address)

				addressOther := identity.NewRecoveryEmailAddress(email+"_other", i.ID)
				i.RecoveryAddresses = append(i.RecoveryAddresses, *addressOther)

				require.NoError(t, p.CreateIdentity(ctx, &i))
				return &i
			}

			t.Run("case=not found", func(t *testing.T) {
				_, err := p.FindRecoveryAddressByValue(ctx, identity.AddressTypeEmail, "does-not-exist")
				require.ErrorIs(t, err, sqlcon.ErrNoRows())

				allAddresses, err := p.FindAllRecoveryAddressesForIdentityByRecoveryAddressValue(ctx, "does-not-exist")
				require.NoError(t, err)
				require.Len(t, allAddresses, 0)
			})

			t.Run("case=create and find", func(t *testing.T) {
				addresses := make([]identity.RecoveryAddress, 15)
				for k := range addresses {
					addresses[k] = createIdentityWithAddresses(t, randx.MustString(16, randx.AlphaLowerNum)+"@ory.sh").RecoveryAddresses[0]
					require.NotEmpty(t, addresses[k].ID)
				}

				compare := func(t *testing.T, expected, actual identity.RecoveryAddress) {
					actual.CreatedAt = actual.CreatedAt.UTC().Truncate(time.Hour * 24)
					actual.UpdatedAt = actual.UpdatedAt.UTC().Truncate(time.Hour * 24)
					expected.CreatedAt = expected.CreatedAt.UTC().Truncate(time.Hour * 24)
					expected.UpdatedAt = expected.UpdatedAt.UTC().Truncate(time.Hour * 24)
					assert.EqualValues(t, expected, actual)
				}

				for k, expected := range addresses {
					t.Run("method=FindRecoveryAddressByValue", func(t *testing.T) {
						t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
							actual, err := p.FindRecoveryAddressByValue(ctx, expected.Via, expected.Value)
							require.NoError(t, err)
							compare(t, expected, *actual)

							t.Run("not if on another network", func(t *testing.T) {
								_, p := testhelpers.NewNetwork(t, ctx, p)
								_, err := p.FindRecoveryAddressByValue(ctx, expected.Via, expected.Value)
								require.ErrorIs(t, err, sqlcon.ErrNoRows())
							})
						})
					})

					t.Run("method=FindAllRecoveryAddressesForIdentityByRecoveryAddressValue", func(t *testing.T) {
						t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
							allAddresses, err := p.FindAllRecoveryAddressesForIdentityByRecoveryAddressValue(ctx, expected.Value)
							require.NoError(t, err)
							require.Len(t, allAddresses, 2)
							sortAddresses(allAddresses)
							require.Equal(t, expected.Value, allAddresses[0].Value)
							require.Equal(t, expected.Value+"_other", allAddresses[1].Value)
						})

						t.Run("not if on another network", func(t *testing.T) {
							_, p := testhelpers.NewNetwork(t, ctx, p)
							allAddresses, err := p.FindAllRecoveryAddressesForIdentityByRecoveryAddressValue(ctx, expected.Value)
							require.NoError(t, err)
							require.Len(t, allAddresses, 0)
						})
					})
				}
			})

			t.Run("case=create and update and find", func(t *testing.T) {
				email := randx.MustString(16, randx.AlphaLowerNum) + "@ory.sh"
				emailLower := strings.ToLower(email)
				id := createIdentityWithAddresses(t, email)

				_, err := p.FindRecoveryAddressByValue(ctx, identity.AddressTypeEmail, email)
				require.NoError(t, err)

				allAddresses, err := p.FindAllRecoveryAddressesForIdentityByRecoveryAddressValue(ctx, emailLower)
				require.NoError(t, err)
				require.Len(t, allAddresses, 2)
				sortAddresses(allAddresses)
				require.Equal(t, allAddresses[0].Value, emailLower)
				require.Equal(t, allAddresses[1].Value, emailLower+"_other")

				t.Run("can not find if on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.FindRecoveryAddressByValue(ctx, identity.AddressTypeEmail, email)
					require.ErrorIs(t, err, sqlcon.ErrNoRows())

					allAddresses, err := p.FindAllRecoveryAddressesForIdentityByRecoveryAddressValue(ctx, emailLower)
					require.NoError(t, err)
					require.Len(t, allAddresses, 0)
				})

				emailNext := randx.MustString(16, randx.AlphaLowerNum) + "@ory.sh"
				id.RecoveryAddresses = []identity.RecoveryAddress{{Via: identity.AddressTypeEmail, Value: emailNext}, {Via: identity.AddressTypeEmail, Value: emailNext + "_other"}}
				require.NoError(t, p.UpdateIdentity(ctx, id))

				_, err = p.FindRecoveryAddressByValue(ctx, identity.AddressTypeEmail, email)
				require.EqualError(t, err, sqlcon.ErrNoRows().Error())

				allAddresses, err = p.FindAllRecoveryAddressesForIdentityByRecoveryAddressValue(ctx, emailLower)
				require.NoError(t, err)
				require.Len(t, allAddresses, 0)

				t.Run("can not find if on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.FindRecoveryAddressByValue(ctx, identity.AddressTypeEmail, email)
					require.ErrorIs(t, err, sqlcon.ErrNoRows())

					allAddresses, err := p.FindAllRecoveryAddressesForIdentityByRecoveryAddressValue(ctx, emailLower)
					require.NoError(t, err)
					require.Len(t, allAddresses, 0)
				})

				emailNextLower := strings.ToLower(emailNext)
				actual, err := p.FindRecoveryAddressByValue(ctx, identity.AddressTypeEmail, emailNext)
				require.NoError(t, err)
				assert.Equal(t, identity.AddressTypeEmail, actual.Via)
				assert.Equal(t, emailNextLower, actual.Value)

				allAddresses, err = p.FindAllRecoveryAddressesForIdentityByRecoveryAddressValue(ctx, emailNextLower)
				require.NoError(t, err)
				require.Len(t, allAddresses, 2)
				sortAddresses(allAddresses)
				assert.Equal(t, identity.AddressTypeEmail, allAddresses[0].Via)
				assert.Equal(t, emailNextLower, allAddresses[0].Value)
				assert.Equal(t, identity.AddressTypeEmail, allAddresses[1].Via)
				assert.Equal(t, emailNextLower+"_other", allAddresses[1].Value)

				t.Run("can not find if on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.FindRecoveryAddressByValue(ctx, identity.AddressTypeEmail, emailNext)
					require.ErrorIs(t, err, sqlcon.ErrNoRows())

					allAddresses, err := p.FindAllRecoveryAddressesForIdentityByRecoveryAddressValue(ctx, emailNextLower)
					require.NoError(t, err)
					require.Len(t, allAddresses, 0)
				})
			})

			t.Run("case=create and update and find case insensitive", func(t *testing.T) {
				id := createIdentityWithAddresses(t, "recovery.TestPersister.Update-case-insensitive@ory.sh")

				_, err := p.FindRecoveryAddressByValue(ctx, identity.AddressTypeEmail, strings.ToUpper("recovery.TestPersister.Update-case-insensitive@ory.sh"))
				require.NoError(t, err)

				t.Run("can not find if on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.FindRecoveryAddressByValue(ctx, identity.AddressTypeEmail, strings.ToUpper("Recovery.TestPersister.Update-case-insensitive@ory.sh"))
					require.ErrorIs(t, err, sqlcon.ErrNoRows())
				})

				id.RecoveryAddresses = []identity.RecoveryAddress{{Via: identity.AddressTypeEmail, Value: "recovery.TestPersister.Update-case-insensitive-next@ory.sh"}}
				require.NoError(t, p.UpdateIdentity(ctx, id))

				_, err = p.FindRecoveryAddressByValue(ctx, identity.AddressTypeEmail, strings.ToUpper("recovery.TestPersister.Update-case-insensitive@ory.sh"))
				require.EqualError(t, err, sqlcon.ErrNoRows().Error())

				t.Run("can not find if on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.FindRecoveryAddressByValue(ctx, identity.AddressTypeEmail, strings.ToUpper("recovery.TestPersister.Update-case-insensitive@ory.sh"))
					require.ErrorIs(t, err, sqlcon.ErrNoRows())
				})

				actual, err := p.FindRecoveryAddressByValue(ctx, identity.AddressTypeEmail, strings.ToUpper("recovery.TestPersister.Update-case-insensitive-next@ory.sh"))
				require.NoError(t, err)
				assert.Equal(t, identity.AddressTypeEmail, actual.Via)
				assert.Equal(t, "recovery.testpersister.update-case-insensitive-next@ory.sh", actual.Value)

				t.Run("can not find if on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.FindRecoveryAddressByValue(ctx, identity.AddressTypeEmail, strings.ToUpper("recovery.TestPersister.Update-case-insensitive-next@ory.sh"))
					require.ErrorIs(t, err, sqlcon.ErrNoRows())
				})
			})

			t.Run("case=find recovery address by phone number with normalization", func(t *testing.T) {
				// Use unique phone numbers to avoid conflicts with other tests
				nonNormalizedPhone := "+1 650-253-0001"
				normalizedPhone := "+16502530001"

				t.Run("create with normalized phone and find with both formats", func(t *testing.T) {
					var i identity.Identity
					require.NoError(t, faker.FakeData(&i))
					i.Traits = []byte(`{"phone":"` + normalizedPhone + `"}`)

					// Create identity with normalized phone
					address := identity.NewRecoverySMSAddress(normalizedPhone, i.ID)
					i.RecoveryAddresses = append(i.RecoveryAddresses, *address)

					require.NoError(t, p.CreateIdentity(ctx, &i))
					createdIDs = append(createdIDs, i.ID)

					// Test 1: Find with normalized phone (exact match in DB)
					actualNorm, err := p.FindRecoveryAddressByValue(ctx, identity.AddressTypeSMS, normalizedPhone)
					require.NoError(t, err, "should find recovery address with normalized phone")
					assert.Equal(t, normalizedPhone, actualNorm.Value)
					assert.Equal(t, identity.AddressTypeSMS, actualNorm.Via)

					// Test 2: Find with non-normalized phone (gets normalized, then matches)
					actualNonNorm, err := p.FindRecoveryAddressByValue(ctx, identity.AddressTypeSMS, nonNormalizedPhone)
					require.NoError(t, err, "should find recovery address with non-normalized phone (gets normalized)")
					assert.Equal(t, normalizedPhone, actualNonNorm.Value)
					assert.Equal(t, actualNorm.ID, actualNonNorm.ID, "should return same address")
				})

				t.Run("create with non-normalized phone and find with both formats", func(t *testing.T) {
					// Use different phone number for this test
					nonNormPhone2 := "+44 20 7946 0123"
					normPhone2 := "+442079460123"

					var i identity.Identity
					require.NoError(t, faker.FakeData(&i))
					i.Traits = []byte(`{"phone":"` + nonNormPhone2 + `"}`)

					// Create identity with non-normalized phone
					// Note: CreateIdentity will normalize the phone during creation
					address := identity.NewRecoverySMSAddress(nonNormPhone2, i.ID)
					i.RecoveryAddresses = append(i.RecoveryAddresses, *address)

					require.NoError(t, p.CreateIdentity(ctx, &i))
					createdIDs = append(createdIDs, i.ID)

					// Test 1: Find with non-normalized phone (gets normalized for search)
					actualNonNorm, err := p.FindRecoveryAddressByValue(ctx, identity.AddressTypeSMS, nonNormPhone2)
					require.NoError(t, err, "should find recovery address with non-normalized phone")
					assert.Equal(t, normPhone2, actualNonNorm.Value, "stored value should be normalized")
					assert.Equal(t, identity.AddressTypeSMS, actualNonNorm.Via)

					// Test 2: Find with normalized phone (direct match)
					actualNorm, err := p.FindRecoveryAddressByValue(ctx, identity.AddressTypeSMS, normPhone2)
					require.NoError(t, err, "should find recovery address with normalized phone")
					assert.Equal(t, normPhone2, actualNorm.Value)
					assert.Equal(t, actualNonNorm.ID, actualNorm.ID, "should return same address")
				})

				t.Run("backward compatibility with legacy non-normalized phone data", func(t *testing.T) {
					// This test simulates legacy data that was stored before phone normalization was implemented
					// The database contains a non-normalized phone number
					legacyPhone := "+1 650-253-0600"

					identityID := x.NewUUID()
					addressID := x.NewUUID()
					now := time.Now()

					// Cleanup after test
					t.Cleanup(func() {
						_ = p.GetConnection(ctx).RawQuery("DELETE FROM identity_recovery_addresses WHERE id = ?", addressID).Exec()
						_ = p.GetConnection(ctx).RawQuery("DELETE FROM identities WHERE id = ?", identityID).Exec()
					})

					// Insert identity with raw SQL
					require.NoError(t, p.GetConnection(ctx).RawQuery(
						"INSERT INTO identities (id, nid, schema_id, traits, created_at, updated_at, state) VALUES (?, ?, ?, ?, ?, ?, ?)",
						identityID, nid, config.DefaultIdentityTraitsSchemaID, sqlxx.JSONRawMessage(`{}`), now, now, identity.StateActive,
					).Exec())

					// Insert recovery address with NON-NORMALIZED phone (legacy data)
					require.NoError(t, p.GetConnection(ctx).RawQuery(
						"INSERT INTO identity_recovery_addresses (id, identity_id, nid, via, value, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
						addressID, identityID, nid, identity.AddressTypeSMS, legacyPhone, now, now,
					).Exec())

					// Test: Find with exact non-normalized phone (what user would enter)
					// Should find the legacy non-normalized phone data
					actual, err := p.FindRecoveryAddressByValue(ctx, identity.AddressTypeSMS, legacyPhone)
					require.NoError(t, err, "should find legacy non-normalized phone data with non-normalized search")
					assert.Equal(t, legacyPhone, actual.Value)
					assert.Equal(t, identity.AddressTypeSMS, actual.Via)
					assert.Equal(t, addressID, actual.ID)
				})

				t.Run("post-migration: legacy recovery address normalized in DB", func(t *testing.T) {
					legacyPhone := "+1 650-253-0700"
					normalizedPhone := "+16502530700"

					identityID := x.NewUUID()
					addressID := x.NewUUID()
					now := time.Now()

					t.Cleanup(func() {
						_ = p.GetConnection(ctx).RawQuery("DELETE FROM identity_recovery_addresses WHERE id = ?", addressID).Exec()
						_ = p.GetConnection(ctx).RawQuery("DELETE FROM identities WHERE id = ?", identityID).Exec()
					})

					require.NoError(t, p.GetConnection(ctx).RawQuery(
						"INSERT INTO identities (id, nid, schema_id, traits, created_at, updated_at, state) VALUES (?, ?, ?, ?, ?, ?, ?)",
						identityID, nid, config.DefaultIdentityTraitsSchemaID, sqlxx.JSONRawMessage(`{}`), now, now, identity.StateActive,
					).Exec())

					// Insert with non-normalized phone, then simulate migration
					require.NoError(t, p.GetConnection(ctx).RawQuery(
						"INSERT INTO identity_recovery_addresses (id, identity_id, nid, via, value, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
						addressID, identityID, nid, identity.AddressTypeSMS, legacyPhone, now, now,
					).Exec())

					// Simulate migration: normalize the value in DB
					require.NoError(t, p.GetConnection(ctx).RawQuery(
						"UPDATE identity_recovery_addresses SET value = ? WHERE id = ?",
						normalizedPhone, addressID,
					).Exec())

					// Find by E.164
					actual, err := p.FindRecoveryAddressByValue(ctx, identity.AddressTypeSMS, normalizedPhone)
					require.NoError(t, err, "should find migrated recovery address by E.164")
					assert.Equal(t, normalizedPhone, actual.Value)

					// Find by original non-normalized format
					actual2, err := p.FindRecoveryAddressByValue(ctx, identity.AddressTypeSMS, legacyPhone)
					require.NoError(t, err, "should find migrated recovery address by non-normalized phone")
					assert.Equal(t, normalizedPhone, actual2.Value)
					assert.Equal(t, actual.ID, actual2.ID)
				})
			})

			t.Run("case=find all recovery addresses by phone number with normalization", func(t *testing.T) {
				t.Run("create with multiple SMS addresses and find all with normalized phone", func(t *testing.T) {
					// Use unique phone numbers to avoid conflicts with other tests
					nonNormalizedPhone1 := "+1 650-253-0100"
					normalizedPhone1 := "+16502530100"
					normalizedPhone2 := "+16502530200"

					var i identity.Identity
					require.NoError(t, faker.FakeData(&i))
					i.Traits = []byte(`{"phone":"` + normalizedPhone1 + `"}`)

					// Create identity with two SMS recovery addresses
					address1 := identity.NewRecoverySMSAddress(normalizedPhone1, i.ID)
					address2 := identity.NewRecoverySMSAddress(normalizedPhone2, i.ID)
					i.RecoveryAddresses = append(i.RecoveryAddresses, *address1, *address2)

					require.NoError(t, p.CreateIdentity(ctx, &i))
					createdIDs = append(createdIDs, i.ID)

					// Test 1: Find all addresses using normalized phone1
					allAddresses, err := p.FindAllRecoveryAddressesForIdentityByRecoveryAddressValue(ctx, normalizedPhone1)
					require.NoError(t, err, "should find all recovery addresses with normalized phone")
					require.Len(t, allAddresses, 2, "should return all recovery addresses for the identity")
					sortAddresses(allAddresses)
					assert.Equal(t, normalizedPhone1, allAddresses[0].Value)
					assert.Equal(t, normalizedPhone2, allAddresses[1].Value)
					assert.Equal(t, identity.AddressTypeSMS, allAddresses[0].Via)
					assert.Equal(t, identity.AddressTypeSMS, allAddresses[1].Via)

					// Test 2: Find all addresses using non-normalized phone1 (gets normalized, then matches)
					allAddresses2, err := p.FindAllRecoveryAddressesForIdentityByRecoveryAddressValue(ctx, nonNormalizedPhone1)
					require.NoError(t, err, "should find all recovery addresses with non-normalized phone")
					require.Len(t, allAddresses2, 2, "should return all recovery addresses when querying with non-normalized phone")
					sortAddresses(allAddresses2)
					assert.Equal(t, normalizedPhone1, allAddresses2[0].Value)
					assert.Equal(t, normalizedPhone2, allAddresses2[1].Value)

					// Test 3: Find all addresses using normalized phone2
					allAddresses3, err := p.FindAllRecoveryAddressesForIdentityByRecoveryAddressValue(ctx, normalizedPhone2)
					require.NoError(t, err)
					require.Len(t, allAddresses3, 2, "should return all recovery addresses when querying with second phone")
					sortAddresses(allAddresses3)
					assert.Equal(t, normalizedPhone1, allAddresses3[0].Value)
					assert.Equal(t, normalizedPhone2, allAddresses3[1].Value)
				})

				t.Run("create with multiple SMS addresses and find all with non-normalized phone", func(t *testing.T) {
					// Use different phone numbers for this test
					nonNormPhone1 := "+1 650-253-0300"
					normPhone1 := "+16502530300"
					normPhone2 := "+16502530400"

					var i identity.Identity
					require.NoError(t, faker.FakeData(&i))
					i.Traits = []byte(`{"phone":"` + nonNormPhone1 + `"}`)

					// Create identity with two SMS recovery addresses
					// Note: CreateIdentity will normalize both phones during creation
					address1 := identity.NewRecoverySMSAddress(nonNormPhone1, i.ID)
					address2 := identity.NewRecoverySMSAddress(normPhone2, i.ID)
					i.RecoveryAddresses = append(i.RecoveryAddresses, *address1, *address2)

					require.NoError(t, p.CreateIdentity(ctx, &i))
					createdIDs = append(createdIDs, i.ID)

					// Test 1: Find all addresses using non-normalized phone1 (gets normalized for search)
					allAddresses, err := p.FindAllRecoveryAddressesForIdentityByRecoveryAddressValue(ctx, nonNormPhone1)
					require.NoError(t, err, "should find all recovery addresses with non-normalized phone")
					require.Len(t, allAddresses, 2, "should return all recovery addresses for the identity")
					sortAddresses(allAddresses)
					assert.Equal(t, normPhone1, allAddresses[0].Value, "stored values should be normalized")
					assert.Equal(t, normPhone2, allAddresses[1].Value)
					assert.Equal(t, identity.AddressTypeSMS, allAddresses[0].Via)
					assert.Equal(t, identity.AddressTypeSMS, allAddresses[1].Via)

					// Test 2: Find all addresses using normalized phone1 (direct match)
					allAddresses2, err := p.FindAllRecoveryAddressesForIdentityByRecoveryAddressValue(ctx, normPhone1)
					require.NoError(t, err)
					require.Len(t, allAddresses2, 2, "should return all recovery addresses when querying with normalized phone")
					sortAddresses(allAddresses2)
					assert.Equal(t, normPhone1, allAddresses2[0].Value)
					assert.Equal(t, normPhone2, allAddresses2[1].Value)
				})

				t.Run("backward compatibility with legacy non-normalized phone data", func(t *testing.T) {
					// This test simulates legacy data that was stored before phone normalization was implemented
					// The database contains non-normalized phone numbers
					//
					// SECURITY NOTE: This function returns ALL recovery addresses for an identity when ANY ONE matches.
					// The implementation must ensure it only matches addresses for the SAME identity and doesn't
					// accidentally return addresses from different identities that might have similar phone numbers.
					legacyPhone1 := "+1 650-253-0700"
					legacyPhone2 := "+1 650-253-0800"

					identityID := x.NewUUID()
					addressID1 := x.NewUUID()
					addressID2 := x.NewUUID()
					now := time.Now()

					// Cleanup after test
					t.Cleanup(func() {
						_ = p.GetConnection(ctx).RawQuery("DELETE FROM identity_recovery_addresses WHERE id IN (?, ?)", addressID1, addressID2).Exec()
						_ = p.GetConnection(ctx).RawQuery("DELETE FROM identities WHERE id = ?", identityID).Exec()
					})

					// Insert identity with raw SQL
					require.NoError(t, p.GetConnection(ctx).RawQuery(
						"INSERT INTO identities (id, nid, schema_id, traits, created_at, updated_at, state) VALUES (?, ?, ?, ?, ?, ?, ?)",
						identityID, nid, config.DefaultIdentityTraitsSchemaID, sqlxx.JSONRawMessage(`{}`), now, now, identity.StateActive,
					).Exec())

					// Insert TWO recovery addresses with NON-NORMALIZED phones (legacy data)
					require.NoError(t, p.GetConnection(ctx).RawQuery(
						"INSERT INTO identity_recovery_addresses (id, identity_id, nid, via, value, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
						addressID1, identityID, nid, identity.AddressTypeSMS, legacyPhone1, now, now,
					).Exec())
					require.NoError(t, p.GetConnection(ctx).RawQuery(
						"INSERT INTO identity_recovery_addresses (id, identity_id, nid, via, value, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
						addressID2, identityID, nid, identity.AddressTypeSMS, legacyPhone2, now, now,
					).Exec())

					// Test: Find all addresses with exact non-normalized phone (what user would enter)
					// Should find all recovery addresses for the identity
					allAddresses, err := p.FindAllRecoveryAddressesForIdentityByRecoveryAddressValue(ctx, legacyPhone1)
					require.NoError(t, err, "should find all recovery addresses with non-normalized search")
					require.Len(t, allAddresses, 2, "should return all recovery addresses for the identity")
					sortAddresses(allAddresses)
					assert.Equal(t, legacyPhone1, allAddresses[0].Value)
					assert.Equal(t, legacyPhone2, allAddresses[1].Value)
					assert.Equal(t, identity.AddressTypeSMS, allAddresses[0].Via)
					assert.Equal(t, identity.AddressTypeSMS, allAddresses[1].Via)
				})
			})
		})

		t.Run("network reference isolation", func(t *testing.T) {
			nid1, p := testhelpers.NewNetwork(t, ctx, p)
			nid2, _ := testhelpers.NewNetwork(t, ctx, p)

			var m []identity.CredentialsTypeTable
			require.NoError(t, p.GetConnection(ctx).All(&m))

			iid := x.NewUUID()
			require.NoError(t, p.GetConnection(ctx).RawQuery("INSERT INTO identities (id, nid, schema_id, traits, created_at, updated_at) VALUES (?, ?, 'default', '{}', ?, ?)", iid, nid1, time.Now(), time.Now()).Exec())

			cid1, cid2 := x.NewUUID(), x.NewUUID()
			require.NoError(t, p.GetConnection(ctx).RawQuery("INSERT INTO identity_credentials (id, identity_id, nid, identity_credential_type_id, created_at, updated_at, config) VALUES (?, ?, ?, ?, ?, ?, '{}')", cid1, iid, nid1, m[0].ID, time.Now(), time.Now()).Exec())
			require.NoError(t, p.GetConnection(ctx).RawQuery("INSERT INTO identity_credentials (id, identity_id, nid, identity_credential_type_id, created_at, updated_at, config) VALUES (?, ?, ?, ?, ?, ?, '{}')", cid2, iid, nid2, m[0].ID, time.Now(), time.Now()).Exec())

			ici1, ici2 := x.NewUUID(), x.NewUUID()
			require.NoError(t, p.GetConnection(ctx).RawQuery("INSERT INTO identity_credential_identifiers (id, identity_id, identity_credential_id, nid, identifier, created_at, updated_at, identity_credential_type_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", ici1, iid, cid1, nid1, "nid1", time.Now(), time.Now(), m[0].ID).Exec())
			require.NoError(t, p.GetConnection(ctx).RawQuery("INSERT INTO identity_credential_identifiers (id, identity_id, identity_credential_id, nid, identifier, created_at, updated_at, identity_credential_type_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", ici2, iid, cid2, nid2, "nid2", time.Now(), time.Now(), m[0].ID).Exec())

			_, err := p.GetIdentity(ctx, nid1, identity.ExpandNothing)
			require.ErrorIs(t, err, sqlcon.ErrNoRows())

			_, err = p.GetIdentityConfidential(ctx, nid1)
			require.ErrorIs(t, err, sqlcon.ErrNoRows())

			i, c, err := p.FindByCredentialsIdentifier(ctx, m[0].Name, "nid1")
			require.NoError(t, err)
			assert.Equal(t, "nid1", c.Identifiers[0])
			require.Len(t, i.Credentials, 1)

			_, _, err = p.FindByCredentialsIdentifier(ctx, m[0].Name, "nid2")
			require.ErrorIs(t, err, sqlcon.ErrNoRows())

			i, err = p.GetIdentityConfidential(ctx, iid)
			require.NoError(t, err)
			require.Len(t, i.Credentials, 1)
			assert.Equal(t, "nid1", i.Credentials[m[0].Name].Identifiers[0])
		})

		t.Run("suite=update-verifiable-addresses-edge-cases", func(t *testing.T) {
			t.Run("case=add new verifiable addresses", func(t *testing.T) {
				initial := passwordIdentity("", x.NewUUID().String())
				originalEmail := "dev+" + uuid.Must(uuid.NewV4()).String() + "+@ory.com"
				new1Email := "dev+" + uuid.Must(uuid.NewV4()).String() + "+@ory.com"
				new2Email := "dev+" + uuid.Must(uuid.NewV4()).String() + "+@ory.com"
				initial.VerifiableAddresses = []identity.VerifiableAddress{
					{Value: originalEmail, Via: identity.AddressTypeEmail, Verified: false, Status: identity.VerifiableAddressStatusPending},
				}
				require.NoError(t, p.CreateIdentity(ctx, initial))
				createdIDs = append(createdIDs, initial.ID)

				fromDB, err := p.GetIdentity(ctx, initial.ID, identity.ExpandDefault)
				require.NoError(t, err)
				require.Len(t, fromDB.VerifiableAddresses, 1)

				// Add two new addresses
				updated := fromDB.CopyWithoutCredentials()
				updated.VerifiableAddresses = append(updated.VerifiableAddresses,
					identity.VerifiableAddress{Value: new1Email, Via: identity.AddressTypeEmail, Verified: false, Status: identity.VerifiableAddressStatusPending},
					identity.VerifiableAddress{Value: new2Email, Via: identity.AddressTypeEmail, Verified: true, Status: identity.VerifiableAddressStatusCompleted},
				)

				require.NoError(t, p.UpdateIdentity(ctx, updated, identity.DiffAgainst(fromDB)))

				actual, err := p.GetIdentity(ctx, initial.ID, identity.ExpandDefault)
				require.NoError(t, err)
				require.Len(t, actual.VerifiableAddresses, 3)

				values := []string{actual.VerifiableAddresses[0].Value, actual.VerifiableAddresses[1].Value, actual.VerifiableAddresses[2].Value}
				assertContainsValues(t, values, []string{originalEmail, new1Email, new2Email}, nil)

				// Verify the new verified address has verified_at set
				for _, addr := range actual.VerifiableAddresses {
					if addr.Value == new2Email {
						assert.True(t, addr.Verified)
						assert.NotNil(t, addr.VerifiedAt)
					}
				}
			})

			t.Run("case=remove all verifiable addresses", func(t *testing.T) {
				email1 := "dev+" + uuid.Must(uuid.NewV4()).String() + "+@ory.com"
				email2 := "dev+" + uuid.Must(uuid.NewV4()).String() + "+@ory.com"
				initial := passwordIdentity("", x.NewUUID().String())
				initial.VerifiableAddresses = []identity.VerifiableAddress{
					{Value: email1, Via: identity.AddressTypeEmail, Verified: false, Status: identity.VerifiableAddressStatusPending},
					{Value: email2, Via: identity.AddressTypeEmail, Verified: true, Status: identity.VerifiableAddressStatusCompleted},
				}
				require.NoError(t, p.CreateIdentity(ctx, initial))
				createdIDs = append(createdIDs, initial.ID)

				fromDB, err := p.GetIdentity(ctx, initial.ID, identity.ExpandDefault)
				require.NoError(t, err)
				require.Len(t, fromDB.VerifiableAddresses, 2)

				// Remove all addresses
				updated := fromDB.CopyWithoutCredentials()
				updated.VerifiableAddresses = []identity.VerifiableAddress{}

				require.NoError(t, p.UpdateIdentity(ctx, updated, identity.DiffAgainst(fromDB)))

				actual, err := p.GetIdentity(ctx, initial.ID, identity.ExpandDefault)
				require.NoError(t, err)
				assert.Len(t, actual.VerifiableAddresses, 0)
			})

			t.Run("case=remove some and add some verifiable addresses", func(t *testing.T) {
				keepEmail := "dev+" + uuid.Must(uuid.NewV4()).String() + "+@ory.com"
				removeEmail := "dev+" + uuid.Must(uuid.NewV4()).String() + "+@ory.com"
				addEmail := "dev+" + uuid.Must(uuid.NewV4()).String() + "+@ory.com"
				initial := passwordIdentity("", x.NewUUID().String())
				initial.VerifiableAddresses = []identity.VerifiableAddress{
					{Value: keepEmail, Via: identity.AddressTypeEmail, Verified: true, Status: identity.VerifiableAddressStatusCompleted},
					{Value: removeEmail, Via: identity.AddressTypeEmail, Verified: false, Status: identity.VerifiableAddressStatusPending},
				}
				require.NoError(t, p.CreateIdentity(ctx, initial))
				createdIDs = append(createdIDs, initial.ID)

				fromDB, err := p.GetIdentity(ctx, initial.ID, identity.ExpandDefault)
				require.NoError(t, err)
				require.Len(t, fromDB.VerifiableAddresses, 2)

				// Keep one, remove one, add one
				updated := fromDB.CopyWithoutCredentials()
				var keptAddress identity.VerifiableAddress
				for _, addr := range updated.VerifiableAddresses {
					if addr.Value == keepEmail {
						keptAddress = addr
						break
					}
				}
				updated.VerifiableAddresses = []identity.VerifiableAddress{
					keptAddress,
					{Value: addEmail, Via: identity.AddressTypeEmail, Verified: false, Status: identity.VerifiableAddressStatusSent},
				}

				require.NoError(t, p.UpdateIdentity(ctx, updated, identity.DiffAgainst(fromDB)))

				actual, err := p.GetIdentity(ctx, initial.ID, identity.ExpandDefault)
				require.NoError(t, err)
				require.Len(t, actual.VerifiableAddresses, 2)

				values := []string{actual.VerifiableAddresses[0].Value, actual.VerifiableAddresses[1].Value}
				assertContainsValues(t, values, []string{keepEmail, addEmail}, []string{removeEmail})
			})

			t.Run("case=update existing verifiable address properties", func(t *testing.T) {
				changeEmail := "dev+" + uuid.Must(uuid.NewV4()).String() + "+@ory.com"
				initial := passwordIdentity("", x.NewUUID().String())
				initial.VerifiableAddresses = []identity.VerifiableAddress{
					{Value: changeEmail, Via: identity.AddressTypeEmail, Verified: false, Status: identity.VerifiableAddressStatusPending},
				}
				require.NoError(t, p.CreateIdentity(ctx, initial))
				createdIDs = append(createdIDs, initial.ID)

				fromDB, err := p.GetIdentity(ctx, initial.ID, identity.ExpandDefault)
				require.NoError(t, err)
				oldAddr := fromDB.VerifiableAddresses[0]
				assert.False(t, oldAddr.Verified)
				assert.Nil(t, oldAddr.VerifiedAt)

				// Change the address value - this should be treated as removal + addition
				updated := fromDB.CopyWithoutCredentials()
				updated.VerifiableAddresses = []identity.VerifiableAddress{
					{Value: changeEmail, Via: identity.AddressTypeEmail, Verified: true, Status: identity.VerifiableAddressStatusCompleted},
				}

				require.NoError(t, p.UpdateIdentity(ctx, updated, identity.DiffAgainst(fromDB)))

				actual, err := p.GetIdentity(ctx, initial.ID, identity.ExpandDefault)
				require.NoError(t, err)
				require.Len(t, actual.VerifiableAddresses, 1)
				assert.Equal(t, changeEmail, actual.VerifiableAddresses[0].Value)
				assert.True(t, actual.VerifiableAddresses[0].Verified)
				assert.NotNil(t, actual.VerifiableAddresses[0].VerifiedAt)
			})

			t.Run("case=replace all verifiable addresses at once", func(t *testing.T) {
				old1Email := "dev+" + uuid.Must(uuid.NewV4()).String() + "+@ory.com"
				old2Email := "dev+" + uuid.Must(uuid.NewV4()).String() + "+@ory.com"
				old3Email := "dev+" + uuid.Must(uuid.NewV4()).String() + "+@ory.com"
				initial := passwordIdentity("", x.NewUUID().String())
				initial.VerifiableAddresses = []identity.VerifiableAddress{
					{Value: old1Email, Via: identity.AddressTypeEmail, Verified: true, Status: identity.VerifiableAddressStatusCompleted},
					{Value: old2Email, Via: identity.AddressTypeEmail, Verified: false, Status: identity.VerifiableAddressStatusPending},
					{Value: old3Email, Via: identity.AddressTypeEmail, Verified: false, Status: identity.VerifiableAddressStatusSent},
				}
				require.NoError(t, p.CreateIdentity(ctx, initial))
				createdIDs = append(createdIDs, initial.ID)

				fromDB, err := p.GetIdentity(ctx, initial.ID, identity.ExpandDefault)
				require.NoError(t, err)
				require.Len(t, fromDB.VerifiableAddresses, 3)

				new1Email := "dev+" + uuid.Must(uuid.NewV4()).String() + "+@ory.com"
				new2Email := "dev+" + uuid.Must(uuid.NewV4()).String() + "+@ory.com"
				// Replace all addresses with new ones
				updated := fromDB.CopyWithoutCredentials()
				updated.VerifiableAddresses = []identity.VerifiableAddress{
					{Value: new1Email, Via: identity.AddressTypeEmail, Verified: false, Status: identity.VerifiableAddressStatusPending},
					{Value: new2Email, Via: identity.AddressTypeEmail, Verified: false, Status: identity.VerifiableAddressStatusPending},
				}

				require.NoError(t, p.UpdateIdentity(ctx, updated, identity.DiffAgainst(fromDB)))

				actual, err := p.GetIdentity(ctx, initial.ID, identity.ExpandDefault)
				require.NoError(t, err)
				require.Len(t, actual.VerifiableAddresses, 2)

				values := []string{actual.VerifiableAddresses[0].Value, actual.VerifiableAddresses[1].Value}
				assertContainsValues(t, values, []string{new1Email, new2Email}, []string{old1Email, old2Email, old3Email})
			})
		})

		t.Run("suite=update-recovery-addresses-edge-cases", func(t *testing.T) {
			t.Run("case=add new recovery addresses", func(t *testing.T) {
				initialEmail := randx.MustString(16, randx.AlphaLowerNum) + "@ory.sh"
				recovery1Email := randx.MustString(16, randx.AlphaLowerNum) + "@ory.sh"
				recovery2Email := randx.MustString(16, randx.AlphaLowerNum) + "@ory.sh"

				initial := passwordIdentity("", x.NewUUID().String())
				initial.RecoveryAddresses = []identity.RecoveryAddress{
					{Value: initialEmail, Via: identity.AddressTypeEmail},
				}
				require.NoError(t, p.CreateIdentity(ctx, initial))
				createdIDs = append(createdIDs, initial.ID)

				fromDB, err := p.GetIdentity(ctx, initial.ID, identity.ExpandDefault)
				require.NoError(t, err)
				require.Len(t, fromDB.RecoveryAddresses, 1)

				// Add two new addresses
				updated := fromDB.CopyWithoutCredentials()
				updated.RecoveryAddresses = append(updated.RecoveryAddresses,
					identity.RecoveryAddress{Value: recovery1Email, Via: identity.AddressTypeEmail},
					identity.RecoveryAddress{Value: recovery2Email, Via: identity.AddressTypeEmail},
				)

				require.NoError(t, p.UpdateIdentity(ctx, updated, identity.DiffAgainst(fromDB)))

				actual, err := p.GetIdentity(ctx, initial.ID, identity.ExpandDefault)
				require.NoError(t, err)
				require.Len(t, actual.RecoveryAddresses, 3)

				values := []string{actual.RecoveryAddresses[0].Value, actual.RecoveryAddresses[1].Value, actual.RecoveryAddresses[2].Value}
				assertContainsValues(t, values, []string{initialEmail, recovery1Email, recovery2Email}, nil)
			})

			t.Run("case=remove all recovery addresses", func(t *testing.T) {
				remove1Email := randx.MustString(16, randx.AlphaLowerNum) + "@ory.sh"
				remove2Email := randx.MustString(16, randx.AlphaLowerNum) + "@ory.sh"

				initial := passwordIdentity("", x.NewUUID().String())
				initial.RecoveryAddresses = []identity.RecoveryAddress{
					{Value: remove1Email, Via: identity.AddressTypeEmail},
					{Value: remove2Email, Via: identity.AddressTypeEmail},
				}
				require.NoError(t, p.CreateIdentity(ctx, initial))
				createdIDs = append(createdIDs, initial.ID)

				fromDB, err := p.GetIdentity(ctx, initial.ID, identity.ExpandDefault)
				require.NoError(t, err)
				require.Len(t, fromDB.RecoveryAddresses, 2)

				// Remove all addresses
				updated := fromDB.CopyWithoutCredentials()
				updated.RecoveryAddresses = []identity.RecoveryAddress{}

				require.NoError(t, p.UpdateIdentity(ctx, updated, identity.DiffAgainst(fromDB)))

				actual, err := p.GetIdentity(ctx, initial.ID, identity.ExpandDefault)
				require.NoError(t, err)
				assert.Len(t, actual.RecoveryAddresses, 0)
			})

			t.Run("case=remove some and add some recovery addresses", func(t *testing.T) {
				keepEmail := randx.MustString(16, randx.AlphaLowerNum) + "@ory.sh"
				removeEmail := randx.MustString(16, randx.AlphaLowerNum) + "@ory.sh"
				addEmail := randx.MustString(16, randx.AlphaLowerNum) + "@ory.sh"

				initial := passwordIdentity("", x.NewUUID().String())
				initial.RecoveryAddresses = []identity.RecoveryAddress{
					{Value: keepEmail, Via: identity.AddressTypeEmail},
					{Value: removeEmail, Via: identity.AddressTypeEmail},
				}
				require.NoError(t, p.CreateIdentity(ctx, initial))
				createdIDs = append(createdIDs, initial.ID)

				fromDB, err := p.GetIdentity(ctx, initial.ID, identity.ExpandDefault)
				require.NoError(t, err)
				require.Len(t, fromDB.RecoveryAddresses, 2)

				// Keep one, remove one, add one
				updated := fromDB.CopyWithoutCredentials()
				var keptAddress identity.RecoveryAddress
				for _, addr := range updated.RecoveryAddresses {
					if addr.Value == keepEmail {
						keptAddress = addr
						break
					}
				}
				updated.RecoveryAddresses = []identity.RecoveryAddress{
					keptAddress,
					{Value: addEmail, Via: identity.AddressTypeEmail},
				}

				require.NoError(t, p.UpdateIdentity(ctx, updated, identity.DiffAgainst(fromDB)))

				actual, err := p.GetIdentity(ctx, initial.ID, identity.ExpandDefault)
				require.NoError(t, err)
				require.Len(t, actual.RecoveryAddresses, 2)

				values := []string{actual.RecoveryAddresses[0].Value, actual.RecoveryAddresses[1].Value}
				assertContainsValues(t, values, []string{keepEmail, addEmail}, []string{removeEmail})
			})

			t.Run("case=replace all recovery addresses at once", func(t *testing.T) {
				old1Email := randx.MustString(16, randx.AlphaLowerNum) + "@ory.sh"
				old2Email := randx.MustString(16, randx.AlphaLowerNum) + "@ory.sh"
				old3Email := randx.MustString(16, randx.AlphaLowerNum) + "@ory.sh"
				new1Email := randx.MustString(16, randx.AlphaLowerNum) + "@ory.sh"
				new2Email := randx.MustString(16, randx.AlphaLowerNum) + "@ory.sh"

				initial := passwordIdentity("", x.NewUUID().String())
				initial.RecoveryAddresses = []identity.RecoveryAddress{
					{Value: old1Email, Via: identity.AddressTypeEmail},
					{Value: old2Email, Via: identity.AddressTypeEmail},
					{Value: old3Email, Via: identity.AddressTypeEmail},
				}
				require.NoError(t, p.CreateIdentity(ctx, initial))
				createdIDs = append(createdIDs, initial.ID)

				fromDB, err := p.GetIdentity(ctx, initial.ID, identity.ExpandDefault)
				require.NoError(t, err)
				require.Len(t, fromDB.RecoveryAddresses, 3)

				// Replace all addresses with new ones
				updated := fromDB.CopyWithoutCredentials()
				updated.RecoveryAddresses = []identity.RecoveryAddress{
					{Value: new1Email, Via: identity.AddressTypeEmail},
					{Value: new2Email, Via: identity.AddressTypeEmail},
				}

				require.NoError(t, p.UpdateIdentity(ctx, updated, identity.DiffAgainst(fromDB)))

				actual, err := p.GetIdentity(ctx, initial.ID, identity.ExpandDefault)
				require.NoError(t, err)
				require.Len(t, actual.RecoveryAddresses, 2)

				values := []string{actual.RecoveryAddresses[0].Value, actual.RecoveryAddresses[1].Value}
				assertContainsValues(t, values, []string{new1Email, new2Email}, []string{old1Email, old2Email, old3Email})
			})
		})

		t.Run("suite=update-credentials-edge-cases", func(t *testing.T) {
			t.Run("case=add new credential type", func(t *testing.T) {
				totpIdentifier := randx.MustString(16, randx.AlphaLowerNum)

				initial := passwordIdentity("", x.NewUUID().String())
				require.NoError(t, p.CreateIdentity(ctx, initial))
				createdIDs = append(createdIDs, initial.ID)

				fromDB, err := p.GetIdentityConfidential(ctx, initial.ID)
				require.NoError(t, err)
				require.Len(t, fromDB.Credentials, 1)
				_, hasPassword := fromDB.Credentials[identity.CredentialsTypePassword]
				assert.True(t, hasPassword)
				oldPasswordCredID := fromDB.Credentials[identity.CredentialsTypePassword].ID

				// Add TOTP credential
				initial.SetCredentials(identity.CredentialsTypeTOTP, identity.Credentials{
					Type:        identity.CredentialsTypeTOTP,
					Identifiers: []string{totpIdentifier},
					Config:      sqlxx.JSONRawMessage(`{"totp_url":"otpauth://totp/test"}`),
				})

				require.NoError(t, p.UpdateIdentity(ctx, initial, identity.DiffAgainst(fromDB)))

				actual, err := p.GetIdentityConfidential(ctx, initial.ID)
				require.NoError(t, err)
				require.Len(t, actual.Credentials, 2)
				_, hasPassword = actual.Credentials[identity.CredentialsTypePassword]
				_, hasTOTP := actual.Credentials[identity.CredentialsTypeTOTP]
				assert.True(t, hasPassword)
				assert.True(t, hasTOTP)
				assert.Equal(t, []string{totpIdentifier}, actual.Credentials[identity.CredentialsTypeTOTP].Identifiers)
				// Verify that the password credential was not recreated (ID should remain the same)
				assert.Equal(t, oldPasswordCredID, actual.Credentials[identity.CredentialsTypePassword].ID, "password credential should not be recreated when adding TOTP")
			})

			t.Run("case=remove all credentials", func(t *testing.T) {
				oidcIdentifier := randx.MustString(16, randx.AlphaLowerNum)

				initial := passwordIdentity("", x.NewUUID().String())
				initial.SetCredentials(identity.CredentialsTypeOIDC, identity.Credentials{
					Type:        identity.CredentialsTypeOIDC,
					Identifiers: []string{oidcIdentifier},
					Config:      sqlxx.JSONRawMessage(`{}`),
				})
				require.NoError(t, p.CreateIdentity(ctx, initial))
				createdIDs = append(createdIDs, initial.ID)

				fromDB, err := p.GetIdentityConfidential(ctx, initial.ID)
				require.NoError(t, err)
				require.Len(t, fromDB.Credentials, 2)

				// Remove all credentials
				initial.Credentials = map[identity.CredentialsType]identity.Credentials{}

				require.NoError(t, p.UpdateIdentity(ctx, initial, identity.DiffAgainst(fromDB)))

				actual, err := p.GetIdentityConfidential(ctx, initial.ID)
				require.NoError(t, err)
				assert.Len(t, actual.Credentials, 0)
			})

			t.Run("case=remove one credential type and keep others", func(t *testing.T) {
				oidcIdentifier := randx.MustString(16, randx.AlphaLowerNum)
				totpIdentifier := randx.MustString(16, randx.AlphaLowerNum)

				initial := passwordIdentity("", x.NewUUID().String())
				initial.SetCredentials(identity.CredentialsTypeOIDC, identity.Credentials{
					Type:        identity.CredentialsTypeOIDC,
					Identifiers: []string{oidcIdentifier},
					Config:      sqlxx.JSONRawMessage(`{}`),
				})
				initial.SetCredentials(identity.CredentialsTypeTOTP, identity.Credentials{
					Type:        identity.CredentialsTypeTOTP,
					Identifiers: []string{totpIdentifier},
					Config:      sqlxx.JSONRawMessage(`{"totp_url":"otpauth://totp/test"}`),
				})
				require.NoError(t, p.CreateIdentity(ctx, initial))
				createdIDs = append(createdIDs, initial.ID)

				fromDB, err := p.GetIdentityConfidential(ctx, initial.ID)
				require.NoError(t, err)
				require.Len(t, fromDB.Credentials, 3)
				oldOIDCCredID := fromDB.Credentials[identity.CredentialsTypeOIDC].ID
				oldTOTPCredID := fromDB.Credentials[identity.CredentialsTypeTOTP].ID

				// Remove password credential, keep OIDC and TOTP
				delete(initial.Credentials, identity.CredentialsTypePassword)

				require.NoError(t, p.UpdateIdentity(ctx, initial, identity.DiffAgainst(fromDB)))

				actual, err := p.GetIdentityConfidential(ctx, initial.ID)
				require.NoError(t, err)
				require.Len(t, actual.Credentials, 2)
				_, hasPassword := actual.Credentials[identity.CredentialsTypePassword]
				_, hasOIDC := actual.Credentials[identity.CredentialsTypeOIDC]
				_, hasTOTP := actual.Credentials[identity.CredentialsTypeTOTP]
				assert.False(t, hasPassword)
				assert.True(t, hasOIDC)
				assert.True(t, hasTOTP)
				// Verify that OIDC and TOTP credentials were not recreated (IDs should remain the same)
				assert.Equal(t, oldOIDCCredID, actual.Credentials[identity.CredentialsTypeOIDC].ID, "OIDC credential should not be recreated when removing password")
				assert.Equal(t, oldTOTPCredID, actual.Credentials[identity.CredentialsTypeTOTP].ID, "TOTP credential should not be recreated when removing password")
			})

			t.Run("case=update credential config and identifiers", func(t *testing.T) {
				oldEmail := randx.MustString(16, randx.AlphaLowerNum) + "@ory.sh"
				newEmail := randx.MustString(16, randx.AlphaLowerNum) + "@ory.sh"
				initial := passwordIdentity("", oldEmail)
				require.NoError(t, p.CreateIdentity(ctx, initial))
				createdIDs = append(createdIDs, initial.ID)

				fromDB, err := p.GetIdentityConfidential(ctx, initial.ID)
				require.NoError(t, err)
				oldCred := fromDB.Credentials[identity.CredentialsTypePassword]

				// Update password credential with new identifier and config
				initial.SetCredentials(identity.CredentialsTypePassword, identity.Credentials{
					Type:        identity.CredentialsTypePassword,
					Identifiers: []string{newEmail},
					Config:      sqlxx.JSONRawMessage(`{"new":"config"}`),
				})

				require.NoError(t, p.UpdateIdentity(ctx, initial, identity.DiffAgainst(fromDB)))

				actual, err := p.GetIdentityConfidential(ctx, initial.ID)
				require.NoError(t, err)
				newCred := actual.Credentials[identity.CredentialsTypePassword]
				assert.NotEqual(t, oldCred.ID, newCred.ID)
				assert.Equal(t, []string{newEmail}, newCred.Identifiers)
				assert.JSONEq(t, `{"new":"config"}`, string(newCred.Config))
			})

			t.Run("case=replace all credentials at once", func(t *testing.T) {
				initial := passwordIdentity("", x.NewUUID().String())
				initial.SetCredentials(identity.CredentialsTypeOIDC, identity.Credentials{
					Type:        identity.CredentialsTypeOIDC,
					Identifiers: []string{"oidc-replace"},
					Config:      sqlxx.JSONRawMessage(`{}`),
				})
				initial.SetCredentials(identity.CredentialsTypeTOTP, identity.Credentials{
					Type:        identity.CredentialsTypeTOTP,
					Identifiers: []string{"totp-replace"},
					Config:      sqlxx.JSONRawMessage(`{"totp_url":"otpauth://totp/test"}`),
				})
				require.NoError(t, p.CreateIdentity(ctx, initial))
				createdIDs = append(createdIDs, initial.ID)

				fromDB, err := p.GetIdentityConfidential(ctx, initial.ID)
				require.NoError(t, err)
				require.Len(t, fromDB.Credentials, 3)

				// Replace all credentials with webauthn
				initial.Credentials = map[identity.CredentialsType]identity.Credentials{
					identity.CredentialsTypeWebAuthn: {
						Type:        identity.CredentialsTypeWebAuthn,
						Identifiers: []string{"webauthn-new"},
						Config:      sqlxx.JSONRawMessage(`{"credentials":[]}`),
					},
				}

				require.NoError(t, p.UpdateIdentity(ctx, initial, identity.DiffAgainst(fromDB)))

				actual, err := p.GetIdentityConfidential(ctx, initial.ID)
				require.NoError(t, err)
				require.Len(t, actual.Credentials, 1)
				_, hasWebAuthn := actual.Credentials[identity.CredentialsTypeWebAuthn]
				assert.True(t, hasWebAuthn)
				assert.Equal(t, []string{"webauthn-new"}, actual.Credentials[identity.CredentialsTypeWebAuthn].Identifiers)
			})

			t.Run("case=update with no changes", func(t *testing.T) {
				initial := passwordIdentity("", x.NewUUID().String())
				initial.SetCredentials(identity.CredentialsTypeOIDC, identity.Credentials{
					Type:        identity.CredentialsTypeOIDC,
					Identifiers: []string{"oidc-no-change"},
					Config:      sqlxx.JSONRawMessage(`{}`),
				})
				initial.SetCredentials(identity.CredentialsTypeTOTP, identity.Credentials{
					Type:        identity.CredentialsTypeTOTP,
					Identifiers: []string{"totp-no-change"},
					Config:      sqlxx.JSONRawMessage(`{"totp_url":"otpauth://totp/test"}`),
				})
				require.NoError(t, p.CreateIdentity(ctx, initial))
				createdIDs = append(createdIDs, initial.ID)

				fromDB, err := p.GetIdentityConfidential(ctx, initial.ID)
				require.NoError(t, err)
				require.Len(t, fromDB.Credentials, 3)
				oldPasswordCredID := fromDB.Credentials[identity.CredentialsTypePassword].ID
				oldOIDCCredID := fromDB.Credentials[identity.CredentialsTypeOIDC].ID
				oldTOTPCredID := fromDB.Credentials[identity.CredentialsTypeTOTP].ID

				// Update without changing anything
				require.NoError(t, p.UpdateIdentity(ctx, initial, identity.DiffAgainst(fromDB)))

				actual, err := p.GetIdentityConfidential(ctx, initial.ID)
				require.NoError(t, err)
				// Verify no credentials were added or removed
				require.Len(t, actual.Credentials, 3, "credential count should not change when nothing is modified")
				// Verify all credential IDs remained the same (nothing was recreated)
				assert.Equal(t, oldPasswordCredID, actual.Credentials[identity.CredentialsTypePassword].ID, "password credential should not be recreated when nothing changes")
				assert.Equal(t, oldOIDCCredID, actual.Credentials[identity.CredentialsTypeOIDC].ID, "OIDC credential should not be recreated when nothing changes")
				assert.Equal(t, oldTOTPCredID, actual.Credentials[identity.CredentialsTypeTOTP].ID, "TOTP credential should not be recreated when nothing changes")
			})

			t.Run("case=update with json whitespace differences", func(t *testing.T) {
				initial := passwordIdentity("", x.NewUUID().String())
				// Create with compact JSON
				initial.SetCredentials(identity.CredentialsTypeOIDC, identity.Credentials{
					Type:        identity.CredentialsTypeOIDC,
					Identifiers: []string{"oidc-whitespace"},
					Config:      sqlxx.JSONRawMessage(`{"foo":"bar","baz":"qux"}`),
				})
				require.NoError(t, p.CreateIdentity(ctx, initial))
				createdIDs = append(createdIDs, initial.ID)

				fromDB, err := p.GetIdentityConfidential(ctx, initial.ID)
				require.NoError(t, err)
				require.Len(t, fromDB.Credentials, 2)
				oldPasswordCredID := fromDB.Credentials[identity.CredentialsTypePassword].ID
				oldOIDCCredID := fromDB.Credentials[identity.CredentialsTypeOIDC].ID

				// Update with same JSON but different whitespace formatting
				initial.SetCredentials(identity.CredentialsTypeOIDC, identity.Credentials{
					Type:        identity.CredentialsTypeOIDC,
					Identifiers: []string{"oidc-whitespace"},
					// Same JSON content but with different whitespace
					Config: sqlxx.JSONRawMessage(`{
						"foo": "bar",
						"baz": "qux"
					}`),
				})

				require.NoError(t, p.UpdateIdentity(ctx, initial, identity.DiffAgainst(fromDB)))

				actual, err := p.GetIdentityConfidential(ctx, initial.ID)
				require.NoError(t, err)
				// Verify no credentials were added or removed
				require.Len(t, actual.Credentials, 2, "credential count should not change")
				// Verify credential IDs remained the same (nothing was recreated despite JSON formatting difference)
				assert.Equal(t, oldPasswordCredID, actual.Credentials[identity.CredentialsTypePassword].ID, "password credential should not be recreated")
				assert.Equal(t, oldOIDCCredID, actual.Credentials[identity.CredentialsTypeOIDC].ID, "OIDC credential should not be recreated when JSON has different whitespace")
			})

			t.Run("case=update traits with fromDatabase parameter", func(t *testing.T) {
				initial := passwordIdentity("", x.NewUUID().String())
				initial.Traits = identity.Traits(`{"email":"initial@ory.sh","name":"Initial Name"}`)
				require.NoError(t, p.CreateIdentity(ctx, initial))
				createdIDs = append(createdIDs, initial.ID)

				fromDB, err := p.GetIdentity(ctx, initial.ID, identity.ExpandDefault)
				require.NoError(t, err)

				// Update traits using DiffAgainst
				updated := fromDB.CopyWithoutCredentials()
				updated.Traits = identity.Traits(`{"email":"updated@ory.sh","name":"Updated Name"}`)

				require.NoError(t, p.UpdateIdentity(ctx, updated, identity.DiffAgainst(fromDB)))

				actual, err := p.GetIdentity(ctx, initial.ID, identity.ExpandDefault)
				require.NoError(t, err)
				assert.JSONEq(t, `{"email":"updated@ory.sh","name":"Updated Name"}`, string(actual.Traits))
			})

			t.Run("case=update without fromDatabase parameter", func(t *testing.T) {
				initial := passwordIdentity("", x.NewUUID().String())
				initial.SetCredentials(identity.CredentialsTypeOIDC, identity.Credentials{
					Type:        identity.CredentialsTypeOIDC,
					Identifiers: []string{"oidc-no-from-db"},
					Config:      sqlxx.JSONRawMessage(`{}`),
				})
				require.NoError(t, p.CreateIdentity(ctx, initial))
				createdIDs = append(createdIDs, initial.ID)

				fromDB, err := p.GetIdentityConfidential(ctx, initial.ID)
				require.NoError(t, err)
				require.Len(t, fromDB.Credentials, 2)
				oldPasswordCredID := fromDB.Credentials[identity.CredentialsTypePassword].ID
				oldOIDCCredID := fromDB.Credentials[identity.CredentialsTypeOIDC].ID

				// Update without providing fromDatabase - should fetch from DB internally
				updated := *fromDB
				updated.SetCredentials(identity.CredentialsTypeTOTP, identity.Credentials{
					Type:        identity.CredentialsTypeTOTP,
					Identifiers: []string{"totp-no-from-db"},
					Config:      sqlxx.JSONRawMessage(`{"totp_url":"otpauth://totp/test"}`),
				})

				require.NoError(t, p.UpdateIdentity(ctx, &updated))

				actual, err := p.GetIdentityConfidential(ctx, initial.ID)
				require.NoError(t, err)
				require.Len(t, actual.Credentials, 3)
				_, hasTOTP := actual.Credentials[identity.CredentialsTypeTOTP]
				assert.True(t, hasTOTP)
				// Verify that password and OIDC credentials were not recreated (IDs should remain the same)
				assert.Equal(t, oldPasswordCredID, actual.Credentials[identity.CredentialsTypePassword].ID, "password credential should not be recreated when adding TOTP without fromDatabase")
				assert.Equal(t, oldOIDCCredID, actual.Credentials[identity.CredentialsTypeOIDC].ID, "OIDC credential should not be recreated when adding TOTP without fromDatabase")
			})
		})

		t.Run("suite=update-combined-changes", func(t *testing.T) {
			t.Run("case=update addresses and credentials simultaneously", func(t *testing.T) {
				initial := passwordIdentity("", x.NewUUID().String())
				initial.VerifiableAddresses = []identity.VerifiableAddress{
					{Value: "combined-verify@ory.sh", Via: identity.AddressTypeEmail, Verified: false, Status: identity.VerifiableAddressStatusPending},
				}
				initial.RecoveryAddresses = []identity.RecoveryAddress{
					{Value: "combined-recovery@ory.sh", Via: identity.AddressTypeEmail},
				}
				require.NoError(t, p.CreateIdentity(ctx, initial))
				createdIDs = append(createdIDs, initial.ID)

				fromDB, err := p.GetIdentityConfidential(ctx, initial.ID)
				require.NoError(t, err)

				// Change everything at once
				updated := *fromDB
				updated.VerifiableAddresses = []identity.VerifiableAddress{
					{Value: "combined-verify-new@ory.sh", Via: identity.AddressTypeEmail, Verified: true, Status: identity.VerifiableAddressStatusCompleted},
				}
				updated.RecoveryAddresses = []identity.RecoveryAddress{
					{Value: "combined-recovery-new@ory.sh", Via: identity.AddressTypeEmail},
				}
				updated.SetCredentials(identity.CredentialsTypeTOTP, identity.Credentials{
					Type:        identity.CredentialsTypeTOTP,
					Identifiers: []string{"combined-totp"},
					Config:      sqlxx.JSONRawMessage(`{"totp_url":"otpauth://totp/test"}`),
				})

				require.NoError(t, p.UpdateIdentity(ctx, &updated, identity.DiffAgainst(fromDB)))

				actual, err := p.GetIdentityConfidential(ctx, initial.ID)
				require.NoError(t, err)
				require.Len(t, actual.VerifiableAddresses, 1)
				require.Len(t, actual.RecoveryAddresses, 1)
				require.Len(t, actual.Credentials, 2)

				assert.Equal(t, "combined-verify-new@ory.sh", actual.VerifiableAddresses[0].Value)
				assert.Equal(t, "combined-recovery-new@ory.sh", actual.RecoveryAddresses[0].Value)
				_, hasTOTP := actual.Credentials[identity.CredentialsTypeTOTP]
				assert.True(t, hasTOTP)
			})
		})
	}
}

func NewTestIdentity(numAddresses int, prefix string, i int) *identity.Identity {
	var (
		verifiableAddresses []identity.VerifiableAddress
		recoveryAddresses   []identity.RecoveryAddress
	)
	traits := struct {
		Emails   []string `json:"emails"`
		Username string   `json:"username"`
	}{}

	verificationStates := []identity.VerifiableAddressStatus{
		identity.VerifiableAddressStatusPending,
		identity.VerifiableAddressStatusSent,
		identity.VerifiableAddressStatusCompleted,
	}

	for j := 0; j < numAddresses; j++ {
		email := fmt.Sprintf("%s-%d-%d@ory.sh", prefix, i, j)
		traits.Emails = append(traits.Emails, email)
		verifiableAddresses = append(verifiableAddresses, identity.VerifiableAddress{
			Value:    email,
			Via:      identity.AddressTypeEmail,
			Verified: j%2 == 0,
			Status:   verificationStates[j%len(verificationStates)],
		})
		recoveryAddresses = append(recoveryAddresses, identity.RecoveryAddress{
			Value: email,
			Via:   identity.AddressTypeEmail,
		})
	}
	traits.Username = traits.Emails[0]
	rawTraits, _ := json.Marshal(traits)

	id := &identity.Identity{
		SchemaID:            "multiple_emails",
		Traits:              rawTraits,
		VerifiableAddresses: verifiableAddresses,
		RecoveryAddresses:   recoveryAddresses,
		State:               "active",
	}
	id.SetCredentials(identity.CredentialsTypePassword, identity.Credentials{
		Type:        identity.CredentialsTypePassword,
		Identifiers: []string{traits.Username},
		Config:      sqlxx.JSONRawMessage(`{}`),
	})

	return id
}
