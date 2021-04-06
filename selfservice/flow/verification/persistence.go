package verification

import (
	"context"
	"testing"

	"github.com/ory/kratos/selfservice/flow"

	"github.com/ory/kratos/ui/container"

	"github.com/bxcodec/faker/v3"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"

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
			require.Len(t, r.UI, 1)
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

			assert.EqualValues(t, expected.ID, actual.ID)
			x.AssertEqualTime(t, expected.IssuedAt, actual.IssuedAt)
			x.AssertEqualTime(t, expected.ExpiresAt, actual.ExpiresAt)
			assert.EqualValues(t, expected.RequestURL, actual.RequestURL)
			require.Equal(t, expected.UI, actual.UI, "expected:\t%s\nactual:\t%s", expected.UI, actual.UI)
		})

		t.Run("case=should create and update a verification request", func(t *testing.T) {
			expected := newFlow(t)
			expected.Type = flow.TypeAPI
			expected.UI = container.New("ory-sh")

			err := p.CreateVerificationFlow(ctx, expected)
			require.NoError(t, err)

			actual, err := p.GetVerificationFlow(ctx, expected.ID)
			require.NoError(t, err)
			assert.Equal(t, flow.TypeAPI, actual.Type)

			actual.UI = container.New("not-ory-sh")
			actual.Type = flow.TypeBrowser

			require.NoError(t, p.UpdateVerificationFlow(ctx, actual))

			actual, err = p.GetVerificationFlow(ctx, actual.ID)
			require.NoError(t, err)
			assert.Equal(t, flow.TypeBrowser, actual.Type)
			assert.Equal(t, "not.ory-sh", actual.UI.Action)
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
			assert.EqualValues(t, expected.UI, actual.UI)
		})
	}
}
