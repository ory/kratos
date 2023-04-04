// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlxx"
)

const (
	VerifiableAddressTypeEmail VerifiableAddressType = AddressTypeEmail
	VerifiableAddressTypePhone VerifiableAddressType = AddressTypePhone

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
	// required: false
	VerifiedAt *sqlxx.NullTime `json:"verified_at,omitempty" faker:"-" db:"verified_at"`

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
	NID        uuid.UUID `json:"-"  faker:"-" db:"nid"`
}

func (v VerifiableAddressType) HTMLFormInputType() string {
	switch v {
	case VerifiableAddressTypeEmail:
		return "email"
	case VerifiableAddressTypePhone:
		return "phone"
	}
	return ""
}

func (a VerifiableAddress) TableName(ctx context.Context) string {
	return "identity_verifiable_addresses"
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

func NewVerifiablePhoneAddress(value string, identity uuid.UUID) *VerifiableAddress {
	return &VerifiableAddress{
		Value:      value,
		Verified:   false,
		Status:     VerifiableAddressStatusPending,
		Via:        VerifiableAddressTypePhone,
		IdentityID: identity,
	}
}

func (a VerifiableAddress) GetID() uuid.UUID {
	return a.ID
}

func (a VerifiableAddress) GetNID() uuid.UUID {
	return a.NID
}

func (a VerifiableAddress) ValidateNID() error {
	return nil
}

// Hash returns a unique string representation for the recovery address.
func (a VerifiableAddress) Hash() string {
	return fmt.Sprintf("%v|%v|%v|%v|%v|%v", a.Value, a.Verified, a.Via, a.Status, a.IdentityID, a.NID)
}
