package identity

import (
	"context"
	"time"

	"github.com/ory/kratos/corp"

	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlxx"
)

const (
	VerifiableAddressTypeEmail VerifiableAddressType = AddressTypeEmail

	VerifiableAddressStatusPending   VerifiableAddressStatus = "pending"
	VerifiableAddressStatusCompleted VerifiableAddressStatus = "completed"
)

type (
	// VerifiableAddressType must not exceed 16 characters as that is the limitation in the SQL Schema.
	VerifiableAddressType string

	// VerifiableAddressStatus must not exceed 16 characters as that is the limitation in the SQL Schema.
	VerifiableAddressStatus string

	// swagger:model verifiableIdentityAddress
	VerifiableAddress struct {
		// required: true
		ID uuid.UUID `json:"id" db:"id" faker:"-"`

		// required: true
		Value string `json:"value" db:"value"`

		// required: true
		Verified bool `json:"verified" db:"verified"`

		// required: true
		EmailInitiated bool `json:"email_initiated" db:"email_initiated"`

		// required: true
		Via VerifiableAddressType `json:"via" db:"via"`

		// required: true
		Status VerifiableAddressStatus `json:"status" db:"status"`

		VerifiedAt sqlxx.NullTime `json:"verified_at" faker:"-" db:"verified_at"`

		// IdentityID is a helper struct field for gobuffalo.pop.
		IdentityID uuid.UUID `json:"-" faker:"-" db:"identity_id"`
		// CreatedAt is a helper struct field for gobuffalo.pop.
		CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`
		// UpdatedAt is a helper struct field for gobuffalo.pop.
		UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`
		NID       uuid.UUID `json:"-"  faker:"-" db:"nid"`
	}
)

func (v VerifiableAddressType) HTMLFormInputType() string {
	switch v {
	case VerifiableAddressTypeEmail:
		return "email"
	}
	return ""
}

func (a VerifiableAddress) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "identity_verifiable_addresses")
}

func NewVerifiableEmailAddress(value string, identity uuid.UUID) *VerifiableAddress {
	return &VerifiableAddress{
		Value:      value,
		Verified:   false,
		Status:     VerifiableAddressStatusPending,
		Via:        VerifiableAddressTypeEmail,
		IdentityID: identity,
	}
}

func (a VerifiableAddress) GetID() uuid.UUID {
	return a.ID
}

func (a VerifiableAddress) GetNID() uuid.UUID {
	return a.NID
}
