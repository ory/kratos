// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sessiontokenexchange

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
)

type Codes struct {
	InitCode     string
	ReturnToCode string
}

type Exchanger struct {
	ID        uuid.UUID     `db:"id"`
	NID       uuid.UUID     `db:"nid"`
	FlowID    uuid.UUID     `db:"flow_id"`
	SessionID uuid.NullUUID `db:"session_id"`

	InitCode     string `db:"init_code"`
	ReturnToCode string `db:"return_to_code"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `db:"created_at"`

	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `db:"updated_at"`
}

func (e *Exchanger) TableName() string {
	return "session_token_exchanges"
}

type (
	Persister interface {
		CreateSessionTokenExchanger(ctx context.Context, flowID uuid.UUID) (e *Exchanger, err error)
		GetExchangerFromCode(ctx context.Context, initCode string, returnToCode string) (*Exchanger, error)
		UpdateSessionOnExchanger(ctx context.Context, flowID uuid.UUID, sessionID uuid.UUID) error
		CodeForFlow(ctx context.Context, flowID uuid.UUID) (codes *Codes, found bool, err error)
		MoveToNewFlow(ctx context.Context, oldFlow, newFlow uuid.UUID) error

		DeleteExpiredExchangers(context.Context, time.Time, int) error
	}

	PersistenceProvider interface {
		SessionTokenExchangePersister() Persister
	}
)
