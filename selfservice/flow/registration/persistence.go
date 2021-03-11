package registration

import (
	"context"
	"testing"

	"github.com/ory/x/assertx"

	"github.com/bxcodec/faker/v3"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/x"
)

type FlowPersister interface {
	UpdateRegistrationFlow(context.Context, *Flow) error
	CreateRegistrationFlow(context.Context, *Flow) error
	GetRegistrationFlow(context.Context, uuid.UUID) (*Flow, error)
}

type FlowPersistenceProvider interface {
	RegistrationFlowPersister() FlowPersister
}

func TestFlowPersister(ctx context.Context, p FlowPersister) func(t *testing.T) {
	var clearids = func(r *Flow) {
		r.ID = uuid.UUID{}
	}

	return func(t *testing.T) {
		t.Run("case=should error when the registration flow does not exist", func(t *testing.T) {
			_, err := p.GetRegistrationFlow(ctx, x.NewUUID())
			require.Error(t, err)
		})

		var newFlow = func(t *testing.T) *Flow {
			var r Flow
			require.NoError(t, faker.FakeData(&r))
			clearids(&r)

			nodes := len(r.UI.Nodes)
			assert.NotZero(t, nodes)

			return &r
		}

		t.Run("case=should create a new registration flow and properly set IDs", func(t *testing.T) {
			r := newFlow(t)
			err := p.CreateRegistrationFlow(ctx, r)
			require.NoError(t, err, "%#v", err)

			assert.NotEqual(t, uuid.Nil, r.ID)
		})

		t.Run("case=should create with set ids", func(t *testing.T) {
			var r Flow
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
			require.Equal(t, expected.UI, actual.UI, "expected:\t%s\nactual:\t%s", expected.UI, actual.UI)
		})

		t.Run("case=should not cause data loss when updating a request without changes", func(t *testing.T) {
			expected := newFlow(t)
			expected.Active = ""
			err := p.CreateRegistrationFlow(ctx, expected)
			require.NoError(t, err)

			actual, err := p.GetRegistrationFlow(ctx, expected.ID)
			require.NoError(t, err)
			require.Len(t, actual.UI.Nodes, 2)
			assertx.EqualAsJSON(t,
				expected.UI,
				actual.UI,
			)

			require.NoError(t, p.UpdateRegistrationFlow(ctx, actual))

			actual, err = p.GetRegistrationFlow(ctx, expected.ID)
			require.NoError(t, err)
			require.Len(t, actual.UI.Nodes, 2)
			assertx.EqualAsJSON(t,
				expected.UI,
				actual.UI,
			)
		})
	}
}
