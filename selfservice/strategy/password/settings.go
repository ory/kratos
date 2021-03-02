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
	"github.com/ory/x/decoderx"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/selfservice/strategy"
	"github.com/ory/kratos/x"
)

const (
	RouteSettings = "/self-service/settings/methods/password"
)

func (s *Strategy) RegisterSettingsRoutes(router *x.RouterPublic) {
	s.d.CSRFHandler().IgnorePath(RouteSettings)

	wrappedSubmmitSettingsFlow := strategy.IsDisabled(s.d, s.SettingsStrategyID(), s.submitSettingsFlow)
	router.POST(RouteSettings, wrappedSubmmitSettingsFlow)
	router.GET(RouteSettings, wrappedSubmmitSettingsFlow)
}

func (s *Strategy) SettingsStrategyID() string {
	return identity.CredentialsTypePassword.String()
}

// nolint:deadcode,unused
// swagger:parameters completeSelfServiceSettingsFlowWithPasswordMethod
type completeSelfServiceSettingsFlowWithPasswordMethod struct {
	// in: body
	Body CompleteSelfServiceSettingsFlowWithPasswordMethod

	// Flow is flow ID.
	//
	// in: query
	Flow string `json:"flow"`
}

type CompleteSelfServiceSettingsFlowWithPasswordMethod struct {
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

func (p *CompleteSelfServiceSettingsFlowWithPasswordMethod) GetFlowID() uuid.UUID {
	return x.ParseUUID(p.Flow)
}

func (p *CompleteSelfServiceSettingsFlowWithPasswordMethod) SetFlowID(rid uuid.UUID) {
	p.Flow = rid.String()
}

// swagger:route POST /self-service/settings/methods/password public completeSelfServiceSettingsFlowWithPasswordMethod
//
// Complete Settings Flow with Username/Email Password Method
//
// Use this endpoint to complete a settings flow by sending an identity's updated password. This endpoint
// behaves differently for API and browser flows.
//
// API-initiated flows expect `application/json` to be sent in the body and respond with
//   - HTTP 200 and an application/json body with the session token on success;
//   - HTTP 302 redirect to a fresh settings flow if the original flow expired with the appropriate error messages set;
//   - HTTP 400 on form validation errors.
//   - HTTP 401 when the endpoint is called without a valid session token.
//   - HTTP 403 when `selfservice.flows.settings.privileged_session_max_age` was reached.
//     Implies that the user needs to re-authenticate.
//
// Browser flows expect `application/x-www-form-urlencoded` to be sent in the body and responds with
//   - a HTTP 302 redirect to the post/after settings URL or the `return_to` value if it was set and if the flow succeeded;
//   - a HTTP 302 redirect to the Settings UI URL with the flow ID containing the validation errors otherwise.
//   - a HTTP 302 redirect to the login endpoint when `selfservice.flows.settings.privileged_session_max_age` was reached.
//
// More information can be found at [ORY Kratos User Settings & Profile Management Documentation](../self-service/flows/user-settings).
//
//     Consumes:
//     - application/json
//     - application/x-www-form-urlencoded
//
//     Produces:
//     - application/json
//
//     Security:
//       sessionToken:
//
//     Schemes: http, https
//
//     Responses:
//       200: settingsViaApiResponse
//       302: emptyResponse
//       400: settingsFlow
//       401: genericError
//       403: genericError
//       500: genericError
func (s *Strategy) submitSettingsFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var p CompleteSelfServiceSettingsFlowWithPasswordMethod
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
	ctxUpdate *settings.UpdateContext, p *CompleteSelfServiceSettingsFlowWithPasswordMethod,
) {
	if err := flow.VerifyRequest(r, ctxUpdate.Flow.Type, s.d.Config(r.Context()).DisableAPIFlowEnforcement(), s.d.GenerateCSRFToken, p.CSRFToken); err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p, err)
		return
	}

	if ctxUpdate.Session.AuthenticatedAt.Add(s.d.Config(r.Context()).SelfServiceFlowSettingsPrivilegedSessionMaxAge()).Before(time.Now()) {
		s.handleSettingsError(w, r, ctxUpdate, p, errors.WithStack(settings.NewFlowNeedsReAuth()))
		return
	}

	if len(p.Password) == 0 {
		s.handleSettingsError(w, r, ctxUpdate, p, schema.NewRequiredError("#/password", "password"))
		return
	}

	hpw, err := s.d.Hasher().Generate(r.Context(), []byte(p.Password))
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
	if err := s.validateCredentials(r.Context(), i, p.Password); err != nil {
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
	hf := &form.HTMLForm{Action: urlx.CopyWithQuery(urlx.AppendPaths(s.d.Config(r.Context()).SelfPublicURL(r), RouteSettings),
		url.Values{"flow": {f.ID.String()}}).String(), Fields: form.Fields{{Name: "password",
		Type: "password", Required: true}}, Method: "POST"}
	hf.SetCSRF(s.d.GenerateCSRFToken(r))

	f.Methods[string(s.ID())] = &settings.FlowMethod{
		Method: string(s.ID()),
		Config: &settings.FlowMethodConfig{FlowMethodConfigurator: &FlowMethod{HTMLForm: hf}},
	}
	return nil
}

func (s *Strategy) handleSettingsError(w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p *CompleteSelfServiceSettingsFlowWithPasswordMethod, err error) {
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
