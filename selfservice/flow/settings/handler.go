package settings

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/nosurf"
	"github.com/pkg/errors"

	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/schema"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	PublicPath        = "/self-service/browser/flows/settings"
	PublicRequestPath = "/self-service/browser/flows/requests/settings"

	ContinuityPrefix = "ory_kratos_settings"
)

func ContinuityKey(id string) string {
	return ContinuityPrefix + "_" + id
}

type (
	handlerDependencies interface {
		x.CSRFProvider
		x.WriterProvider
		x.LoggingProvider

		continuity.ManagementProvider

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
		SettingsHandler() *Handler
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
	redirect := session.RedirectOnUnauthenticated(h.c.SelfServiceFlowLoginUI().String())
	public.GET(PublicPath, h.d.SessionHandler().IsAuthenticated(h.initUpdateSettings, redirect))
	public.GET(PublicRequestPath, h.d.SessionHandler().IsAuthenticated(h.publicFetchUpdateSettingsRequest, redirect))
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(PublicRequestPath, h.adminFetchUpdateSettingsRequest)
}

// swagger:route GET /self-service/browser/flows/settings public initializeSelfServiceSettingsFlow
//
// Initialize Browser-Based Settings Flow
//
// This endpoint initializes a browser-based settings flow. Once initialized, the browser will be redirected to
// `selfservice.flows.settings.ui_url` with the request ID set as a query parameter. If no valid user session exists, a login
// flow will be initialized.
//
// > This endpoint is NOT INTENDED for API clients and only works
// with browsers (Chrome, Firefox, ...).
//
// More information can be found at [ORY Kratos User Settings & Profile Management Documentation](../self-service/flows/user-settings).
//
//     Schemes: http, https
//
//     Responses:
//       302: emptyResponse
//       500: genericError
func (h *Handler) initUpdateSettings(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	s, err := h.d.SessionManager().FetchFromRequest(r.Context(), r)
	if err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	req := NewRequest(h.c.SelfServiceFlowSettingsRequestLifespan(), r, s)

	if err := h.CreateRequest(w, r, s, req); err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	http.Redirect(w, r, req.URL(h.c.SelfServiceFlowSettingsUI()).String(), http.StatusFound)
}

func (h *Handler) CreateRequest(w http.ResponseWriter, r *http.Request, sess *session.Session, req *Request) error {
	for _, strategy := range h.d.SettingsStrategies() {
		if err := h.d.ContinuityManager().Abort(r.Context(), w, r, ContinuityKey(strategy.SettingsStrategyID())); err != nil {
			return err
		}

		if err := strategy.PopulateSettingsMethod(r, sess, req); err != nil {
			return err
		}
	}

	if err := h.d.SettingsRequestPersister().CreateSettingsRequest(r.Context(), req); err != nil {
		return err
	}

	return nil
}

// nolint:deadcode,unused
// swagger:parameters getSelfServiceBrowserSettingsRequest
type getSelfServiceBrowserSettingsRequestParameters struct {
	// Request is the Login Request ID
	//
	// The value for this parameter comes from `request` URL Query parameter sent to your
	// application (e.g. `/settingss?request=abcde`).
	//
	// required: true
	// in: query
	Request string `json:"request"`
}

// swagger:route GET /self-service/browser/flows/requests/settings common public admin getSelfServiceBrowserSettingsRequest
//
// Get the Request Context of Browser-Based Settings Flows
//
// When accessing this endpoint through ORY Kratos' Public API, ensure that cookies are set as they are required
// for checking the auth session. To prevent scanning attacks, the public endpoint does not return 404 status codes
// but instead 403 or 500.
//
// More information can be found at [ORY Kratos User Settings & Profile Management Documentation](../self-service/flows/user-settings).
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: settingsRequest
//       403: genericError
//       404: genericError
//       410: genericError
//       500: genericError
func (h *Handler) publicFetchUpdateSettingsRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.fetchUpdateSettingsRequest(w, r, true); err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}
}

func (h *Handler) adminFetchUpdateSettingsRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.fetchUpdateSettingsRequest(w, r, false); err != nil {
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

func (h *Handler) fetchUpdateSettingsRequest(w http.ResponseWriter, r *http.Request, checkSession bool) error {
	rid := x.ParseUUID(r.URL.Query().Get("request"))
	pr, err := h.d.SettingsRequestPersister().GetSettingsRequest(r.Context(), rid)
	if err != nil {
		return h.wrapErrorForbidden(err, checkSession)
	}

	if checkSession {
		sess, err := h.d.SessionManager().FetchFromRequest(r.Context(), r)
		if err != nil {
			return h.wrapErrorForbidden(err, checkSession)
		}

		if pr.IdentityID != sess.Identity.ID {
			return errors.WithStack(herodot.ErrForbidden.WithReasonf("The request was made for another identity and has been blocked for security reasons."))
		}
	}

	if pr.ExpiresAt.Before(time.Now().UTC()) {
		return errors.WithStack(x.ErrGone.
			WithReason("The settings request has expired. Redirect the user to the login endpoint to initialize a new session.").
			WithDetail("redirect_to", urlx.AppendPaths(h.c.SelfPublicURL(), PublicPath).String()))
	}

	h.d.Writer().Write(w, r, pr)
	return nil
}
