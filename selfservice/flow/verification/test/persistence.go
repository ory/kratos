package test

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/persistence"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/assertx"
	"github.com/ory/x/sqlcon"
)

func TestFlowPersister(ctx context.Context, conf *config.Config, p interface {
	persistence.Persister
}) func(t *testing.T) {
	var clearids = func(r *verification.Flow) {
		r.ID = uuid.UUID{}
	}

	return func(t *testing.T) {
		nid, p := testhelpers.NewNetwork(t, p)

		conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/identity.schema.json")

		t.Run("case=should error when the verification request does not exist", func(t *testing.T) {
			_, err := p.GetVerificationFlow(ctx, x.NewUUID())
			require.Error(t, err)
		})

		var newFlow = func(t *testing.T) *verification.Flow {
			var r verification.Flow
			require.NoError(t, faker.FakeData(&r))
			clearids(&r)
			return &r
		}

		t.Run("case=should create a new verification flow", func(t *testing.T) {
			r := newFlow(t)
			err := p.CreateVerificationFlow(ctx, r)
			require.NoError(t, err, "%#v", err)
			require.Equal(t, nid, r.NID)

			t.Run("fail to find on other network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, p)
				_, err := p.GetVerificationFlow(ctx, r.ID)
				require.ErrorIs(t, err, sqlcon.ErrNoRows)
			})
		})

		t.Run("case=should create with set ids", func(t *testing.T) {
			var r verification.Flow
			require.NoError(t, faker.FakeData(&r))
			require.NoError(t, p.CreateVerificationFlow(ctx, &r))
			require.Equal(t, nid, r.NID)

			t.Run("fail to find on other network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, p)
				_, err := p.GetVerificationFlow(ctx, r.ID)
				require.ErrorIs(t, err, sqlcon.ErrNoRows)
			})
		})

		t.Run("case=should create and fetch a verification request", func(t *testing.T) {
			expected := newFlow(t)
			require.NoError(t, p.CreateVerificationFlow(ctx, expected))
			require.Equal(t, nid, expected.NID)

			t.Run("fail to find on other network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, p)
				_, err := p.GetVerificationFlow(ctx, expected.ID)
				require.ErrorIs(t, err, sqlcon.ErrNoRows)
			})

			actual, err := p.GetVerificationFlow(ctx, expected.ID)
			require.NoError(t, err)

			assert.EqualValues(t, expected.ID, actual.ID)
			x.AssertEqualTime(t, expected.IssuedAt, actual.IssuedAt)
			x.AssertEqualTime(t, expected.ExpiresAt, actual.ExpiresAt)
			assert.EqualValues(t, expected.RequestURL, actual.RequestURL)
			assertx.EqualAsJSON(t, expected.UI, actual.UI, "expected:\t%s\nactual:\t%s", expected.UI, actual.UI)
		})

		t.Run("case=should create and update a verification request", func(t *testing.T) {
			expected := newFlow(t)
			expected.UI.Nodes = node.Nodes{}
			expected.UI.Nodes.Append(node.NewInputField("zab", nil, node.DefaultGroup, "bar", node.WithInputAttributes(func(a *node.InputAttributes) {
				a.Pattern = "baz"
			})))

			expected.UI.Nodes.Append(node.NewInputField("foo", nil, node.DefaultGroup, "bar", node.WithInputAttributes(func(a *node.InputAttributes) {
				a.Pattern = "baz"
			})))

			err := p.CreateVerificationFlow(ctx, expected)
			require.NoError(t, err)

			expected.UI.Action = "/new-action"
			expected.UI.Nodes.Append(
				node.NewInputField("zab", nil, node.DefaultGroup, "zab", node.WithInputAttributes(func(a *node.InputAttributes) {
					a.Pattern = "zab"
				})))

			expected.RequestURL = "/new-request-url"

			t.Run("fail to find on other network", func(t *testing.T) {
				_, other := testhelpers.NewNetwork(t, p)
				require.ErrorIs(t, other.UpdateVerificationFlow(ctx, expected), sqlcon.ErrNoRows)

				actual, err := p.GetVerificationFlow(ctx, expected.ID)
				require.NoError(t, err)
				assert.NotEqual(t, "/new-request-url", actual.RequestURL)
			})

			require.NoError(t, p.UpdateVerificationFlow(ctx, expected))

			actual, err := p.GetVerificationFlow(ctx, expected.ID)
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

		t.Run("case=should not cause data loss when updating a request without changes", func(t *testing.T) {
			expected := newFlow(t)
			err := p.CreateVerificationFlow(ctx, expected)
			require.NoError(t, err)

			actual, err := p.GetVerificationFlow(ctx, expected.ID)
			require.NoError(t, err)

			require.NoError(t, p.UpdateVerificationFlow(ctx, actual))

			actual, err = p.GetVerificationFlow(ctx, expected.ID)
			require.NoError(t, err)
			assertx.EqualAsJSON(t, expected.UI, actual.UI)
		})
	}
}
