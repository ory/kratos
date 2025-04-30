// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package verification

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
		ExecuteVerificationPreHook(w http.ResponseWriter, r *http.Request, a *Flow) error
	}
	PreHookExecutorFunc func(w http.ResponseWriter, r *http.Request, a *Flow) error

	PostHookExecutor interface {
		ExecutePostVerificationHook(w http.ResponseWriter, r *http.Request, a *Flow, i *identity.Identity) error
	}
	PostHookExecutorFunc func(w http.ResponseWriter, r *http.Request, a *Flow, i *identity.Identity) error

	HooksProvider interface {
		PostVerificationHooks(ctx context.Context) ([]PostHookExecutor, error)
		PreVerificationHooks(ctx context.Context) ([]PreHookExecutor, error)
	}
)

func PostHookVerificationExecutorNames(e []PostHookExecutor) []string {
	names := make([]string, len(e))
	for k, ee := range e {
		names[k] = fmt.Sprintf("%T", ee)
	}
	return names
}

func (f PreHookExecutorFunc) ExecuteVerificationPreHook(w http.ResponseWriter, r *http.Request, a *Flow) error {
	return f(w, r, a)
}

func (f PostHookExecutorFunc) ExecutePostVerificationHook(w http.ResponseWriter, r *http.Request, a *Flow, i *identity.Identity) error {
	return f(w, r, a, i)
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
		VerificationExecutor() *HookExecutor
	}
)

func NewHookExecutor(d executorDependencies) *HookExecutor {
	return &HookExecutor{
		d: d,
	}
}

func (e *HookExecutor) PreVerificationHook(w http.ResponseWriter, r *http.Request, a *Flow) error {
	hooks, err := e.d.PreVerificationHooks(r.Context())
	if err != nil {
		return err
	}
	for _, executor := range hooks {
		if err := executor.ExecuteVerificationPreHook(w, r, a); err != nil {
			return err
		}
	}

	return nil
}

func (e *HookExecutor) PostVerificationHook(w http.ResponseWriter, r *http.Request, a *Flow, i *identity.Identity) error {
	e.d.Logger().
		WithRequest(r).
		WithField("identity_id", i.ID).
		Debug("Running ExecutePostVerificationHooks.")
	hooks, err := e.d.PostVerificationHooks(r.Context())
	if err != nil {
		return err
	}
	for k, executor := range hooks {
		if err := executor.ExecutePostVerificationHook(w, r, a, i); err != nil {
			return flow.HandleHookError(w, r, a, i.Traits, node.LinkGroup, err, e.d, e.d)
		}

		e.d.Logger().WithRequest(r).
			WithField("executor", fmt.Sprintf("%T", executor)).
			WithField("executor_position", k).
			WithField("executors", PostHookVerificationExecutorNames(hooks)).
			WithField("identity_id", i.ID).
			Debug("ExecutePostVerificationHook completed successfully.")
	}

	trace.SpanFromContext(r.Context()).AddEvent(events.NewVerificationSucceeded(r.Context(), a.ID, i.ID, string(a.Type), a.Active.String()))

	e.d.Logger().
		WithRequest(r).
		WithField("identity_id", i.ID).
		Debug("Post verification execution hooks completed successfully.")

	return nil
}
