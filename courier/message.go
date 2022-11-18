// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/stringsx"
)

// A Message's Status
//
// swagger:model courierMessageStatus
type MessageStatus int

const (
	MessageStatusQueued MessageStatus = iota + 1
	MessageStatusSent
	MessageStatusProcessing
	MessageStatusAbandoned
)

const (
	messageStatusQueuedText     = "queued"
	messageStatusSentText       = "sent"
	messageStatusProcessingText = "processing"
	messageStatusAbandonedText  = "abandoned"
)

func ToMessageStatus(str string) (MessageStatus, error) {
	switch s := stringsx.SwitchExact(str); {
	case s.AddCase(MessageStatusQueued.String()):
		return MessageStatusQueued, nil
	case s.AddCase(MessageStatusSent.String()):
		return MessageStatusSent, nil
	case s.AddCase(MessageStatusProcessing.String()):
		return MessageStatusProcessing, nil
	case s.AddCase(MessageStatusAbandoned.String()):
		return MessageStatusAbandoned, nil
	default:
		return 0, errors.WithStack(herodot.ErrBadRequest.WithWrap(s.ToUnknownCaseErr()).WithReason("Message status is not valid"))
	}
}

func (ms MessageStatus) String() string {
	switch ms {
	case MessageStatusQueued:
		return messageStatusQueuedText
	case MessageStatusSent:
		return messageStatusSentText
	case MessageStatusProcessing:
		return messageStatusProcessingText
	case MessageStatusAbandoned:
		return messageStatusAbandonedText
	default:
		return ""
	}
}

func (ms MessageStatus) IsValid() error {
	switch ms {
	case MessageStatusQueued, MessageStatusSent, MessageStatusProcessing, MessageStatusAbandoned:
		return nil
	default:
		return errors.WithStack(herodot.ErrBadRequest.WithReason("Message status is not valid"))
	}
}

func (ms MessageStatus) MarshalJSON() ([]byte, error) {
	if err := ms.IsValid(); err != nil {
		return nil, err
	}
	return json.Marshal(ms.String())
}

func (ms *MessageStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	s, err := ToMessageStatus(str)

	if err != nil {
		return err
	}

	*ms = s
	return nil
}

// A Message's Type
//
// It can either be `email` or `phone`
//
// swagger:model courierMessageType
type MessageType int

const (
	MessageTypeEmail MessageType = iota + 1
	MessageTypePhone
)

const (
	messageTypeEmailText = "email"
	messageTypePhoneText = "phone"
)

func ToMessageType(str string) (MessageType, error) {
	switch s := stringsx.SwitchExact(str); {
	case s.AddCase(messageTypeEmailText):
		return MessageTypeEmail, nil
	case s.AddCase(messageTypePhoneText):
		return MessageTypePhone, nil
	default:
		return 0, errors.WithStack(herodot.ErrBadRequest.WithWrap(s.ToUnknownCaseErr()).WithReason("Message type is not valid"))
	}
}

func (mt MessageType) String() string {
	switch mt {
	case MessageTypeEmail:
		return messageTypeEmailText
	case MessageTypePhone:
		return messageTypePhoneText
	default:
		return ""
	}
}

func (mt MessageType) IsValid() error {
	switch mt {
	case MessageTypeEmail, MessageTypePhone:
		return nil
	default:
		return errors.WithStack(herodot.ErrBadRequest.WithReason("Message type is not valid"))
	}
}

func (mt MessageType) MarshalJSON() ([]byte, error) {
	if err := mt.IsValid(); err != nil {
		return nil, err
	}
	return json.Marshal(mt.String())
}

func (mt *MessageType) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	t, err := ToMessageType(str)
	if err != nil {
		return err
	}

	*mt = t
	return nil
}

// swagger:model message
type Message struct {
	ID           uuid.UUID     `json:"id" faker:"-" db:"id"`
	NID          uuid.UUID     `json:"-" faker:"-" db:"nid"`
	Status       MessageStatus `json:"status" db:"status"`
	Type         MessageType   `json:"type" db:"type"`
	Recipient    string        `json:"recipient" db:"recipient"`
	Body         string        `json:"body" db:"body"`
	Subject      string        `json:"subject" db:"subject"`
	TemplateType TemplateType  `json:"template_type" db:"template_type"`
	TemplateData []byte        `json:"-" db:"template_data"`
	SendCount    int           `json:"send_count" db:"send_count"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"created_at" faker:"-" db:"created_at"`
	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"updated_at" faker:"-" db:"updated_at"`
}

func (m Message) TableName(ctx context.Context) string {
	return "courier_messages"
}

func (m *Message) GetID() uuid.UUID {
	return m.ID
}

func (m *Message) GetNID() uuid.UUID {
	return m.NID
}
