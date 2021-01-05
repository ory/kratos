package errorx

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/ory/kratos/corp/tablename"

	"github.com/gofrs/uuid"
)

// swagger:model errorContainer
type ErrorContainer struct {
	// ID of the error container.
	//
	// required: true
	ID uuid.UUID `db:"id" json:"id"`

	CSRFToken string `db:"csrf_token" json:"-"`

	// Errors in the container
	//
	// required: true
	Errors json.RawMessage `json:"errors" db:"errors"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" db:"created_at"`

	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" db:"updated_at"`

	SeenAt  sql.NullTime `json:"-" db:"seen_at"`
	WasSeen bool         `json:"-" db:"was_seen"`
}

func (e ErrorContainer) TableName(ctx context.Context) string {
	return tablename.Contextualize(ctx, "selfservice_errors")
}
