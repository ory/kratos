package hooks

import (
	"net/http"

	"github.com/ory/hive-cloud/hive/selfservice"
	"github.com/ory/hive-cloud/hive/session"
)

var _ selfservice.HookLoginPostExecutor = new(SessionIssuer)
var _ selfservice.HookRegistrationPostExecutor = new(SessionIssuer)

type sessionIssuerDependencies interface {
	session.ManagementProvider
}

type SessionIssuer struct {
	r sessionIssuerDependencies
}

func NewSessionIssuer(r sessionIssuerDependencies) *SessionIssuer {
	return &SessionIssuer{r: r}
}

func (e *SessionIssuer) ExecuteRegistrationPostHook(w http.ResponseWriter, r *http.Request, a *selfservice.RegistrationRequest, s *session.Session) error {
	if err := e.r.SessionManager().Create(s); err != nil {
		return err
	}
	return e.r.SessionManager().SaveToRequest(s, w, r)
}

func (e *SessionIssuer) ExecuteLoginPostHook(w http.ResponseWriter, r *http.Request, a *selfservice.LoginRequest, s *session.Session) error {
	if err := e.r.SessionManager().Create(s); err != nil {
		return err
	}
	return e.r.SessionManager().SaveToRequest(s, w, r)
}
