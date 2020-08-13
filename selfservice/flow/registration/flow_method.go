package registration

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/ory/x/sqlxx"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/form"
)

// swagger:model registrationRequestMethod
type FlowMethod struct {
	// Method contains the flow method's credentials type.
	Method identity.CredentialsType `json:"method" faker:"string" db:"method"`

	// Config is the credential type's config.
	Config *FlowMethodConfig `json:"config" db:"config"`

	// ID is a helper struct field for gobuffalo.pop.
	ID uuid.UUID `json:"-" faker:"-" db:"id"`

	// FlowID is a helper struct field for gobuffalo.pop.
	FlowID uuid.UUID `json:"-" faker:"-" db:"selfservice_registration_flow_id"`

	// Flow is a helper struct field for gobuffalo.pop.
	Flow *Flow `json:"-" faker:"-" belongs_to:"selfservice_registration_flow" fk_id:"FlowID"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" faker:"-" db:"created_at"`

	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" faker:"-" db:"updated_at"`
}

func (u FlowMethod) TableName() string {
	return "selfservice_registration_flow_methods"
}

type FlowMethodsRaw []FlowMethod // workaround for https://github.com/gobuffalo/pop/pull/478
type FlowMethods map[identity.CredentialsType]*FlowMethod

func (u FlowMethods) TableName() string {
	// This must be stay a value receiver, using a pointer receiver will cause issues with pop.
	return "selfservice_registration_flow_methods"
}

func (u FlowMethodsRaw) TableName() string {
	// This must be stay a value receiver, using a pointer receiver will cause issues with pop.
	return "selfservice_registration_flow_methods"
}

// swagger:ignore
type FlowMethodConfigurator interface {
	form.ErrorParser
	form.FieldSetter
	form.FieldUnsetter
	form.ValueSetter
	form.Resetter
	form.MessageResetter
	form.CSRFSetter
	form.FieldSorter
	form.MessageAdder
}

// swagger:model registrationRequestMethodConfig
type FlowMethodConfig struct {
	// swagger:ignore
	FlowMethodConfigurator

	flowMethodConfigMock
}

// swagger:model registrationRequestMethodConfigPayload
type flowMethodConfigMock struct {
	*form.HTMLForm

	// Providers is set for the "oidc" registration method.
	Providers []form.Field `json:"providers" faker:"len=3"`
}

func (c *FlowMethodConfig) Scan(value interface{}) error {
	return sqlxx.JSONScan(c, value)
}

func (c *FlowMethodConfig) Value() (driver.Value, error) {
	return sqlxx.JSONValue(c)
}

func (c *FlowMethodConfig) UnmarshalJSON(data []byte) error {
	c.FlowMethodConfigurator = form.NewHTMLForm("")
	return json.Unmarshal(data, c.FlowMethodConfigurator)
}

func (c *FlowMethodConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.FlowMethodConfigurator)
}
