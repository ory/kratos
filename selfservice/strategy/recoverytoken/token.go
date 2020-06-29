package recoverytoken

import (
	"time"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/x"
)

type Token struct {
	// ID represents the tokens's unique ID.
	//
	// required: true
	// type: string
	// format: uuid
	ID uuid.UUID `json:"id" db:"id" faker:"-"`

	// Token represents the recovery token. It can not be longer than 64 chars!
	Token string `json:"-" db:"token"`

	// RecoveryAddress links this token to a recovery address.
	RecoveryAddress *identity.RecoveryAddress `json:"recovery_address" belongs_to:"identity_recovery_addresses" fk_id:"RecoveryAddressID"`

	// RecoveryAddress links this token to a recovery request.
	Request *recovery.Request `json:"request" belongs_to:"identity_recovery_requests" fk_id:"RequestID"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`
	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`
	// RecoveryAddressID is a helper struct field for gobuffalo.pop.
	RecoveryAddressID uuid.UUID `json:"-" faker:"-" db:"identity_recovery_address_id"`
	// RequestID is a helper struct field for gobuffalo.pop.
	RequestID uuid.UUID `json:"-" faker:"-" db:"selfservice_recovery_request_id"`
}

func (Token) TableName() string {
	return "identity_recovery_tokens"
}

func NewToken(token string, ra *identity.RecoveryAddress, req *recovery.Request) *Token {
	return &Token{
		ID:              x.NewUUID(),
		Token:           token,
		RecoveryAddress: ra,
		Request:         req,
	}
}
