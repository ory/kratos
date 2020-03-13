package errorx

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
)

// swagger:model errorContainer
type ErrorContainer struct {
	ID uuid.UUID `db:"id" rw:"r" json:"id"`

	CSRFToken string `db:"csrf_token" json:"-"`

	Errors json.RawMessage `json:"errors" db:"errors"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" db:"created_at"`

	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" db:"updated_at"`

	SeenAt  sql.NullTime `json:"-" db:"seen_at"`
	WasSeen bool         `json:"-" db:"was_seen"`
}

func (e ErrorContainer) TableName() string {
	return "selfservice_errors"
}
