// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
)

type LoginStatus string

const (
	LoginStatusSuccess LoginStatus = "success"
	LoginStatusFailure LoginStatus = "failure"
)

type LoginAttempt struct {
	// ID represents the login attempt's unique ID.
	//
	// required: true
	// type: string
	// format: uuid
	ID uuid.UUID `json:"id" db:"id" faker:"-"`

	// IdentityID represents the identity's unique ID.
	IdentityID uuid.UUID `json:"-" db:"identity_id"`

	// The status of the login attempt (success or failure).
	Status LoginStatus `json:"-" db:"status"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`
	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`
}

func (LoginAttempt) TableName(context.Context) string {
	return "selfservice_login_attempts"
}
