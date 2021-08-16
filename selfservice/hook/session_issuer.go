// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hook

import (
	"github.com/ory/kratos/driver/config"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/pkg/errors"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

var (
	_ registration.PostHookPostPersistExecutor = new(SessionIssuer)
)

type (
	sessionIssuerDependencies interface {
		config.Provider
		session.ManagementProvider
		session.PersistenceProvider
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
	s.AuthenticatedAt = time.Now().UTC()
	if err := e.r.SessionPersister().UpsertSession(r.Context(), s); err != nil {
		return err
	}

	if a.Type == flow.TypeAPI {
		e.r.Writer().Write(w, r, &registration.APIFlowResponse{
			Session:  s,
			Token:    s.Token,
			Identity: s.Identity,
		})
		return errors.WithStack(registration.ErrHookAbortFlow)
	}

	isWebView, err := flow.IsWebViewFlow(r.Context(), e.r.Config(), a)
	if err != nil {
		return err
	}
	if isWebView {
		response := &registration.APIFlowResponse{Session: s.Declassify(), Token: s.Token}

		w.Header().Set("Content-Type", "application/json")
		returnTo, err := url.Parse(a.ReturnTo)
		if err != nil {
			return err
		}
		returnTo.Path = path.Join(returnTo.Path, "success")
		query := returnTo.Query()
		query.Set("session_token", s.Token)
		returnTo.RawQuery = query.Encode()
		w.Header().Set("Location", returnTo.String())
		e.r.Writer().WriteCode(w, r, http.StatusSeeOther, response)

		return errors.WithStack(registration.ErrHookAbortFlow)
	}

	// cookie is issued both for browser and for SPA flows
	if err := e.r.SessionManager().IssueCookie(r.Context(), w, r, s); err != nil {
		return err
	}

	// SPA flows additionally send the session
	if x.IsJSONRequest(r) {
		e.r.Writer().Write(w, r, &registration.APIFlowResponse{
			Session:  s,
			Identity: s.Identity,
		})
		return errors.WithStack(registration.ErrHookAbortFlow)
	}

	return nil
}
