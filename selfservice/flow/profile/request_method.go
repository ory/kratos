package profile

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/persistence/aliases"
	"github.com/ory/kratos/selfservice/form"
)

// swagger:model profileRequestForm
type RequestMethod struct {
	// Method contains the request credentials type.
	Method string `json:"method" db:"method"`

	// Config is the credential type's config.
	Config *RequestMethodConfig `json:"config" db:"config"`

	// ID is a helper struct field for gobuffalo.pop.
	ID uuid.UUID `json:"-" db:"id" rw:"r"`

	// RequestID is a helper struct field for gobuffalo.pop.
	RequestID uuid.UUID `json:"-" db:"selfservice_profile_management_request_id"`

	// Request is a helper struct field for gobuffalo.pop.
	Request *Request `json:"-" belongs_to:"selfservice_profile_management_request" fk_id:"RequestID"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" db:"created_at"`

	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" db:"updated_at"`
}

func (u RequestMethod) TableName() string {
	return "selfservice_profile_management_request_methods"
}

type RequestMethodsRaw []RequestMethod // workaround for https://github.com/gobuffalo/pop/pull/478
type RequestForms map[string]*RequestMethod

func (u RequestForms) TableName() string {
	// This must be stay a value receiver, using a pointer receiver will cause issues with pop.
	return "selfservice_profile_management_request_methods"
}

func (u RequestMethodsRaw) TableName() string {
	// This must be stay a value receiver, using a pointer receiver will cause issues with pop.
	return "selfservice_profile_management_request_methods"
}

// swagger:ignore
type RequestMethodConfigurator interface {
	form.ErrorParser
	form.FieldSetter
	form.ValueSetter
	form.Resetter
	form.CSRFSetter
	form.FieldSorter
	form.ErrorAdder
}

// swagger:type profileManagementRequestConfigPayload
type RequestMethodConfig struct {
	// swagger:ignore
	RequestMethodConfigurator

	swaggerRequestMethodConfig
}

// swagger:model profileManagementRequestConfigPayload
type swaggerRequestMethodConfig struct {
	*form.HTMLForm
}

func (c *RequestMethodConfig) Scan(value interface{}) error {
	return aliases.JSONScan(c, value)
}

func (c *RequestMethodConfig) Value() (driver.Value, error) {
	return aliases.JSONValue(c)
}

func (c *RequestMethodConfig) UnmarshalJSON(data []byte) error {
	c.RequestMethodConfigurator = form.NewHTMLForm("")
	return json.Unmarshal(data, c.RequestMethodConfigurator)
}

func (c *RequestMethodConfig) MarshalJSON() ([]byte, error) {
	out, err := json.Marshal(c.RequestMethodConfigurator)
	return out, err
}
