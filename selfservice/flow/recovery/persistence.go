package recovery

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ory/kratos/ui/container"

	"github.com/ory/kratos/ui/node"

	"github.com/bxcodec/faker/v3"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"

	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

type (
	FlowPersister interface {
		CreateRecoveryFlow(context.Context, *Flow) error
		GetRecoveryFlow(ctx context.Context, id uuid.UUID) (*Flow, error)
		UpdateRecoveryFlow(context.Context, *Flow) error
	}
	FlowPersistenceProvider interface {
		RecoveryFlowPersister() FlowPersister
	}
)

func TestFlowPersister(conf *config.Config, p interface {
	FlowPersister
	identity.PrivilegedPool
}) func(t *testing.T) {
	var clearids = func(r *Flow) {
		r.ID = uuid.UUID{}
	}

	ctx := context.Background()
	return func(t *testing.T) {
		conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/identity.schema.json")

		t.Run("case=should error when the recovery request does not exist", func(t *testing.T) {
			_, err := p.GetRecoveryFlow(ctx, x.NewUUID())
			require.Error(t, err)
		})

		var newFlow = func(t *testing.T) *Flow {
			var r Flow
			require.NoError(t, faker.FakeData(&r))
			clearids(&r)
			return &r
		}

		t.Run("case=should create a new recovery request", func(t *testing.T) {
			r := newFlow(t)
			err := p.CreateRecoveryFlow(ctx, r)
			require.NoError(t, err, "%#v", err)
		})

		t.Run("case=should create with set ids", func(t *testing.T) {
			var r Flow
			require.NoError(t, faker.FakeData(&r))
			require.NoError(t, p.CreateRecoveryFlow(ctx, &r))
		})

		t.Run("case=should create and fetch a recovery request", func(t *testing.T) {
			expected := newFlow(t)
			err := p.CreateRecoveryFlow(ctx, expected)
			require.NoError(t, err)

			actual, err := p.GetRecoveryFlow(ctx, expected.ID)
			require.NoError(t, err)

			fexpected, _ := json.Marshal(expected.Methods[StrategyRecoveryLinkName].Config)
			factual, _ := json.Marshal(actual.Methods[StrategyRecoveryLinkName].Config)

			require.NotEmpty(t, actual.Methods[StrategyRecoveryLinkName].Config.FlowMethodConfigurator.(*container.Container).Action)
			assert.EqualValues(t, expected.ID, actual.ID)
			assert.JSONEq(t, string(fexpected), string(factual))
			x.AssertEqualTime(t, expected.IssuedAt, actual.IssuedAt)
			x.AssertEqualTime(t, expected.ExpiresAt, actual.ExpiresAt)
			assert.EqualValues(t, expected.RequestURL, actual.RequestURL)
		})

		t.Run("case=should create and update a recovery request", func(t *testing.T) {
			expected := newFlow(t)
			expected.Methods[StrategyRecoveryLinkName] = &FlowMethod{
				Method: StrategyRecoveryLinkName, Config: &FlowMethodConfig{FlowMethodConfigurator: &container.Container{Nodes: node.Nodes{
					// v0.5: {Name: "zab", Type: "bar", Pattern: "baz"},
					node.NewInputField("zab", nil, node.DefaultGroup, "bar", node.WithInputAttributes(func(a *node.InputAttributes) {
						a.Pattern = "baz"
					})),
				}}}}
			expected.Methods["password"] = &FlowMethod{
				Method: "password", Config: &FlowMethodConfig{FlowMethodConfigurator: &container.Container{Nodes: node.Nodes{
					// v0.5: {Name: "foo", Type: "bar", Pattern: "baz"},
					node.NewInputField("foo", nil, node.DefaultGroup, "bar", node.WithInputAttributes(func(a *node.InputAttributes) {
						a.Pattern = "baz"
					})),
				}}}}
			err := p.CreateRecoveryFlow(ctx, expected)
			require.NoError(t, err)

			expected.Methods[StrategyRecoveryLinkName].Config.FlowMethodConfigurator.(*container.Container).Action = "/new-action"
			expected.Methods["password"].Config.FlowMethodConfigurator.(*container.Container).Nodes = node.Nodes{
				// v0.5: {Name: "zab", Type: "zab", Pattern: "zab"},
				node.NewInputField("zab", nil, node.DefaultGroup, "zab", node.WithInputAttributes(func(a *node.InputAttributes) {
					a.Pattern = "zab"
				})),
			}
			expected.RequestURL = "/new-request-url"
			expected.Active = StrategyRecoveryLinkName
			expected.Messages.Add(text.NewRecoveryEmailSent())
			require.NoError(t, p.UpdateRecoveryFlow(ctx, expected))

			actual, err := p.GetRecoveryFlow(ctx, expected.ID)
			require.NoError(t, err)

			assert.Equal(t, "/new-action", actual.Methods[StrategyRecoveryLinkName].Config.FlowMethodConfigurator.(*container.Container).Action)
			assert.Equal(t, "/new-request-url", actual.RequestURL)
			assert.Equal(t, StrategyRecoveryLinkName, actual.Active.String())
			assert.Equal(t, expected.Messages, actual.Messages)
			assert.EqualValues(t, node.Nodes{
				// v0.5: {Name: "zab", Type: "zab", Pattern: "zab"},
				node.NewInputField("zab", nil, node.DefaultGroup, "zab", node.WithInputAttributes(func(a *node.InputAttributes) {
					a.Pattern = "zab"
				})),
			}, actual.
				Methods["password"].Config.FlowMethodConfigurator.(*container.Container).Nodes)
			assert.EqualValues(t, node.Nodes{
				// v0.5: {Name: "zab", Type: "bar", Pattern: "baz"},
				node.NewInputField("zab", nil, node.DefaultGroup, "bar", node.WithInputAttributes(func(a *node.InputAttributes) {
					a.Pattern = "baz"
				})),
			}, actual.
				Methods[StrategyRecoveryLinkName].Config.FlowMethodConfigurator.(*container.Container).Nodes)
		})

		t.Run("case=should not cause data loss when updating a request without changes", func(t *testing.T) {
			expected := newFlow(t)
			err := p.CreateRecoveryFlow(ctx, expected)
			require.NoError(t, err)

			actual, err := p.GetRecoveryFlow(ctx, expected.ID)
			require.NoError(t, err)
			assert.Len(t, actual.Methods, 1)

			require.NoError(t, p.UpdateRecoveryFlow(ctx, actual))

			actual, err = p.GetRecoveryFlow(ctx, expected.ID)
			require.NoError(t, err)
			require.Len(t, actual.Methods, 1)

			js, _ := json.Marshal(actual.Methods)
			assert.Equal(t, expected.Methods[StrategyRecoveryLinkName].Config.FlowMethodConfigurator.(*container.Container).Action,
				actual.Methods[StrategyRecoveryLinkName].Config.FlowMethodConfigurator.(*container.Container).Action, "%s", js)
		})
	}
}
