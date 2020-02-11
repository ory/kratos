package verify

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/x/randx"
)

const (
	ViaEmail Via = "email"

	StatusPending   Status = "pending"
	StatusCompleted Status = "completed"
	// StatusDelivered = "delivered"

	// codeEntropy sets the number of characters used for generating verification codes. This must not be
	// changed to another value as we only have 32 characters available in the SQL schema.
	codeEntropy = 32
)

type (
	// Via must not exceed 16 characters as that is the limitation in the SQL Schema.
	Via string

	// Status must not exceed 16 characters as that is the limitation in the SQL Schema.
	Status string

	Address struct {
		// required: true
		ID uuid.UUID `json:"id" db:"id" faker:"uuid" rw:"r"`
		// required: true
		Value string `json:"value" db:"value"`

		// required: true
		Verified bool `json:"verified" db:"verified"`

		// required: true
		Via Via `json:"via" db:"via"`

		// required: true
		VerifiedAt *time.Time `json:"verified_at" faker:"-" db:"verified_at"`

		// required: true
		ExpiresAt time.Time `json:"expires_at" faker:"time_type" db:"expires_at"`

		// required: true
		Status Status `json:"status" db:"status"`

		// IdentityID is a helper struct field for gobuffalo.pop.
		IdentityID uuid.UUID `json:"-" faker:"-" db:"identity_id"`
		// CreatedAt is a helper struct field for gobuffalo.pop.
		CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`
		// UpdatedAt is a helper struct field for gobuffalo.pop.
		UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`
		// Code is the verification code, never to be shared as JSON
		Code string `json:"-" db:"code"`
	}
)

func (v Via) HTMLFormInputType() string {
	switch v {
	case ViaEmail:
		return "email"
	}
	return ""
}

func (a Address) TableName() string {
	return "selfservice_verification_addresses"
}

func NewVerifyCode() (string, error) {
	code, err := randx.RuneSequence(codeEntropy, randx.AlphaNum)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return string(code), nil
}

func MustNewEmailAddress(
	value string,
	identity uuid.UUID,
	expiresIn time.Duration,
) *Address {
	a, err := NewEmailAddress(value, identity, expiresIn)
	if err != nil {
		panic(err)
	}
	return a
}

func NewEmailAddress(
	value string,
	identity uuid.UUID,
	expiresIn time.Duration,
) (*Address, error) {
	code, err := NewVerifyCode()
	if err != nil {
		return nil, err
	}

	return &Address{
		Code:       code,
		Value:      value,
		Verified:   false,
		Status:     StatusPending,
		Via:        ViaEmail,
		ExpiresAt:  time.Now().Add(expiresIn).UTC(),
		IdentityID: identity,
	}, nil
}
