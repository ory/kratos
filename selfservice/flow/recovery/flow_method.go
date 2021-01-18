package recovery

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"github.com/ory/kratos/selfservice/flow"
	"time"

	"github.com/ory/kratos/corp"

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

func (u FlowMethod) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "selfservice_recovery_flow_methods")
}

type FlowMethods map[string]*FlowMethod

func (u FlowMethods) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "selfservice_recovery_flow_methods")
}

// swagger:ignore
type FlowMethodConfigurator interface {
	flow.MethodConfigurator
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
