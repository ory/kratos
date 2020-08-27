package recovery

import (
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	"github.com/ory/x/sqlxx"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/x"
)

var (
	ErrRequestExpired = herodot.ErrBadRequest.
				WithError("recovery request expired").
				WithReasonf(`The recovery request has expired. Please restart the flow.`)
	ErrHookAbortRequest             = errors.New("aborted recovery hook execution")
	ErrRequestNeedsReAuthentication = herodot.ErrForbidden.WithReasonf("The login session is too old and thus not allowed to update these fields. Please re-authenticate.")
)

type (
	errorHandlerDependencies interface {
		errorx.ManagementProvider
		x.WriterProvider
		x.LoggingProvider

		RequestPersistenceProvider
	}

	ErrorHandlerProvider interface{ RecoveryRequestErrorHandler() *ErrorHandler }

	ErrorHandler struct {
		d errorHandlerDependencies
		c configuration.Provider
	}
)

func NewErrorHandler(d errorHandlerDependencies, c configuration.Provider) *ErrorHandler {
	return &ErrorHandler{
		d: d,
		c: c,
	}
}

func (s *ErrorHandler) HandleRecoveryError(
	w http.ResponseWriter,
	r *http.Request,
	rr *Flow,
	err error,
	method string,
) {
	s.d.Audit().
		WithError(err).
		WithRequest(r).
		WithField("recovery_request", rr).
		Info("Encountered self-service recovery error.")

	if rr == nil {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	} else if x.IsJSONRequest(r) {
		s.d.Writer().WriteError(w, r, err)
		return
	}

	if _, ok := rr.Methods[method]; !ok {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, errors.WithStack(x.PseudoPanic.WithReasonf("Expected recovery method %s to exist.", method)))
		return
	}

	rr.Active = sqlxx.NullString(method)
	if err := rr.Methods[method].Config.ParseError(err); err != nil {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	if err := s.d.RecoveryRequestPersister().UpdateRecoveryRequest(r.Context(), rr); err != nil {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	http.Redirect(w, r,
		urlx.CopyWithQuery(s.c.SelfServiceFlowRecoveryUI(), url.Values{"request": {rr.ID.String()}}).String(),
		http.StatusFound,
	)
}
