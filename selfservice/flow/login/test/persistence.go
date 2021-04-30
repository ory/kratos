package test

import (
	"context"
	"testing"
	"time"

	"github.com/ory/kratos/internal/testhelpers"

	"github.com/bxcodec/faker/v3"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/persistence"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/x"
	"github.com/ory/x/assertx"
	"github.com/ory/x/sqlcon"
)

func TestFlowPersister(ctx context.Context, p persistence.Persister) func(t *testing.T) {
	var clearids = func(r *login.Flow) {
		r.ID = uuid.UUID{}
	}

	return func(t *testing.T) {
		_, p := testhelpers.NewNetworkUnlessExisting(t, ctx, p)

		t.Run("case=should error when the login flow does not exist", func(t *testing.T) {
			_, err := p.GetLoginFlow(ctx, x.NewUUID())
			require.Error(t, err)
		})

		var newFlow = func(t *testing.T) *login.Flow {
			var r login.Flow
			require.NoError(t, faker.FakeData(&r))
			clearids(&r)
			return &r
		}

		t.Run("case=should create with set ids", func(t *testing.T) {
			var r login.Flow
			require.NoError(t, faker.FakeData(&r))
			require.NoError(t, p.CreateLoginFlow(ctx, &r))
		})

		t.Run("case=should create a new login flow and properly set IDs", func(t *testing.T) {
			r := newFlow(t)
			err := p.CreateLoginFlow(ctx, r)
			require.NoError(t, err, "%#v", err)
			assert.NotEqual(t, uuid.Nil, r.ID)
		})

		t.Run("case=should create and fetch a login flow", func(t *testing.T) {
			expected := newFlow(t)
			err := p.CreateLoginFlow(ctx, expected)
			require.NoError(t, err)

			actual, err := p.GetLoginFlow(ctx, expected.ID)
			require.NoError(t, err)

			assert.EqualValues(t, expected.ID, actual.ID)
			x.AssertEqualTime(t, expected.IssuedAt, actual.IssuedAt)
			x.AssertEqualTime(t, expected.ExpiresAt, actual.ExpiresAt)
			assert.EqualValues(t, expected.RequestURL, actual.RequestURL)
			assert.EqualValues(t, expected.Active, actual.Active)
			assertx.EqualAsJSON(t, expected.UI, actual.UI, "expected:\t%s\nactual:\t%s", expected.UI, actual.UI)
		})

		t.Run("case=should properly set the flow type", func(t *testing.T) {
			expected := newFlow(t)
			expected.Forced = true
			expected.Type = flow.TypeAPI
			expected.UI = container.New("ory-sh")

			err := p.CreateLoginFlow(ctx, expected)
			require.NoError(t, err)

			actual, err := p.GetLoginFlow(ctx, expected.ID)
			require.NoError(t, err)
			assert.Equal(t, flow.TypeAPI, actual.Type)

			actual.UI = container.New("not-ory-sh")
			actual.Type = flow.TypeBrowser
			actual.Forced = true

			require.NoError(t, p.UpdateLoginFlow(ctx, actual))

			actual, err = p.GetLoginFlow(ctx, actual.ID)
			require.NoError(t, err)
			assert.Equal(t, flow.TypeBrowser, actual.Type)
			assert.True(t, actual.Forced)
			assert.Equal(t, "not-ory-sh", actual.UI.Action)
		})

		t.Run("case=should not cause data loss when updating a request without changes", func(t *testing.T) {
			expected := newFlow(t)
			err := p.CreateLoginFlow(ctx, expected)
			require.NoError(t, err)

			actual, err := p.GetLoginFlow(ctx, expected.ID)
			require.NoError(t, err)

			require.NoError(t, p.UpdateLoginFlow(ctx, actual))

			actual, err = p.GetLoginFlow(ctx, expected.ID)
			require.NoError(t, err)
			assertx.EqualAsJSON(t, expected.UI, actual.UI)
		})

		t.Run("case=network", func(t *testing.T) {
			id := x.NewUUID()
			nid, p := testhelpers.NewNetwork(t, ctx, p)

			t.Run("sets id on creation", func(t *testing.T) {
				expected := &login.Flow{ID: id, IssuedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)}
				require.NoError(t, p.CreateLoginFlow(ctx, expected))
				assert.EqualValues(t, id, expected.ID)
				assert.EqualValues(t, nid, expected.NID)

				actual, err := p.GetLoginFlow(ctx, id)
				require.NoError(t, err)
				assert.EqualValues(t, id, actual.ID)
				assert.EqualValues(t, nid, actual.NID)
			})

			t.Run("can not update id", func(t *testing.T) {
				expected, err := p.GetLoginFlow(ctx, id)
				require.NoError(t, err)
				require.NoError(t, p.UpdateLoginFlow(ctx, expected))

				actual, err := p.GetLoginFlow(ctx, id)
				require.NoError(t, err)
				assert.EqualValues(t, id, actual.ID)
				assert.EqualValues(t, nid, actual.NID)
			})

			t.Run("can not force on another network", func(t *testing.T) {
				expected, err := p.GetLoginFlow(ctx, id)
				require.NoError(t, err)

				_, other := testhelpers.NewNetwork(t, ctx, p)

				expected.RequestURL = "updated"
				require.Error(t, other.ForceLoginFlow(ctx, expected.ID))

				_, err = other.GetLoginFlow(ctx, id)
				require.ErrorIs(t, err, sqlcon.ErrNoRows)

				actual, err := p.GetLoginFlow(ctx, id)
				require.NoError(t, err)
				require.False(t, actual.Forced)
			})

			t.Run("can not update on another network", func(t *testing.T) {
				expected, err := p.GetLoginFlow(ctx, id)
				require.NoError(t, err)

				_, other := testhelpers.NewNetwork(t, ctx, p)

				expected.RequestURL = "updated"
				require.Error(t, other.UpdateLoginFlow(ctx, expected))

				_, err = other.GetLoginFlow(ctx, id)
				require.ErrorIs(t, err, sqlcon.ErrNoRows)

				actual, err := p.GetLoginFlow(ctx, id)
				require.NoError(t, err)
				require.NotEqual(t, "updated", actual.RequestURL)
			})

			t.Run("can not get on another network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, ctx, p)
				_, err := p.GetLoginFlow(ctx, id)
				require.ErrorIs(t, err, sqlcon.ErrNoRows)
			})
		})
	}
}
