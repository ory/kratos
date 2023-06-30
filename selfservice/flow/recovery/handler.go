// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"net/http"
	"time"

	"github.com/ory/nosurf"

	"github.com/ory/kratos/schema"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/ui/node"

	"github.com/ory/herodot"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/config"
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

	RouteSubmitFlow = "/self-service/recovery"
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
		x.CSRFProvider
		config.Provider
		ErrorHandlerProvider
		HookExecutorProvider
	}
	Handler struct {
		d handlerDependencies
	}
)

func NewHandler(d handlerDependencies) *Handler {
	return &Handler{d: d}
}

func (h *Handler) RegisterPublicRoutes(public *x.RouterPublic) {
	h.d.CSRFHandler().IgnorePath(RouteInitAPIFlow)
	h.d.CSRFHandler().IgnorePath(RouteSubmitFlow)

	redirect := session.RedirectOnAuthenticated(h.d)
	public.GET(RouteInitBrowserFlow, h.d.SessionHandler().IsNotAuthenticated(h.createBrowserRecoveryFlow, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if x.IsJSONRequest(r) {
			h.d.Writer().WriteError(w, r, errors.WithStack(ErrAlreadyLoggedIn))
		} else {
			redirect(w, r, ps)
		}
	}))

	public.GET(RouteInitAPIFlow, h.d.SessionHandler().IsNotAuthenticated(h.createNativeRecoveryFlow,
		session.RespondWithJSONErrorOnAuthenticated(h.d.Writer(), ErrAlreadyLoggedIn)))

	public.GET(RouteGetFlow, h.getRecoveryFlow)

	public.GET(RouteSubmitFlow, h.updateRecoveryFlow)
	public.POST(RouteSubmitFlow, h.updateRecoveryFlow)
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(RouteInitBrowserFlow, x.RedirectToPublicRoute(h.d))
	admin.GET(RouteInitAPIFlow, x.RedirectToPublicRoute(h.d))
	admin.GET(RouteGetFlow, x.RedirectToPublicRoute(h.d))
	admin.GET(RouteSubmitFlow, x.RedirectToPublicRoute(h.d))
	admin.POST(RouteSubmitFlow, x.RedirectToPublicRoute(h.d))
}

// swagger:route GET /self-service/recovery/api frontend createNativeRecoveryFlow
//
// # Create Recovery Flow for Native Apps
//
// This endpoint initiates a recovery flow for API clients such as mobile devices, smart TVs, and so on.
//
// If a valid provided session cookie or session token is provided, a 400 Bad Request error.
//
// To fetch an existing recovery flow call `/self-service/recovery/flows?flow=<flow_id>`.
//
// You MUST NOT use this endpoint in client-side (Single Page Apps, ReactJS, AngularJS) nor server-side (Java Server
// Pages, NodeJS, PHP, Golang, ...) browser applications. Using this endpoint in these applications will make
// you vulnerable to a variety of CSRF attacks.
//
// This endpoint MUST ONLY be used in scenarios such as native mobile apps (React Native, Objective C, Swift, Java, ...).
//
// More information can be found at [Ory Kratos Account Recovery Documentation](../self-service/flows/account-recovery).
//
//	Schemes: http, https
//
//	Responses:
//	  200: recoveryFlow
//	  400: errorGeneric
//	  default: errorGeneric
func (h *Handler) createNativeRecoveryFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if !h.d.Config().SelfServiceFlowRecoveryEnabled(r.Context()) {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Recovery is not allowed because it was disabled.")))
		return
	}
	activeRecoveryStrategy, err := h.d.GetActiveRecoveryStrategy(r.Context())
	if err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	req, err := NewFlow(h.d.Config(), h.d.Config().SelfServiceFlowRecoveryRequestLifespan(r.Context()), h.d.GenerateCSRFToken(r), r, activeRecoveryStrategy, flow.TypeAPI)
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	if err := h.d.RecoveryExecutor().PreRecoveryHook(w, r, req); err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	if err := h.d.RecoveryFlowPersister().CreateRecoveryFlow(r.Context(), req); err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	h.d.Writer().Write(w, r, req)
}

// Create Browser Recovery Flow Parameters
//
// swagger:parameters createBrowserRecoveryFlow
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type createBrowserRecoveryFlow struct {
	// The URL to return the browser to after the flow was completed.
	//
	// in: query
	ReturnTo string `json:"return_to"`
}

// swagger:route GET /self-service/recovery/browser frontend createBrowserRecoveryFlow
//
// # Create Recovery Flow for Browsers
//
// This endpoint initializes a browser-based account recovery flow. Once initialized, the browser will be redirected to
// `selfservice.flows.recovery.ui_url` with the flow ID set as the query parameter `?flow=`. If a valid user session
// exists, the browser is returned to the configured return URL.
//
// If this endpoint is called via an AJAX request, the response contains the recovery flow without any redirects
// or a 400 bad request error if the user is already authenticated.
//
// This endpoint is NOT INTENDED for clients that do not have a browser (Chrome, Firefox, ...) as cookies are needed.
//
// More information can be found at [Ory Kratos Account Recovery Documentation](../self-service/flows/account-recovery).
//
//	Schemes: http, https
//
//	Responses:
//	  200: recoveryFlow
//	  303: emptyResponse
//	  400: errorGeneric
//	  default: errorGeneric
func (h *Handler) createBrowserRecoveryFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	if !h.d.Config().SelfServiceFlowRecoveryEnabled(r.Context()) {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Recovery is not allowed because it was disabled.")))
		return
	}
	activeRecoveryStrategy, err := h.d.GetActiveRecoveryStrategy(r.Context())
	if err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	f, err := NewFlow(h.d.Config(), h.d.Config().SelfServiceFlowRecoveryRequestLifespan(r.Context()), h.d.GenerateCSRFToken(r), r, activeRecoveryStrategy, flow.TypeBrowser)
	if err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	if err := h.d.RecoveryExecutor().PreRecoveryHook(w, r, f); err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	if err := h.d.RecoveryFlowPersister().CreateRecoveryFlow(r.Context(), f); err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	redirTo := f.AppendTo(h.d.Config().SelfServiceFlowRecoveryUI(r.Context())).String()
	x.AcceptToRedirectOrJSON(w, r, h.d.Writer(), f, redirTo)
}

// Get Recovery Flow Parameters
//
// swagger:parameters getRecoveryFlow
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type getRecoveryFlow struct {
	// The Flow ID
	//
	// The value for this parameter comes from `request` URL Query parameter sent to your
	// application (e.g. `/recovery?flow=abcde`).
	//
	// required: true
	// in: query
	FlowID string `json:"id"`

	// HTTP Cookies
	//
	// When using the SDK in a browser app, on the server side you must include the HTTP Cookie Header
	// sent by the client to your server here. This ensures that CSRF and session cookies are respected.
	//
	// in: header
	// name: Cookie
	Cookies string `json:"Cookie"`
}

// swagger:route GET /self-service/recovery/flows frontend getRecoveryFlow
//
// # Get Recovery Flow
//
// This endpoint returns a recovery flow's context with, for example, error details and other information.
//
// Browser flows expect the anti-CSRF cookie to be included in the request's HTTP Cookie Header.
// For AJAX requests you must ensure that cookies are included in the request or requests will fail.
//
// If you use the browser-flow for server-side apps, the services need to run on a common top-level-domain
// and you need to forward the incoming HTTP Cookie header to this endpoint:
//
//	```js
//	// pseudo-code example
//	router.get('/recovery', async function (req, res) {
//	  const flow = await client.getRecoveryFlow(req.header('Cookie'), req.query['flow'])
//
//	  res.render('recovery', flow)
//	})
//	```
//
// More information can be found at [Ory Kratos Account Recovery Documentation](../self-service/flows/account-recovery).
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Responses:
//	  200: recoveryFlow
//	  404: errorGeneric
//	  410: errorGeneric
//	  default: errorGeneric
func (h *Handler) getRecoveryFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if !h.d.Config().SelfServiceFlowRecoveryEnabled(r.Context()) {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Recovery is not allowed because it was disabled.")))
		return
	}

	rid := x.ParseUUID(r.URL.Query().Get("id"))
	f, err := h.d.RecoveryFlowPersister().GetRecoveryFlow(r.Context(), rid)
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	// Browser flows must include the CSRF token
	//
	// Resolves: https://github.com/ory/kratos/issues/1282
	if f.Type.IsBrowser() && !f.DangerousSkipCSRFCheck && !nosurf.VerifyToken(h.d.GenerateCSRFToken(r), f.CSRFToken) {
		h.d.Writer().WriteError(w, r, x.CSRFErrorReason(r, h.d))
		return
	}

	if f.ExpiresAt.Before(time.Now().UTC()) {
		if f.Type == flow.TypeBrowser {
			redirectURL := flow.GetFlowExpiredRedirectURL(r.Context(), h.d.Config(), RouteInitBrowserFlow, f.ReturnTo)

			h.d.Writer().WriteError(w, r, errors.WithStack(x.ErrGone.
				WithReason("The recovery flow has expired. Redirect the user to the recovery flow init endpoint to initialize a new recovery flow.").
				WithDetail("redirect_to", redirectURL.String()).
				WithDetail("return_to", f.ReturnTo)))
			return
		}
		h.d.Writer().WriteError(w, r, errors.WithStack(x.ErrGone.
			WithReason("The recovery flow has expired. Call the recovery flow init API endpoint to initialize a new recovery flow.").
			WithDetail("api", urlx.AppendPaths(h.d.Config().SelfPublicURL(r.Context()), RouteInitAPIFlow).String())))
		return
	}

	h.d.Writer().Write(w, r, f)
}

// Update Recovery Flow Parameters
//
// swagger:parameters updateRecoveryFlow
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type updateRecoveryFlow struct {
	// The Recovery Flow ID
	//
	// The value for this parameter comes from `flow` URL Query parameter sent to your
	// application (e.g. `/recovery?flow=abcde`).
	//
	// required: true
	// in: query
	Flow string `json:"flow"`

	// Recovery Token
	//
	// The recovery token which completes the recovery request. If the token
	// is invalid (e.g. expired) an error will be shown to the end-user.
	//
	// This parameter is usually set in a link and not used by any direct API call.
	//
	// in: query
	Token string `json:"token" form:"token"`

	// in: body
	// required: true
	Body updateRecoveryFlowBody

	// HTTP Cookies
	//
	// When using the SDK in a browser app, on the server side you must include the HTTP Cookie Header
	// sent by the client to your server here. This ensures that CSRF and session cookies are respected.
	//
	// in: header
	// name: Cookie
	Cookies string `json:"Cookie"`
}

// Update Recovery Flow Request Body
//
// swagger:model updateRecoveryFlowBody
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type updateRecoveryFlowBody struct{}

// swagger:route POST /self-service/recovery frontend updateRecoveryFlow
//
// # Complete Recovery Flow
//
// Use this endpoint to complete a recovery flow. This endpoint
// behaves differently for API and browser flows and has several states:
//
//   - `choose_method` expects `flow` (in the URL query) and `email` (in the body) to be sent
//     and works with API- and Browser-initiated flows.
//   - For API clients and Browser clients with HTTP Header `Accept: application/json` it either returns a HTTP 200 OK when the form is valid and HTTP 400 OK when the form is invalid.
//     and a HTTP 303 See Other redirect with a fresh recovery flow if the flow was otherwise invalid (e.g. expired).
//   - For Browser clients without HTTP Header `Accept` or with `Accept: text/*` it returns a HTTP 303 See Other redirect to the Recovery UI URL with the Recovery Flow ID appended.
//   - `sent_email` is the success state after `choose_method` for the `link` method and allows the user to request another recovery email. It
//     works for both API and Browser-initiated flows and returns the same responses as the flow in `choose_method` state.
//   - `passed_challenge` expects a `token` to be sent in the URL query and given the nature of the flow ("sending a recovery link")
//     does not have any API capabilities. The server responds with a HTTP 303 See Other redirect either to the Settings UI URL
//     (if the link was valid) and instructs the user to update their password, or a redirect to the Recover UI URL with
//     a new Recovery Flow ID which contains an error message that the recovery link was invalid.
//
// More information can be found at [Ory Kratos Account Recovery Documentation](../self-service/flows/account-recovery).
//
//		Consumes:
//		- application/json
//		- application/x-www-form-urlencoded
//
//		Produces:
//		- application/json
//
//		Schemes: http, https
//
//	    Responses:
//	      200: recoveryFlow
//	      303: emptyResponse
//	      400: recoveryFlow
//	      410: errorGeneric
//	      422: errorBrowserLocationChangeRequired
//	      default: errorGeneric
func (h *Handler) updateRecoveryFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	rid, err := flow.GetFlowID(r)
	if err != nil {
		h.d.RecoveryFlowErrorHandler().WriteFlowError(w, r, nil, node.DefaultGroup, err)
		return
	}

	f, err := h.d.RecoveryFlowPersister().GetRecoveryFlow(r.Context(), rid)
	if errors.Is(err, sqlcon.ErrNoRows) {
		h.d.RecoveryFlowErrorHandler().WriteFlowError(w, r, nil, node.DefaultGroup, errors.WithStack(herodot.ErrNotFound.WithReasonf("The recovery request could not be found. Please restart the flow.")))
		return
	} else if err != nil {
		h.d.RecoveryFlowErrorHandler().WriteFlowError(w, r, nil, node.DefaultGroup, err)
		return
	}

	if err := f.Valid(); err != nil {
		h.d.RecoveryFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, err)
		return
	}

	var g node.UiNodeGroup
	var found bool
	for _, ss := range h.d.AllRecoveryStrategies() {
		err := ss.Recover(w, r, f)
		if errors.Is(err, flow.ErrStrategyNotResponsible) {
			continue
		} else if errors.Is(err, flow.ErrCompletedByStrategy) {
			return
		} else if err != nil {
			h.d.RecoveryFlowErrorHandler().WriteFlowError(w, r, f, ss.NodeGroup(), err)
			return
		}

		found = true
		g = ss.NodeGroup()
		break
	}

	if !found {
		h.d.RecoveryFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, errors.WithStack(schema.NewNoRecoveryStrategyResponsible()))
		return
	}

	if f.Type == flow.TypeBrowser && !x.IsJSONRequest(r) {
		http.Redirect(w, r, f.AppendTo(h.d.Config().SelfServiceFlowRecoveryUI(r.Context())).String(), http.StatusSeeOther)
		return
	}

	updatedFlow, err := h.d.RecoveryFlowPersister().GetRecoveryFlow(r.Context(), f.ID)
	if err != nil {
		h.d.RecoveryFlowErrorHandler().WriteFlowError(w, r, f, g, err)
		return
	}

	h.d.Writer().Write(w, r, updatedFlow)
}
