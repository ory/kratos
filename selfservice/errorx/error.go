package errorx

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/ory/kratos/corp"

	"github.com/gofrs/uuid"
)

// swagger:model selfServiceErrorContainer
type ErrorContainer struct {
	// ID of the error container.
	//
	// required: true
	ID  uuid.UUID `db:"id" json:"id"`
	NID uuid.UUID `json:"-" db:"nid"`

	// Errors in the container
	//
	// required: true
	Errors json.RawMessage `json:"errors" db:"errors"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	SeenAt    sql.NullTime `json:"-" db:"seen_at"`
	WasSeen   bool         `json:"-" db:"was_seen"`
	CSRFToken string       `db:"csrf_token" json:"-"`
}

func (e ErrorContainer) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "selfservice_errors")
}
