// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package errorx

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
)

type (
	Persister interface {
		// CreateErrorContainer adds an error to the manager and returns a
		// unique identifier or an error if insertion fails.
		CreateErrorContainer(ctx context.Context, csrfToken string, err error) (uuid.UUID, error)

		// ReadErrorContainer returns an error by its unique identifier and
		// marks the error as read. If an error occurs during retrieval the
		// second return parameter is an error.
		ReadErrorContainer(ctx context.Context, id uuid.UUID) (*ErrorContainer, error)

		// ClearErrorContainers clears read containers that are older than a
		// certain amount of time. If force is set to true, unread errors will
		// be cleared as well.
		ClearErrorContainers(ctx context.Context, olderThan time.Duration, force bool) error
	}

	PersistenceProvider interface {
		SelfServiceErrorPersister() Persister
	}
)
