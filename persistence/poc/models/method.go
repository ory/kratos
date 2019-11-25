package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gofrs/uuid"

	"github.com/ory/kratos/persistence/aliases"
	"github.com/ory/kratos/selfservice/form"
)

type Method struct {
	ID uuid.UUID `json:"id" db:"id" rw:"r"`

	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAtF time.Time `json:"updated_at" db:"updated_at"`

	RequestID uuid.UUID `json:"-" db:"request_id"`
	Request   *Request  `json:"request" belongs_to:"request" fk_id:"request_id"`

	Config *MethodConfig `json:"config" db:"config"`
}

type MethodConfigurator interface {
	form.ErrorParser
}

var _ json.Marshaler = new(MethodConfig)
var _ json.Marshaler = new(MethodConfig)

type MethodConfig struct {
	MethodConfigurator
}

func (c *MethodConfig) Scan(value interface{}) error {
	return aliases.JSONScan(c, value)
}

func (c *MethodConfig) Value() (driver.Value, error) {
	return aliases.JSONValue(c)
}

func (c *MethodConfig) UnmarshalJSON(data []byte) error {
	c.MethodConfigurator = new(form.HTMLForm)
	return json.Unmarshal(data, c.MethodConfigurator)
}

func (c *MethodConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.MethodConfigurator)
}

// String is not required by pop and may be deleted
func (m Method) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}

// Methods is not required by pop and may be deleted
type Methods []Method

// String is not required by pop and may be deleted
func (m Methods) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (m *Method) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (m *Method) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (m *Method) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
