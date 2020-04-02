package continuity

import (
	"context"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/sqlcon"
	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"
)

type PersistenceProvider interface {
	ContinuityPersister() Persister
}

type Persister interface {
	SaveContinuitySession(ctx context.Context, c *Container) error
	GetContinuitySession(ctx context.Context, id uuid.UUID) (*Container, error)
	DeleteContinuitySession(ctx context.Context, id uuid.UUID) error
}

func TestPersister(p interface {
	Persister
	identity.PrivilegedPool
}) func(t *testing.T) {
	var createIdentity = func(t *testing.T) *identity.Identity {
		id := identity.Identity{ID: x.NewUUID()}
		require.NoError(t, p.CreateIdentity(context.Background(), &id))
		return &id
	}

	var createContainer = func(t *testing.T) Container {
		m := sqlxx.NullJSONRawMessage(`{"foo":"bar"}`)
		return Container{Name: "foo", IdentityID: x.PointToUUID(createIdentity(t).ID),
			ExpiresAt: time.Now().Add(time.Hour).UTC().Truncate(time.Second),
			Payload:   m,
		}
	}

	return func(t *testing.T) {
		t.Run("case=not found", func(t *testing.T) {
			_, err := p.GetContinuitySession(context.Background(), x.NewUUID())
			require.EqualError(t, err, sqlcon.ErrNoRows.Error())
		})

		t.Run("case=save and find", func(t *testing.T) {
			expected := createContainer(t)
			require.NoError(t, p.SaveContinuitySession(context.Background(), &expected))

			actual, err := p.GetContinuitySession(context.Background(), expected.ID)
			require.NoError(t, err)
			assert.EqualValues(t, expected.UTC(), actual.UTC())
		})

		t.Run("case=save and delete", func(t *testing.T) {
			expected := createContainer(t)

			require.NoError(t, p.SaveContinuitySession(context.Background(), &expected))
			require.NoError(t, p.DeleteContinuitySession(context.Background(), expected.ID))

			_, err := p.GetContinuitySession(context.Background(), expected.ID)
			require.EqualError(t, err, sqlcon.ErrNoRows.Error())
		})
	}
}
