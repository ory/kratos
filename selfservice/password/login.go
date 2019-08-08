package password

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/gojsonschema"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/hive/schema"
	"github.com/ory/hive/selfservice"
	"github.com/ory/hive/session"
	"github.com/ory/hive/x"
)

const (
	LoginPath = "/auth/browser/methods/password/login"
)

func (s *Strategy) setLoginRoutes(r *x.RouterPublic) {
	if _, _, ok := r.Lookup("POST", LoginPath); !ok {
		r.POST(LoginPath, s.d.SessionHandler().IsNotAuthenticated(s.handleLogin, session.RedirectOnAuthenticated(s.c)))
	}
}

func (s *Strategy) handleLoginError(w http.ResponseWriter, r *http.Request, rr *selfservice.LoginRequest, err error) {
	s.d.SelfServiceRequestErrorHandler().HandleLoginError(w, r, CredentialsType, rr, err,
		&selfservice.ErrorHandlerOptions{
			AdditionalKeys: map[string]interface{}{
				selfservice.CSRFTokenName: s.cg(r),
			},
			IgnoreValuesForKeys: []string{"password"},
		},
	)
}

func (s *Strategy) handleLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	rid := r.URL.Query().Get("request")
	if len(rid) == 0 {
		s.handleLoginError(w, r, nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("The request Code is missing.")))
		return
	}

	ar, err := s.d.LoginRequestManager().GetLoginRequest(r.Context(), rid)
	if err != nil {
		s.handleLoginError(w, r, nil, err)
		return
	}

	var p LoginFormPayload
	if err := r.ParseForm(); err != nil {
		s.handleLoginError(w, r, ar, errors.WithStack(herodot.ErrBadRequest.WithDebug(err.Error()).WithReasonf("Unable to parse HTTP form request: %s", err.Error())))
		return
	}
	if err := s.dc.Decode(&p, r.PostForm); err != nil {
		s.handleLoginError(w, r, ar, errors.WithStack(herodot.ErrBadRequest.WithDebug(err.Error()).WithReasonf("Unable to parse HTTP form payload: %s", err.Error())))
		return
	}

	if len(p.Identifier) == 0 {
		s.handleLoginError(w, r, ar, errors.WithStack(schema.NewRequiredError("", gojsonschema.NewJsonContext("identifier", nil))))
		return
	}

	if len(p.Password) == 0 {
		s.handleLoginError(w, r, ar, errors.WithStack(schema.NewRequiredError("", gojsonschema.NewJsonContext("password", nil))))
		return
	}

	if err := ar.Valid(); err != nil {
		s.d.ErrorManager().ForwardError(w, r, err)
		return
	}

	i, c, err := s.d.IdentityPool().FindByCredentialsIdentifier(r.Context(), s.ID(), p.Identifier)
	if err != nil {
		s.handleLoginError(w, r, ar, errors.WithStack(schema.NewInvalidCredentialsError()))
		return
	}

	var o CredentialsConfig
	d := json.NewDecoder(bytes.NewBuffer(c.Options))
	if err := d.Decode(&o); err != nil {
		s.d.ErrorManager().ForwardError(w, r, herodot.ErrInternalServerError.WithReason("The password credentials could not be decoded properly").WithDebug(err.Error()))
		return
	}

	if err := s.d.PasswordHasher().Compare([]byte(p.Password), []byte(o.HashedPassword)); err != nil {
		s.handleLoginError(w, r, ar, errors.WithStack(schema.NewInvalidCredentialsError()))
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

	sr.Methods[CredentialsType] = &selfservice.DefaultRequestMethod{
		Method: CredentialsType,
		Config: &RequestMethodConfig{
			Action: action.String(),
			Fields: selfservice.FormFields{
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
				selfservice.CSRFTokenName: {
					Name:     selfservice.CSRFTokenName,
					Type:     "hidden",
					Required: true,
					Value:    s.cg(r),
				},
			},
		},
	}
	return nil
}
