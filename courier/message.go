package courier

import (
	"context"
	"time"

	"github.com/ory/kratos/corp"
	"github.com/ory/kratos/driver/config"

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

type Message struct {
	ID           uuid.UUID     `json:"-" faker:"-" db:"id"`
	Status       MessageStatus `json:"-" db:"status"`
	Type         MessageType   `json:"-" db:"type"`
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

type EmailMessage struct {
	From      string
	FromName  string
	Recipient string
	Subject   string
	BodyHTML  string
}

func PopulateEmailMessage(c *config.Config, t EmailTemplate) (*EmailMessage, error) {
	from := c.CourierSMTPFrom()
	fromName := c.CourierSMTPFromName()
	recipient, err := t.EmailRecipient()
	if err != nil {
		return nil, err
	}

	subject, err := t.EmailSubject()
	if err != nil {
		return nil, err
	}

	body, err := t.EmailBody()
	if err != nil {
		return nil, err
	}
	return &EmailMessage{
		From:      from,
		FromName:  fromName,
		Recipient: recipient,
		Subject:   subject,
		BodyHTML:  body,
	}, nil
}
