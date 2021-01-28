package login

import (
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/ory/kratos/driver/config"
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
		config.Provider
		session.ManagementProvider
		session.PersistenceProvider
		x.WriterProvider
		x.LoggingProvider

		HooksProvider
	}
	HookExecutor struct {
		d executorDependencies
	}
	HookExecutorProvider interface {
		LoginHookExecutor() *HookExecutor
	}
)

func PostHookExecutorNames(e []PostHookExecutor) []string {
	names := make([]string, len(e))
	for k, ee := range e {
		names[k] = fmt.Sprintf("%T", ee)
	}
	return names
}

func NewHookExecutor(d executorDependencies) *HookExecutor {
	return &HookExecutor{d: d}
}

func (e *HookExecutor) PostLoginHook(w http.ResponseWriter, r *http.Request, ct identity.CredentialsType, a *Flow, i *identity.Identity) error {
	s := session.NewActiveSession(i, e.d.Config(r.Context()), time.Now().UTC()).Declassify()

	e.d.Logger().
		WithRequest(r).
		WithField("identity_id", i.ID).
		WithField("flow_method", ct).
		Debug("Running ExecuteLoginPostHook.")
	for k, executor := range e.d.PostLoginHooks(ct) {
		if err := executor.ExecuteLoginPostHook(w, r, a, s); err != nil {
			if errors.Is(err, ErrHookAbortFlow) {
				e.d.Logger().
					WithRequest(r).
					WithField("executor", fmt.Sprintf("%T", executor)).
					WithField("executor_position", k).
					WithField("executors", PostHookExecutorNames(e.d.PostLoginHooks(ct))).
					WithField("identity_id", i.ID).
					WithField("flow_method", ct).
					Debug("A ExecuteLoginPostHook hook aborted early.")
				return nil
			}
			return err
		}

		e.d.Logger().
			WithRequest(r).
			WithField("executor", fmt.Sprintf("%T", executor)).
			WithField("executor_position", k).
			WithField("executors", PostHookExecutorNames(e.d.PostLoginHooks(ct))).
			WithField("identity_id", i.ID).
			WithField("flow_method", ct).
			Debug("ExecuteLoginPostHook completed successfully.")
	}

	if a.Type == flow.TypeAPI {
		if err := e.d.SessionPersister().CreateSession(r.Context(), s); err != nil {
			return errors.WithStack(err)
		}
		e.d.Audit().
			WithRequest(r).
			WithField("session_id", s.ID).
			WithField("identity_id", i.ID).
			Info("Identity authenticated successfully and was issued an ORY Kratos Session Token.")

		e.d.Writer().Write(w, r, &APIFlowResponse{Session: s, Token: s.Token})
		return nil
	}

	if err := e.d.SessionManager().CreateAndIssueCookie(r.Context(), w, r, s); err != nil {
		return errors.WithStack(err)
	}

	e.d.Audit().
		WithRequest(r).
		WithField("identity_id", i.ID).
		WithField("session_id", s.ID).
		Info("Identity authenticated successfully and was issued an ORY Kratos Session Cookie.")
	return x.SecureContentNegotiationRedirection(w, r, s.Declassify(), a.RequestURL,
		e.d.Writer(), e.d.Config(r.Context()), x.SecureRedirectOverrideDefaultReturnTo(e.d.Config(r.Context()).SelfServiceFlowLoginReturnTo(ct.String())))
}

func (e *HookExecutor) PreLoginHook(w http.ResponseWriter, r *http.Request, a *Flow) error {
	for _, executor := range e.d.PreLoginHooks() {
		if err := executor.ExecuteLoginPreHook(w, r, a); err != nil {
			return err
		}
	}

	return nil
}
