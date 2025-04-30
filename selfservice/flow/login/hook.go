// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ory/kratos/x/nosurfx"
	"github.com/ory/kratos/x/redir"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hydra"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/sessiontokenexchange"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/ui/container"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/events"
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
		PreLoginHooks(ctx context.Context) ([]PreHookExecutor, error)
		PostLoginHooks(ctx context.Context, credentialsType identity.CredentialsType) ([]PostHookExecutor, error)
	}
)

type (
	executorDependencies interface {
		config.Provider
		hydra.Provider
		identity.PrivilegedPoolProvider
		identity.ManagementProvider
		session.ManagementProvider
		session.PersistenceProvider
		nosurfx.CSRFTokenGeneratorProvider
		x.WriterProvider
		x.LoggingProvider
		x.TracingProvider
		sessiontokenexchange.PersistenceProvider
		HandlerProvider

		FlowPersistenceProvider
		HooksProvider
		StrategyProvider
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

func (e *HookExecutor) checkAAL(ctx context.Context, s *session.Session, a *Flow) error {
	err := e.d.SessionManager().DoesSessionSatisfy(ctx, s, e.d.Config().SessionWhoAmIAAL(ctx))
	if err == nil {
		return nil
	}

	if aalErr := new(session.ErrAALNotSatisfied); errors.As(err, &aalErr) {
		if a != nil && aalErr.PassReturnToAndLoginChallengeParameters(a.RequestURL) != nil {
			_ = aalErr.WithDetail("pass_request_params_error", "failed to pass request parameters to aalErr.RedirectTo")
		}
		return aalErr
	}

	return err
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
	f *Flow,
	i *identity.Identity,
	s *session.Session,
	provider string,
) (err error) {
	ctx := r.Context()
	ctx, span := e.d.Tracer(ctx).Tracer().Start(ctx, "HookExecutor.PostLoginHook")
	r = r.WithContext(ctx)
	defer otelx.End(span, &err)

	// We need to set the identity here because we check the available AAL in maybeLinkCredentials.
	s.IdentityID = i.ID
	s.Identity = i

	if err := e.maybeLinkCredentials(ctx, s, i, f); err != nil {
		return err
	}

	if err := e.d.SessionManager().ActivateSession(r, s, i, time.Now().UTC()); err != nil {
		return err
	}

	c := e.d.Config()
	// Verify the redirect URL before we do any other processing.
	returnTo, err := redir.SecureRedirectTo(r,
		c.SelfServiceBrowserDefaultReturnTo(ctx),
		redir.SecureRedirectReturnTo(f.ReturnTo),
		redir.SecureRedirectUseSourceURL(f.RequestURL),
		redir.SecureRedirectAllowURLs(c.SelfServiceBrowserAllowedReturnToDomains(ctx)),
		redir.SecureRedirectAllowSelfServiceURLs(c.SelfPublicURL(ctx)),
		redir.SecureRedirectOverrideDefaultReturnTo(c.SelfServiceFlowLoginReturnTo(ctx, f.Active.String())),
	)
	if err != nil {
		return err
	}
	span.SetAttributes(otelx.StringAttrs(map[string]string{
		"return_to":       returnTo.String(),
		"flow_type":       string(flow.TypeBrowser),
		"redirect_reason": "login successful",
	})...)

	if f.Type == flow.TypeBrowser && x.IsJSONRequest(r) {
		f.AddContinueWith(flow.NewContinueWithRedirectBrowserTo(returnTo.String()))
	}

	classified := s
	s = s.Declassified()

	e.d.Logger().
		WithRequest(r).
		WithField("identity_id", i.ID).
		WithField("flow_method", f.Active).
		Debug("Running ExecuteLoginPostHook.")
	hooks, err := e.d.PostLoginHooks(ctx, f.Active)
	if err != nil {
		return err
	}
	for k, executor := range hooks {
		if err := executor.ExecuteLoginPostHook(w, r, g, f, s); err != nil {
			if errors.Is(err, ErrHookAbortFlow) {
				e.d.Logger().
					WithRequest(r).
					WithField("executor", fmt.Sprintf("%T", executor)).
					WithField("executor_position", k).
					WithField("executors", PostHookExecutorNames(hooks)).
					WithField("identity_id", i.ID).
					WithField("flow_method", f.Active).
					Debug("A ExecuteLoginPostHook hook aborted early.")

				span.SetAttributes(attribute.String("redirect_reason", "aborted by hook"), attribute.String("executor", fmt.Sprintf("%T", executor)))

				return nil
			}
			return e.handleLoginError(w, r, g, f, i, err)
		}

		e.d.Logger().
			WithRequest(r).
			WithField("executor", fmt.Sprintf("%T", executor)).
			WithField("executor_position", k).
			WithField("executors", PostHookExecutorNames(hooks)).
			WithField("identity_id", i.ID).
			WithField("flow_method", f.Active).
			Debug("ExecuteLoginPostHook completed successfully.")
	}

	if f.Type == flow.TypeAPI {
		span.SetAttributes(attribute.String("flow_type", string(flow.TypeAPI)))
		if err := e.d.SessionPersister().UpsertSession(ctx, s); err != nil {
			return errors.WithStack(err)
		}
		e.d.Audit().
			WithRequest(r).
			WithField("session_id", s.ID).
			WithField("identity_id", i.ID).
			Info("Identity authenticated successfully and was issued an Ory Kratos Session Token.")

		span.AddEvent(events.NewLoginSucceeded(ctx, &events.LoginSucceededOpts{
			SessionID:    s.ID,
			IdentityID:   i.ID,
			FlowID:       f.ID,
			FlowType:     string(f.Type),
			RequestedAAL: string(f.RequestedAAL),
			IsRefresh:    f.Refresh,
			Method:       f.Active.String(),
			SSOProvider:  provider,
		}))
		if f.IDToken != "" {
			// We don't want to redirect with the code, if the flow was submitted with an ID token.
			// This is the case for Sign in with native Apple SDK or Google SDK.
		} else if handled, err := e.d.SessionManager().MaybeRedirectAPICodeFlow(w, r, f, s.ID, g); err != nil {
			return errors.WithStack(err)
		} else if handled {
			return nil
		}

		response := &APIFlowResponse{
			Session:      s,
			Token:        s.Token,
			ContinueWith: f.ContinueWith(),
		}
		if e.checkAAL(ctx, classified, f) != nil {
			// If AAL is not satisfied, we omit the identity to preserve the user's privacy in case of a phishing attack.
			response.Session.Identity = nil
		}

		e.d.Writer().Write(w, r, response)
		return nil
	}

	if err := e.d.SessionManager().UpsertAndIssueCookie(ctx, w, r, s); err != nil {
		return errors.WithStack(err)
	}

	e.d.Audit().
		WithRequest(r).
		WithField("identity_id", i.ID).
		WithField("session_id", s.ID).
		Info("Identity authenticated successfully and was issued an Ory Kratos Session Cookie.")

	span.AddEvent(events.NewLoginSucceeded(ctx, &events.LoginSucceededOpts{
		SessionID:  s.ID,
		FlowID:     f.ID,
		IdentityID: i.ID, FlowType: string(f.Type), RequestedAAL: string(f.RequestedAAL), IsRefresh: f.Refresh, Method: f.Active.String(),
		SSOProvider: provider,
	}))

	if x.IsJSONRequest(r) {
		span.SetAttributes(attribute.String("flow_type", "spa"))

		// Browser flows rely on cookies. Adding tokens in the mix will confuse consumers.
		s.Token = ""

		// If we detect that whoami would require a higher AAL, we redirect!
		if err := e.checkAAL(ctx, classified, f); err != nil {
			if aalErr := new(session.ErrAALNotSatisfied); errors.As(err, &aalErr) {
				if data, _ := flow.DuplicateCredentials(f); data == nil {
					span.SetAttributes(attribute.String("return_to", aalErr.RedirectTo), attribute.String("redirect_reason", "requires aal2"))
					e.d.Writer().WriteError(w, r, flow.NewBrowserLocationChangeRequiredError(aalErr.RedirectTo))
					return nil
				}

				// Special case: If we are in a flow that wants to link credentials, we create a
				// new login flow here that asks for the require AAL, but also copies over the
				// internal context and the organization ID.
				r.URL, err = url.Parse(aalErr.RedirectTo)
				if err != nil {
					return errors.WithStack(err)
				}
				newFlow, _, err := e.d.LoginHandler().NewLoginFlow(w, r, flow.TypeBrowser,
					WithInternalContext(f.InternalContext),
					WithOrganizationID(f.OrganizationID),
				)
				if err != nil {
					return errors.WithStack(err)
				}

				x.SendFlowCompletedAsRedirectOrJSON(w, r, e.d.Writer(), newFlow, newFlow.AppendTo(e.d.Config().SelfServiceFlowLoginUI(ctx)).String())
				return nil
			}
			return err
		}

		// If Kratos is used as a Hydra login provider, we need to redirect back to Hydra by returning a 422 status
		// with the post login challenge URL as the body.
		if f.OAuth2LoginChallenge != "" {
			postChallengeURL, err := e.d.Hydra().AcceptLoginRequest(ctx,
				hydra.AcceptLoginRequestParams{
					LoginChallenge:        string(f.OAuth2LoginChallenge),
					IdentityID:            i.ID.String(),
					SessionID:             s.ID.String(),
					AuthenticationMethods: s.AMR,
				})
			if err != nil {
				return err
			}
			span.SetAttributes(attribute.String("return_to", postChallengeURL), attribute.String("redirect_reason", "oauth2 login challenge"))
			e.d.Writer().WriteError(w, r, flow.NewBrowserLocationChangeRequiredError(postChallengeURL))
			return nil
		}

		response := &APIFlowResponse{
			Session:      s,
			ContinueWith: f.ContinueWith(),
		}
		e.d.Writer().Write(w, r, response)
		return nil
	}

	// If we detect that whoami would require a higher AAL, we redirect!
	if err := e.checkAAL(ctx, classified, f); err != nil {
		if aalErr := new(session.ErrAALNotSatisfied); errors.As(err, &aalErr) {
			if data, _ := flow.DuplicateCredentials(f); data == nil {
				http.Redirect(w, r, aalErr.RedirectTo, http.StatusSeeOther)
				return nil
			}

			// Special case: If we are in a flow that wants to link credentials, we create a
			// new login flow here that asks for the require AAL, but also copies over the
			// internal context and the organization ID.
			r.URL, err = url.Parse(aalErr.RedirectTo)
			if err != nil {
				return errors.WithStack(err)
			}
			newFlow, _, err := e.d.LoginHandler().NewLoginFlow(w, r, flow.TypeBrowser,
				WithInternalContext(f.InternalContext),
				WithOrganizationID(f.OrganizationID),
			)
			if err != nil {
				return errors.WithStack(err)
			}

			x.SendFlowCompletedAsRedirectOrJSON(w, r, e.d.Writer(), newFlow, newFlow.AppendTo(e.d.Config().SelfServiceFlowLoginUI(ctx)).String())
			return nil
		}
		return errors.WithStack(err)
	}

	finalReturnTo := returnTo.String()
	if f.OAuth2LoginChallenge != "" {
		rt, err := e.d.Hydra().AcceptLoginRequest(ctx,
			hydra.AcceptLoginRequestParams{
				LoginChallenge:        string(f.OAuth2LoginChallenge),
				IdentityID:            i.ID.String(),
				SessionID:             s.ID.String(),
				AuthenticationMethods: s.AMR,
			})
		if err != nil {
			return err
		}
		finalReturnTo = rt
		span.SetAttributes(attribute.String("return_to", rt), attribute.String("redirect_reason", "oauth2 login challenge"))
	} else if f.ReturnToVerification != "" {
		finalReturnTo = f.ReturnToVerification
		span.SetAttributes(attribute.String("redirect_reason", "verification requested"))
	}

	redir.ContentNegotiationRedirection(w, r, s, e.d.Writer(), finalReturnTo)
	return nil
}

func (e *HookExecutor) PreLoginHook(w http.ResponseWriter, r *http.Request, a *Flow) error {
	hooks, err := e.d.PreLoginHooks(r.Context())
	if err != nil {
		return err
	}
	for _, h := range hooks {
		if err := h.ExecuteLoginPreHook(w, r, a); err != nil {
			return err
		}
	}

	return nil
}

// maybeLinkCredentials links the identity with the credentials of the inner context of the login flow.
func (e *HookExecutor) maybeLinkCredentials(ctx context.Context, sess *session.Session, ident *identity.Identity, loginFlow *Flow) (err error) {
	ctx, span := e.d.Tracer(ctx).Tracer().Start(ctx, "HookExecutor.PostLoginHook.maybeLinkCredentials")
	defer otelx.End(span, &err)

	if e.checkAAL(ctx, sess, loginFlow) != nil {
		// we don't yet want to link credentials because the required AAL is not satisfied
		return nil
	}

	lc, err := flow.DuplicateCredentials(loginFlow)
	if err != nil {
		return err
	} else if lc == nil {
		return nil
	}

	if err = e.checkDuplicateCredentialsIdentifierMatch(ctx, ident, lc.DuplicateIdentifier); err != nil {
		return err
	}
	strategy, err := e.d.AllLoginStrategies().Strategy(lc.CredentialsType)
	if err != nil {
		return err
	}

	linkableStrategy, ok := strategy.(LinkableStrategy)
	if !ok {
		// This should never happen because we check for this in the registration flow.
		return errors.Errorf("strategy is not linkable: %T", linkableStrategy)
	}

	if err := linkableStrategy.Link(ctx, ident, lc.CredentialsConfig); err != nil {
		return err
	}

	if err = linkableStrategy.CompletedLogin(sess, lc); err != nil {
		return err
	}

	return nil
}

func (e *HookExecutor) checkDuplicateCredentialsIdentifierMatch(ctx context.Context, i *identity.Identity, match string) error {
	if len(i.Credentials) == 0 {
		if err := e.d.PrivilegedIdentityPool().HydrateIdentityAssociations(ctx, i, identity.ExpandCredentials); err != nil {
			return err
		}
	}

	for _, credentials := range i.Credentials {
		for _, identifier := range credentials.Identifiers {
			if identifier == match {
				return nil
			}
		}
	}
	return schema.NewLinkedCredentialsDoNotMatch()
}
