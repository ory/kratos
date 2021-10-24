package hook

import (
	"net/http"

	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/ui/node"
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

func (e *SessionDestroyer) ExecuteLoginPostHook(_ http.ResponseWriter, r *http.Request, _ node.Group, _ *login.Flow, s *session.Session) error {
	if err := e.r.SessionPersister().DeleteSessionsByIdentity(r.Context(), s.Identity.ID); err != nil {
		return err
	}
	return nil
}
