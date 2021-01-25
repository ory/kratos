package text

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/ory/x/sqlxx"
)

// swagger:model uiTexts
type Messages []Message

func (h *Messages) Scan(value interface{}) error {
	return sqlxx.JSONScan(h, value)
}

func (h Messages) Value() (driver.Value, error) {
	return sqlxx.JSONValue(&h)
}

func (h *Messages) Add(m *Message) Messages {
	*h = append(*h, *m)
	return *h
}

func (h *Messages) Set(m *Message) Messages {
	*h = Messages{*m}
	return *h
}

func (h *Messages) Clear() Messages {
	*h = *new(Messages)
	return *h
}

// swagger:model uiText
type Message struct {
	// The message ID.
	//
	// required: true
	ID ID `json:"id"`

	// The message text. Written in american english.
	//
	// required: true
	Text string `json:"text"`

	// The message type.
	//
	// required: true
	Type Type `json:"type"`

	// The message's context. Useful when customizing messages.
	Context json.RawMessage `json:"context,omitempty" faker:"-"`
}

func (m *Message) Scan(value interface{}) error {
	return sqlxx.JSONScan(m, value)
}

func (m Message) Value() (driver.Value, error) {
	return sqlxx.JSONValue(&m)
}
