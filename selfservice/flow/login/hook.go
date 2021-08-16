// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hydra"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/otelx/semconv"
)

type (
	PreHookExecutor interface {
		ExecuteLoginPreHook(w http.ResponseWriter, r *http.Request, a *Flow) error
	}

	PostHookExecutor interface {
		ExecuteLoginPostHook(w http.ResponseWriter, r *http.Request, g node.UiNodeGroup, a *Flow, s *session.Session) error
	}

	HooksProvider interface {
		PreLoginHooks(ctx context.Context) []PreHookExecutor
		PostLoginHooks(ctx context.Context, credentialsType identity.CredentialsType) []PostHookExecutor
	}
)

type (
	executorDependencies interface {
		config.Provider
		hydra.HydraProvider
		session.ManagementProvider
		session.PersistenceProvider
		x.CSRFTokenGeneratorProvider
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

func (e *HookExecutor) requiresAAL2(r *http.Request, s *session.Session, a *Flow) (bool, error) {
	err := e.d.SessionManager().DoesSessionSatisfy(r, s, e.d.Config().SessionWhoAmIAAL(r.Context()))

	if aalErr := new(session.ErrAALNotSatisfied); errors.As(err, &aalErr) {
		if aalErr.PassReturnToAndLoginChallengeParameters(a.RequestURL) != nil {
			_ = aalErr.WithDetail("pass_request_params_error", "failed to pass request parameters to aalErr.RedirectTo")
		}
		return true, aalErr
	} else if err != nil {
		return true, errors.WithStack(err)
	}

	return false, nil
}

func (e *HookExecutor) handleLoginError(_ http.ResponseWriter, r *http.Request, g node.UiNodeGroup, f *Flow, i *identity.Identity, flowError error) error {
	if f != nil {
		if i != nil {
			cont, err := container.NewFromStruct("", g, i.Traits, "traits")
			if err != nil {
				e.d.Logger().WithError(err).Warn("could not update flow UI")
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

func (e *HookExecutor) PostLoginHook(w http.ResponseWriter, r *http.Request, g node.UiNodeGroup, a *Flow, i *identity.Identity, s *session.Session) error {
	if err := s.Activate(r, i, e.d.Config(), time.Now().UTC()); err != nil {
		return err
	}

	// Verify the redirect URL before we do any other processing.
	c := e.d.Config()
	returnTo, err := x.SecureRedirectTo(r, c.SelfServiceBrowserDefaultReturnTo(r.Context()),
		x.SecureRedirectReturnTo(a.ReturnTo),
		x.SecureRedirectUseSourceURL(a.RequestURL),
		x.SecureRedirectAllowURLs(c.SelfServiceBrowserAllowedReturnToDomains(r.Context())),
		x.SecureRedirectAllowSelfServiceURLs(c.SelfPublicURL(r.Context())),
		x.SecureRedirectOverrideDefaultReturnTo(e.d.Config().SelfServiceFlowLoginReturnTo(r.Context(), a.Active.String())),
	)
	if err != nil {
		return err
	}

	s = s.Declassify()

	e.d.Logger().
		WithRequest(r).
		WithField("identity_id", i.ID).
		WithField("flow_method", a.Active).
		Debug("Running ExecuteLoginPostHook.")
	for k, executor := range e.d.PostLoginHooks(r.Context(), a.Active) {
		if err := executor.ExecuteLoginPostHook(w, r, g, a, s); err != nil {
			if errors.Is(err, ErrHookAbortFlow) {
				e.d.Logger().
					WithRequest(r).
					WithField("executor", fmt.Sprintf("%T", executor)).
					WithField("executor_position", k).
					WithField("executors", PostHookExecutorNames(e.d.PostLoginHooks(r.Context(), a.Active))).
					WithField("identity_id", i.ID).
					WithField("flow_method", a.Active).
					Debug("A ExecuteLoginPostHook hook aborted early.")
				return nil
			}
			return e.handleLoginError(w, r, g, a, i, err)
		}

		e.d.Logger().
			WithRequest(r).
			WithField("executor", fmt.Sprintf("%T", executor)).
			WithField("executor_position", k).
			WithField("executors", PostHookExecutorNames(e.d.PostLoginHooks(r.Context(), a.Active))).
			WithField("identity_id", i.ID).
			WithField("flow_method", a.Active).
			Debug("ExecuteLoginPostHook completed successfully.")
	}

	if a.Type == flow.TypeAPI {
		if err := e.d.SessionPersister().UpsertSession(r.Context(), s); err != nil {
			return errors.WithStack(err)
		}
		e.d.Audit().
			WithRequest(r).
			WithField("session_id", s.ID).
			WithField("identity_id", i.ID).
			Info("Identity authenticated successfully and was issued an Ory Kratos Session Token.")
		trace.SpanFromContext(r.Context()).AddEvent(
			semconv.EventSessionIssued,
			trace.WithAttributes(
				attribute.String(semconv.AttrIdentityID, i.ID.String()),
				attribute.String(semconv.AttrNID, i.NID.String()),
				attribute.String("flow", string(flow.TypeAPI)),
			),
		)

		response := &APIFlowResponse{Session: s, Token: s.Token}
		if required, _ := e.requiresAAL2(r, s, a); required {
			// If AAL is not satisfied, we omit the identity to preserve the user's privacy in case of a phishing attack.
			response.Session.Identity = nil
		}

		e.d.Writer().Write(w, r, response)
		return nil
	}

	if err := e.d.SessionManager().UpsertAndIssueCookie(r.Context(), w, r, s); err != nil {
		return errors.WithStack(err)
	}

	e.d.Audit().
		WithRequest(r).
		WithField("identity_id", i.ID).
		WithField("session_id", s.ID).
		Info("Identity authenticated successfully and was issued an Ory Kratos Session Cookie.")
	trace.SpanFromContext(r.Context()).AddEvent(
		semconv.EventSessionIssued,
		trace.WithAttributes(
			attribute.String(semconv.AttrIdentityID, i.ID.String()),
			attribute.String(semconv.AttrNID, i.NID.String()),
			attribute.String("flow", string(flow.TypeBrowser)),
		),
	)

	if x.IsJSONRequest(r) {
		// Browser flows rely on cookies. Adding tokens in the mix will confuse consumers.
		s.Token = ""

		response := &APIFlowResponse{Session: s}
		if required, _ := e.requiresAAL2(r, s, a); required {
			// If AAL is not satisfied, we omit the identity to preserve the user's privacy in case of a phishing attack.
			response.Session.Identity = nil
		}
		e.d.Writer().Write(w, r, response)
		return nil
	}

	// If we detect that whoami would require a higher AAL, we redirect!
	if _, err := e.requiresAAL2(r, s, a); err != nil {
		if aalErr := new(session.ErrAALNotSatisfied); errors.As(err, &aalErr) {
			http.Redirect(w, r, aalErr.RedirectTo, http.StatusSeeOther)
			return nil
		}
		return errors.WithStack(err)
	}

	finalReturnTo := returnTo.String()
	if a.OAuth2LoginChallenge.Valid {
		rt, err := e.d.Hydra().AcceptLoginRequest(r.Context(), a.OAuth2LoginChallenge.UUID, i.ID.String(), s.AMR)
		if err != nil {
			return err
		}
		finalReturnTo = rt
	}

	isWebView, err := flow.IsWebViewFlow(r.Context(), e.d.Config(), a)
	if err != nil {
		return err
	}
	if isWebView {
		response := &APIFlowResponse{Session: s.Declassify(), Token: s.Token}
		required, err := e.requiresAAL2(r, s, a)
		if err != nil {
			return err
		}
		if required {
			// If AAL is not satisfied, we omit the identity to preserve the user's privacy in case of a phishing attack.
			response.Session.Identity = nil
		}
		w.Header().Set("Content-Type", "application/json")
		returnTo.Path = path.Join(returnTo.Path, "success")
		query := returnTo.Query()
		query.Set("session_token", s.Token)
		returnTo.RawQuery = query.Encode()
		w.Header().Set("Location", returnTo.String())
		e.d.Writer().WriteCode(w, r, http.StatusSeeOther, response)
	} else {
		x.ContentNegotiationRedirection(w, r, s.Declassify(), e.d.Writer(), finalReturnTo)
	}
	return nil
}

func (e *HookExecutor) PreLoginHook(w http.ResponseWriter, r *http.Request, a *Flow) error {
	for _, executor := range e.d.PreLoginHooks(r.Context()) {
		if err := executor.ExecuteLoginPreHook(w, r, a); err != nil {
			return err
		}
	}

	return nil
}
