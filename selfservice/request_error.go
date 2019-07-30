package selfservice

import (
	"net/http"
	"net/url"

	"github.com/golang/gddo/httputil"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/stringslice"
	"github.com/ory/x/urlx"

	"github.com/ory/hive/driver/configuration"
	"github.com/ory/hive/errorx"
	"github.com/ory/hive/identity"
	"github.com/ory/hive/schema"
	"github.com/ory/hive/x"
)

type requestErrorHandlerDependencies interface {
	errorx.ManagementProvider
	x.WriterProvider
	x.LoggingProvider
	RegistrationRequestManagementProvider
	LoginRequestManagementProvider
}

type RequestErrorHandler struct {
	d requestErrorHandlerDependencies
	c configuration.Provider
}

func NewRequestErrorHandler(d requestErrorHandlerDependencies, c configuration.Provider) *RequestErrorHandler {
	return &RequestErrorHandler{
		d: d,
		c: c,
	}
}

type RequestErrorHandlerOptions struct {
	// IgnoreValuesForKeys will not set the values for the given keys. This is useful for passwords,
	// csrf_tokens, and so on.
	IgnoreValuesForKeys []string

	AdditionalKeys url.Values
}

type RequestErrorHandlerProvider interface {
	SelfServiceRequestErrorHandler() *RequestErrorHandler
}

func (s *RequestErrorHandler) HandleRegistrationError(
	w http.ResponseWriter,
	r *http.Request,
	ct identity.CredentialsType,
	rr *RegistrationRequest,
	err error,
	opts *RequestErrorHandlerOptions,
) {
	if rr == nil {
		s.d.ErrorManager().ForwardError(w, r, err)
		return
	}

	if opts == nil {
		opts = new(RequestErrorHandlerOptions)
	}

	// TODO improve this https://github.com/ory/hive/issues/44
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
		return
	}

	method, ok := rr.Methods[ct]
	if !ok {
		s.d.ErrorManager().ForwardError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithError(`Expected method "%s" to exist in request. This is a bug in the code and should be reported on GitHub.`)))
		return
	}

	config := method.Config
	config.Reset()

	switch et := errors.Cause(err).(type) {
	case *herodot.DefaultError:
		switch et.Error() {
		case ErrRegistrationRequestExpired.Error():
			config.SetError(et.Reason())
		default:
			s.d.ErrorManager().ForwardError(w, r, err)
			return
		}
	case schema.ResultErrors:
		config.Reset()

		for k := range r.PostForm {
			if !stringslice.Has(opts.IgnoreValuesForKeys, k) {
				config.GetFormFields().SetValue(k, r.PostForm.Get(k))
			}
		}

		for k := range opts.AdditionalKeys {
			config.GetFormFields().SetValue(k, opts.AdditionalKeys.Get(k))
		}

		for k, e := range et {
			herodot.DefaultErrorLogger(s.d.Logger(), err).Warnf("A form error occurred during registration (%d of %d): %s", k+1, len(et), e.String())
			name := e.Field()
			config.GetFormFields().SetError(name, e.String())
		}
	default:
		s.d.ErrorManager().ForwardError(w, r, err)
		return
	}

	if err := s.d.RegistrationRequestManager().UpdateRegistrationRequest(r.Context(), rr.ID, ct, config); err != nil {
		s.d.ErrorManager().ForwardError(w, r, err)
		return
	}

	http.Redirect(w,
		r,
		urlx.CopyWithQuery(s.c.RegisterURL(), url.Values{"request": {rr.ID}}).String(),
		http.StatusFound,
	)
}

func (s *RequestErrorHandler) HandleLoginError(
	w http.ResponseWriter,
	r *http.Request,
	ct identity.CredentialsType,
	rr *LoginRequest,
	err error,
	opts *RequestErrorHandlerOptions,
) {
	if rr == nil {
		s.d.ErrorManager().ForwardError(w, r, err)
		return
	}

	if opts == nil {
		opts = new(RequestErrorHandlerOptions)
	}

	// TODO improve this https://github.com/ory/hive/issues/44
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
		return
	}

	method, ok := rr.Methods[ct]
	if !ok {
		s.d.ErrorManager().ForwardError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithError(`Expected method "%s" to exist in request. This is a bug in the code and should be reported on GitHub.`)))
		return
	}

	config := method.Config
	config.Reset()

	switch et := errors.Cause(err).(type) {
	case *herodot.DefaultError:
		switch et.Error() {
		case ErrIDTokenMissing.Error():
			fallthrough
		case ErrScopeMissing.Error():
			fallthrough
		case ErrLoginRequestExpired.Error():
			config.SetError(et.Reason())
		default:
			s.d.ErrorManager().ForwardError(w, r, err)
			return
		}
	case schema.ResultErrors:
		for k := range r.PostForm {
			if !stringslice.Has(opts.IgnoreValuesForKeys, k) {
				config.GetFormFields().SetValue(k, r.PostForm.Get(k))
			}
		}

		for k := range opts.AdditionalKeys {
			config.GetFormFields().SetValue(k, opts.AdditionalKeys.Get(k))
		}

		for k, e := range et {
			switch e.Type() {
			case "invalid_credentials":
				config.SetError(e.Description())
			default:
				herodot.DefaultErrorLogger(s.d.Logger(), err).Warnf("A form error occurred during registration (%d of %d): %s", k+1, len(et), e.String())
				config.GetFormFields().SetError(e.Field(), e.String())
			}
		}
	default:
		s.d.ErrorManager().ForwardError(w, r, err)
		return
	}

	if err := s.d.LoginRequestManager().UpdateLoginRequest(r.Context(), rr.ID, ct, config); err != nil {
		s.d.ErrorManager().ForwardError(w, r, err)
		return
	}

	http.Redirect(w,
		r,
		urlx.CopyWithQuery(s.c.LoginURL(), url.Values{"request": {rr.ID}}).String(),
		http.StatusFound,
	)
}
