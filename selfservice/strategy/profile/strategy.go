package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"runtime/debug"

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
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	PublicSettingsProfilePath = "/self-service/browser/flows/settings/strategies/profile"
)

var _ settings.Strategy = new(Strategy)
var pkgName = reflect.TypeOf(Strategy{}).PkgPath()

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
		settings.RequestPersistenceProvider
		settings.StrategyProvider
		settings.HooksProvider

		IdentityTraitsSchemas() schema.Schemas
	}
	Strategy struct {
		c configuration.Provider
		d strategyDependencies
	}
)

// swagger:model settingsProfileFormConfig
type SettingsProfileRequestMethod struct {
	*form.HTMLForm
}

func NewStrategy(d strategyDependencies, c configuration.Provider) *Strategy {
	return &Strategy{c: c, d: d}
}

func (s *Strategy) SettingsStrategyID() string {
	return settings.StrategyProfile
}

func (s *Strategy) RegisterSettingsRoutes(public *x.RouterPublic) {
	redirect := session.RedirectOnUnauthenticated(s.c.SelfServiceFlowLoginUI().String())
	public.POST(PublicSettingsProfilePath, s.d.SessionHandler().IsAuthenticated(s.handleSubmit, redirect))
	public.GET(PublicSettingsProfilePath, s.d.SessionHandler().IsAuthenticated(s.handleSubmit, redirect))
}

func (s *Strategy) PopulateSettingsMethod(r *http.Request, ss *session.Session, pr *settings.Request) error {
	traitsSchema, err := s.c.IdentityTraitsSchemas().FindSchemaByID(ss.Identity.SchemaID)
	if err != nil {
		return err
	}

	// use a schema compiler that disables identifiers
	schemaCompiler := jsonschema.NewCompiler()

	f, err := form.NewHTMLFormFromJSONSchema(urlx.CopyWithQuery(
		urlx.AppendPaths(s.c.SelfPublicURL(), PublicSettingsProfilePath),
		url.Values{"request": {pr.ID.String()}},
	).String(), traitsSchema.URL, "", schemaCompiler)
	if err != nil {
		return err
	}

	f.SetValuesFromJSON(json.RawMessage(ss.Identity.Traits), "traits")
	f.SetCSRF(s.d.GenerateCSRFToken(r))

	if err := f.SortFields(traitsSchema.URL); err != nil {
		return err
	}

	pr.Methods[s.SettingsStrategyID()] = &settings.RequestMethod{
		Method: s.SettingsStrategyID(),
		Config: &settings.RequestMethodConfig{RequestMethodConfigurator: &SettingsProfileRequestMethod{HTMLForm: f}},
	}
	return nil
}

// swagger:route POST /self-service/browser/flows/settings/strategies/profile public completeSelfServiceBrowserSettingsProfileStrategyFlow
//
// Complete the Browser-Based Settings Flow for Profile Data
//
// This endpoint completes a browser-based settings flow. This is usually achieved by POSTing data to this
// endpoint.
//
// If the provided profile data is valid against the Identity's Traits JSON Schema, the data will be updated and
// the browser redirected to `url.settings_ui` for further steps.
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
func (s *Strategy) handleSubmit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var p completeSelfServiceBrowserSettingsStrategyProfileFlowPayload
	ctxUpdate, err := settings.PrepareUpdate(s.d, w, r, settings.ContinuityKey(s.SettingsStrategyID()), &p)
	s.d.Logger().
		WithField("err", fmt.Sprintf("%+v", err)).
		WithField("ctxUpdate", fmt.Sprintf("%+v", ctxUpdate)).
		WithField("package", pkgName).
		WithField("stack_trace", fmt.Sprintf("%s", debug.Stack())).
		Debug("handleSubmit: PrepareUpdate")
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

	if err := decoderx.NewHTTP().Decode(r, &p,
		decoderx.HTTPFormDecoder(),
		option,
		decoderx.HTTPDecoderSetValidatePayloads(false),
		decoderx.HTTPDecoderSetIgnoreParseErrorsStrategy(decoderx.ParseErrorIgnore),
	); err != nil {
		s.handleSettingsError(w, r, ctxUpdate, nil, &p, err)
		return
	}

	// Reset after decoding form
	p.SetRequestID(ctxUpdate.Request.ID)

	s.continueFlow(w, r, ctxUpdate, &p)
}

func (s *Strategy) continueFlow(w http.ResponseWriter, r *http.Request, ctxUpdate *settings.UpdateContext, p *completeSelfServiceBrowserSettingsStrategyProfileFlowPayload) {
	if len(p.Traits) == 0 {
		s.handleSettingsError(w, r, ctxUpdate, nil, p, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Did not receive any value changes.")))
		return
	}

	if err := s.hydrateForm(r, ctxUpdate.Request, ctxUpdate.Session, p.Traits); err != nil {
		s.d.SettingsRequestErrorHandler().HandleSettingsError(w, r, ctxUpdate.Request, err, settings.StrategyProfile)
		return
	}

	update, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), ctxUpdate.Session.Identity.ID)
	if err != nil {
		s.d.SettingsRequestErrorHandler().HandleSettingsError(w, r, ctxUpdate.Request, err, settings.StrategyProfile)
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

func (s *Strategy) hydrateForm(r *http.Request, ar *settings.Request, ss *session.Session, traits json.RawMessage) error {
	action := urlx.CopyWithQuery(
		urlx.AppendPaths(s.c.SelfPublicURL(), PublicSettingsProfilePath),
		url.Values{"request": {ar.ID.String()}},
	)

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
func (s *Strategy) handleSettingsError(w http.ResponseWriter, r *http.Request, puc *settings.UpdateContext, traits json.RawMessage, p *completeSelfServiceBrowserSettingsStrategyProfileFlowPayload, err error) {
	if errors.Is(err, settings.ErrRequestNeedsReAuthentication) {
		s.d.Logger().
			WithField("err", fmt.Sprintf("%+v", err)).
			WithField("key", settings.ContinuityKey(s.SettingsStrategyID())).
			WithField("package", pkgName).
			WithField("stack_trace", fmt.Sprintf("%s", debug.Stack())).
			Debug("handleSettingsError: Pause")

		if err := s.d.ContinuityManager().Pause(r.Context(), w, r,
			settings.ContinuityKey(s.SettingsStrategyID()),
			settings.ContinuityOptions(p, puc.Session.Identity)...); err != nil {
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
func (s *Strategy) newSettingsProfileDecoder(i *identity.Identity) (decoderx.HTTPDecoderOption, error) {
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

	ss, err := s.d.IdentityTraitsSchemas().GetByID(i.SchemaID)
	if err != nil {
		return nil, err
	}
	raw, err := sjson.SetBytes(
		[]byte(registrationFormPayloadSchema),
		"properties.traits.$ref",
		ss.URL.String()+"#/properties/traits",
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
