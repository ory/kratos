package code

import (
	"database/sql"

	"github.com/gofrs/uuid"
)

type LoginRegistrationCode struct {
	// ID is the primary key
	//
	// required: true
	// type: string
	// format: uuid
	ID uuid.UUID `json:"id" db:"id" faker:"-"`

	// CodeHMAC represents the HMACed value of the login/registration code
	CodeHMAC string `json:"-" db:"code_hmac"`

	// UsedAt is the timestamp of when the code was used or null if it wasn't yet
	UsedAt sql.NullTime `json:"-" db:"used_at"`
}
