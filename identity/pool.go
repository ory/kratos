package identity

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"

	"github.com/ory/x/sqlxx"

	"github.com/ory/x/errorsx"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/schema"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/x"
)

type (
	Pool interface {
		// ListIdentities lists all identities in the store given the page and itemsPerPage.
		ListIdentities(ctx context.Context, page, itemsPerPage int) ([]Identity, error)

		// CountIdentities counts the number of identities in the store.
		CountIdentities(ctx context.Context) (int64, error)

		// GetIdentity returns an identity by its id. Will return an error if the identity does not exist or backend
		// connectivity is broken.
		GetIdentity(context.Context, uuid.UUID) (*Identity, error)

		// FindVerifiableAddressByValue returns a matching address or sql.ErrNoRows if no address could be found.
		FindVerifiableAddressByValue(ctx context.Context, via VerifiableAddressType, address string) (*VerifiableAddress, error)

		// FindRecoveryAddressByValue returns a matching address or sql.ErrNoRows if no address could be found.
		FindRecoveryAddressByValue(ctx context.Context, via RecoveryAddressType, address string) (*RecoveryAddress, error)
	}

	PoolProvider interface {
		IdentityPool() Pool
	}

	PrivilegedPoolProvider interface {
		PrivilegedIdentityPool() PrivilegedPool
	}

	PrivilegedPool interface {
		Pool

		// FindByCredentialsIdentifier returns an identity by querying for it's credential identifiers.
		FindByCredentialsIdentifier(ctx context.Context, ct CredentialsType, match string) (*Identity, *Credentials, error)

		// Delete removes an identity by its id. Will return an error
		// if identity exists, backend connectivity is broken, or trait validation fails.
		DeleteIdentity(context.Context, uuid.UUID) error

		// UpdateVerifiableAddress
		UpdateVerifiableAddress(ctx context.Context, address *VerifiableAddress) error

		// Create creates an identity. It is capable of setting credentials without encoding. Will return an error
		// if identity exists, backend connectivity is broken, or trait validation fails.
		CreateIdentity(context.Context, *Identity) error

		// UpdateIdentity updates an identity including its confidential / privileged / protected data.
		UpdateIdentity(context.Context, *Identity) error

		// GetIdentityConfidential returns the identity including it's raw credentials. This should only be used internally.
		GetIdentityConfidential(context.Context, uuid.UUID) (*Identity, error)

		// ListVerifiableAddresses lists all tracked verifiable addresses, regardless of whether they are already verified
		// or not.
		ListVerifiableAddresses(ctx context.Context, page, itemsPerPage int) ([]VerifiableAddress, error)

		// ListRecoveryAddresses lists all tracked recovery addresses.
		ListRecoveryAddresses(ctx context.Context, page, itemsPerPage int) ([]RecoveryAddress, error)
	}
)

func TestPool(conf *config.Config, p interface {
	PrivilegedPool
}) func(t *testing.T) {
	return func(t *testing.T) {
		exampleServerURL := urlx.ParseOrPanic("http://example.com")
		conf.MustSet(config.ViperKeyPublicBaseURL, exampleServerURL.String())
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
		conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, defaultSchema.RawURL)
		conf.MustSet(config.ViperKeyIdentitySchemas, []config.SchemaConfig{{
			ID:  altSchema.ID,
			URL: altSchema.RawURL,
		}})

		var createdIDs []uuid.UUID

		var passwordIdentity = func(schemaID string, credentialsID string) *Identity {
			i := NewIdentity(schemaID)
			i.SetCredentials(CredentialsTypePassword, Credentials{
				Type: CredentialsTypePassword, Identifiers: []string{credentialsID},
				Config: sqlxx.JSONRawMessage(`{"foo":"bar"}`),
			})
			return i
		}

		var oidcIdentity = func(schemaID string, credentialsID string) *Identity {
			i := NewIdentity(schemaID)
			i.SetCredentials(CredentialsTypeOIDC, Credentials{
				Type: CredentialsTypeOIDC, Identifiers: []string{credentialsID},
				Config: sqlxx.JSONRawMessage(`{}`),
			})
			return i
		}

		var assertEqual = func(t *testing.T, expected, actual *Identity) {
			assert.Empty(t, actual.Credentials)
			require.Equal(t, expected.Traits, actual.Traits)
			require.Equal(t, expected.ID, actual.ID)
		}

		t.Run("case=should create and set missing ID", func(t *testing.T) {
			i := NewIdentity(config.DefaultIdentityTraitsSchemaID)
			i.SetCredentials(CredentialsTypeOIDC, Credentials{
				Type: CredentialsTypeOIDC, Identifiers: []string{x.NewUUID().String()},
				Config: sqlxx.JSONRawMessage(`{}`),
			})
			i.ID = uuid.Nil
			require.NoError(t, p.CreateIdentity(context.Background(), i))
			assert.NotEqual(t, uuid.Nil, i.ID)
			createdIDs = append(createdIDs, i.ID)

			count, err := p.CountIdentities(context.Background())
			require.NoError(t, err)
			assert.EqualValues(t, 1, count)
		})

		t.Run("case=create with default values", func(t *testing.T) {
			expected := passwordIdentity("", "id-1")
			require.NoError(t, p.CreateIdentity(context.Background(), expected))
			createdIDs = append(createdIDs, expected.ID)

			actual, err := p.GetIdentity(context.Background(), expected.ID)
			require.NoError(t, err)

			assert.Equal(t, expected.ID, actual.ID)
			assert.Equal(t, config.DefaultIdentityTraitsSchemaID, actual.SchemaID)
			assert.Equal(t, defaultSchema.SchemaURL(exampleServerURL).String(), actual.SchemaURL)
			assertEqual(t, expected, actual)

			count, err := p.CountIdentities(context.Background())
			require.NoError(t, err)
			assert.EqualValues(t, 2, count)
		})

		t.Run("case=should error when the identity ID does not exist", func(t *testing.T) {
			_, err := p.GetIdentity(context.Background(), uuid.UUID{})
			require.Error(t, err)

			_, err = p.GetIdentity(context.Background(), x.NewUUID())
			require.Error(t, err)

			_, err = p.GetIdentityConfidential(context.Background(), x.NewUUID())
			require.Error(t, err)
		})

		t.Run("case=create and keep set values", func(t *testing.T) {
			expected := passwordIdentity(altSchema.ID, "id-2")
			require.NoError(t, p.CreateIdentity(context.Background(), expected))
			createdIDs = append(createdIDs, expected.ID)

			actual, err := p.GetIdentity(context.Background(), expected.ID)
			require.NoError(t, err)
			assert.Equal(t, altSchema.ID, actual.SchemaID)
			assert.Equal(t, altSchema.SchemaURL(exampleServerURL).String(), actual.SchemaURL)
			assertEqual(t, expected, actual)

			actual, err = p.GetIdentityConfidential(context.Background(), expected.ID)
			require.NoError(t, err)
			require.Equal(t, expected.Traits, actual.Traits)
			require.Equal(t, expected.ID, actual.ID)

			assert.Empty(t, actual.CredentialsCollection)
			assert.NotEmpty(t, actual.Credentials)
			assert.NotEmpty(t, expected.Credentials)

			for m, expected := range expected.Credentials {
				assert.Equal(t, expected.ID, actual.Credentials[m].ID)
				assert.JSONEq(t, string(expected.Config), string(actual.Credentials[m].Config))
				assert.Equal(t, expected.Identifiers, actual.Credentials[m].Identifiers)
				assert.Equal(t, expected.Type, actual.Credentials[m].Type)
			}
		})

		t.Run("case=fail on duplicate credential identifiers if type is password", func(t *testing.T) {
			initial := passwordIdentity("", "foo@bar.com")
			require.NoError(t, p.CreateIdentity(context.Background(), initial))
			createdIDs = append(createdIDs, initial.ID)

			for _, ids := range []string{"foo@bar.com", "fOo@bar.com", "FOO@bar.com", "foo@Bar.com"} {
				expected := passwordIdentity("", ids)
				err := p.CreateIdentity(context.Background(), expected)
				require.Error(t, err)
				require.True(t, errors.Is(err, sqlcon.ErrUniqueViolation), "%+v", err)

				_, err = p.GetIdentity(context.Background(), expected.ID)
				require.Error(t, err)
			}
		})

		t.Run("case=fail on duplicate credential identifiers if type is oidc", func(t *testing.T) {
			initial := oidcIdentity("", "oidc-1")
			require.NoError(t, p.CreateIdentity(context.Background(), initial))
			createdIDs = append(createdIDs, initial.ID)

			expected := oidcIdentity("", "oidc-1")
			require.Error(t, p.CreateIdentity(context.Background(), expected))

			_, err := p.GetIdentity(context.Background(), expected.ID)
			require.Error(t, err)

			second := oidcIdentity("", "OIDC-1")
			require.NoError(t, p.CreateIdentity(context.Background(), second), "should work because oidc is not case-sensitive")
			createdIDs = append(createdIDs, second.ID)
		})

		t.Run("case=create with invalid traits data", func(t *testing.T) {
			expected := oidcIdentity("", x.NewUUID().String())
			expected.Traits = Traits(`{"bar":123}`) // bar should be a string
			err := p.CreateIdentity(context.Background(), expected)
			require.Error(t, err)
			assert.Contains(t, fmt.Sprintf("%+v", err.Error()), "malformed")
		})

		t.Run("case=get classified credentials", func(t *testing.T) {
			initial := oidcIdentity("", x.NewUUID().String())
			initial.SetCredentials(CredentialsTypeOIDC, Credentials{
				Type: CredentialsTypeOIDC, Identifiers: []string{"aylmao-oidc"},
				Config: sqlxx.JSONRawMessage(`{"ay":"lmao"}`),
			})
			require.NoError(t, p.CreateIdentity(context.Background(), initial))
			createdIDs = append(createdIDs, initial.ID)

			initial, err := p.GetIdentityConfidential(context.Background(), initial.ID)
			require.NoError(t, err)
			require.NotEqual(t, uuid.Nil, initial.ID)
			require.NotEmpty(t, initial.Credentials)
		})

		t.Run("case=update an identity and set credentials", func(t *testing.T) {
			initial := oidcIdentity("", x.NewUUID().String())
			require.NoError(t, p.CreateIdentity(context.Background(), initial))
			createdIDs = append(createdIDs, initial.ID)

			assert.Equal(t, config.DefaultIdentityTraitsSchemaID, initial.SchemaID)
			assert.Equal(t, defaultSchema.SchemaURL(exampleServerURL).String(), initial.SchemaURL)

			expected := initial.CopyWithoutCredentials()
			expected.SetCredentials(CredentialsTypePassword, Credentials{
				Type:        CredentialsTypePassword,
				Identifiers: []string{"ignore-me"},
				Config:      sqlxx.JSONRawMessage(`{"oh":"nono"}`),
			})
			expected.Traits = Traits(`{"update":"me"}`)
			expected.SchemaID = altSchema.ID
			require.NoError(t, p.UpdateIdentity(context.Background(), expected))

			actual, err := p.GetIdentityConfidential(context.Background(), expected.ID)
			require.NoError(t, err)
			assert.Equal(t, altSchema.ID, actual.SchemaID)
			assert.Equal(t, altSchema.SchemaURL(exampleServerURL).String(), actual.SchemaURL)
			assert.NotEmpty(t, actual.Credentials[CredentialsTypePassword])
			assert.Empty(t, actual.Credentials[CredentialsTypeOIDC])

			assert.Equal(t, expected.Credentials[CredentialsTypeOIDC], actual.Credentials[CredentialsTypeOIDC])
		})

		t.Run("case=fail to update because validation fails", func(t *testing.T) {
			initial := oidcIdentity("", x.NewUUID().String())

			require.NoError(t, p.CreateIdentity(context.Background(), initial))
			createdIDs = append(createdIDs, initial.ID)

			initial.Traits = Traits(`{"bar":123}`)
			err := p.UpdateIdentity(context.Background(), initial)
			require.Error(t, err)
			require.Contains(t, err.Error(), "malformed")
		})

		t.Run("case=should fail to insert identity because credentials from traits exist", func(t *testing.T) {
			first := passwordIdentity("", "test-identity@ory.sh")
			first.Traits = Traits(`{}`)
			require.NoError(t, p.CreateIdentity(context.Background(), first))
			createdIDs = append(createdIDs, first.ID)

			second := passwordIdentity("", "test-identity@ory.sh")
			require.Error(t, p.CreateIdentity(context.Background(), second))
		})

		t.Run("case=should fail to update identity because credentials exist", func(t *testing.T) {
			first := passwordIdentity("", x.NewUUID().String())
			first.Traits = Traits(`{}`)
			require.NoError(t, p.CreateIdentity(context.Background(), first))
			createdIDs = append(createdIDs, first.ID)

			c := first.Credentials[CredentialsTypePassword]
			c.Identifiers = []string{"test-identity@ory.sh"}
			first.Credentials[CredentialsTypePassword] = c
			require.Error(t, p.UpdateIdentity(context.Background(), first))
		})

		t.Run("case=should succeed to update credentials from traits", func(t *testing.T) {
			expected := passwordIdentity("", x.NewUUID().String())
			require.NoError(t, p.CreateIdentity(context.Background(), expected))
			createdIDs = append(createdIDs, expected.ID)

			expected.Traits = Traits(`{"email":"update-test-identity@ory.sh"}`)
			require.NoError(t, p.UpdateIdentity(context.Background(), expected))

			actual, err := p.GetIdentityConfidential(context.Background(), expected.ID)
			require.NoError(t, err)

			assert.Equal(t, expected.Credentials[CredentialsTypePassword].Identifiers, actual.Credentials[CredentialsTypePassword].Identifiers)
		})

		t.Run("case=delete an identity", func(t *testing.T) {
			expected := passwordIdentity("", x.NewUUID().String())
			require.NoError(t, p.CreateIdentity(context.Background(), expected))
			require.NoError(t, p.DeleteIdentity(context.Background(), expected.ID))

			_, err := p.GetIdentity(context.Background(), expected.ID)
			require.Error(t, err)
		})

		t.Run("case=create with empty credentials config", func(t *testing.T) {
			// This test covers a case where the config value of a credentials setting is empty. This causes
			// issues with postgres' json field.
			expected := passwordIdentity("", x.NewUUID().String())
			expected.SetCredentials(CredentialsTypePassword, Credentials{
				Type:        CredentialsTypePassword,
				Identifiers: []string{"id-missing-creds-config"},
				Config:      sqlxx.JSONRawMessage(``),
			})
			require.NoError(t, p.CreateIdentity(context.Background(), expected))
			createdIDs = append(createdIDs, expected.ID)
		})

		t.Run("case=list", func(t *testing.T) {
			is, err := p.ListIdentities(context.Background(), 0, 25)
			require.NoError(t, err)
			assert.Len(t, is, len(createdIDs))
			for _, id := range createdIDs {
				var found bool
				for _, i := range is {
					if i.ID == id {
						found = true
					}
				}
				assert.True(t, found, id)
			}
		})

		t.Run("case=find identity by its credentials identifier", func(t *testing.T) {
			expected := passwordIdentity("", "find-credentials-identifier@ory.sh")
			expected.Traits = Traits(`{}`)

			require.NoError(t, p.CreateIdentity(context.Background(), expected))
			createdIDs = append(createdIDs, expected.ID)

			actual, creds, err := p.FindByCredentialsIdentifier(context.Background(), CredentialsTypePassword, "find-credentials-identifier@ory.sh")
			require.NoError(t, err)

			assert.EqualValues(t, expected.Credentials[CredentialsTypePassword].ID, creds.ID)
			assert.EqualValues(t, expected.Credentials[CredentialsTypePassword].Identifiers, creds.Identifiers)
			assert.JSONEq(t, string(expected.Credentials[CredentialsTypePassword].Config), string(creds.Config))
			// assert.EqualValues(t, expected.Credentials[CredentialsTypePassword].CreatedAt.Unix(), creds.CreatedAt.Unix())
			// assert.EqualValues(t, expected.Credentials[CredentialsTypePassword].UpdatedAt.Unix(), creds.UpdatedAt.Unix())

			expected.Credentials = nil
			assertEqual(t, expected, actual)
		})

		t.Run("case=find identity by its credentials case insensitive emails", func(t *testing.T) {
			expected := passwordIdentity("", "cAsEinSensiTivE@ory.Sh")
			expected.Traits = Traits(`{}`)

			require.NoError(t, p.CreateIdentity(context.Background(), expected))
			createdIDs = append(createdIDs, expected.ID)

			actual, creds, err := p.FindByCredentialsIdentifier(context.Background(), CredentialsTypePassword, "CASEINSENSITIVE@ory.Sh")
			require.NoError(t, err)

			assert.EqualValues(t, expected.Credentials[CredentialsTypePassword].ID, creds.ID)
			assert.EqualValues(t, []string{"caseinsensitive@ory.sh"}, creds.Identifiers)
			assert.JSONEq(t, string(expected.Credentials[CredentialsTypePassword].Config), string(creds.Config))

			expected.Credentials = nil
			assertEqual(t, expected, actual)
		})

		t.Run("suite=verifiable-address", func(t *testing.T) {
			createIdentityWithAddresses := func(t *testing.T, email string) VerifiableAddress {
				var i Identity
				require.NoError(t, faker.FakeData(&i))

				address := NewVerifiableEmailAddress(email, i.ID)
				i.VerifiableAddresses = append(i.VerifiableAddresses, *address)

				require.NoError(t, p.CreateIdentity(context.Background(), &i))
				return i.VerifiableAddresses[0]
			}

			t.Run("case=not found", func(t *testing.T) {
				_, err := p.FindVerifiableAddressByValue(context.Background(), VerifiableAddressTypeEmail, "does-not-exist")
				require.Equal(t, sqlcon.ErrNoRows, errorsx.Cause(err))
			})

			t.Run("case=create and find", func(t *testing.T) {
				addresses := make([]VerifiableAddress, 15)
				for k := range addresses {
					addresses[k] = createIdentityWithAddresses(t, "recovery.TestPersister.Create"+strconv.Itoa(k)+"@ory.sh")
					require.NotEmpty(t, addresses[k].ID)
				}

				compare := func(t *testing.T, expected, actual VerifiableAddress) {
					actual.CreatedAt = actual.CreatedAt.UTC().Truncate(time.Hour * 24)
					actual.UpdatedAt = actual.UpdatedAt.UTC().Truncate(time.Hour * 24)
					expected.CreatedAt = expected.CreatedAt.UTC().Truncate(time.Hour * 24)
					expected.UpdatedAt = expected.UpdatedAt.UTC().Truncate(time.Hour * 24)
					assert.EqualValues(t, expected, actual)
				}

				for k, expected := range addresses {
					t.Run("method=FindVerifiableAddressByValue", func(t *testing.T) {
						t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
							actual, err := p.FindVerifiableAddressByValue(context.Background(), expected.Via, expected.Value)
							require.NoError(t, err)
							compare(t, expected, *actual)
						})
					})
				}
			})

			t.Run("case=update", func(t *testing.T) {
				address := createIdentityWithAddresses(t, "verification.TestPersister.Update@ory.sh")

				address.Value = "new-code"
				require.NoError(t, p.UpdateVerifiableAddress(context.Background(), &address))

				actual, err := p.FindVerifiableAddressByValue(context.Background(), address.Via, address.Value)
				require.NoError(t, err)
				assert.Equal(t, "new-code", actual.Value)
			})

			t.Run("case=create and update and find", func(t *testing.T) {
				var i Identity
				require.NoError(t, faker.FakeData(&i))

				address := NewVerifiableEmailAddress("verification.TestPersister.Update-Identity@ory.sh", i.ID)
				i.VerifiableAddresses = append(i.VerifiableAddresses, *address)
				require.NoError(t, p.CreateIdentity(context.Background(), &i))

				_, err := p.FindVerifiableAddressByValue(context.Background(), VerifiableAddressTypeEmail, "verification.TestPersister.Update-Identity@ory.sh")
				require.NoError(t, err)

				address = NewVerifiableEmailAddress("verification.TestPersister.Update-Identity-next@ory.sh", i.ID)
				i.VerifiableAddresses = []VerifiableAddress{*address}
				require.NoError(t, p.UpdateIdentity(context.Background(), &i))

				_, err = p.FindVerifiableAddressByValue(context.Background(), VerifiableAddressTypeEmail, "verification.TestPersister.Update-Identity@ory.sh")
				require.EqualError(t, err, sqlcon.ErrNoRows.Error())

				actual, err := p.FindVerifiableAddressByValue(context.Background(), VerifiableAddressTypeEmail, "verification.TestPersister.Update-Identity-next@ory.sh")
				require.NoError(t, err)

				assert.Equal(t, VerifiableAddressTypeEmail, actual.Via)
				assert.Equal(t, "verification.TestPersister.Update-Identity-next@ory.sh", actual.Value)
			})
		})

		t.Run("suite=recovery-address", func(t *testing.T) {
			createIdentityWithAddresses := func(t *testing.T, email string) *Identity {
				var i Identity
				require.NoError(t, faker.FakeData(&i))
				i.Traits = []byte(`{"email":"` + email + `"}`)
				address := NewRecoveryEmailAddress(email, i.ID)
				i.RecoveryAddresses = append(i.RecoveryAddresses, *address)
				require.NoError(t, p.CreateIdentity(context.Background(), &i))
				return &i
			}

			t.Run("case=not found", func(t *testing.T) {
				_, err := p.FindRecoveryAddressByValue(context.Background(), RecoveryAddressTypeEmail, "does-not-exist")
				require.Equal(t, sqlcon.ErrNoRows, errorsx.Cause(err))
			})

			t.Run("case=create and find", func(t *testing.T) {
				addresses := make([]RecoveryAddress, 15)
				for k := range addresses {
					addresses[k] = createIdentityWithAddresses(t, "recovery.TestPersister.Create"+strconv.Itoa(k)+"@ory.sh").RecoveryAddresses[0]
					require.NotEmpty(t, addresses[k].ID)
				}

				compare := func(t *testing.T, expected, actual RecoveryAddress) {
					actual.CreatedAt = actual.CreatedAt.UTC().Truncate(time.Hour * 24)
					actual.UpdatedAt = actual.UpdatedAt.UTC().Truncate(time.Hour * 24)
					expected.CreatedAt = expected.CreatedAt.UTC().Truncate(time.Hour * 24)
					expected.UpdatedAt = expected.UpdatedAt.UTC().Truncate(time.Hour * 24)
					assert.EqualValues(t, expected, actual)
				}

				for k, expected := range addresses {
					t.Run("method=FindVerifiableAddressByValue", func(t *testing.T) {
						t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
							actual, err := p.FindRecoveryAddressByValue(context.Background(), expected.Via, expected.Value)
							require.NoError(t, err)
							compare(t, expected, *actual)
						})
					})
				}
			})

			t.Run("case=create and update and find", func(t *testing.T) {
				identity := createIdentityWithAddresses(t, "recovery.TestPersister.Update@ory.sh")

				_, err := p.FindRecoveryAddressByValue(context.Background(), RecoveryAddressTypeEmail, "recovery.TestPersister.Update@ory.sh")
				require.NoError(t, err)

				identity.RecoveryAddresses = []RecoveryAddress{{Via: RecoveryAddressTypeEmail, Value: "recovery.TestPersister.Update-next@ory.sh"}}
				require.NoError(t, p.UpdateIdentity(context.Background(), identity))

				_, err = p.FindRecoveryAddressByValue(context.Background(), RecoveryAddressTypeEmail, "recovery.TestPersister.Update@ory.sh")
				require.EqualError(t, err, sqlcon.ErrNoRows.Error())

				actual, err := p.FindRecoveryAddressByValue(context.Background(), RecoveryAddressTypeEmail, "recovery.TestPersister.Update-next@ory.sh")
				require.NoError(t, err)

				assert.Equal(t, RecoveryAddressTypeEmail, actual.Via)
				assert.Equal(t, "recovery.TestPersister.Update-next@ory.sh", actual.Value)
			})
		})
	}
}
