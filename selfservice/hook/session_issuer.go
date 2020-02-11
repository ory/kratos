package hook

import (
	"net/http"

	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/session"
)

var (
	_ login.PostHookExecutor        = new(SessionIssuer)
	_ registration.PostHookExecutor = new(SessionIssuer)
)

type (
	sessionIssuerDependencies interface {
		session.ManagementProvider
		session.PersistenceProvider
	}
	SessionIssuer struct {
		r sessionIssuerDependencies
	}
)

func NewSessionIssuer(r sessionIssuerDependencies) *SessionIssuer {
	return &SessionIssuer{r: r}
}

func (e *SessionIssuer) ExecuteRegistrationPostHook(w http.ResponseWriter, r *http.Request, a *registration.Request, s *session.Session) error {
	if err := e.r.SessionPersister().CreateSession(r.Context(), s); err != nil {
		return err
	}
	return e.r.SessionManager().SaveToRequest(r.Context(), s, w, r)
}

func (e *SessionIssuer) ExecuteLoginPostHook(w http.ResponseWriter, r *http.Request, a *login.Request, s *session.Session) error {
	if err := e.r.SessionPersister().CreateSession(r.Context(), s); err != nil {
		return err
	}
	return e.r.SessionManager().SaveToRequest(r.Context(), s, w, r)
}
