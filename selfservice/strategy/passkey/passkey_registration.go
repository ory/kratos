package passkey

import (
	"net/http"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/x"
)

func (s *Strategy) RegisterRegistrationRoutes(public *x.RouterPublic) {
	//TODO implement me
	panic("implement me")
}

func (s *Strategy) PopulateRegistrationMethod(r *http.Request, sr *registration.Flow) error {
	//TODO implement me
	panic("implement me")
}

func (s *Strategy) Register(w http.ResponseWriter, r *http.Request, f *registration.Flow, i *identity.Identity) (err error) {
	//TODO implement me
	panic("implement me")
}
