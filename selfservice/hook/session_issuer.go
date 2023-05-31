// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hook

import (
	"context"
	"net/http"
	"time"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/ui/node"

	"github.com/pkg/errors"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/sessiontokenexchange"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"github.com/ory/x/otelx"
)

var (
	_ registration.PostHookPostPersistExecutor = new(SessionIssuer)
)

type (
	sessionIssuerDependencies interface {
		session.ManagementProvider
		session.PersistenceProvider
		sessiontokenexchange.PersistenceProvider
		config.Provider
		x.WriterProvider
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
	s.AuthenticatedAt = time.Now().UTC()
	if err := e.r.SessionPersister().UpsertSession(r.Context(), s); err != nil {
		return err
	}

	if a.Type == flow.TypeAPI {
		if s.AuthenticatedVia(identity.CredentialsTypeOIDC) {
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
		return errors.WithStack(registration.ErrHookAbortFlow)
	}

	// cookie is issued both for browser and for SPA flows
	if err := e.r.SessionManager().IssueCookie(r.Context(), w, r, s); err != nil {
		return err
	}

	// SPA flows additionally send the session
	if x.IsJSONRequest(r) {
		e.r.Writer().Write(w, r, &registration.APIFlowResponse{
			Session:      s,
			Identity:     s.Identity,
			ContinueWith: a.ContinueWithItems,
		})
		return errors.WithStack(registration.ErrHookAbortFlow)
	}

	return nil
}
