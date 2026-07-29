// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

// Package transaction holds the persistence transaction contracts. It is a leaf
// package so that importing it does not pull the SQL layer into API-only
// consumers such as the CLI.
package transaction

import (
	"context"

	"github.com/ory/pop/v6"
)

type (
	PersistenceProvider interface {
		TransactionalPersisterProvider() Persister
	}

	Persister interface {
		Transaction(ctx context.Context, callback func(ctx context.Context, connection *pop.Connection) error) error
	}
)
