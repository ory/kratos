package oid2

import (
	"context"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"net/http"
)

var _ login.Strategy = new(Strategy)

// TODO #3631 implement OID2 login

func (s *Strategy) RegisterLoginRoutes(publicRouter *x.RouterPublic) {

}

func (s *Strategy) PopulateLoginMethod(r *http.Request, requestedAAL identity.AuthenticatorAssuranceLevel, sr *login.Flow) error {
	return nil
}

func (s *Strategy) Login(w http.ResponseWriter, r *http.Request, f *login.Flow, sess *session.Session) (i *identity.Identity, err error) {
	return nil, nil
}

func (s *Strategy) CompletedAuthenticationMethod(ctx context.Context, methods session.AuthenticationMethods) session.AuthenticationMethod {
	return session.AuthenticationMethod{
		Method: s.ID(),
		AAL:    identity.AuthenticatorAssuranceLevel1,
	}
}
