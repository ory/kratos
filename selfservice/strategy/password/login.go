package password

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	LoginPath = "/self-service/browser/flows/login/strategies/password"
)

func (s *Strategy) RegisterLoginRoutes(r *x.RouterPublic) {
	r.POST(LoginPath, s.d.SessionHandler().IsNotAuthenticated(s.handleLogin, session.RedirectOnAuthenticated(s.c)))
}

func (s *Strategy) handleLoginError(w http.ResponseWriter, r *http.Request, rr *login.Request, err error) {
	if rr != nil {
		if method, ok := rr.Methods[identity.CredentialsTypePassword]; ok {
			method.Config.Reset()
			method.Config.SetValue("identifier", r.PostForm.Get("identifier"))
			method.Config.SetCSRF(s.cg(r))
			rr.Methods[identity.CredentialsTypePassword] = method
		}
	}

	s.d.LoginRequestErrorHandler().HandleLoginError(w, r, identity.CredentialsTypePassword, rr, err)
}

func (s *Strategy) handleLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	rid := x.ParseUUID(r.URL.Query().Get("request"))
	if x.IsZeroUUID(rid) {
		s.handleLoginError(w, r, nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("The request query parameter is missing or invalid.")))
		return
	}

	ar, err := s.d.LoginRequestPersister().GetLoginRequest(r.Context(), rid)
	if err != nil {
		s.handleLoginError(w, r, nil, err)
		return
	}

	var p LoginFormPayload
	if err := r.ParseForm(); err != nil {
		s.handleLoginError(w, r, ar, errors.WithStack(herodot.ErrBadRequest.WithDebug(err.Error()).WithReasonf("Unable to parse HTTP form request: %s", err.Error())))
		return
	}

	p.Identifier = r.PostForm.Get("identifier")
	p.Password = r.PostForm.Get("password")

	if len(p.Identifier) == 0 {
		s.handleLoginError(w, r, ar, schema.NewRequiredError("#/", "identifier"))
		return
	}

	if len(p.Password) == 0 {
		s.handleLoginError(w, r, ar, schema.NewRequiredError("#/", "password"))
		return
	}

	if err := ar.Valid(); err != nil {
		s.handleLoginError(w, r, ar, err)
		return
	}

	i, c, err := s.d.IdentityPool().FindByCredentialsIdentifier(r.Context(), s.ID(), p.Identifier)
	if err != nil {
		s.handleLoginError(w, r, ar, errors.WithStack(schema.NewInvalidCredentialsError()))
		return
	}

	var o CredentialsConfig
	d := json.NewDecoder(bytes.NewBuffer(c.Config))
	if err := d.Decode(&o); err != nil {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, herodot.ErrInternalServerError.WithReason("The password credentials could not be decoded properly").WithDebug(err.Error()))
		return
	}

	if err := s.d.PasswordHasher().Compare([]byte(p.Password), []byte(o.HashedPassword)); err != nil {
		s.handleLoginError(w, r, ar, errors.WithStack(schema.NewInvalidCredentialsError()))
		return
	}

	if err := s.d.LoginHookExecutor().PostLoginHook(w, r,
		s.d.PostLoginHooks(identity.CredentialsTypePassword), ar, i); err != nil {
		s.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}
}

func (s *Strategy) PopulateLoginMethod(r *http.Request, sr *login.Request) error {
	if err := r.ParseForm(); err != nil {
		return errors.WithStack(herodot.ErrBadRequest.WithReasonf("Unable to decode POST body: %s", err))
	}

	action := urlx.CopyWithQuery(
		urlx.AppendPaths(s.c.SelfPublicURL(), LoginPath),
		url.Values{"request": {sr.ID.String()}},
	)

	f := &form.HTMLForm{
		Action: action.String(),
		Method: "POST",
		Fields: form.Fields{
			{
				Name:     "identifier",
				Type:     "text",
				Required: true,
			},
			{
				Name:     "password",
				Type:     "password",
				Required: true,
			},
		},
	}
	f.SetCSRF(s.cg(r))

	sr.Methods[identity.CredentialsTypePassword] = &login.RequestMethod{
		Method: identity.CredentialsTypePassword,
		Config: &login.RequestMethodConfig{RequestMethodConfigurator: &RequestMethod{HTMLForm: f}},
	}
	return nil
}
