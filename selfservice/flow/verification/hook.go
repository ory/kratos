package verification

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
)

type (
	PostHookExecutor interface {
		ExecutePostVerificationHook(w http.ResponseWriter, r *http.Request, a *Flow, i *identity.Identity) error
	}
	PostHookExecutorFunc func(w http.ResponseWriter, r *http.Request, a *Flow, i *identity.Identity) error

	HooksProvider interface {
		PostVerificationHooks(ctx context.Context) []PostHookExecutor
	}
)

func PostHookVerificationExecutorNames(e []PostHookExecutor) []string {
	names := make([]string, len(e))
	for k, ee := range e {
		names[k] = fmt.Sprintf("%T", ee)
	}
	return names
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
		x.CSRFTokenGeneratorProvider
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

func (e *HookExecutor) handleVerificationError(_ http.ResponseWriter, r *http.Request, f *Flow, i *identity.Identity, flowError error) error {
	if f != nil {
		if i != nil {
			cont, err := container.NewFromStruct("", node.RecoveryLinkGroup, i.Traits, "traits")
			if err != nil {
				e.d.Logger().WithError(err)("could not update flow UI")
				return err
			}

			for _, n := range cont.Nodes {
				// we only set the value and not the whole field because we want to keep types from the initial form generation
				f.UI.Nodes.SetValueAttribute(n.ID(), n.Attributes.GetValue())
			}
		}

		if f.Type == flow.TypeBrowser {
			f.UI.SetCSRF(e.d.GenerateCSRFToken(r))
		}
	}

	return flowError
}

func (e *HookExecutor) PostVerificationHook(w http.ResponseWriter, r *http.Request, a *Flow, i *identity.Identity) error {
	e.d.Logger().
		WithRequest(r).
		WithField("identity_id", i.ID).
		Debug("Running ExecutePostVerificationHooks.")
	for k, executor := range e.d.PostVerificationHooks(r.Context()) {
		if err := executor.ExecutePostVerificationHook(w, r, a, i); err != nil {
			return e.handleVerificationError(w, r, a, i, err)
		}

		e.d.Logger().WithRequest(r).
			WithField("executor", fmt.Sprintf("%T", executor)).
			WithField("executor_position", k).
			WithField("executors", PostHookVerificationExecutorNames(e.d.PostVerificationHooks(r.Context()))).
			WithField("identity_id", i.ID).
			Debug("ExecutePostVerificationHook completed successfully.")
	}

	e.d.Logger().
		WithRequest(r).
		WithField("identity_id", i.ID).
		Debug("Post verification execution hooks completed successfully.")

	return nil
}
