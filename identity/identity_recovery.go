// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"fmt"
	"time"

	"github.com/gofrs/uuid"
)

type (
	// RecoveryAddressStatus must not exceed 16 characters as that is the limitation in the SQL Schema.
	RecoveryAddressStatus string

	// swagger:model recoveryIdentityAddress
	RecoveryAddress struct {
		ID uuid.UUID `json:"id" db:"id" faker:"-"`

		// required: true
		Value string `json:"value" db:"value"`

		// required: true
		Via string `json:"via" db:"via"`

		// BreakGlassForOrganization, when set to an organization ID, allows this
		// recovery address to bypass SSO enforcement for that organization. This
		// enables designated users to recover their account via email when the
		// SSO provider is unavailable.
		BreakGlassForOrganization uuid.NullUUID `json:"break_glass_for_organization,omitzero" db:"break_glass_for_organization"`

		// IdentityID is a helper struct field for gobuffalo.pop.
		IdentityID uuid.UUID `json:"-" faker:"-" db:"identity_id"`
		// CreatedAt is a helper struct field for gobuffalo.pop.
		CreatedAt time.Time `json:"created_at" faker:"-" db:"created_at"`
		// UpdatedAt is a helper struct field for gobuffalo.pop.
		UpdatedAt time.Time `json:"updated_at" faker:"-" db:"updated_at"`
		NID       uuid.UUID `json:"-"  faker:"-" db:"nid"`
	}
)

func (a RecoveryAddress) TableName() string { return "identity_recovery_addresses" }
func (a RecoveryAddress) GetID() uuid.UUID  { return a.ID }

// Signature returns a unique string representation for the recovery address.
func (a RecoveryAddress) Signature() string {
	return fmt.Sprintf("%v|%v|%v|%v|%v", a.Value, a.Via, a.BreakGlassForOrganization, a.IdentityID, a.NID)
}

func NewRecoveryEmailAddress(
	value string,
	identity uuid.UUID,
) *RecoveryAddress {
	return &RecoveryAddress{
		Value:      value,
		Via:        AddressTypeEmail,
		IdentityID: identity,
	}
}

func NewRecoverySMSAddress(
	value string,
	identity uuid.UUID,
) *RecoveryAddress {
	return &RecoveryAddress{
		Value:      value,
		Via:        AddressTypeSMS,
		IdentityID: identity,
	}
}
