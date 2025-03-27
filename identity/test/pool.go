// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	confighelpers "github.com/ory/kratos/driver/config/testhelpers"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/persistence"
	idpersistence "github.com/ory/kratos/persistence/sql/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/x"
	"github.com/ory/x/assertx"
	"github.com/ory/x/crdbx"
	"github.com/ory/x/errorsx"
	"github.com/ory/x/pagination/keysetpagination"
	"github.com/ory/x/randx"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"
)

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
		ctx := confighelpers.WithConfigValues(ctx, map[string]any{
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
			expected := passwordIdentity("", "id-1")
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
				require.ErrorIs(t, err, sqlcon.ErrNoRows)

				count, err := p.CountIdentities(ctx)
				require.NoError(t, err)
				assert.EqualValues(t, int64(0), count)
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
					assert.ErrorIs(t, err, sqlcon.ErrUniqueViolation)
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
			expected := passwordIdentity(altSchema.ID, "id-2")
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
				require.ErrorIs(t, err, sqlcon.ErrNoRows)

				_, err = p.GetIdentityConfidential(ctx, expected.ID)
				require.ErrorIs(t, err, sqlcon.ErrNoRows)
			})
		})

		t.Run("case=fail on duplicate credential identifiers if type is password", func(t *testing.T) {
			initial := passwordIdentity("", "foo@bar.com")
			require.NoError(t, p.CreateIdentity(ctx, initial))
			createdIDs = append(createdIDs, initial.ID)

			for _, ids := range []string{"foo@bar.com", "fOo@bar.com", "FOO@bar.com", "foo@Bar.com"} {
				expected := passwordIdentity("", ids)
				err := p.CreateIdentity(ctx, expected)
				require.ErrorIs(t, err, sqlcon.ErrUniqueViolation, "%+v", err)

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
			initial := oidcIdentity("", "oidc-1")
			require.NoError(t, p.CreateIdentity(ctx, initial))
			createdIDs = append(createdIDs, initial.ID)

			expected := oidcIdentity("", "oidc-1")
			require.Error(t, p.CreateIdentity(ctx, expected))

			_, err := p.GetIdentity(ctx, expected.ID, identity.ExpandNothing)
			require.Error(t, err)

			second := oidcIdentity("", "OIDC-1")
			require.NoError(t, p.CreateIdentity(ctx, second), "should work because oidc is not case-sensitive")
			createdIDs = append(createdIDs, second.ID)

			t.Run("succeeds on different network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, ctx, p)
				expected := oidcIdentity("", "oidc-1")
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
				require.ErrorIs(t, err, sqlcon.ErrNoRows)
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
				require.ErrorIs(t, p.UpdateIdentity(ctx, expected), sqlcon.ErrNoRows)
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
			first := passwordIdentity("", "test-identity@ory.sh")
			first.Traits = identity.Traits(`{}`)
			require.NoError(t, p.CreateIdentity(ctx, first))
			createdIDs = append(createdIDs, first.ID)

			second := passwordIdentity("", "test-identity@ory.sh")
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
				c.Identifiers = []string{"test-identity@ory.sh"}
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
				require.ErrorIs(t, p.DeleteIdentity(ctx, expected.ID), sqlcon.ErrNoRows)

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
			expected := passwordIdentity("", "find-credentials-identifier@ory.sh")
			expected.Traits = identity.Traits(`{}`)

			require.NoError(t, p.CreateIdentity(ctx, expected))
			createdIDs = append(createdIDs, expected.ID)

			actual, creds, err := p.FindByCredentialsIdentifier(ctx, identity.CredentialsTypePassword, "find-credentials-identifier@ory.sh")
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
				_, _, err := p.FindByCredentialsIdentifier(ctx, identity.CredentialsTypePassword, "find-credentials-identifier@ory.sh")
				require.ErrorIs(t, err, sqlcon.ErrNoRows)
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
				require.ErrorIs(t, err, sqlcon.ErrNoRows)
			})
		})

		t.Run("case=find identity only by credentials identifier", func(t *testing.T) {
			expected := passwordIdentity("", "find-credentials-identifier-only@ory.sh")
			expected.Traits = identity.Traits(`{}`)

			require.NoError(t, p.CreateIdentity(ctx, expected))
			createdIDs = append(createdIDs, expected.ID)

			actual, err := p.FindIdentityByCredentialIdentifier(ctx, "find-credentials-IDENTIFIER-only@ory.sh", false)
			require.NoError(t, err)

			expected.Credentials = nil
			assertEqual(t, expected, actual)

			t.Run("not if on another network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, ctx, p)
				_, err := p.FindIdentityByCredentialIdentifier(ctx, "find-credentials-IDENTIFIER-only@ory.sh", false)
				require.ErrorIs(t, err, sqlcon.ErrNoRows)
			})
		})

		t.Run("case=find identity only by credentials identifier case sensitive", func(t *testing.T) {
			expected := passwordIdentity("", "find-credentials-identifier-only-ci@ory.sh")
			expected.Traits = identity.Traits(`{}`)

			require.NoError(t, p.CreateIdentity(ctx, expected))
			createdIDs = append(createdIDs, expected.ID)

			_, err := p.FindIdentityByCredentialIdentifier(ctx, "find-credentials-IDENTIFIER-only-ci@ory.sh", true)
			require.ErrorIs(t, err, sqlcon.ErrNoRows)

			actual, err := p.FindIdentityByCredentialIdentifier(ctx, "find-credentials-identifier-only-ci@ory.sh", true)
			require.NoError(t, err)

			expected.Credentials = nil
			assertEqual(t, expected, actual)

			t.Run("not if on another network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, ctx, p)
				_, err := p.FindIdentityByCredentialIdentifier(ctx, "find-credentials-identifier-only-ci@ory.sh", true)
				require.ErrorIs(t, err, sqlcon.ErrNoRows)
			})
		})

		t.Run("case=find identity by its credentials respects cases", func(t *testing.T) {
			caseSensitive := "6Q(%ZKd~8u_(5uea@ory.sh"
			caseInsensitiveWithSpaces := " 6Q(%ZKD~8U_(5uea@ORY.sh "

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
			identifier := x.NewUUID().String()
			expected := passwordIdentity("", strings.ToUpper(identifier))
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
				require.ErrorIs(t, err, sqlcon.ErrNoRows)
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
				_, err := p.FindVerifiableAddressByValue(ctx, identity.VerifiableAddressTypeEmail, "does-not-exist")
				require.Equal(t, sqlcon.ErrNoRows, errorsx.Cause(err))
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
								require.ErrorIs(t, err, sqlcon.ErrNoRows)
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
					require.ErrorIs(t, p.UpdateVerifiableAddress(ctx, &address), sqlcon.ErrNoRows)
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

				_, err := p.FindVerifiableAddressByValue(ctx, identity.VerifiableAddressTypeEmail, "verification.TestPersister.Update-Identity@ory.sh")
				require.NoError(t, err)

				t.Run("can not find if on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.FindVerifiableAddressByValue(ctx, identity.VerifiableAddressTypeEmail, "verification.TestPersister.Update-Identity@ory.sh")
					require.ErrorIs(t, err, sqlcon.ErrNoRows)
				})

				address = identity.NewVerifiableEmailAddress("verification.TestPersister.Update-Identity-next@ory.sh", i.ID)
				i.VerifiableAddresses = []identity.VerifiableAddress{*address}
				require.NoError(t, p.UpdateIdentity(ctx, &i))

				_, err = p.FindVerifiableAddressByValue(ctx, identity.VerifiableAddressTypeEmail, "verification.TestPersister.Update-Identity@ory.sh")
				require.EqualError(t, err, sqlcon.ErrNoRows.Error())

				t.Run("can not find if on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.FindVerifiableAddressByValue(ctx, identity.VerifiableAddressTypeEmail, "verification.TestPersister.Update-Identity@ory.sh")
					require.ErrorIs(t, err, sqlcon.ErrNoRows)
				})

				actual, err := p.FindVerifiableAddressByValue(ctx, identity.VerifiableAddressTypeEmail, "verification.TestPersister.Update-Identity-next@ory.sh")
				require.NoError(t, err)
				assert.Equal(t, identity.VerifiableAddressTypeEmail, actual.Via)
				assert.Equal(t, "verification.testpersister.update-identity-next@ory.sh", actual.Value)

				t.Run("can not find if on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.FindVerifiableAddressByValue(ctx, identity.VerifiableAddressTypeEmail, "verification.TestPersister.Update-Identity-next@ory.sh")
					require.ErrorIs(t, err, sqlcon.ErrNoRows)
				})
			})

			t.Run("case=create and update and find case insensitive", func(t *testing.T) {
				var i identity.Identity
				require.NoError(t, faker.FakeData(&i))

				address := identity.NewVerifiableEmailAddress("verification.TestPersister.Update-Identity-case-insensitive@ory.sh", i.ID)
				i.VerifiableAddresses = append(i.VerifiableAddresses, *address)
				require.NoError(t, p.CreateIdentity(ctx, &i))

				_, err := p.FindVerifiableAddressByValue(ctx, identity.VerifiableAddressTypeEmail, strings.ToUpper("verification.TestPersister.Update-Identity-case-insensitive@ory.sh"))
				require.NoError(t, err)

				t.Run("can not find if on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.FindVerifiableAddressByValue(ctx, identity.VerifiableAddressTypeEmail, strings.ToUpper("verification.TestPersister.Update-Identity-case-insensitive@ory.sh"))
					require.ErrorIs(t, err, sqlcon.ErrNoRows)
				})

				address = identity.NewVerifiableEmailAddress("verification.TestPersister.Update-Identity-case-insensitive-next@ory.sh", i.ID)
				i.VerifiableAddresses = []identity.VerifiableAddress{*address}
				require.NoError(t, p.UpdateIdentity(ctx, &i))

				_, err = p.FindVerifiableAddressByValue(ctx, identity.VerifiableAddressTypeEmail, strings.ToUpper("verification.TestPersister.Update-Identity-case-insensitive@ory.sh"))
				require.EqualError(t, err, sqlcon.ErrNoRows.Error())

				t.Run("can not find if on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.FindVerifiableAddressByValue(ctx, identity.VerifiableAddressTypeEmail, strings.ToUpper("verification.TestPersister.Update-Identity-case-insensitive@ory.sh"))
					require.ErrorIs(t, err, sqlcon.ErrNoRows)
				})

				actual, err := p.FindVerifiableAddressByValue(ctx, identity.VerifiableAddressTypeEmail, strings.ToUpper("verification.TestPersister.Update-Identity-case-insensitive-next@ory.sh"))
				require.NoError(t, err)
				assert.Equal(t, identity.VerifiableAddressTypeEmail, actual.Via)
				assert.Equal(t, "verification.testpersister.update-identity-case-insensitive-next@ory.sh", actual.Value)

				t.Run("can not find if on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.FindVerifiableAddressByValue(ctx, identity.VerifiableAddressTypeEmail, "verification.TestPersister.Update-Identity-case-insensitive-next@ory.sh")
					require.ErrorIs(t, err, sqlcon.ErrNoRows)
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
			createIdentityWithAddresses := func(t *testing.T, email string) *identity.Identity {
				var i identity.Identity
				require.NoError(t, faker.FakeData(&i))
				i.Traits = []byte(`{"email":"` + email + `"}`)
				address := identity.NewRecoveryEmailAddress(email, i.ID)
				i.RecoveryAddresses = append(i.RecoveryAddresses, *address)
				require.NoError(t, p.CreateIdentity(ctx, &i))
				return &i
			}

			t.Run("case=not found", func(t *testing.T) {
				_, err := p.FindRecoveryAddressByValue(ctx, identity.RecoveryAddressTypeEmail, "does-not-exist")
				require.Equal(t, sqlcon.ErrNoRows, errorsx.Cause(err))
			})

			t.Run("case=create and find", func(t *testing.T) {
				addresses := make([]identity.RecoveryAddress, 15)
				for k := range addresses {
					addresses[k] = createIdentityWithAddresses(t, "recovery.TestPersister.Create"+strconv.Itoa(k)+"@ory.sh").RecoveryAddresses[0]
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
					t.Run("method=FindVerifiableAddressByValue", func(t *testing.T) {
						t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
							actual, err := p.FindRecoveryAddressByValue(ctx, expected.Via, expected.Value)
							require.NoError(t, err)
							compare(t, expected, *actual)

							t.Run("not if on another network", func(t *testing.T) {
								_, p := testhelpers.NewNetwork(t, ctx, p)
								_, err := p.FindRecoveryAddressByValue(ctx, expected.Via, expected.Value)
								require.ErrorIs(t, err, sqlcon.ErrNoRows)
							})
						})
					})
				}
			})

			t.Run("case=create and update and find", func(t *testing.T) {
				id := createIdentityWithAddresses(t, "recovery.TestPersister.Update@ory.sh")

				_, err := p.FindRecoveryAddressByValue(ctx, identity.RecoveryAddressTypeEmail, "recovery.TestPersister.Update@ory.sh")
				require.NoError(t, err)

				t.Run("can not find if on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.FindRecoveryAddressByValue(ctx, identity.RecoveryAddressTypeEmail, "Recovery.TestPersister.Update@ory.sh")
					require.ErrorIs(t, err, sqlcon.ErrNoRows)
				})

				id.RecoveryAddresses = []identity.RecoveryAddress{{Via: identity.RecoveryAddressTypeEmail, Value: "recovery.TestPersister.Update-next@ory.sh"}}
				require.NoError(t, p.UpdateIdentity(ctx, id))

				_, err = p.FindRecoveryAddressByValue(ctx, identity.RecoveryAddressTypeEmail, "recovery.TestPersister.Update@ory.sh")
				require.EqualError(t, err, sqlcon.ErrNoRows.Error())

				t.Run("can not find if on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.FindRecoveryAddressByValue(ctx, identity.RecoveryAddressTypeEmail, "recovery.TestPersister.Update@ory.sh")
					require.ErrorIs(t, err, sqlcon.ErrNoRows)
				})

				actual, err := p.FindRecoveryAddressByValue(ctx, identity.RecoveryAddressTypeEmail, "recovery.TestPersister.Update-next@ory.sh")
				require.NoError(t, err)
				assert.Equal(t, identity.RecoveryAddressTypeEmail, actual.Via)
				assert.Equal(t, "recovery.testpersister.update-next@ory.sh", actual.Value)

				t.Run("can not find if on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.FindRecoveryAddressByValue(ctx, identity.RecoveryAddressTypeEmail, "recovery.TestPersister.Update-next@ory.sh")
					require.ErrorIs(t, err, sqlcon.ErrNoRows)
				})
			})

			t.Run("case=create and update and find case insensitive", func(t *testing.T) {
				id := createIdentityWithAddresses(t, "recovery.TestPersister.Update-case-insensitive@ory.sh")

				_, err := p.FindRecoveryAddressByValue(ctx, identity.RecoveryAddressTypeEmail, strings.ToUpper("recovery.TestPersister.Update-case-insensitive@ory.sh"))
				require.NoError(t, err)

				t.Run("can not find if on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.FindRecoveryAddressByValue(ctx, identity.RecoveryAddressTypeEmail, strings.ToUpper("Recovery.TestPersister.Update-case-insensitive@ory.sh"))
					require.ErrorIs(t, err, sqlcon.ErrNoRows)
				})

				id.RecoveryAddresses = []identity.RecoveryAddress{{Via: identity.RecoveryAddressTypeEmail, Value: "recovery.TestPersister.Update-case-insensitive-next@ory.sh"}}
				require.NoError(t, p.UpdateIdentity(ctx, id))

				_, err = p.FindRecoveryAddressByValue(ctx, identity.RecoveryAddressTypeEmail, strings.ToUpper("recovery.TestPersister.Update-case-insensitive@ory.sh"))
				require.EqualError(t, err, sqlcon.ErrNoRows.Error())

				t.Run("can not find if on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.FindRecoveryAddressByValue(ctx, identity.RecoveryAddressTypeEmail, strings.ToUpper("recovery.TestPersister.Update-case-insensitive@ory.sh"))
					require.ErrorIs(t, err, sqlcon.ErrNoRows)
				})

				actual, err := p.FindRecoveryAddressByValue(ctx, identity.RecoveryAddressTypeEmail, strings.ToUpper("recovery.TestPersister.Update-case-insensitive-next@ory.sh"))
				require.NoError(t, err)
				assert.Equal(t, identity.RecoveryAddressTypeEmail, actual.Via)
				assert.Equal(t, "recovery.testpersister.update-case-insensitive-next@ory.sh", actual.Value)

				t.Run("can not find if on another network", func(t *testing.T) {
					_, p := testhelpers.NewNetwork(t, ctx, p)
					_, err := p.FindRecoveryAddressByValue(ctx, identity.RecoveryAddressTypeEmail, strings.ToUpper("recovery.TestPersister.Update-case-insensitive-next@ory.sh"))
					require.ErrorIs(t, err, sqlcon.ErrNoRows)
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
			require.NoError(t, p.GetConnection(ctx).RawQuery("INSERT INTO identity_credential_identifiers (id, identity_credential_id, nid, identifier, created_at, updated_at, identity_credential_type_id) VALUES (?, ?, ?, ?, ?, ?, ?)", ici1, cid1, nid1, "nid1", time.Now(), time.Now(), m[0].ID).Exec())
			require.NoError(t, p.GetConnection(ctx).RawQuery("INSERT INTO identity_credential_identifiers (id, identity_credential_id, nid, identifier, created_at, updated_at, identity_credential_type_id) VALUES (?, ?, ?, ?, ?, ?, ?)", ici2, cid2, nid2, "nid2", time.Now(), time.Now(), m[0].ID).Exec())

			_, err := p.GetIdentity(ctx, nid1, identity.ExpandNothing)
			require.ErrorIs(t, err, sqlcon.ErrNoRows)

			_, err = p.GetIdentityConfidential(ctx, nid1)
			require.ErrorIs(t, err, sqlcon.ErrNoRows)

			i, c, err := p.FindByCredentialsIdentifier(ctx, m[0].Name, "nid1")
			require.NoError(t, err)
			assert.Equal(t, "nid1", c.Identifiers[0])
			require.Len(t, i.Credentials, 1)

			_, _, err = p.FindByCredentialsIdentifier(ctx, m[0].Name, "nid2")
			require.ErrorIs(t, err, sqlcon.ErrNoRows)

			i, err = p.GetIdentityConfidential(ctx, iid)
			require.NoError(t, err)
			require.Len(t, i.Credentials, 1)
			assert.Equal(t, "nid1", i.Credentials[m[0].Name].Identifiers[0])
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
			Via:      identity.VerifiableAddressTypeEmail,
			Verified: j%2 == 0,
			Status:   verificationStates[j%len(verificationStates)],
		})
		recoveryAddresses = append(recoveryAddresses, identity.RecoveryAddress{
			Value: email,
			Via:   identity.RecoveryAddressTypeEmail,
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
