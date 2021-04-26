package login

import (
	"net/http"
	"net/url"
	"time"

	"github.com/ory/kratos/ui/node"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/text"

	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/x"
)

var (
	ErrHookAbortFlow   = errors.New("aborted login hook execution")
	ErrAlreadyLoggedIn = herodot.ErrBadRequest.WithReason("A valid session was detected and thus login is not possible. Did you forget to set `?refresh=true`?")
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
			WithError("login flow expired").
			WithReasonf(`The login flow has expired. Please restart the flow.`).
			WithReasonf("The login flow expired %.2f minutes ago, please try again.", ago.Minutes()),
	}
}

func NewFlowErrorHandler(d errorHandlerDependencies) *ErrorHandler {
	return &ErrorHandler{d: d}
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

	if e := new(FlowExpiredError); errors.As(err, &e) {
		// create new flow because the old one is not valid
		a, err := s.d.LoginHandler().NewLoginFlow(w, r, f.Type)
		if err != nil {
			// failed to create a new session and redirect to it, handle that error as a new one
			s.WriteFlowError(w, r, f, group, err)
			return
		}

		a.UI.Messages.Add(text.NewErrorValidationLoginFlowExpired(e.ago))
		if err := s.d.LoginFlowPersister().UpdateLoginFlow(r.Context(), a); err != nil {
			s.forward(w, r, a, err)
			return
		}

		if f.Type == flow.TypeAPI {
			http.Redirect(w, r, urlx.CopyWithQuery(urlx.AppendPaths(s.d.Config(r.Context()).SelfPublicURL(r),
				RouteGetFlow), url.Values{"id": {a.ID.String()}}).String(), http.StatusFound)
		} else {
			http.Redirect(w, r, a.AppendTo(s.d.Config(r.Context()).SelfServiceFlowLoginUI()).String(), http.StatusFound)
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

	if f.Type == flow.TypeBrowser {
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
