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
	RequestPersister interface {
		CreateRecoveryRequest(context.Context, *Flow) error
		GetRecoveryRequest(ctx context.Context, id uuid.UUID) (*Flow, error)
		UpdateRecoveryRequest(context.Context, *Flow) error
	}
	RequestPersistenceProvider interface {
		RecoveryRequestPersister() RequestPersister
	}
)

func TestRequestPersister(p interface {
	RequestPersister
	identity.PrivilegedPool
}) func(t *testing.T) {
	viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/identity.schema.json")

	var clearids = func(r *Flow) {
		r.ID = uuid.UUID{}
	}

	return func(t *testing.T) {
		t.Run("case=should error when the recovery request does not exist", func(t *testing.T) {
			_, err := p.GetRecoveryRequest(context.Background(), x.NewUUID())
			require.Error(t, err)
		})

		var newRequest = func(t *testing.T) *Flow {
			var r Flow
			require.NoError(t, faker.FakeData(&r))
			clearids(&r)
			return &r
		}

		t.Run("case=should create a new recovery request", func(t *testing.T) {
			r := newRequest(t)
			err := p.CreateRecoveryRequest(context.Background(), r)
			require.NoError(t, err, "%#v", err)
		})

		t.Run("case=should create with set ids", func(t *testing.T) {
			var r Flow
			require.NoError(t, faker.FakeData(&r))
			require.NoError(t, p.CreateRecoveryRequest(context.Background(), &r))
		})

		t.Run("case=should create and fetch a recovery request", func(t *testing.T) {
			expected := newRequest(t)
			err := p.CreateRecoveryRequest(context.Background(), expected)
			require.NoError(t, err)

			actual, err := p.GetRecoveryRequest(context.Background(), expected.ID)
			require.NoError(t, err)

			fexpected, _ := json.Marshal(expected.Methods[StrategyRecoveryTokenName].Config)
			factual, _ := json.Marshal(actual.Methods[StrategyRecoveryTokenName].Config)

			require.NotEmpty(t, actual.Methods[StrategyRecoveryTokenName].Config.RequestMethodConfigurator.(*form.HTMLForm).Action)
			assert.EqualValues(t, expected.ID, actual.ID)
			assert.JSONEq(t, string(fexpected), string(factual))
			x.AssertEqualTime(t, expected.IssuedAt, actual.IssuedAt)
			x.AssertEqualTime(t, expected.ExpiresAt, actual.ExpiresAt)
			assert.EqualValues(t, expected.RequestURL, actual.RequestURL)
		})

		t.Run("case=should create and update a recovery request", func(t *testing.T) {
			expected := newRequest(t)
			expected.Methods[StrategyRecoveryTokenName] = &FlowMethod{
				Method: StrategyRecoveryTokenName, Config: &RequestMethodConfig{RequestMethodConfigurator: &form.HTMLForm{Fields: []form.Field{{
					Name: "zab", Type: "bar", Pattern: "baz"}}}}}
			expected.Methods["password"] = &FlowMethod{
				Method: "password", Config: &RequestMethodConfig{RequestMethodConfigurator: &form.HTMLForm{Fields: []form.Field{{
					Name: "foo", Type: "bar", Pattern: "baz"}}}}}
			err := p.CreateRecoveryRequest(context.Background(), expected)
			require.NoError(t, err)

			expected.Methods[StrategyRecoveryTokenName].Config.RequestMethodConfigurator.(*form.HTMLForm).Action = "/new-action"
			expected.Methods["password"].Config.RequestMethodConfigurator.(*form.HTMLForm).Fields = []form.Field{{
				Name: "zab", Type: "zab", Pattern: "zab"}}
			expected.RequestURL = "/new-request-url"
			expected.Active = StrategyRecoveryTokenName
			expected.Messages.Add(text.NewRecoveryEmailSent())
			require.NoError(t, p.UpdateRecoveryRequest(context.Background(), expected))

			actual, err := p.GetRecoveryRequest(context.Background(), expected.ID)
			require.NoError(t, err)

			assert.Equal(t, "/new-action", actual.Methods[StrategyRecoveryTokenName].Config.RequestMethodConfigurator.(*form.HTMLForm).Action)
			assert.Equal(t, "/new-request-url", actual.RequestURL)
			assert.Equal(t, StrategyRecoveryTokenName, actual.Active.String())
			assert.Equal(t, expected.Messages, actual.Messages)
			assert.EqualValues(t, []form.Field{{Name: "zab", Type: "zab", Pattern: "zab"}}, actual.
				Methods["password"].Config.RequestMethodConfigurator.(*form.HTMLForm).Fields)
			assert.EqualValues(t, []form.Field{{Name: "zab", Type: "bar", Pattern: "baz"}}, actual.
				Methods[StrategyRecoveryTokenName].Config.RequestMethodConfigurator.(*form.HTMLForm).Fields)
		})

		t.Run("case=should not cause data loss when updating a request without changes", func(t *testing.T) {
			t.Logf("Needs implementation")
			t.FailNow()
			// expected := newFlow(t)
			// delete(expected.Methods, identity.CredentialsTypeOIDC)
			// err := p.CreateLoginFlow(context.Background(), expected)
			// require.NoError(t, err)
			//
			// actual, err := p.GetLoginFlow(context.Background(), expected.ID)
			// require.NoError(t, err)
			// assert.Len(t, actual.Methods, 1)
			//
			// require.NoError(t, p.UpdateLoginFlow(context.Background(), actual))
			//
			// actual, err = p.GetLoginFlow(context.Background(), expected.ID)
			// require.NoError(t, err)
			// require.Len(t, actual.Methods, 2)
			// assert.EqualValues(t, identity.CredentialsTypePassword, actual.Active)
			//
			// js, _ := json.Marshal(actual.Methods)
			// assert.Equal(t, string(identity.CredentialsTypePassword), actual.Methods[identity.CredentialsTypePassword].Config.FlowMethodConfigurator.(*form.HTMLForm).Action, "%s", js)
			// assert.Equal(t, string(identity.CredentialsTypeOIDC), actual.Methods[identity.CredentialsTypeOIDC].Config.FlowMethodConfigurator.(*form.HTMLForm).Action)
		})
	}
}
