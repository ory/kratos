package courier

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

var ErrQueueEmpty = errors.New("queue is empty")

type (
	Persister interface {
		AddMessage(context.Context, *Message) error

		NextMessages(context.Context, uint8) ([]Message, error)

		SetMessageStatus(context.Context, uuid.UUID, MessageStatus) error

		LatestQueuedMessage(ctx context.Context) (*Message, error)
	}

	PersistenceProvider interface {
		CourierPersister() Persister
	}
)
