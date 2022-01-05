package login

import (
	"net/http"
	"time"

	"github.com/gofrs/uuid"

	"github.com/ory/herodot"
	"github.com/ory/kratos/text"
	"github.com/ory/x/stringsx"

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

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(RouteInitBrowserFlow, x.RedirectToPublicRoute(h.d))
	admin.GET(RouteInitAPIFlow, x.RedirectToPublicRoute(h.d))
	admin.GET(RouteGetFlow, x.RedirectToPublicRoute(h.d))

	admin.POST(RouteSubmitFlow, x.RedirectToPublicRoute(h.d))
	admin.GET(RouteSubmitFlow, x.RedirectToPublicRoute(h.d))
}

func (h *Handler) NewLoginFlow(w http.ResponseWriter, r *http.Request, ft flow.Type) (*Flow, error) {
	conf := h.d.Config(r.Context())
	f, err := NewFlow(conf, conf.SelfServiceFlowLoginRequestLifespan(), h.d.GenerateCSRFToken(r), r, ft)
	if err != nil {
		return nil, err
	}

	if f.RequestedAAL == "" {
		f.RequestedAAL = identity.AuthenticatorAssuranceLevel1
	}

	switch cs := stringsx.SwitchExact(string(f.RequestedAAL)); {
	case cs.AddCase(string(identity.AuthenticatorAssuranceLevel1)):
		f.RequestedAAL = identity.AuthenticatorAssuranceLevel1
	case cs.AddCase(string(identity.AuthenticatorAssuranceLevel2)):
		f.RequestedAAL = identity.AuthenticatorAssuranceLevel2
	default:
		return nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Unable to parse AuthenticationMethod Assurance Level (AAL): %s", cs.ToUnknownCaseErr()))
	}

	// We assume an error means the user has no session
	sess, err := h.d.SessionManager().FetchFromRequest(r.Context(), r)
	if e := new(session.ErrNoActiveSessionFound); errors.As(err, &e) {
		// No session exists yet

		// We can not request an AAL > 1 because we must first verify the first factor.
		if f.RequestedAAL > identity.AuthenticatorAssuranceLevel1 {
			return nil, errors.WithStack(ErrSessionRequiredForHigherAAL)
		}

		goto preLoginHook
	} else if err != nil {
		// Some other error happened - return that one.
		return nil, err
	} else {
		// A session exists already
		if f.Refresh {
			// We are refreshing so let's continue
			goto preLoginHook
		}

		// We are not refreshing - so are we requesting MFA?

		// If level is 1 we are not requesting AAL -> we are logged in already.
		if f.RequestedAAL == identity.AuthenticatorAssuranceLevel1 {
			return nil, errors.WithStack(ErrAlreadyLoggedIn)
		}

		// We are requesting an assurance level which the session already has. So we are not upgrading the session
		// in which case we want to return an error.
		if f.RequestedAAL <= sess.AuthenticatorAssuranceLevel {
			return nil, errors.WithStack(ErrAlreadyLoggedIn)
		}

		// Looks like we are requesting an AAL which is higher than what the session has.
		goto preLoginHook
	}

preLoginHook:
	if f.Refresh {
		f.UI.Messages.Set(text.NewInfoLoginReAuth())
	}

	if sess != nil && f.RequestedAAL > sess.AuthenticatorAssuranceLevel && f.RequestedAAL > identity.AuthenticatorAssuranceLevel1 {
		f.UI.Messages.Add(text.NewInfoLoginMFA())
	}

	for _, s := range h.d.LoginStrategies(r.Context()) {
		if err := s.PopulateLoginMethod(r, f.RequestedAAL, f); err != nil {
			return nil, err
		}
	}

	if err := sortNodes(f.UI.Nodes); err != nil {
		return nil, err
	}

	if f.Type == flow.TypeBrowser {
		f.UI.SetCSRF(h.d.GenerateCSRFToken(r))
	}

	if err := h.d.LoginHookExecutor().PreLoginHook(w, r, f); err != nil {
		return nil, err
	}

	if err := h.d.LoginFlowPersister().CreateLoginFlow(r.Context(), f); err != nil {
		return nil, err
	}

	return f, nil
}

func (h *Handler) FromOldFlow(w http.ResponseWriter, r *http.Request, of Flow) (*Flow, error) {
	nf, err := h.NewLoginFlow(w, r, of.Type)
	if err != nil {
		return nil, err
	}

	nf.RequestURL = of.RequestURL
	return nf, nil
}

// nolint:deadcode,unused
// swagger:parameters initializeSelfServiceLoginFlowWithoutBrowser
type initializeSelfServiceLoginFlowWithoutBrowser struct {
	// Refresh a login session
	//
	// If set to true, this will refresh an existing login session by
	// asking the user to sign in again. This will reset the
	// authenticated_at time of the session.
	//
	// in: query
	Refresh bool `json:"refresh"`

	// Request a Specific AuthenticationMethod Assurance Level
	//
	// Use this parameter to upgrade an existing session's authenticator assurance level (AAL). This
	// allows you to ask for multi-factor authentication. When an identity sign in using e.g. username+password,
	// the AAL is 1. If you wish to "upgrade" the session's security by asking the user to perform TOTP / WebAuth/ ...
	// you would set this to "aal2".
	//
	// in: query
	RequestAAL identity.AuthenticatorAssuranceLevel `json:"aal"`

	// The Session Token of the Identity performing the settings flow.
	//
	// in: header
	SessionToken string `json:"X-Session-Token"`
}

// swagger:route GET /self-service/login/api v0alpha2 initializeSelfServiceLoginFlowWithoutBrowser
//
// Initialize Login Flow for APIs, Services, Apps, ...
//
// This endpoint initiates a login flow for API clients that do not use a browser, such as mobile devices, smart TVs, and so on.
//
// If a valid provided session cookie or session token is provided, a 400 Bad Request error
// will be returned unless the URL query parameter `?refresh=true` is set.
//
// To fetch an existing login flow call `/self-service/login/flows?flow=<flow_id>`.
//
// You MUST NOT use this endpoint in client-side (Single Page Apps, ReactJS, AngularJS) nor server-side (Java Server
// Pages, NodeJS, PHP, Golang, ...) browser applications. Using this endpoint in these applications will make
// you vulnerable to a variety of CSRF attacks, including CSRF login attacks.
//
// In the case of an error, the `error.id` of the JSON response body can be one of:
//
// - `session_already_available`: The user is already signed in.
// - `session_aal1_required`: Multi-factor auth (e.g. 2fa) was requested but the user has no session yet.
// - `security_csrf_violation`: Unable to fetch the flow because a CSRF violation occurred.
//
// This endpoint MUST ONLY be used in scenarios such as native mobile apps (React Native, Objective C, Swift, Java, ...).
//
// More information can be found at [Ory Kratos User Login and User Registration Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-login-user-registration).
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: selfServiceLoginFlow
//       400: jsonError
//       500: jsonError
func (h *Handler) initAPIFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	f, err := h.NewLoginFlow(w, r, flow.TypeAPI)
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	h.d.Writer().Write(w, r, f)
}

// nolint:deadcode,unused
// swagger:parameters initializeSelfServiceLoginFlowForBrowsers
type initializeSelfServiceLoginFlowForBrowsers struct {
	// Refresh a login session
	//
	// If set to true, this will refresh an existing login session by
	// asking the user to sign in again. This will reset the
	// authenticated_at time of the session.
	//
	// in: query
	Refresh bool `json:"refresh"`

	// Request a Specific AuthenticationMethod Assurance Level
	//
	// Use this parameter to upgrade an existing session's authenticator assurance level (AAL). This
	// allows you to ask for multi-factor authentication. When an identity sign in using e.g. username+password,
	// the AAL is 1. If you wish to "upgrade" the session's security by asking the user to perform TOTP / WebAuth/ ...
	// you would set this to "aal2".
	//
	// in: query
	RequestAAL identity.AuthenticatorAssuranceLevel `json:"aal"`

	// The URL to return the browser to after the flow was completed.
	//
	// in: query
	ReturnTo string `json:"return_to"`
}

// swagger:route GET /self-service/login/browser v0alpha2 initializeSelfServiceLoginFlowForBrowsers
//
// Initialize Login Flow for Browsers
//
// This endpoint initializes a browser-based user login flow. This endpoint will set the appropriate
// cookies and anti-CSRF measures required for browser-based flows.
//
// If this endpoint is opened as a link in the browser, it will be redirected to
// `selfservice.flows.login.ui_url` with the flow ID set as the query parameter `?flow=`. If a valid user session
// exists already, the browser will be redirected to `urls.default_redirect_url` unless the query parameter
// `?refresh=true` was set.
//
// If this endpoint is called via an AJAX request, the response contains the flow without a redirect. In the
// case of an error, the `error.id` of the JSON response body can be one of:
//
// - `session_already_available`: The user is already signed in.
// - `session_aal1_required`: Multi-factor auth (e.g. 2fa) was requested but the user has no session yet.
// - `security_csrf_violation`: Unable to fetch the flow because a CSRF violation occurred.
// - `security_identity_mismatch`: The requested `?return_to` address is not allowed to be used. Adjust this in the configuration!
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
//       200: selfServiceLoginFlow
//       302: emptyResponse
//       400: jsonError
//       500: jsonError
func (h *Handler) initBrowserFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	a, err := h.NewLoginFlow(w, r, flow.TypeBrowser)
	if errors.Is(err, ErrAlreadyLoggedIn) {
		returnTo, redirErr := x.SecureRedirectTo(r, h.d.Config(r.Context()).SelfServiceBrowserDefaultReturnTo(),
			x.SecureRedirectAllowSelfServiceURLs(h.d.Config(r.Context()).SelfPublicURL()),
			x.SecureRedirectAllowURLs(h.d.Config(r.Context()).SelfServiceBrowserWhitelistedReturnToDomains()),
		)
		if redirErr != nil {
			h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, redirErr)
			return
		}

		x.AcceptToRedirectOrJSON(w, r, h.d.Writer(), err, returnTo.String())
		return
	} else if err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	x.AcceptToRedirectOrJSON(w, r, h.d.Writer(), a, a.AppendTo(h.d.Config(r.Context()).SelfServiceFlowLoginUI()).String())
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

// swagger:route GET /self-service/login/flows v0alpha2 getSelfServiceLoginFlow
//
// Get Login Flow
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
// This request may fail due to several reasons. The `error.id` can be one of:
//
// - `session_already_available`: The user is already signed in.
// - `self_service_flow_expired`: The flow is expired and you should request a new one.
//
// More information can be found at [Ory Kratos User Login and User Registration Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-login-user-registration).
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: selfServiceLoginFlow
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
			h.d.Writer().WriteError(w, r, errors.WithStack(x.ErrGone.WithID(text.ErrIDSelfServiceFlowExpired).
				WithReason("The login flow has expired. Redirect the user to the login flow init endpoint to initialize a new login flow.").
				WithDetail("redirect_to", urlx.AppendPaths(h.d.Config(r.Context()).SelfPublicURL(), RouteInitBrowserFlow).String())))
			return
		}
		h.d.Writer().WriteError(w, r, errors.WithStack(x.ErrGone.WithID(text.ErrIDSelfServiceFlowExpired).
			WithReason("The login flow has expired. Call the login flow init API endpoint to initialize a new login flow.").
			WithDetail("api", urlx.AppendPaths(h.d.Config(r.Context()).SelfPublicURL(), RouteInitAPIFlow).String())))
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

	// The Session Token of the Identity performing the settings flow.
	//
	// in: header
	SessionToken string `json:"X-Session-Token"`
}

// swagger:model submitSelfServiceLoginFlowBody
// nolint:deadcode,unused
type submitSelfServiceLoginFlowBody struct{}

// swagger:route POST /self-service/login v0alpha2 submitSelfServiceLoginFlow
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
// If this endpoint is called with `Accept: application/json` in the header, the response contains the flow without a redirect. In the
// case of an error, the `error.id` of the JSON response body can be one of:
//
// - `session_already_available`: The user is already signed in.
// - `security_csrf_violation`: Unable to fetch the flow because a CSRF violation occurred.
// - `security_identity_mismatch`: The requested `?return_to` address is not allowed to be used. Adjust this in the configuration!
// - `browser_location_change_required`: Usually sent when an AJAX request indicates that the browser needs to open a specific URL.
//		Most likely used in Social Sign In flows.
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
//       200: successfulSelfServiceLoginWithoutBrowser
//       302: emptyResponse
//       400: selfServiceLoginFlow
//       422: selfServiceBrowserLocationChangeRequiredError
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

	sess, err := h.d.SessionManager().FetchFromRequest(r.Context(), r)
	if err == nil {
		if f.Refresh {
			// If we want to refresh, continue the login
			goto continueLogin
		}

		if f.RequestedAAL > sess.AuthenticatorAssuranceLevel {
			// If we want to upgrade AAL, continue the login
			goto continueLogin
		}

		if x.IsJSONRequest(r) || f.Type == flow.TypeAPI {
			// We are not upgrading AAL, nor are we refreshing. Error!
			h.d.LoginFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, errors.WithStack(ErrAlreadyLoggedIn))
			return
		}

		http.Redirect(w, r, h.d.Config(r.Context()).SelfServiceBrowserDefaultReturnTo().String(), http.StatusSeeOther)
		return
	} else if e := new(session.ErrNoActiveSessionFound); errors.As(err, &e) {
		// Only failure scenario here is if we try to upgrade the session to a higher AAL without actually
		// having a session.
		if f.RequestedAAL > identity.AuthenticatorAssuranceLevel1 {
			h.d.LoginFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, errors.WithStack(ErrSessionRequiredForHigherAAL))
			return
		}

		sess = session.NewInactiveSession()
	} else {
		h.d.LoginFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, err)
		return
	}

continueLogin:
	if err := f.Valid(); err != nil {
		h.d.LoginFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, err)
		return
	}

	var i *identity.Identity
	for _, ss := range h.d.AllLoginStrategies() {
		interim, err := ss.Login(w, r, f, sess)
		if errors.Is(err, flow.ErrStrategyNotResponsible) {
			continue
		} else if errors.Is(err, flow.ErrCompletedByStrategy) {
			return
		} else if err != nil {
			h.d.LoginFlowErrorHandler().WriteFlowError(w, r, f, ss.NodeGroup(), err)
			return
		}

		// What can happen is that we re-authenticate as another user. In this case, we need to use a completely fresh
		// session!
		if sess.IdentityID != uuid.Nil && sess.IdentityID != interim.ID {
			sess = session.NewInactiveSession()
		}

		sess.CompletedLoginFor(ss.ID())
		i = interim
		break
	}

	if i == nil {
		h.d.LoginFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, errors.WithStack(schema.NewNoLoginStrategyResponsible()))
		return
	}

	if err := h.d.LoginHookExecutor().PostLoginHook(w, r, f, i, sess); err != nil {
		if errors.Is(err, ErrAddressNotVerified) {
			h.d.LoginFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, errors.WithStack(schema.NewAddressNotVerifiedError()))
			return
		}

		h.d.LoginFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, err)
		return
	}
}
