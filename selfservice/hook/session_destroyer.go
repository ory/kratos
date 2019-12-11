package hook

import (
	"net/http"

	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/session"
)

var _ login.PostHookExecutor = new(SessionDestroyer)
var _ registration.PostHookExecutor = new(SessionIssuer)

type sessionDestroyerDependencies interface {
	session.ManagementProvider
	session.PersistenceProvider
}

type SessionDestroyer struct {
	r sessionDestroyerDependencies
}

func NewSessionDestroyer(r sessionDestroyerDependencies) *SessionDestroyer {
	return &SessionDestroyer{r: r}
}

func (e *SessionDestroyer) ExecuteRegistrationPostHook(w http.ResponseWriter, r *http.Request, a *registration.Request, s *session.Session) error {
	return nil
}

func (e *SessionDestroyer) ExecuteLoginPostHook(w http.ResponseWriter, r *http.Request, a *login.Request, s *session.Session) error {
	if err := e.r.SessionPersister().DeleteSessionsFor(r.Context(), s.Identity.ID); err != nil {
		return err
	}
	return nil
}
