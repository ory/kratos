// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package settings

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ory/kratos/session"

	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/schema"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/x"
)

type (
	PreHookExecutor interface {
		ExecuteSettingsPreHook(w http.ResponseWriter, r *http.Request, a *Flow) error
	}
	PreHookExecutorFunc func(w http.ResponseWriter, r *http.Request, a *Flow) error

	PostHookPrePersistExecutor interface {
		ExecuteSettingsPrePersistHook(w http.ResponseWriter, r *http.Request, a *Flow, s *identity.Identity) error
	}
	PostHookPrePersistExecutorFunc func(w http.ResponseWriter, r *http.Request, a *Flow, s *identity.Identity) error

	PostHookPostPersistExecutor interface {
		ExecuteSettingsPostPersistHook(w http.ResponseWriter, r *http.Request, a *Flow, s *identity.Identity) error
	}
	PostHookPostPersistExecutorFunc func(w http.ResponseWriter, r *http.Request, a *Flow, s *identity.Identity) error

	HooksProvider interface {
		PreSettingsHooks(ctx context.Context) []PreHookExecutor
		PostSettingsPrePersistHooks(ctx context.Context, settingsType string) []PostHookPrePersistExecutor
		PostSettingsPostPersistHooks(ctx context.Context, settingsType string) []PostHookPostPersistExecutor
	}

	executorDependencies interface {
		identity.ManagementProvider
		identity.ValidationProvider
		session.ManagementProvider
		config.Provider

		HandlerProvider
		HooksProvider
		FlowPersistenceProvider

		x.CSRFTokenGeneratorProvider
		x.LoggingProvider
		x.WriterProvider
	}
	HookExecutor struct {
		d executorDependencies
	}
	HookExecutorProvider interface {
		SettingsHookExecutor() *HookExecutor
	}
)

func (f PreHookExecutorFunc) ExecuteSettingsPreHook(w http.ResponseWriter, r *http.Request, a *Flow) error {
	return f(w, r, a)
}

func (f PostHookPrePersistExecutorFunc) ExecuteSettingsPrePersistHook(w http.ResponseWriter, r *http.Request, a *Flow, s *identity.Identity) error {
	return f(w, r, a, s)
}

func (f PostHookPostPersistExecutorFunc) ExecuteSettingsPostPersistHook(w http.ResponseWriter, r *http.Request, a *Flow, s *identity.Identity) error {
	return f(w, r, a, s)
}

func PostHookPostPersistExecutorNames(e []PostHookPostPersistExecutor) []string {
	names := make([]string, len(e))
	for k, ee := range e {
		names[k] = fmt.Sprintf("%T", ee)
	}
	return names
}

func PostHookPrePersistExecutorNames(e []PostHookPrePersistExecutor) []string {
	names := make([]string, len(e))
	for k, ee := range e {
		names[k] = fmt.Sprintf("%T", ee)
	}
	return names
}

func NewHookExecutor(d executorDependencies) *HookExecutor {
	return &HookExecutor{d: d}
}

type PostSettingsHookOption func(o *postSettingsHookOptions)

type postSettingsHookOptions struct {
	cb func(ctxUpdate *UpdateContext) error
}

func WithCallback(cb func(ctxUpdate *UpdateContext) error) func(o *postSettingsHookOptions) {
	return func(o *postSettingsHookOptions) {
		o.cb = cb
	}
}

func (e *HookExecutor) handleSettingsError(_ http.ResponseWriter, r *http.Request, settingsType string, f *Flow, i *identity.Identity, flowError error) error {
	if f != nil {
		if i != nil {
			var group node.UiNodeGroup
			switch settingsType {
			case "password":
				group = node.PasswordGroup
			case "oidc":
				group = node.OpenIDConnectGroup
			}

			cont, err := container.NewFromStruct("", group, i.Traits, "traits")
			if err != nil {
				e.d.Logger().WithError(err).Error("could not update flow UI")
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

func (e *HookExecutor) PostSettingsHook(w http.ResponseWriter, r *http.Request, settingsType string, ctxUpdate *UpdateContext, i *identity.Identity, opts ...PostSettingsHookOption) error {
	e.d.Logger().
		WithRequest(r).
		WithField("identity_id", i.ID).
		WithField("flow_method", settingsType).
		Debug("Running PostSettingsPrePersistHooks.")

	// Verify the redirect URL before we do any other processing.
	c := e.d.Config()
	returnTo, err := x.SecureRedirectTo(r, c.SelfServiceBrowserDefaultReturnTo(r.Context()),
		x.SecureRedirectUseSourceURL(ctxUpdate.Flow.RequestURL),
		x.SecureRedirectAllowURLs(c.SelfServiceBrowserAllowedReturnToDomains(r.Context())),
		x.SecureRedirectAllowSelfServiceURLs(c.SelfPublicURL(r.Context())),
		x.SecureRedirectOverrideDefaultReturnTo(
			e.d.Config().SelfServiceFlowSettingsReturnTo(r.Context(), settingsType,
				ctxUpdate.Flow.AppendTo(e.d.Config().SelfServiceFlowSettingsUI(r.Context())))),
	)
	if err != nil {
		return err
	}

	hookOptions := new(postSettingsHookOptions)
	for _, f := range opts {
		f(hookOptions)
	}

	for k, executor := range e.d.PostSettingsPrePersistHooks(r.Context(), settingsType) {
		logFields := logrus.Fields{
			"executor":          fmt.Sprintf("%T", executor),
			"executor_position": k,
			"executors":         PostHookPrePersistExecutorNames(e.d.PostSettingsPrePersistHooks(r.Context(), settingsType)),
			"identity_id":       i.ID,
			"flow_method":       settingsType,
		}

		if err := executor.ExecuteSettingsPrePersistHook(w, r, ctxUpdate.Flow, i); err != nil {
			if errors.Is(err, ErrHookAbortFlow) {
				e.d.Logger().WithRequest(r).WithFields(logFields).
					Debug("A ExecuteSettingsPrePersistHook hook aborted early.")
				return nil
			}
			var group node.UiNodeGroup
			switch settingsType {
			case "password":
				group = node.PasswordGroup
			case "oidc":
				group = node.OpenIDConnectGroup
			}
			var traits identity.Traits
			if i != nil {
				traits = i.Traits
			}
			return flow.HandleHookError(w, r, ctxUpdate.Flow, traits, group, err, e.d, e.d)
		}

		e.d.Logger().WithRequest(r).WithFields(logFields).Debug("ExecuteSettingsPrePersistHook completed successfully.")
	}

	options := []identity.ManagerOption{identity.ManagerExposeValidationErrorsForInternalTypeAssertion}
	ttl := e.d.Config().SelfServiceFlowSettingsPrivilegedSessionMaxAge(r.Context())
	if ctxUpdate.Session.AuthenticatedAt.Add(ttl).After(time.Now()) {
		options = append(options, identity.ManagerAllowWriteProtectedTraits)
	}

	if err := e.d.IdentityManager().Update(r.Context(), i, options...); err != nil {
		if errors.Is(err, identity.ErrProtectedFieldModified) {
			e.d.Logger().WithError(err).Debug("Modifying protected field requires re-authentication.")
			return errors.WithStack(NewFlowNeedsReAuth())
		}
		if errors.Is(err, sqlcon.ErrUniqueViolation) {
			return schema.NewDuplicateCredentialsError()
		}
		return err
	}
	e.d.Audit().
		WithRequest(r).
		WithField("identity_id", i.ID).
		Debug("An identity's settings have been updated.")

	ctxUpdate.UpdateIdentity(i)
	ctxUpdate.Flow.State = StateSuccess
	if hookOptions.cb != nil {
		if err := hookOptions.cb(ctxUpdate); err != nil {
			return err
		}
	}

	newFlow, err := e.d.SettingsHandler().NewFlow(w, r, i, ctxUpdate.Flow.Type)
	if err != nil {
		return err
	}

	ctxUpdate.Flow.UI = newFlow.UI
	ctxUpdate.Flow.UI.ResetMessages()
	ctxUpdate.Flow.UI.AddMessage(node.DefaultGroup, text.NewInfoSelfServiceSettingsUpdateSuccess())
	if err := e.d.SettingsFlowPersister().UpdateSettingsFlow(r.Context(), ctxUpdate.Flow); err != nil {
		return err
	}

	for k, executor := range e.d.PostSettingsPostPersistHooks(r.Context(), settingsType) {
		if err := executor.ExecuteSettingsPostPersistHook(w, r, ctxUpdate.Flow, i); err != nil {
			if errors.Is(err, ErrHookAbortFlow) {
				e.d.Logger().
					WithRequest(r).
					WithField("executor", fmt.Sprintf("%T", executor)).
					WithField("executor_position", k).
					WithField("executors", PostHookPostPersistExecutorNames(e.d.PostSettingsPostPersistHooks(r.Context(), settingsType))).
					WithField("identity_id", i.ID).
					WithField("flow_method", settingsType).
					Debug("A ExecuteSettingsPostPersistHook hook aborted early.")
				return nil
			}
			return e.handleSettingsError(w, r, settingsType, ctxUpdate.Flow, i, err)
		}

		e.d.Logger().WithRequest(r).
			WithField("executor", fmt.Sprintf("%T", executor)).
			WithField("executor_position", k).
			WithField("executors", PostHookPostPersistExecutorNames(e.d.PostSettingsPostPersistHooks(r.Context(), settingsType))).
			WithField("identity_id", i.ID).
			WithField("flow_method", settingsType).
			Debug("ExecuteSettingsPostPersistHook completed successfully.")
	}

	e.d.Logger().
		WithRequest(r).
		WithField("identity_id", i.ID).
		WithField("flow_method", settingsType).
		Debug("Completed all PostSettingsPrePersistHooks and PostSettingsPostPersistHooks.")

	if ctxUpdate.Flow.Type == flow.TypeAPI {
		updatedFlow, err := e.d.SettingsFlowPersister().GetSettingsFlow(r.Context(), ctxUpdate.Flow.ID)
		if err != nil {
			return err
		}

		e.d.Writer().Write(w, r, updatedFlow)
		return nil
	}

	if err := e.d.SessionManager().IssueCookie(r.Context(), w, r, ctxUpdate.Session); err != nil {
		return errors.WithStack(err)
	}

	if x.IsJSONRequest(r) {
		updatedFlow, err := e.d.SettingsFlowPersister().GetSettingsFlow(r.Context(), ctxUpdate.Flow.ID)
		if err != nil {
			return err
		}

		e.d.Writer().Write(w, r, updatedFlow)
		return nil
	}

	x.ContentNegotiationRedirection(w, r, i.CopyWithoutCredentials(), e.d.Writer(), returnTo.String())
	return nil
}

func (e *HookExecutor) PreSettingsHook(w http.ResponseWriter, r *http.Request, a *Flow) error {
	for _, executor := range e.d.PreSettingsHooks(r.Context()) {
		if err := executor.ExecuteSettingsPreHook(w, r, a); err != nil {
			return err
		}
	}

	return nil
}
