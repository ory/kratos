package registration

import (
	"context"
	"net/http"
	"net/url"
	"time"

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

		RequestPersistenceProvider
		HandlerProvider
	}

	ErrorHandlerProvider interface{ RegistrationRequestErrorHandler() *ErrorHandler }

	ErrorHandler struct {
		d errorHandlerDependencies
		c configuration.Provider
	}

	requestExpiredError struct {
		*herodot.DefaultError
		ago time.Duration
	}
)

func newRequestExpiredError(ago time.Duration) *requestExpiredError {
	return &requestExpiredError{
		ago: ago,
		DefaultError: herodot.ErrBadRequest.
			WithError("registration request expired").
			WithReasonf(`The registration request has expired. Please restart the flow.`).
			WithReasonf("The registration request expired %.2f minutes ago, please try again.", ago.Minutes()),
	}
}

func NewErrorHandler(d errorHandlerDependencies, c configuration.Provider) *ErrorHandler {
	return &ErrorHandler{
		d: d,
		c: c,
	}
}

func (s *ErrorHandler) HandleRegistrationError(
	w http.ResponseWriter,
	r *http.Request,
	ct identity.CredentialsType,
	rr *Request,
	err error,
) {
	s.d.Audit().
		WithError(err).
		WithRequest(r).
		WithField("registration_request", rr).
		Info("Encountered self-service request error.")

	if e := new(requestExpiredError); errors.As(err, &e) {
		// create new request because the old one is not valid
		a, err := s.d.RegistrationHandler().NewRegistrationRequest(w, r)
		if err != nil {
			// failed to create a new session and redirect to it, handle that error as a new one
			s.HandleRegistrationError(w, r, ct, rr, err)
			return
		}

		a.Messages.Add(text.NewErrorValidationRegistrationRequestExpired(e.ago))
		if err := s.d.RegistrationRequestPersister().UpdateRegistrationRequest(context.TODO(), a); err != nil {
			redirTo, err := s.d.SelfServiceErrorManager().Create(r.Context(), w, r, err)
			if err != nil {
				s.HandleRegistrationError(w, r, ct, rr, err)
				return
			}
			http.Redirect(w, r, redirTo, http.StatusFound)
			return
		}

		http.Redirect(w, r, urlx.CopyWithQuery(s.c.RegisterURL(), url.Values{"request": {a.ID.String()}}).String(), http.StatusFound)
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

	if err := s.d.RegistrationRequestPersister().UpdateRegistrationRequestMethod(r.Context(), rr.ID, ct, method); err != nil {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	http.Redirect(w, r,
		urlx.CopyWithQuery(s.c.RegisterURL(), url.Values{"request": {rr.ID.String()}}).String(),
		http.StatusFound,
	)
}
