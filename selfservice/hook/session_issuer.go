package hook

import (
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

var (
	_ registration.PostHookPostPersistExecutor = new(SessionIssuer)
)

type (
	sessionIssuerDependencies interface {
		session.ManagementProvider
		session.PersistenceProvider
		x.WriterProvider
	}
	SessionIssuerProvider interface {
		HookSessionIssuer() *SessionIssuer
	}
	SessionIssuer struct {
		r sessionIssuerDependencies
	}
)

func NewSessionIssuer(r sessionIssuerDependencies) *SessionIssuer {
	return &SessionIssuer{r: r}
}

func (e *SessionIssuer) ExecutePostRegistrationPostPersistHook(w http.ResponseWriter, r *http.Request, a *registration.Flow, s *session.Session) error {
	s.AuthenticatedAt = time.Now().UTC()
	if err := e.r.SessionPersister().CreateSession(r.Context(), s); err != nil {
		return err
	}

	if a.Type == flow.TypeAPI {
		e.r.Writer().Write(w, r, &registration.APIFlowResponse{
			Session: s, Token: s.Token,
			Identity: s.Identity,
		})
		return errors.WithStack(registration.ErrHookAbortFlow)
	}

	return e.r.SessionManager().IssueCookie(r.Context(), w, r, s)
}
