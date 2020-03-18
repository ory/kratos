package profile

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/ory/kratos/schema"

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
	PublicProfileManagementPath        = "/self-service/browser/flows/profile"
	PublicProfileManagementRequestPath = "/self-service/browser/flows/requests/profile"
	AdminBrowserProfileRequestPath     = "/self-service/browser/flows/requests/profile"
	PublicProfileManagementUpdatePath  = "/self-service/browser/flows/profile/update"
)

type (
	handlerDependencies interface {
		x.CSRFProvider
		x.WriterProvider
		x.LoggingProvider

		session.HandlerProvider
		session.ManagementProvider

		identity.ValidationProvider
		identity.ManagementProvider
		identity.PrivilegedPoolProvider

		errorx.ManagementProvider

		ErrorHandlerProvider
		RequestPersistenceProvider
		StrategyProvider

		IdentityTraitsSchemas() schema.Schemas
	}
	HandlerProvider interface {
		ProfileManagementHandler() *Handler
	}
	Handler struct {
		c    configuration.Provider
		d    handlerDependencies
		csrf x.CSRFToken
	}
)

func NewHandler(d handlerDependencies, c configuration.Provider) *Handler {
	return &Handler{d: d, c: c, csrf: nosurf.Token}
}

func (h *Handler) RegisterPublicRoutes(public *x.RouterPublic) {
	redirect := session.RedirectOnUnauthenticated(h.c.LoginURL().String())
	public.GET(PublicProfileManagementPath, h.d.SessionHandler().IsAuthenticated(h.initUpdateProfile, redirect))
	public.GET(PublicProfileManagementRequestPath, h.d.SessionHandler().IsAuthenticated(h.publicFetchUpdateProfileRequest, redirect))
	public.POST(PublicProfileManagementUpdatePath, h.d.SessionHandler().IsAuthenticated(h.completeProfileManagementFlow, redirect))
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(AdminBrowserProfileRequestPath, h.adminFetchUpdateProfileRequest)
}

// swagger:route GET /self-service/browser/flows/profile public initializeSelfServiceProfileManagementFlow
//
// Initialize browser-based profile management flow
//
// This endpoint initializes a browser-based profile management flow. Once initialized, the browser will be redirected to
// `urls.profile_ui` with the request ID set as a query parameter. If no valid user session exists, a login
// flow will be initialized.
//
// > This endpoint is NOT INTENDED for API clients and only works
// with browsers (Chrome, Firefox, ...).
//
// More information can be found at [ORY Kratos Profile Management Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-profile-management).
//
//     Schemes: http, https
//
//     Responses:
//       302: emptyResponse
//       500: genericError
func (h *Handler) initUpdateProfile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	s, err := h.d.SessionManager().FetchFromRequest(r.Context(), w, r)
	if err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	a := NewRequest(h.c.SelfServiceProfileRequestLifespan(), r, s)
	for _, strategy := range h.d.ProfileManagementStrategies() {
		if err := strategy.PopulateProfileManagementMethod(r, s, a); err != nil {
			h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
			return
		}
	}

	if err := h.d.ProfileRequestPersister().CreateProfileRequest(r.Context(), a); err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	http.Redirect(w, r,
		urlx.CopyWithQuery(h.c.ProfileURL(), url.Values{"request": {a.ID.String()}}).String(),
		http.StatusFound,
	)
}

// nolint:deadcode,unused
// swagger:parameters getSelfServiceBrowserProfileManagementRequest
type getSelfServiceBrowserLoginRequestParameters struct {
	// Request is the Login Request ID
	//
	// The value for this parameter comes from `request` URL Query parameter sent to your
	// application (e.g. `/login?request=abcde`).
	//
	// required: true
	// in: query
	Request string `json:"request"`
}

// swagger:route GET /self-service/browser/flows/requests/profile common public admin getSelfServiceBrowserProfileManagementRequest
//
// Get the request context of browser-based profile management flows
//
// When accessing this endpoint through ORY Kratos' Public API, ensure that cookies are set as they are required
// for checking the auth session. To prevent scanning attacks, the public endpoint does not return 404 status codes
// but instead 403 or 500.
//
// More information can be found at [ORY Kratos Profile Management Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-profile-management).
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: profileManagementRequest
//       403: genericError
//       404: genericError
//       410: genericError
//       500: genericError
func (h *Handler) publicFetchUpdateProfileRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.fetchUpdateProfileRequest(w, r, true); err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}
}

func (h *Handler) adminFetchUpdateProfileRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.fetchUpdateProfileRequest(w, r, false); err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}
}

func (h *Handler) wrapErrorForbidden(err error, shouldWrap bool) error {
	if shouldWrap {
		return herodot.ErrForbidden.WithReasonf("Access privileges are missing, invalid, or not sufficient to access this endpoint.").WithTrace(err).WithDebugf("%s", err)
	}

	return err
}

func (h *Handler) fetchUpdateProfileRequest(w http.ResponseWriter, r *http.Request, checkSession bool) error {
	rid := x.ParseUUID(r.URL.Query().Get("request"))
	pr, err := h.d.ProfileRequestPersister().GetProfileRequest(r.Context(), rid)
	if err != nil {
		return h.wrapErrorForbidden(err, checkSession)
	}

	if checkSession {
		sess, err := h.d.SessionManager().FetchFromRequest(r.Context(), w, r)
		if err != nil {
			return h.wrapErrorForbidden(err, checkSession)
		}

		if pr.IdentityID != sess.Identity.ID {
			return errors.WithStack(herodot.ErrForbidden.WithReasonf("The request was made for another identity and has been blocked for security reasons."))
		}
	}

	if pr.ExpiresAt.Before(time.Now()) {
		return errors.WithStack(x.ErrGone.
			WithReason("The profile management request has expired. Redirect the user to the login endpoint to initialize a new session.").
			WithDetail("redirect_to", urlx.AppendPaths(h.c.SelfPublicURL(), PublicProfileManagementPath).String()))
	}

	h.d.Writer().Write(w, r, pr)
	return nil
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

// swagger:route POST /self-service/browser/flows/profile/update public completeSelfServiceBrowserProfileManagementFlow
//
// Complete the browser-based profile management flows
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

	var p completeSelfServiceBrowserProfileManagementFlowPayload
	if err := decoderx.NewHTTP().Decode(r, &p,
		decoderx.HTTPFormDecoder(),
		option,
		decoderx.HTTPDecoderSetValidatePayloads(false),
		decoderx.HTTPDecoderSetIgnoreParseErrorsStrategy(decoderx.ParseErrorIgnore),
	); err != nil {
		h.handleProfileManagementError(w, r, nil, s.Identity.Traits, err)
		return
	}

	rid := r.URL.Query().Get("request")
	if len(rid) == 0 {
		h.handleProfileManagementError(w, r, nil, s.Identity.Traits, errors.WithStack(herodot.ErrBadRequest.WithReasonf("The request query parameter is missing.")))
		return
	}

	ar, err := h.d.ProfileRequestPersister().GetProfileRequest(r.Context(), x.ParseUUID(rid))
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

	if s.AuthenticatedAt.After(time.Now()) {
		h.handleProfileManagementError(w, r, ar, s.Identity.Traits, errors.WithStack(
			herodot.ErrInternalServerError.
				WithReason("There was a configuration error, please contact the administrator.").
				WithDebugf("session.AuthenticatedAt was %fs in the future. This should not happen.", time.Since(s.AuthenticatedAt).Seconds())))
		return
	}

	identityManagerOptions := []identity.ManagerOption{identity.ManagerExposeValidationErrors}
	if time.Since(s.AuthenticatedAt) < h.c.SelfServicePrivilegedSessionMaxAge() {
		identityManagerOptions = append(identityManagerOptions, identity.ManagerAllowWriteProtectedTraits)
	}

	if err := h.d.IdentityManager().UpdateTraits(r.Context(), s.Identity.ID, identity.Traits(p.Traits), identityManagerOptions...); err != nil {
		h.handleProfileManagementError(w, r, ar, identity.Traits(p.Traits), err)
		return
	}

	action := urlx.CopyWithQuery(
		urlx.AppendPaths(h.c.SelfPublicURL(), PublicProfileManagementUpdatePath),
		url.Values{"request": {ar.ID.String()}},
	)
	ar.Methods[FormTraitsID].Config.Reset()
	ar.UpdateSuccessful = true
	for _, field := range form.NewHTMLFormFromJSON(action.String(), p.Traits, "traits").Fields {
		ar.Methods[FormTraitsID].Config.SetField(field)
	}
	ar.Methods[FormTraitsID].Config.SetCSRF(nosurf.Token(r))

	traitsSchema, err := h.c.IdentityTraitsSchemas().FindSchemaByID(s.Identity.TraitsSchemaID)
	if err != nil {
		h.handleProfileManagementError(w, r, ar, identity.Traits(p.Traits), err)
		return
	}

	if err = ar.Methods[FormTraitsID].Config.SortFields(traitsSchema.URL, "traits"); err != nil {
		h.handleProfileManagementError(w, r, ar, identity.Traits(p.Traits), err)
		return
	}

	if err := h.d.ProfileRequestPersister().UpdateProfileRequest(r.Context(), ar); err != nil {
		h.handleProfileManagementError(w, r, ar, identity.Traits(p.Traits), err)
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
		action := urlx.CopyWithQuery(
			urlx.AppendPaths(h.c.SelfPublicURL(), PublicProfileManagementUpdatePath),
			url.Values{"request": {rr.ID.String()}},
		)

		rr.Methods[FormTraitsID].Config.Reset()
		rr.UpdateSuccessful = false

		if traits != nil {
			for _, field := range form.NewHTMLFormFromJSON(action.String(), json.RawMessage(traits), "traits").Fields {
				rr.Methods[FormTraitsID].Config.SetField(field)
			}
		}
		rr.Methods[FormTraitsID].Config.SetCSRF(nosurf.Token(r))

		// try to sort, might fail if the error before was sorting related
		traitsSchema, err := h.c.IdentityTraitsSchemas().FindSchemaByID(rr.Identity.TraitsSchemaID)
		if err != nil {
			h.d.ProfileRequestRequestErrorHandler().HandleProfileManagementError(w, r, rr, err)
			return
		}

		if err := rr.Methods[FormTraitsID].Config.SortFields(traitsSchema.URL, "traits"); err != nil {
			h.d.ProfileRequestRequestErrorHandler().HandleProfileManagementError(w, r, rr, err)
			return
		}

		if err := h.d.ProfileRequestPersister().UpdateProfileRequest(r.Context(), rr); err != nil {
			h.d.ProfileRequestRequestErrorHandler().HandleProfileManagementError(w, r, rr, err)
			return
		}
	}

	h.d.ProfileRequestRequestErrorHandler().HandleProfileManagementError(w, r, rr, err)
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
