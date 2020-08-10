package login

import (
	"net/http"
	"time"

	"github.com/ory/kratos/text"

	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/x"
)

var (
	ErrHookAbortRequest = errors.New("aborted login hook execution")
)

type (
	errorHandlerDependencies interface {
		errorx.ManagementProvider
		x.WriterProvider
		x.LoggingProvider

		RequestPersistenceProvider
		HandlerProvider
	}

	ErrorHandlerProvider interface{ LoginRequestErrorHandler() *ErrorHandler }

	ErrorHandler struct {
		d errorHandlerDependencies
		c configuration.Provider
	}

	flowExpiredError struct {
		*herodot.DefaultError
		ago time.Duration
	}
)

func newRequestExpiredError(ago time.Duration) *flowExpiredError {
	return &flowExpiredError{
		ago: ago,
		DefaultError: herodot.ErrBadRequest.
			WithError("login flow expired").
			WithReasonf(`The login flow has expired. Please restart the flow.`).
			WithReasonf("The login flow expired %.2f minutes ago, please try again.", ago.Minutes()),
	}
}

func NewErrorHandler(d errorHandlerDependencies, c configuration.Provider) *ErrorHandler {
	return &ErrorHandler{
		d: d,
		c: c,
	}
}

func (s *ErrorHandler) HandleLoginError(
	w http.ResponseWriter,
	r *http.Request,
	ct identity.CredentialsType,
	rr *Flow,
	err error,
) {
	s.d.Audit().
		WithError(err).
		WithRequest(r).
		WithField("login_flow", rr).
		Info("Encountered self-service login error.")

	if e := new(flowExpiredError); errors.As(err, &e) {
		// create new flow because the old one is not valid
		a, err := s.d.LoginHandler().NewLoginFlow(w, r, rr.Type)
		if err != nil {
			// failed to create a new session and redirect to it, handle that error as a new one
			s.HandleLoginError(w, r, ct, rr, err)
			return
		}

		a.Messages.Add(text.NewErrorValidationLoginRequestExpired(e.ago))
		if err := s.d.LoginRequestPersister().UpdateLoginRequest(r.Context(), a); err != nil {
			redirTo, err := s.d.SelfServiceErrorManager().Create(r.Context(), w, r, err)
			if err != nil {
				s.HandleLoginError(w, r, ct, rr, err)
				return
			}
			http.Redirect(w, r, redirTo, http.StatusFound)
			return
		}

		http.Redirect(w, r, a.AppendTo(s.c.SelfServiceFlowLoginUI()).String(), http.StatusFound)
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
		s.d.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithDebugf("Methods: %+v", rr.Methods).WithErrorf(`Expected login method "%s" to exist in request. This is a bug in the code and should be reported on GitHub.`, ct)))
		return
	}

	if err := method.Config.ParseError(err); err != nil {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	if err := s.d.LoginRequestPersister().UpdateLoginRequestMethod(r.Context(), rr.ID, ct, method); err != nil {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	http.Redirect(w, r, rr.AppendTo(s.c.SelfServiceFlowLoginUI()).String(), http.StatusFound)
}
