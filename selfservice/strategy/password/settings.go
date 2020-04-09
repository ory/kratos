package password

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/x/sqlxx"

	"github.com/ory/herodot"
	"github.com/ory/x/errorsx"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	SettingsPath          = "/self-service/browser/flows/settings/strategies/password"
	continuityKeySettings = "settings_password"
)

func (s *Strategy) RegisterSettingsRoutes(router *x.RouterPublic) {
	router.POST(SettingsPath, s.submitSettingsFlow)
	router.GET(SettingsPath, s.submitSettingsFlow)
}

func (s *Strategy) SettingsStrategyID() string {
	return identity.CredentialsTypePassword.String()
}

// swagger:model completeSelfServiceBrowserSettingsPasswordFlowPayload
type completeSelfServiceBrowserSettingsPasswordFlowPayload struct {
	// Password is the updated password
	//
	// type: string
	// in: body
	// required: true
	Password string `json:"password"`

	// RequestID is request ID.
	//
	// in: query
	RequestID string `json:"request_id"`
}

// swagger:route POST /self-service/browser/flows/settings/strategies/password public completeSelfServiceBrowserSettingsPasswordStrategyFlow
//
// Complete the browser-based settings flow for the password strategy
//
// This endpoint completes a browser-based settings flow. This is usually achieved by POSTing data to this
// endpoint.
//
// > This endpoint is NOT INTENDED for API clients and only works with browsers (Chrome, Firefox, ...) and HTML Forms.
//
// More information can be found at [ORY Kratos User Settings & Profile Management Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-settings-profile-management).
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
func (s *Strategy) submitSettingsFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ss, err := s.d.SessionManager().FetchFromRequest(r.Context(), w, r)
	if err != nil {
		s.handleSettingsError(w, r, nil, ss, nil, err)
		return
	}

	var p completeSelfServiceBrowserSettingsPasswordFlowPayload
	if _, err := s.d.ContinuityManager().Continue(r.Context(), r,
		continuityKeySettings,
		continuity.WithIdentity(ss.Identity),
		continuity.WithPayload(&p)); err == nil {
		s.completeSettingsFlow(w, r, ss, &p)
		return
	}

	rid := r.URL.Query().Get("request")
	if len(rid) == 0 {
		s.handleSettingsError(w, r, nil, ss, &p, errors.WithStack(herodot.ErrBadRequest.WithReasonf("The request query parameter is missing.")))
		return
	}

	if err := r.ParseForm(); err != nil {
		s.handleSettingsError(w, r, nil, ss, &p, errors.WithStack(err))
		return
	}

	p = completeSelfServiceBrowserSettingsPasswordFlowPayload{
		RequestID: rid,
		Password:  r.PostFormValue("password"),
	}
	if err := s.d.ContinuityManager().Pause(
		r.Context(), w, r,
		continuityKeySettings,
		continuity.WithPayload(&p),
		continuity.WithIdentity(ss.Identity),
		continuity.WithLifespan(time.Minute*15),
	); err != nil {
		s.handleSettingsError(w, r, nil, ss, &p, errors.WithStack(err))
		return
	}

	s.completeSettingsFlow(w, r, ss, &p)
}

func (s *Strategy) completeSettingsFlow(
	w http.ResponseWriter, r *http.Request,
	ss *session.Session, p *completeSelfServiceBrowserSettingsPasswordFlowPayload,
) {

	ar, err := s.d.SettingsRequestPersister().GetSettingsRequest(r.Context(), x.ParseUUID(p.RequestID))
	if err != nil {
		s.handleSettingsError(w, r, nil, ss, p, err)
		return
	}

	if err := ar.Valid(ss); err != nil {
		s.handleSettingsError(w, r, ar, ss, p, err)
		return
	}

	if ss.AuthenticatedAt.Add(s.c.SelfServicePrivilegedSessionMaxAge()).Before(time.Now()) {
		s.handleSettingsError(w, r, nil, ss, p, errors.WithStack(settings.ErrRequestNeedsReAuthentication))
		return
	}

	if len(p.Password) == 0 {
		s.handleSettingsError(w, r, ar, ss, p, schema.NewRequiredError("#/", "password"))
		return
	}

	hpw, err := s.d.PasswordHasher().Generate([]byte(p.Password))
	if err != nil {
		s.handleSettingsError(w, r, ar, ss, p, err)
		return
	}

	co, err := json.Marshal(&CredentialsConfig{HashedPassword: string(hpw)})
	if err != nil {
		s.handleSettingsError(w, r, ar, ss, p, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to encode password options to JSON: %s", err)))
		return
	}

	i, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), ss.Identity.ID)
	if err != nil {
		s.handleSettingsError(w, r, ar, ss, p, err)
		return
	}

	c, ok := i.GetCredentials(s.ID())
	if !ok {
		c = &identity.Credentials{Type: s.ID(),
			// Prevent duplicates
			Identifiers: []string{x.NewUUID().String()}}
	}

	c.Config = co
	i.SetCredentials(s.ID(), *c)
	if err := s.validateCredentials(i, p.Password); err != nil {
		s.handleSettingsError(w, r, ar, ss, p, err)
		return
	}

	if err := s.d.SettingsExecutor().PostSettingsHook(w, r,
		s.d.PostSettingsHooks(s.SettingsStrategyID()),
		ar, ss, i,
	); errorsx.Cause(err) == settings.ErrHookAbortRequest {
		return
	} else if err != nil {
		s.handleSettingsError(w, r, ar, ss, p, err)
		return
	}

	if len(w.Header().Get("Location")) == 0 {
		http.Redirect(w, r,
			urlx.CopyWithQuery(s.c.SettingsURL(), url.Values{"request": {ar.ID.String()}}).String(),
			http.StatusFound,
		)
	}
}

func (s *Strategy) PopulateSettingsMethod(r *http.Request, ss *session.Session, pr *settings.Request) error {
	f := &form.HTMLForm{
		Action: urlx.CopyWithQuery(urlx.AppendPaths(s.c.SelfPublicURL(), SettingsPath),
			url.Values{"request": {pr.ID.String()}},
		).String(),
		Fields: form.Fields{{Name: "password", Type: "password", Required: true}}, Method: "POST",
	}
	f.SetCSRF(s.d.GenerateCSRFToken(r))

	pr.Methods[string(s.ID())] = &settings.RequestMethod{
		Method: string(s.ID()),
		Config: &settings.RequestMethodConfig{RequestMethodConfigurator: &RequestMethod{HTMLForm: f}},
	}
	pr.Active = sqlxx.NullString(s.ID())
	return nil
}

func (s *Strategy) handleSettingsError(w http.ResponseWriter, r *http.Request, rr *settings.Request, ss *session.Session, p *completeSelfServiceBrowserSettingsPasswordFlowPayload, err error) {
	if errors.Is(err, settings.ErrRequestNeedsReAuthentication) {
		if _, err := s.d.ContinuityManager().Continue(r.Context(), r,
			continuityKeySettings,
			continuity.WithIdentity(ss.Identity),
			continuity.WithPayload(&p),
		); err != nil {
			s.d.SettingsRequestErrorHandler().HandleSettingsError(w, r, rr, err, s.SettingsStrategyID())
			return
		}
	}

	if rr != nil {
		rr.Methods[s.SettingsStrategyID()].Config.Reset()
		rr.Methods[s.SettingsStrategyID()].Config.SetCSRF(s.d.GenerateCSRFToken(r))
	}

	s.d.SettingsRequestErrorHandler().HandleSettingsError(w, r, rr, err, s.SettingsStrategyID())
}
