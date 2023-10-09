// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package errorx

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
)

// swagger:model flowError
type ErrorContainer struct {
	// ID of the error container.
	//
	// required: true
	ID  uuid.UUID `db:"id" json:"id"`
	NID uuid.UUID `json:"-" db:"nid"`

	// The error
	Errors json.RawMessage `json:"error" db:"errors"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	SeenAt    sql.NullTime `json:"-" db:"seen_at"`
	WasSeen   bool         `json:"-" db:"was_seen"`
	CSRFToken string       `db:"csrf_token" json:"-"`
}

func (e ErrorContainer) TableName(ctx context.Context) string {
	return "selfservice_errors"
}
