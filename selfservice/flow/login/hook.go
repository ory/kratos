package login

import (
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

type (
	PreHookExecutor interface {
		ExecuteLoginPreHook(w http.ResponseWriter, r *http.Request, a *Flow) error
	}

	PostHookExecutor interface {
		ExecuteLoginPostHook(w http.ResponseWriter, r *http.Request, a *Flow, s *session.Session) error
	}

	HooksProvider interface {
		PreLoginHooks() []PreHookExecutor
		PostLoginHooks(credentialsType identity.CredentialsType) []PostHookExecutor
	}
)

type (
	executorDependencies interface {
		HooksProvider
		session.ManagementProvider
		session.PersistenceProvider
		x.WriterProvider
		x.LoggingProvider
	}
	HookExecutor struct {
		d executorDependencies
		c configuration.Provider
	}
	HookExecutorProvider interface {
		LoginHookExecutor() *HookExecutor
	}
)

// Contains the Session and Session Token for API Based Authentication
//
// swagger:model sessionTokenContainer
type SessionTokenContainer struct {
	// The Session Token
	//
	// A session token is equivalent to a session cookie, but it can be sent in the HTTP Authorization
	// Header:
	//
	// 		Authorization: bearer <session-token>
	//
	// The session token is only issued for API flows, not for Browser flows!
	//
	// required: true
	Token string `json:"session_token"`

	// The Session
	//
	// The session contains information about the user, the session device, and so on.
	// This is only available for API flows, not for Browser flows!
	//
	// required: true
	Session *session.Session `json:"session"`
}

func NewHookExecutor(d executorDependencies, c configuration.Provider) *HookExecutor {
	return &HookExecutor{
		d: d,
		c: c,
	}
}

func (e *HookExecutor) PostLoginHook(w http.ResponseWriter, r *http.Request, ct identity.CredentialsType, a *Flow, i *identity.Identity) error {
	s := session.NewActiveSession(i, e.c, time.Now().UTC()).Declassify()

	for _, executor := range e.d.PostLoginHooks(ct) {
		if err := executor.ExecuteLoginPostHook(w, r, a, s); err != nil {
			if errors.Is(err, ErrHookAbortRequest) {
				e.d.Logger().Warn("A successful login attempt was aborted because a hook returned ErrHookAbortRequest.")
				return nil
			}
			return err
		}
	}

	if a.Type == flow.TypeAPI {
		if err := e.d.SessionPersister().CreateSession(r.Context(), s); err != nil {
			return errors.WithStack(err)
		}

		e.d.Writer().Write(w, r, &SessionTokenContainer{Session: s, Token: s.Token})
		return nil
	}

	if err := e.d.SessionManager().CreateAndIssueCookie(r.Context(), w, r, s); err != nil {
		return errors.WithStack(err)
	}

	return x.SecureContentNegotiationRedirection(w, r, s.Declassify(), a.RequestURL,
		e.d.Writer(), e.c, x.SecureRedirectOverrideDefaultReturnTo(e.c.SelfServiceFlowLoginReturnTo(ct.String())))
}

func (e *HookExecutor) PreLoginHook(w http.ResponseWriter, r *http.Request, a *Flow) error {
	for _, executor := range e.d.PreLoginHooks() {
		if err := executor.ExecuteLoginPreHook(w, r, a); err != nil {
			return err
		}
	}

	return nil
}
