// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"net/http"

	"github.com/gofrs/uuid"
	"go.opentelemetry.io/otel/attribute"

	"github.com/ory/x/otelx"

	"go.opentelemetry.io/otel/trace"

	"github.com/ory/kratos/hydra"
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
	ErrHookAbortFlow      = errors.New("aborted login hook execution")
	ErrAlreadyLoggedIn    = herodot.ErrBadRequest.WithID(text.ErrIDAlreadyLoggedIn).WithError("you are already logged in").WithReason("A valid session was detected and thus login is not possible. Did you forget to set `?refresh=true`?")
	ErrAddressNotVerified = herodot.ErrBadRequest.WithID(text.ErrIDAddressNotVerified).WithError("your email or phone address is not yet verified").WithReason("Your account's email or phone address are not verified yet. Please check your email or phone inbox or re-request verification.")

	// ErrSessionHasAALAlready is returned when one attempts to upgrade the AAL of an active session which already has that AAL.
	ErrSessionHasAALAlready = herodot.ErrUnauthorized.WithID(text.ErrIDSessionHasAALAlready).WithError("session has the requested authenticator assurance level already").WithReason("The session has the requested AAL already.")

	// ErrSessionRequiredForHigherAAL is returned when someone requests AAL2 or AAL3 even though no active session exists yet.
	ErrSessionRequiredForHigherAAL = herodot.ErrUnauthorized.WithID(text.ErrIDSessionRequiredForHigherAAL).WithError("aal2 and aal3 can only be requested if a session exists already").WithReason("You can not requested a higher AAL (AAL2/AAL3) without an active session.")
)

type (
	errorHandlerDependencies interface {
		errorx.ManagementProvider
		x.WriterProvider
		x.LoggingProvider
		x.TracingProvider
		config.Provider
		sessiontokenexchange.PersistenceProvider
		hydra.Provider

		FlowPersistenceProvider
		HandlerProvider
	}

	ErrorHandlerProvider interface{ LoginFlowErrorHandler() *ErrorHandler }

	ErrorHandler struct {
		d errorHandlerDependencies
	}
)

func NewFlowErrorHandler(d errorHandlerDependencies) *ErrorHandler {
	return &ErrorHandler{d: d}
}

func (s *ErrorHandler) PrepareReplacementForExpiredFlow(w http.ResponseWriter, r *http.Request, f *Flow, err error) (*flow.ExpiredError, error) {
	errExpired := new(flow.ExpiredError)
	if !errors.As(err, &errExpired) {
		return nil, nil
	}
	// create new flow because the old one is not valid
	newFlow, err := s.d.LoginHandler().FromOldFlow(w, r, *f)
	if err != nil {
		return nil, err
	}

	newFlow.UI.Messages.Add(text.NewErrorValidationLoginFlowExpired(errExpired.ExpiredAt))
	if err := s.d.LoginFlowPersister().UpdateLoginFlow(r.Context(), newFlow); err != nil {
		return nil, err
	}

	return errExpired.WithFlow(newFlow), nil
}

func (s *ErrorHandler) WriteFlowError(w http.ResponseWriter, r *http.Request, f *Flow, group node.UiNodeGroup, err error) {
	ctx, span := s.d.Tracer(r.Context()).Tracer().Start(r.Context(), "selfservice.flow.login.ErrorHandler.WriteFlowError",
		trace.WithAttributes(attribute.String("error", err.Error())))
	r = r.WithContext(ctx)
	defer otelx.End(span, &err)

	logger := s.d.Audit().
		WithError(err).
		WithRequest(r).
		WithField("login_flow", f.ToLoggerField())

	logger.
		Info("Encountered self-service login error.")

	if f == nil {
		trace.SpanFromContext(r.Context()).AddEvent(events.NewLoginFailed(r.Context(), uuid.Nil, "", "", false, err))
		s.forward(w, r, nil, err)
		return
	}

	span.SetAttributes(attribute.String("flow_id", f.ID.String()))
	trace.SpanFromContext(r.Context()).AddEvent(events.NewLoginFailed(r.Context(), f.ID, string(f.Type), string(f.RequestedAAL), f.Refresh, err))

	if expired, inner := s.PrepareReplacementForExpiredFlow(w, r, f, err); inner != nil {
		s.WriteFlowError(w, r, f, group, inner)
		return
	} else if expired != nil {
		if f.Type == flow.TypeAPI || x.IsJSONRequest(r) {
			s.d.Writer().WriteError(w, r, expired)
		} else {
			http.Redirect(w, r, expired.GetFlow().AppendTo(s.d.Config().SelfServiceFlowLoginUI(r.Context())).String(), http.StatusSeeOther)
		}
		return
	}

	f.UI.ResetMessages()
	if err := f.UI.ParseError(group, err); err != nil {
		s.forward(w, r, f, err)
		return
	}

	if err := sortNodes(r.Context(), f.UI.Nodes); err != nil {
		s.forward(w, r, f, err)
		return
	}

	if err := s.d.LoginFlowPersister().UpdateLoginFlow(r.Context(), f); err != nil {
		s.forward(w, r, f, err)
		return
	}

	if f.Type == flow.TypeBrowser && !x.IsJSONRequest(r) {
		http.Redirect(w, r, f.AppendTo(s.d.Config().SelfServiceFlowLoginUI(r.Context())).String(), http.StatusSeeOther)
		return
	}

	_, hasCode, _ := s.d.SessionTokenExchangePersister().CodeForFlow(r.Context(), f.ID)
	if f.Type == flow.TypeAPI && hasCode && group == node.OpenIDConnectGroup {
		http.Redirect(w, r, f.ReturnTo, http.StatusSeeOther)
		return
	}

	updatedFlow, innerErr := s.d.LoginFlowPersister().GetLoginFlow(r.Context(), f.ID)
	if innerErr != nil {
		s.forward(w, r, updatedFlow, innerErr)
		return
	}

	// If the flow has an OAuth2LoginChallenge, we try to fetch the login request from Hydra because we return the flow object as JSON.
	if updatedFlow.OAuth2LoginChallenge != "" {
		hlr, err := s.d.Hydra().GetLoginRequest(ctx, string(updatedFlow.OAuth2LoginChallenge))
		if err != nil {
			// We don't redirect back to the third party on errors because Hydra doesn't
			// give us the 3rd party return_uri when it redirects to the login UI.
			s.forward(w, r, updatedFlow, err)
			return
		}
		updatedFlow.HydraLoginRequest = hlr
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
