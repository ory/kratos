package login

import (
	"net/http"
	"time"

	"github.com/ory/nosurf"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/x/decoderx"

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

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {}

func (h *Handler) NewLoginFlow(w http.ResponseWriter, r *http.Request, flow flow.Type) (*Flow, error) {
	conf := h.d.Config(r.Context())
	f := NewFlow(conf, conf.SelfServiceFlowLoginRequestLifespan(), h.d.GenerateCSRFToken(r), r, flow)
	for _, s := range h.d.LoginStrategies(r.Context()) {
		if err := s.PopulateLoginMethod(r, f); err != nil {
			return nil, err
		}
	}

	if err := sortNodes(f.UI.Nodes); err != nil {
		return nil, err
	}

	if err := h.d.LoginHookExecutor().PreLoginHook(w, r, f); err != nil {
		return nil, err
	}

	if err := h.d.LoginFlowPersister().CreateLoginFlow(r.Context(), f); err != nil {
		return nil, err
	}
	return f, nil
}

// nolint:deadcode,unused
// swagger:parameters initializeSelfServiceLoginForBrowsers initializeSelfServiceLoginWithoutBrowser
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

// swagger:route GET /self-service/login/api public initializeSelfServiceLoginWithoutBrowser
//
// Initialize Login Flow for APIs, Services, Apps, ...
//
// This endpoint initiates a login flow for API clients that do not use a browser, such as mobile devices, smart TVs, and so on.
//
// :::info
//
// This endpoint is EXPERIMENTAL and subject to potential breaking changes in the future.
//
// :::
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
// More information can be found at [Ory Kratos User Login and User Registration Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-login-user-registration).
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: loginFlow
//       400: jsonError
//       500: jsonError
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

// swagger:route GET /self-service/login/browser public initializeSelfServiceLoginForBrowsers
//
// Initialize Login Flow for Browsers
//
// This endpoint initializes a browser-based user login flow. This endpoint will set the appropriate
// cookies and anti-CSRF measures required for browser-based flows.
//
// :::info
//
// This endpoint is EXPERIMENTAL and subject to potential breaking changes in the future.
//
// :::
//
// If this endpoint is opened as a link in the browser, it will be redirected to
// `selfservice.flows.login.ui_url` with the flow ID set as the query parameter `?flow=`. If a valid user session
// exists already, the browser will be redirected to `urls.default_redirect_url` unless the query parameter
// `?refresh=true` was set.
//
// If this endpoint is called via an AJAX request, the response contains the login flow without a redirect.
//
// This endpoint is NOT INTENDED for clients that do not have a browser (Chrome, Firefox, ...) as cookies are needed.
//
// More information can be found at [Ory Kratos User Login and User Registration Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-login-user-registration).
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: loginFlow
//       302: emptyResponse
//       500: jsonError
func (h *Handler) initBrowserFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	a, err := h.NewLoginFlow(w, r, flow.TypeBrowser)
	if err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	// we assume an error means the user has no session
	if _, err := h.d.SessionManager().FetchFromRequest(r.Context(), r); err != nil {
		x.AcceptToRedirectOrJson(w, r, h.d.Writer(), a, a.AppendTo(h.d.Config(r.Context()).SelfServiceFlowLoginUI()).String())
		return
	}

	if a.Forced {
		if err := h.d.LoginFlowPersister().ForceLoginFlow(r.Context(), a.ID); err != nil {
			h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
			return
		}
		x.AcceptToRedirectOrJson(w, r, h.d.Writer(), a, a.AppendTo(h.d.Config(r.Context()).SelfServiceFlowLoginUI()).String())
		return
	}

	if x.IsJSONRequest(r) {
		h.d.Writer().WriteError(w, r, errors.WithStack(ErrAlreadyLoggedIn))
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

	http.Redirect(w, r, returnTo.String(), http.StatusSeeOther)
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

	// HTTP Cookies
	//
	// When using the SDK on the server side you must include the HTTP Cookie Header
	// originally sent to your HTTP handler here.
	//
	// in: header
	// name: Cookie
	Cookies string `json:"cookie"`
}

// swagger:route GET /self-service/login/flows public getSelfServiceLoginFlow
//
// Get Login Flow
//
// :::info
//
// This endpoint is EXPERIMENTAL and subject to potential breaking changes in the future.
//
// :::
//
// This endpoint returns a login flow's context with, for example, error details and other information.
//
// Browser flows expect the anti-CSRF cookie to be included in the request's HTTP Cookie Header.
// For AJAX requests you must ensure that cookies are included in the request or requests will fail.
//
// If you use the browser-flow for server-side apps, the services need to run on a common top-level-domain
// and you need to forward the incoming HTTP Cookie header to this endpoint:
//
//	```js
//	// pseudo-code example
//	router.get('/login', async function (req, res) {
//	  const flow = await client.getSelfServiceLoginFlow(req.header('cookie'), req.query['flow'])
//
//    res.render('login', flow)
//	})
//	```
//
// More information can be found at [Ory Kratos User Login and User Registration Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-login-user-registration).
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: loginFlow
//       403: jsonError
//       404: jsonError
//       410: jsonError
//       500: jsonError
func (h *Handler) fetchFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ar, err := h.d.LoginFlowPersister().GetLoginFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("id")))
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	// Browser flows must include the CSRF token
	//
	// Resolves: https://github.com/ory/kratos/issues/1282
	if ar.Type == flow.TypeBrowser && !nosurf.VerifyToken(h.d.GenerateCSRFToken(r), ar.CSRFToken) {
		h.d.Writer().WriteError(w, r, x.CSRFErrorReason(r, h.d))
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
	Flow string `json:"flow"`

	// in: body
	Body submitSelfServiceLoginFlowBody
}

// swagger:model submitSelfServiceLoginFlow
// nolint:deadcode,unused
type submitSelfServiceLoginFlowBody struct{}

// swagger:route POST /self-service/login public submitSelfServiceLoginFlow
//
// Submit a Login Flow
//
// :::info
//
// This endpoint is EXPERIMENTAL and subject to potential breaking changes in the future.
//
// :::
//
// Use this endpoint to complete a login flow. This endpoint
// behaves differently for API and browser flows.
//
// API flows expect `application/json` to be sent in the body and responds with
//   - HTTP 200 and a application/json body with the session token on success;
//   - HTTP 302 redirect to a fresh login flow if the original flow expired with the appropriate error messages set;
//   - HTTP 400 on form validation errors.
//
// Browser flows expect a Content-Type of `application/x-www-form-urlencoded` or `application/json` to be sent in the body and respond with
//   - a HTTP 302 redirect to the post/after login URL or the `return_to` value if it was set and if the login succeeded;
//   - a HTTP 302 redirect to the login UI URL with the flow ID containing the validation errors otherwise.
//
// Browser flows with an accept header of `application/json` will not redirect but instead respond with
//   - HTTP 200 and a application/json body with the signed in identity and a `Set-Cookie` header on success;
//   - HTTP 302 redirect to a fresh login flow if the original flow expired with the appropriate error messages set;
//   - HTTP 400 on form validation errors.
//
// More information can be found at [Ory Kratos User Login and User Registration Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-login-user-registration).
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
//       500: jsonError
func (h *Handler) submitFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	rid, err := flow.GetFlowID(r)
	if err != nil {
		h.d.LoginFlowErrorHandler().WriteFlowError(w, r, nil, node.DefaultGroup, err)
		return
	}

	f, err := h.d.LoginFlowPersister().GetLoginFlow(r.Context(), rid)
	if err != nil {
		h.d.LoginFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, err)
		return
	}

	if _, err := h.d.SessionManager().FetchFromRequest(r.Context(), r); err == nil && !f.Forced {
		if f.Type == flow.TypeBrowser {
			http.Redirect(w, r, h.d.Config(r.Context()).SelfServiceBrowserDefaultReturnTo().String(), http.StatusFound)
			return
		}

		h.d.Writer().WriteError(w, r, errors.WithStack(ErrAlreadyLoggedIn))
		return
	}

	if err := f.Valid(); err != nil {
		h.d.LoginFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, err)
		return
	}

	var i *identity.Identity
	for _, ss := range h.d.AllLoginStrategies() {
		interim, err := ss.Login(w, r, f)
		if errors.Is(err, flow.ErrStrategyNotResponsible) {
			continue
		} else if errors.Is(err, flow.ErrCompletedByStrategy) {
			return
		} else if err != nil {
			h.d.LoginFlowErrorHandler().WriteFlowError(w, r, f, ss.NodeGroup(), err)
			return
		}

		i = interim
		break
	}

	if i == nil {
		h.d.LoginFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, errors.WithStack(schema.NewNoLoginStrategyResponsible()))
		return
	}

	// TODO Handle n+1 authentication factor

	if err := h.d.LoginHookExecutor().PostLoginHook(w, r, f, i); err != nil {
		if err == ErrAddressNotVerified {
			h.d.LoginFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, errors.WithStack(schema.NewAddressNotVerifiedError()))
		} else {
			h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		}
		return
	}
}
