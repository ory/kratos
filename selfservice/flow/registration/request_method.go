package registration

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/persistence/aliases"
	"github.com/ory/kratos/selfservice/form"
)

// swagger:model registrationRequestMethod
type RequestMethod struct {
	// Method contains the request credentials type.
	Method identity.CredentialsType `json:"method"`

	// Config is the credential type's config.
	Config *RequestMethodConfig `json:"config"`

	// ID is a helper struct field for gobuffalo.pop.
	ID uuid.UUID `json:"-" db:"id"`
	// RequestID is a helper struct field for gobuffalo.pop.
	RequestID uuid.UUID `json:"-" db:"registration_request_id"`
	// Request is a helper struct field for gobuffalo.pop.
	Request *Request `json:"-" belongs_to:"registration_request" fk_id:"registration_request_id"`
}

func (u *RequestMethod) TableName() string {
	return "registration_request_methods"
}

// swagger:ignore
type RequestMethodConfigurator interface {
	form.ErrorParser
	form.FieldSetter
	form.ValueSetter
	form.Resetter
	form.CSRFSetter
}

// swagger:model registrationRequestMethodConfig
type RequestMethodConfig struct {
	RequestMethodConfigurator
}

func (c *RequestMethodConfig) Scan(value interface{}) error {
	return aliases.JSONScan(c, value)
}

func (c *RequestMethodConfig) Value() (driver.Value, error) {
	return aliases.JSONValue(c)
}

func (c *RequestMethodConfig) UnmarshalJSON(data []byte) error {
	c.RequestMethodConfigurator = new(form.HTMLForm)
	return json.Unmarshal(data, c.RequestMethodConfigurator)
}

func (c *RequestMethodConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.RequestMethodConfigurator)
}
