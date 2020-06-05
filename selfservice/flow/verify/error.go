package verify

import (
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/text"
	"github.com/ory/kratos/x"
)

type (
	errorHandlerDependencies interface {
		errorx.ManagementProvider
		x.WriterProvider
		x.LoggingProvider
		x.CSRFTokenGeneratorProvider
		PersistenceProvider
	}

	ErrorHandlerProvider interface{ VerificationRequestErrorHandler() *ErrorHandler }

	ErrorHandler struct {
		d errorHandlerDependencies
		c configuration.Provider
	}

	errRequestExpired struct {
		*herodot.DefaultError
		ago time.Duration
	}
)

func newErrRequestRequired(ago time.Duration) error {
	return errors.WithStack(&errRequestExpired{ago: ago, DefaultError: herodot.ErrBadRequest.
		WithError("verify request expired").
		WithReasonf("The verification request expired %.2f minutes ago, please try again.", ago.Minutes())})
}

func NewErrorHandler(d errorHandlerDependencies, c configuration.Provider) *ErrorHandler {
	return &ErrorHandler{
		d: d,
		c: c,
	}
}

func (s *ErrorHandler) HandleVerificationError(
	w http.ResponseWriter,
	r *http.Request,
	rr *Request,
	err error,
) {
	s.d.Audit().
		WithError(err).
		WithRequest(r).
		WithField("verify_request", rr).
		Info("Encountered self-service verification error.")

	if rr == nil {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	} else if x.IsJSONRequest(r) {
		s.d.Writer().WriteError(w, r, err)
		return
	}

	if e := new(errRequestExpired); errors.As(err, &e) {
		a := NewRequest(
			s.c.SelfServiceSettingsRequestLifespan(), r, rr.Via,
			urlx.AppendPaths(s.c.SelfPublicURL(), PublicVerificationRequestPath), s.d.GenerateCSRFToken,
		)

		a.Form.AddMessage(text.NewErrorValidationVerificationRequestExpired(e.ago))
		if err := s.d.VerificationPersister().CreateVerificationRequest(r.Context(), a); err != nil {
			s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
			return
		}

		http.Redirect(w, r,
			urlx.CopyWithQuery(s.c.VerificationURL(), url.Values{"request": {a.ID.String()}}).String(),
			http.StatusFound,
		)
		return
	}

	if err := rr.Form.ParseError(err); err != nil {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	if err := s.d.VerificationPersister().UpdateVerificationRequest(r.Context(), rr); err != nil {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	http.Redirect(w, r,
		urlx.CopyWithQuery(s.c.VerificationURL(), url.Values{"request": {rr.ID.String()}}).String(),
		http.StatusFound,
	)
}
