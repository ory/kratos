package passkey

import (
	"net/http"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/x"
)

func (s *Strategy) RegisterLoginRoutes(*x.RouterPublic) {}

func (s *Strategy) PopulateLoginMethod(r *http.Request, requestedAAL identity.AuthenticatorAssuranceLevel, sr *login.Flow) error {
	//TODO implement me
	panic("implement me")
}

func (s *Strategy) Login(w http.ResponseWriter, r *http.Request, f *login.Flow, identityID uuid.UUID) (i *identity.Identity, err error) {
	//TODO implement me
	panic("implement me")
}
