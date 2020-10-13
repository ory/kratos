package link

import (
	"time"

	"github.com/gofrs/uuid"
	errors "github.com/pkg/errors"

	"github.com/ory/x/randx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/x"
)

type RecoveryToken struct {
	// ID represents the tokens's unique ID.
	//
	// required: true
	// type: string
	// format: uuid
	ID uuid.UUID `json:"id" db:"id" faker:"-"`

	// Token represents the recovery token. It can not be longer than 64 chars!
	Token string `json:"-" db:"token"`

	// RecoveryAddress links this token to a recovery address.
	// required: true
	RecoveryAddress *identity.RecoveryAddress `json:"recovery_address" belongs_to:"identity_recovery_addresses" fk_id:"RecoveryAddressID"`

	// ExpiresAt is the time (UTC) when the token expires.
	// required: true
	ExpiresAt time.Time `json:"expires_at" faker:"time_type" db:"expires_at"`

	// IssuedAt is the time (UTC) when the token was issued.
	// required: true
	IssuedAt time.Time `json:"issued_at" faker:"time_type" db:"issued_at"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`
	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`
	// RecoveryAddressID is a helper struct field for gobuffalo.pop.
	RecoveryAddressID uuid.UUID `json:"-" faker:"-" db:"identity_recovery_address_id"`
	// FlowID is a helper struct field for gobuffalo.pop.
	FlowID uuid.NullUUID `json:"-" faker:"-" db:"selfservice_recovery_flow_id"`
}

func (RecoveryToken) TableName() string {
	return "identity_recovery_tokens"
}

func NewSelfServiceRecoveryToken(address *identity.RecoveryAddress, f *recovery.Flow) *RecoveryToken {
	return &RecoveryToken{
		ID:              x.NewUUID(),
		Token:           randx.MustString(32, randx.AlphaNum),
		RecoveryAddress: address,
		ExpiresAt:       f.ExpiresAt,
		IssuedAt:        time.Now().UTC(),
		FlowID:          uuid.NullUUID{UUID: f.ID, Valid: true}}
}

func NewRecoveryToken(address *identity.RecoveryAddress, expiresIn time.Duration) *RecoveryToken {
	now := time.Now().UTC()
	return &RecoveryToken{
		ID:              x.NewUUID(),
		Token:           randx.MustString(32, randx.AlphaNum),
		RecoveryAddress: address,
		ExpiresAt:       now.Add(expiresIn),
		IssuedAt:        now,
	}
}

func (f *RecoveryToken) Valid() error {
	if f.ExpiresAt.Before(time.Now()) {
		return errors.WithStack(recovery.NewFlowExpiredError(f.ExpiresAt))
	}
	return nil
}
