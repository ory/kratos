package profile

import (
	"net/http"
	"net/url"
	"time"

	"github.com/ory/kratos/schema"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/nosurf"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	PublicProfileManagementPath = "/self-service/browser/flows/profile"
	BrowserProfileRequestPath   = "/self-service/browser/flows/requests/profile"
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
	public.GET(BrowserProfileRequestPath, h.d.SessionHandler().IsAuthenticated(h.publicFetchUpdateProfileRequest, redirect))
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(BrowserProfileRequestPath, h.adminFetchUpdateProfileRequest)
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
