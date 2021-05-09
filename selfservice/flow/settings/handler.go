package settings

import (
	"net/http"
	"time"

	"github.com/ory/kratos/ui/node"
	"github.com/ory/x/sqlcon"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"github.com/ory/nosurf"
	"github.com/ory/x/urlx"
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

	public.GET(RouteInitBrowserFlow, h.d.SessionHandler().IsAuthenticated(h.initBrowserFlow, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		http.Redirect(w, r, h.d.Config(r.Context()).SelfServiceFlowLoginUI().String(), http.StatusFound)
	}))

	public.GET(RouteInitAPIFlow, h.d.SessionHandler().IsAuthenticated(h.initApiFlow, nil))
	public.GET(RouteGetFlow, h.d.SessionHandler().IsAuthenticated(h.fetchPublicFlow, OnUnauthenticated(h.d)))

	public.POST(RouteSubmitFlow, h.d.SessionHandler().IsAuthenticated(h.submitSettingsFlow, OnUnauthenticated(h.d)))
	public.GET(RouteSubmitFlow, h.d.SessionHandler().IsAuthenticated(h.submitSettingsFlow, OnUnauthenticated(h.d)))
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(RouteGetFlow, h.fetchAdminFlow)
}

func (h *Handler) NewFlow(w http.ResponseWriter, r *http.Request, i *identity.Identity, ft flow.Type) (*Flow, error) {
	f := NewFlow(h.d.Config(r.Context()), h.d.Config(r.Context()).SelfServiceFlowSettingsFlowLifespan(), r, i, ft)
	for _, strategy := range h.d.SettingsStrategies(r.Context()) {
		if err := h.d.ContinuityManager().Abort(r.Context(), w, r, ContinuityKey(strategy.SettingsStrategyID())); err != nil {
			return nil, err
		}

		if err := strategy.PopulateSettingsMethod(r, i, f); err != nil {
			return nil, err
		}
	}

	if err := sortNodes(f.UI.Nodes, h.d.Config(r.Context()).DefaultIdentityTraitsSchemaURL().String()); err != nil {
		return nil, err
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
// You must provide a valid Ory Kratos Session Token for this endpoint to respond with HTTP 200 OK.
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
// More information can be found at [Ory Kratos User Settings & Profile Management Documentation](../self-service/flows/user-settings).
//
//     Schemes: http, https
//
//     Security:
//       sessionToken:
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

// swagger:route GET /self-service/settings/browser public initializeSelfServiceSettingsViaBrowserFlow
//
// Initialize Settings Flow for Browsers
//
// This endpoint initializes a browser-based user settings flow. Once initialized, the browser will be redirected to
// `selfservice.flows.settings.ui_url` with the flow ID set as the query parameter `?flow=`. If no valid
// Ory Kratos Session Cookie is included in the request, a login flow will be initialized.
//
// :::note
//
// This endpoint is NOT INTENDED for API clients and only works with browsers (Chrome, Firefox, ...).
//
// :::
//
// More information can be found at [Ory Kratos User Settings & Profile Management Documentation](../self-service/flows/user-settings).
//
//     Schemes: http, https
//
//     Security:
//       sessionToken:
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

	http.Redirect(w, r, f.AppendTo(h.d.Config(r.Context()).SelfServiceFlowSettingsUI()).String(), http.StatusFound)
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

// swagger:route GET /self-service/settings/flows public admin getSelfServiceSettingsFlow
//
// Get Settings Flow
//
// When accessing this endpoint through Ory Kratos' Public API you must ensure that either the Ory Kratos Session Cookie
// or the Ory Kratos Session Token are set. The public endpoint does not return 404 status codes
// but instead 403 or 500 to improve data privacy.
//
// You can access this endpoint without credentials when using Ory Kratos' Admin API.
//
// More information can be found at [Ory Kratos User Settings & Profile Management Documentation](../self-service/flows/user-settings).
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Security:
//       sessionToken:
//
//     Responses:
//       200: settingsFlow
//       403: genericError
//       404: genericError
//       410: genericError
//       500: genericError
func (h *Handler) fetchPublicFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
				WithDetail("redirect_to", urlx.AppendPaths(h.d.Config(r.Context()).SelfPublicURL(r), RouteInitBrowserFlow).String())))
			return nil
		}
		h.d.Writer().WriteError(w, r, errors.WithStack(x.ErrGone.
			WithReason("The settings flow has expired. Call the settings flow init API endpoint to initialize a new settings flow.").
			WithDetail("api", urlx.AppendPaths(h.d.Config(r.Context()).SelfPublicURL(r), RouteInitAPIFlow).String())))
		return nil
	}

	h.d.Writer().Write(w, r, pr)
	return nil
}

// nolint:deadcode,unused
// swagger:parameters submitSelfServiceSettingsFlow
type submitSelfServiceSettingsFlow struct {
	// The Settings Flow ID
	//
	// The value for this parameter comes from `flow` URL Query parameter sent to your
	// application (e.g. `/settings?flow=abcde`).
	//
	// required: true
	// in: query
	Flow string `json:"flow"`

	// in: body
	Body submitSelfServiceSettingsFlowBody
}

// swagger:model submitSelfServiceSettingsFlow
// nolint:deadcode,unused
type submitSelfServiceSettingsFlowBody struct{}

// swagger:route POST /self-service/settings public submitSelfServiceSettingsFlow
//
// Complete Settings Flow
//
// Use this endpoint to complete a settings flow by sending an identity's updated password. This endpoint
// behaves differently for API and browser flows.
//
// API-initiated flows expect `application/json` to be sent in the body and respond with
//   - HTTP 200 and an application/json body with the session token on success;
//   - HTTP 302 redirect to a fresh settings flow if the original flow expired with the appropriate error messages set;
//   - HTTP 400 on form validation errors.
//   - HTTP 401 when the endpoint is called without a valid session token.
//   - HTTP 403 when `selfservice.flows.settings.privileged_session_max_age` was reached.
//     Implies that the user needs to re-authenticate.
//
// Browser flows expect `application/x-www-form-urlencoded` to be sent in the body and responds with
//   - a HTTP 302 redirect to the post/after settings URL or the `return_to` value if it was set and if the flow succeeded;
//   - a HTTP 302 redirect to the Settings UI URL with the flow ID containing the validation errors otherwise.
//   - a HTTP 302 redirect to the login endpoint when `selfservice.flows.settings.privileged_session_max_age` was reached.
//
// More information can be found at [Ory Kratos User Settings & Profile Management Documentation](../self-service/flows/user-settings).
//
//     Consumes:
//     - application/json
//     - application/x-www-form-urlencoded
//
//     Produces:
//     - application/json
//
//     Security:
//       sessionToken:
//
//     Schemes: http, https
//
//     Responses:
//       200: settingsViaApiResponse
//       302: emptyResponse
//       400: settingsFlow
//       401: genericError
//       403: genericError
//       500: genericError
func (h *Handler) submitSettingsFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
		if f.Type == flow.TypeBrowser {
			http.Redirect(w, r, h.d.Config(r.Context()).SelfServiceFlowLoginUI().String(), http.StatusFound)
			return
		}

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

	// if the submission was from a recovery session, we have to end the recovery session and
	// reissue a new session which isn't a recovery session anymore
	if ss.Recovery {
		if err := h.d.SessionManager().PurgeFromRequest(r.Context(), w, r); err != nil {
			h.d.SettingsFlowErrorHandler().WriteFlowError(w, r, node.DefaultGroup, f, ss.Identity, err)
			return
		}
		newSession := session.NewActiveSession(ss.Identity, h.d.Config(r.Context()), time.Now())

		// this could sit nicely behind a flag to leave the user signed out after the password change
		if err := h.d.SessionManager().CreateAndIssueCookie(r.Context(), w, r, newSession); err != nil {
			h.d.SettingsFlowErrorHandler().WriteFlowError(w, r, node.DefaultGroup, f, ss.Identity, err)
			return
		}
	}

	if updateContext == nil {
		c := &UpdateContext{Session: ss, Flow: f}
		h.d.SettingsFlowErrorHandler().WriteFlowError(w, r, node.DefaultGroup, f, c.GetIdentityToUpdate(), errors.WithStack(schema.NewNoSettingsStrategyResponsible()))
		return
	}

	if err := h.d.SettingsHookExecutor().PostSettingsHook(w, r, s, updateContext, updateContext.GetIdentityToUpdate()); err != nil {
		h.d.SettingsFlowErrorHandler().WriteFlowError(w, r, node.DefaultGroup, f, ss.Identity, err)
		return
	}
}
