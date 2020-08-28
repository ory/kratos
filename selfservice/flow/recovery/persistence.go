package recovery

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

type (
	FlowPersister interface {
		CreateRecoveryFlow(context.Context, *Flow) error
		GetRecoveryFlow(ctx context.Context, id uuid.UUID) (*Flow, error)
		UpdateRecoveryFlow(context.Context, *Flow) error
	}
	RequestPersistenceProvider interface {
		RecoveryFlowPersister() FlowPersister
	}
)

func TestRequestPersister(p interface {
	FlowPersister
	identity.PrivilegedPool
}) func(t *testing.T) {
	viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/identity.schema.json")

	var clearids = func(r *Flow) {
		r.ID = uuid.UUID{}
	}

	return func(t *testing.T) {
		t.Run("case=should error when the recovery request does not exist", func(t *testing.T) {
			_, err := p.GetRecoveryFlow(context.Background(), x.NewUUID())
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
			err := p.CreateRecoveryFlow(context.Background(), r)
			require.NoError(t, err, "%#v", err)
		})

		t.Run("case=should create with set ids", func(t *testing.T) {
			var r Flow
			require.NoError(t, faker.FakeData(&r))
			require.NoError(t, p.CreateRecoveryFlow(context.Background(), &r))
		})

		t.Run("case=should create and fetch a recovery request", func(t *testing.T) {
			expected := newFlow(t)
			err := p.CreateRecoveryFlow(context.Background(), expected)
			require.NoError(t, err)

			actual, err := p.GetRecoveryFlow(context.Background(), expected.ID)
			require.NoError(t, err)

			fexpected, _ := json.Marshal(expected.Methods[StrategyRecoveryTokenName].Config)
			factual, _ := json.Marshal(actual.Methods[StrategyRecoveryTokenName].Config)

			require.NotEmpty(t, actual.Methods[StrategyRecoveryTokenName].Config.FlowMethodConfigurator.(*form.HTMLForm).Action)
			assert.EqualValues(t, expected.ID, actual.ID)
			assert.JSONEq(t, string(fexpected), string(factual))
			x.AssertEqualTime(t, expected.IssuedAt, actual.IssuedAt)
			x.AssertEqualTime(t, expected.ExpiresAt, actual.ExpiresAt)
			assert.EqualValues(t, expected.RequestURL, actual.RequestURL)
		})

		t.Run("case=should create and update a recovery request", func(t *testing.T) {
			expected := newFlow(t)
			expected.Methods[StrategyRecoveryTokenName] = &FlowMethod{
				Method: StrategyRecoveryTokenName, Config: &FlowMethodConfig{FlowMethodConfigurator: &form.HTMLForm{Fields: []form.Field{{
					Name: "zab", Type: "bar", Pattern: "baz"}}}}}
			expected.Methods["password"] = &FlowMethod{
				Method: "password", Config: &FlowMethodConfig{FlowMethodConfigurator: &form.HTMLForm{Fields: []form.Field{{
					Name: "foo", Type: "bar", Pattern: "baz"}}}}}
			err := p.CreateRecoveryFlow(context.Background(), expected)
			require.NoError(t, err)

			expected.Methods[StrategyRecoveryTokenName].Config.FlowMethodConfigurator.(*form.HTMLForm).Action = "/new-action"
			expected.Methods["password"].Config.FlowMethodConfigurator.(*form.HTMLForm).Fields = []form.Field{{
				Name: "zab", Type: "zab", Pattern: "zab"}}
			expected.RequestURL = "/new-request-url"
			expected.Active = StrategyRecoveryTokenName
			expected.Messages.Add(text.NewRecoveryEmailSent())
			require.NoError(t, p.UpdateRecoveryFlow(context.Background(), expected))

			actual, err := p.GetRecoveryFlow(context.Background(), expected.ID)
			require.NoError(t, err)

			assert.Equal(t, "/new-action", actual.Methods[StrategyRecoveryTokenName].Config.FlowMethodConfigurator.(*form.HTMLForm).Action)
			assert.Equal(t, "/new-request-url", actual.RequestURL)
			assert.Equal(t, StrategyRecoveryTokenName, actual.Active.String())
			assert.Equal(t, expected.Messages, actual.Messages)
			assert.EqualValues(t, []form.Field{{Name: "zab", Type: "zab", Pattern: "zab"}}, actual.
				Methods["password"].Config.FlowMethodConfigurator.(*form.HTMLForm).Fields)
			assert.EqualValues(t, []form.Field{{Name: "zab", Type: "bar", Pattern: "baz"}}, actual.
				Methods[StrategyRecoveryTokenName].Config.FlowMethodConfigurator.(*form.HTMLForm).Fields)
		})

		t.Run("case=should not cause data loss when updating a request without changes", func(t *testing.T) {
			expected := newFlow(t)
			err := p.CreateRecoveryFlow(context.Background(), expected)
			require.NoError(t, err)

			actual, err := p.GetRecoveryFlow(context.Background(), expected.ID)
			require.NoError(t, err)
			assert.Len(t, actual.Methods, 1)

			require.NoError(t, p.UpdateRecoveryFlow(context.Background(), actual))

			actual, err = p.GetRecoveryFlow(context.Background(), expected.ID)
			require.NoError(t, err)
			require.Len(t, actual.Methods, 1)

			js, _ := json.Marshal(actual.Methods)
			assert.Equal(t, expected.Methods[StrategyRecoveryTokenName].Config.FlowMethodConfigurator.(*form.HTMLForm).Action,
				actual.Methods[StrategyRecoveryTokenName].Config.FlowMethodConfigurator.(*form.HTMLForm).Action, "%s", js)
		})
	}
}
