package profile

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/markbates/pkger"
	"github.com/pkg/errors"
	"github.com/tidwall/sjson"

	"github.com/ory/herodot"
	"github.com/ory/jsonschema/v3"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	RouteSettings = "/self-service/settings/methods/profile"
)

var _ settings.Strategy = new(Strategy)

type (
	strategyDependencies interface {
		x.CSRFProvider
		x.CSRFTokenGeneratorProvider
		x.WriterProvider
		x.LoggingProvider

		continuity.ManagementProvider

		session.HandlerProvider
		session.ManagementProvider

		identity.ValidationProvider
		identity.ManagementProvider
		identity.PrivilegedPoolProvider

		errorx.ManagementProvider

		settings.HookExecutorProvider
		settings.ErrorHandlerProvider
		settings.FlowPersistenceProvider
		settings.StrategyProvider
		settings.HooksProvider

		IdentityTraitsSchemas() schema.Schemas
	}
	Strategy struct {
		c  configuration.Provider
		d  strategyDependencies
		dc *decoderx.HTTP
	}
)

// swagger:model settingsProfileFormConfig
type SettingsProfileRequestMethod struct {
	*form.HTMLForm
}

func NewStrategy(d strategyDependencies, c configuration.Provider) *Strategy {
	return &Strategy{c: c, d: d, dc: decoderx.NewHTTP()}
}

func (s *Strategy) SettingsStrategyID() string {
	return settings.StrategyProfile
}

func (s *Strategy) RegisterSettingsRoutes(public *x.RouterPublic) {
	public.POST(RouteSettings, s.d.SessionHandler().IsAuthenticated(s.handleSubmit, settings.OnUnauthenticated(s.c, s.d)))
	public.GET(RouteSettings, s.d.SessionHandler().IsAuthenticated(s.handleSubmit, settings.OnUnauthenticated(s.c, s.d)))
}

func (s *Strategy) PopulateSettingsMethod(r *http.Request, id *identity.Identity, pr *settings.Flow) error {
	traitsSchema, err := s.c.IdentityTraitsSchemas().FindSchemaByID(id.SchemaID)
	if err != nil {
		return err
	}

	// use a schema compiler that disables identifiers
	schemaCompiler := jsonschema.NewCompiler()

	f, err := form.NewHTMLFormFromJSONSchema(urlx.CopyWithQuery(
		urlx.AppendPaths(s.c.SelfPublicURL(), RouteSettings),
		url.Values{"flow": {pr.ID.String()}},
	).String(), traitsSchema.URL, "", schemaCompiler)
	if err != nil {
		return err
	}

	f.SetValuesFromJSON(json.RawMessage(id.Traits), "traits")
	f.SetCSRF(s.d.GenerateCSRFToken(r))

	if err := f.SortFields(traitsSchema.URL); err != nil {
		return err
	}

	pr.Methods[s.SettingsStrategyID()] = &settings.FlowMethod{
		Method: s.SettingsStrategyID(),
		Config: &settings.FlowMethodConfig{FlowMethodConfigurator: &SettingsProfileRequestMethod{HTMLForm: f}},
	}
	return nil
}

// swagger:route POST /self-service/settings/methods/profile public completeSelfServiceSettingsFlowWithProfileMethod
//
// Complete Settings Flow with Profile Method
//
// Use this endpoint to complete a settings flow by sending an identity's updated traits. This endpoint
// behaves differently for API and browser flows.
//
// API-initiated flows expect `application/json` to be sent in the body and respond with
//   - HTTP 200 and an application/json body with the session token on success;
//   - HTTP 302 redirect to a fresh settings flow if the original flow expired with the appropriate error messages set;
//   - HTTP 400 on form validation errors.
//   - HTTP 401 when the endpoint is called without a valid session token.
//   - HTTP 403 when `selfservice.flows.settings.privileged_session_max_age` was reached and a sensitive field was
// 	   updated (e.g. recovery email). Implies that the user needs to re-authenticate.
//
// Browser flows expect `application/x-www-form-urlencoded` to be sent in the body and responds with
//   - a HTTP 302 redirect to the post/after settings URL or the `return_to` value if it was set and if the flow succeeded;
//   - a HTTP 302 redirect to the settings UI URL with the flow ID containing the validation errors otherwise.
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
//     - sessionToken
//
//     Schemes: http, https
//
//     Responses:
//       200: settingsViaApiResponse
//       200: settingsFlow
//       302: emptyResponse
//       400: settingsFlow
//       401: genericError
//       403: genericError
//       500: genericError
func (s *Strategy) handleSubmit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var p CompleteSelfServiceBrowserSettingsProfileStrategyFlow
	ctxUpdate, err := settings.PrepareUpdate(s.d, w, r, settings.ContinuityKey(s.SettingsStrategyID()), &p)
	if errors.Is(err, settings.ErrContinuePreviousAction) {
		s.continueFlow(w, r, ctxUpdate, &p)
		return
	} else if err != nil {
		s.handleSettingsError(w, r, ctxUpdate, nil, &p, err)
		return
	}

	option, err := s.newSettingsProfileDecoder(ctxUpdate.Session.Identity)
	if err != nil {
		s.handleSettingsError(w, r, ctxUpdate, nil, &p, err)
		return
	}

	if err := s.dc.Decode(r, &p, option,
		decoderx.HTTPDecoderSetValidatePayloads(false),
		decoderx.HTTPDecoderJSONFollowsFormFormat(),
	); err != nil {
		s.handleSettingsError(w, r, ctxUpdate, nil, &p, err)
		return
	}

	// Reset after decoding form
	p.SetFlowID(ctxUpdate.Flow.ID)

	s.continueFlow(w, r, ctxUpdate, &p)
}

func (s *Strategy) continueFlow(w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p *CompleteSelfServiceBrowserSettingsProfileStrategyFlow) {
	if err := flow.VerifyRequest(r, ctxUpdate.Flow.Type, s.d.GenerateCSRFToken, p.CSRFToken); err != nil {
		s.handleSettingsError(w, r, ctxUpdate, nil, p, err)
		return
	}

	if len(p.Traits) == 0 {
		s.handleSettingsError(w, r, ctxUpdate, nil, p, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Did not receive any value changes.")))
		return
	}

	if err := s.hydrateForm(r, ctxUpdate.Flow, ctxUpdate.Session, p.Traits); err != nil {
		s.d.SettingsFlowErrorHandler().WriteFlowError(w, r, settings.StrategyProfile, ctxUpdate.Flow, ctxUpdate.Session.Identity, err)
		return
	}

	update, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), ctxUpdate.Session.Identity.ID)
	if err != nil {
		s.d.SettingsFlowErrorHandler().WriteFlowError(w, r, settings.StrategyProfile, ctxUpdate.Flow, ctxUpdate.Session.Identity, err)
		return
	}

	update.Traits = identity.Traits(p.Traits)
	if err := s.d.SettingsHookExecutor().PostSettingsHook(w, r,
		settings.StrategyProfile, ctxUpdate, update); err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p.Traits, p, err)
		return
	}
}

// Complete profile update payload
//
// swagger:parameters completeSelfServiceBrowserSettingsProfileStrategyFlow
// nolint:deadcode,unused
type completeSelfServiceBrowserSettingsProfileStrategyFlowParameters struct {
	// Request is the request ID.
	//
	// required: true
	// in: query
	// format: uuid
	Flow string `json:"flow"`

	// in: body
	// required: true
	Body CompleteSelfServiceBrowserSettingsProfileStrategyFlow
}

// nolint:deadcode,unused
type CompleteSelfServiceBrowserSettingsProfileStrategyFlow struct {
	// Traits contains all of the identity's traits.
	Traits json.RawMessage `json:"traits"`

	// FlowIDRequestID is the flow ID.
	//
	// swagger:ignore
	FlowID string `json:"flow"`

	// The Anti-CSRF Token
	//
	// This token is only required when performing browser flows.
	CSRFToken string `json:"csrf_token"`
}

func (p *CompleteSelfServiceBrowserSettingsProfileStrategyFlow) GetFlowID() uuid.UUID {
	return x.ParseUUID(p.FlowID)
}

func (p *CompleteSelfServiceBrowserSettingsProfileStrategyFlow) SetFlowID(rid uuid.UUID) {
	p.FlowID = rid.String()
}

func (s *Strategy) hydrateForm(r *http.Request, ar *settings.Flow, ss *session.Session, traits json.RawMessage) error {
	action := urlx.CopyWithQuery(urlx.AppendPaths(s.c.SelfPublicURL(), RouteSettings),
		url.Values{"flow": {ar.ID.String()}})

	ar.Methods[settings.StrategyProfile].Config.Reset()
	if traits != nil {
		for _, field := range form.NewHTMLFormFromJSON(action.String(), traits, "traits").Fields {
			ar.Methods[settings.StrategyProfile].Config.SetField(field)
		}
	}
	ar.Methods[settings.StrategyProfile].Config.SetCSRF(s.d.GenerateCSRFToken(r))

	traitsSchema, err := s.c.IdentityTraitsSchemas().FindSchemaByID(ss.Identity.SchemaID)
	if err != nil {
		return err
	}

	if err = ar.Methods[settings.StrategyProfile].Config.SortFields(traitsSchema.URL); err != nil {
		return err
	}

	return nil
}

// handleSettingsError is a convenience function for handling all types of errors that may occur (e.g. validation error)
// during a settings request.
func (s *Strategy) handleSettingsError(w http.ResponseWriter, r *http.Request, puc *settings.UpdateContext, traits json.RawMessage, p *CompleteSelfServiceBrowserSettingsProfileStrategyFlow, err error) {
	if e := new(settings.FlowNeedsReAuth); errors.As(err, &e) {
		if err := s.d.ContinuityManager().Pause(r.Context(), w, r,
			settings.ContinuityKey(s.SettingsStrategyID()),
			settings.ContinuityOptions(p, puc.Session.Identity)...); err != nil {
			s.d.SettingsFlowErrorHandler().WriteFlowError(w, r, s.SettingsStrategyID(), puc.Flow, puc.Session.Identity, err)
			return
		}
	}

	if puc.Flow != nil {
		if traits == nil {
			traits = json.RawMessage(puc.GetSessionIdentity().Traits)
		}

		if err := s.hydrateForm(r, puc.Flow, puc.Session, traits); err != nil {
			s.d.SettingsFlowErrorHandler().WriteFlowError(w, r, s.SettingsStrategyID(), puc.Flow, puc.GetSessionIdentity(), err)
			return
		}
	}

	s.d.SettingsFlowErrorHandler().WriteFlowError(w, r, s.SettingsStrategyID(), puc.Flow, puc.GetSessionIdentity(), err)
}

// newSettingsProfileDecoder returns a decoderx.HTTPDecoderOption with a JSON Schema for type assertion and
// validation.
func (s *Strategy) newSettingsProfileDecoder(i *identity.Identity) (decoderx.HTTPDecoderOption, error) {
	ss, err := s.d.IdentityTraitsSchemas().GetByID(i.SchemaID)
	if err != nil {
		return nil, err
	}
	raw, err := sjson.SetBytes(x.MustPkgerRead(pkger.Open(
		"/selfservice/strategy/password/.schema/settings.schema.json")),
		"properties.traits.$ref", ss.URL.String()+"#/properties/traits")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	o, err := decoderx.HTTPRawJSONSchemaCompiler(raw)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return o, nil
}
