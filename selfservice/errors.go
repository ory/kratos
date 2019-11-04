package selfservice

import (
	"net/http"
	"net/url"

	"github.com/golang/gddo/httputil"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/x/stringslice"

	"github.com/ory/hive/driver/configuration"
	"github.com/ory/hive/identity"
	"github.com/ory/hive/schema"
	"github.com/ory/hive/selfservice/errorx"
	"github.com/ory/hive/x"
)

var (
	ErrIDTokenMissing = herodot.ErrBadRequest.
				WithError("authentication failed because id_token is missing").
				WithReasonf(`Authentication failed because no id_token was returned. Please accept the "openid" permission and try again.`)

	ErrScopeMissing = herodot.ErrBadRequest.
			WithError("authentication failed because a required scope was not granted").
			WithReasonf(`Unable to finish because one or more permissions were not granted. Please retry and accept all permissions.`)

	ErrLoginRequestExpired = herodot.ErrBadRequest.
				WithError("login request expired").
				WithReasonf(`The login request has expired. Please restart the flow.`)

	ErrRegistrationRequestExpired = herodot.ErrBadRequest.
					WithError("registration request expired").
					WithReasonf(`The registration request has expired. Please restart the flow.`)
)

type (
	errorHandlerDependencies interface {
		errorx.ManagementProvider
		x.WriterProvider
		x.LoggingProvider
		RegistrationRequestManagementProvider
		LoginRequestManagementProvider
	}

	RequestErrorHandlerProvider interface{ SelfServiceRequestErrorHandler() *ErrorHandler }

	ErrorHandler struct {
		d  errorHandlerDependencies
		c  configuration.Provider
		bd *BodyDecoder
	}

	ErrorHandlerOptions struct {
		// IgnoreValuesForKeys will not set the values for the given keys. This is useful for passwords,
		// csrf_tokens, and so on.
		IgnoreValuesForKeys []string

		AdditionalKeys map[string]interface{}
	}
)

func mergeErrorHandlerOptions(opts *ErrorHandlerOptions) *ErrorHandlerOptions {
	if opts != nil {
		return opts
	}
	return new(ErrorHandlerOptions)
}

func NewErrorHandler(d errorHandlerDependencies, c configuration.Provider) *ErrorHandler {
	return &ErrorHandler{
		d:  d,
		c:  c,
		bd: NewBodyDecoder(),
	}
}

func (s *ErrorHandler) json(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) bool {
	// TODO improve this https://github.com/ory/hive/issues/44 #44 #61
	if httputil.NegotiateContentType(
		r,
		[]string{"application/json", "text/html", "text/*", "*/*"},
		"text/*",
	) == "application/json" {
		switch errors.Cause(err).(type) {
		case schema.ResultErrors:
			s.d.Writer().WriteErrorCode(w, r, http.StatusBadRequest, err)
		default:
			s.d.Writer().WriteError(w, r, err)
		}
		return true
	}

	return false
}

func (s *ErrorHandler) handleHerodotError(err *herodot.DefaultError, config RequestMethodConfig) error {
	switch err.Error() {
	case ErrIDTokenMissing.Error():
		config.AddError(&FormError{Message: err.Reason()})
	case ErrScopeMissing.Error():
		config.AddError(&FormError{Message: err.Reason()})
	case ErrRegistrationRequestExpired.Error():
		config.AddError(&FormError{Message: err.Reason()})
	case ErrLoginRequestExpired.Error():
		config.AddError(&FormError{Message: err.Reason()})
	default:
		return err
	}

	return nil
}

func (s *ErrorHandler) handleValidationError(r *http.Request, err schema.ResultErrors, config RequestMethodConfig, opts *ErrorHandlerOptions) error {
	for k := range r.PostForm {
		if !stringslice.Has(opts.IgnoreValuesForKeys, k) {
			config.GetFormFields().SetValue(k, s.bd.ParseFormFieldOr(r.PostForm[k], r.PostForm.Get(k)))
		}
	}

	for k, v := range opts.AdditionalKeys {
		config.GetFormFields().SetValue(k, v)
	}

	for k, e := range err {
		herodot.DefaultErrorLogger(s.d.Logger(), err).Debugf("A validation error was caught (%d of %d): %s", k+1, len(err), e.String())
		switch e.Type() {
		case "invalid_credentials":
			config.AddError(&FormError{Message: e.Description()})
		default:
			fe := &FormError{Field: e.Field(), Message: e.String()}
			config.AddError(fe)
			config.GetFormFields().SetError(e.Field(), fe)
		}
	}

	return nil
}

func (s *ErrorHandler) handleError(
	r *http.Request,
	ct identity.CredentialsType,
	methods map[identity.CredentialsType]*DefaultRequestMethod,
	err error,
	opts *ErrorHandlerOptions,
) (*RequestMethodConfig, error) {
	method, ok := methods[ct]
	if !ok {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithError(`Expected method "%s" to exist in request. This is a bug in the code and should be reported on GitHub.`))
	}

	config := method.Config
	config.Reset()
	switch e := errors.Cause(err).(type) {
	case *herodot.DefaultError:
		return &config, s.handleHerodotError(e, config)
	case schema.ResultErrors:
		return &config, s.handleValidationError(r, e, config, opts)
	}

	return &config, err
}

func (s *ErrorHandler) HandleRegistrationError(
	w http.ResponseWriter,
	r *http.Request,
	ct identity.CredentialsType,
	rr *RegistrationRequest,
	err error,
	opts *ErrorHandlerOptions,
) {
	opts = mergeErrorHandlerOptions(opts)

	if rr == nil {
		s.d.ErrorManager().ForwardError(r.Context(), w, r, err)
		return
	} else if s.json(w, r, err) {
		return
	}

	config, err := s.handleError(r, ct, rr.Methods, err, opts)
	if err != nil {
		s.d.ErrorManager().ForwardError(r.Context(), w, r, err)
		return
	}

	if err := s.d.RegistrationRequestManager().UpdateRegistrationRequest(r.Context(), rr.ID, ct, *config); err != nil {
		s.d.ErrorManager().ForwardError(r.Context(), w, r, err)
		return
	}

	http.Redirect(w, r,
		urlx.CopyWithQuery(s.c.RegisterURL(), url.Values{"request": {rr.ID}}).String(),
		http.StatusFound,
	)
}

func (s *ErrorHandler) HandleLoginError(
	w http.ResponseWriter,
	r *http.Request,
	ct identity.CredentialsType,
	rr *LoginRequest,
	err error,
	opts *ErrorHandlerOptions,
) {
	opts = mergeErrorHandlerOptions(opts)

	if rr == nil {
		s.d.ErrorManager().ForwardError(r.Context(), w, r, err)
		return
	} else if s.json(w, r, err) {
		return
	}

	config, err := s.handleError(r, ct, rr.Methods, err, opts)
	if err != nil {
		s.d.ErrorManager().ForwardError(r.Context(), w, r, err)
		return
	}

	if err := s.d.LoginRequestManager().UpdateLoginRequest(r.Context(), rr.ID, ct, *config); err != nil {
		s.d.ErrorManager().ForwardError(r.Context(), w, r, err)
		return
	}

	http.Redirect(w, r,
		urlx.CopyWithQuery(s.c.LoginURL(), url.Values{"request": {rr.ID}}).String(),
		http.StatusFound,
	)
}
