package login

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/ory/kratos/corp"

	"github.com/ory/x/sqlxx"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/form"
)

// swagger:model loginFlowMethod
type FlowMethod struct {
	// Method contains the methods' credentials type.
	//
	// required: true
	Method identity.CredentialsType `json:"method" db:"method"`

	// Config is the credential type's config.
	//
	// required: true
	Config *FlowMethodConfig `json:"config" db:"config"`

	// ID is a helper struct field for gobuffalo.pop.
	ID uuid.UUID `json:"-" db:"id"`

	// FlowID is a helper struct field for gobuffalo.pop.
	FlowID uuid.UUID `json:"-" db:"selfservice_login_flow_id"`

	// Flow is a helper struct field for gobuffalo.pop.
	Flow *Flow `json:"-" belongs_to:"selfservice_login_flow" fk_id:"FlowID"`

	// CreatedAt is a helper struct field for gobuffalo.pop.
	CreatedAt time.Time `json:"-" db:"created_at"`

	// UpdatedAt is a helper struct field for gobuffalo.pop.
	UpdatedAt time.Time `json:"-" db:"updated_at"`
}

func (u FlowMethod) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "selfservice_login_flow_methods")
}

type FlowMethods map[identity.CredentialsType]*FlowMethod

func (u FlowMethods) TableName(ctx context.Context) string {
	return corp.ContextualizeTableName(ctx, "selfservice_login_flow_methods")
}

// swagger:ignore
type FlowMethodConfigurator interface {
	form.ErrorParser
	form.ValueSetter
	form.Resetter
	form.MessageResetter
	form.CSRFSetter
	form.MessageAdder
}

// swagger:model loginFlowMethodConfig
type FlowMethodConfig struct {
	// swagger:ignore
	FlowMethodConfigurator

	FlowMethodConfigMock
}

// swagger:model loginFlowMethodConfigPayload
type FlowMethodConfigMock struct {
	*form.HTMLForm

	// Providers is set for the "oidc" flow method.
	Providers []form.Field `json:"providers"`
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
