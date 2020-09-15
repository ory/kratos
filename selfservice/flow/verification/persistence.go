package verification

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
	FlowPersistenceProvider interface {
		VerificationFlowPersister() FlowPersister
	}
	FlowPersister interface {
		CreateVerificationFlow(context.Context, *Flow) error
		GetVerificationFlow(ctx context.Context, id uuid.UUID) (*Flow, error)
		UpdateVerificationFlow(context.Context, *Flow) error
	}
)

func TestFlowPersister(p interface {
	FlowPersister
	identity.PrivilegedPool
}) func(t *testing.T) {
	viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/identity.schema.json")

	var clearids = func(r *Flow) {
		r.ID = uuid.UUID{}
	}

	return func(t *testing.T) {
		t.Run("case=should error when the verification request does not exist", func(t *testing.T) {
			_, err := p.GetVerificationFlow(context.Background(), x.NewUUID())
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
			err := p.CreateVerificationFlow(context.Background(), r)
			require.NoError(t, err, "%#v", err)
		})

		t.Run("case=should create with set ids", func(t *testing.T) {
			var r Flow
			require.NoError(t, faker.FakeData(&r))
			require.NoError(t, p.CreateVerificationFlow(context.Background(), &r))
		})

		t.Run("case=should create and fetch a verification request", func(t *testing.T) {
			expected := newFlow(t)
			err := p.CreateVerificationFlow(context.Background(), expected)
			require.NoError(t, err)

			actual, err := p.GetVerificationFlow(context.Background(), expected.ID)
			require.NoError(t, err)

			fexpected, _ := json.Marshal(expected.Methods[StrategyVerificationLinkName].Config)
			factual, _ := json.Marshal(actual.Methods[StrategyVerificationLinkName].Config)

			require.NotEmpty(t, actual.Methods[StrategyVerificationLinkName].Config.FlowMethodConfigurator.(*form.HTMLForm).Action)
			assert.EqualValues(t, expected.ID, actual.ID)
			assert.JSONEq(t, string(fexpected), string(factual))
			x.AssertEqualTime(t, expected.IssuedAt, actual.IssuedAt)
			x.AssertEqualTime(t, expected.ExpiresAt, actual.ExpiresAt)
			assert.EqualValues(t, expected.RequestURL, actual.RequestURL)
		})

		t.Run("case=should create and update a verification request", func(t *testing.T) {
			expected := newFlow(t)
			expected.Methods[StrategyVerificationLinkName] = &FlowMethod{
				Method: StrategyVerificationLinkName, Config: &FlowMethodConfig{FlowMethodConfigurator: &form.HTMLForm{Fields: []form.Field{{
					Name: "zab", Type: "bar", Pattern: "baz"}}}}}
			expected.Methods["password"] = &FlowMethod{
				Method: "password", Config: &FlowMethodConfig{FlowMethodConfigurator: &form.HTMLForm{Fields: []form.Field{{
					Name: "foo", Type: "bar", Pattern: "baz"}}}}}
			err := p.CreateVerificationFlow(context.Background(), expected)
			require.NoError(t, err)

			expected.Methods[StrategyVerificationLinkName].Config.FlowMethodConfigurator.(*form.HTMLForm).Action = "/new-action"
			expected.Methods["password"].Config.FlowMethodConfigurator.(*form.HTMLForm).Fields = []form.Field{{
				Name: "zab", Type: "zab", Pattern: "zab"}}
			expected.RequestURL = "/new-request-url"
			expected.Active = StrategyVerificationLinkName
			expected.Messages.Add(text.NewVerificationEmailSent())
			require.NoError(t, p.UpdateVerificationFlow(context.Background(), expected))

			actual, err := p.GetVerificationFlow(context.Background(), expected.ID)
			require.NoError(t, err)

			assert.Equal(t, "/new-action", actual.Methods[StrategyVerificationLinkName].Config.FlowMethodConfigurator.(*form.HTMLForm).Action)
			assert.Equal(t, "/new-request-url", actual.RequestURL)
			assert.Equal(t, StrategyVerificationLinkName, actual.Active.String())
			assert.Equal(t, expected.Messages, actual.Messages)
			assert.EqualValues(t, []form.Field{{Name: "zab", Type: "zab", Pattern: "zab"}}, actual.
				Methods["password"].Config.FlowMethodConfigurator.(*form.HTMLForm).Fields)
			assert.EqualValues(t, []form.Field{{Name: "zab", Type: "bar", Pattern: "baz"}}, actual.
				Methods[StrategyVerificationLinkName].Config.FlowMethodConfigurator.(*form.HTMLForm).Fields)
		})

		t.Run("case=should not cause data loss when updating a request without changes", func(t *testing.T) {
			expected := newFlow(t)
			err := p.CreateVerificationFlow(context.Background(), expected)
			require.NoError(t, err)

			actual, err := p.GetVerificationFlow(context.Background(), expected.ID)
			require.NoError(t, err)
			assert.Len(t, actual.Methods, 1)

			require.NoError(t, p.UpdateVerificationFlow(context.Background(), actual))

			actual, err = p.GetVerificationFlow(context.Background(), expected.ID)
			require.NoError(t, err)
			require.Len(t, actual.Methods, 1)

			js, _ := json.Marshal(actual.Methods)
			assert.Equal(t, expected.Methods[StrategyVerificationLinkName].Config.FlowMethodConfigurator.(*form.HTMLForm).Action,
				actual.Methods[StrategyVerificationLinkName].Config.FlowMethodConfigurator.(*form.HTMLForm).Action, "%s", js)
		})
	}
}
