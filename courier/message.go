package courier

import (
	"context"
	"time"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/corp"
)

type MessageStatus int

const (
	MessageStatusQueued MessageStatus = iota + 1
	MessageStatusSent
	MessageStatusProcessing
	MessageStatusAbandoned
)

type MessageType int

const (
	MessageTypeEmail MessageType = iota + 1
	MessageTypePhone
)

// swagger:ignore
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
	return corp.ContextualizeTableName(ctx, "courier_messages")
}

func (m *Message) GetID() uuid.UUID {
	return m.ID
}

func (m *Message) GetNID() uuid.UUID {
	return m.NID
}
