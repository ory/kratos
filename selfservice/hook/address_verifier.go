package hook

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/session"
)

var _ login.PostHookExecutor = new(SessionDestroyer)

var ErrAddressNotVerified = errors.New("address not verified yet")

type AddressVerifier struct{}

func NewAddressVerifier() *AddressVerifier {
	return &AddressVerifier{}
}

func (e *AddressVerifier) ExecuteLoginPostHook(_ http.ResponseWriter, _ *http.Request, _ *login.Flow, s *session.Session) error {
	// all addresses, the user is using for identification purposes, must be verified
	for _, va := range s.Identity.VerifiableAddresses {
		if !va.Verified {
			return ErrAddressNotVerified
		}
	}
	return nil
}
