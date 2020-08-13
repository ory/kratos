package registration

import (
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
	ErrHookAbortFlow   = errors.New("aborted registration hook execution")
	ErrAlreadyLoggedIn = herodot.ErrBadRequest.WithReason("A valid session was detected and thus registration is not possible.")
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
	f *Flow,
	err error,
) {
	s.d.Audit().
		WithError(err).
		WithRequest(r).
		WithField("registration_flow", f).
		Info("Encountered self-service flow error.")

	if f == nil {
		s.forward(w, r, nil, err)
		return
	}

	if e := new(FlowExpiredError); errors.As(err, &e) {
		// create new flow because the old one is not valid
		a, err := s.d.RegistrationHandler().NewRegistrationFlow(w, r, f.Type)
		if err != nil {
			// failed to create a new session and redirect to it, handle that error as a new one
			s.WriteFlowError(w, r, ct, f, err)
			return
		}

		a.Messages.Add(text.NewErrorValidationRegistrationFlowExpired(e.ago))
		if err := s.d.RegistrationFlowPersister().UpdateRegistrationFlow(r.Context(), a); err != nil {
			s.forward(w, r, a, err)
			return
		}

		if f.Type == flow.TypeAPI {
			http.Redirect(w, r, urlx.CopyWithQuery(urlx.AppendPaths(s.c.SelfPublicURL(),
				RouteGetFlow), url.Values{"id": {a.ID.String()}}).String(), http.StatusFound)
		} else {
			http.Redirect(w, r, a.AppendTo(s.c.SelfServiceFlowRegistrationUI()).String(), http.StatusFound)
		}
		return
	}

	method, ok := f.Methods[ct]
	if !ok {
		s.forward(w, r, f, errors.WithStack(herodot.ErrInternalServerError.
			WithErrorf(`Expected registration method "%s" to exist in flow. This is a bug in the code and should be reported on GitHub.`, ct)))
		return
	}

	if err := method.Config.ParseError(err); err != nil {
		s.forward(w, r, f, err)
		return
	}

	if err := s.d.RegistrationFlowPersister().UpdateRegistrationFlowMethod(r.Context(), f.ID, ct, method); err != nil {
		s.forward(w, r, f, err)
		return
	}

	if f.Type == flow.TypeBrowser {
		http.Redirect(w, r, f.AppendTo(s.c.SelfServiceFlowRegistrationUI()).String(), http.StatusFound)
		return
	}

	innerRegistrationFlow, innerErr := s.d.RegistrationFlowPersister().GetRegistrationFlow(r.Context(), f.ID)
	if innerErr != nil {
		s.forward(w, r, innerRegistrationFlow, innerErr)
	}

	s.d.Writer().WriteCode(w, r, x.RecoverStatusCode(err, http.StatusBadRequest), innerRegistrationFlow)
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
