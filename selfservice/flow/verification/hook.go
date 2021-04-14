package verification

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

type (
	PostHookExecutor interface {
		ExecutePostVerificationHook(w http.ResponseWriter, r *http.Request, a *Flow, s *session.Session) error
	}
	PostHookExecutorFunc func(w http.ResponseWriter, r *http.Request, a *Flow, s *session.Session) error

	HooksProvider interface {
		PostVerificationHooks() []PostHookExecutor
	}
)

func PostHookVerificationExecutorNames(e []PostHookExecutor) []string {
	names := make([]string, len(e))
	for k, ee := range e {
		names[k] = fmt.Sprintf("%T", ee)
	}
	return names
}

func (f PostHookExecutorFunc) ExecutePostVerificationHook(w http.ResponseWriter, r *http.Request, a *Flow, s *session.Session) error {
	return f(w, r, a, s)
}

type (
	executorDependencies interface {
		identity.ManagementProvider
		identity.ValidationProvider
		session.PersistenceProvider
		HooksProvider
		x.LoggingProvider
		x.WriterProvider
	}
	HookExecutor struct {
		d executorDependencies
		c config.Provider
	}
	HookExecutorProvider interface {
		VerificationExecutor() *HookExecutor
	}
)

func NewHookExecutor(d executorDependencies) *HookExecutor {
	return &HookExecutor{
		d: d,
	}
}

func (e *HookExecutor) PostVerificationHook(w http.ResponseWriter, r *http.Request, a *Flow, i *identity.Identity) error {
	s := session.NewActiveSession(i, e.c.Config(r.Context()), time.Now().UTC())
	e.d.Logger().
		WithRequest(r).
		WithField("identity_id", i.ID).
		Debug("Running ExecutePostVerificationHooks.")
	for k, executor := range e.d.PostVerificationHooks() {
		if err := executor.ExecutePostVerificationHook(w, r, a, s); err != nil {
			return err
		}

		e.d.Logger().WithRequest(r).
			WithField("executor", fmt.Sprintf("%T", executor)).
			WithField("executor_position", k).
			WithField("executors", PostHookVerificationExecutorNames(e.d.PostVerificationHooks())).
			WithField("identity_id", i.ID).
			Debug("ExecutePostVerificationHook completed successfully.")
	}

	e.d.Logger().
		WithRequest(r).
		WithField("identity_id", i.ID).
		Debug("Post verification execution hooks completed successfully.")

	return nil
}
