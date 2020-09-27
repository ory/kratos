package recovery

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	RouteInitBrowserFlow = "/self-service/recovery/browser"
	RouteInitAPIFlow     = "/self-service/recovery/api"
	RouteGetFlow         = "/self-service/recovery/flows"
)

type (
	HandlerProvider interface {
		RecoveryHandler() *Handler
	}
	handlerDependencies interface {
		errorx.ManagementProvider
		identity.ManagementProvider
		identity.PrivilegedPoolProvider
		session.HandlerProvider
		StrategyProvider
		FlowPersistenceProvider
		x.CSRFTokenGeneratorProvider
		x.WriterProvider
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
	redirect := session.RedirectOnAuthenticated(h.c)
	public.GET(RouteInitBrowserFlow, h.d.SessionHandler().IsNotAuthenticated(h.initBrowserFlow, redirect))
	public.GET(RouteInitAPIFlow, h.d.SessionHandler().IsNotAuthenticated(h.initAPIFlow,
		session.RespondWithJSONErrorOnAuthenticated(h.d.Writer(), ErrAlreadyLoggedIn)))
	public.GET(RouteGetFlow, h.fetch)
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(RouteGetFlow, h.fetch)
}

// swagger:route GET /self-service/recovery/api public initializeSelfServiceRecoveryViaAPIFlow
//
// Initialize Recovery Flow for API Clients
//
// This endpoint initiates a recovery flow for API clients such as mobile devices, smart TVs, and so on.
//
// If a valid provided session cookie or session token is provided, a 400 Bad Request error.
//
// To fetch an existing recovery flow call `/self-service/recovery/flows?flow=<flow_id>`.
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
// More information can be found at [ORY Kratos Account Recovery Documentation](../self-service/flows/account-recovery.mdx).
//
//     Schemes: http, https
//
//     Security:
//     - sessionToken
//
//     Responses:
//       200: recoveryFlow
//       500: genericError
//       400: genericError
func (h *Handler) initAPIFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	req, err := NewFlow(h.c.SelfServiceFlowRecoveryRequestLifespan(), h.d.GenerateCSRFToken(r), r, h.d.RecoveryStrategies(), flow.TypeAPI)
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	if err := h.d.RecoveryFlowPersister().CreateRecoveryFlow(r.Context(), req); err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	h.d.Writer().Write(w, r, req)
}

// swagger:route GET /self-service/recovery/browser public initializeSelfServiceRecoveryViaBrowserFlow
//
// Initialize Recovery Flow for Browser Clients
//
// This endpoint initializes a browser-based account recovery flow. Once initialized, the browser will be redirected to
// `selfservice.flows.recovery.ui_url` with the flow ID set as the query parameter `?flow=`. If a valid user session
// exists, the browser is returned to the configured return URL.
//
// This endpoint is NOT INTENDED for API clients and only works with browsers (Chrome, Firefox, ...).
//
// More information can be found at [ORY Kratos Account Recovery Documentation](../self-service/flows/account-recovery.mdx).
//
//     Schemes: http, https
//
//     Security:
//     - sessionToken
//
//     Responses:
//       302: emptyResponse
//       500: genericError
func (h *Handler) initBrowserFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	req, err := NewFlow(h.c.SelfServiceFlowRecoveryRequestLifespan(), h.d.GenerateCSRFToken(r), r, h.d.RecoveryStrategies(), flow.TypeBrowser)
	if err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	if err := h.d.RecoveryFlowPersister().CreateRecoveryFlow(r.Context(), req); err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	http.Redirect(w, r, req.AppendTo(h.c.SelfServiceFlowRecoveryUI()).String(), http.StatusFound)
}

// nolint:deadcode,unused
// swagger:parameters getSelfServiceRecoveryFlow
type getSelfServiceRecoveryFlowParameters struct {
	// The Flow ID
	//
	// The value for this parameter comes from `request` URL Query parameter sent to your
	// application (e.g. `/recovery?flow=abcde`).
	//
	// required: true
	// in: query
	FlowID string `json:"id"`
}

// swagger:route GET /self-service/recovery/flows public admin getSelfServiceRecoveryFlow
//
// Get information about a recovery flow
//
// This endpoint returns a recovery flow's context with, for example, error details and other information.
//
// More information can be found at [ORY Kratos Account Recovery Documentation](../self-service/flows/account-recovery.mdx).
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: recoveryFlow
//       404: genericError
//       410: genericError
//       500: genericError
func (h *Handler) fetch(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	rid := x.ParseUUID(r.URL.Query().Get("id"))
	req, err := h.d.RecoveryFlowPersister().GetRecoveryFlow(r.Context(), rid)
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	if req.ExpiresAt.Before(time.Now().UTC()) {
		if req.Type == flow.TypeBrowser {
			h.d.Writer().WriteError(w, r, errors.WithStack(x.ErrGone.
				WithReason("The recovery flow has expired. Redirect the user to the recovery flow init endpoint to initialize a new recovery flow.").
				WithDetail("redirect_to", urlx.AppendPaths(h.c.SelfPublicURL(), RouteInitBrowserFlow).String())))
			return
		}
		h.d.Writer().WriteError(w, r, errors.WithStack(x.ErrGone.
			WithReason("The recovery flow has expired. Call the recovery flow init API endpoint to initialize a new recovery flow.").
			WithDetail("api", urlx.AppendPaths(h.c.SelfPublicURL(), RouteInitAPIFlow).String())))
		return
	}

	h.d.Writer().Write(w, r, req)
}
