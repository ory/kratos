// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"context"
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
)

type VerificationCode struct {
	// ID represents the code's unique ID.
	//
	// required: true
	// type: string
	// format: uuid
	ID uuid.UUID `json:"id" db:"id" faker:"-"`

	// CodeHMAC represents the HMACed value of the verification code
	CodeHMAC string `json:"-" db:"code_hmac"`

	// UsedAt is the timestamp of when the code was used or null if it wasn't yet
	UsedAt sql.NullTime `json:"-" db:"used_at"`

	// VerifiableAddress links this code to a verification address.
	// required: true
	VerifiableAddress *identity.VerifiableAddress `json:"verification_address" belongs_to:"identity_verifiable_addresses"`

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
	// VerifiableAddressID is a helper struct field for gobuffalo.pop.
	VerifiableAddressID uuid.NullUUID `json:"-" faker:"-" db:"identity_verifiable_address_id"`
	// FlowID is a helper struct field for gobuffalo.pop.
	FlowID uuid.UUID `json:"-" faker:"-" db:"selfservice_verification_flow_id"`
	NID    uuid.UUID `json:"-" faker:"-" db:"nid"`
}

func (VerificationCode) TableName(context.Context) string {
	return "identity_verification_codes"
}

// Validate validates the state of the verification code
//
// - If the code is expired, `flow.ExpiredError` is returned
// - If the code was already used `ErrCodeAlreadyUsed` is returnd
// - Otherwise, `nil` is returned
func (f *VerificationCode) Validate() error {
	if f.ExpiresAt.Before(time.Now().UTC()) {
		return errors.WithStack(flow.NewFlowExpiredError(f.ExpiresAt))
	}
	if f.UsedAt.Valid {
		return errors.WithStack(ErrCodeAlreadyUsed)
	}
	return nil
}

type CreateVerificationCodeParams struct {
	// Code represents the recovery code
	RawCode string

	// ExpiresIn is the lifetime of the code
	ExpiresIn time.Duration

	// VerifiableAddress is the address to be verified
	VerifiableAddress *identity.VerifiableAddress

	// FlowID is the id of the current verification flow
	FlowID uuid.UUID
}
