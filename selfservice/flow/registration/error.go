package registration

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/text"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/x"
)

var (
	ErrHookAbortRequest = errors.New("aborted registration hook execution")
)

type (
	errorHandlerDependencies interface {
		errorx.ManagementProvider
		x.WriterProvider
		x.LoggingProvider

		FlowPersistenceProvider
		HandlerProvider
	}

	ErrorHandlerProvider interface{ RegistrationFlowErrorHandler() *ErrorHandler }

	ErrorHandler struct {
		d errorHandlerDependencies
		c configuration.Provider
	}

	FlowExpiredError struct {
		*herodot.DefaultError
		ago time.Duration
	}
)

func NewFlowExpiredError(ago time.Duration) *FlowExpiredError {
	return &FlowExpiredError{
		ago: ago,
		DefaultError: herodot.ErrBadRequest.
			WithError("registration flow expired").
			WithReasonf(`The registration flow has expired. Please restart the flow.`).
			WithReasonf("The registration flow expired %.2f minutes ago, please try again.", ago.Minutes()),
	}
}

func NewErrorHandler(d errorHandlerDependencies, c configuration.Provider) *ErrorHandler {
	return &ErrorHandler{
		d: d,
		c: c,
	}
}

func (s *ErrorHandler) WriteFlowError(
	w http.ResponseWriter,
	r *http.Request,
	ct identity.CredentialsType,
	rr *Flow,
	err error,
) {
	s.d.Audit().
		WithError(err).
		WithRequest(r).
		WithField("registration_flow", rr).
		Info("Encountered self-service flow error.")

	if e := new(FlowExpiredError); errors.As(err, &e) {
		// create new flow because the old one is not valid
		a, err := s.d.RegistrationHandler().NewRegistrationFlow(w, r)
		if err != nil {
			// failed to create a new session and redirect to it, handle that error as a new one
			s.WriteFlowError(w, r, ct, rr, err)
			return
		}

		a.Messages.Add(text.NewErrorValidationRegistrationFlowExpired(e.ago))
		if err := s.d.RegistrationFlowPersister().UpdateRegistrationFlow(context.TODO(), a); err != nil {
			redirTo, err := s.d.SelfServiceErrorManager().Create(r.Context(), w, r, err)
			if err != nil {
				s.WriteFlowError(w, r, ct, rr, err)
				return
			}
			http.Redirect(w, r, redirTo, http.StatusFound)
			return
		}

		http.Redirect(w, r, urlx.CopyWithQuery(s.c.SelfServiceFlowRegisterUI(), url.Values{"request": {a.ID.String()}}).String(), http.StatusFound)
		return
	}

	if rr == nil {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	} else if x.IsJSONRequest(r) {
		s.d.Writer().WriteError(w, r, err)
		return
	}

	method, ok := rr.Methods[ct]
	if !ok {
		s.d.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithDebugf("Methods: %+v", rr.Methods).WithErrorf(`Expected registration method "%s" to exist in request. This is a bug in the code and should be reported on GitHub.`, ct)))
		return
	}

	if err := method.Config.ParseError(err); err != nil {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	if err := s.d.RegistrationFlowPersister().UpdateRegistrationFlowMethod(r.Context(), rr.ID, ct, method); err != nil {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	http.Redirect(w, r,
		urlx.CopyWithQuery(s.c.SelfServiceFlowRegisterUI(), url.Values{"request": {rr.ID.String()}}).String(),
		http.StatusFound,
	)
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

