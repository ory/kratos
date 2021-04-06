package settings

import (
	"context"
	"encoding/json"
	"github.com/ory/x/assertx"
	"testing"

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

			factual, _ := json.Marshal(actual.UI)
			fexpected, _ := json.Marshal(expected.UI)

			require.NotEmpty(t, actual.UI.Action)
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
			expected.UI.Nodes = node.Nodes{}
			expected.UI.Nodes.Append(node.NewInputField("zab", nil, node.DefaultGroup, "bar", node.WithInputAttributes(func(a *node.InputAttributes) {
				a.Pattern = "baz"
			})))

			expected.UI.Nodes.Append(node.NewInputField("foo", nil, node.DefaultGroup, "bar", node.WithInputAttributes(func(a *node.InputAttributes) {
				a.Pattern = "baz"
			})))

			err := p.CreateSettingsFlow(ctx, expected)
			require.NoError(t, err)

			expected.UI.Action = "/new-action"
			expected.UI.Nodes.Append(
				node.NewInputField("zab", nil, node.DefaultGroup, "zab", node.WithInputAttributes(func(a *node.InputAttributes) {
					a.Pattern = "zab"
				})))

			expected.RequestURL = "/new-request-url"
			require.NoError(t, p.UpdateSettingsFlow(ctx, expected))

			actual, err := p.GetSettingsFlow(ctx, expected.ID)
			require.NoError(t, err)

			assert.Equal(t, "/new-action", actual.UI.Action)
			assert.Equal(t, "/new-request-url", actual.RequestURL)
			assertx.EqualAsJSON(t, node.Nodes{
				// v0.5: {Name: "zab", Type: "zab", Pattern: "zab"},
				node.NewInputField("zab", nil, node.DefaultGroup, "bar", node.WithInputAttributes(func(a *node.InputAttributes) {
					a.Pattern = "baz"
				})),
				node.NewInputField("foo", nil, node.DefaultGroup, "bar", node.WithInputAttributes(func(a *node.InputAttributes) {
					a.Pattern = "baz"
				})),
				// v0.5: {Name: "zab", Type: "bar", Pattern: "baz"},
				node.NewInputField("zab", nil, node.DefaultGroup, "zab", node.WithInputAttributes(func(a *node.InputAttributes) {
					a.Pattern = "zab"
				})),
			}, actual.UI.Nodes)
		})
	}
}
