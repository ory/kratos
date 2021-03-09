package login

import (
	"github.com/ory/herodot"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/x/decoderx"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	RouteInitBrowserFlow = "/self-service/login/browser"
	RouteInitAPIFlow     = "/self-service/login/api"

	RouteGetFlow = "/self-service/login/flows"

	RouteSubmitFlow = "/self-service/login"
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
		x.CSRFProvider
		config.Provider
		ErrorHandlerProvider
	}
	HandlerProvider interface {
		LoginHandler() *Handler
	}
	Handler struct {
		d  handlerDependencies
		hd *decoderx.HTTP
	}
)

func NewHandler(d handlerDependencies) *Handler {
	return &Handler{d: d, hd: decoderx.NewHTTP()}
}

func (h *Handler) RegisterPublicRoutes(public *x.RouterPublic) {
	h.d.CSRFHandler().IgnorePath(RouteInitAPIFlow)
	h.d.CSRFHandler().IgnorePath(RouteSubmitFlow)

	public.GET(RouteInitBrowserFlow, h.initBrowserFlow)
	public.GET(RouteInitAPIFlow, h.initAPIFlow)
	public.GET(RouteGetFlow, h.fetchFlow)

	public.POST(RouteSubmitFlow, h.submitFlow)
	public.GET(RouteSubmitFlow, h.submitFlow)
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(RouteGetFlow, h.fetchFlow)
}

func (h *Handler) NewLoginFlow(w http.ResponseWriter, r *http.Request, flow flow.Type) (*Flow, error) {
	conf := h.d.Config(r.Context())
	a := NewFlow(conf, conf.SelfServiceFlowLoginRequestLifespan(), h.d.GenerateCSRFToken(r), r, flow)
	for _, s := range h.d.LoginStrategies(r.Context()) {
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
		http.Redirect(w, r, a.AppendTo(h.d.Config(r.Context()).SelfServiceFlowLoginUI()).String(), http.StatusFound)
		return
	}

	if a.Forced {
		if err := h.d.LoginFlowPersister().ForceLoginFlow(r.Context(), a.ID); err != nil {
			h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
			return
		}
		http.Redirect(w, r, a.AppendTo(h.d.Config(r.Context()).SelfServiceFlowLoginUI()).String(), http.StatusFound)
		return
	}

	returnTo, err := x.SecureRedirectTo(r, h.d.Config(r.Context()).SelfServiceBrowserDefaultReturnTo(),
		x.SecureRedirectAllowSelfServiceURLs(h.d.Config(r.Context()).SelfPublicURL(r)),
		x.SecureRedirectAllowURLs(h.d.Config(r.Context()).SelfServiceBrowserWhitelistedReturnToDomains()),
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
				WithDetail("redirect_to", urlx.AppendPaths(h.d.Config(r.Context()).SelfPublicURL(r), RouteInitBrowserFlow).String())))
			return
		}
		h.d.Writer().WriteError(w, r, errors.WithStack(x.ErrGone.
			WithReason("The login flow has expired. Call the login flow init API endpoint to initialize a new login flow.").
			WithDetail("api", urlx.AppendPaths(h.d.Config(r.Context()).SelfPublicURL(r), RouteInitAPIFlow).String())))
		return
	}

	h.d.Writer().Write(w, r, ar)
}

// nolint:deadcode,unused
// swagger:parameters submitSelfServiceLoginFlow
type submitSelfServiceLoginFlow struct {
	// The Login Flow ID
	//
	// The value for this parameter comes from `flow` URL Query parameter sent to your
	// application (e.g. `/login?flow=abcde`).
	//
	// required: true
	// in: query
	ID string `json:"id"`
}

// swagger:route POST /self-service/login/flows public submitSelfServiceLoginFlow
//
// Submit a Login Flow
//
// Use this endpoint to complete a login flow. This endpoint
// behaves differently for API and browser flows.
//
// API flows expect `application/json` to be sent in the body and responds with
//   - HTTP 200 and a application/json body with the session token on success;
//   - HTTP 302 redirect to a fresh login flow if the original flow expired with the appropriate error messages set;
//   - HTTP 400 on form validation errors.
//
// Browser flows expect `application/x-www-form-urlencoded` to be sent in the body and responds with
//   - a HTTP 302 redirect to the post/after login URL or the `return_to` value if it was set and if the login succeeded;
//   - a HTTP 302 redirect to the login UI URL with the flow ID containing the validation errors otherwise.
//
// More information can be found at [ORY Kratos User Login and User Registration Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-login-user-registration).
//
//     Schemes: http, https
//
//     Consumes:
//     - application/json
//     - application/x-www-form-urlencoded
//
//     Produces:
//     - application/json
//
//     Header:
//     - Set-Cookie
//
//     Responses:
//       200: loginViaApiResponse
//       302: emptyResponse
//       400: loginFlow
//       500: genericError
func (h *Handler) submitFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	rid := x.ParseUUID(r.URL.Query().Get("flow"))
	if x.IsZeroUUID(rid) {
		h.d.LoginFlowErrorHandler().WriteFlowError(w, r, nil, node.DefaultGroup, errors.WithStack(herodot.ErrBadRequest.WithReasonf("The flow query parameter is missing or invalid.")))
		return
	}

	f, err := h.d.LoginFlowPersister().GetLoginFlow(r.Context(), rid)
	if err != nil {
		h.d.LoginFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, err)
		return
	}

	// TODO first decoding here - we need to CSRF token

	if _, err := h.d.SessionManager().FetchFromRequest(r.Context(), r); err == nil && !f.Forced {
		if f.Type == flow.TypeBrowser {
			http.Redirect(w, r, h.d.Config(r.Context()).SelfServiceBrowserDefaultReturnTo().String(), http.StatusFound)
			return
		}

		h.d.LoginFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, err)
		return
	}

	if err := f.Valid(); err != nil {
		h.d.LoginFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, err)
		return
	}

	var i *identity.Identity
	var s identity.CredentialsType
	for _, ss := range h.d.LoginStrategies() {
		interim, err := ss.Login(w, r, f)
		if errors.Is(err, ErrStrategyNotResponsible) {
			continue
		} else if err != nil {
			h.d.LoginFlowErrorHandler().WriteFlowError(w, r, f, ss.NodeGroup(), err)
			return
		}

		i = interim
		s = ss.ID()
		break
	}

	if i == nil {
		h.d.LoginFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, errors.WithStack(schema.NewNoLoginStrategyResponsible()))
		return
	}

	// TODO Handle n+1 authentication factor

	if err := h.d.LoginHookExecutor().PostLoginHook(w, r, s, f, i); err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}
}
