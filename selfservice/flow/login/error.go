package login

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
		config.Provider

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
	e := new(flow.ExpiredError)
	if !errors.As(err, &e) {
		return nil, nil
	}
	// create new flow because the old one is not valid
	a, err := s.d.LoginHandler().FromOldFlow(w, r, *f)
	if err != nil {
		return nil, err
	}

	a.UI.Messages.Add(text.NewErrorValidationLoginFlowExpired(e.Ago))
	if err := s.d.LoginFlowPersister().UpdateLoginFlow(r.Context(), a); err != nil {
		return nil, err
	}

	return e.WithFlow(a), nil
}

func (s *ErrorHandler) WriteFlowError(w http.ResponseWriter, r *http.Request, f *Flow, group node.Group, err error) {
	s.d.Audit().
		WithError(err).
		WithRequest(r).
		WithField("login_flow", f).
		Info("Encountered self-service login error.")

	if f == nil {
		s.forward(w, r, nil, err)
		return
	}

	if expired, inner := s.PrepareReplacementForExpiredFlow(w, r, f, err); inner != nil {
		s.WriteFlowError(w, r, f, group, inner)
		return
	} else if expired != nil {
		if f.Type == flow.TypeAPI || x.IsJSONRequest(r) {
			s.d.Writer().WriteError(w, r, expired)
		} else {
			http.Redirect(w, r, expired.GetFlow().AppendTo(s.d.Config(r.Context()).SelfServiceFlowLoginUI()).String(), http.StatusSeeOther)
		}
		return
	}

	f.UI.ResetMessages()
	if err := f.UI.ParseError(group, err); err != nil {
		s.forward(w, r, f, err)
		return
	}

	if err := sortNodes(f.UI.Nodes); err != nil {
		s.forward(w, r, f, err)
		return
	}

	if err := s.d.LoginFlowPersister().UpdateLoginFlow(r.Context(), f); err != nil {
		s.forward(w, r, f, err)
		return
	}

	if f.Type == flow.TypeBrowser && !x.IsJSONRequest(r) {
		http.Redirect(w, r, f.AppendTo(s.d.Config(r.Context()).SelfServiceFlowLoginUI()).String(), http.StatusFound)
		return
	}

	updatedFlow, innerErr := s.d.LoginFlowPersister().GetLoginFlow(r.Context(), f.ID)
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
