package verification

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ory/kratos/ui/container"

	"github.com/davecgh/go-spew/spew"

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
	FlowPersistenceProvider interface {
		VerificationFlowPersister() FlowPersister
	}
	FlowPersister interface {
		CreateVerificationFlow(context.Context, *Flow) error
		GetVerificationFlow(ctx context.Context, id uuid.UUID) (*Flow, error)
		UpdateVerificationFlow(context.Context, *Flow) error
	}
)

func TestFlowPersister(ctx context.Context, conf *config.Config, p interface {
	FlowPersister
	identity.PrivilegedPool
}) func(t *testing.T) {
	var clearids = func(r *Flow) {
		r.ID = uuid.UUID{}
	}

	return func(t *testing.T) {
		conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/identity.schema.json")

		t.Run("case=should error when the verification request does not exist", func(t *testing.T) {
			_, err := p.GetVerificationFlow(ctx, x.NewUUID())
			require.Error(t, err)
		})

		var newFlow = func(t *testing.T) *Flow {
			var r Flow
			require.NoError(t, faker.FakeData(&r))
			clearids(&r)
			require.Len(t, r.Methods, 1)
			return &r
		}

		t.Run("case=should create a new verification flow", func(t *testing.T) {
			r := newFlow(t)
			err := p.CreateVerificationFlow(ctx, r)
			require.NoError(t, err, "%#v", err)
		})

		t.Run("case=should create with set ids", func(t *testing.T) {
			var r Flow
			require.NoError(t, faker.FakeData(&r))
			require.NoError(t, p.CreateVerificationFlow(ctx, &r))
		})

		t.Run("case=should create and fetch a verification request", func(t *testing.T) {
			expected := newFlow(t)
			err := p.CreateVerificationFlow(ctx, expected)
			require.NoError(t, err)

			actual, err := p.GetVerificationFlow(ctx, expected.ID)
			require.NoError(t, err)

			fexpected, _ := json.Marshal(expected.Methods[StrategyVerificationLinkName].Config)
			factual, _ := json.Marshal(actual.Methods[StrategyVerificationLinkName].Config)

			require.NotEmpty(t, actual.Methods[StrategyVerificationLinkName].Config.FlowMethodConfigurator.(*container.Container).Action)
			assert.EqualValues(t, expected.ID, actual.ID)
			assert.JSONEq(t, string(fexpected), string(factual))
			x.AssertEqualTime(t, expected.IssuedAt, actual.IssuedAt)
			x.AssertEqualTime(t, expected.ExpiresAt, actual.ExpiresAt)
			assert.EqualValues(t, expected.RequestURL, actual.RequestURL)
		})

		t.Run("case=should create and update a verification request", func(t *testing.T) {
			expected := newFlow(t)
			expected.Methods[StrategyVerificationLinkName] = &FlowMethod{
				Method: StrategyVerificationLinkName, Config: &FlowMethodConfig{FlowMethodConfigurator: &container.Container{
					Nodes: node.Nodes{
						// v0.5: {Name: "zab", Type: "bar", Pattern: "baz"},
						node.NewInputField("zab", nil, node.DefaultGroup, "bar", node.WithInputAttributes(func(a *node.InputAttributes) {
							a.Pattern = "baz"
						})),
					}}}}
			expected.Methods["password"] = &FlowMethod{
				Method: "password", Config: &FlowMethodConfig{FlowMethodConfigurator: &container.Container{
					Nodes: node.Nodes{
						// v0.5: {Name: "foo", Type: "bar", Pattern: "baz"},
						node.NewInputField("foo", nil, node.DefaultGroup, "bar", node.WithInputAttributes(func(a *node.InputAttributes) {
							a.Pattern = "baz"
						})),
					}}}}

			spew.Dump(expected)
			err := p.CreateVerificationFlow(ctx, expected)
			require.NoError(t, err)

			expected.Methods[StrategyVerificationLinkName].Config.FlowMethodConfigurator.(*container.Container).Action = "/new-action"
			expected.Methods["password"].Config.FlowMethodConfigurator.(*container.Container).Nodes = node.Nodes{
				// {Name: "zab", Type: "zab", Pattern: "zab"},
				node.NewInputField("zab", nil, node.DefaultGroup, "zab", node.WithInputAttributes(func(a *node.InputAttributes) {
					a.Pattern = "zab"
				})),
			}
			expected.RequestURL = "/new-request-url"
			expected.Active = StrategyVerificationLinkName
			expected.Messages.Add(text.NewVerificationEmailSent())
			require.NoError(t, p.UpdateVerificationFlow(ctx, expected))

			actual, err := p.GetVerificationFlow(ctx, expected.ID)
			require.NoError(t, err)

			assert.Equal(t, "/new-action", actual.Methods[StrategyVerificationLinkName].Config.FlowMethodConfigurator.(*container.Container).Action)
			assert.Equal(t, "/new-request-url", actual.RequestURL)
			assert.Equal(t, StrategyVerificationLinkName, actual.Active.String())
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
				Methods[StrategyVerificationLinkName].Config.FlowMethodConfigurator.(*container.Container).Nodes)
		})

		t.Run("case=should not cause data loss when updating a request without changes", func(t *testing.T) {
			expected := newFlow(t)
			err := p.CreateVerificationFlow(ctx, expected)
			require.NoError(t, err)

			actual, err := p.GetVerificationFlow(ctx, expected.ID)
			require.NoError(t, err)
			assert.Len(t, actual.Methods, 1)

			require.NoError(t, p.UpdateVerificationFlow(ctx, actual))

			actual, err = p.GetVerificationFlow(ctx, expected.ID)
			require.NoError(t, err)
			require.Len(t, actual.Methods, 1)

			js, _ := json.Marshal(actual.Methods)
			assert.Equal(t, expected.Methods[StrategyVerificationLinkName].Config.FlowMethodConfigurator.(*container.Container).Action,
				actual.Methods[StrategyVerificationLinkName].Config.FlowMethodConfigurator.(*container.Container).Action, "%s", js)
		})
	}
}
