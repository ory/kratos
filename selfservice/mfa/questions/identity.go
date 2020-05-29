package questions

import (
	"time"

	"github.com/gofrs/uuid"
)

type (
	RecoverySecurityAnswers []RecoverySecurityAnswer
	RecoverySecurityAnswer  struct {
		// required: true
		ID uuid.UUID `json:"id" db:"id" faker:"-"`

		Key    string `json:"key" db:"key"`
		Answer string `json:"answer" db:"answer"`

		// IdentityID is a helper struct field for gobuffalo.pop.
		IdentityID uuid.UUID `json:"-" faker:"-" db:"identity_id"`
		// CreatedAt is a helper struct field for gobuffalo.pop.
		CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`
		// UpdatedAt is a helper struct field for gobuffalo.pop.
		UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`
	}
)

func (a RecoverySecurityAnswers) TableName() string {
	return "identity_recovery_addresses"
}
