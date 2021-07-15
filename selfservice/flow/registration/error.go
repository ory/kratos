package registration

import (
	"net/http"
	"net/url"
	"time"

	"github.com/ory/kratos/ui/node"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/text"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/config"
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
		config.Provider

		FlowPersistenceProvider
		HandlerProvider
	}

	ErrorHandlerProvider interface{ RegistrationFlowErrorHandler() *ErrorHandler }

	ErrorHandler struct {
		d errorHandlerDependencies
	}

	FlowExpiredError struct {
		*herodot.DefaultError
		ago time.Duration
	}
)

func NewFlowExpiredError(at time.Time) *FlowExpiredError {
	ago := time.Since(at)
	return &FlowExpiredError{
		ago: ago,
		DefaultError: herodot.ErrBadRequest.
			WithError("registration flow expired").
			WithReasonf(`The registration flow has expired. Please restart the flow.`).
			WithReasonf("The registration flow expired %.2f minutes ago, please try again.", ago.Minutes()),
	}
}

func NewErrorHandler(d errorHandlerDependencies) *ErrorHandler {
	return &ErrorHandler{d: d}
}

func (s *ErrorHandler) WriteFlowError(
	w http.ResponseWriter,
	r *http.Request,
	f *Flow,
	group node.Group,
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
			s.WriteFlowError(w, r, f, group, err)
			return
		}

		a.UI.AddMessage(group, text.NewErrorValidationRegistrationFlowExpired(e.ago))
		if err := s.d.RegistrationFlowPersister().UpdateRegistrationFlow(r.Context(), a); err != nil {
			s.forward(w, r, a, err)
			return
		}

		if f.Type == flow.TypeAPI || x.IsJSONRequest(r) {
			http.Redirect(w, r, urlx.CopyWithQuery(urlx.AppendPaths(s.d.Config(r.Context()).SelfPublicURL(r),
				RouteGetFlow), url.Values{"id": {a.ID.String()}}).String(), http.StatusSeeOther)
		} else {
			http.Redirect(w, r, a.AppendTo(s.d.Config(r.Context()).SelfServiceFlowRegistrationUI()).String(), http.StatusSeeOther)
		}
		return
	}

	f.UI.ResetMessages()
	if err := f.UI.ParseError(group, err); err != nil {
		s.forward(w, r, f, err)
		return
	}

	if err := SortNodes(f.UI.Nodes, s.d.Config(r.Context()).DefaultIdentityTraitsSchemaURL().String()); err != nil {
		s.forward(w, r, f, err)
		return
	}

	if err := s.d.RegistrationFlowPersister().UpdateRegistrationFlow(r.Context(), f); err != nil {
		s.forward(w, r, f, err)
		return
	}

	if f.Type == flow.TypeBrowser && !x.IsJSONRequest(r) {
		http.Redirect(w, r, f.AppendTo(s.d.Config(r.Context()).SelfServiceFlowRegistrationUI()).String(), http.StatusFound)
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
