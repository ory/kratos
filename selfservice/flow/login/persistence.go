package login

import (
	"context"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/x"
)

type (
	FlowPersister interface {
		UpdateLoginFlow(context.Context, *Flow) error
		CreateLoginFlow(context.Context, *Flow) error
		GetLoginFlow(context.Context, uuid.UUID) (*Flow, error)
		UpdateLoginFlowMethod(context.Context, uuid.UUID, identity.CredentialsType, *FlowMethod) error
		ForceLoginFlow(ctx context.Context, id uuid.UUID) error
	}
	FlowPersistenceProvider interface {
		LoginFlowPersister() FlowPersister
	}
)

func TestFlowPersister(p FlowPersister) func(t *testing.T) {
	var clearids = func(r *Flow) {
		r.ID = uuid.UUID{}
		for k := range r.Methods {
			r.Methods[k].ID = uuid.UUID{}
		}
	}

	return func(t *testing.T) {
		t.Run("case=should error when the login flow does not exist", func(t *testing.T) {
			_, err := p.GetLoginFlow(context.Background(), x.NewUUID())
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

		t.Run("case=should create with set ids", func(t *testing.T) {
			var r Flow
			require.NoError(t, faker.FakeData(&r))
			require.NoError(t, p.CreateLoginFlow(context.Background(), &r))
		})

		t.Run("case=should create a new login flow and properly set IDs", func(t *testing.T) {
			r := newFlow(t)
			methods := len(r.Methods)
			err := p.CreateLoginFlow(context.Background(), r)
			require.NoError(t, err, "%#v", err)

			assert.Nil(t, r.MethodsRaw)
			assert.NotEqual(t, uuid.Nil, r.ID)
			for _, m := range r.Methods {
				assert.NotEqual(t, uuid.Nil, m.ID)
			}
			assert.Len(t, r.Methods, methods)
		})

		t.Run("case=should create and fetch a login flow", func(t *testing.T) {
			expected := newFlow(t)
			err := p.CreateLoginFlow(context.Background(), expected)
			require.NoError(t, err)

			actual, err := p.GetLoginFlow(context.Background(), expected.ID)
			require.NoError(t, err)
			assert.Empty(t, actual.MethodsRaw)

			assert.EqualValues(t, expected.ID, actual.ID)
			x.AssertEqualTime(t, expected.IssuedAt, actual.IssuedAt)
			x.AssertEqualTime(t, expected.ExpiresAt, actual.ExpiresAt)
			assert.EqualValues(t, expected.RequestURL, actual.RequestURL)
			assert.EqualValues(t, expected.Active, actual.Active)
			require.Equal(t, len(expected.Methods), len(actual.Methods), "expected:\t%s\nactual:\t%s", expected.Methods, actual.Methods)
		})

		t.Run("case=should properly set the flow type", func(t *testing.T) {
			expected := newFlow(t)
			expected.Forced = true
			expected.Type = flow.TypeAPI
			expected.Methods = map[identity.CredentialsType]*FlowMethod{
				identity.CredentialsTypeOIDC: {
					Method: identity.CredentialsTypeOIDC,
					Config: &FlowMethodConfig{FlowMethodConfigurator: form.NewHTMLForm(string(identity.CredentialsTypeOIDC))},
				},
				identity.CredentialsTypePassword: {
					Method: identity.CredentialsTypePassword,
					Config: &FlowMethodConfig{FlowMethodConfigurator: form.NewHTMLForm(string(identity.CredentialsTypePassword))},
				},
			}
			err := p.CreateLoginFlow(context.Background(), expected)
			require.NoError(t, err)

			actual, err := p.GetLoginFlow(context.Background(), expected.ID)
			require.NoError(t, err)
			assert.Equal(t, flow.TypeAPI, actual.Type)

			actual.Methods = map[identity.CredentialsType]*FlowMethod{identity.CredentialsTypeOIDC: {
				Method: identity.CredentialsTypeOIDC,
				Config: &FlowMethodConfig{FlowMethodConfigurator: form.NewHTMLForm("ory-sh")},
			}}
			actual.Type = flow.TypeBrowser
			actual.Forced = true

			require.NoError(t, p.UpdateLoginFlow(context.Background(), actual))

			actual, err = p.GetLoginFlow(context.Background(), actual.ID)
			require.NoError(t, err)
			assert.Equal(t, flow.TypeBrowser, actual.Type)
			assert.True(t, actual.Forced)
			require.Len(t, actual.Methods, 1)
			assert.Equal(t, "ory-sh",
				actual.Methods[identity.CredentialsTypeOIDC].Config.
					FlowMethodConfigurator.(*form.HTMLForm).Action)
		})

		t.Run("case=should properly update a flow", func(t *testing.T) {
			expected := newFlow(t)
			expected.Type = flow.TypeAPI
			err := p.CreateLoginFlow(context.Background(), expected)
			require.NoError(t, err)

			actual, err := p.GetLoginFlow(context.Background(), expected.ID)
			require.NoError(t, err)
			assert.Equal(t, flow.TypeAPI, actual.Type)
		})

		t.Run("case=should update a login flow", func(t *testing.T) {
			expected := newFlow(t)
			delete(expected.Methods, identity.CredentialsTypeOIDC)
			err := p.CreateLoginFlow(context.Background(), expected)
			require.NoError(t, err)

			actual, err := p.GetLoginFlow(context.Background(), expected.ID)
			require.NoError(t, err)
			assert.Len(t, actual.Methods, 1)

			require.NoError(t, p.UpdateLoginFlowMethod(context.Background(), expected.ID, identity.CredentialsTypeOIDC, &FlowMethod{
				Method: identity.CredentialsTypeOIDC,
				Config: &FlowMethodConfig{FlowMethodConfigurator: form.NewHTMLForm(string(identity.CredentialsTypeOIDC))},
			}))

			require.NoError(t, p.UpdateLoginFlowMethod(context.Background(), expected.ID, identity.CredentialsTypePassword, &FlowMethod{
				Method: identity.CredentialsTypePassword,
				Config: &FlowMethodConfig{FlowMethodConfigurator: form.NewHTMLForm(string(identity.CredentialsTypePassword))},
			}))

			actual, err = p.GetLoginFlow(context.Background(), expected.ID)
			require.NoError(t, err)
			require.Len(t, actual.Methods, 2)
			assert.EqualValues(t, identity.CredentialsTypePassword, actual.Active)

			assert.Equal(t, string(identity.CredentialsTypePassword), actual.Methods[identity.CredentialsTypePassword].Config.FlowMethodConfigurator.(*form.HTMLForm).Action)
			assert.Equal(t, string(identity.CredentialsTypeOIDC), actual.Methods[identity.CredentialsTypeOIDC].Config.FlowMethodConfigurator.(*form.HTMLForm).Action)
		})
	}
}
