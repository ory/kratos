package identity

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/x/randx"
)

const (
	VerifiableAddressTypeEmail VerifiableAddressType = "email"

	VerifiableAddressStatusPending   VerifiableAddressStatus = "pending"
	VerifiableAddressStatusCompleted VerifiableAddressStatus = "completed"

	// codeEntropy sets the number of characters used for generating verification codes. This must not be
	// changed to another value as we only have 32 characters available in the SQL schema.
	codeEntropy = 32
)

type (
	// VerifiableAddressType must not exceed 16 characters as that is the limitation in the SQL Schema.
	VerifiableAddressType string

	// VerifiableAddressStatus must not exceed 16 characters as that is the limitation in the SQL Schema.
	VerifiableAddressStatus string

	// swagger:model verifiableIdentityAddress
	VerifiableAddress struct {
		// required: true
		ID uuid.UUID `json:"id" db:"id" faker:"uuid" rw:"r"`

		// required: true
		Value string `json:"value" db:"value"`

		// required: true
		Verified bool `json:"verified" db:"verified"`

		// required: true
		Via VerifiableAddressType `json:"via" db:"via"`

		VerifiedAt *time.Time `json:"verified_at" faker:"-" db:"verified_at"`

		// required: true
		ExpiresAt time.Time `json:"expires_at" faker:"time_type" db:"expires_at"`

		// IdentityID is a helper struct field for gobuffalo.pop.
		IdentityID uuid.UUID `json:"-" faker:"-" db:"identity_id"`
		// CreatedAt is a helper struct field for gobuffalo.pop.
		CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`
		// UpdatedAt is a helper struct field for gobuffalo.pop.
		UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`
		// Code is the verification code, never to be shared as JSON
		Code   string                  `json:"-" db:"code"`
		Status VerifiableAddressStatus `json:"-" db:"status"`
	}
)

func (v VerifiableAddressType) HTMLFormInputType() string {
	switch v {
	case VerifiableAddressTypeEmail:
		return "email"
	}
	return ""
}

func (a VerifiableAddress) TableName() string {
	return "identity_verifiable_addresses"
}

func NewVerifyCode() (string, error) {
	code, err := randx.RuneSequence(codeEntropy, randx.AlphaNum)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return string(code), nil
}

func NewVerifiableEmailAddress(
	value string,
	identity uuid.UUID,
	expiresIn time.Duration,
) (*VerifiableAddress, error) {
	code, err := NewVerifyCode()
	if err != nil {
		return nil, err
	}

	return &VerifiableAddress{
		Code:       code,
		Value:      value,
		Verified:   false,
		Status:     VerifiableAddressStatusPending,
		Via:        VerifiableAddressTypeEmail,
		ExpiresAt:  time.Now().Add(expiresIn).UTC(),
		IdentityID: identity,
	}, nil
}
