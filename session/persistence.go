package session

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"
)

type PersistenceProvider interface {
	SessionPersister() Persister
}

type Persister interface {
	// GetSession retrieves a session from the store.
	GetSession(ctx context.Context, sid uuid.UUID) (*Session, error)

	// ListSessionsByIdentity retrieves sessions for an identity from the store.
	ListSessionsByIdentity(ctx context.Context, iID uuid.UUID, active *bool, page, perPage int, except uuid.UUID) ([]*Session, error)

	// UpsertSession inserts or updates a session into / in the store.
	UpsertSession(ctx context.Context, s *Session) error

	// DeleteSession removes a session from the store.
	DeleteSession(ctx context.Context, id uuid.UUID) error

	// DeleteSessionsByIdentity removes all active session from the store for the given identity.
	DeleteSessionsByIdentity(ctx context.Context, identity uuid.UUID) error

	// GetSessionByToken gets the session associated with the given token.
	//
	// Functionality is similar to GetSession but accepts a session token
	// instead of a session ID.
	GetSessionByToken(context.Context, string) (*Session, error)

	// DeleteSessionByToken deletes a session associated with the given token.
	//
	// Functionality is similar to DeleteSession but accepts a session token
	// instead of a session ID.
	DeleteSessionByToken(context.Context, string) error

	// RevokeSessionByToken marks a session inactive with the given token.
	RevokeSessionByToken(ctx context.Context, token string) error

	// RevokeSession marks a given session inactive.
	RevokeSession(ctx context.Context, iID, sID uuid.UUID) error

	// RevokeSessionsIdentityExcept marks all except the given session of an identity inactive. It returns the number of sessions that were revoked.
	RevokeSessionsIdentityExcept(ctx context.Context, iID, sID uuid.UUID) (int, error)
}

func TestPersister(ctx context.Context, conf *config.Config, p interface {
	Persister
	identity.PrivilegedPool
}) func(t *testing.T) {
	return func(t *testing.T) {
		conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/identity.schema.json")

		t.Run("case=not found", func(t *testing.T) {
			_, err := p.GetSession(ctx, x.NewUUID())
			require.Error(t, err)
		})

		t.Run("case=create session", func(t *testing.T) {
			var expected Session
			require.NoError(t, faker.FakeData(&expected))
			expected.Active = true
			require.NoError(t, p.CreateIdentity(ctx, expected.Identity))

			assert.Equal(t, uuid.Nil, expected.ID)
			require.NoError(t, p.UpsertSession(ctx, &expected))
			assert.NotEqual(t, uuid.Nil, expected.ID)

			check := func(actual *Session, err error) {
				require.NoError(t, err)
				assert.Equal(t, expected.Identity.ID, actual.Identity.ID)
				assert.NotEmpty(t, actual.Identity.SchemaURL)
				assert.NotEmpty(t, actual.Identity.SchemaID)
				assert.Equal(t, expected.ID, actual.ID)
				assert.Equal(t, expected.Active, actual.Active)
				assert.Equal(t, expected.Token, actual.Token)
				assert.EqualValues(t, expected.ExpiresAt.Unix(), actual.ExpiresAt.Unix())
				assert.Equal(t, expected.AuthenticatedAt.Unix(), actual.AuthenticatedAt.Unix())
				assert.Equal(t, expected.IssuedAt.Unix(), actual.IssuedAt.Unix())
			}

			t.Run("method=get by id", func(t *testing.T) {
				check(p.GetSession(ctx, expected.ID))
			})

			t.Run("method=get by token", func(t *testing.T) {
				check(p.GetSessionByToken(ctx, expected.Token))
			})
		})

		t.Run("case=delete session", func(t *testing.T) {
			var expected Session
			require.NoError(t, faker.FakeData(&expected))
			require.NoError(t, p.CreateIdentity(ctx, expected.Identity))
			require.NoError(t, p.UpsertSession(ctx, &expected))

			require.NoError(t, p.DeleteSession(ctx, expected.ID))
			_, err := p.GetSession(ctx, expected.ID)
			require.Error(t, err)
		})

		t.Run("case=delete session by token", func(t *testing.T) {
			var expected Session
			require.NoError(t, faker.FakeData(&expected))
			require.NoError(t, p.CreateIdentity(ctx, expected.Identity))
			require.NoError(t, p.UpsertSession(ctx, &expected))

			require.NoError(t, p.DeleteSessionByToken(ctx, expected.Token))
			_, err := p.GetSession(ctx, expected.ID)
			require.Error(t, err)
		})

		t.Run("case=revoke session by token", func(t *testing.T) {
			var expected Session
			require.NoError(t, faker.FakeData(&expected))
			expected.Active = true
			require.NoError(t, p.CreateIdentity(ctx, expected.Identity))
			require.NoError(t, p.UpsertSession(ctx, &expected))

			actual, err := p.GetSession(ctx, expected.ID)
			require.NoError(t, err)
			assert.True(t, actual.Active)

			require.NoError(t, p.RevokeSessionByToken(ctx, expected.Token))

			actual, err = p.GetSession(ctx, expected.ID)
			require.NoError(t, err)
			assert.False(t, actual.Active)
		})

		t.Run("case=delete session for", func(t *testing.T) {
			var expected1 Session
			var expected2 Session
			require.NoError(t, faker.FakeData(&expected1))
			require.NoError(t, p.CreateIdentity(ctx, expected1.Identity))

			require.NoError(t, p.UpsertSession(ctx, &expected1))

			require.NoError(t, faker.FakeData(&expected2))
			expected2.Identity = expected1.Identity
			expected2.IdentityID = expected1.IdentityID
			require.NoError(t, p.UpsertSession(ctx, &expected2))

			require.NoError(t, p.DeleteSessionsByIdentity(ctx, expected2.IdentityID))
			_, err := p.GetSession(ctx, expected1.ID)
			require.Error(t, err)
			_, err = p.GetSession(ctx, expected2.ID)
			require.Error(t, err)
		})
	}
}
