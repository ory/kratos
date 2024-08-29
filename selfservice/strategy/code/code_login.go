// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"

	"github.com/ory/kratos/selfservice/flow"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/identity"
)

// swagger:ignore
type LoginCode struct {
	// ID represents the tokens's unique ID.
	//
	// required: true
	// type: string
	// format: uuid
	ID uuid.UUID `json:"id" db:"id" faker:"-"`

	// Address represents the address that the code was sent to.
	// this can be an email address or a phone number.
	Address string `json:"-" db:"address"`

	// AddressType represents the type of the address
	// this can be an email address or a phone number.
	AddressType identity.CodeChannel `json:"-" db:"address_type"`

	// CodeHMAC represents the HMACed value of the verification code
	CodeHMAC string `json:"-" db:"code"`

	// UsedAt is the timestamp of when the code was used or null if it wasn't yet
	UsedAt sql.NullTime `json:"-" db:"used_at"`

	// ExpiresAt is the time (UTC) when the token expires.
	// required: true
	ExpiresAt time.Time `json:"expires_at" faker:"time_type" db:"expires_at"`

	// IssuedAt is the time (UTC) when the token was issued.
	// required: true
	IssuedAt time.Time `json:"issued_at" faker:"time_type" db:"issued_at"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`

	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`

	// FlowID is a helper struct field for gobuffalo.pop.
	FlowID uuid.UUID `json:"-" faker:"-" db:"selfservice_login_flow_id"`

	NID        uuid.UUID `json:"-"  faker:"-" db:"nid"`
	IdentityID uuid.UUID `json:"identity_id" faker:"-" db:"identity_id"`
}

func (LoginCode) TableName(ctx context.Context) string {
	return "identity_login_codes"
}

func (f *LoginCode) Validate() error {
	if f == nil {
		return errors.WithStack(ErrCodeNotFound)
	}
	if f.ExpiresAt.Before(time.Now().UTC()) {
		return errors.WithStack(flow.NewFlowExpiredError(f.ExpiresAt))
	}
	if f.UsedAt.Valid {
		return errors.WithStack(ErrCodeAlreadyUsed)
	}
	return nil
}

func (f *LoginCode) GetHMACCode() string {
	return f.CodeHMAC
}

func (f *LoginCode) GetID() uuid.UUID {
	return f.ID
}

// swagger:ignore
type CreateLoginCodeParams struct {
	// Address is the email address or phone number the code should be sent to.
	// required: true
	Address string

	// AddressType is the type of the address (email or phone number).
	// required: true
	AddressType identity.CodeChannel

	// Code represents the recovery code
	// required: true
	RawCode string

	// ExpiresAt is the time (UTC) when the code expires.
	// required: true
	ExpiresIn time.Duration

	// FlowID is a helper struct field for gobuffalo.pop.
	// required: true
	FlowID uuid.UUID

	// IdentityID is the identity that this code is for
	// required: true
	IdentityID uuid.UUID
}
