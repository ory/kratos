package password

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	SettingsPath = "/self-service/browser/flows/settings/strategies/password"

	continuityPrefix = "ory_kratos_settings_password"
)

func continuityKeySettings(rid string) string {
	// Use one individual container per request ID to prevent resuming other request IDs.
	return continuityPrefix + "." + rid
}

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

func (p *completeSelfServiceBrowserSettingsPasswordFlowPayload) GetRequestID() uuid.UUID {
	return x.ParseUUID(p.RequestID)
}

func (p *completeSelfServiceBrowserSettingsPasswordFlowPayload) SetRequestID(rid uuid.UUID) {
	p.RequestID = rid.String()
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
	var p completeSelfServiceBrowserSettingsPasswordFlowPayload
	ctxUpdate, err := settings.PrepareUpdate(s.d, r, continuityPrefix, &p)
	if errors.Is(err, settings.ErrContinuePreviousAction) {
		s.continueSettingsFlow(w, r, ctxUpdate, &p)
		return
	} else if err != nil {
		s.handleSettingsError(w, r, ctxUpdate, &p, err)
		return
	}

	p.RequestID = ctxUpdate.Request.ID.String()
	p.Password = r.PostFormValue("password")
	s.continueSettingsFlow(w, r, ctxUpdate, &p)
}

func (s *Strategy) continueSettingsFlow(
	w http.ResponseWriter, r *http.Request,
	ctxUpdate *settings.UpdateContext, p *completeSelfServiceBrowserSettingsPasswordFlowPayload,
) {
	if ctxUpdate.Session.AuthenticatedAt.Add(s.c.SelfServicePrivilegedSessionMaxAge()).Before(time.Now()) {
		s.handleSettingsError(w, r, ctxUpdate, p, errors.WithStack(settings.ErrRequestNeedsReAuthentication))
		return
	}

	if len(p.Password) == 0 {
		s.handleSettingsError(w, r, ctxUpdate, p, schema.NewRequiredError("#/", "password"))
		return
	}

	hpw, err := s.d.PasswordHasher().Generate([]byte(p.Password))
	if err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p, err)
		return
	}

	co, err := json.Marshal(&CredentialsConfig{HashedPassword: string(hpw)})
	if err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to encode password options to JSON: %s", err)))
		return
	}

	i, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), ctxUpdate.Session.Identity.ID)
	if err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p, err)
		return
	}

	c, ok := i.GetCredentials(s.ID())
	if !ok {
		c = &identity.Credentials{Type: s.ID(),
			// We need to insert a random identifier now...
			Identifiers: []string{x.NewUUID().String()}}
	}

	c.Config = co
	i.SetCredentials(s.ID(), *c)
	if err := s.validateCredentials(i, p.Password); err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p, err)
		return
	}

	if err := s.d.SettingsHookExecutor().PostSettingsHook(w, r,
		s.SettingsStrategyID(), ctxUpdate, i); err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p, err)
		return
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
	return nil
}

func (s *Strategy) handleSettingsError(w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p *completeSelfServiceBrowserSettingsPasswordFlowPayload, err error) {
	if errors.Is(err, settings.ErrRequestNeedsReAuthentication) {
		if err := s.d.ContinuityManager().Pause(r.Context(), w, r,
			continuityKeySettings(r.URL.Query().Get("request")), settings.ContinuityOptions(p, ctxUpdate.Session.Identity)...); err != nil {
			s.d.SettingsRequestErrorHandler().HandleSettingsError(w, r, ctxUpdate.Request, err, s.SettingsStrategyID())
			return
		}
	}

	if ctxUpdate.Request != nil {
		ctxUpdate.Request.Methods[s.SettingsStrategyID()].Config.Reset()
		ctxUpdate.Request.Methods[s.SettingsStrategyID()].Config.SetCSRF(s.d.GenerateCSRFToken(r))
	}

	s.d.SettingsRequestErrorHandler().HandleSettingsError(w, r, ctxUpdate.Request, err, s.SettingsStrategyID())
}
