package courier

import (
	"context"
	"time"

	"github.com/ory/kratos/corp"

	"github.com/gofrs/uuid"
)

type MessageStatus int

const (
	MessageStatusQueued MessageStatus = iota + 1
	MessageStatusSent
	MessageStatusProcessing
)

type MessageType int

const (
	MessageTypeEmail MessageType = iota + 1
)

// swagger:ignore
type Message struct {
	ID           uuid.UUID     `json:"-" faker:"-" db:"id"`
	NID          uuid.UUID     `json:"-"  faker:"-" db:"nid"`
	Status       MessageStatus `json:"-" db:"status"`
	Type         MessageType   `json:"-" db:"type"`
	Recipient    string        `json:"-" db:"recipient"`
	Body         string        `json:"-" db:"body"`
	Subject      string        `json:"-" db:"subject"`
	TemplateType TemplateType  `json:"-" db:"template_type"`
	TemplateData []byte        `json:"-" db:"template_data"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`
	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`
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
