package login

import (
	"context"
	"testing"

	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/require"
	"github.com/gofrs/uuid"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"
)

type (
	RequestPersister interface {
		CreateLoginRequest(context.Context, *Request) error
		GetLoginRequest(context.Context, uuid.UUID) (*Request, error)
		UpdateLoginRequest(context.Context, uuid.UUID, identity.CredentialsType, *RequestMethod) error
	}
	RequestPersistenceProvider interface {
		LoginRequestPersister() RequestPersister
	}
)

func TestRequestPersister(t *testing.T, p RequestPersister) {
	// nbr := func() *Request {
	// 	return &Request{
	// 		ID:             uuid.New().String(),
	// 		IssuedAt:       time.Now().UTC().Round(time.Second),
	// 		ExpiresAt:      time.Now().Add(time.Hour).UTC().Round(time.Second),
	// 		RequestURL:     "https://www.ory.sh/request",
	// 		RequestHeaders: http.Header{"Content-Type": {"application/json"}},
	// 		// Disable Active as this value is initially empty (NULL).
	// 		// Active:         identity.CredentialsTypePassword,
	// 		Methods: map[identity.CredentialsType]*CredentialsRequest{
	// 			identity.CredentialsTypePassword: {
	// 				Method: identity.CredentialsTypePassword,
	// 				Config: password.NewRequestMethodConfig(),
	// 			},
	// 			identity.CredentialsTypeOIDC: {
	// 				Method: identity.CredentialsTypeOIDC,
	// 				Config: oidc.NewRequestMethodConfig(),
	// 			},
	// 		},
	// 	}
	// }
	//
	// assertUpdated := func(t *testing.T, expected, actual Request) {
	// 	assert.EqualValues(t, identity.CredentialsTypePassword, actual.Active)
	// 	assert.EqualValues(t, "bar", actual.Methods[identity.CredentialsTypeOIDC].Config.(*oidc.RequestMethodConfig).Action)
	// 	assert.EqualValues(t, "foo", actual.Methods[identity.CredentialsTypePassword].Config.(*password.RequestMethodConfig).Action)
	// }

	t.Run("case=should error when the login request does not exist", func(t *testing.T) {
		_, err := p.GetLoginRequest(context.Background(), x.NewUUID())
		require.NoError(t, err)
	})

	t.Run("case=", func(t *testing.T) {
		var r Request
		require.NoError(t, faker.FakeData(&r))
		t.Logf("%+v", r)
	})

	// r := LoginRequest{Request: nbr()}
	// require.NoError(t, p.CreateLoginRequest(context.Background(), &r))
	//
	// g, err := p.GetLoginRequest(context.Background(), r.ID)
	// require.NoError(t, err)
	// assert.EqualValues(t, r, *g)
	//
	// require.NoError(t, p.UpdateLoginRequest(context.Background(), r.ID, identity.CredentialsTypeOIDC, &oidc.RequestMethod{Action: "bar"}))
	// require.NoError(t, p.UpdateLoginRequest(context.Background(), r.ID, identity.CredentialsTypePassword, &password.RequestMethod{Action: "foo"}))
	//
	// g, err = p.GetLoginRequest(context.Background(), r.ID)
	// require.NoError(t, err)
	// assertUpdated(t, *r.Request, *g.Request)
}
