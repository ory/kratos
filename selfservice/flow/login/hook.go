// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/trace"

	"github.com/ory/kratos/x/events"

	"github.com/pkg/errors"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hydra"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/sessiontokenexchange"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/otelx"
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
		hydra.Provider
		session.ManagementProvider
		session.PersistenceProvider
		x.CSRFTokenGeneratorProvider
		x.WriterProvider
		x.LoggingProvider
		x.TracingProvider
		sessiontokenexchange.PersistenceProvider

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

func (e *HookExecutor) PostLoginHook(
	w http.ResponseWriter,
	r *http.Request,
	g node.UiNodeGroup,
	a *Flow,
	i *identity.Identity,
	s *session.Session,
	provider string,
) (err error) {
	ctx := r.Context()
	ctx, span := e.d.Tracer(ctx).Tracer().Start(ctx, "HookExecutor.PostLoginHook")
	r = r.WithContext(ctx)
	defer otelx.End(span, &err)

	if err := s.Activate(r, i, e.d.Config(), time.Now().UTC()); err != nil {
		return err
	}

	c := e.d.Config()
	// Verify the redirect URL before we do any other processing.
	returnTo, err := x.SecureRedirectTo(r,
		c.SelfServiceBrowserDefaultReturnTo(r.Context()),
		x.SecureRedirectReturnTo(a.ReturnTo),
		x.SecureRedirectUseSourceURL(a.RequestURL),
		x.SecureRedirectAllowURLs(c.SelfServiceBrowserAllowedReturnToDomains(r.Context())),
		x.SecureRedirectAllowSelfServiceURLs(c.SelfPublicURL(r.Context())),
		x.SecureRedirectOverrideDefaultReturnTo(c.SelfServiceFlowLoginReturnTo(r.Context(), a.Active.String())),
	)

	if err != nil {
		return err
	}

	classified := s
	s = s.Declassified()

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

		trace.SpanFromContext(r.Context()).AddEvent(events.NewLoginSucceeded(r.Context(), &events.LoginSucceededOpts{
			SessionID:    s.ID,
			IdentityID:   i.ID,
			FlowType:     string(a.Type),
			RequestedAAL: string(a.RequestedAAL),
			IsRefresh:    a.Refresh,
			Method:       a.Active.String(),
			SSOProvider:  provider,
		}))
		if handled, err := e.d.SessionManager().MaybeRedirectAPICodeFlow(w, r, a, s.ID, g); err != nil {
			return errors.WithStack(err)
		} else if handled {
			return nil
		}

		response := &APIFlowResponse{Session: s, Token: s.Token}
		if required, _ := e.requiresAAL2(r, classified, a); required {
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

	trace.SpanFromContext(r.Context()).AddEvent(events.NewLoginSucceeded(r.Context(), &events.LoginSucceededOpts{
		SessionID:  s.ID,
		IdentityID: i.ID, FlowType: string(a.Type), RequestedAAL: string(a.RequestedAAL), IsRefresh: a.Refresh, Method: a.Active.String(),
		SSOProvider: provider,
	}))

	if x.IsJSONRequest(r) {
		// Browser flows rely on cookies. Adding tokens in the mix will confuse consumers.
		s.Token = ""

		// If Kratos is used as a Hydra login provider, we need to redirect back to Hydra by returning a 422 status
		// with the post login challenge URL as the body.
		if a.OAuth2LoginChallenge != "" {
			postChallengeURL, err := e.d.Hydra().AcceptLoginRequest(r.Context(),
				hydra.AcceptLoginRequestParams{
					LoginChallenge:        string(a.OAuth2LoginChallenge),
					IdentityID:            i.ID.String(),
					SessionID:             s.ID.String(),
					AuthenticationMethods: s.AMR,
				})
			if err != nil {
				return err
			}
			e.d.Writer().WriteError(w, r, flow.NewBrowserLocationChangeRequiredError(postChallengeURL))
			return nil
		}

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
	if a.OAuth2LoginChallenge != "" {
		rt, err := e.d.Hydra().AcceptLoginRequest(r.Context(),
			hydra.AcceptLoginRequestParams{
				LoginChallenge:        string(a.OAuth2LoginChallenge),
				IdentityID:            i.ID.String(),
				SessionID:             s.ID.String(),
				AuthenticationMethods: s.AMR,
			})
		if err != nil {
			return err
		}
		finalReturnTo = rt
	}

	x.ContentNegotiationRedirection(w, r, s, e.d.Writer(), finalReturnTo)
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
