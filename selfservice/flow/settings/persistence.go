package settings

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

	"github.com/ory/kratos/x"
)

type (
	FlowPersister interface {
		CreateSettingsFlow(context.Context, *Flow) error
		GetSettingsFlow(ctx context.Context, id uuid.UUID) (*Flow, error)
		UpdateSettingsFlow(context.Context, *Flow) error
		UpdateSettingsFlowMethod(context.Context, uuid.UUID, string, *FlowMethod) error
	}
	FlowPersistenceProvider interface {
		SettingsFlowPersister() FlowPersister
	}
)

func TestRequestPersister(ctx context.Context, conf *config.Config, p interface {
	FlowPersister
	identity.PrivilegedPool
}) func(t *testing.T) {
	var clearids = func(r *Flow) {
		r.ID = uuid.UUID{}
		r.Identity.ID = uuid.UUID{}
		r.IdentityID = uuid.UUID{}
	}

	return func(t *testing.T) {
		conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/identity.schema.json")

		t.Run("case=should error when the settings request does not exist", func(t *testing.T) {
			_, err := p.GetSettingsFlow(ctx, x.NewUUID())
			require.Error(t, err)
		})

		var newFlow = func(t *testing.T) *Flow {
			var r Flow
			require.NoError(t, faker.FakeData(&r))
			clearids(&r)
			require.NoError(t, p.CreateIdentity(ctx, r.Identity))
			return &r
		}

		t.Run("case=should create a new settings request", func(t *testing.T) {
			r := newFlow(t)
			err := p.CreateSettingsFlow(ctx, r)
			require.NoError(t, err, "%#v", err)
		})

		t.Run("case=should create with set ids", func(t *testing.T) {
			var r Flow
			require.NoError(t, faker.FakeData(&r))
			require.NoError(t, p.CreateIdentity(ctx, r.Identity))
			require.NoError(t, p.CreateSettingsFlow(ctx, &r))
		})

		t.Run("case=should create and fetch a settings request", func(t *testing.T) {
			expected := newFlow(t)
			err := p.CreateSettingsFlow(ctx, expected)
			require.NoError(t, err)

			actual, err := p.GetSettingsFlow(ctx, expected.ID)
			require.NoError(t, err)

			factual, _ := json.Marshal(actual.Methods[StrategyProfile].Config)
			fexpected, _ := json.Marshal(expected.Methods[StrategyProfile].Config)

			require.NotEmpty(t, actual.Methods[StrategyProfile].Config.FlowMethodConfigurator.(*container.Container).Action)
			assert.EqualValues(t, expected.ID, actual.ID)
			assert.JSONEq(t, string(fexpected), string(factual))
			x.AssertEqualTime(t, expected.IssuedAt, actual.IssuedAt)
			x.AssertEqualTime(t, expected.ExpiresAt, actual.ExpiresAt)
			assert.EqualValues(t, expected.RequestURL, actual.RequestURL)
			assert.EqualValues(t, expected.Identity.ID, actual.Identity.ID)
			assert.EqualValues(t, expected.Identity.Traits, actual.Identity.Traits)
			assert.EqualValues(t, expected.Identity.SchemaID, actual.Identity.SchemaID)
			assert.Empty(t, actual.Identity.Credentials)
		})

		t.Run("case=should fail to create if identity does not exist", func(t *testing.T) {
			var expected Flow
			require.NoError(t, faker.FakeData(&expected))
			clearids(&expected)
			expected.Identity = nil
			expected.IdentityID = uuid.Nil
			err := p.CreateSettingsFlow(ctx, &expected)
			require.Error(t, err, "%+s", expected)
		})

		t.Run("case=should create and update a settings request", func(t *testing.T) {
			expected := newFlow(t)
			expected.Methods["oidc"] = &FlowMethod{
				Method: "oidc", Config: &FlowMethodConfig{FlowMethodConfigurator: &container.Container{
					// v0.5: Fields: []form.Field{{Name: "zab", Type: "bar", Pattern: "baz"}},
					Nodes: node.Nodes{node.NewInputField("zab", nil, node.DefaultGroup, "bar", node.WithInputAttributes(func(a *node.InputAttributes) {
						a.Pattern = "baz"
					}))},
				}}}
			expected.Methods["password"] = &FlowMethod{
				Method: "password", Config: &FlowMethodConfig{FlowMethodConfigurator: &container.Container{
					// v0.5: Fields: []form.Field{{Name: "foo", Type: "bar", Pattern: "baz"}},
					Nodes: node.Nodes{node.NewInputField("foo", nil, node.DefaultGroup, "bar", node.WithInputAttributes(func(a *node.InputAttributes) {
						a.Pattern = "baz"
					}))},
				}}}
			err := p.CreateSettingsFlow(ctx, expected)
			require.NoError(t, err)

			expected.Methods[StrategyProfile].Config.FlowMethodConfigurator.(*container.Container).Action = "/new-action"
			expected.Methods["password"].Config.FlowMethodConfigurator.(*container.Container).Nodes = node.Nodes{
				// v0.5: {Name: "zab", Type: "zab", Pattern: "zab"},
				node.NewInputField("zab", nil, node.DefaultGroup, "zab", node.WithInputAttributes(func(a *node.InputAttributes) {
					a.Pattern = "zab"
				})),
			}
			expected.RequestURL = "/new-request-url"
			require.NoError(t, p.UpdateSettingsFlow(ctx, expected))

			actual, err := p.GetSettingsFlow(ctx, expected.ID)
			require.NoError(t, err)

			assert.Equal(t, "/new-action", actual.Methods[StrategyProfile].Config.FlowMethodConfigurator.(*container.Container).Action)
			assert.Equal(t, "/new-request-url", actual.RequestURL)
			assert.EqualValues(t, node.Nodes{
				// v0.5: {Name: "zab", Type: "zab", Pattern: "zab"},
				node.NewInputField("zab", nil, node.DefaultGroup, "zab", node.WithInputAttributes(func(a *node.InputAttributes) {
					a.Pattern = "zab"
				})),
			}, actual.Methods["password"].Config.FlowMethodConfigurator.(*container.Container).Nodes)
			assert.EqualValues(t, node.Nodes{
				// v0.5: {Name: "zab", Type: "bar", Pattern: "baz"},
				node.NewInputField("zab", nil, node.DefaultGroup, "bar", node.WithInputAttributes(func(a *node.InputAttributes) {
					a.Pattern = "baz"
				})),
			}, actual.Methods["oidc"].Config.FlowMethodConfigurator.(*container.Container).Nodes)
		})

		t.Run("case=should update a settings flow method", func(t *testing.T) {
			expected := newFlow(t)
			delete(expected.Methods, identity.CredentialsTypeOIDC.String())
			delete(expected.Methods, StrategyProfile)

			err := p.CreateSettingsFlow(ctx, expected)
			require.NoError(t, err)

			actual, err := p.GetSettingsFlow(ctx, expected.ID)
			require.NoError(t, err)
			assert.Len(t, actual.Methods, 1)

			require.NoError(t, p.UpdateSettingsFlowMethod(ctx, expected.ID, identity.CredentialsTypeOIDC.String(), &FlowMethod{
				Method: identity.CredentialsTypeOIDC.String(),
				Config: &FlowMethodConfig{FlowMethodConfigurator: container.New(string(identity.CredentialsTypeOIDC))},
			}))

			require.NoError(t, p.UpdateSettingsFlowMethod(ctx, expected.ID, identity.CredentialsTypePassword.String(), &FlowMethod{
				Method: identity.CredentialsTypePassword.String(),
				Config: &FlowMethodConfig{FlowMethodConfigurator: container.New(string(identity.CredentialsTypePassword))},
			}))

			actual, err = p.GetSettingsFlow(ctx, expected.ID)
			require.NoError(t, err)
			require.Len(t, actual.Methods, 2)
			assert.EqualValues(t, identity.CredentialsTypePassword, actual.Active)

			assert.Equal(t, string(identity.CredentialsTypePassword), actual.Methods[identity.CredentialsTypePassword.String()].Config.FlowMethodConfigurator.(*container.Container).Action)
			assert.Equal(t, string(identity.CredentialsTypeOIDC), actual.Methods[identity.CredentialsTypeOIDC.String()].Config.FlowMethodConfigurator.(*container.Container).Action)
		})
	}
}
