package hook

import (
	"net/http"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/session"
)

var _ login.PostHookExecutor = new(SessionDestroyer)

type AddressVerifier struct{}

func NewAddressVerifier() *AddressVerifier {
	return &AddressVerifier{}
}

func (e *AddressVerifier) ExecuteLoginPostHook(_ http.ResponseWriter, _ *http.Request, f *login.Flow, s *session.Session) error {
	// if the login happens using the password method, there must be at least one verified address
	if f.Active != identity.CredentialsTypePassword {
		return nil
	}

	// TODO: can this happen at all?
	if len(s.Identity.VerifiableAddresses) == 0 {
		return login.ErrAddressNotVerified
	}

	addressVerified := false
	for _, va := range s.Identity.VerifiableAddresses {
		if va.Verified {
			addressVerified = true
			break
		}
	}

	if !addressVerified {
		return login.ErrAddressNotVerified
	}

	return nil
}
