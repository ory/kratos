package recovery

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/ory/x/sqlxx"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/selfservice/form"
)

// swagger:model recoveryFlowMethod
type FlowMethod struct {
	// Method contains the request credentials type.
	//
	// required: true
	Method string `json:"method" db:"method"`

	// Config is the credential type's config.
	//
	// required: true
	Config *FlowMethodConfig `json:"config" db:"config"`

	// ID is a helper struct field for gobuffalo.pop.
	ID uuid.UUID `json:"-" db:"id"`

	// FlowID is a helper struct field for gobuffalo.pop.
	FlowID uuid.UUID `json:"-" db:"selfservice_recovery_flow_id"`

	// Flow is a helper struct field for gobuffalo.pop.
	Flow *Flow `json:"-" belongs_to:"selfservice_flow_request" fk_id:"FlowID"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" db:"created_at"`

	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" db:"updated_at"`
}

func (u FlowMethod) TableName() string {
	return "selfservice_recovery_flow_methods"
}

type FlowMethodsRaw []FlowMethod // workaround for https://github.com/gobuffalo/pop/pull/478
type FlowMethods map[string]*FlowMethod

func (u FlowMethods) TableName() string {
	// This must be stay a value receiver, using a pointer receiver will cause issues with pop.
	return "selfservice_recovery_flow_methods"
}

func (u FlowMethodsRaw) TableName() string {
	// This must be stay a value receiver, using a pointer receiver will cause issues with pop.
	return "selfservice_recovery_flow_methods"
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

// swagger:model recoveryFlowMethodConfig
type FlowMethodConfig struct {
	// swagger:ignore
	FlowMethodConfigurator

	FlowMethodConfigMock
}

// swagger:model recoveryFlowMethodConfigPayload
type FlowMethodConfigMock struct {
	*form.HTMLForm
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
	out, err := json.Marshal(c.FlowMethodConfigurator)
	return out, err
}
