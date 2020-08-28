package registration

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/gobuffalo/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/x"
)

type FlowPersister interface {
	UpdateRegistrationFlow(context.Context, *Flow) error
	CreateRegistrationFlow(context.Context, *Flow) error
	GetRegistrationFlow(context.Context, uuid.UUID) (*Flow, error)
	UpdateRegistrationFlowMethod(context.Context, uuid.UUID, identity.CredentialsType, *FlowMethod) error
}

type FlowPersistenceProvider interface {
	RegistrationFlowPersister() FlowPersister
}

func TestFlowPersister(p FlowPersister) func(t *testing.T) {
	var clearids = func(r *Flow) {
		r.ID = uuid.UUID{}
		for k := range r.Methods {
			r.Methods[k].ID = uuid.UUID{}
		}
	}

	return func(t *testing.T) {
		t.Run("case=should error when the registration flow does not exist", func(t *testing.T) {
			_, err := p.GetRegistrationFlow(context.Background(), x.NewUUID())
			require.Error(t, err)
		})

		var newFlow = func(t *testing.T) *Flow {
			var r Flow
			require.NoError(t, faker.FakeData(&r))
			clearids(&r)

			methods := len(r.Methods)
			assert.NotZero(t, methods)

			return &r
		}

		t.Run("case=should create a new registration flow and properly set IDs", func(t *testing.T) {
			r := newFlow(t)
			methods := len(r.Methods)
			err := p.CreateRegistrationFlow(context.Background(), r)
			require.NoError(t, err, "%#v", err)

			assert.Nil(t, r.MethodsRaw)
			assert.NotEqual(t, uuid.Nil, r.ID)
			for _, m := range r.Methods {
				assert.NotEqual(t, uuid.Nil, m.ID)
			}
			assert.Len(t, r.Methods, methods)
		})

		t.Run("case=should create with set ids", func(t *testing.T) {
			var r Flow
			require.NoError(t, faker.FakeData(&r))
			require.NoError(t, p.CreateRegistrationFlow(context.Background(), &r))
		})

		t.Run("case=should create and fetch a registration flow", func(t *testing.T) {
			expected := newFlow(t)
			err := p.CreateRegistrationFlow(context.Background(), expected)
			require.NoError(t, err)

			actual, err := p.GetRegistrationFlow(context.Background(), expected.ID)
			require.NoError(t, err)
			assert.Empty(t, actual.MethodsRaw)

			assert.EqualValues(t, expected.ID, actual.ID)
			x.AssertEqualTime(t, expected.IssuedAt, actual.IssuedAt)
			x.AssertEqualTime(t, expected.ExpiresAt, actual.ExpiresAt)
			assert.EqualValues(t, expected.RequestURL, actual.RequestURL)
			assert.EqualValues(t, expected.Active, actual.Active)
			require.Equal(t, len(expected.Methods), len(actual.Methods), "expected:\t%s\nactual:\t%s", expected.Methods, actual.Methods)
		})

		t.Run("case=should update a registration flow", func(t *testing.T) {
			expected := newFlow(t)
			delete(expected.Methods, identity.CredentialsTypeOIDC)
			err := p.CreateRegistrationFlow(context.Background(), expected)
			require.NoError(t, err)

			actual, err := p.GetRegistrationFlow(context.Background(), expected.ID)
			require.NoError(t, err)
			assert.Len(t, actual.Methods, 1)

			require.NoError(t, p.UpdateRegistrationFlowMethod(context.Background(), expected.ID, identity.CredentialsTypeOIDC, &FlowMethod{
				Method: identity.CredentialsTypeOIDC,
				Config: &FlowMethodConfig{FlowMethodConfigurator: form.NewHTMLForm(string(identity.CredentialsTypeOIDC))},
			}))

			require.NoError(t, p.UpdateRegistrationFlowMethod(context.Background(), expected.ID, identity.CredentialsTypePassword, &FlowMethod{
				Method: identity.CredentialsTypePassword,
				Config: &FlowMethodConfig{FlowMethodConfigurator: form.NewHTMLForm(string(identity.CredentialsTypePassword))},
			}))

			actual, err = p.GetRegistrationFlow(context.Background(), expected.ID)
			require.NoError(t, err)
			require.Len(t, actual.Methods, 2)
			assert.EqualValues(t, identity.CredentialsTypePassword, actual.Active)

			js, _ := json.Marshal(actual.Methods)
			assert.Equal(t, string(identity.CredentialsTypePassword), actual.Methods[identity.CredentialsTypePassword].Config.FlowMethodConfigurator.(*form.HTMLForm).Action, "%s", js)
			assert.Equal(t, string(identity.CredentialsTypeOIDC), actual.Methods[identity.CredentialsTypeOIDC].Config.FlowMethodConfigurator.(*form.HTMLForm).Action)
		})

		t.Run("case=should not cause data loss when updating a request without changes", func(t *testing.T) {
			expected := newFlow(t)
			expected.Active = ""
			err := p.CreateRegistrationFlow(context.Background(), expected)
			require.NoError(t, err)

			actual, err := p.GetRegistrationFlow(context.Background(), expected.ID)
			require.NoError(t, err)
			assert.Len(t, actual.Methods, 2)

			require.NoError(t, p.UpdateRegistrationFlow(context.Background(), actual))

			actual, err = p.GetRegistrationFlow(context.Background(), expected.ID)
			require.NoError(t, err)
			require.Len(t, actual.Methods, 2)

			assert.Equal(t,
				expected.Methods[identity.CredentialsTypePassword].Config.FlowMethodConfigurator.(*form.HTMLForm).Action,
				actual.Methods[identity.CredentialsTypePassword].Config.FlowMethodConfigurator.(*form.HTMLForm).Action,
			)
			assert.Equal(t,
				expected.Methods[identity.CredentialsTypeOIDC].Config.FlowMethodConfigurator.(*form.HTMLForm).Action,
				actual.Methods[identity.CredentialsTypeOIDC].Config.FlowMethodConfigurator.(*form.HTMLForm).Action,
			)
		})
	}
}
