package password

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/golang/gddo/httputil"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/gojsonschema"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/hive/schema"
	"github.com/ory/hive/selfservice"
	"github.com/ory/hive/x"
)

const (
	LoginPath = "/auth/browser/login/methods/password"
)

func (s *Strategy) setLoginRoutes(r *x.RouterPublic) {
	if _, _, ok := r.Lookup("POST", LoginPath); !ok {
		r.POST(LoginPath, s.handleLogin)
	}
}

func (s *Strategy) handleError(w http.ResponseWriter, r *http.Request, rr *selfservice.LoginRequest, err error) {
	if rr == nil {
		rr = NewBlankLoginRequest("")
	}

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

	var tc = func() *LoginRequestMethodConfig {
		if method, ok := rr.Methods[CredentialsType]; !ok {
			panic(fmt.Sprintf(`*selfservice.LoginRequest.Methods must have CredentialsType "%s" but did not: %+v`, CredentialsType, rr.Methods))
		} else if mc, ok := method.Config.(*LoginRequestMethodConfig); !ok {
			panic(fmt.Sprintf(`*selfservice.LoginRequest.Methods[%s].Config must be of type "*LoginRequestMethodConfig" but got: %T`, CredentialsType, method.Config))
		} else {
			return mc
		}
	}

	switch et := errors.Cause(err).(type) {
	case *herodot.DefaultError:
		if et.Error() == selfservice.ErrLoginRequestExpired.Error() {
			config := tc()
			config.Reset()
			config.Error = et.Reason()
			if err := s.d.LoginRequestManager().UpdateLoginRequest(r.Context(), rr.ID, CredentialsType, config); err != nil {
				s.d.ErrorManager().ForwardError(w, r, err)
				return
			}

			http.Redirect(w,
				r,
				urlx.CopyWithQuery(s.c.LoginURL(), url.Values{"request": {rr.ID}}).String(),
				http.StatusFound,
			)
			return
		}
		s.d.ErrorManager().ForwardError(w, r, err)
		return
	case schema.ResultErrors:
		config := tc()
		config.Reset()
		for k := range tidyForm(r.PostForm) {
			config.Fields.SetValue(k, r.PostForm.Get(k))
		}
		config.Fields.SetValue(csrfTokenName, s.cg(r))

		for k, e := range et {
			switch e.Type() {
			case "invalid_credentials":
				config.Error = e.Description()
			default:
				herodot.DefaultErrorLogger(s.d.Logger(), err).Warnf("A form error occurred during login (%d of %d): %s", k+1, len(et), e.String())
				config.Fields.SetError(
					e.Field(),
					e.String(),
				)
			}
		}

		if err := s.d.LoginRequestManager().UpdateLoginRequest(r.Context(), rr.ID, CredentialsType, config); err != nil {
			s.d.ErrorManager().ForwardError(w, r, err)
			return
		}

		http.Redirect(w,
			r,
			urlx.CopyWithQuery(s.c.LoginURL(), url.Values{"request": {rr.ID}}).String(),
			http.StatusFound,
		)
	default:
		s.d.ErrorManager().ForwardError(w, r, err)
	}
}

func (s *Strategy) handleLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	rid := r.URL.Query().Get("request")
	if len(rid) == 0 {
		s.handleError(w, r, nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("The request ID is missing.")))
		return
	}

	ar, err := s.d.LoginRequestManager().GetLoginRequest(r.Context(), rid)
	if err != nil {
		s.handleError(w, r, NewBlankLoginRequest(rid), err)
		return
	}

	var p LoginFormPayload
	if err := r.ParseForm(); err != nil {
		s.handleError(w, r, ar, errors.WithStack(herodot.ErrBadRequest.WithDebug(err.Error()).WithReasonf("Unable to parse HTTP form request: %s", err.Error())))
		return
	}
	if err := s.dc.Decode(&p, r.PostForm); err != nil {
		s.handleError(w, r, ar, errors.WithStack(herodot.ErrBadRequest.WithDebug(err.Error()).WithReasonf("Unable to parse HTTP form payload: %s", err.Error())))
		return
	}

	if len(p.Identifier) == 0 {
		s.handleError(w, r, ar, errors.WithStack(schema.NewRequiredError("", gojsonschema.NewJsonContext("identifier", nil))))
		return
	}

	if len(p.Password) == 0 {
		s.handleError(w, r, ar, errors.WithStack(schema.NewRequiredError("", gojsonschema.NewJsonContext("password", nil))))
		return
	}

	if err := ar.Valid(); err != nil {
		s.d.ErrorManager().ForwardError(w, r, err)
		return
	}

	i, c, err := s.d.IdentityPool().FindByCredentialsIdentifier(r.Context(), s.ID(), p.Identifier)
	if err != nil {
		s.handleError(w, r, ar, errors.WithStack(schema.NewInvalidCredentialsError()))
		return
	}

	var o CredentialsConfig
	d := json.NewDecoder(bytes.NewBuffer(c.Options))
	if err := d.Decode(&o); err != nil {
		s.d.ErrorManager().ForwardError(w, r, herodot.ErrInternalServerError.WithReason("The password credentials could not be decoded properly").WithDebug(err.Error()))
		return
	}

	if err := s.d.PasswordHasher().Compare([]byte(p.Password), []byte(o.HashedPassword)); err != nil {
		s.handleError(w, r, ar, errors.WithStack(schema.NewInvalidCredentialsError()))
		return
	}

	if err := s.d.LoginExecutor().PostLoginHook(w, r,
		s.d.PostLoginHooks(CredentialsType), ar, i); err != nil {
		s.d.ErrorManager().ForwardError(w, r, err)
		return
	}
}

func (s *Strategy) PopulateLoginMethod(r *http.Request, sr *selfservice.LoginRequest) error {
	action := urlx.CopyWithQuery(
		urlx.AppendPaths(s.c.SelfPublicURL(), LoginPath),
		url.Values{"request": {sr.ID}},
	)

	sr.Methods[CredentialsType] = &selfservice.LoginRequestMethod{
		Method: CredentialsType,
		Config: &LoginRequestMethodConfig{
			Action: action.String(),
			Fields: FormFields{
				"identifier": {
					Name:     "identifier",
					Type:     "text",
					Required: true,
				},
				"password": {
					Name:     "password",
					Type:     "password",
					Required: true,
				},
				csrfTokenName: {
					Name:     csrfTokenName,
					Type:     "hidden",
					Required: true,
					Value:    s.cg(r),
				},
			},
		},
	}
	return nil
}
