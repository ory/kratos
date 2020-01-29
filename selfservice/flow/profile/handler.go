package profile

import (
	"encoding/json"
	"net/http"
	"net/url"
	"sort"

	"github.com/ory/kratos/schema"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/nosurf"
	"github.com/pkg/errors"
	"github.com/tidwall/sjson"

	"github.com/ory/herodot"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	BrowserProfilePath        = "/self-service/browser/flows/profile"
	BrowserProfileRequestPath = "/self-service/browser/flows/requests/profile"
)

type (
	handlerDependencies interface {
		x.CSRFProvider
		x.WriterProvider
		x.LoggingProvider

		session.HandlerProvider
		session.ManagementProvider

		identity.PoolProvider
		identity.ValidationProvider

		errorx.ManagementProvider

		ErrorHandlerProvider
		RequestPersistenceProvider

		IdentityTraitsSchemas() schema.Schemas
	}
	HandlerProvider interface {
		ProfileManagementHandler() *Handler
	}
	Handler struct {
		c configuration.Provider
		d handlerDependencies
	}
)

func NewHandler(d handlerDependencies, c configuration.Provider) *Handler {
	return &Handler{d: d, c: c}
}

func (h *Handler) RegisterPublicRoutes(public *x.RouterPublic) {
	redirect := session.RedirectOnUnauthenticated(h.c.LoginURL().String())
	public.GET(BrowserProfilePath, h.d.SessionHandler().IsAuthenticated(h.initUpdateProfile, redirect))
	public.GET(BrowserProfileRequestPath, h.d.SessionHandler().IsAuthenticated(h.fetchUpdateProfileRequest, redirect))
	public.POST(BrowserProfilePath, h.d.SessionHandler().IsAuthenticated(h.completeProfileManagementFlow, redirect))
}

// swagger:route GET /profiles public initializeProfileManagementFlow
//
// Initialize Profile Management Flow
//
// This endpoint initializes a profile update flow. This endpoint **should not be called from a programatic API**
// but instead for the, for example, browser. It will redirect the user agent (e.g. browser) to the
// configured login UI, appending the login challenge.
//
// If the user-agent does not have a valid authentication session, a 302 code will be returned which
// redirects to the initializeLoginFlow endpoint, appending this page as the return_to value.
//
// For an in-depth look at ORY Krato's profile management flow, head over to: https://www.ory.sh/docs/kratos/selfservice/profile
//
//     Schemes: http, https
//
//     Responses:
//       302: emptyResponse
//       500: genericError
func (h *Handler) initUpdateProfile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	s, err := h.d.SessionManager().FetchFromRequest(r.Context(), w, r)
	if err != nil {
		h.d.SelfServiceErrorManager().ForwardError(r.Context(), w, r, err)
		return
	}

	a := NewRequest(h.c.SelfServiceProfileRequestLifespan(), r, s)
	a.Form = form.NewHTMLFormFromJSON(urlx.AppendPaths(h.c.SelfPublicURL(), BrowserProfilePath).String(), json.RawMessage(s.Identity.Traits), "traits")
	if err := h.d.ProfileRequestPersister().CreateProfileRequest(r.Context(), a); err != nil {
		h.d.SelfServiceErrorManager().ForwardError(r.Context(), w, r, err)
		return
	}

	http.Redirect(w, r,
		urlx.CopyWithQuery(h.c.ProfileURL(), url.Values{"request": {a.ID.String()}}).String(),
		http.StatusFound,
	)
}

// swagger:parameters getProfileManagementRequest
type (
	// nolint:deadcode,unused
	getProfileManagementRequestParameters struct {
		// Request should be set to the value of the `request` query parameter
		// by the profile management UI.
		//
		// in: query
		// required: true
		Request string `json:"request"`
	}
)

// swagger:route GET /profiles/requests admin getProfileManagementRequest
//
// Get Profile Management Request
//
// This endpoint returns a profile management request's context with, for example, error details and
// other information.
//
// It can be used from a server or other applications running in a privileged network with access to
// ORY Kratos' admin port.
//
// If you wish to access this endpoint from e.g. a SPA instead, please call this path at the public port
// and make sure to include cookies in that request.
//
// For an in-depth look at ORY Krato's profile management flow, head over to: https://www.ory.sh/docs/kratos/selfservice/profile
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: profileManagementRequest
//       302: emptyResponse
//       500: genericError
//
// nolint:deadcode,unused
func fetchUpdateProfileRequestAdmin() {}

// swagger:route GET /profiles/requests public getProfileManagementRequest
//
// Get Profile Management Request (via cookie)
//
// This endpoint returns a profile management request's context with, for example, error details and
// other information.
//
// It can be used from a Single Page Application or other applications running on a client device.
// The request must be made with valid authentication cookies or it will fail!
//
// If you wish to access this endpoint without the valid cookies (e.g. as part of a server)
// please call this path at the admin port.
//
// For an in-depth look at ORY Krato's profile management flow, head over to: https://www.ory.sh/docs/kratos/selfservice/profile
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: profileManagementRequest
//       302: emptyResponse
//       500: genericError
func (h *Handler) fetchUpdateProfileRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	rid := x.ParseUUID(r.URL.Query().Get("request"))
	ar, err := h.d.ProfileRequestPersister().GetProfileRequest(r.Context(), rid)
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	sess, err := h.d.SessionManager().FetchFromRequest(r.Context(), w, r)
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	if ar.IdentityID != sess.Identity.ID {
		h.d.Writer().WriteError(w, r, errors.WithStack(herodot.ErrForbidden.WithReasonf("The request was made for another identity and has been blocked for security reasons.")))
		return
	}

	ar.Form.SetField("request", form.Field{
		Name:     "request",
		Type:     "hidden",
		Required: true,
		Value:    rid,
	})
	ar.Form.SetCSRF(nosurf.Token(r))
	sort.Sort(ar.Form.Fields)
	h.d.Writer().Write(w, r, ar)
}

type (
	// swagger:parameters completeProfileManagementFlow
	// nolint:deadcode,unused
	completeProfileManagementParameters struct {
		// in: body
		// required: true
		Body completeProfileManagementPayload
	}

	// swagger:model completeProfileManagementPayload
	// nolint:deadcode,unused
	completeProfileManagementPayload struct {
		// Traits contains all of the identity's traits.
		//
		// type: string
		// format: binary
		// required: true
		Traits json.RawMessage `json:"traits"`

		// Request is the request ID.
		//
		// type: string
		// required: true
		Request uuid.UUID `json:"request"`
	}
)

// swagger:route POST /profiles public completeProfileManagementFlow
//
// Complete Profile Management Flow
//
// This endpoint returns a login request's context with, for example, error details and
// other information.
//
// For an in-depth look at ORY Krato's profile management flow, head over to: https://www.ory.sh/docs/kratos/selfservice/profile
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
func (h *Handler) completeProfileManagementFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	s, err := h.d.SessionManager().FetchFromRequest(r.Context(), w, r)
	if err != nil {
		h.handleProfileManagementError(w, r, nil, nil, err)
		return
	}

	option, err := h.newProfileManagementDecoder(s.Identity)
	if err != nil {
		h.handleProfileManagementError(w, r, nil, s.Identity.Traits, err)
		return
	}

	var p completeProfileManagementPayload
	if err := decoderx.NewHTTP().Decode(r, &p,
		decoderx.HTTPFormDecoder(),
		option,
		decoderx.HTTPDecoderSetValidatePayloads(false),
		decoderx.HTTPDecoderSetIgnoreParseErrorsStrategy(decoderx.ParseErrorIgnore),
	); err != nil {
		h.handleProfileManagementError(w, r, nil, s.Identity.Traits, err)
		return
	}

	if x.IsZeroUUID(p.Request) {
		h.handleProfileManagementError(w, r, nil, s.Identity.Traits, errors.WithStack(herodot.ErrBadRequest.WithReasonf("The request query parameter is missing.")))
		return
	}

	ar, err := h.d.ProfileRequestPersister().GetProfileRequest(r.Context(), p.Request)
	if err != nil {
		h.handleProfileManagementError(w, r, nil, s.Identity.Traits, err)
		return
	}

	if err := ar.Valid(s); err != nil {
		h.handleProfileManagementError(w, r, ar, s.Identity.Traits, err)
		return
	}

	if len(p.Traits) == 0 {
		h.handleProfileManagementError(w, r, ar, s.Identity.Traits, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Did not receive any value changes.")))
		return
	}

	// identity.TraitsSchemaURL

	creds, err := h.d.IdentityPool().GetIdentityConfidential(r.Context(), s.Identity.ID)
	if err != nil {
		h.handleProfileManagementError(w, r, ar, identity.Traits(p.Traits), err)
		return
	}

	i := *s.Identity
	i.Traits = identity.Traits(p.Traits)
	i.Credentials = creds.CopyCredentials()

	// If credential identifiers have changed we need to block this action UNLESS
	// the identity has been authenticated in that request:
	//
	// - https://security.stackexchange.com/questions/24291/why-do-we-ask-for-a-users-existing-password-when-changing-their-password

	// We need to make sure that the identity has a valid schema before passing it down to the identity pool.
	if err := h.d.IdentityValidator().Validate(&i); err != nil {
		h.handleProfileManagementError(w, r, ar, i.Traits, err)
		return
	}

	// Check if any credentials-related field changed.
	if !i.CredentialsEqual(creds.Credentials) {

		// !! WARNING !!
		//
		// This will leak the credential options which may include the hashed password. Do not use seriously:
		//
		//	h.d.Logger().
		//	 	WithField("original_credentials", fmt.Sprintf("%+v", creds.Credentials)).
		//	 	WithField("updated_credentials", fmt.Sprintf("%+v", i.Credentials)).
		//	 	Trace("Credentials changed unexpectedly in CompleteProfileManagementFlow.")

		h.handleProfileManagementError(w, r, ar, i.Traits,
			errors.WithStack(
				herodot.ErrInternalServerError.
					WithReasonf(`A field was modified that updates one or more credentials-related settings. These fields can only be updated as part of a "Change your password", or "Link authentication methods" flow which requires prior authentication. This is a configuration error.`)),
		)
		return
	}

	if err := h.d.IdentityPool().UpdateIdentity(r.Context(), &i); err != nil {
		h.handleProfileManagementError(w, r, ar, i.Traits, err)
		return
	}

	ar.Form.Reset()
	ar.UpdateSuccessful = true
	for _, field := range form.NewHTMLFormFromJSON("", json.RawMessage(i.Traits), "traits").Fields {
		ar.Form.SetField(field.Name, field)
	}
	ar.Form.SetValue("request", r.Form.Get("request"))
	ar.Form.SetCSRF(nosurf.Token(r))
	sort.Sort(ar.Form.Fields)

	if err := h.d.ProfileRequestPersister().UpdateProfileRequest(r.Context(), ar); err != nil {
		h.handleProfileManagementError(w, r, ar, i.Traits, err)
		return
	}

	http.Redirect(w, r,
		urlx.CopyWithQuery(h.c.ProfileURL(), url.Values{"request": {ar.ID.String()}}).String(),
		http.StatusFound,
	)
}

// handleProfileManagementError is a convenience function for handling all types of errors that may occur (e.g. validation error)
// during a profile management request.
func (h *Handler) handleProfileManagementError(w http.ResponseWriter, r *http.Request, rr *Request, traits identity.Traits, err error) {
	if rr != nil {
		rr.Form.Reset()
		rr.UpdateSuccessful = false

		if traits != nil {
			for _, field := range form.NewHTMLFormFromJSON("", json.RawMessage(traits), "traits").Fields {
				rr.Form.SetField(field.Name, field)
			}
		}
		rr.Form.SetValue("request", r.Form.Get("request"))
		rr.Form.SetCSRF(nosurf.Token(r))
		sort.Sort(rr.Form.Fields)
	}

	h.d.ProfileRequestRequestErrorHandler().HandleProfileManagementError(w, r, identity.CredentialsTypePassword, rr, err)
}

// newProfileManagementDecoder returns a decoderx.HTTPDecoderOption with a JSON Schema for type assertion and
// validation.
func (h *Handler) newProfileManagementDecoder(i *identity.Identity) (decoderx.HTTPDecoderOption, error) {
	const registrationFormPayloadSchema = `
{
  "$id": "./selfservice/profile/decoder.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["traits"],
  "properties": {
    "request": { "type": "string" },
    "traits": {}
  }
}
`

	s, err := h.d.IdentityTraitsSchemas().GetByID(i.TraitsSchemaID)
	if err != nil {
		return nil, err
	}
	raw, err := sjson.SetBytes(
		[]byte(registrationFormPayloadSchema),
		"properties.traits.$ref",
		s.URL.String(),
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
