package settings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
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
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	StrategyProfile           = "profile"
	PublicSettingsProfilePath = "/self-service/browser/flows/settings/strategies/profile"

	continuityPrefix = "ory_kratos_settings_profile"
)

func strategyProfileContinuityNameFromRequest(r *http.Request) string {
	// Use one individual container per request ID to prevent resuming other request IDs.
	return strategyProfileContinuityName(r.URL.Query().Get("request"))
}

func strategyProfileContinuityName(rid string) string {
	// Use one individual container per request ID to prevent resuming other request IDs.
	return continuityPrefix + "." + rid
}

var _ Strategy = new(StrategyTraits)

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

		HookExecutorProvider
		ErrorHandlerProvider
		RequestPersistenceProvider
		StrategyProvider
		HooksProvider

		IdentityTraitsSchemas() schema.Schemas
	}
	StrategyTraits struct {
		c configuration.Provider
		d strategyDependencies
	}
)

// swagger:model traitsFormConfig
type TraitsRequestMethod struct {
	*form.HTMLForm
}

func NewStrategyTraits(d strategyDependencies, c configuration.Provider) *StrategyTraits {
	return &StrategyTraits{c: c, d: d}
}

func (s *StrategyTraits) SettingsStrategyID() string {
	return StrategyProfile
}

func (s *StrategyTraits) RegisterSettingsRoutes(public *x.RouterPublic) {
	redirect := session.RedirectOnUnauthenticated(s.c.LoginURL().String())
	public.POST(PublicSettingsProfilePath, s.d.SessionHandler().IsAuthenticated(s.handleSubmit, redirect))
	public.GET(PublicSettingsProfilePath, s.d.SessionHandler().IsAuthenticated(s.handleSubmit, redirect))
}

func (s *StrategyTraits) PopulateSettingsMethod(r *http.Request, ss *session.Session, pr *Request) error {
	traitsSchema, err := s.c.IdentityTraitsSchemas().FindSchemaByID(ss.Identity.TraitsSchemaID)
	if err != nil {
		return err
	}

	// use a schema compiler that disables identifiers
	schemaCompiler := jsonschema.NewCompiler()

	f, err := form.NewHTMLFormFromJSONSchema(urlx.CopyWithQuery(
		urlx.AppendPaths(s.c.SelfPublicURL(), PublicSettingsProfilePath),
		url.Values{"request": {pr.ID.String()}},
	).String(), traitsSchema.URL, "traits", schemaCompiler)
	if err != nil {
		return err
	}

	f.SetValuesFromJSON(json.RawMessage(ss.Identity.Traits), "traits")
	f.SetCSRF(s.d.GenerateCSRFToken(r))

	if err := f.SortFields(traitsSchema.URL, "traits"); err != nil {
		return err
	}

	pr.Methods[s.SettingsStrategyID()] = &RequestMethod{
		Method: s.SettingsStrategyID(),
		Config: &RequestMethodConfig{RequestMethodConfigurator: &TraitsRequestMethod{HTMLForm: f}},
	}
	return nil
}

// swagger:route POST /self-service/browser/flows/settings/strategies/profile public completeSelfServiceBrowserSettingsProfileStrategyFlow
//
// Complete the browser-based settings flow for profile data
//
// This endpoint completes a browser-based settings flow. This is usually achieved by POSTing data to this
// endpoint.
//
// If the provided profile data is valid against the Identity's Traits JSON Schema, the data will be updated and
// the browser redirected to `url.settings_ui` for further steps.
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
func (s *StrategyTraits) handleSubmit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var p completeSelfServiceBrowserSettingsStrategyProfileFlowPayload
	ctxUpdate, err := PrepareUpdate(s.d, r, continuityPrefix, &p)
	if errors.Is(err, ErrContinuePreviousAction) {
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

	if err := decoderx.NewHTTP().Decode(r, &p,
		decoderx.HTTPFormDecoder(),
		option,
		decoderx.HTTPDecoderSetValidatePayloads(false),
		decoderx.HTTPDecoderSetIgnoreParseErrorsStrategy(decoderx.ParseErrorIgnore),
	); err != nil {
		s.handleSettingsError(w, r, ctxUpdate, nil, &p, err)
		return
	}

	p.RequestID = p.GetRequestID().String()
	s.continueFlow(w, r, ctxUpdate, &p)
}

func (s *StrategyTraits) continueFlow(w http.ResponseWriter, r *http.Request, ctxUpdate *UpdateContext, p *completeSelfServiceBrowserSettingsStrategyProfileFlowPayload) {
	if len(p.Traits) == 0 {
		s.handleSettingsError(w, r, ctxUpdate, nil, p, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Did not receive any value changes.")))
		return
	}

	if err := s.hydrateForm(r, ctxUpdate.Request, ctxUpdate.Session, p.Traits); err != nil {
		s.d.SettingsRequestErrorHandler().HandleSettingsError(w, r, ctxUpdate.Request, err, StrategyProfile)
		return
	}

	update, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), ctxUpdate.Session.Identity.ID)
	if err != nil {
		s.d.SettingsRequestErrorHandler().HandleSettingsError(w, r, ctxUpdate.Request, err, StrategyProfile)
		return
	}

	update.Traits = identity.Traits(p.Traits)
	if err := s.d.SettingsHookExecutor().PostSettingsHook(w, r,
		StrategyProfile, ctxUpdate, update); err != nil {
		s.handleSettingsError(w, r, ctxUpdate, p.Traits, p, err)
		return
	}
}

// Complete profile update payload
//
// swagger:parameters completeSelfServiceBrowserSettingsProfileStrategyFlow
// nolint:deadcode,unused
type completeSelfServiceBrowserSettingsProfileStrategyFlow struct {
	// Request is the request ID.
	//
	// required: true
	// in: query
	// format: uuid
	Request string `json:"request"`

	// in: body
	// required: true
	Body completeSelfServiceBrowserSettingsStrategyProfileFlowPayload
}

// swagger:model completeSelfServiceBrowserSettingsStrategyProfileFlowPayload
// nolint:deadcode,unused
type completeSelfServiceBrowserSettingsStrategyProfileFlowPayload struct {
	// Traits contains all of the identity's traits.
	//
	// type: string
	// format: binary
	// required: true
	Traits json.RawMessage `json:"traits"`

	// RequestID is request ID.
	//
	// in: query
	RequestID string `json:"request_id"`

	rid uuid.UUID
}

func (p *completeSelfServiceBrowserSettingsStrategyProfileFlowPayload) GetRequestID() uuid.UUID {
	return p.rid
}

func (p *completeSelfServiceBrowserSettingsStrategyProfileFlowPayload) SetRequestID(rid uuid.UUID) {
	p.rid = rid
}

func (s *StrategyTraits) hydrateForm(r *http.Request, ar *Request, ss *session.Session, traits json.RawMessage) error {
	action := urlx.CopyWithQuery(
		urlx.AppendPaths(s.c.SelfPublicURL(), PublicSettingsProfilePath),
		url.Values{"request": {ar.ID.String()}},
	)

	ar.Methods[StrategyProfile].Config.Reset()
	if traits != nil {
		for _, field := range form.NewHTMLFormFromJSON(action.String(), traits, "traits").Fields {
			ar.Methods[StrategyProfile].Config.SetField(field)
		}
	}
	ar.Methods[StrategyProfile].Config.SetCSRF(s.d.GenerateCSRFToken(r))

	traitsSchema, err := s.c.IdentityTraitsSchemas().FindSchemaByID(ss.Identity.TraitsSchemaID)
	if err != nil {
		return err
	}

	if err = ar.Methods[StrategyProfile].Config.SortFields(traitsSchema.URL, "traits"); err != nil {
		return err
	}

	return nil
}

// handleSettingsError is a convenience function for handling all types of errors that may occur (e.g. validation error)
// during a settings request.
func (s *StrategyTraits) handleSettingsError(w http.ResponseWriter, r *http.Request, puc *UpdateContext, traits json.RawMessage, p *completeSelfServiceBrowserSettingsStrategyProfileFlowPayload, err error) {
	if errors.Is(err, ErrRequestNeedsReAuthentication) {
		if err := s.d.ContinuityManager().Pause(r.Context(), w, r, strategyProfileContinuityNameFromRequest(r),
			ContinuityOptions(p, puc.Session.Identity)...); err != nil {
			s.d.SettingsRequestErrorHandler().HandleSettingsError(w, r, puc.Request, err, s.SettingsStrategyID())
			return
		}
	}

	if puc.Request != nil {
		if traits == nil {
			traits = json.RawMessage(puc.Session.Identity.Traits)
		}

		if err := s.hydrateForm(r, puc.Request, puc.Session, traits); err != nil {
			s.d.SettingsRequestErrorHandler().HandleSettingsError(w, r, puc.Request, err, s.SettingsStrategyID())
			return
		}
	}

	s.d.SettingsRequestErrorHandler().HandleSettingsError(w, r, puc.Request, err, s.SettingsStrategyID())
}

// newSettingsProfileDecoder returns a decoderx.HTTPDecoderOption with a JSON Schema for type assertion and
// validation.
func (s *StrategyTraits) newSettingsProfileDecoder(i *identity.Identity) (decoderx.HTTPDecoderOption, error) {
	const registrationFormPayloadSchema = `
{
  "$id": "./selfservice/settings/decoder.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["traits"],
  "properties": {
    "traits": {}
  }
}
`

	ss, err := s.d.IdentityTraitsSchemas().GetByID(i.TraitsSchemaID)
	if err != nil {
		return nil, err
	}
	raw, err := sjson.SetBytes(
		[]byte(registrationFormPayloadSchema),
		"properties.traits.$ref",
		ss.URL.String(),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	o, err := decoderx.HTTPRawJSONSchemaCompiler(raw)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return o, nil
}
