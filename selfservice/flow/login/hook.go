package login

import (
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
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

func NewHookExecutor(d executorDependencies, c configuration.Provider) *HookExecutor {
	return &HookExecutor{
		d: d,
		c: c,
	}
}

func (e *HookExecutor) PostLoginHook(w http.ResponseWriter, r *http.Request, ct identity.CredentialsType, a *Flow, i *identity.Identity) error {
	s := session.NewSession(i, e.c, time.Now().UTC()).Declassify()

	for _, executor := range e.d.PostLoginHooks(ct) {
		if err := executor.ExecuteLoginPostHook(w, r, a, s); err != nil {
			if errors.Is(err, ErrHookAbortRequest) {
				e.d.Logger().Warn("A successful login attempt was aborted because a hook returned ErrHookAbortRequest.")
				return nil
			}
			return err
		}
	}

	if err := e.d.SessionManager().CreateToRequest(r.Context(), w, r, s); err != nil {
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
