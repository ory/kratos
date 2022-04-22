package continuity

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
)

type PersistenceProvider interface {
	ContinuityPersister() Persister
}

type Persister interface {
	SaveContinuitySession(ctx context.Context, c *Container) error
	GetContinuitySession(ctx context.Context, id uuid.UUID) (*Container, error)
	DeleteContinuitySession(ctx context.Context, id uuid.UUID) error
	DeleteExpiredContinuitySessions(context.Context, time.Time, int) error
}
