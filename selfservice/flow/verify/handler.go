package verify

import (
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/nosurf"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/jsonschema/v3"
	"github.com/ory/x/errorsx"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/x"
)

const (
	PublicVerificationPath        = "/self-service/browser/flows/verification/:via"
	PublicVerificationRequestPath = "/self-service/browser/flows/requests/verification"
	PublicVerificationConfirmPath = "/self-service/browser/flows/verification/:code"
)

type (
	handlerDependencies interface {
		x.CSRFTokenGeneratorProvider
		errorx.ManagementProvider
		x.WriterProvider

		PersistenceProvider
		ManagementProvider
		ErrorHandlerProvider
	}
	Handler struct {
		d handlerDependencies
		c configuration.Provider
	}
)

func NewHandler(d handlerDependencies, c configuration.Provider) *Handler {
	return &Handler{c: c, d: d}
}

func (h *Handler) RegisterPublicRoutes(public *x.RouterPublic) {
	public.GET(PublicVerificationPath, h.init)
	public.GET(PublicVerificationRequestPath, h.publicFetch)
	public.POST(PublicVerificationRequestPath, h.complete)
	public.GET(PublicVerificationConfirmPath, h.verify)
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(PublicVerificationRequestPath, h.adminFetch)
}

// nolint:deadcode,unused
// swagger:parameters initializeSelfServiceVerificationFlowParameters
type initializeSelfServiceVerificationFlowParameters struct {
	// What to verify
	//
	// Currently only "email" is supported.
	//
	// required: true
	// in: path
	Via string `json:"via"`
}

// swagger:route GET /self-service/browser/flows/verification/{via} public initializeSelfServiceBrowserVerificationFlow
//
// Initialize browser-based verification flow
//
// This endpoint initializes a browser-based profile management flow. Once initialized, the browser will be redirected to
// `urls.profile_ui` with the request ID set as a query parameter. If no valid user session exists, a login
// flow will be initialized.
//
// > This endpoint is NOT INTENDED for API clients and only works
// with browsers (Chrome, Firefox, ...).
//
// More information can be found at [ORY Kratos Email and Phone Verification Documentation](https://www.ory.sh/docs/kratos/selfservice/flows/verify-email-account-activation).
//
//     Schemes: http, https
//
//     Responses:
//       302: emptyResponse
//       500: genericError
func (h *Handler) init(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	via, err := h.toVia(ps)
	if err != nil {
		h.handleError(w, r, nil, err)
		return
	}

	a := NewRequest(
		h.c.SelfServiceProfileRequestLifespan(), r, via,
		urlx.AppendPaths(h.c.SelfPublicURL(), PublicVerificationRequestPath), h.d.GenerateCSRFToken,
	)

	if err := h.d.VerificationPersister().CreateVerifyRequest(r.Context(), a); err != nil {
		h.handleError(w, r, nil, err)
		return
	}

	http.Redirect(w, r,
		urlx.CopyWithQuery(h.c.VerificationURL(), url.Values{"request": {a.ID.String()}}).String(),
		http.StatusFound,
	)
}

// nolint:deadcode,unused
// swagger:parameters getSelfServiceVerificationRequest
type getSelfServiceVerificationRequestParameters struct {
	// Request is the Request ID
	//
	// The value for this parameter comes from `request` URL Query parameter sent to your
	// application (e.g. `/verify?request=abcde`).
	//
	// required: true
	// in: query
	Request string `json:"request"`
}

// swagger:route GET /self-service/browser/flows/requests/verification common public admin getSelfServiceVerificationRequest
//
// Get the request context of browser-based verification flows
//
// When accessing this endpoint through ORY Kratos' Public API, ensure that cookies are set as they are required
// for checking the auth session. To prevent scanning attacks, the public endpoint does not return 404 status codes
// but instead 403 or 500.
//
// More information can be found at [ORY Kratos Email and Phone Verification Documentation](https://www.ory.sh/docs/kratos/selfservice/flows/verify-email-account-activation).
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: verificationRequest
//       403: genericError
//       404: genericError
//       500: genericError
func (h *Handler) publicFetch(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if err := h.fetch(w, r, true); err != nil {
		h.d.Writer().WriteError(w, r, herodot.ErrForbidden.WithReasonf("Access privileges are missing, invalid, or not sufficient to access this endpoint.").WithTrace(err).WithDebugf("%s", err))
		return
	}
}

func (h *Handler) adminFetch(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if err := h.fetch(w, r, false); err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}
}

func (h *Handler) fetch(w http.ResponseWriter, r *http.Request, mustVerify bool) error {
	rid := x.ParseUUID(r.URL.Query().Get("request"))
	ar, err := h.d.VerificationPersister().GetVerifyRequest(r.Context(), rid)
	if err != nil {
		return err
	}

	if mustVerify && !nosurf.VerifyToken(h.d.GenerateCSRFToken(r), ar.CSRFToken) {
		return errors.WithStack(x.ErrInvalidCSRFToken)
	}

	h.d.Writer().Write(w, r, ar)
	return nil
}

// nolint:deadcode,unused
// swagger:parameters completeSelfServiceBrowserVerificationFlow
type completeSelfServiceBrowserVerificationFlowParameters struct {
	// Request is the Request ID
	//
	// The value for this parameter comes from `request` URL Query parameter sent to your
	// application (e.g. `/verify?request=abcde`).
	//
	// required: true
	// in: query
	Request string `json:"request"`
}

// swagger:route POST /self-service/browser/flows/requests/verification public completeSelfServiceBrowserVerificationFlow
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
// More information can be found at [ORY Kratos Email and Phone Verification Documentation](https://www.ory.sh/docs/kratos/selfservice/flows/verify-email-account-activation).
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
func (h *Handler) complete(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if err := r.ParseForm(); err != nil {
		h.handleError(w, r, nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Unable to parse the request: %s", err)))
		return
	}

	rid := r.URL.Query().Get("request")
	if len(rid) == 0 {
		h.handleError(w, r, nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("The request query parameter is missing.")))
		return
	}

	vr, err := h.d.VerificationPersister().GetVerifyRequest(r.Context(), x.ParseUUID(rid))
	if err != nil {
		h.handleError(w, r, nil, err)
		return
	}

	switch vr.Via {
	case ViaEmail:
		h.completeViaEmail(w, r, vr)

	default:
		h.handleError(w, r, vr, errors.WithStack(herodot.ErrInternalServerError.WithDebugf("Ended up with an invalid VerifyRequest.Via: %s", vr.Via)))
		return
	}
}

func (h *Handler) completeViaEmail(w http.ResponseWriter, r *http.Request, vr *Request) {
	panic("not implemented")

	to := r.PostForm.Get("to_verify")
	if !jsonschema.Formats["email"](to) {
		h.handleError(w, r, vr, errors.WithStack(schema.NewInvalidFormatError("#/to_verify", "email", to)))
		return
	}

	address, err := h.d.VerificationPersister().FindAddressByValue(r.Context(), ViaEmail, to)
	if err != nil {
		if errorsx.Cause(err) == sqlcon.ErrNoRows {
			if err := h.d.VerificationManager().sendToUnknownAddress(r.Context(), ViaEmail, to); err != nil {
				h.handleError(w, r, vr, err)
				return
			}
			return
		}
		h.handleError(w, r, vr, err)
		return
	}

	h.d.VerificationManager().sendCodeToKnownAddress(r.Context(), address)
}

// nolint:deadcode,unused
// swagger:parameters selfServiceBrowserVerify
type selfServiceBrowserVerifyParameters struct {
	// required: true
	// in: path
	Code string `json:"code"`
}

// swagger:route GET /self-service/browser/flows/verification/{code} public selfServiceBrowserVerify
//
// Complete the browser-based verification flows
//
// This endpoint completes a browser-based verification flow.
//
// > This endpoint is NOT INTENDED for API clients and only works with browsers (Chrome, Firefox, ...) and HTML Forms.
//
// More information can be found at [ORY Kratos Email and Phone Verification Documentation](https://www.ory.sh/docs/kratos/selfservice/flows/verify-email-account-activation).
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
func (h *Handler) verify(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	panic("not implemented")
	if err := h.d.VerificationManager().Verify(r.Context(), ps.ByName("code")); err != nil {

		// TODO ADD RETRY LINK?
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	// TODO CHECK EXPIRY

	http.Redirect(w, r, h.c.DefaultReturnToURL().String(), http.StatusFound)
}

// handleError is a convenience function for handling all types of errors that may occur (e.g. validation error).
func (h *Handler) handleError(w http.ResponseWriter, r *http.Request, rr *Request, err error) {
	if rr != nil {
		rr.Form.Reset()
		rr.Form.SetCSRF(h.d.GenerateCSRFToken(r))
	}

	h.d.VerificationRequestErrorHandler().HandleVerificationError(w, r, rr, err)
}

func (h *Handler) toVia(ps httprouter.Params) (Via, error) {
	v := ps.ByName("via")
	switch Via(v) {
	case ViaEmail:
		return ViaEmail, nil
	}
	return "", errors.WithStack(herodot.ErrBadRequest.WithReasonf("Verification only works for email but got: %s", v))
}
