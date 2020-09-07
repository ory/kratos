package text

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/ory/x/sqlxx"
)

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

type Message struct {
	ID       ID              `json:"id"`
	Text     string          `json:"text"`
	Type     Type            `json:"type"`
	Context  json.RawMessage `json:"context,omitempty" faker:"-"`
	I18nText string          `json:"i18nText"`
	I18nData json.RawMessage `json:"i18nData"`
}

func (m *Message) Scan(value interface{}) error {
	return sqlxx.JSONScan(m, value)
}

func (m Message) Value() (driver.Value, error) {
	return sqlxx.JSONValue(&m)
}
