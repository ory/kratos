package login

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/x/errorsx"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/x"
)

var (
	ErrHookAbortRequest = errors.New("abort hook")
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

	requestExpiredError struct {
		*herodot.DefaultError
	}
)

func newRequestExpiredError(since time.Duration) requestExpiredError {
	return requestExpiredError{
		herodot.ErrBadRequest.
			WithError("login request expired").
			WithReasonf(`The login request has expired. Please restart the flow.`).
			WithReasonf("The login request expired %.2f minutes ago, please try again.", since.Minutes()),
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
	rr *Request,
	err error,
) {
	s.d.Logger().WithError(err).
		WithField("details", fmt.Sprintf("%+v", err)).
		WithField("credentials_type", ct).
		WithField("login_request", rr).
		Warn("Encountered login error.")

	if _, ok := errorsx.Cause(err).(requestExpiredError); ok {
		// create new request because the old one is not valid
		if err = s.d.LoginHandler().NewLoginRequest(w, r, func(a *Request) (string, error) {
			for name, method := range a.Methods {
				method.Config.AddError(&form.Error{Message: "Your session expired, please try again."})
				if err := s.d.LoginRequestPersister().UpdateLoginRequestMethod(r.Context(), a.ID, name, method); err != nil {
					return s.d.SelfServiceErrorManager().Create(r.Context(), w, r, err)
				}
				a.Methods[name] = method
			}

			return urlx.CopyWithQuery(s.c.LoginURL(), url.Values{"request": {a.ID.String()}}).String(), nil
		}); err != nil {
			// failed to create a new session and redirect to it, handle that error as a new one
			s.HandleLoginError(w, r, ct, rr, err)
		}
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
		s.d.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithErrorf(`Expected method "%s" to exist in request. This is a bug in the code and should be reported on GitHub.`, ct)))
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

	http.Redirect(w, r,
		urlx.CopyWithQuery(s.c.LoginURL(), url.Values{"request": {rr.ID.String()}}).String(),
		http.StatusFound,
	)
}
