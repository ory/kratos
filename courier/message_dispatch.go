// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"time"

	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlxx"
)

// swagger:enum CourierMessageDispatchStatus
type CourierMessageDispatchStatus string

const (
	CourierMessageDispatchStatusFailed  CourierMessageDispatchStatus = "failed"
	CourierMessageDispatchStatusSuccess CourierMessageDispatchStatus = "success"
)

// MessageDispatch represents an attempt of sending a courier message
// It contains the status of the attempt (failed or successful) and the error if any occured
//
// swagger:model messageDispatch
type MessageDispatch struct {
	// The ID of this message dispatch
	// required: true
	ID uuid.UUID `json:"id" db:"id"`

	// The ID of the message being dispatched
	// required: true
	MessageID uuid.UUID `json:"message_id" db:"message_id"`

	// The status of this dispatch
	// Either "failed" or "success"
	// required: true
	Status CourierMessageDispatchStatus `json:"status" db:"status"`

	// An optional error
	Error sqlxx.JSONRawMessage `json:"error,omitempty" db:"error"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	// required: true
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	// UpdatedAt is a helper struct field for gobuffalo.pop.
	// required: true
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	NID uuid.UUID `json:"-" db:"nid"`
}

func (MessageDispatch) TableName() string {
	return "courier_message_dispatches"
}
