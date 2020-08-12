package session

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"
)

type PersistenceProvider interface {
	SessionPersister() Persister
}

type Persister interface {
	// GetSession retrieves a session from the store.
	GetSession(ctx context.Context, sid uuid.UUID) (*Session, error)

	// CreateSession adds a session to the store.
	CreateSession(ctx context.Context, s *Session) error

	// DeleteSession removes a session from the store
	DeleteSession(ctx context.Context, id uuid.UUID) error

	// DeleteSessionsFor removes all active session from the store for the given identity.
	DeleteSessionsFor(ctx context.Context, identity uuid.UUID) error

	// GetSessionFromToken gets the session associated with the given token.
	//
	// Functionality is similar to GetSession but accepts a session token
	// instead of a session ID.
	GetSessionFromToken(context.Context, string) (*Session, error)

	// DeleteSessionFromToken deletes a session associated with the given token.
	//
	// Functionality is similar to DeleteSession but accepts a session token
	// instead of a session ID.
	DeleteSessionFromToken(context.Context, string) error
}

func TestPersister(p interface {
	Persister
	identity.PrivilegedPool
}) func(t *testing.T) {
	return func(t *testing.T) {
		viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/identity.schema.json")

		t.Run("case=not found", func(t *testing.T) {
			_, err := p.GetSession(context.Background(), x.NewUUID())
			require.Error(t, err)
		})

		t.Run("case=create session", func(t *testing.T) {
			var expected Session
			require.NoError(t, faker.FakeData(&expected))
			require.NoError(t, p.CreateIdentity(context.Background(), expected.Identity))

			assert.Equal(t, uuid.Nil, expected.ID)
			require.NoError(t, p.CreateSession(context.Background(), &expected))
			assert.NotEqual(t, uuid.Nil, expected.ID)

			check := func(actual *Session, err error) {
				require.NoError(t, err)
				assert.Equal(t, expected.Identity.ID, actual.Identity.ID)
				assert.NotEmpty(t, actual.Identity.SchemaURL)
				assert.NotEmpty(t, actual.Identity.SchemaID)
				assert.Equal(t, expected.ID, actual.ID)
				assert.Equal(t, expected.Token, actual.Token)
				assert.EqualValues(t, expected.ExpiresAt.Unix(), actual.ExpiresAt.Unix())
				assert.Equal(t, expected.AuthenticatedAt.Unix(), actual.AuthenticatedAt.Unix())
				assert.Equal(t, expected.IssuedAt.Unix(), actual.IssuedAt.Unix())
			}

			t.Run("method=get by id", func(t *testing.T) {
				check(p.GetSession(context.Background(), expected.ID))
			})

			t.Run("method=get by token", func(t *testing.T) {
				check(p.GetSessionFromToken(context.Background(), expected.Token))
			})
		})

		t.Run("case=delete session", func(t *testing.T) {
			var expected Session
			require.NoError(t, faker.FakeData(&expected))
			require.NoError(t, p.CreateIdentity(context.Background(), expected.Identity))
			require.NoError(t, p.CreateSession(context.Background(), &expected))

			require.NoError(t, p.DeleteSession(context.Background(), expected.ID))
			_, err := p.GetSession(context.Background(), expected.ID)
			require.Error(t, err)
		})

		t.Run("case=delete session from token", func(t *testing.T) {
			var expected Session
			require.NoError(t, faker.FakeData(&expected))
			require.NoError(t, p.CreateIdentity(context.Background(), expected.Identity))
			require.NoError(t, p.CreateSession(context.Background(), &expected))

			require.NoError(t, p.DeleteSessionFromToken(context.Background(), expected.Token))
			_, err := p.GetSession(context.Background(), expected.ID)
			require.Error(t, err)
		})

		t.Run("case=delete session for", func(t *testing.T) {
			var expected1 Session
			var expected2 Session
			require.NoError(t, faker.FakeData(&expected1))
			require.NoError(t, p.CreateIdentity(context.Background(), expected1.Identity))

			require.NoError(t, p.CreateSession(context.Background(), &expected1))

			require.NoError(t, faker.FakeData(&expected2))
			expected2.Identity = expected1.Identity
			expected2.IdentityID = expected1.IdentityID
			require.NoError(t, p.CreateSession(context.Background(), &expected2))

			require.NoError(t, p.DeleteSessionsFor(context.Background(), expected2.IdentityID))
			_, err := p.GetSession(context.Background(), expected1.ID)
			require.Error(t, err)
			_, err = p.GetSession(context.Background(), expected2.ID)
			require.Error(t, err)
		})
	}
}
