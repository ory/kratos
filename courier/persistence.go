// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

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

		IncrementMessageSendCount(context.Context, uuid.UUID) error

		// ListMessages lists all messages in the store given the page, itemsPerPage, status and recipient.
		// Returns list of messages, total count of messages satisfied by given filter, and error if any
		ListMessages(ctx context.Context, filter ListCourierMessagesParameters) ([]Message, int64, error)
	}
	PersistenceProvider interface {
		CourierPersister() Persister
	}
)
