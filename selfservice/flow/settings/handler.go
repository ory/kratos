// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package settings

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/nosurf"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
)

const (
	RouteInitBrowserFlow = "/self-service/settings/browser"
	RouteInitAPIFlow     = "/self-service/settings/api"
	RouteGetFlow         = "/self-service/settings/flows"

	RouteSubmitFlow = "/self-service/settings"

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

		config.Provider

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
		HookExecutorProvider

		schema.IdentityTraitsProvider
	}
	HandlerProvider interface {
		SettingsHandler() *Handler
	}
	Handler struct {
		d    handlerDependencies
		csrf x.CSRFToken
	}
)

func NewHandler(d handlerDependencies) *Handler {
	return &Handler{d: d, csrf: nosurf.Token}
}

func (h *Handler) RegisterPublicRoutes(public *x.RouterPublic) {
	h.d.CSRFHandler().IgnorePath(RouteInitAPIFlow)
	h.d.CSRFHandler().IgnorePath(RouteSubmitFlow)

	public.GET(RouteInitBrowserFlow, h.d.SessionHandler().IsAuthenticated(h.createBrowserSettingsFlow, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if x.IsJSONRequest(r) {
			h.d.Writer().WriteError(w, r, session.NewErrNoActiveSessionFound())
		} else {
			loginFlowUrl := h.d.Config().SelfPublicURL(r.Context()).JoinPath(login.RouteInitBrowserFlow).String()
			redirectUrl, err := x.TakeOverReturnToParameter(r.URL.String(), loginFlowUrl)
			if err != nil {
				http.Redirect(w, r, h.d.Config().SelfServiceFlowLoginUI(r.Context()).String(), http.StatusSeeOther)
			} else {
				http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
			}
		}
	}))

	public.GET(RouteInitAPIFlow, h.d.SessionHandler().IsAuthenticated(h.createNativeSettingsFlow, nil))
	public.GET(RouteGetFlow, h.d.SessionHandler().IsAuthenticated(h.getSettingsFlow, OnUnauthenticated(h.d)))

	public.POST(RouteSubmitFlow, h.d.SessionHandler().IsAuthenticated(h.updateSettingsFlow, OnUnauthenticated(h.d)))
	public.GET(RouteSubmitFlow, h.d.SessionHandler().IsAuthenticated(h.updateSettingsFlow, OnUnauthenticated(h.d)))
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(RouteInitBrowserFlow, x.RedirectToPublicRoute(h.d))

	admin.GET(RouteInitAPIFlow, x.RedirectToPublicRoute(h.d))
	admin.GET(RouteGetFlow, x.RedirectToPublicRoute(h.d))

	admin.POST(RouteSubmitFlow, x.RedirectToPublicRoute(h.d))
	admin.GET(RouteSubmitFlow, x.RedirectToPublicRoute(h.d))
}

func (h *Handler) NewFlow(w http.ResponseWriter, r *http.Request, i *identity.Identity, ft flow.Type) (*Flow, error) {
	f, err := NewFlow(h.d.Config(), h.d.Config().SelfServiceFlowSettingsFlowLifespan(r.Context()), r, i, ft)
	if err != nil {
		return nil, err
	}

	if err := h.d.SettingsHookExecutor().PreSettingsHook(w, r, f); err != nil {
		return nil, err
	}

	for _, strategy := range h.d.SettingsStrategies(r.Context()) {
		if err := h.d.ContinuityManager().Abort(r.Context(), w, r, ContinuityKey(strategy.SettingsStrategyID())); err != nil {
			return nil, err
		}

		if err := strategy.PopulateSettingsMethod(r, i, f); err != nil {
			return nil, err
		}
	}

	ds, err := h.d.Config().DefaultIdentityTraitsSchemaURL(r.Context())
	if err != nil {
		return nil, err
	}

	if err := sortNodes(r.Context(), f.UI.Nodes, ds.String()); err != nil {
		return nil, err
	}

	if err := h.d.SettingsFlowPersister().CreateSettingsFlow(r.Context(), f); err != nil {
		return nil, err
	}

	return f, nil
}

func (h *Handler) FromOldFlow(w http.ResponseWriter, r *http.Request, i *identity.Identity, of Flow) (*Flow, error) {
	nf, err := h.NewFlow(w, r, i, of.Type)
	if err != nil {
		return nil, err
	}

	nf.RequestURL = of.RequestURL
	return nf, nil
}

// Create Native Settings Flow Parameters
//
// swagger:parameters createNativeSettingsFlow
// nolint:deadcode,unused
type createNativeSettingsFlow struct {
	// The Session Token of the Identity performing the settings flow.
	//
	// in: header
	SessionToken string `json:"X-Session-Token"`
}

// swagger:route GET /self-service/settings/api frontend createNativeSettingsFlow
//
// # Create Settings Flow for Native Apps
//
// This endpoint initiates a settings flow for API clients such as mobile devices, smart TVs, and so on.
// You must provide a valid Ory Kratos Session Token for this endpoint to respond with HTTP 200 OK.
//
// To fetch an existing settings flow call `/self-service/settings/flows?flow=<flow_id>`.
//
// You MUST NOT use this endpoint in client-side (Single Page Apps, ReactJS, AngularJS) nor server-side (Java Server
// Pages, NodeJS, PHP, Golang, ...) browser applications. Using this endpoint in these applications will make
// you vulnerable to a variety of CSRF attacks.
//
// Depending on your configuration this endpoint might return a 403 error if the session has a lower Authenticator
// Assurance Level (AAL) than is possible for the identity. This can happen if the identity has password + webauthn
// credentials (which would result in AAL2) but the session has only AAL1. If this error occurs, ask the user
// to sign in with the second factor or change the configuration.
//
// In the case of an error, the `error.id` of the JSON response body can be one of:
//
// - `security_csrf_violation`: Unable to fetch the flow because a CSRF violation occurred.
// - `session_inactive`: No Ory Session was found - sign in a user first.
//
// This endpoint MUST ONLY be used in scenarios such as native mobile apps (React Native, Objective C, Swift, Java, ...).
//
// More information can be found at [Ory Kratos User Settings & Profile Management Documentation](../self-service/flows/user-settings).
//
//	   Schemes: http, https
//
//	   Responses:
//		  200: settingsFlow
//		  400: errorGeneric
//		  default: errorGeneric
func (h *Handler) createNativeSettingsFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	s, err := h.d.SessionManager().FetchFromRequest(r.Context(), r)
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	if err := h.d.SessionManager().DoesSessionSatisfy(r, s, h.d.Config().SelfServiceSettingsRequiredAAL(r.Context())); err != nil {
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

// Create Browser Settings Flow Parameters
//
// nolint:deadcode,unused
// swagger:parameters createBrowserSettingsFlow
type createBrowserSettingsFlow struct {
	// The URL to return the browser to after the flow was completed.
	//
	// in: query
	ReturnTo string `json:"return_to"`

	// HTTP Cookies
	//
	// When using the SDK in a browser app, on the server side you must include the HTTP Cookie Header
	// sent by the client to your server here. This ensures that CSRF and session cookies are respected.
	//
	// in: header
	// name: Cookie
	Cookies string `json:"Cookie"`
}

// swagger:route GET /self-service/settings/browser frontend createBrowserSettingsFlow
//
// # Create Settings Flow for Browsers
//
// This endpoint initializes a browser-based user settings flow. Once initialized, the browser will be redirected to
// `selfservice.flows.settings.ui_url` with the flow ID set as the query parameter `?flow=`. If no valid
// Ory Kratos Session Cookie is included in the request, a login flow will be initialized.
//
// If this endpoint is opened as a link in the browser, it will be redirected to
// `selfservice.flows.settings.ui_url` with the flow ID set as the query parameter `?flow=`. If no valid user session
// was set, the browser will be redirected to the login endpoint.
//
// If this endpoint is called via an AJAX request, the response contains the settings flow without any redirects
// or a 401 forbidden error if no valid session was set.
//
// Depending on your configuration this endpoint might return a 403 error if the session has a lower Authenticator
// Assurance Level (AAL) than is possible for the identity. This can happen if the identity has password + webauthn
// credentials (which would result in AAL2) but the session has only AAL1. If this error occurs, ask the user
// to sign in with the second factor (happens automatically for server-side browser flows) or change the configuration.
//
// If this endpoint is called via an AJAX request, the response contains the flow without a redirect. In the
// case of an error, the `error.id` of the JSON response body can be one of:
//
// - `security_csrf_violation`: Unable to fetch the flow because a CSRF violation occurred.
// - `session_inactive`: No Ory Session was found - sign in a user first.
// - `security_identity_mismatch`: The requested `?return_to` address is not allowed to be used. Adjust this in the configuration!
//
// This endpoint is NOT INTENDED for clients that do not have a browser (Chrome, Firefox, ...) as cookies are needed.
//
// More information can be found at [Ory Kratos User Settings & Profile Management Documentation](../self-service/flows/user-settings).
//
//	Schemes: http, https
//
//	Responses:
//	  200: settingsFlow
//	  303: emptyResponse
//	  400: errorGeneric
//	  401: errorGeneric
//	  403: errorGeneric
//	  default: errorGeneric
func (h *Handler) createBrowserSettingsFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	s, err := h.d.SessionManager().FetchFromRequest(r.Context(), r)
	if err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	if err := h.d.SessionManager().DoesSessionSatisfy(r, s, h.d.Config().SelfServiceSettingsRequiredAAL(r.Context())); err != nil {
		h.d.SettingsFlowErrorHandler().WriteFlowError(w, r, node.DefaultGroup, nil, nil, err)
		return
	}

	f, err := h.NewFlow(w, r, s.Identity, flow.TypeBrowser)
	if err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	redirTo := f.AppendTo(h.d.Config().SelfServiceFlowSettingsUI(r.Context())).String()
	x.AcceptToRedirectOrJSON(w, r, h.d.Writer(), f, redirTo)
}

// Get Settings Flow
//
// nolint:deadcode,unused
// swagger:parameters getSettingsFlow
type getSettingsFlow struct {
	// ID is the Settings Flow ID
	//
	// The value for this parameter comes from `flow` URL Query parameter sent to your
	// application (e.g. `/settings?flow=abcde`).
	//
	// required: true
	// in: query
	ID string `json:"id"`

	// The Session Token
	//
	// When using the SDK in an app without a browser, please include the
	// session token here.
	//
	// in: header
	SessionToken string `json:"X-Session-Token"`

	// HTTP Cookies
	//
	// When using the SDK in a browser app, on the server side you must include the HTTP Cookie Header
	// sent by the client to your server here. This ensures that CSRF and session cookies are respected.
	//
	// in: header
	// name: Cookie
	Cookies string `json:"Cookie"`
}

// swagger:route GET /self-service/settings/flows frontend getSettingsFlow
//
// # Get Settings Flow
//
// When accessing this endpoint through Ory Kratos' Public API you must ensure that either the Ory Kratos Session Cookie
// or the Ory Kratos Session Token are set.
//
// Depending on your configuration this endpoint might return a 403 error if the session has a lower Authenticator
// Assurance Level (AAL) than is possible for the identity. This can happen if the identity has password + webauthn
// credentials (which would result in AAL2) but the session has only AAL1. If this error occurs, ask the user
// to sign in with the second factor or change the configuration.
//
// You can access this endpoint without credentials when using Ory Kratos' Admin API.
//
// If this endpoint is called via an AJAX request, the response contains the flow without a redirect. In the
// case of an error, the `error.id` of the JSON response body can be one of:
//
//   - `security_csrf_violation`: Unable to fetch the flow because a CSRF violation occurred.
//   - `session_inactive`: No Ory Session was found - sign in a user first.
//   - `security_identity_mismatch`: The flow was interrupted with `session_refresh_required` but apparently some other
//     identity logged in instead.
//
// More information can be found at [Ory Kratos User Settings & Profile Management Documentation](../self-service/flows/user-settings).
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Responses:
//	  200: settingsFlow
//	  401: errorGeneric
//	  403: errorGeneric
//	  404: errorGeneric
//	  410: errorGeneric
//	  default: errorGeneric
func (h *Handler) getSettingsFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.fetchFlow(w, r); err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}
}

func (h *Handler) fetchFlow(w http.ResponseWriter, r *http.Request) error {
	rid := x.ParseUUID(r.URL.Query().Get("id"))
	pr, err := h.d.SettingsFlowPersister().GetSettingsFlow(r.Context(), rid)
	if err != nil {
		return err
	}

	sess, err := h.d.SessionManager().FetchFromRequest(r.Context(), r)
	if err != nil {
		return err
	}

	if pr.IdentityID != sess.Identity.ID {
		return errors.WithStack(herodot.ErrForbidden.WithID(text.ErrIDInitiatedBySomeoneElse).WithReasonf("The request was made for another identity and has been blocked for security reasons."))
	}

	if err := h.d.SessionManager().DoesSessionSatisfy(r, sess, h.d.Config().SelfServiceSettingsRequiredAAL(r.Context())); err != nil {
		return err
	}

	if pr.ExpiresAt.Before(time.Now().UTC()) {
		if pr.Type == flow.TypeBrowser {
			redirectURL := flow.GetFlowExpiredRedirectURL(r.Context(), h.d.Config(), RouteInitBrowserFlow, pr.ReturnTo)

			h.d.Writer().WriteError(w, r, errors.WithStack(x.ErrGone.
				WithReason("The settings flow has expired. Redirect the user to the settings flow init endpoint to initialize a new settings flow.").
				WithDetail("redirect_to", redirectURL.String()).
				WithDetail("return_to", pr.ReturnTo)))
			return nil
		}
		h.d.Writer().WriteError(w, r, errors.WithStack(x.ErrGone.
			WithReason("The settings flow has expired. Call the settings flow init API endpoint to initialize a new settings flow.").
			WithDetail("api", urlx.AppendPaths(h.d.Config().SelfPublicURL(r.Context()), RouteInitAPIFlow).String())))
		return nil
	}

	h.d.Writer().Write(w, r, pr)
	return nil
}

// Update Settings Flow Parameters
//
// nolint:deadcode,unused
// swagger:parameters updateSettingsFlow
type updateSettingsFlow struct {
	// The Settings Flow ID
	//
	// The value for this parameter comes from `flow` URL Query parameter sent to your
	// application (e.g. `/settings?flow=abcde`).
	//
	// required: true
	// in: query
	Flow string `json:"flow"`

	// in: body
	// required: true
	Body updateSettingsFlowBody

	// The Session Token of the Identity performing the settings flow.
	//
	// in: header
	SessionToken string `json:"X-Session-Token"`

	// HTTP Cookies
	//
	// When using the SDK in a browser app, on the server side you must include the HTTP Cookie Header
	// sent by the client to your server here. This ensures that CSRF and session cookies are respected.
	//
	// in: header
	// name: Cookie
	Cookies string `json:"Cookie"`
}

// Update Settings Flow Request Body
//
// swagger:model updateSettingsFlowBody
// nolint:deadcode,unused
type updateSettingsFlowBody struct{}

// swagger:route POST /self-service/settings frontend updateSettingsFlow
//
// # Complete Settings Flow
//
// Use this endpoint to complete a settings flow by sending an identity's updated password. This endpoint
// behaves differently for API and browser flows.
//
// API-initiated flows expect `application/json` to be sent in the body and respond with
//   - HTTP 200 and an application/json body with the session token on success;
//   - HTTP 303 redirect to a fresh settings flow if the original flow expired with the appropriate error messages set;
//   - HTTP 400 on form validation errors.
//   - HTTP 401 when the endpoint is called without a valid session token.
//   - HTTP 403 when `selfservice.flows.settings.privileged_session_max_age` was reached or the session's AAL is too low.
//     Implies that the user needs to re-authenticate.
//
// Browser flows without HTTP Header `Accept` or with `Accept: text/*` respond with
//   - a HTTP 303 redirect to the post/after settings URL or the `return_to` value if it was set and if the flow succeeded;
//   - a HTTP 303 redirect to the Settings UI URL with the flow ID containing the validation errors otherwise.
//   - a HTTP 303 redirect to the login endpoint when `selfservice.flows.settings.privileged_session_max_age` was reached or the session's AAL is too low.
//
// Browser flows with HTTP Header `Accept: application/json` respond with
//   - HTTP 200 and a application/json body with the signed in identity and a `Set-Cookie` header on success;
//   - HTTP 303 redirect to a fresh login flow if the original flow expired with the appropriate error messages set;
//   - HTTP 401 when the endpoint is called without a valid session cookie.
//   - HTTP 403 when the page is accessed without a session cookie or the session's AAL is too low.
//   - HTTP 400 on form validation errors.
//
// Depending on your configuration this endpoint might return a 403 error if the session has a lower Authenticator
// Assurance Level (AAL) than is possible for the identity. This can happen if the identity has password + webauthn
// credentials (which would result in AAL2) but the session has only AAL1. If this error occurs, ask the user
// to sign in with the second factor (happens automatically for server-side browser flows) or change the configuration.
//
// If this endpoint is called with a `Accept: application/json` HTTP header, the response contains the flow without a redirect. In the
// case of an error, the `error.id` of the JSON response body can be one of:
//
//   - `session_refresh_required`: The identity requested to change something that needs a privileged session. Redirect
//     the identity to the login init endpoint with query parameters `?refresh=true&return_to=<the-current-browser-url>`,
//     or initiate a refresh login flow otherwise.
//   - `security_csrf_violation`: Unable to fetch the flow because a CSRF violation occurred.
//   - `session_inactive`: No Ory Session was found - sign in a user first.
//   - `security_identity_mismatch`: The flow was interrupted with `session_refresh_required` but apparently some other
//     identity logged in instead.
//   - `security_identity_mismatch`: The requested `?return_to` address is not allowed to be used. Adjust this in the configuration!
//   - `browser_location_change_required`: Usually sent when an AJAX request indicates that the browser needs to open a specific URL.
//     Most likely used in Social Sign In flows.
//
// More information can be found at [Ory Kratos User Settings & Profile Management Documentation](../self-service/flows/user-settings).
//
//	Consumes:
//	- application/json
//	- application/x-www-form-urlencoded
//
//	Produces:
//	- application/json
//
//	Security:
//	  sessionToken:
//
//	Schemes: http, https
//
//	Responses:
//	  200: settingsFlow
//	  303: emptyResponse
//	  400: settingsFlow
//	  401: errorGeneric
//	  403: errorGeneric
//	  410: errorGeneric
//	  422: errorBrowserLocationChangeRequired
//	  default: errorGeneric
func (h *Handler) updateSettingsFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	rid, err := GetFlowID(r)
	if err != nil {
		h.d.SettingsFlowErrorHandler().WriteFlowError(w, r, node.DefaultGroup, nil, nil, err)
		return
	}

	f, err := h.d.SettingsFlowPersister().GetSettingsFlow(r.Context(), rid)
	if errors.Is(err, sqlcon.ErrNoRows) {
		h.d.SettingsFlowErrorHandler().WriteFlowError(w, r, node.DefaultGroup, nil, nil, errors.WithStack(herodot.ErrNotFound.WithReasonf("The settings request could not be found. Please restart the flow.")))
		return
	} else if err != nil {
		h.d.SettingsFlowErrorHandler().WriteFlowError(w, r, node.DefaultGroup, nil, nil, err)
		return
	}

	ss, err := h.d.SessionManager().FetchFromRequest(r.Context(), r)
	if err != nil {
		h.d.SettingsFlowErrorHandler().WriteFlowError(w, r, node.DefaultGroup, f, nil, err)
		return
	}

	if err := h.d.SessionManager().DoesSessionSatisfy(r, ss, h.d.Config().SelfServiceSettingsRequiredAAL(r.Context())); err != nil {
		h.d.SettingsFlowErrorHandler().WriteFlowError(w, r, node.DefaultGroup, f, nil, err)
		return
	}

	if err := f.Valid(ss); err != nil {
		h.d.SettingsFlowErrorHandler().WriteFlowError(w, r, node.DefaultGroup, f, ss.Identity, err)
		return
	}

	var s string
	var updateContext *UpdateContext
	for _, strat := range h.d.AllSettingsStrategies() {
		uc, err := strat.Settings(w, r, f, ss)
		if errors.Is(err, flow.ErrStrategyNotResponsible) {
			continue
		} else if errors.Is(err, flow.ErrCompletedByStrategy) {
			return
		} else if err != nil {
			h.d.SettingsFlowErrorHandler().WriteFlowError(w, r, strat.NodeGroup(), f, ss.Identity, err)
			return
		}

		s = strat.SettingsStrategyID()
		updateContext = uc
		break
	}

	if updateContext == nil {
		h.d.SettingsFlowErrorHandler().WriteFlowError(w, r, node.DefaultGroup, f, ss.Identity, errors.WithStack(schema.NewNoSettingsStrategyResponsible()))
		return
	}

	i, err := updateContext.GetIdentityToUpdate()
	if err != nil {
		// An identity to update must always be present.
		h.d.SettingsFlowErrorHandler().WriteFlowError(w, r, node.DefaultGroup, f, ss.Identity, err)
		return
	}

	if err := h.d.SettingsHookExecutor().PostSettingsHook(w, r, s, updateContext, i); err != nil {
		h.d.SettingsFlowErrorHandler().WriteFlowError(w, r, node.DefaultGroup, f, ss.Identity, err)
		return
	}
}
