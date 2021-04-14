package test

import (
	"context"
	"testing"
	"time"

	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/persistence"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/ui/node"

	"github.com/ory/x/assertx"

	"github.com/bxcodec/faker/v3"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/x"
)

func TestFlowPersister(ctx context.Context, p persistence.Persister) func(t *testing.T) {
	var clearids = func(r *registration.Flow) {
		r.ID = uuid.UUID{}
	}

	return func(t *testing.T) {
		_, p := testhelpers.NewNetworkUnlessExisting(t, ctx, p)

		t.Run("case=should error when the registration flow does not exist", func(t *testing.T) {
			_, err := p.GetRegistrationFlow(ctx, x.NewUUID())
			require.Error(t, err)
		})

		var newFlow = func(t *testing.T) *registration.Flow {
			var r registration.Flow
			require.NoError(t, faker.FakeData(&r))
			clearids(&r)
			return &r
		}

		t.Run("case=should create a new registration flow and properly set IDs", func(t *testing.T) {
			r := newFlow(t)
			err := p.CreateRegistrationFlow(ctx, r)
			require.NoError(t, err, "%#v", err)

			assert.NotEqual(t, uuid.Nil, r.ID)
		})

		t.Run("case=should create with set ids", func(t *testing.T) {
			var r registration.Flow
			require.NoError(t, faker.FakeData(&r))
			require.NoError(t, p.CreateRegistrationFlow(ctx, &r))
		})

		t.Run("case=should create and fetch a registration flow", func(t *testing.T) {
			expected := newFlow(t)
			err := p.CreateRegistrationFlow(ctx, expected)
			require.NoError(t, err)

			actual, err := p.GetRegistrationFlow(ctx, expected.ID)
			require.NoError(t, err)

			assert.EqualValues(t, expected.ID, actual.ID)
			x.AssertEqualTime(t, expected.IssuedAt, actual.IssuedAt)
			x.AssertEqualTime(t, expected.ExpiresAt, actual.ExpiresAt)
			assert.EqualValues(t, expected.RequestURL, actual.RequestURL)
			assert.EqualValues(t, expected.Active, actual.Active)
			assertx.EqualAsJSON(t, expected.UI, actual.UI, "expected:\t%s\nactual:\t%s", expected.UI, actual.UI)
		})

		t.Run("case=should not cause data loss when updating a request without changes", func(t *testing.T) {
			expected := newFlow(t)
			expected.UI.Nodes = node.Nodes{}
			expected.UI.Nodes.Append(node.NewInputField("zab", nil, node.DefaultGroup, "bar", node.WithInputAttributes(func(a *node.InputAttributes) {
				a.Pattern = "baz"
			})))

			expected.UI.Nodes.Append(node.NewInputField("foo", nil, node.DefaultGroup, "bar", node.WithInputAttributes(func(a *node.InputAttributes) {
				a.Pattern = "baz"
			})))

			err := p.CreateRegistrationFlow(ctx, expected)
			require.NoError(t, err)

			expected.UI.Action = "/new-action"
			expected.UI.Nodes.Append(
				node.NewInputField("zab", nil, node.DefaultGroup, "zab", node.WithInputAttributes(func(a *node.InputAttributes) {
					a.Pattern = "zab"
				})))

			expected.RequestURL = "/new-request-url"
			require.NoError(t, p.UpdateRegistrationFlow(ctx, expected))

			actual, err := p.GetRegistrationFlow(ctx, expected.ID)
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

		t.Run("case=network", func(t *testing.T) {
			id := x.NewUUID()
			nid, p := testhelpers.NewNetwork(t, ctx, p)

			t.Run("sets id on creation", func(t *testing.T) {
				now := time.Now()
				expected := &registration.Flow{ID: id, ExpiresAt: now.Add(time.Hour), IssuedAt: now}
				require.NoError(t, p.CreateRegistrationFlow(ctx, expected))
				assert.EqualValues(t, id, expected.ID)
				assert.EqualValues(t, nid, expected.NID)

				actual, err := p.GetRegistrationFlow(ctx, id)
				require.NoError(t, err)
				assert.EqualValues(t, id, actual.ID)
				assert.EqualValues(t, nid, actual.NID)
			})

			t.Run("can not update id", func(t *testing.T) {
				expected, err := p.GetRegistrationFlow(ctx, id)
				require.NoError(t, err)
				require.NoError(t, p.UpdateRegistrationFlow(ctx, expected))

				actual, err := p.GetRegistrationFlow(ctx, id)
				require.NoError(t, err)
				assert.EqualValues(t, id, actual.ID)
				assert.EqualValues(t, nid, actual.NID)
			})

			t.Run("can not update on another network", func(t *testing.T) {
				expected, err := p.GetRegistrationFlow(ctx, id)
				require.NoError(t, err)

				_, other := testhelpers.NewNetwork(t, ctx, p)

				expected.RequestURL = "updated"
				require.Error(t, other.UpdateRegistrationFlow(ctx, expected))

				_, err = other.GetRegistrationFlow(ctx, id)
				require.ErrorIs(t, err, sqlcon.ErrNoRows)

				actual, err := p.GetRegistrationFlow(ctx, id)
				require.NoError(t, err)
				require.NotEqual(t, "updated", actual.RequestURL)
			})

			t.Run("can not get on another network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, ctx, p)
				_, err := p.GetRegistrationFlow(ctx, id)
				require.ErrorIs(t, err, sqlcon.ErrNoRows)
			})
		})
	}
}
