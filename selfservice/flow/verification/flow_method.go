package verification

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/ory/kratos/ui/container"

	"github.com/ory/kratos/selfservice/flow"

	"github.com/ory/kratos/corp"

	"github.com/ory/x/sqlxx"

	"github.com/gofrs/uuid"
)

// swagger:model verificationFlowMethod
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
	FlowID uuid.UUID `json:"-" db:"selfservice_verification_flow_id"`

	// Flow is a helper struct field for gobuffalo.pop.
	Flow *Flow `json:"-" belongs_to:"selfservice_flow_request" fk_id:"FlowID"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" db:"created_at"`

	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" db:"updated_at"`
}

func (u FlowMethod) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "selfservice_verification_flow_methods")
}

type FlowMethods map[string]*FlowMethod

func (u FlowMethods) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "selfservice_verification_flow_methods")
}

// swagger:ignore
type FlowMethodConfigurator interface {
	flow.MethodConfigurator
}

// swagger:model verificationFlowMethodConfig
type FlowMethodConfig struct {
	// swagger:ignore
	FlowMethodConfigurator

	FlowMethodConfigMock
}

// swagger:model verificationFlowMethodConfigPayload
type FlowMethodConfigMock struct {
	*container.Container
}

func (c *FlowMethodConfig) Scan(value interface{}) error {
	return sqlxx.JSONScan(c, value)
}

func (c *FlowMethodConfig) Value() (driver.Value, error) {
	return sqlxx.JSONValue(c)
}

func (c *FlowMethodConfig) UnmarshalJSON(data []byte) error {
	c.FlowMethodConfigurator = container.New("")
	return json.Unmarshal(data, c.FlowMethodConfigurator)
}

func (c *FlowMethodConfig) MarshalJSON() ([]byte, error) {
	out, err := json.Marshal(c.FlowMethodConfigurator)
	return out, err
}
