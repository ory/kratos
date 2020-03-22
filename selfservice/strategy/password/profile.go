package password

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/errorsx"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow/profile"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	ProfilePath = "/self-service/browser/flows/profile/strategies/password"
)

func (s *Strategy) RegisterProfileManagementRoutes(router *x.RouterPublic) {
	router.POST(ProfilePath, s.completeProfileManagementFlow)
}

func (s *Strategy) ProfileManagementStrategyID() string {
	return string(identity.CredentialsTypePassword)
}

// swagger:route POST /self-service/browser/flows/profile/strategies/profile public completeSelfServiceBrowserProfileManagementPasswordStrategyFlow
//
// Complete the browser-based profile management flow for the password strategy
//
// This endpoint completes a browser-based profile management flow. This is usually achieved by POSTing data to this
// endpoint.
//
// > This endpoint is NOT INTENDED for API clients and only works with browsers (Chrome, Firefox, ...) and HTML Forms.
//
// More information can be found at [ORY Kratos Profile Management Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-profile-management).
//
//     Consumes:
//     - application/json
//     - application/x-www-form-urlencoded
//
//     Schemes: http, https
//
//     Responses:
//       302: emptyResponse
//       500: genericError
func (s *Strategy) completeProfileManagementFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ss, err := s.d.SessionManager().FetchFromRequest(r.Context(), w, r)
	if err != nil {
		s.handleProfileManagementError(w, r, nil, nil, err)
		return
	}

	rid := r.URL.Query().Get("request")
	if len(rid) == 0 {
		s.handleProfileManagementError(w, r, nil, ss.Identity.Traits, errors.WithStack(herodot.ErrBadRequest.WithReasonf("The request query parameter is missing.")))
		return
	}

	ar, err := s.d.ProfileRequestPersister().GetProfileRequest(r.Context(), x.ParseUUID(rid))
	if err != nil {
		s.handleProfileManagementError(w, r, nil, ss.Identity.Traits, err)
		return
	}

	if err := ar.Valid(ss); err != nil {
		s.handleProfileManagementError(w, r, ar, ss.Identity.Traits, err)
		return
	}

	if ss.AuthenticatedAt.Add(s.c.SelfServicePrivilegedSessionMaxAge()).Before(time.Now()) {
		s.handleProfileManagementError(w, r, nil, ss.Identity.Traits, errors.WithStack(herodot.ErrForbidden.WithReasonf("The session is too old and thus not allowed to update the password.")))
		return
	}

	if err := r.ParseForm(); err != nil {
		s.handleProfileManagementError(w, r, ar, ss.Identity.Traits, errors.WithStack(err))
		return
	}

	password := r.PostForm.Get("password")
	if len(password) == 0 {
		s.handleProfileManagementError(w, r, ar, ss.Identity.Traits, schema.NewRequiredError("#/", "password"))
		return
	}

	hpw, err := s.d.PasswordHasher().Generate([]byte(password))
	if err != nil {
		s.handleProfileManagementError(w, r, ar, ss.Identity.Traits, err)
		return
	}

	co, err := json.Marshal(&CredentialsConfig{HashedPassword: string(hpw)})
	if err != nil {
		s.handleProfileManagementError(w, r, ar, ss.Identity.Traits, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to encode password options to JSON: %s", err)))
		return
	}

	i, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), ss.Identity.ID)
	if err != nil {
		s.handleProfileManagementError(w, r, ar, ss.Identity.Traits, err)
		return
	}

	c, ok := i.GetCredentials(s.ID())
	if !ok {
		s.handleProfileManagementError(w, r, ar, ss.Identity.Traits, errors.WithStack(herodot.ErrBadRequest.WithReasonf("A password change request was made but this identity does not use a password flow.")))
		return
	}

	c.Config = co
	i.SetCredentials(s.ID(), *c)
	if err := s.validateCredentials(i, password); err != nil {
		s.handleProfileManagementError(w, r, ar, ss.Identity.Traits, err)
		return
	}

	if err := s.d.ProfileManagementExecutor().PostProfileManagementHook(w, r,
		s.d.PostProfileManagementHooks(profile.StrategyTraitsID),
		ar, ss, i,
	); errorsx.Cause(err) == profile.ErrHookAbortRequest {
		return
	} else if err != nil {
		s.handleProfileManagementError(w, r, ar, ss.Identity.Traits, err)
		return
	}

	if len(w.Header().Get("Location")) == 0 {
		http.Redirect(w, r,
			urlx.CopyWithQuery(s.c.ProfileURL(), url.Values{"request": {ar.ID.String()}}).String(),
			http.StatusFound,
		)
	}
}

func (s *Strategy) PopulateProfileManagementMethod(r *http.Request, ss *session.Session, pr *profile.Request) error {
	i, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), ss.Identity.ID)
	if err != nil {
		return err
	}

	if _, ok := i.GetCredentials(identity.CredentialsTypePassword); !ok {
		// Ignore this method if no password flow is set up.
		return nil
	}

	f := &form.HTMLForm{
		Action: urlx.CopyWithQuery(
			urlx.AppendPaths(s.c.SelfPublicURL(), ProfilePath),
			url.Values{"request": {pr.ID.String()}},
		).String(),
		Method: "POST",
		Fields: form.Fields{
			{
				Name:     "password",
				Type:     "password",
				Required: true,
			},
		},
	}
	f.SetCSRF(s.d.GenerateCSRFToken(r))

	pr.Methods[string(s.ID())] = &profile.RequestMethod{
		Method: string(s.ID()),
		Config: &profile.RequestMethodConfig{RequestMethodConfigurator: &RequestMethod{HTMLForm: f}},
	}
	return nil
}

func (s *Strategy) handleProfileManagementError(w http.ResponseWriter, r *http.Request, rr *profile.Request, traits identity.Traits, err error) {
	if rr != nil {
		rr.Methods[s.ProfileManagementStrategyID()].Config.Reset()
	}

	s.d.ProfileRequestRequestErrorHandler().HandleProfileManagementError(w, r, rr, err, string(identity.CredentialsTypePassword))
}
