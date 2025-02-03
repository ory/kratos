// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hook

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel/trace"

	"github.com/ory/kratos/hydra"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x/events"

	"github.com/pkg/errors"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/sessiontokenexchange"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"github.com/ory/x/otelx"
)

var _ registration.PostHookPostPersistExecutor = new(SessionIssuer)

type (
	sessionIssuerDependencies interface {
		session.ManagementProvider
		session.PersistenceProvider
		sessiontokenexchange.PersistenceProvider
		config.Provider
		x.WriterProvider
		hydra.Provider
	}
	SessionIssuerProvider interface {
		HookSessionIssuer() *SessionIssuer
	}
	SessionIssuer struct {
		r sessionIssuerDependencies
	}
)

func NewSessionIssuer(r sessionIssuerDependencies) *SessionIssuer {
	return &SessionIssuer{r: r}
}

func (e *SessionIssuer) ExecutePostRegistrationPostPersistHook(w http.ResponseWriter, r *http.Request, a *registration.Flow, s *session.Session) error {
	return otelx.WithSpan(r.Context(), "selfservice.hook.SessionIssuer.ExecutePostRegistrationPostPersistHook", func(ctx context.Context) error {
		return e.executePostRegistrationPostPersistHook(w, r.WithContext(ctx), a, s)
	})
}

func (e *SessionIssuer) executePostRegistrationPostPersistHook(w http.ResponseWriter, r *http.Request, a *registration.Flow, s *session.Session) error {
	if a.Type == flow.TypeAPI {
		// We don't want to redirect with the code, if the flow was submitted with an ID token.
		// This is the case for Sign in with native Apple SDK or Google SDK.
		if s.AuthenticatedVia(identity.CredentialsTypeOIDC) && a.IDToken == "" {
			if handled, err := e.r.SessionManager().MaybeRedirectAPICodeFlow(w, r, a, s.ID, node.OpenIDConnectGroup); err != nil {
				return errors.WithStack(err)
			} else if handled {
				return nil
			}
		}

		a.AddContinueWith(flow.NewContinueWithSetToken(s.Token))
		e.r.Writer().Write(w, r, &registration.APIFlowResponse{
			Session:      s,
			Token:        s.Token,
			Identity:     s.Identity,
			ContinueWith: a.ContinueWithItems,
		})

		trace.SpanFromContext(r.Context()).AddEvent(events.NewLoginSucceeded(r.Context(), &events.LoginSucceededOpts{
			SessionID:  s.ID,
			IdentityID: s.Identity.ID,
			FlowID:     a.ID,
			FlowType:   string(a.Type),
			Method:     a.Active.String(),
		}))

		return errors.WithStack(registration.ErrHookAbortFlow)
	}

	// cookie is issued both for browser and for SPA flows
	if err := e.r.SessionManager().IssueCookie(r.Context(), w, r, s); err != nil {
		return err
	}

	trace.SpanFromContext(r.Context()).AddEvent(events.NewLoginSucceeded(r.Context(), &events.LoginSucceededOpts{
		SessionID:  s.ID,
		IdentityID: s.Identity.ID,
		FlowID:     a.ID,
		FlowType:   string(a.Type),
		Method:     a.Active.String(),
	}))

	// SPA flows additionally send the session
	if x.IsJSONRequest(r) {
		if err := e.acceptLoginChallenge(r.Context(), a, s, s.Identity); err != nil {
			return err
		}
		e.r.Writer().Write(w, r, &registration.APIFlowResponse{
			Session:      s,
			Identity:     s.Identity,
			ContinueWith: a.ContinueWithItems,
		})
		return errors.WithStack(registration.ErrHookAbortFlow)
	}

	return nil
}

func (e *SessionIssuer) acceptLoginChallenge(ctx context.Context, registrationFlow *registration.Flow, s *session.Session, i *identity.Identity) error {
	// If Kratos is used as a Hydra login provider, we need to redirect back to Hydra by using the continue_with items
	// with the post login challenge URL as the body.
	// We only do this if the flow did not create a verification flow (e.g. verification is disabled or not active due to it being a code flow).
	// Since the session issuer hook must be the last hook in the flow, we can safely assume that the verification flow was already added (if it was)
	if registrationFlow.OAuth2LoginChallenge != "" && !willVerificationFollow(registrationFlow) {
		postChallengeURL, err := e.r.Hydra().AcceptLoginRequest(ctx,
			hydra.AcceptLoginRequestParams{
				LoginChallenge:        string(registrationFlow.OAuth2LoginChallenge),
				IdentityID:            i.ID.String(),
				SessionID:             s.ID.String(),
				AuthenticationMethods: s.AMR,
			})
		if err != nil {
			return err
		}
		cw := []flow.ContinueWith{}
		for _, i := range registrationFlow.ContinueWithItems {
			// Filter any continueWithRedirectBrowserTo items out of the list
			// We will add a new one at the end of the flow
			// as the OAuth2 login challenge should be the last step in the flow
			if i.GetAction() != string(flow.ContinueWithActionRedirectBrowserToString) {
				cw = append(cw, i)
			}
		}
		registrationFlow.ContinueWithItems = append(cw, flow.NewContinueWithRedirectBrowserTo(postChallengeURL))
	}
	return nil
}

// willVerificationFollow returns true if the flow's continue with items contain a verification UI.
func willVerificationFollow(f *registration.Flow) bool {
	for _, i := range f.ContinueWithItems {
		if i.GetAction() == string(flow.ContinueWithActionShowVerificationUIString) {
			return true
		}
	}
	return false
}
