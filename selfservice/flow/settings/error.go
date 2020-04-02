package settings

import (
	"fmt"
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
				WithError("settings request expired").
				WithReasonf(`The settings request has expired. Please restart the flow.`)
	ErrHookAbortRequest             = errors.New("abort hook")
	ErrRequestNeedsReAuthentication = herodot.ErrForbidden.WithReasonf("The login session is too old and thus not allowed to update these fields. Please re-authenticate.")
)

type (
	errorHandlerDependencies interface {
		errorx.ManagementProvider
		x.WriterProvider
		x.LoggingProvider

		RequestPersistenceProvider
	}

	ErrorHandlerProvider interface{ SettingsRequestErrorHandler() *ErrorHandler }

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

//
// func (s *ErrorHandler) reauthenticate(
// 	w http.ResponseWriter,
// 	r *http.Request,
// 	rr *Request,
// 	err error,
// 	method string) {
//
// 	if err := s.d.SettingsRequestPersister().UpdateSettingsRequest(r.Context(), rr); err != nil {
// 		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
// 		return
// 	}
//
// }

func (s *ErrorHandler) HandleSettingsError(
	w http.ResponseWriter,
	r *http.Request,
	rr *Request,
	err error,
	method string,
) {
	s.d.Logger().WithError(err).
		WithField("details", fmt.Sprintf("%+v", err)).
		WithField("settings_request", rr).
		Warn("Encountered settings error.")

	if rr == nil {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	} else if x.IsJSONRequest(r) {
		s.d.Writer().WriteError(w, r, err)
		return
	}

	if _, ok := rr.Methods[method]; !ok {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Expected settings method %s to exist.", method)))
		return
	}

	rr.Active = sqlxx.NullString(method)
	// if errorsx.Cause(err) == ErrRequestNeedsReAuthentication {
	// 	s.reauthenticate(w, r, rr, err, method)
	// 	return
	// }

	if err := rr.Methods[method].Config.ParseError(err); err != nil {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	if err := s.d.SettingsRequestPersister().UpdateSettingsRequest(r.Context(), rr); err != nil {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	http.Redirect(w, r,
		urlx.CopyWithQuery(s.c.SettingsURL(), url.Values{"request": {rr.ID.String()}}).String(),
		http.StatusFound,
	)
}
