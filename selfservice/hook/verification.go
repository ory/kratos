package hook

import (
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/x"
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
		config.Provider
		x.CSRFTokenGeneratorProvider
		verification.StrategyProvider
	}
	Verifier struct {
		r verifierDependencies
	}
)

func NewVerifier(r verifierDependencies) *Verifier {
	return &Verifier{r: r}
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

		f, err := verification.NewFlow(e.r.Config(r.Context()), e.r.Config(r.Context()).SelfServiceFlowVerificationRequestLifespan(),
			e.r.GenerateCSRFToken(r), r, e.r.VerificationStrategies(r.Context()), flow.TypeBrowser)
		if err != nil {
			return err
		}

		token := link.NewVerificationToken(address, e.r.Config(r.Context()).SelfServiceFlowVerificationRequestLifespan())
		if err := e.r.VerificationTokenPersister().CreateVerificationToken(r.Context(), token); err != nil {
			return err
		}

		if err := e.r.LinkSender().SendVerificationTokenTo(r.Context(), f, address, token); err != nil {
			return err
		}
	}

	return nil
}
