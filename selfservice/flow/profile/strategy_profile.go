package profile

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/tidwall/sjson"

	"github.com/ory/herodot"
	"github.com/ory/jsonschema/v3"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/errorsx"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	StrategyTraitsID                  = "profile"
	PublicProfileManagementUpdatePath = "/self-service/browser/flows/profile/strategies/profile"
)

var _ Strategy = new(StrategyTraits)

type (
	strategyDependencies interface {
		x.CSRFProvider
		x.CSRFTokenGeneratorProvider
		x.WriterProvider
		x.LoggingProvider

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

func (s *StrategyTraits) ProfileManagementStrategyID() string {
	return StrategyTraitsID
}

func (s *StrategyTraits) RegisterProfileManagementRoutes(public *x.RouterPublic) {
	redirect := session.RedirectOnUnauthenticated(s.c.LoginURL().String())
	public.POST(PublicProfileManagementUpdatePath, s.d.SessionHandler().IsAuthenticated(s.completeProfileManagementFlow, redirect))
}

func (s *StrategyTraits) PopulateProfileManagementMethod(r *http.Request, ss *session.Session, pr *Request) error {
	traitsSchema, err := s.c.IdentityTraitsSchemas().FindSchemaByID(ss.Identity.TraitsSchemaID)
	if err != nil {
		return err
	}

	// use a schema compiler that disables identifiers
	schemaCompiler := jsonschema.NewCompiler()
	registerNewDisableIdentifiersExtension(schemaCompiler)

	f, err := form.NewHTMLFormFromJSONSchema(urlx.CopyWithQuery(
		urlx.AppendPaths(s.c.SelfPublicURL(), PublicProfileManagementUpdatePath),
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

	pr.Methods[s.ProfileManagementStrategyID()] = &RequestMethod{
		Method: s.ProfileManagementStrategyID(),
		Config: &RequestMethodConfig{RequestMethodConfigurator: &TraitsRequestMethod{HTMLForm: f}},
	}
	return nil
}

// swagger:route POST /self-service/browser/flows/profile/strategies/profile public completeSelfServiceBrowserProfileManagementProfileStrategyFlow
//
// Complete the browser-based profile management flow for profile data
//
// This endpoint completes a browser-based profile management flow. This is usually achieved by POSTing data to this
// endpoint.
//
// If the provided profile data is valid against the Identity's Traits JSON Schema, the data will be updated and
// the browser redirected to `url.profile_ui` for further steps.
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
func (s *StrategyTraits) completeProfileManagementFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ss, err := s.d.SessionManager().FetchFromRequest(r.Context(), w, r)
	if err != nil {
		s.handleProfileManagementError(w, r, nil, nil, nil, err)
		return
	}

	option, err := s.newProfileManagementDecoder(ss.Identity)
	if err != nil {
		s.handleProfileManagementError(w, r, nil, ss, json.RawMessage(ss.Identity.Traits), err)
		return
	}

	var p completeSelfServiceBrowserProfileManagementFlowPayload
	if err := decoderx.NewHTTP().Decode(r, &p,
		decoderx.HTTPFormDecoder(),
		option,
		decoderx.HTTPDecoderSetValidatePayloads(false),
		decoderx.HTTPDecoderSetIgnoreParseErrorsStrategy(decoderx.ParseErrorIgnore),
	); err != nil {
		s.handleProfileManagementError(w, r, nil, ss, json.RawMessage(ss.Identity.Traits), err)
		return
	}

	rid := r.URL.Query().Get("request")
	if len(rid) == 0 {
		s.handleProfileManagementError(w, r, nil, ss, json.RawMessage(ss.Identity.Traits), errors.WithStack(herodot.ErrBadRequest.WithReasonf("The request query parameter is missing.")))
		return
	}

	ar, err := s.d.ProfileRequestPersister().GetProfileRequest(r.Context(), x.ParseUUID(rid))
	if err != nil {
		s.handleProfileManagementError(w, r, nil, ss, json.RawMessage(ss.Identity.Traits), err)
		return
	}

	if err := ar.Valid(ss); err != nil {
		s.handleProfileManagementError(w, r, ar, ss, json.RawMessage(ss.Identity.Traits), err)
		return
	}

	if len(p.Traits) == 0 {
		s.handleProfileManagementError(w, r, ar, ss, json.RawMessage(ss.Identity.Traits), errors.WithStack(herodot.ErrBadRequest.WithReasonf("Did not receive any value changes.")))
		return
	}

	if err := s.hydrateForm(r, ar, ss, p.Traits); err != nil {
		s.d.ProfileRequestRequestErrorHandler().HandleProfileManagementError(w, r, ar, err, StrategyTraitsID)
		return
	}

	update, err := s.d.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), ss.Identity.ID)
	if err != nil {
		s.d.ProfileRequestRequestErrorHandler().HandleProfileManagementError(w, r, ar, err, StrategyTraitsID)
		return
	}
	update.Traits = identity.Traits(p.Traits)

	if err := s.d.ProfileManagementExecutor().PostProfileManagementHook(w, r,
		s.d.PostProfileManagementHooks(StrategyTraitsID),
		ar, ss, update,
	); errorsx.Cause(err) == ErrHookAbortRequest {
		return
	} else if err != nil {
		s.handleProfileManagementError(w, r, ar, ss, p.Traits, err)
		return
	}

	if len(w.Header().Get("Location")) == 0 {
		http.Redirect(w, r,
			urlx.CopyWithQuery(s.c.ProfileURL(), url.Values{"request": {ar.ID.String()}}).String(),
			http.StatusFound,
		)
	}
}

// Complete profile update payload
//
// swagger:parameters completeSelfServiceBrowserProfileManagementFlow
// nolint:deadcode,unused
type completeProfileManagementParameters struct {
	// Request is the request ID.
	//
	// required: true
	// in: query
	// format: uuid
	Request string `json:"request"`

	// in: body
	// required: true
	Body completeSelfServiceBrowserProfileManagementFlowPayload
}

// swagger:model completeSelfServiceBrowserProfileManagementFlowPayload
// nolint:deadcode,unused
type completeSelfServiceBrowserProfileManagementFlowPayload struct {
	// Traits contains all of the identity's traits.
	//
	// type: string
	// format: binary
	// required: true
	Traits json.RawMessage `json:"traits"`
}

func (s *StrategyTraits) hydrateForm(r *http.Request, ar *Request, ss *session.Session, traits json.RawMessage) error {
	action := urlx.CopyWithQuery(
		urlx.AppendPaths(s.c.SelfPublicURL(), PublicProfileManagementUpdatePath),
		url.Values{"request": {ar.ID.String()}},
	)

	ar.Methods[StrategyTraitsID].Config.Reset()
	if traits != nil {
		for _, field := range form.NewHTMLFormFromJSON(action.String(), traits, "traits").Fields {
			ar.Methods[StrategyTraitsID].Config.SetField(field)
		}
	}
	ar.Methods[StrategyTraitsID].Config.SetCSRF(s.d.GenerateCSRFToken(r))

	traitsSchema, err := s.c.IdentityTraitsSchemas().FindSchemaByID(ss.Identity.TraitsSchemaID)
	if err != nil {
		return err
	}

	if err = ar.Methods[StrategyTraitsID].Config.SortFields(traitsSchema.URL, "traits"); err != nil {
		return err
	}

	return nil
}

// handleProfileManagementError is a convenience function for handling all types of errors that may occur (e.g. validation error)
// during a profile management request.
func (s *StrategyTraits) handleProfileManagementError(w http.ResponseWriter, r *http.Request, rr *Request, ss *session.Session, traits json.RawMessage, err error) {
	if rr != nil {
		if err := s.hydrateForm(r, rr, ss, traits); err != nil {
			s.d.ProfileRequestRequestErrorHandler().HandleProfileManagementError(w, r, rr, err, StrategyTraitsID)
			return
		}

		if err := s.d.ProfileRequestPersister().UpdateProfileRequest(r.Context(), rr); err != nil {
			s.d.ProfileRequestRequestErrorHandler().HandleProfileManagementError(w, r, rr, err, StrategyTraitsID)
			return
		}
	}

	s.d.ProfileRequestRequestErrorHandler().HandleProfileManagementError(w, r, rr, err, StrategyTraitsID)
}

// newProfileManagementDecoder returns a decoderx.HTTPDecoderOption with a JSON Schema for type assertion and
// validation.
func (s *StrategyTraits) newProfileManagementDecoder(i *identity.Identity) (decoderx.HTTPDecoderOption, error) {
	const registrationFormPayloadSchema = `
{
  "$id": "./selfservice/profile/decoder.schema.json",
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
