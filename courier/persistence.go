package courier

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/ory/x/pagination/keysetpagination"
	"github.com/pkg/errors"
)

var ErrQueueEmpty = errors.New("queue is empty")

type (
	Persister interface {
		AddMessage(context.Context, *Message) error

		NextMessages(context.Context, uint8) ([]Message, error)

		SetMessageStatus(context.Context, uuid.UUID, MessageStatus) error

		LatestQueuedMessage(ctx context.Context) (*Message, error)

		IncrementMessageSendCount(context.Context, uuid.UUID) error

		// ListMessages lists all messages in the store given the page, itemsPerPage, status and recipient.
		// Returns list of messages, total count of messages satisfied by given filter, and error if any
		ListMessages(context.Context, MessagesFilter, []keysetpagination.Option) ([]Message, int64, *keysetpagination.Paginator, error)
	}
	PersistenceProvider interface {
		CourierPersister() Persister
	}
)
