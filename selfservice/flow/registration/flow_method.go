package registration

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/ory/kratos/ui/container"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/ui/node"

	"github.com/ory/kratos/corp"

	"github.com/ory/x/sqlxx"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/identity"
)

// swagger:model registrationFlowMethod
type FlowMethod struct {
	// Method contains the flow method's credentials type.
	//
	// required: true
	Method identity.CredentialsType `json:"method" faker:"string" db:"method"`

	// Config is the credential type's config.
	//
	// required: true
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

func (u FlowMethod) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "selfservice_registration_flow_methods")
}

type FlowMethods map[identity.CredentialsType]*FlowMethod

func (u FlowMethods) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "selfservice_registration_flow_methods")
}

// swagger:ignore
type FlowMethodConfigurator interface {
	flow.MethodConfigurator
}

// swagger:model registrationFlowMethodConfig
type FlowMethodConfig struct {
	// swagger:ignore
	FlowMethodConfigurator

	flowMethodConfigMock
}

// swagger:model registrationFlowMethodConfigPayload
type flowMethodConfigMock struct {
	*container.Container

	// Providers is set for the "oidc" registration method.
	Providers []node.Nodes `json:"providers" faker:"len=3"`
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
	return json.Marshal(c.FlowMethodConfigurator)
}
