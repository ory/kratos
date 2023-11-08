// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/x/pagination/keysetpagination"
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
		ListMessages(context.Context, ListCourierMessagesParameters, []keysetpagination.Option) ([]Message, int64, *keysetpagination.Paginator, error)

		// FetchMessage returns a message with the id or nil and an error if not found
		FetchMessage(context.Context, uuid.UUID) (*Message, error)

		// Records an attempt of sending out a courier message
		// Returns an error if it fails
		RecordDispatch(ctx context.Context, msgID uuid.UUID, status CourierMessageDispatchStatus, err error) error
	}
	PersistenceProvider interface {
		CourierPersister() Persister
	}
)
