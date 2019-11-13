package profile

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/x"
)

var (
	ErrRequestExpired = herodot.ErrBadRequest.
		WithError("profile management request expired").
		WithReasonf(`The profile management request has expired. Please restart the flow.`)
)

type (
	errorHandlerDependencies interface {
		errorx.ManagementProvider
		x.WriterProvider
		x.LoggingProvider

		RequestPersistenceProvider
	}

	ErrorHandlerProvider interface{ ProfileRequestRequestErrorHandler() *ErrorHandler }

	ErrorHandler struct {
		d  errorHandlerDependencies
		c  configuration.Provider
		bd *x.BodyDecoder
	}
)

func NewErrorHandler(d errorHandlerDependencies, c configuration.Provider) *ErrorHandler {
	return &ErrorHandler{
		d: d,
		c: c,
	}
}

func (s *ErrorHandler) HandleProfileManagementError(
	w http.ResponseWriter,
	r *http.Request,
	ct identity.CredentialsType,
	rr *Request,
	err error,
) {
	s.d.Logger().WithError(err).
		WithField("details", fmt.Sprintf("%+v", err)).
		WithField("profile_request", rr).
		Warn("Encountered profile management error.")

	if rr == nil {
		s.d.ErrorManager().ForwardError(r.Context(), w, r, err)
		return
	} else if x.IsJSONRequest(r) {
		s.d.Writer().WriteError(w, r, err)
		return
	}

	if err := rr.Form.ParseError(err); err != nil {
		s.d.ErrorManager().ForwardError(r.Context(), w, r, err)
		return
	}

	if err := s.d.ProfileRequestPersister().UpdateProfileRequest(r.Context(), rr.ID, rr); err != nil {
		s.d.ErrorManager().ForwardError(r.Context(), w, r, err)
		return
	}

	http.Redirect(w, r,
		urlx.CopyWithQuery(s.c.ProfileURL(), url.Values{"request": {rr.ID}}).String(),
		http.StatusFound,
	)
}
