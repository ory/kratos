package registration

import (
	"context"
	"testing"

	"github.com/bxcodec/faker"
	"github.com/gobuffalo/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/x"
)

type RequestPersister interface {
	CreateRegistrationRequest(context.Context, *Request) error
	GetRegistrationRequest(context.Context, uuid.UUID) (*Request, error)
	UpdateRegistrationRequest(context.Context, uuid.UUID, identity.CredentialsType, *RequestMethod) error
}

type RequestPersistenceProvider interface {
	RegistrationRequestPersister() RequestPersister
}


func TestRequestPersister(p RequestPersister) func(t *testing.T) {
	var clearids = func(r *Request) {
		r.ID = uuid.UUID{}
		for k := range r.Methods {
			r.Methods[k].ID = uuid.UUID{}
		}
	}

	return func(t *testing.T) {
		t.Run("case=should error when the registration request does not exist", func(t *testing.T) {
			_, err := p.GetRegistrationRequest(context.Background(), x.NewUUID())
			require.Error(t, err)
		})

		var newRequest = func(t *testing.T) *Request {
			var r Request
			require.NoError(t, faker.FakeData(&r))
			clearids(&r)

			methods := len(r.Methods)
			assert.NotZero(t, methods)

			return &r
		}

		t.Run("case=should create a new registration request and properly set IDs", func(t *testing.T) {
			r := newRequest(t)
			methods := len(r.Methods)
			err := p.CreateRegistrationRequest(context.Background(), r)
			require.NoError(t, err, "%#v", err)

			assert.Nil(t, r.MethodsRaw)
			assert.NotEmpty(t, r.ID)
			for _, m := range r.Methods {
				assert.NotEmpty(t, m.ID)
			}
			assert.Len(t, r.Methods, methods)
		})

		t.Run("case=should create and fetch a registration request", func(t *testing.T) {
			expected := newRequest(t)
			err := p.CreateRegistrationRequest(context.Background(), expected)
			require.NoError(t, err)

			actual, err := p.GetRegistrationRequest(context.Background(), expected.ID)
			require.NoError(t, err)
			assert.Empty(t, actual.MethodsRaw)

			assert.EqualValues(t, expected.ID, actual.ID)
			assert.EqualValues(t, expected.ExpiresAt, actual.ExpiresAt)
			assert.EqualValues(t, expected.IssuedAt, actual.IssuedAt)
			assert.EqualValues(t, expected.RequestURL, actual.RequestURL)
			assert.EqualValues(t, expected.Active, actual.Active)
			require.Equal(t, len(expected.Methods), len(actual.Methods), "expected:\t%s\nactual:\t%s", expected.Methods, actual.Methods)
		})

		t.Run("case=should update a registration request", func(t *testing.T) {
			expected := newRequest(t)
			delete(expected.Methods, identity.CredentialsTypeOIDC)
			err := p.CreateRegistrationRequest(context.Background(), expected)
			require.NoError(t, err)

			actual, err := p.GetRegistrationRequest(context.Background(), expected.ID)
			require.NoError(t, err)
			assert.Len(t, actual.Methods, 1)

			require.NoError(t, p.UpdateRegistrationRequest(context.Background(), expected.ID, identity.CredentialsTypeOIDC, &RequestMethod{
				Method: identity.CredentialsTypeOIDC,
				Config: &RequestMethodConfig{ form.NewHTMLForm( string(identity.CredentialsTypeOIDC))},
			}))

			require.NoError(t, p.UpdateRegistrationRequest(context.Background(), expected.ID, identity.CredentialsTypePassword, &RequestMethod{
				Method: identity.CredentialsTypePassword,
				Config: &RequestMethodConfig{form.NewHTMLForm(string(identity.CredentialsTypePassword))},
			}))

			actual, err = p.GetRegistrationRequest(context.Background(), expected.ID)
			require.NoError(t, err)
			require.Len(t, actual.Methods, 2)

			assert.Equal(t, string(identity.CredentialsTypePassword), actual.Methods[identity.CredentialsTypePassword].Config.RequestMethodConfigurator.(*form.HTMLForm).Action)
			assert.Equal(t, string(identity.CredentialsTypeOIDC), actual.Methods[identity.CredentialsTypeOIDC].Config.RequestMethodConfigurator.(*form.HTMLForm).Action)
		})
	}
}
