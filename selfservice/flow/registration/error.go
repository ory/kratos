// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package registration

import (
	"net/http"

	"go.opentelemetry.io/otel/trace"

	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/sessiontokenexchange"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x/events"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/text"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/x"
)

var (
	ErrHookAbortFlow                  = errors.New("aborted registration hook execution")
	ErrAlreadyLoggedIn                = herodot.ErrBadRequest.WithID(text.ErrIDAlreadyLoggedIn).WithError("you are already logged in").WithReason("A valid session was detected and thus registration is not possible.")
	ErrRegistrationDisabled           = herodot.ErrBadRequest.WithID(text.ErrIDSelfServiceFlowDisabled).WithError("registration flow disabled").WithReason("Registration is not allowed because it was disabled.")
	ErrDuplicateCredentials           = herodot.ErrBadRequest.WithError(text.NewErrorValidationDuplicateCredentials().Text)
	ErrDuplicateCredentialsOnOIDCLink = herodot.ErrBadRequest.WithError(text.NewErrorValidationDuplicateCredentialsOnOIDCLink().Text)
)

type (
	errorHandlerDependencies interface {
		errorx.ManagementProvider
		x.WriterProvider
		x.LoggingProvider
		config.Provider

		sessiontokenexchange.PersistenceProvider
		FlowPersistenceProvider
		HandlerProvider
	}

	ErrorHandlerProvider interface{ RegistrationFlowErrorHandler() *ErrorHandler }

	ErrorHandler struct {
		d errorHandlerDependencies
	}
)

func NewErrorHandler(d errorHandlerDependencies) *ErrorHandler {
	return &ErrorHandler{d: d}
}

func (s *ErrorHandler) PrepareReplacementForExpiredFlow(w http.ResponseWriter, r *http.Request, f *Flow, err error) (*flow.ExpiredError, error) {
	e := new(flow.ExpiredError)
	if !errors.As(err, &e) {
		return nil, nil
	}
	// create new flow because the old one is not valid
	a, err := s.d.RegistrationHandler().FromOldFlow(w, r, *f)
	if err != nil {
		return nil, err
	}

	a.UI.Messages.Add(text.NewErrorValidationRegistrationFlowExpired(e.ExpiredAt))
	if err := s.d.RegistrationFlowPersister().UpdateRegistrationFlow(r.Context(), a); err != nil {
		return nil, err
	}

	return e.WithFlow(a), nil
}
func (s *ErrorHandler) WriteFlowError(
	w http.ResponseWriter,
	r *http.Request,
	f *Flow,
	group node.UiNodeGroup,
	err error,
) {

	if errors.Is(err, ErrDuplicateCredentials) {
		err = schema.NewDuplicateCredentialsError()
	}

	s.d.Audit().
		WithError(err).
		WithRequest(r).
		WithField("registration_flow", f).
		Info("Encountered self-service flow error.")

	if f == nil {
		trace.SpanFromContext(r.Context()).AddEvent(events.NewRegistrationFailed(r.Context(), "", ""))
		s.forward(w, r, nil, err)
		return
	}
	trace.SpanFromContext(r.Context()).AddEvent(events.NewRegistrationFailed(r.Context(), string(f.Type), f.Active.String()))

	if expired, inner := s.PrepareReplacementForExpiredFlow(w, r, f, err); inner != nil {
		s.forward(w, r, f, err)
		return
	} else if expired != nil {
		if f.Type == flow.TypeAPI || x.IsJSONRequest(r) {
			s.d.Writer().WriteError(w, r, expired)
		} else {
			http.Redirect(w, r, expired.GetFlow().AppendTo(s.d.Config().SelfServiceFlowRegistrationUI(r.Context())).String(), http.StatusSeeOther)
		}
		return
	}

	f.UI.ResetMessages()
	if err := f.UI.ParseError(group, err); err != nil {
		s.forward(w, r, f, err)
		return
	}

	ds, err := s.d.Config().DefaultIdentityTraitsSchemaURL(r.Context())
	if err != nil {
		s.forward(w, r, f, err)
		return
	}

	if err := SortNodes(r.Context(), f.UI.Nodes, ds.String()); err != nil {
		s.forward(w, r, f, err)
		return
	}

	if err := s.d.RegistrationFlowPersister().UpdateRegistrationFlow(r.Context(), f); err != nil {
		s.forward(w, r, f, err)
		return
	}

	if f.Type == flow.TypeBrowser && !x.IsJSONRequest(r) {
		http.Redirect(w, r, f.AppendTo(s.d.Config().SelfServiceFlowRegistrationUI(r.Context())).String(), http.StatusFound)
		return
	}
	if _, hasCode, _ := s.d.SessionTokenExchangePersister().CodeForFlow(r.Context(), f.ID); group == node.OpenIDConnectGroup && f.Type == flow.TypeAPI && hasCode {
		http.Redirect(w, r, f.ReturnTo, http.StatusSeeOther)
		return
	}

	updatedFlow, innerErr := s.d.RegistrationFlowPersister().GetRegistrationFlow(r.Context(), f.ID)
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

	if rr.Type == flow.TypeAPI {
		s.d.Writer().WriteErrorCode(w, r, x.RecoverStatusCode(err, http.StatusBadRequest), err)
	} else {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
	}
}
