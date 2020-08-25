package password

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/ory/herodot"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/urlx"
	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/x"
)

const (
	RouteSettings = "/self-service/settings/methods/password"
)

func (s *Strategy) RegisterSettingsRoutes(router *x.RouterPublic) {
	s.d.CSRFHandler().ExemptPath(RouteSettings)
	router.POST(RouteSettings, s.submitSettingsFlow)
	router.GET(RouteSettings, s.submitSettingsFlow)
}

func (s *Strategy) SettingsStrategyID() string {
	return identity.CredentialsTypePassword.String()
}

// swagger:parameters completeSelfServiceSettingsFlowWithPasswordMethod
type completeSelfServiceSettingsFlowWithPasswordMethod struct {
	// in: body
	Payload SettingsFlowPayload

	// Flow is flow ID.
	//
	// in: query
	Flow string `json:"flow"`
}

type SettingsFlowPayload struct {
	// Password is the updated password
	//
	// type: string
	// required: true
	Password string `json:"password"`

	// CSRFToken is the anti-CSRF token
	//
	// type: string
	CSRFToken string `json:"csrf_token"`

	// Flow is flow ID.
	//
	// swagger:ignore
	Flow string `json:"flow"`
}

func (p *SettingsFlowPayload) GetFlowID() uuid.UUID {
	return x.ParseUUID(p.Flow)
}

func (p *SettingsFlowPayload) SetFlowID(rid uuid.UUID) {
	p.Flow = rid.String()
}

// swagger:route POST /self-service/settings/methods/password public completeSelfServiceSettingsFlowWithPasswordMethod
//
// Complete the browser-based settings flow for the password strategy
//
// This endpoint completes a browser-based settings flow. This is usually achieved by POSTing data to this
// endpoint.
//
// > This endpoint is NOT INTENDED for API clients and only works with browsers (Chrome, Firefox, ...) and HTML Forms.
//
// More information can be found at [ORY Kratos User Settings & Profile Management Documentation](../self-service/flows/user-settings).
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
	var p SettingsFlowPayload
	ctxUpdate, err := settings.PrepareUpdate(s.d, w, r, settings.ContinuityKey(s.SettingsStrategyID()), &p)
	if errors.Is(err, settings.ErrContinuePreviousAction) {
		s.continueSettingsFlow(w, r, ctxUpdate, &p)
		return
	} else if err != nil {
		s.handleSettingsError(w, r, ctxUpdate, &p, err)
		return
	}

	if err := s.decodeSettingsFlow(r, &p); err != nil {
		s.handleSettingsError(w, r, ctxUpdate, &p, err)
		return
	}

	// This does not come from the payload!
	p.Flow = ctxUpdate.Flow.ID.String()
	s.continueSettingsFlow(w, r, ctxUpdate, &p)
}

func (s *Strategy) decodeSettingsFlow(r *http.Request, dest interface{}) error {
	compiler, err := decoderx.HTTPRawJSONSchemaCompiler(settingsSchema)
	if err != nil {
		return errors.WithStack(err)
	}

	return decoderx.NewHTTP().Decode(r, dest, compiler,
		decoderx.HTTPDecoderSetValidatePayloads(false),
		decoderx.HTTPDecoderJSONFollowsFormFormat(),
	)
}

func (s *Strategy) continueSettingsFlow(
	w http.ResponseWriter, r *http.Request,
	ctxUpdate *settings.UpdateContext, p *SettingsFlowPayload,
) {
	if err := flow.VerifyRequest(r,ctxUpdate.Flow.Type,s.d.GenerateCSRFToken,p.CSRFToken); err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p, err)
		return
	}

	if ctxUpdate.Session.AuthenticatedAt.Add(s.c.SelfServiceFlowSettingsPrivilegedSessionMaxAge()).Before(time.Now()) {
		s.handleSettingsError(w, r, ctxUpdate, p, errors.WithStack(settings.NewFlowNeedsReAuth()))
		return
	}

	if len(p.Password) == 0 {
		s.handleSettingsError(w, r, ctxUpdate, p, schema.NewRequiredError("#/password", "password"))
		return
	}

	hpw, err := s.d.Hasher().Generate([]byte(p.Password))
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

func (s *Strategy) PopulateSettingsMethod(r *http.Request, _ *identity.Identity, f *settings.Flow) error {
	hf := &form.HTMLForm{Action: urlx.CopyWithQuery(urlx.AppendPaths(s.c.SelfPublicURL(), RouteSettings),
		url.Values{"flow": {f.ID.String()}}).String(), Fields: form.Fields{{Name: "password",
			Type: "password", Required: true}}, Method: "POST"}
	hf.SetCSRF(s.d.GenerateCSRFToken(r))

	f.Methods[string(s.ID())] = &settings.FlowMethod{
		Method: string(s.ID()),
		Config: &settings.FlowMethodConfig{FlowMethodConfigurator: &FlowMethod{HTMLForm: hf}},
	}
	return nil
}

func (s *Strategy) handleSettingsError(w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p *SettingsFlowPayload, err error) {
	// Do not pause flow if the flow type is an API flow as we can't save cookies in those flows.
	if e := new(settings.FlowNeedsReAuth); errors.As(err, &e) && ctxUpdate.Flow != nil && ctxUpdate.Flow.Type == flow.TypeBrowser {
		if err := s.d.ContinuityManager().Pause(r.Context(), w, r,
			settings.ContinuityKey(s.SettingsStrategyID()), settings.ContinuityOptions(p, ctxUpdate.Session.Identity)...); err != nil {
			s.d.SettingsFlowErrorHandler().WriteFlowError(w, r, s.SettingsStrategyID(), ctxUpdate.Flow, ctxUpdate.Session.Identity, err)
			return
		}
	}

	var id *identity.Identity
	if ctxUpdate.Flow != nil {
		ctxUpdate.Flow.Methods[s.SettingsStrategyID()].Config.Reset()
		ctxUpdate.Flow.Methods[s.SettingsStrategyID()].Config.SetCSRF(s.d.GenerateCSRFToken(r))
		id = ctxUpdate.Session.Identity
	}

	s.d.SettingsFlowErrorHandler().WriteFlowError(w, r, s.SettingsStrategyID(), ctxUpdate.Flow, id, err)
}
