// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"time"

	"github.com/gofrs/uuid"
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
	ID uuid.UUID `json:"id" db:"dispatch_id"`

	// The ID of the message being dispatched
	MessageID uuid.UUID `json:"message_id" db:"message_id"`

	// The status of this dispatch
	// Either "failed" or "success"
	Status CourierMessageDispatchStatus `json:"status" db:"status"`

	// An optional error
	Error string `json:"error" db:"error"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	NID uuid.UUID `json:"-" db:"nid"`
}

func (MessageDispatch) TableName() string {
	return "courier_message_dispatches"
}
