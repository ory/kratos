package verification

import (
	"net/http"
	"net/url"

	"github.com/ory/kratos/ui/node"

	"github.com/pkg/errors"

	"github.com/ory/x/sqlxx"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x"
)

var (
	ErrHookAbortFlow = errors.New("aborted verification hook execution")
)

type (
	errorHandlerDependencies interface {
		errorx.ManagementProvider
		x.WriterProvider
		x.LoggingProvider
		x.CSRFTokenGeneratorProvider
		config.Provider
		FlowPersistenceProvider
		StrategyProvider
	}

	ErrorHandlerProvider interface {
		VerificationFlowErrorHandler() *ErrorHandler
	}

	ErrorHandler struct {
		d errorHandlerDependencies
	}
)

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
		WithField("verification_flow", f).
		Info("Encountered self-service verification error.")

	if f == nil {
		s.forward(w, r, nil, err)
		return
	}

	if e := new(flow.ExpiredError); errors.As(err, &e) {
		// create new flow because the old one is not valid
		a, err := FromOldFlow(s.d.Config(r.Context()), s.d.Config(r.Context()).SelfServiceFlowVerificationRequestLifespan(),
			s.d.GenerateCSRFToken(r), r, s.d.VerificationStrategies(r.Context()), f)
		if err != nil {
			// failed to create a new session and redirect to it, handle that error as a new one
			s.WriteFlowError(w, r, f, group, err)
			return
		}

		a.UI.Messages.Add(text.NewErrorValidationVerificationFlowExpired(e.Ago))
		if err := s.d.VerificationFlowPersister().CreateVerificationFlow(r.Context(), a); err != nil {
			s.forward(w, r, a, err)
			return
		}

		if f.Type == flow.TypeAPI || x.IsJSONRequest(r) {
			http.Redirect(w, r, urlx.CopyWithQuery(urlx.AppendPaths(s.d.Config(r.Context()).SelfPublicURL(),
				RouteGetFlow), url.Values{"id": {a.ID.String()}}).String(), http.StatusSeeOther)
		} else {
			http.Redirect(w, r, a.AppendTo(s.d.Config(r.Context()).SelfServiceFlowVerificationUI()).String(), http.StatusSeeOther)
		}
		return
	}

	if err := f.UI.ParseError(group, err); err != nil {
		s.forward(w, r, f, err)
		return
	}

	f.Active = sqlxx.NullString(group)
	if err := s.d.VerificationFlowPersister().UpdateVerificationFlow(r.Context(), f); err != nil {
		s.forward(w, r, f, err)
		return
	}

	if f.Type == flow.TypeBrowser && !x.IsJSONRequest(r) {
		http.Redirect(w, r, f.AppendTo(s.d.Config(r.Context()).SelfServiceFlowVerificationUI()).String(), http.StatusSeeOther)
		return
	}

	updatedFlow, innerErr := s.d.VerificationFlowPersister().GetVerificationFlow(r.Context(), f.ID)
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

	if rr.Type == flow.TypeAPI || x.IsJSONRequest(r) {
		s.d.Writer().WriteErrorCode(w, r, x.RecoverStatusCode(err, http.StatusBadRequest), err)
	} else {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
	}
}
