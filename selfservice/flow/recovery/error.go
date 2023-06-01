// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"net/http"
	"net/url"

	"go.opentelemetry.io/otel/trace"

	"github.com/ory/kratos/x/events"

	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/ui/node"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

var (
	ErrHookAbortFlow   = errors.New("aborted recovery hook execution")
	ErrAlreadyLoggedIn = herodot.ErrBadRequest.WithID(text.ErrIDAlreadyLoggedIn).WithReason("A valid session was detected and thus recovery is not possible.")
)

type (
	errorHandlerDependencies interface {
		errorx.ManagementProvider
		x.WriterProvider
		x.LoggingProvider
		x.CSRFTokenGeneratorProvider
		config.Provider
		StrategyProvider

		FlowPersistenceProvider
	}

	ErrorHandlerProvider interface {
		RecoveryFlowErrorHandler() *ErrorHandler
	}

	ErrorHandler struct {
		d errorHandlerDependencies
	}
)

func NewErrorHandler(d errorHandlerDependencies) *ErrorHandler {
	return &ErrorHandler{d: d}
}

func (s *ErrorHandler) WriteFlowError(
	w http.ResponseWriter,
	r *http.Request,
	f *Flow,
	group node.UiNodeGroup,
	err error,
) {
	s.d.Audit().
		WithError(err).
		WithRequest(r).
		WithField("recovery_flow", f).
		Info("Encountered self-service recovery error.")

	if f == nil {
		trace.SpanFromContext(r.Context()).AddEvent(events.NewRecoveryFailed(r.Context(), "", ""))
		s.forward(w, r, nil, err)
		return
	}

	trace.SpanFromContext(r.Context()).AddEvent(events.NewRecoveryFailed(r.Context(), string(f.Type), f.Active.String()))

	if e := new(flow.ExpiredError); errors.As(err, &e) {
		strategy, err := s.d.RecoveryStrategies(r.Context()).Strategy(f.Active.String())
		if err != nil {
			strategy, err = s.d.GetActiveRecoveryStrategy(r.Context())
			// Can't retry the recovery if no strategy has been set
			if err != nil {
				s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
				return
			}
		}
		// create new flow because the old one is not valid
		a, err := FromOldFlow(s.d.Config(), s.d.Config().SelfServiceFlowRecoveryRequestLifespan(r.Context()), s.d.GenerateCSRFToken(r), r, strategy, *f)
		if err != nil {
			// failed to create a new session and redirect to it, handle that error as a new one
			s.WriteFlowError(w, r, f, group, err)
			return
		}

		a.UI.Messages.Add(text.NewErrorValidationRecoveryFlowExpired(e.ExpiredAt))
		if err := s.d.RecoveryFlowPersister().CreateRecoveryFlow(r.Context(), a); err != nil {
			s.forward(w, r, a, err)
			return
		}

		// We need to use the new flow, as that flow will be a browser flow. Bug fix for:
		//
		// https://github.com/ory/kratos/issues/2049!!
		if a.Type == flow.TypeAPI || x.IsJSONRequest(r) {
			http.Redirect(w, r, urlx.CopyWithQuery(urlx.AppendPaths(s.d.Config().SelfPublicURL(r.Context()),
				RouteGetFlow), url.Values{"id": {a.ID.String()}}).String(), http.StatusSeeOther)
		} else {
			http.Redirect(w, r, a.AppendTo(s.d.Config().SelfServiceFlowRecoveryUI(r.Context())).String(), http.StatusSeeOther)
		}
		return
	}

	f.UI.ResetMessages()
	if err := f.UI.ParseError(group, err); err != nil {
		s.forward(w, r, f, err)
		return
	}

	f.Active = sqlxx.NullString(group)
	if err := s.d.RecoveryFlowPersister().UpdateRecoveryFlow(r.Context(), f); err != nil {
		s.forward(w, r, f, err)
		return
	}

	if f.Type == flow.TypeBrowser && !x.IsJSONRequest(r) {
		http.Redirect(w, r, f.AppendTo(s.d.Config().SelfServiceFlowRecoveryUI(r.Context())).String(), http.StatusSeeOther)
		return
	}

	updatedFlow, innerErr := s.d.RecoveryFlowPersister().GetRecoveryFlow(r.Context(), f.ID)
	if innerErr != nil {
		s.forward(w, r, updatedFlow, innerErr)
	}

	s.d.Writer().WriteCode(w, r, x.RecoverStatusCode(err, http.StatusBadRequest), updatedFlow)
}

func (s *ErrorHandler) forward(w http.ResponseWriter, r *http.Request, rr *Flow, err error) {
	if rr == nil {
		if x.IsJSONRequest(r) {
			s.d.Writer().WriteError(w, r, err)
			return
		}
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	if rr.Type == flow.TypeAPI || x.IsJSONRequest(r) {
		s.d.Writer().WriteErrorCode(w, r, x.RecoverStatusCode(err, http.StatusBadRequest), err)
	} else {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
	}
}
