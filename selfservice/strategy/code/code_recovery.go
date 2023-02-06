// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"context"
	"database/sql"
	"time"

	"github.com/gofrs/uuid"

	"github.com/ory/herodot"

	"github.com/ory/kratos/identity"
)

type RecoveryCodeType int

const (
	RecoveryCodeTypeAdmin RecoveryCodeType = iota + 1
	RecoveryCodeTypeSelfService
)

var (
	ErrCodeNotFound          = herodot.ErrNotFound.WithReasonf("unknown code")
	ErrCodeAlreadyUsed       = herodot.ErrBadRequest.WithReasonf("The code was already used. Please request another code.")
	ErrCodeSubmittedTooOften = herodot.ErrBadRequest.WithReasonf("The request was submitted too often. Please request another code.")
)

type RecoveryCode struct {
	// ID represents the code's unique ID.
	//
	// required: true
	// type: string
	// format: uuid
	ID uuid.UUID `json:"id" db:"id" faker:"-"`

	// CodeHMAC represents the HMACed value of the recovery code
	CodeHMAC string `json:"-" db:"code"`

	// UsedAt is the timestamp of when the code was used or null if it wasn't yet
	UsedAt sql.NullTime `json:"-" db:"used_at"`

	// RecoveryAddress links this code to a recovery address.
	// required: true
	RecoveryAddress *identity.RecoveryAddress `json:"recovery_address" belongs_to:"identity_recovery_addresses" fk_id:"RecoveryAddressID"`

	// CodeType is the type of the code - either "admin" or "selfservice"
	CodeType RecoveryCodeType `json:"-" faker:"-" db:"code_type"`

	// ExpiresAt is the time (UTC) when the code expires.
	// required: true
	ExpiresAt time.Time `json:"expires_at" faker:"time_type" db:"expires_at"`

	// IssuedAt is the time (UTC) when the code was issued.
	// required: true
	IssuedAt time.Time `json:"issued_at" faker:"time_type" db:"issued_at"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`
	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`
	// RecoveryAddressID is a helper struct field for gobuffalo.pop.
	RecoveryAddressID uuid.NullUUID `json:"-" faker:"-" db:"identity_recovery_address_id"`
	// FlowID is a helper struct field for gobuffalo.pop.
	FlowID     uuid.UUID `json:"-" faker:"-" db:"selfservice_recovery_flow_id"`
	NID        uuid.UUID `json:"-" faker:"-" db:"nid"`
	IdentityID uuid.UUID `json:"identity_id" faker:"-" db:"identity_id"`
}

func (RecoveryCode) TableName(ctx context.Context) string {
	return "identity_recovery_codes"
}

func (f RecoveryCode) IsExpired() bool {
	return f.ExpiresAt.Before(time.Now())
}

func (r RecoveryCode) WasUsed() bool {
	return r.UsedAt.Valid
}

func (f RecoveryCode) IsValid() bool {
	return !f.IsExpired() && !f.WasUsed()
}

type CreateRecoveryCodeParams struct {
	// Code represents the recovery code
	RawCode string

	// CodeType is the type of the code - either "admin" or "selfservice"
	CodeType RecoveryCodeType

	// ExpiresAt is the time (UTC) when the code expires.
	// required: true
	ExpiresIn time.Duration

	// RecoveryAddressID is a helper struct field for gobuffalo.pop.
	RecoveryAddress *identity.RecoveryAddress

	// FlowID is a helper struct field for gobuffalo.pop.
	FlowID uuid.UUID

	IdentityID uuid.UUID
}
