package settings

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/nosurf"
	"github.com/pkg/errors"

	"github.com/ory/x/urlx"

	"github.com/ory/herodot"

	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	RouteInitBrowserFlow = "/self-service/settings/browser"
	RouteInitAPIFlow     = "/self-service/settings/api"
	RouteGetFlow         = "/self-service/settings/flows"

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
		FlowPersistenceProvider
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
	public.GET(RouteInitBrowserFlow, h.d.SessionHandler().IsAuthenticated(h.initBrowserFlow, redirect))
	public.GET(RouteInitAPIFlow, h.d.SessionHandler().IsAuthenticated(h.initApiFlow, nil))

	public.GET(RouteGetFlow, h.d.SessionHandler().IsAuthenticated(h.fetchPublicFLow, OnUnauthenticated(h.c, h.d)))
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(RouteGetFlow, h.fetchAdminFlow)
}

func (h *Handler) NewFlow(w http.ResponseWriter, r *http.Request, i *identity.Identity, ft flow.Type) (*Flow, error) {
	f := NewFlow(h.c.SelfServiceFlowSettingsFlowLifespan(), r, i, ft)
	for _, strategy := range h.d.SettingsStrategies() {
		if err := h.d.ContinuityManager().Abort(r.Context(), w, r, ContinuityKey(strategy.SettingsStrategyID())); err != nil {
			return nil, err
		}

		if err := strategy.PopulateSettingsMethod(r, i, f); err != nil {
			return nil, err
		}
	}

	if err := h.d.SettingsFlowPersister().CreateSettingsFlow(r.Context(), f); err != nil {
		return nil, err
	}

	return f, nil
}

// swagger:route GET /self-service/settings/api public initializeSelfServiceSettingsViaAPIFlow
//
// Initialize Settings Flow for API Clients
//
// This endpoint initiates a settings flow for API clients such as mobile devices, smart TVs, and so on.
// You must provide a valid ORY Kratos Session Token for this endpoint to respond with HTTP 200 OK.
//
// To fetch an existing settings flow call `/self-service/settings/flows?flow=<flow_id>`.
//
// :::warning
//
// You MUST NOT use this endpoint in client-side (Single Page Apps, ReactJS, AngularJS) nor server-side (Java Server
// Pages, NodeJS, PHP, Golang, ...) browser applications. Using this endpoint in these applications will make
// you vulnerable to a variety of CSRF attacks.
//
// This endpoint MUST ONLY be used in scenarios such as native mobile apps (React Native, Objective C, Swift, Java, ...).
//
// :::
//
// More information can be found at [ORY Kratos User Settings & Profile Management Documentation](../self-service/flows/user-settings).
//
//     Schemes: http, https
//
//     Security:
//     - sessionToken
//
//     Responses:
//       200: settingsFlow
//       400: genericError
//       500: genericError
func (h *Handler) initApiFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	s, err := h.d.SessionManager().FetchFromRequest(r.Context(), r)
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	f, err := h.NewFlow(w, r, s.Identity, flow.TypeAPI)
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	h.d.Writer().Write(w, r, f)
}

// swagger:route GET /self-service/settings/browser/flows public initializeSelfServiceSettingsViaBrowserFlow
//
// Initialize Settings Flow for Browsers
//
// This endpoint initializes a browser-based user settings flow. Once initialized, the browser will be redirected to
// `selfservice.flows.settings.ui_url` with the flow ID set as the query parameter `?flow=`. If no valid
// ORY Kratos Session Cookie is included in the request, a login flow will be initialized.
//
// :::note
//
// This endpoint is NOT INTENDED for API clients and only works with browsers (Chrome, Firefox, ...).
//
// :::
//
// More information can be found at [ORY Kratos User Settings & Profile Management Documentation](../self-service/flows/user-settings).
//
//     Schemes: http, https
//
//     Security:
//     - sessionToken
//
//     Responses:
//       302: emptyResponse
//       500: genericError
func (h *Handler) initBrowserFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	s, err := h.d.SessionManager().FetchFromRequest(r.Context(), r)
	if err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	f, err := h.NewFlow(w, r, s.Identity, flow.TypeBrowser)
	if err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	http.Redirect(w, r, f.AppendTo(h.c.SelfServiceFlowSettingsUI()).String(), http.StatusFound)
}

// nolint:deadcode,unused
// swagger:parameters getSelfServiceSettingsFlow
type getSelfServiceSettingsFlowParameters struct {
	// ID is the Settings Flow ID
	//
	// The value for this parameter comes from `flow` URL Query parameter sent to your
	// application (e.g. `/settings?flow=abcde`).
	//
	// required: true
	// in: query
	ID string `json:"id"`
}

// swagger:route GET /self-service/settings/flows common public admin getSelfServiceSettingsFlow
//
// Get Settings Flow
//
// When accessing this endpoint through ORY Kratos' Public API you must ensure that either the ORY Kratos Session Cookie
// or the ORY Kratos Session Token are set. The public endpoint does not return 404 status codes
// but instead 403 or 500 to improve data privacy.
//
// You can access this endpoint without credentials when using ORY Kratos' Admin API.
//
// More information can be found at [ORY Kratos User Settings & Profile Management Documentation](../self-service/flows/user-settings).
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Security:
//     - sessionToken
//
//     Responses:
//       200: settingsFlow
//       403: genericError
//       404: genericError
//       410: genericError
//       500: genericError
func (h *Handler) fetchPublicFLow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.fetchFlow(w, r, true); err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}
}

func (h *Handler) fetchAdminFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.fetchFlow(w, r, false); err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}
}

func (h *Handler) wrapErrorForbidden(err error, shouldWrap bool) error {
	if shouldWrap {
		return herodot.ErrForbidden.
			WithReasonf("Access privileges are missing, invalid, or not sufficient to access this endpoint.").
			WithTrace(err).WithDebugf("%s", err)
	}

	return err
}

func (h *Handler) fetchFlow(w http.ResponseWriter, r *http.Request, checkSession bool) error {
	rid := x.ParseUUID(r.URL.Query().Get("id"))
	pr, err := h.d.SettingsFlowPersister().GetSettingsFlow(r.Context(), rid)
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
		if pr.Type == flow.TypeBrowser {
			h.d.Writer().WriteError(w, r, errors.WithStack(x.ErrGone.
				WithReason("The settings flow has expired. Redirect the user to the settings flow init endpoint to initialize a new settings flow.").
				WithDetail("redirect_to", urlx.AppendPaths(h.c.SelfPublicURL(), RouteInitBrowserFlow).String())))
			return nil
		}
		h.d.Writer().WriteError(w, r, errors.WithStack(x.ErrGone.
			WithReason("The settings flow has expired. Call the settings flow init API endpoint to initialize a new settings flow.").
			WithDetail("api", urlx.AppendPaths(h.c.SelfPublicURL(), RouteInitAPIFlow).String())))
		return nil
	}

	h.d.Writer().Write(w, r, pr)
	return nil
}
