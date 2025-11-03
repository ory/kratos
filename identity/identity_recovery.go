// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"fmt"
	"time"

	"github.com/gofrs/uuid"
)

const (
	RecoveryAddressTypeEmail RecoveryAddressType = AddressTypeEmail
	RecoveryAddressTypeSMS   RecoveryAddressType = AddressTypeSMS
)

type (
	// RecoveryAddressType must not exceed 16 characters as that is the limitation in the SQL Schema.
	RecoveryAddressType string

	// RecoveryAddressStatus must not exceed 16 characters as that is the limitation in the SQL Schema.
	RecoveryAddressStatus string

	// swagger:model recoveryIdentityAddress
	RecoveryAddress struct {
		ID uuid.UUID `json:"id" db:"id" faker:"-"`

		// required: true
		Value string `json:"value" db:"value"`

		// required: true
		Via RecoveryAddressType `json:"via" db:"via"`

		// IdentityID is a helper struct field for gobuffalo.pop.
		IdentityID uuid.UUID `json:"-" faker:"-" db:"identity_id"`
		// CreatedAt is a helper struct field for gobuffalo.pop.
		CreatedAt time.Time `json:"created_at" faker:"-" db:"created_at"`
		// UpdatedAt is a helper struct field for gobuffalo.pop.
		UpdatedAt time.Time `json:"updated_at" faker:"-" db:"updated_at"`
		NID       uuid.UUID `json:"-"  faker:"-" db:"nid"`
	}
)

func (v RecoveryAddressType) HTMLFormInputType() string {
	switch v {
	case RecoveryAddressTypeEmail:
		return "email"
	case RecoveryAddressTypeSMS:
		return "tel"
	}
	return ""
}

func (a RecoveryAddress) TableName() string { return "identity_recovery_addresses" }
func (a RecoveryAddress) GetID() uuid.UUID  { return a.ID }

// Signature returns a unique string representation for the recovery address.
func (a RecoveryAddress) Signature() string {
	return fmt.Sprintf("%v|%v|%v|%v", a.Value, a.Via, a.IdentityID, a.NID)
}

func NewRecoveryEmailAddress(
	value string,
	identity uuid.UUID,
) *RecoveryAddress {
	return &RecoveryAddress{
		Value:      value,
		Via:        RecoveryAddressTypeEmail,
		IdentityID: identity,
	}
}

func NewRecoverySMSAddress(
	value string,
	identity uuid.UUID,
) *RecoveryAddress {
	return &RecoveryAddress{
		Value:      value,
		Via:        RecoveryAddressTypeSMS,
		IdentityID: identity,
	}
}
