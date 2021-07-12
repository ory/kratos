package errorx

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
)

type (
	Persister interface {
		// Add adds an error to the manager and returns a unique identifier or an error if insertion fails.
		Add(ctx context.Context, csrfToken string, err error) (uuid.UUID, error)

		// Read returns an error by its unique identifier and marks the error as read. If an error occurs during retrieval
		// the second return parameter is an error.
		Read(ctx context.Context, id uuid.UUID) (*ErrorContainer, error)

		// Clear clears read containers that are older than a certain amount of time. If force is set to true, unread
		// errors will be cleared as well.
		Clear(ctx context.Context, olderThan time.Duration, force bool) error
	}

	PersistenceProvider interface {
		SelfServiceErrorPersister() Persister
	}
)
