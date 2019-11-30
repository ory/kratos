package identity

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"
	"github.com/ory/x/errorsx"

	"github.com/ory/herodot"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/x"
)

type (
	Pool interface {
		// FindByCredentialsIdentifier returns an identity by querying for it's credential identifiers.
		FindByCredentialsIdentifier(ctx context.Context, ct CredentialsType, match string) (*Identity, *Credentials, error)

		// Create creates an identity. It is capable of setting credentials without encoding. Will return an error
		// if identity exists, backend connectivity is broken, or trait validation fails.
		CreateIdentity(context.Context, *Identity) error

		ListIdentities(ctx context.Context, limit, offset int) ([]Identity, error)

		// UpdateIdentityConfidential updates an identities confidential data. It is capable of setting credentials without encoding. Will return an error
		// if identity exists, backend connectivity is broken, or trait validation fails.
		//
		// Because this will overwrite credentials you always need to update the identity using `GetClassified`.
		UpdateIdentityConfidential(context.Context, *Identity) error

		// Update updates an identity excluding its confidential data. It is capable of setting credentials without encoding. Will return an error
		// if identity exists, backend connectivity is broken, or trait validation fails.
		//
		// This update procedure works well with `Get`.
		UpdateIdentity(context.Context, *Identity) error

		// Delete removes an identity by its id. Will return an error
		// if identity exists, backend connectivity is broken, or trait validation fails.
		DeleteIdentity(context.Context, uuid.UUID) error

		// Get returns an identity by its id. Will return an error if the identity does not exist or backend
		// connectivity is broken.
		GetIdentity(context.Context, uuid.UUID) (*Identity, error)

		// GetClassified returns the identity including it's raw credentials. This should only be used internally.
		GetIdentityConfidential(context.Context, uuid.UUID) (*Identity, error)
	}

	PoolProvider interface {
		IdentityPool() Pool
	}

	abstractPool struct {
		c configuration.Provider
		d ValidationProvider
	}
)

func TestPool(p Pool) func(t *testing.T) {
	return func(t *testing.T) {
		viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/identity.schema.json")

		var createdIDs []uuid.UUID

		var passwordIdentity = func(schemaURL string, credentialsID string) *Identity {
			i := NewIdentity(schemaURL)
			i.SetCredentials(CredentialsTypePassword, Credentials{
				Type: CredentialsTypePassword, Identifiers: []string{credentialsID},
				Config: json.RawMessage(`{"foo":"bar"}`),
			})
			return i
		}

		var oidcIdentity = func(schemaURL string, credentialsID string) *Identity {
			i := NewIdentity(schemaURL)
			i.SetCredentials(CredentialsTypeOIDC, Credentials{
				Type: CredentialsTypeOIDC, Identifiers: []string{credentialsID},
				Config: json.RawMessage(`{}`),
			})
			return i
		}

		var assertEqual = func(t *testing.T, expected, actual *Identity) {
			assert.Empty(t, actual.Credentials)
			require.Equal(t, expected.Traits, actual.Traits)
			require.Equal(t, expected.ID, actual.ID)
		}

		t.Run("case=should create and set missing ID", func(t *testing.T) {
			i := NewIdentity("")
			i.SetCredentials(CredentialsTypeOIDC, Credentials{
				Type: CredentialsTypeOIDC, Identifiers: []string{x.NewUUID().String()},
				Config: json.RawMessage(`{}`),
			})
			i.ID = uuid.Nil
			require.NoError(t, p.CreateIdentity(context.Background(), i))
			assert.NotEqual(t, uuid.Nil, i.ID)
			createdIDs = append(createdIDs, i.ID)
		})

		t.Run("case=create with default values", func(t *testing.T) {
			expected := passwordIdentity("", "id-1")
			require.NoError(t, p.CreateIdentity(context.Background(), expected))
			createdIDs = append(createdIDs, expected.ID)

			actual, err := p.GetIdentity(context.Background(), expected.ID)
			require.NoError(t, err)

			assert.Equal(t, expected.ID, actual.ID)
			assert.Equal(t, "file://./stub/identity.schema.json", actual.TraitsSchemaURL)
			assertEqual(t, expected, actual)
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
			expected := passwordIdentity("file://./stub/identity-2.schema.json", "id-2")
			require.NoError(t, p.CreateIdentity(context.Background(), expected))
			createdIDs = append(createdIDs, expected.ID)

			actual, err := p.GetIdentity(context.Background(), expected.ID)
			require.NoError(t, err)
			assert.Equal(t, "file://./stub/identity-2.schema.json", actual.TraitsSchemaURL)
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
				assert.Equal(t, expected.Config, actual.Credentials[m].Config)
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
				require.Error(t, p.CreateIdentity(context.Background(), expected))

				_, err := p.GetIdentity(context.Background(), expected.ID)
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
				Config: json.RawMessage(`{"ay":"lmao"}`),
			})
			require.NoError(t, p.CreateIdentity(context.Background(), initial))
			createdIDs = append(createdIDs, initial.ID)

			initial, err := p.GetIdentityConfidential(context.Background(), initial.ID)
			require.NoError(t, err)
			require.NotEqual(t, uuid.Nil, initial.ID)
			require.NotEmpty(t, initial.Credentials)
		})

		t.Run("case=fail to update an identity because credentials changed but update was called", func(t *testing.T) {
			initial := oidcIdentity("", x.NewUUID().String())
			require.NoError(t, p.CreateIdentity(context.Background(), initial))
			createdIDs = append(createdIDs, initial.ID)

			assert.Equal(t, "file://./stub/identity.schema.json", initial.TraitsSchemaURL)

			toUpdate := initial.CopyWithoutCredentials()
			toUpdate.SetCredentials(CredentialsTypePassword, Credentials{
				Type:        CredentialsTypePassword,
				Identifiers: []string{"ignore-me"},
				Config:      json.RawMessage(`{"oh":"nono"}`),
			})
			toUpdate.Traits = Traits(`{"update":"me"}`)
			toUpdate.TraitsSchemaURL = "file://./stub/identity-2.schema.json"

			err := p.UpdateIdentity(context.Background(), toUpdate)
			require.Error(t, err)
			assert.Contains(t, fmt.Sprintf("%+v", err), "A field was modified that updates one or more credentials-related settings.")

			actual, err := p.GetIdentityConfidential(context.Background(), toUpdate.ID)
			require.NoError(t, err)
			assert.Equal(t, "file://./stub/identity.schema.json", actual.TraitsSchemaURL)
			assert.Empty(t, actual.Credentials[CredentialsTypePassword])
			assert.NotEmpty(t, actual.Credentials[CredentialsTypeOIDC])
		})

		t.Run("case=update an identity and set credentials", func(t *testing.T) {
			initial := oidcIdentity("", x.NewUUID().String())
			require.NoError(t, p.CreateIdentity(context.Background(), initial))
			createdIDs = append(createdIDs, initial.ID)

			assert.Equal(t, "file://./stub/identity.schema.json", initial.TraitsSchemaURL)

			expected := initial.CopyWithoutCredentials()
			expected.SetCredentials(CredentialsTypePassword, Credentials{
				Type:        CredentialsTypePassword,
				Identifiers: []string{"ignore-me"},
				Config:      json.RawMessage(`{"oh":"nono"}`),
			})
			expected.Traits = Traits(`{"update":"me"}`)
			expected.TraitsSchemaURL = "file://./stub/identity-2.schema.json"
			require.NoError(t, p.UpdateIdentityConfidential(context.Background(), expected))

			actual, err := p.GetIdentityConfidential(context.Background(), expected.ID)
			require.NoError(t, err)
			assert.Equal(t, "file://./stub/identity-2.schema.json", actual.TraitsSchemaURL)
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
			first := passwordIdentity("", x.NewUUID().String())
			first.Traits = Traits(`{"email":"test-identity@ory.sh"}`)
			require.NoError(t, p.CreateIdentity(context.Background(), first))
			createdIDs = append(createdIDs, first.ID)

			second := passwordIdentity("", x.NewUUID().String())
			require.NoError(t, p.CreateIdentity(context.Background(), second))
			createdIDs = append(createdIDs, second.ID)

			second.Traits = Traits(`{"email":"test-identity@ory.sh"}`)
			require.Error(t, p.UpdateIdentityConfidential(context.Background(), second))
		})

		t.Run("case=should succeed to update credentials from traits", func(t *testing.T) {
			expected := passwordIdentity("", x.NewUUID().String())
			require.NoError(t, p.CreateIdentity(context.Background(), expected))
			createdIDs = append(createdIDs, expected.ID)

			expected.Traits = Traits(`{"email":"update-test-identity@ory.sh"}`)
			require.NoError(t, p.UpdateIdentityConfidential(context.Background(), expected))

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
				Config:      json.RawMessage(``),
			})
			require.NoError(t, p.CreateIdentity(context.Background(), expected))
			createdIDs = append(createdIDs, expected.ID)
		})

		t.Run("case=list", func(t *testing.T) {
			is, err := p.ListIdentities(context.Background(), 25, 0)
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
			expected := passwordIdentity("file://./stub/identity.schema.json", x.NewUUID().String())
			expected.Traits = Traits(`{"email": "find-credentials-identifier@ory.sh"}`)

			require.NoError(t, p.CreateIdentity(context.Background(), expected))
			createdIDs = append(createdIDs, expected.ID)

			actual, creds, err := p.FindByCredentialsIdentifier(context.Background(), CredentialsTypePassword, "find-credentials-identifier@ory.sh")
			require.NoError(t, err)

			assert.EqualValues(t, expected.Credentials[CredentialsTypePassword].ID, creds.ID)
			assert.EqualValues(t, expected.Credentials[CredentialsTypePassword].Identifiers, creds.Identifiers)
			assert.EqualValues(t, expected.Credentials[CredentialsTypePassword].Config, creds.Config)
			assert.EqualValues(t, expected.Credentials[CredentialsTypePassword].CreatedAt.Unix(), creds.CreatedAt.Unix())
			assert.EqualValues(t, expected.Credentials[CredentialsTypePassword].UpdatedAt.Unix(), creds.UpdatedAt.Unix())

			expected.Credentials = nil
			assertEqual(t, expected, actual)
		})
	}
}

func newAbstractPool(c configuration.Provider, d ValidationProvider) *abstractPool {
	return &abstractPool{c: c, d: d}
}

func (p *abstractPool) augment(i Identity) *Identity {
	if i.TraitsSchemaURL == "" {
		i.TraitsSchemaURL = p.c.DefaultIdentityTraitsSchemaURL().String()
	}

	return &i
}

func (p *abstractPool) declassify(i Identity) *Identity {
	return i.CopyWithoutCredentials()
}

func (p *abstractPool) declassifyAll(i []Identity) []Identity {
	declassified := make([]Identity, len(i))
	for k, ii := range i {
		declassified[k] = *ii.CopyWithoutCredentials()
	}
	return declassified
}

func (p *abstractPool) Validate(i *Identity) error {
	if err := p.d.IdentityValidator().Validate(i); err != nil {
		if _, ok := errorsx.Cause(err).(schema.ResultErrors); ok {
			return errors.WithStack(herodot.ErrBadRequest.WithReasonf("%s", err))
		}
		return err
	}

	return nil
}
