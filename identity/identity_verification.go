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
	VerifiableAddressStatusSent      VerifiableAddressStatus = "sent"
	VerifiableAddressStatusCompleted VerifiableAddressStatus = "completed"
)

// VerifiableAddressType must not exceed 16 characters as that is the limitation in the SQL Schema
//
// swagger:model identityVerifiableAddressType
type VerifiableAddressType string

// VerifiableAddressStatus must not exceed 16 characters as that is the limitation in the SQL Schema
//
// swagger:model identityVerifiableAddressStatus
type VerifiableAddressStatus string

// VerifiableAddress is an identity's verifiable address
//
// swagger:model verifiableIdentityAddress
type VerifiableAddress struct {
	// The ID
	//
	// required: true
	ID uuid.UUID `json:"id" db:"id" faker:"-"`

	// The address value
	//
	// example foo@user.com
	// required: true
	Value string `json:"value" db:"value"`

	// Indicates if the address has already been verified
	//
	// example: true
	// required: true
	Verified bool `json:"verified" db:"verified"`

	// The delivery method
	//
	// enum: ["email"]
	// example: email
	// required: true
	Via VerifiableAddressType `json:"via" db:"via"`

	// The verified address status
	//
	// enum: ["pending","sent","completed"]
	// example: sent
	// required: true
	Status VerifiableAddressStatus `json:"status" db:"status"`

	// When the address was verified
	//
	// example: 2014-01-01T23:28:56.782Z
	VerifiedAt sqlxx.NullTime `json:"verified_at" faker:"-" db:"verified_at"`

	// When this entry was created
	//
	// example: 2014-01-01T23:28:56.782Z
	CreatedAt time.Time `json:"created_at" faker:"-" db:"created_at"`

	// When this entry was last updated
	//
	// example: 2014-01-01T23:28:56.782Z
	UpdatedAt time.Time `json:"updated_at" faker:"-" db:"updated_at"`

	// IdentityID is a helper struct field for gobuffalo.pop.
	IdentityID uuid.UUID `json:"-" faker:"-" db:"identity_id"`
	// CreatedAt is a helper struct field for gobuffalo.pop.
	NID uuid.UUID `json:"-"  faker:"-" db:"nid"`
}

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
