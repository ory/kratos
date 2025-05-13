// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ory/kratos/x/nosurfx"

	"go.opentelemetry.io/otel/trace"

	"github.com/ory/kratos/x/events"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
)

type (
	PreHookExecutor interface {
		ExecuteRecoveryPreHook(w http.ResponseWriter, r *http.Request, a *Flow) error
	}
	PreHookExecutorFunc func(w http.ResponseWriter, r *http.Request, a *Flow) error

	PostHookExecutor interface {
		ExecutePostRecoveryHook(w http.ResponseWriter, r *http.Request, a *Flow, s *session.Session) error
	}
	PostHookExecutorFunc func(w http.ResponseWriter, r *http.Request, a *Flow, s *session.Session) error

	HooksProvider interface {
		PreRecoveryHooks(ctx context.Context) ([]PreHookExecutor, error)
		PostRecoveryHooks(ctx context.Context) ([]PostHookExecutor, error)
	}
)

func PostHookRecoveryExecutorNames(e []PostHookExecutor) []string {
	names := make([]string, len(e))
	for k, ee := range e {
		names[k] = fmt.Sprintf("%T", ee)
	}
	return names
}

func (f PreHookExecutorFunc) ExecuteRecoveryPreHook(w http.ResponseWriter, r *http.Request, a *Flow) error {
	return f(w, r, a)
}

func (f PostHookExecutorFunc) ExecutePostRecoveryHook(w http.ResponseWriter, r *http.Request, a *Flow, s *session.Session) error {
	return f(w, r, a, s)
}

type (
	executorDependencies interface {
		config.Provider
		identity.ManagementProvider
		identity.ValidationProvider
		session.PersistenceProvider
		HooksProvider
		nosurfx.CSRFTokenGeneratorProvider
		x.LoggingProvider
		x.WriterProvider
	}

	HookExecutor struct {
		d executorDependencies
	}

	HookExecutorProvider interface {
		RecoveryExecutor() *HookExecutor
	}
)

func NewHookExecutor(d executorDependencies) *HookExecutor {
	return &HookExecutor{
		d: d,
	}
}

func (e *HookExecutor) PostRecoveryHook(w http.ResponseWriter, r *http.Request, a *Flow, s *session.Session) error {
	logger := e.d.Logger().
		WithRequest(r)

	if s.Identity != nil {
		logger = logger.WithField("identity_id", s.Identity.ID)
	}

	logger.Debug("Running ExecutePostRecoveryHooks.")
	hooks, err := e.d.PostRecoveryHooks(r.Context())
	if err != nil {
		return err
	}
	for k, executor := range hooks {
		if err := executor.ExecutePostRecoveryHook(w, r, a, s); err != nil {
			var traits identity.Traits
			if s.Identity != nil {
				traits = s.Identity.Traits
			}
			return flow.HandleHookError(w, r, a, traits, node.LinkGroup, err, e.d, e.d)
		}

		logger.
			WithField("executor", fmt.Sprintf("%T", executor)).
			WithField("executor_position", k).
			WithField("executors", PostHookRecoveryExecutorNames(hooks)).
			Debug("ExecutePostRecoveryHook completed successfully.")
	}

	trace.SpanFromContext(r.Context()).AddEvent(events.NewRecoverySucceeded(r.Context(), a.ID, s.Identity.ID, string(a.Type), a.Active.String()))

	logger.Debug("Post recovery execution hooks completed successfully.")

	return nil
}

func (e *HookExecutor) PreRecoveryHook(w http.ResponseWriter, r *http.Request, a *Flow) error {
	hooks, err := e.d.PreRecoveryHooks(r.Context())
	if err != nil {
		return err
	}
	for _, executor := range hooks {
		if err := executor.ExecuteRecoveryPreHook(w, r, a); err != nil {
			return err
		}
	}

	return nil
}
