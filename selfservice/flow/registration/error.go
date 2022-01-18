package registration

import (
	"net/http"

	"github.com/ory/kratos/ui/node"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/text"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/x"
)

var (
	ErrHookAbortFlow        = errors.New("aborted registration hook execution")
	ErrAlreadyLoggedIn      = herodot.ErrBadRequest.WithID(text.ErrIDAlreadyLoggedIn).WithError("you are already logged in").WithReason("A valid session was detected and thus registration is not possible.")
	ErrRegistrationDisabled = herodot.ErrBadRequest.WithID(text.ErrIDSelfServiceFlowDisabled).WithError("registration flow disabled").WithReason("Registration is not allowed because it was disabled.")
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

	a.UI.Messages.Add(text.NewErrorValidationRegistrationFlowExpired(e.Ago))
	if err := s.d.RegistrationFlowPersister().UpdateRegistrationFlow(r.Context(), a); err != nil {
		return nil, err
	}

	return e.WithFlow(a), nil
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

	if expired, inner := s.PrepareReplacementForExpiredFlow(w, r, f, err); inner != nil {
		s.forward(w, r, f, err)
		return
	} else if expired != nil {
		if f.Type == flow.TypeAPI || x.IsJSONRequest(r) {
			s.d.Writer().WriteError(w, r, expired)
		} else {
			http.Redirect(w, r, expired.GetFlow().AppendTo(s.d.Config(r.Context()).SelfServiceFlowRegistrationUI()).String(), http.StatusSeeOther)
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
