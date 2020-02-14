package hook

import (
	"net/http"

	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/verify"
	"github.com/ory/kratos/session"
)

var _ registration.PostHookExecutor = new(Verifier)

type (
	verifierDependencies interface {
		verify.SenderProvider
	}
	Verifier struct {
		r verifierDependencies
	}
)

func NewVerifier(r verifierDependencies) *Verifier {
	return &Verifier{r: r}
}

func (e *Verifier) ExecuteRegistrationPostHook(w http.ResponseWriter, r *http.Request, _ *registration.Request, s *session.Session) error {
	// Ths is called after the identity has been created so we can safely assume that all addresses are available
	// already.

	for _, address := range s.Identity.Addresses {
		if err := e.r.VerificationSender().SendCode(r.Context(), address.Via, address.Value); err != nil {
			return err
		}
	}

	return nil
}
