package hook

import (
	"net/http"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/flow/verify"
	"github.com/ory/kratos/session"
)

var _ registration.PostHookPostPersistExecutor = new(Verifier)
var _ settings.PostHookPostPersistExecutor = new(Verifier)

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

func (e *Verifier) ExecutePostRegistrationPostPersistHook(_ http.ResponseWriter, r *http.Request, _ *registration.Request, s *session.Session) error {
	return e.do(r, s.Identity)
}

func (e *Verifier) ExecuteSettingsPostPersistHook(w http.ResponseWriter, r *http.Request, a *settings.Request, i *identity.Identity) error {
	return e.do(r, i)
}

func (e *Verifier) do(r *http.Request, i *identity.Identity) error {
	// Ths is called after the identity has been created so we can safely assume that all addresses are available
	// already.

	for k, address := range i.VerifiableAddresses {
		sent, err := e.r.VerificationSender().SendCode(r.Context(), address.Via, address.Value)
		if err != nil {
			return err
		}
		i.VerifiableAddresses[k] = *sent
	}

	return nil
}
