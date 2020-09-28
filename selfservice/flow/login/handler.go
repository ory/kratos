package login

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	RouteInitBrowserFlow = "/self-service/login/browser"
	RouteInitAPIFlow     = "/self-service/login/api"

	RouteGetFlow = "/self-service/login/flows"
)

type (
	handlerDependencies interface {
		HookExecutorProvider
		FlowPersistenceProvider
		errorx.ManagementProvider
		StrategyProvider
		session.HandlerProvider
		session.ManagementProvider
		x.WriterProvider
		x.CSRFTokenGeneratorProvider
	}
	HandlerProvider interface {
		LoginHandler() *Handler
	}
	Handler struct {
		d handlerDependencies
		c configuration.Provider
	}
)

func NewHandler(d handlerDependencies, c configuration.Provider) *Handler {
	return &Handler{d: d, c: c}
}

func (h *Handler) RegisterPublicRoutes(public *x.RouterPublic) {
	public.GET(RouteInitBrowserFlow, h.initBrowserFlow)
	public.GET(RouteInitAPIFlow, h.initAPIFlow)
	public.GET(RouteGetFlow, h.fetchFlow)
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(RouteGetFlow, h.fetchFlow)
}

func (h *Handler) NewLoginFlow(w http.ResponseWriter, r *http.Request, flow flow.Type) (*Flow, error) {
	a := NewFlow(h.c.SelfServiceFlowLoginRequestLifespan(), h.d.GenerateCSRFToken(r), r, flow)
	for _, s := range h.d.LoginStrategies() {
		if err := s.PopulateLoginMethod(r, a); err != nil {
			return nil, err
		}
	}

	if err := h.d.LoginHookExecutor().PreLoginHook(w, r, a); err != nil {
		return nil, err
	}

	if err := h.d.LoginFlowPersister().CreateLoginFlow(r.Context(), a); err != nil {
		return nil, err
	}
	return a, nil
}

// nolint:deadcode,unused
// swagger:parameters initializeSelfServiceBrowserLoginFlow initializeSelfServiceLoginViaAPIFlow
type initializeSelfServiceBrowserLoginFlow struct {
	// Refresh a login session
	//
	// If set to true, this will refresh an existing login session by
	// asking the user to sign in again. This will reset the
	// authenticated_at time of the session.
	//
	// in: query
	Refresh bool `json:"refresh"`
}

// swagger:route GET /self-service/login/api public initializeSelfServiceLoginViaAPIFlow
//
// Initialize Login Flow for API clients
//
// This endpoint initiates a login flow for API clients such as mobile devices, smart TVs, and so on.
//
// If a valid provided session cookie or session token is provided, a 400 Bad Request error
// will be returned unless the URL query parameter `?refresh=true` is set.
//
// To fetch an existing login flow call `/self-service/login/flows?flow=<flow_id>`.
//
// :::warning
//
// You MUST NOT use this endpoint in client-side (Single Page Apps, ReactJS, AngularJS) nor server-side (Java Server
// Pages, NodeJS, PHP, Golang, ...) browser applications. Using this endpoint in these applications will make
// you vulnerable to a variety of CSRF attacks, including CSRF login attacks.
//
// This endpoint MUST ONLY be used in scenarios such as native mobile apps (React Native, Objective C, Swift, Java, ...).
//
// :::
//
// More information can be found at [ORY Kratos User Login and User Registration Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-login-user-registration).
//
//     Schemes: http, https
//
//     Responses:
//       200: loginFlow
//       500: genericError
//       400: genericError
func (h *Handler) initAPIFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	a, err := h.NewLoginFlow(w, r, flow.TypeAPI)
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	// we assume an error means the user has no session
	if _, err := h.d.SessionManager().FetchFromRequest(r.Context(), r); err != nil {
		h.d.Writer().Write(w, r, a)
		return
	}

	if a.Forced {
		if err := h.d.LoginFlowPersister().ForceLoginFlow(r.Context(), a.ID); err != nil {
			h.d.Writer().WriteError(w, r, err)
			return
		}
		h.d.Writer().Write(w, r, a)
		return
	}

	h.d.Writer().WriteError(w, r, errors.WithStack(ErrAlreadyLoggedIn))
}

// swagger:route GET /self-service/login/browser public initializeSelfServiceLoginViaBrowserFlow
//
// Initialize Login Flow for browsers
//
// This endpoint initializes a browser-based user login flow. Once initialized, the browser will be redirected to
// `selfservice.flows.login.ui_url` with the flow ID set as the query parameter `?flow=`. If a valid user session
// exists already, the browser will be redirected to `urls.default_redirect_url` unless the query parameter
// `?refresh=true` was set.
//
// This endpoint is NOT INTENDED for API clients and only works with browsers (Chrome, Firefox, ...).
//
// More information can be found at [ORY Kratos User Login and User Registration Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-login-user-registration).
//
//     Schemes: http, https
//
//     Responses:
//       302: emptyResponse
//       500: genericError
func (h *Handler) initBrowserFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	a, err := h.NewLoginFlow(w, r, flow.TypeBrowser)

	if err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	// we assume an error means the user has no session
	if _, err := h.d.SessionManager().FetchFromRequest(r.Context(), r); err != nil {
		http.Redirect(w, r, a.AppendTo(h.c.SelfServiceFlowLoginUI()).String(), http.StatusFound)
		return
	}

	if a.Forced {
		if err := h.d.LoginFlowPersister().ForceLoginFlow(r.Context(), a.ID); err != nil {
			h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
			return
		}
		http.Redirect(w, r, a.AppendTo(h.c.SelfServiceFlowLoginUI()).String(), http.StatusFound)
		return
	}

	returnTo, err := x.SecureRedirectTo(r, h.c.SelfServiceBrowserDefaultReturnTo(),
		x.SecureRedirectAllowSelfServiceURLs(h.c.SelfPublicURL()),
		x.SecureRedirectAllowURLs(h.c.SelfServiceBrowserWhitelistedReturnToDomains()),
	)
	if err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	http.Redirect(w, r, returnTo.String(), http.StatusFound)
}

// nolint:deadcode,unused
// swagger:parameters getSelfServiceLoginFlow
type getSelfServiceLoginFlow struct {
	// The Login Flow ID
	//
	// The value for this parameter comes from `flow` URL Query parameter sent to your
	// application (e.g. `/login?flow=abcde`).
	//
	// required: true
	// in: query
	ID string `json:"id"`
}

// swagger:route GET /self-service/login/flows public admin getSelfServiceLoginFlow
//
// Get Login Flow
//
// This endpoint returns a login flow's context with, for example, error details and other information.
//
// More information can be found at [ORY Kratos User Login and User Registration Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-login-user-registration).
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: loginFlow
//       403: genericError
//       404: genericError
//       410: genericError
//       500: genericError
func (h *Handler) fetchFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ar, err := h.d.LoginFlowPersister().GetLoginFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("id")))
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	if ar.ExpiresAt.Before(time.Now()) {
		if ar.Type == flow.TypeBrowser {
			h.d.Writer().WriteError(w, r, errors.WithStack(x.ErrGone.
				WithReason("The login flow has expired. Redirect the user to the login flow init endpoint to initialize a new login flow.").
				WithDetail("redirect_to", urlx.AppendPaths(h.c.SelfPublicURL(), RouteInitBrowserFlow).String())))
			return
		}
		h.d.Writer().WriteError(w, r, errors.WithStack(x.ErrGone.
			WithReason("The login flow has expired. Call the login flow init API endpoint to initialize a new login flow.").
			WithDetail("api", urlx.AppendPaths(h.c.SelfPublicURL(), RouteInitAPIFlow).String())))
		return
	}

	h.d.Writer().Write(w, r, ar)
}
