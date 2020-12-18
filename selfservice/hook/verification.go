package hook

import (
	"net/http"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/strategy/link"
	"github.com/ory/kratos/session"
)

var _ registration.PostHookPostPersistExecutor = new(Verifier)
var _ settings.PostHookPostPersistExecutor = new(Verifier)

type (
	verifierDependencies interface {
		link.SenderProvider
		link.VerificationTokenPersistenceProvider
	}
	Verifier struct {
		r verifierDependencies
		c *config.Config
	}
)

func NewVerifier(r verifierDependencies, c *config.Config) *Verifier {
	return &Verifier{r: r, c: c}
}

func (e *Verifier) ExecutePostRegistrationPostPersistHook(_ http.ResponseWriter, r *http.Request, _ *registration.Flow, s *session.Session) error {
	return e.do(r, s.Identity)
}

func (e *Verifier) ExecuteSettingsPostPersistHook(w http.ResponseWriter, r *http.Request, a *settings.Flow, i *identity.Identity) error {
	return e.do(r, i)
}

func (e *Verifier) do(r *http.Request, i *identity.Identity) error {
	// Ths is called after the identity has been created so we can safely assume that all addresses are available
	// already.

	for k := range i.VerifiableAddresses {
		address := &i.VerifiableAddresses[k]
		if address.Verified {
			continue
		}

		token := link.NewVerificationToken(address, e.c.SelfServiceFlowVerificationRequestLifespan())
		if err := e.r.VerificationTokenPersister().CreateVerificationToken(r.Context(), token); err != nil {
			return err
		}

		if err := e.r.LinkSender().SendVerificationTokenTo(r.Context(), address, token); err != nil {
			return err
		}
	}

	return nil
}
