// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

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
	DeleteExpiredContinuitySessions(ctx context.Context, deleteOlder time.Time, pageSize int) error
}
