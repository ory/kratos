// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package verification

import (
	"net/http"
	"time"

	"github.com/ory/nosurf"

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
	"github.com/ory/kratos/x"
)

const (
	RouteInitBrowserFlow = "/self-service/verification/browser"
	RouteInitAPIFlow     = "/self-service/verification/api"
	RouteGetFlow         = "/self-service/verification/flows"

	RouteSubmitFlow = "/self-service/verification"
)

type (
	HandlerProvider interface {
		VerificationHandler() *Handler
	}
	handlerDependencies interface {
		errorx.ManagementProvider
		identity.ManagementProvider
		identity.PrivilegedPoolProvider
		config.Provider

		x.CSRFTokenGeneratorProvider
		x.WriterProvider
		x.CSRFProvider

		FlowPersistenceProvider
		ErrorHandlerProvider
		StrategyProvider
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

	public.GET(RouteInitBrowserFlow, h.createBrowserVerificationFlow)
	public.GET(RouteInitAPIFlow, h.createNativeVerificationFlow)
	public.GET(RouteGetFlow, h.getVerificationFlow)

	public.POST(RouteSubmitFlow, h.updateVerificationFlow)
	public.GET(RouteSubmitFlow, h.updateVerificationFlow)
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(RouteInitBrowserFlow, x.RedirectToPublicRoute(h.d))
	admin.GET(RouteInitAPIFlow, x.RedirectToPublicRoute(h.d))
	admin.GET(RouteGetFlow, x.RedirectToPublicRoute(h.d))

	admin.POST(RouteSubmitFlow, x.RedirectToPublicRoute(h.d))
	admin.GET(RouteSubmitFlow, x.RedirectToPublicRoute(h.d))
}

type FlowOption func(f *Flow)

func WithFlowReturnTo(returnTo string) FlowOption {
	return func(f *Flow) {
		f.ReturnTo = returnTo
	}
}

func (h *Handler) NewVerificationFlow(w http.ResponseWriter, r *http.Request, ft flow.Type, opts ...FlowOption) (*Flow, error) {
	strategy, err := h.d.GetActiveVerificationStrategy(r.Context())
	if err != nil {
		return nil, err
	}

	f, err := NewFlow(h.d.Config(), h.d.Config().SelfServiceFlowVerificationRequestLifespan(r.Context()), h.d.GenerateCSRFToken(r), r, strategy, ft)
	if err != nil {
		return nil, err
	}
	for _, o := range opts {
		o(f)
	}

	if err := h.d.VerificationExecutor().PreVerificationHook(w, r, f); err != nil {
		return nil, err
	}

	if err := h.d.VerificationFlowPersister().CreateVerificationFlow(r.Context(), f); err != nil {
		return nil, err
	}

	return f, nil
}

// swagger:route GET /self-service/verification/api frontend createNativeVerificationFlow
//
// # Create Verification Flow for Native Apps
//
// This endpoint initiates a verification flow for API clients such as mobile devices, smart TVs, and so on.
//
// To fetch an existing verification flow call `/self-service/verification/flows?flow=<flow_id>`.
//
// You MUST NOT use this endpoint in client-side (Single Page Apps, ReactJS, AngularJS) nor server-side (Java Server
// Pages, NodeJS, PHP, Golang, ...) browser applications. Using this endpoint in these applications will make
// you vulnerable to a variety of CSRF attacks.
//
// This endpoint MUST ONLY be used in scenarios such as native mobile apps (React Native, Objective C, Swift, Java, ...).
//
// More information can be found at [Ory Email and Phone Verification Documentation](https://www.ory.sh/docs/kratos/self-service/flows/verify-email-account-activation).
//
//	Schemes: http, https
//
//	Responses:
//	  200: verificationFlow
//	  400: errorGeneric
//	  default: errorGeneric
func (h *Handler) createNativeVerificationFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if !h.d.Config().SelfServiceFlowVerificationEnabled(r.Context()) {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Verification is not allowed because it was disabled.")))
		return
	}

	req, err := h.NewVerificationFlow(w, r, flow.TypeAPI)
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	h.d.Writer().Write(w, r, req)
}

// Create Browser Verification Flow Parameters
//
// swagger:parameters createBrowserVerificationFlow
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type createBrowserVerificationFlow struct {
	// The URL to return the browser to after the flow was completed.
	//
	// in: query
	ReturnTo string `json:"return_to"`
}

// swagger:route GET /self-service/verification/browser frontend createBrowserVerificationFlow
//
// # Create Verification Flow for Browser Clients
//
// This endpoint initializes a browser-based account verification flow. Once initialized, the browser will be redirected to
// `selfservice.flows.verification.ui_url` with the flow ID set as the query parameter `?flow=`.
//
// If this endpoint is called via an AJAX request, the response contains the recovery flow without any redirects.
//
// This endpoint is NOT INTENDED for API clients and only works with browsers (Chrome, Firefox, ...).
//
// More information can be found at [Ory Kratos Email and Phone Verification Documentation](https://www.ory.sh/docs/kratos/self-service/flows/verify-email-account-activation).
//
//	Schemes: http, https
//
//	Responses:
//	  200: verificationFlow
//	  303: emptyResponse
//	  default: errorGeneric
func (h *Handler) createBrowserVerificationFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if !h.d.Config().SelfServiceFlowVerificationEnabled(r.Context()) {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Verification is not allowed because it was disabled.")))
		return
	}

	req, err := h.NewVerificationFlow(w, r, flow.TypeBrowser)
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	redirTo := req.AppendTo(h.d.Config().SelfServiceFlowVerificationUI(r.Context())).String()
	x.AcceptToRedirectOrJSON(w, r, h.d.Writer(), req, redirTo)
}

// Get Verification Flow Parameters
//
// swagger:parameters getVerificationFlow
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type getVerificationFlow struct {
	// The Flow ID
	//
	// The value for this parameter comes from `request` URL Query parameter sent to your
	// application (e.g. `/verification?flow=abcde`).
	//
	// required: true
	// in: query
	FlowID string `json:"id"`

	// HTTP Cookies
	//
	// When using the SDK on the server side you must include the HTTP Cookie Header
	// originally sent to your HTTP handler here.
	//
	// in: header
	// name: Cookie
	Cookie string `json:"cookie"`
}

// swagger:route GET /self-service/verification/flows frontend getVerificationFlow
//
// # Get Verification Flow
//
// This endpoint returns a verification flow's context with, for example, error details and other information.
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
//	  const flow = await client.getVerificationFlow(req.header('cookie'), req.query['flow'])
//
//	  res.render('verification', flow)
//	})
//	```
//
// More information can be found at [Ory Kratos Email and Phone Verification Documentation](https://www.ory.sh/docs/kratos/self-service/flows/verify-email-account-activation).
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Responses:
//	  200: verificationFlow
//	  403: errorGeneric
//	  404: errorGeneric
//	  default: errorGeneric
func (h *Handler) getVerificationFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if !h.d.Config().SelfServiceFlowVerificationEnabled(r.Context()) {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Verification is not allowed because it was disabled.")))
		return
	}

	rid := x.ParseUUID(r.URL.Query().Get("id"))
	req, err := h.d.VerificationFlowPersister().GetVerificationFlow(r.Context(), rid)
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	// Browser flows must include the CSRF token
	//
	// Resolves: https://github.com/ory/kratos/issues/1282
	if req.Type == flow.TypeBrowser && !nosurf.VerifyToken(h.d.GenerateCSRFToken(r), req.CSRFToken) {
		h.d.Writer().WriteError(w, r, x.CSRFErrorReason(r, h.d))
		return
	}

	if req.ExpiresAt.Before(time.Now().UTC()) {
		if req.Type == flow.TypeBrowser {
			redirectURL := flow.GetFlowExpiredRedirectURL(r.Context(), h.d.Config(), RouteInitBrowserFlow, req.ReturnTo)

			h.d.Writer().WriteError(w, r, errors.WithStack(x.ErrGone.
				WithReason("The verification flow has expired. Redirect the user to the verification flow init endpoint to initialize a new verification flow.").
				WithDetail("redirect_to", redirectURL.String()).
				WithDetail("return_to", req.ReturnTo)))
			return
		}
		h.d.Writer().WriteError(w, r, errors.WithStack(x.ErrGone.
			WithReason("The verification flow has expired. Call the verification flow init API endpoint to initialize a new verification flow.").
			WithDetail("api", urlx.AppendPaths(h.d.Config().SelfPublicURL(r.Context()), RouteInitAPIFlow).String())))
		return
	}

	h.d.Writer().Write(w, r, req)
}

// Update Verification Flow Parameters
//
// swagger:parameters updateVerificationFlow
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type updateVerificationFlow struct {
	// The Verification Flow ID
	//
	// The value for this parameter comes from `flow` URL Query parameter sent to your
	// application (e.g. `/verification?flow=abcde`).
	//
	// required: true
	// in: query
	Flow string `json:"flow"`

	// Verification Token
	//
	// The verification token which completes the verification request. If the token
	// is invalid (e.g. expired) an error will be shown to the end-user.
	//
	// This parameter is usually set in a link and not used by any direct API call.
	//
	// in: query
	Token string `json:"token" form:"token"`

	// in: body
	// required: true
	Body updateVerificationFlowBody

	// HTTP Cookies
	//
	// When using the SDK in a browser app, on the server side you must include the HTTP Cookie Header
	// sent by the client to your server here. This ensures that CSRF and session cookies are respected.
	//
	// in: header
	// name: Cookie
	Cookies string `json:"Cookie"`
}

// Update Verification Flow Request Body
//
// swagger:model updateVerificationFlowBody
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type updateVerificationFlowBody struct{}

// swagger:route POST /self-service/verification frontend updateVerificationFlow
//
// # Complete Verification Flow
//
// Use this endpoint to complete a verification flow. This endpoint
// behaves differently for API and browser flows and has several states:
//
//   - `choose_method` expects `flow` (in the URL query) and `email` (in the body) to be sent
//     and works with API- and Browser-initiated flows.
//   - For API clients and Browser clients with HTTP Header `Accept: application/json` it either returns a HTTP 200 OK when the form is valid and HTTP 400 OK when the form is invalid
//     and a HTTP 303 See Other redirect with a fresh verification flow if the flow was otherwise invalid (e.g. expired).
//   - For Browser clients without HTTP Header `Accept` or with `Accept: text/*` it returns a HTTP 303 See Other redirect to the Verification UI URL with the Verification Flow ID appended.
//   - `sent_email` is the success state after `choose_method` when using the `link` method and allows the user to request another verification email. It
//     works for both API and Browser-initiated flows and returns the same responses as the flow in `choose_method` state.
//   - `passed_challenge` expects a `token` to be sent in the URL query and given the nature of the flow ("sending a verification link")
//     does not have any API capabilities. The server responds with a HTTP 303 See Other redirect either to the Settings UI URL
//     (if the link was valid) and instructs the user to update their password, or a redirect to the Verification UI URL with
//     a new Verification Flow ID which contains an error message that the verification link was invalid.
//
// More information can be found at [Ory Kratos Email and Phone Verification Documentation](https://www.ory.sh/docs/kratos/self-service/flows/verify-email-account-activation).
//
//	Consumes:
//	- application/json
//	- application/x-www-form-urlencoded
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Responses:
//	  200: verificationFlow
//	  303: emptyResponse
//	  400: verificationFlow
//	  410: errorGeneric
//	  default: errorGeneric
func (h *Handler) updateVerificationFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	rid, err := flow.GetFlowID(r)
	if err != nil {
		h.d.VerificationFlowErrorHandler().WriteFlowError(w, r, nil, node.DefaultGroup, err)
		return
	}

	f, err := h.d.VerificationFlowPersister().GetVerificationFlow(r.Context(), rid)
	if errors.Is(err, sqlcon.ErrNoRows) {
		h.d.VerificationFlowErrorHandler().WriteFlowError(w, r, nil, node.DefaultGroup, errors.WithStack(herodot.ErrNotFound.WithReasonf("The verification request could not be found. Please restart the flow.")))
		return
	} else if err != nil {
		h.d.VerificationFlowErrorHandler().WriteFlowError(w, r, nil, node.DefaultGroup, err)
		return
	}

	if err := f.Valid(); err != nil {
		h.d.VerificationFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, err)
		return
	}

	var g node.UiNodeGroup
	var found bool
	for _, ss := range h.d.AllVerificationStrategies() {
		// If an active strategy is set, but it does not match the current strategy, that strategy is not responsible anyways.
		if f.Active.String() != "" && f.Active.String() != ss.VerificationStrategyID() {
			continue
		}

		err := ss.Verify(w, r, f)
		if errors.Is(err, flow.ErrStrategyNotResponsible) {
			continue
		} else if errors.Is(err, flow.ErrCompletedByStrategy) {
			return
		} else if err != nil {
			h.d.VerificationFlowErrorHandler().WriteFlowError(w, r, f, ss.VerificationNodeGroup(), err)
			return
		}

		found = true
		g = ss.VerificationNodeGroup()
		break
	}

	if !found {
		h.d.VerificationFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, errors.WithStack(schema.NewNoVerificationStrategyResponsible()))
		return
	}

	if x.IsBrowserRequest(r) {
		http.Redirect(w, r, f.AppendTo(h.d.Config().SelfServiceFlowVerificationUI(r.Context())).String(), http.StatusSeeOther)
		return
	}

	updatedFlow, err := h.d.VerificationFlowPersister().GetVerificationFlow(r.Context(), f.ID)
	if err != nil {
		h.d.VerificationFlowErrorHandler().WriteFlowError(w, r, f, g, err)
		return
	}

	h.d.Writer().Write(w, r, updatedFlow)
}
