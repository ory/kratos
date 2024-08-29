// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"context"

	"github.com/gobuffalo/pop/v6"
)

type (
	TransactionPersistenceProvider interface {
		TransactionalPersisterProvider() TransactionalPersister
	}

	TransactionalPersister interface {
		Transaction(ctx context.Context, callback func(ctx context.Context, connection *pop.Connection) error) error
	}
)
