package recovery

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
	FlowPersister interface {
		CreateRecoveryFlow(context.Context, *Flow) error
		GetRecoveryFlow(ctx context.Context, id uuid.UUID) (*Flow, error)
		UpdateRecoveryFlow(context.Context, *Flow) error
	}
	FlowPersistenceProvider interface {
		RecoveryFlowPersister() FlowPersister
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

			assert.EqualValues(t, expected.ID, actual.ID)
			x.AssertEqualTime(t, expected.IssuedAt, actual.IssuedAt)
			x.AssertEqualTime(t, expected.ExpiresAt, actual.ExpiresAt)
			assert.EqualValues(t, expected.RequestURL, actual.RequestURL)
			require.Equal(t, expected.UI, actual.UI, "expected:\t%s\nactual:\t%s", expected.UI, actual.UI)
		})

		t.Run("case=should create and update a recovery request", func(t *testing.T) {

			expected := newFlow(t)
			expected.Type = flow.TypeAPI
			expected.UI = container.New("ory-sh")

			err := p.CreateRecoveryFlow(ctx, expected)
			require.NoError(t, err)

			actual, err := p.GetRecoveryFlow(ctx, expected.ID)
			require.NoError(t, err)
			assert.Equal(t, flow.TypeAPI, actual.Type)

			actual.UI = container.New("not-ory-sh")
			actual.Type = flow.TypeBrowser

			require.NoError(t, p.UpdateRecoveryFlow(ctx, actual))

			actual, err = p.GetRecoveryFlow(ctx, actual.ID)
			require.NoError(t, err)
			assert.Equal(t, flow.TypeBrowser, actual.Type)
			assert.Equal(t, "not.ory-sh", actual.UI.Action)
		})

		t.Run("case=should not cause data loss when updating a request without changes", func(t *testing.T) {
			expected := newFlow(t)
			err := p.CreateRecoveryFlow(ctx, expected)
			require.NoError(t, err)

			actual, err := p.GetRecoveryFlow(ctx, expected.ID)
			require.NoError(t, err)

			require.NoError(t, p.UpdateRecoveryFlow(ctx, actual))

			actual, err = p.GetRecoveryFlow(ctx, expected.ID)
			require.NoError(t, err)
			assert.EqualValues(t, expected.UI, actual.UI)
		})
	}
}
