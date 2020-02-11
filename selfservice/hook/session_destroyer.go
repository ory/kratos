package hook

import (
	"net/http"

	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/session"
)

var _ login.PostHookExecutor = new(SessionDestroyer)

type (
	sessionDestroyerDependencies interface {
		session.ManagementProvider
		session.PersistenceProvider
	}
	SessionDestroyer struct {
		r sessionDestroyerDependencies
	}
)

func NewSessionDestroyer(r sessionDestroyerDependencies) *SessionDestroyer {
	return &SessionDestroyer{r: r}
}

func (e *SessionDestroyer) ExecuteLoginPostHook(_ http.ResponseWriter, r *http.Request, _ *login.Request, s *session.Session) error {
	if err := e.r.SessionPersister().DeleteSessionsFor(r.Context(), s.Identity.ID); err != nil {
		return err
	}
	return nil
}
