package identity

import (
	"time"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/otp"
)

const (
	RecoveryAddressTypeEmail RecoveryAddressType = "email"
)

type (
	// RecoveryAddressType must not exceed 16 characters as that is the limitation in the SQL Schema.
	RecoveryAddressType string

	// RecoveryAddressStatus must not exceed 16 characters as that is the limitation in the SQL Schema.
	RecoveryAddressStatus string

	// swagger:model recoveryIdentityAddress
	RecoveryAddress struct {
		// required: true
		ID uuid.UUID `json:"id" db:"id" faker:"uuid" rw:"r"`

		// required: true
		Value string `json:"value" db:"value"`

		// required: true
		Via RecoveryAddressType `json:"via" db:"via"`

		// required: true
		ExpiresAt time.Time `json:"expires_at" faker:"time_type" db:"expires_at"`

		// IdentityID is a helper struct field for gobuffalo.pop.
		IdentityID uuid.UUID `json:"-" faker:"-" db:"identity_id"`
		// CreatedAt is a helper struct field for gobuffalo.pop.
		CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`
		// UpdatedAt is a helper struct field for gobuffalo.pop.
		UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`
		// Code is the recovery code, never to be shared as JSON
		Code string `json:"-" db:"code"`
	}
)

func (v RecoveryAddressType) HTMLFormInputType() string {
	switch v {
	case RecoveryAddressTypeEmail:
		return "email"
	}
	return ""
}

func (a RecoveryAddress) TableName() string {
	return "identity_recovery_addresses"
}

func NewRecoveryEmailAddress(
	value string,
	identity uuid.UUID,
	expiresIn time.Duration,
) (*RecoveryAddress, error) {
	code, err := otp.New()
	if err != nil {
		return nil, err
	}

	return &RecoveryAddress{
		Code:       code,
		Value:      value,
		Via:        RecoveryAddressTypeEmail,
		ExpiresAt:  time.Now().Add(expiresIn).UTC(),
		IdentityID: identity,
	}, nil
}
