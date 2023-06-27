// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gofrs/uuid"

	"github.com/ory/herodot"
	hydraclientgo "github.com/ory/hydra-client-go/v2"
	"github.com/ory/kratos/hydra"
	"github.com/ory/kratos/selfservice/sessiontokenexchange"
	"github.com/ory/kratos/text"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/stringsx"

	"github.com/ory/nosurf"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/x/decoderx"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"github.com/ory/x/urlx"
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
		hydra.Provider
		StrategyProvider
		session.HandlerProvider
		session.ManagementProvider
		x.WriterProvider
		x.CSRFTokenGeneratorProvider
		x.CSRFProvider
		config.Provider
		ErrorHandlerProvider
		sessiontokenexchange.PersistenceProvider
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

	public.GET(RouteInitBrowserFlow, h.createBrowserLoginFlow)
	public.GET(RouteInitAPIFlow, h.createNativeLoginFlow)
	public.GET(RouteGetFlow, h.getLoginFlow)

	public.POST(RouteSubmitFlow, h.updateLoginFlow)
	public.GET(RouteSubmitFlow, h.updateLoginFlow)
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

func WithFormErrorMessage(messages []text.Message) FlowOption {
	return func(f *Flow) {
		for i := range messages {
			f.UI.Messages.Add(&messages[i])
		}
	}
}

func (h *Handler) NewLoginFlow(w http.ResponseWriter, r *http.Request, ft flow.Type, opts ...FlowOption) (*Flow, *session.Session, error) {
	conf := h.d.Config()
	f, err := NewFlow(conf, conf.SelfServiceFlowLoginRequestLifespan(r.Context()), h.d.GenerateCSRFToken(r), r, ft)
	if err != nil {
		return nil, nil, err
	}
	for _, o := range opts {
		o(f)
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
		return nil, nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Unable to parse AuthenticationMethod Assurance Level (AAL): %s", cs.ToUnknownCaseErr()))
	}

	// We assume an error means the user has no session
	sess, err := h.d.SessionManager().FetchFromRequest(r.Context(), r)
	if e := new(session.ErrNoActiveSessionFound); errors.As(err, &e) {
		// No session exists yet
		if ft == flow.TypeAPI && r.URL.Query().Get("return_session_token_exchange_code") == "true" {
			e, err := h.d.SessionTokenExchangePersister().CreateSessionTokenExchanger(r.Context(), f.ID)
			if err != nil {
				return nil, nil, errors.WithStack(herodot.ErrInternalServerError.WithWrap(err))
			}
			f.SessionTokenExchangeCode = e.InitCode
		}

		// We can not request an AAL > 1 because we must first verify the first factor.
		if f.RequestedAAL > identity.AuthenticatorAssuranceLevel1 {
			return nil, nil, errors.WithStack(ErrSessionRequiredForHigherAAL)
		}

		// We are setting refresh to false if no session exists.
		f.Refresh = false

		goto preLoginHook
	} else if err != nil {
		// Some other error happened - return that one.
		return nil, nil, err
	} else {
		// A session exists already
		if f.Refresh {
			// We are refreshing so let's continue
			goto preLoginHook
		}

		// We are not refreshing - so are we requesting MFA?

		// If level is 1 we are not requesting AAL -> we are logged in already.
		if f.RequestedAAL == identity.AuthenticatorAssuranceLevel1 {
			return nil, sess, errors.WithStack(ErrAlreadyLoggedIn)
		}

		// We are requesting an assurance level which the session already has. So we are not upgrading the session
		// in which case we want to return an error.
		if f.RequestedAAL <= sess.AuthenticatorAssuranceLevel {
			return nil, sess, errors.WithStack(ErrAlreadyLoggedIn)
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

	var s Strategy
	for _, s = range h.d.LoginStrategies(r.Context()) {
		if err := s.PopulateLoginMethod(r, f.RequestedAAL, f); err != nil {
			return nil, nil, err
		}
	}

	if err := sortNodes(r.Context(), f.UI.Nodes); err != nil {
		return nil, nil, err
	}

	if f.Type == flow.TypeBrowser {
		f.UI.SetCSRF(h.d.GenerateCSRFToken(r))
	}

	if err := h.d.LoginHookExecutor().PreLoginHook(w, r, f); err != nil {
		h.d.LoginFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, err)
		return f, sess, nil
	}

	if err := h.d.LoginFlowPersister().CreateLoginFlow(r.Context(), f); err != nil {
		return nil, nil, err
	}

	return f, nil, nil
}

func (h *Handler) FromOldFlow(w http.ResponseWriter, r *http.Request, of Flow) (*Flow, error) {
	nf, _, err := h.NewLoginFlow(w, r, of.Type)
	if err != nil {
		return nil, err
	}

	nf.RequestURL = of.RequestURL
	return nf, nil
}

// Create Native Login Flow Parameters
//
// swagger:parameters createNativeLoginFlow
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type createNativeLoginFlow struct {
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

	// EnableSessionTokenExchangeCode requests the login flow to include a code that can be used to retrieve the session token
	// after the login flow has been completed.
	//
	// in: query
	EnableSessionTokenExchangeCode bool `json:"return_session_token_exchange_code"`

	// The URL to return the browser to after the flow was completed.
	//
	// in: query
	ReturnTo string `json:"return_to"`
}

// swagger:route GET /self-service/login/api frontend createNativeLoginFlow
//
// # Create Login Flow for Native Apps
//
// This endpoint initiates a login flow for native apps that do not use a browser, such as mobile devices, smart TVs, and so on.
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
// More information can be found at [Ory Kratos User Login](https://www.ory.sh/docs/kratos/self-service/flows/user-login) and [User Registration Documentation](https://www.ory.sh/docs/kratos/self-service/flows/user-registration).
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Responses:
//	  200: loginFlow
//	  400: errorGeneric
//	  default: errorGeneric
func (h *Handler) createNativeLoginFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	f, _, err := h.NewLoginFlow(w, r, flow.TypeAPI)
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	h.d.Writer().Write(w, r, f)
}

// Initialize Browser Login Flow Parameters
//
// swagger:parameters createBrowserLoginFlow
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type createBrowserLoginFlow struct {
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

	// HTTP Cookies
	//
	// When using the SDK in a browser app, on the server side you must include the HTTP Cookie Header
	// sent by the client to your server here. This ensures that CSRF and session cookies are respected.
	//
	// in: header
	// name: Cookie
	Cookies string `json:"Cookie"`

	// An optional Hydra login challenge. If present, Kratos will cooperate with
	// Ory Hydra to act as an OAuth2 identity provider.
	//
	// The value for this parameter comes from `login_challenge` URL Query parameter sent to your
	// application (e.g. `/login?login_challenge=abcde`).
	//
	// required: false
	// in: query
	HydraLoginChallenge string `json:"login_challenge"`
}

// swagger:route GET /self-service/login/browser frontend createBrowserLoginFlow
//
// # Create Login Flow for Browsers
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
// The optional query parameter login_challenge is set when using Kratos with
// Hydra in an OAuth2 flow. See the oauth2_provider.url configuration
// option.
//
// This endpoint is NOT INTENDED for clients that do not have a browser (Chrome, Firefox, ...) as cookies are needed.
//
// More information can be found at [Ory Kratos User Login](https://www.ory.sh/docs/kratos/self-service/flows/user-login) and [User Registration Documentation](https://www.ory.sh/docs/kratos/self-service/flows/user-registration).
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Responses:
//	  200: loginFlow
//	  303: emptyResponse
//	  400: errorGeneric
//	  default: errorGeneric
func (h *Handler) createBrowserLoginFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var (
		hydraLoginRequest   *hydraclientgo.OAuth2LoginRequest
		hydraLoginChallenge sqlxx.NullString
	)
	if r.URL.Query().Has("login_challenge") {
		var err error
		hydraLoginChallenge, err = hydra.GetLoginChallengeID(h.d.Config(), r)
		if err != nil {
			h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
			return
		}

		hydraLoginRequest, err = h.d.Hydra().GetLoginRequest(r.Context(), string(hydraLoginChallenge))
		if err != nil {
			h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, errors.WithStack(herodot.ErrInternalServerError.WithReason("Failed to retrieve OAuth 2.0 login request.")))
			return
		}

		if !hydraLoginRequest.GetSkip() {
			q := r.URL.Query()
			q.Set("refresh", "true")
			r.URL.RawQuery = q.Encode()
		}

		// on OAuth2 flows, we need to use the RequestURL
		// as the ReturnTo URL.
		// This is because a user might want to switch between
		// different flows, such as login to registration and login to recovery.
		// After completing a complex flow, such as recovery, we want the user
		// to be redirected back to the original OAuth2 login flow.
		if hydraLoginRequest.RequestUrl != "" && h.d.Config().OAuth2ProviderOverrideReturnTo(r.Context()) {
			// replace the return_to query parameter
			q := r.URL.Query()
			q.Set("return_to", hydraLoginRequest.RequestUrl)
			r.URL.RawQuery = q.Encode()
		}
	}

	a, sess, err := h.NewLoginFlow(w, r, flow.TypeBrowser)
	if errors.Is(err, ErrAlreadyLoggedIn) {
		if hydraLoginRequest != nil {
			if !hydraLoginRequest.GetSkip() {
				h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, errors.WithStack(herodot.ErrInternalServerError.WithReason("ErrAlreadyLoggedIn indicated we can skip login, but Hydra asked us to refresh")))
				return
			}

			rt, err := h.d.Hydra().AcceptLoginRequest(r.Context(), string(hydraLoginChallenge), sess.IdentityID.String(), sess.AMR)
			if err != nil {
				h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
				return
			}
			returnTo, err := url.Parse(rt)
			if err != nil {
				h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to parse URL: %s", rt)))
				return
			}
			x.AcceptToRedirectOrJSON(w, r, h.d.Writer(), err, returnTo.String())
			return
		}

		returnTo, redirErr := x.SecureRedirectTo(r, h.d.Config().SelfServiceBrowserDefaultReturnTo(r.Context()),
			x.SecureRedirectAllowSelfServiceURLs(h.d.Config().SelfPublicURL(r.Context())),
			x.SecureRedirectAllowURLs(h.d.Config().SelfServiceBrowserAllowedReturnToDomains(r.Context())),
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

	a.HydraLoginRequest = hydraLoginRequest

	x.AcceptToRedirectOrJSON(w, r, h.d.Writer(), a, a.AppendTo(h.d.Config().SelfServiceFlowLoginUI(r.Context())).String())
}

// Get Login Flow Parameters
//
// swagger:parameters getLoginFlow
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type getLoginFlow struct {
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
	// When using the SDK in a browser app, on the server side you must include the HTTP Cookie Header
	// sent by the client to your server here. This ensures that CSRF and session cookies are respected.
	//
	// in: header
	// name: Cookie
	Cookies string `json:"Cookie"`
}

// swagger:route GET /self-service/login/flows frontend getLoginFlow
//
// # Get Login Flow
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
//	  const flow = await client.getLoginFlow(req.header('cookie'), req.query['flow'])
//
//	  res.render('login', flow)
//	})
//	```
//
// This request may fail due to several reasons. The `error.id` can be one of:
//
// - `session_already_available`: The user is already signed in.
// - `self_service_flow_expired`: The flow is expired and you should request a new one.
//
// More information can be found at [Ory Kratos User Login](https://www.ory.sh/docs/kratos/self-service/flows/user-login) and [User Registration Documentation](https://www.ory.sh/docs/kratos/self-service/flows/user-registration).
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Responses:
//	  200: loginFlow
//	  403: errorGeneric
//	  404: errorGeneric
//	  410: errorGeneric
//	  default: errorGeneric
func (h *Handler) getLoginFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
			redirectURL := flow.GetFlowExpiredRedirectURL(r.Context(), h.d.Config(), RouteInitBrowserFlow, ar.ReturnTo)

			h.d.Writer().WriteError(w, r, errors.WithStack(x.ErrGone.WithID(text.ErrIDSelfServiceFlowExpired).
				WithReason("The login flow has expired. Redirect the user to the login flow init endpoint to initialize a new login flow.").
				WithDetail("redirect_to", redirectURL.String()).
				WithDetail("return_to", ar.ReturnTo)))
			return
		}
		h.d.Writer().WriteError(w, r, errors.WithStack(x.ErrGone.WithID(text.ErrIDSelfServiceFlowExpired).
			WithReason("The login flow has expired. Call the login flow init API endpoint to initialize a new login flow.").
			WithDetail("api", urlx.AppendPaths(h.d.Config().SelfPublicURL(r.Context()), RouteInitAPIFlow).String())))
		return
	}

	if ar.OAuth2LoginChallenge != "" {
		hlr, err := h.d.Hydra().GetLoginRequest(r.Context(), string(ar.OAuth2LoginChallenge))
		if err != nil {
			// We don't redirect back to the third party on errors because Hydra doesn't
			// give us the 3rd party return_uri when it redirects to the login UI.
			h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
			return
		}
		ar.HydraLoginRequest = hlr
	}

	h.d.Writer().Write(w, r, ar)
}

// Update Login Flow Parameters
//
// swagger:parameters updateLoginFlow
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type updateLoginFlow struct {
	// The Login Flow ID
	//
	// The value for this parameter comes from `flow` URL Query parameter sent to your
	// application (e.g. `/login?flow=abcde`).
	//
	// required: true
	// in: query
	Flow string `json:"flow"`

	// in: body
	// required: true
	Body updateLoginFlowBody

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

// swagger:model updateLoginFlowBody
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type updateLoginFlowBody struct{}

// swagger:route POST /self-service/login frontend updateLoginFlow
//
// # Submit a Login Flow
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
//   - HTTP 410 if the original flow expired with the appropriate error messages set and optionally a `use_flow_id` parameter in the body;
//   - HTTP 400 on form validation errors.
//
// Browser flows expect a Content-Type of `application/x-www-form-urlencoded` or `application/json` to be sent in the body and respond with
//   - a HTTP 303 redirect to the post/after login URL or the `return_to` value if it was set and if the login succeeded;
//   - a HTTP 303 redirect to the login UI URL with the flow ID containing the validation errors otherwise.
//
// Browser flows with an accept header of `application/json` will not redirect but instead respond with
//   - HTTP 200 and a application/json body with the signed in identity and a `Set-Cookie` header on success;
//   - HTTP 303 redirect to a fresh login flow if the original flow expired with the appropriate error messages set;
//   - HTTP 400 on form validation errors.
//
// If this endpoint is called with `Accept: application/json` in the header, the response contains the flow without a redirect. In the
// case of an error, the `error.id` of the JSON response body can be one of:
//
//   - `session_already_available`: The user is already signed in.
//   - `security_csrf_violation`: Unable to fetch the flow because a CSRF violation occurred.
//   - `security_identity_mismatch`: The requested `?return_to` address is not allowed to be used. Adjust this in the configuration!
//   - `browser_location_change_required`: Usually sent when an AJAX request indicates that the browser needs to open a specific URL.
//     Most likely used in Social Sign In flows.
//
// More information can be found at [Ory Kratos User Login](https://www.ory.sh/docs/kratos/self-service/flows/user-login) and [User Registration Documentation](https://www.ory.sh/docs/kratos/self-service/flows/user-registration).
//
//	Schemes: http, https
//
//	Consumes:
//	- application/json
//	- application/x-www-form-urlencoded
//
//	Produces:
//	- application/json
//
//	Header:
//	- Set-Cookie
//
//	Responses:
//	  200: successfulNativeLogin
//	  303: emptyResponse
//	  400: loginFlow
//	  410: errorGeneric
//	  422: errorBrowserLocationChangeRequired
//	  default: errorGeneric
func (h *Handler) updateLoginFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

		http.Redirect(w, r, h.d.Config().SelfServiceBrowserDefaultReturnTo(r.Context()).String(), http.StatusSeeOther)
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
	var group node.UiNodeGroup
	for _, ss := range h.d.AllLoginStrategies() {
		interim, err := ss.Login(w, r, f, sess.IdentityID)
		group = ss.NodeGroup()
		if errors.Is(err, flow.ErrStrategyNotResponsible) {
			continue
		} else if errors.Is(err, flow.ErrCompletedByStrategy) {
			return
		} else if err != nil {
			h.d.LoginFlowErrorHandler().WriteFlowError(w, r, f, group, err)
			return
		}

		// What can happen is that we re-authenticate as another user. In this case, we need to use a completely fresh
		// session!
		if sess.IdentityID != uuid.Nil && sess.IdentityID != interim.ID {
			sess = session.NewInactiveSession()
		}

		method := ss.CompletedAuthenticationMethod(r.Context())
		sess.CompletedLoginFor(method.Method, method.AAL)
		i = interim
		break
	}

	if i == nil {
		h.d.LoginFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, errors.WithStack(schema.NewNoLoginStrategyResponsible()))
		return
	}

	if err := h.d.LoginHookExecutor().PostLoginHook(w, r, group, f, i, sess, ""); err != nil {
		if errors.Is(err, ErrAddressNotVerified) {
			h.d.LoginFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, errors.WithStack(schema.NewAddressNotVerifiedError()))
			return
		}

		h.d.LoginFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, err)
		return
	}
}
