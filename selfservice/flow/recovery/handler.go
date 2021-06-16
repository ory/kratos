package recovery

import (
	"net/http"
	"time"

	"github.com/ory/kratos/schema"

	"github.com/ory/kratos/ui/node"
	"github.com/ory/x/sqlcon"

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
	public.GET(RouteInitBrowserFlow, h.d.SessionHandler().IsNotAuthenticated(h.initBrowserFlow, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if x.IsJSONRequest(r) {
			h.d.Writer().WriteError(w, r, errors.WithStack(ErrAlreadyLoggedIn))
		} else {
			redirect(w, r, ps)
		}
	}))

	public.GET(RouteInitAPIFlow, h.d.SessionHandler().IsNotAuthenticated(h.initAPIFlow,
		session.RespondWithJSONErrorOnAuthenticated(h.d.Writer(), ErrAlreadyLoggedIn)))

	public.GET(RouteGetFlow, h.fetch)

	public.GET(RouteSubmitFlow, h.d.SessionHandler().IsNotAuthenticated(h.submitFlow, redirect))
	public.POST(RouteSubmitFlow, h.d.SessionHandler().IsNotAuthenticated(h.submitFlow, redirect))
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(RouteGetFlow, h.fetch)
}

// swagger:route GET /self-service/recovery/api public initializeSelfServiceRecoveryWithoutBrowser
//
// Initialize Recovery Flow for APIs, Services, Apps, ...
//
// :::info
//
// This endpoint is EXPERIMENTAL and subject to potential breaking changes in the future.
//
// :::
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
// More information can be found at [Ory Kratos Account Recovery Documentation](../self-service/flows/account-recovery.mdx).
//
//     Schemes: http, https
//
//     Responses:
//       200: recoveryFlow
//       500: jsonError
//       400: jsonError
func (h *Handler) initAPIFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if !h.d.Config(r.Context()).SelfServiceFlowRecoveryEnabled() {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Recovery is not allowed because it was disabled.")))
		return
	}

	req, err := NewFlow(h.d.Config(r.Context()), h.d.Config(r.Context()).SelfServiceFlowRecoveryRequestLifespan(), h.d.GenerateCSRFToken(r), r, h.d.RecoveryStrategies(r.Context()), flow.TypeAPI)
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

// swagger:route GET /self-service/recovery/browser public initializeSelfServiceRecoveryForBrowsers
//
// Initialize Recovery Flow for Browsers
//
// :::info
//
// This endpoint is EXPERIMENTAL and subject to potential breaking changes in the future.
//
// :::
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
// More information can be found at [Ory Kratos Account Recovery Documentation](../self-service/flows/account-recovery.mdx).
//
//     Schemes: http, https
//
//     Responses:
//       200: recoveryFlow
//       302: emptyResponse
//       400: jsonError
//       500: jsonError
func (h *Handler) initBrowserFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if !h.d.Config(r.Context()).SelfServiceFlowRecoveryEnabled() {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Recovery is not allowed because it was disabled.")))
		return
	}

	f, err := NewFlow(h.d.Config(r.Context()), h.d.Config(r.Context()).SelfServiceFlowRecoveryRequestLifespan(), h.d.GenerateCSRFToken(r), r, h.d.RecoveryStrategies(r.Context()), flow.TypeBrowser)
	if err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	if err := h.d.RecoveryFlowPersister().CreateRecoveryFlow(r.Context(), f); err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	redirTo := f.AppendTo(h.d.Config(r.Context()).SelfServiceFlowRecoveryUI()).String()
	x.AcceptToRedirectOrJson(w, r, h.d.Writer(), f, redirTo)
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
// Get Recovery Flow
//
// :::info
//
// This endpoint is EXPERIMENTAL and subject to potential breaking changes in the future.
//
// :::
//
// This endpoint returns a recovery flow's context with, for example, error details and other information.
//
// More information can be found at [Ory Kratos Account Recovery Documentation](../self-service/flows/account-recovery.mdx).
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: recoveryFlow
//       404: jsonError
//       410: jsonError
//       500: jsonError
func (h *Handler) fetch(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if !h.d.Config(r.Context()).SelfServiceFlowRecoveryEnabled() {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Recovery is not allowed because it was disabled.")))
		return
	}

	rid := x.ParseUUID(r.URL.Query().Get("id"))
	f, err := h.d.RecoveryFlowPersister().GetRecoveryFlow(r.Context(), rid)
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	if f.ExpiresAt.Before(time.Now().UTC()) {
		if f.Type == flow.TypeBrowser {
			h.d.Writer().WriteError(w, r, errors.WithStack(x.ErrGone.
				WithReason("The recovery flow has expired. Redirect the user to the recovery flow init endpoint to initialize a new recovery flow.").
				WithDetail("redirect_to", urlx.AppendPaths(h.d.Config(r.Context()).SelfPublicURL(r), RouteInitBrowserFlow).String())))
			return
		}
		h.d.Writer().WriteError(w, r, errors.WithStack(x.ErrGone.
			WithReason("The recovery flow has expired. Call the recovery flow init API endpoint to initialize a new recovery flow.").
			WithDetail("api", urlx.AppendPaths(h.d.Config(r.Context()).SelfPublicURL(r), RouteInitAPIFlow).String())))
		return
	}

	h.d.Writer().Write(w, r, f)
}

// nolint:deadcode,unused
// swagger:parameters submitSelfServiceRecoveryFlow
type submitSelfServiceRecoveryFlow struct {
	// The Registration Flow ID
	//
	// The value for this parameter comes from `flow` URL Query parameter sent to your
	// application (e.g. `/registration?flow=abcde`).
	//
	// required: true
	// in: query
	Flow string `json:"flow"`

	// in: body
	Body submitSelfServiceRecoveryFlowBody
}

// swagger:model submitSelfServiceRecoveryFlowBody
type submitSelfServiceRecoveryFlowBody struct {
	// Method supports `link` only right now.
	//
	// enum:
	// - link
	// required: true
	Method string `json:"method"`

	// Email is the email to which to send the Recovery message to.
	//
	// required: true
	Email string `json:"email"`
}

// swagger:route POST /self-service/recovery public submitSelfServiceRecoveryFlow
//
// Complete Recovery Flow
//
// Use this endpoint to complete a recovery flow. This endpoint
// behaves differently for API and browser flows and has several states:
//
// - `choose_method` expects `flow` (in the URL query) and `email` (in the body) to be sent
//   and works with API- and Browser-initiated flows.
//	 - For API clients and Browser clients with HTTP Header `Accept: application/json` it either returns a HTTP 200 OK when the form is valid and HTTP 400 OK when the form is invalid.
//     and a HTTP 302 Found redirect with a fresh recovery flow if the flow was otherwise invalid (e.g. expired).
//	 - For Browser clients without HTTP Header `Accept` or with `Accept: text/*` it returns a HTTP 302 Found redirect to the Recovery UI URL with the Recovery Flow ID appended.
// - `sent_email` is the success state after `choose_method` for the `link` method and allows the user to request another recovery email. It
//   works for both API and Browser-initiated flows and returns the same responses as the flow in `choose_method` state.
// - `passed_challenge` expects a `token` to be sent in the URL query and given the nature of the flow ("sending a recovery link")
//   does not have any API capabilities. The server responds with a HTTP 302 Found redirect either to the Settings UI URL
//   (if the link was valid) and instructs the user to update their password, or a redirect to the Recover UI URL with
//   a new Recovery Flow ID which contains an error message that the recovery link was invalid.
//
// More information can be found at [Ory Kratos Account Recovery Documentation](../self-service/flows/account-recovery.mdx).
//
//     Consumes:
//     - application/json
//     - application/x-www-form-urlencoded
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: recoveryFlow
//       400: recoveryFlow
//       302: emptyResponse
//       500: jsonError
func (h *Handler) submitFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	var g node.Group
	var found bool
	for _, ss := range h.d.AllRecoveryStrategies() {
		err := ss.Recover(w, r, f)
		if errors.Is(err, flow.ErrStrategyNotResponsible) {
			continue
		} else if errors.Is(err, flow.ErrCompletedByStrategy) {
			return
		} else if err != nil {
			h.d.RecoveryFlowErrorHandler().WriteFlowError(w, r, f, ss.RecoveryNodeGroup(), err)
			return
		}

		found = true
		g = ss.RecoveryNodeGroup()
		break
	}

	if !found {
		h.d.RecoveryFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, errors.WithStack(schema.NewNoRecoveryStrategyResponsible()))
		return
	}

	if f.Type == flow.TypeBrowser && !x.IsJSONRequest(r) {
		http.Redirect(w, r, f.AppendTo(h.d.Config(r.Context()).SelfServiceFlowRecoveryUI()).String(), http.StatusSeeOther)
		return
	}

	updatedFlow, err := h.d.RecoveryFlowPersister().GetRecoveryFlow(r.Context(), f.ID)
	if err != nil {
		h.d.RecoveryFlowErrorHandler().WriteFlowError(w, r, f, g, err)
		return
	}

	h.d.Writer().Write(w, r, updatedFlow)
}
