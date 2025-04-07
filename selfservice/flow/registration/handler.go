// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package registration

import (
	"net/http"
	"net/url"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	hydraclientgo "github.com/ory/hydra-client-go/v2"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hydra"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/sessiontokenexchange"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/nosurf"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"
)

const (
	RouteInitBrowserFlow = "/self-service/registration/browser"
	RouteInitAPIFlow     = "/self-service/registration/api"

	RouteGetFlow = "/self-service/registration/flows"

	RouteSubmitFlow = "/self-service/registration"
)

type (
	handlerDependencies interface {
		config.Provider
		errorx.ManagementProvider
		hydra.Provider
		session.HandlerProvider
		session.ManagementProvider
		x.WriterProvider
		x.CSRFTokenGeneratorProvider
		x.CSRFProvider
		StrategyProvider
		HookExecutorProvider
		FlowPersistenceProvider
		ErrorHandlerProvider
		sessiontokenexchange.PersistenceProvider
		x.LoggingProvider
	}
	HandlerProvider interface {
		RegistrationHandler() *Handler
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

	public.GET(RouteInitBrowserFlow, h.createBrowserRegistrationFlow)
	public.GET(RouteInitAPIFlow, h.d.SessionHandler().IsNotAuthenticated(h.createNativeRegistrationFlow,
		session.RespondWithJSONErrorOnAuthenticated(h.d.Writer(), errors.WithStack(ErrAlreadyLoggedIn))))

	public.GET(RouteGetFlow, h.getRegistrationFlow)

	public.POST(RouteSubmitFlow, h.d.SessionHandler().IsNotAuthenticated(h.updateRegistrationFlow, h.onAuthenticated))
	public.GET(RouteSubmitFlow, h.d.SessionHandler().IsNotAuthenticated(h.updateRegistrationFlow, h.onAuthenticated))
}

func (h *Handler) onAuthenticated(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	handler := session.RedirectOnAuthenticated(h.d)
	if x.IsJSONRequest(r) {
		handler = session.RespondWithJSONErrorOnAuthenticated(h.d.Writer(), ErrAlreadyLoggedIn)
	}

	handler(w, r, ps)
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

func WithFlowOAuth2LoginChallenge(loginChallenge string) FlowOption {
	return func(f *Flow) {
		f.OAuth2LoginChallenge = sqlxx.NullString(loginChallenge)
	}
}

func (h *Handler) NewRegistrationFlow(w http.ResponseWriter, r *http.Request, ft flow.Type, opts ...FlowOption) (*Flow, error) {
	if !h.d.Config().SelfServiceFlowRegistrationEnabled(r.Context()) {
		return nil, errors.WithStack(ErrRegistrationDisabled)
	}

	f, err := NewFlow(h.d.Config(), h.d.Config().SelfServiceFlowRegistrationRequestLifespan(r.Context()), h.d.GenerateCSRFToken(r), r, ft)
	if err != nil {
		return nil, err
	}
	for _, o := range opts {
		o(f)
	}

	if ft == flow.TypeAPI && r.URL.Query().Get("return_session_token_exchange_code") == "true" {
		e, err := h.d.SessionTokenExchangePersister().CreateSessionTokenExchanger(r.Context(), f.ID)
		if err != nil {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithWrap(err))
		}
		f.SessionTokenExchangeCode = e.InitCode
	}

	for _, s := range h.d.RegistrationStrategies(r.Context(), PrepareOrganizations(r, f)...) {
		var populateErr error

		switch strategy := s.(type) {
		case FormHydrator:
			if h.d.Config().SelfServiceFlowRegistrationTwoSteps(r.Context()) {
				populateErr = strategy.PopulateRegistrationMethodProfile(r, f)
			} else {
				populateErr = strategy.PopulateRegistrationMethod(r, f)
			}
		case UnifiedFormHydrator:
			populateErr = strategy.PopulateRegistrationMethod(r, f)
		default:
			populateErr = errors.WithStack(x.PseudoPanic.WithReasonf("A registration strategy was expected to implement one of the interfaces UnifiedFormHydrator or FormHydrator but did not."))
		}

		if populateErr != nil {
			return nil, populateErr
		}
	}

	ds, err := h.d.Config().DefaultIdentityTraitsSchemaURL(r.Context())
	if err != nil {
		return nil, err
	}

	if err := SortNodes(r.Context(), f.UI.Nodes, ds.String()); err != nil {
		return nil, err
	}

	if err := h.d.RegistrationExecutor().PreRegistrationHook(w, r, f); err != nil {
		return nil, err
	}

	if err := h.d.RegistrationFlowPersister().CreateRegistrationFlow(r.Context(), f); err != nil {
		return nil, err
	}

	return f, nil
}

func (h *Handler) FromOldFlow(w http.ResponseWriter, r *http.Request, of Flow) (*Flow, error) {
	nf, err := h.NewRegistrationFlow(w, r, of.Type)
	if err != nil {
		return nil, err
	}

	nf.RequestURL = of.RequestURL
	return nf, nil
}

// swagger:route GET /self-service/registration/api frontend createNativeRegistrationFlow
//
// # Create Registration Flow for Native Apps
//
// This endpoint initiates a registration flow for API clients such as mobile devices, smart TVs, and so on.
//
// If a valid provided session cookie or session token is provided, a 400 Bad Request error
// will be returned unless the URL query parameter `?refresh=true` is set.
//
// To fetch an existing registration flow call `/self-service/registration/flows?flow=<flow_id>`.
//
// You MUST NOT use this endpoint in client-side (Single Page Apps, ReactJS, AngularJS) nor server-side (Java Server
// Pages, NodeJS, PHP, Golang, ...) browser applications. Using this endpoint in these applications will make
// you vulnerable to a variety of CSRF attacks.
//
// In the case of an error, the `error.id` of the JSON response body can be one of:
//
// - `session_already_available`: The user is already signed in.
// - `security_csrf_violation`: Unable to fetch the flow because a CSRF violation occurred.
//
// This endpoint MUST ONLY be used in scenarios such as native mobile apps (React Native, Objective C, Swift, Java, ...).
//
// More information can be found at [Ory Kratos User Login](https://www.ory.sh/docs/kratos/self-service/flows/user-login) and [User Registration Documentation](https://www.ory.sh/docs/kratos/self-service/flows/user-registration).
//
//	Schemes: http, https
//
//	Responses:
//	  200: registrationFlow
//	  400: errorGeneric
//	  default: errorGeneric
func (h *Handler) createNativeRegistrationFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	a, err := h.NewRegistrationFlow(w, r, flow.TypeAPI)
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	h.d.Writer().Write(w, r, a)
}

// Create Native Registration Flow Parameters
//
// swagger:parameters createNativeRegistrationFlow
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type createNativeRegistrationFlow struct {
	// EnableSessionTokenExchangeCode requests the login flow to include a code that can be used to retrieve the session token
	// after the login flow has been completed.
	//
	// in: query
	EnableSessionTokenExchangeCode bool `json:"return_session_token_exchange_code"`

	// The URL to return the browser to after the flow was completed.
	//
	// in: query
	ReturnTo string `json:"return_to"`

	// An optional organization ID that should be used to register this user.
	// This parameter is only effective in the Ory Network.
	//
	// required: false
	// in: query
	Organization string `json:"organization"`
}

// Create Browser Registration Flow Parameters
//
// swagger:parameters createBrowserRegistrationFlow
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type createBrowserRegistrationFlow struct {
	// The URL to return the browser to after the flow was completed.
	//
	// in: query
	ReturnTo string `json:"return_to"`

	// Ory OAuth 2.0 Login Challenge.
	//
	// If set will cooperate with Ory OAuth2 and OpenID to act as an OAuth2 server / OpenID Provider.
	//
	// The value for this parameter comes from `login_challenge` URL Query parameter sent to your
	// application (e.g. `/registration?login_challenge=abcde`).
	//
	// This feature is compatible with Ory Hydra when not running on the Ory Network.
	//
	// required: false
	// in: query
	LoginChallenge string `json:"login_challenge"`

	// The URL to return the browser to after the verification flow was completed.
	//
	// After the registration flow is completed, the user will be sent a verification email.
	// Upon completing the verification flow, this URL will be used to override the default
	// `selfservice.flows.verification.after.default_redirect_to` value.
	//
	// required: false
	// in: query
	AfterVerificationReturnTo string `json:"after_verification_return_to"`

	// An optional organization ID that should be used to register this user.
	// This parameter is only effective in the Ory Network.
	//
	// required: false
	// in: query
	Organization string `json:"organization"`
}

// swagger:route GET /self-service/registration/browser frontend createBrowserRegistrationFlow
//
// # Create Registration Flow for Browsers
//
// This endpoint initializes a browser-based user registration flow. This endpoint will set the appropriate
// cookies and anti-CSRF measures required for browser-based flows.
//
// If this endpoint is opened as a link in the browser, it will be redirected to
// `selfservice.flows.registration.ui_url` with the flow ID set as the query parameter `?flow=`. If a valid user session
// exists already, the browser will be redirected to `urls.default_redirect_url`.
//
// If this endpoint is called via an AJAX request, the response contains the flow without a redirect. In the
// case of an error, the `error.id` of the JSON response body can be one of:
//
// - `session_already_available`: The user is already signed in.
// - `security_csrf_violation`: Unable to fetch the flow because a CSRF violation occurred.
// - `security_identity_mismatch`: The requested `?return_to` address is not allowed to be used. Adjust this in the configuration!
//
// If this endpoint is called via an AJAX request, the response contains the registration flow without a redirect.
//
// This endpoint is NOT INTENDED for clients that do not have a browser (Chrome, Firefox, ...) as cookies are needed.
//
// More information can be found at [Ory Kratos User Login](https://www.ory.sh/docs/kratos/self-service/flows/user-login) and [User Registration Documentation](https://www.ory.sh/docs/kratos/self-service/flows/user-registration).
//
//	Schemes: http, https
//
//	Produces:
//	- application/json
//
//	Responses:
//	  200: registrationFlow
//	  303: emptyResponse
//	  default: errorGeneric
func (h *Handler) createBrowserRegistrationFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	a, err := h.NewRegistrationFlow(w, r, flow.TypeBrowser)
	if err != nil {
		h.d.SelfServiceErrorManager().Forward(ctx, w, r, err)
		return
	}

	var (
		hydraLoginRequest   *hydraclientgo.OAuth2LoginRequest
		hydraLoginChallenge sqlxx.NullString
	)
	if r.URL.Query().Has("login_challenge") {
		var err error
		hydraLoginChallenge, err = hydra.GetLoginChallengeID(h.d.Config(), r)
		if err != nil {
			h.d.SelfServiceErrorManager().Forward(ctx, w, r, err)
			return
		}

		hydraLoginRequest, err = h.d.Hydra().GetLoginRequest(ctx, hydraLoginChallenge.String())
		if err != nil {
			h.d.SelfServiceErrorManager().Forward(ctx, w, r, err)
			return
		}

		// on OAuth2 flows, we need to use the RequestURL
		// as the ReturnTo URL.
		// This is because a user might want to switch between
		// different flows, such as login to registration and login to recovery.
		// After completing a complex flow, such as recovery, we want the user
		// to be redirected back to the original OAuth2 login flow.
		if hydraLoginRequest.RequestUrl != "" && h.d.Config().OAuth2ProviderOverrideReturnTo(r.Context()) {
			q := r.URL.Query()
			// replace the return_to query parameter
			q.Set("return_to", hydraLoginRequest.RequestUrl)
			r.URL.RawQuery = q.Encode()
		}
	}

	if sess, err := h.d.SessionManager().FetchFromRequest(ctx, r); err == nil {
		if hydraLoginRequest != nil {
			if hydraLoginRequest.GetSkip() {
				rt, err := h.d.Hydra().AcceptLoginRequest(r.Context(),
					hydra.AcceptLoginRequestParams{
						LoginChallenge:        string(hydraLoginChallenge),
						IdentityID:            sess.IdentityID.String(),
						SessionID:             sess.ID.String(),
						AuthenticationMethods: sess.AMR,
					})
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

			// hydra indicates that we cannot skip the login request
			// so we must perform the login flow.
			// we directly go to the login handler from here
			// copy over any query parameters, such as `return_to` and `login_challenge`
			loginURL := urlx.CopyWithQuery(urlx.AppendPaths(h.d.Config().SelfPublicURL(ctx), "/self-service/login/browser"), x.RequestURL(r).Query())
			http.Redirect(
				w,
				r,
				loginURL.String(),
				http.StatusFound,
			)
			return
		}

		if x.IsJSONRequest(r) {
			h.d.Writer().WriteError(w, r, errors.WithStack(ErrAlreadyLoggedIn))
			return
		}

		returnTo, redirErr := x.SecureRedirectTo(r, h.d.Config().SelfServiceBrowserDefaultReturnTo(ctx),
			x.SecureRedirectAllowSelfServiceURLs(h.d.Config().SelfPublicURL(ctx)),
			x.SecureRedirectAllowURLs(h.d.Config().SelfServiceBrowserAllowedReturnToDomains(ctx)),
		)
		if redirErr != nil {
			h.d.SelfServiceErrorManager().Forward(ctx, w, r, redirErr)
			return
		}

		http.Redirect(w, r, returnTo.String(), http.StatusSeeOther)
		return
	}

	redirTo := a.AppendTo(h.d.Config().SelfServiceFlowRegistrationUI(ctx)).String()
	x.AcceptToRedirectOrJSON(w, r, h.d.Writer(), a, redirTo)
}

// Get Registration Flow Parameters
//
// swagger:parameters getRegistrationFlow
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type getRegistrationFlow struct {
	// The Registration Flow ID
	//
	// The value for this parameter comes from `flow` URL Query parameter sent to your
	// application (e.g. `/registration?flow=abcde`).
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

// swagger:route GET /self-service/registration/flows frontend getRegistrationFlow
//
// # Get Registration Flow
//
// This endpoint returns a registration flow's context with, for example, error details and other information.
//
// Browser flows expect the anti-CSRF cookie to be included in the request's HTTP Cookie Header.
// For AJAX requests you must ensure that cookies are included in the request or requests will fail.
//
// If you use the browser-flow for server-side apps, the services need to run on a common top-level-domain
// and you need to forward the incoming HTTP Cookie header to this endpoint:
//
//	```js
//	// pseudo-code example
//	router.get('/registration', async function (req, res) {
//	  const flow = await client.getRegistrationFlow(req.header('cookie'), req.query['flow'])
//
//	  res.render('registration', flow)
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
//	  200: registrationFlow
//	  403: errorGeneric
//	  404: errorGeneric
//	  410: errorGeneric
//	  default: errorGeneric
func (h *Handler) getRegistrationFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if !h.d.Config().SelfServiceFlowRegistrationEnabled(r.Context()) {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, errors.WithStack(ErrRegistrationDisabled))
		return
	}

	ar, err := h.d.RegistrationFlowPersister().GetRegistrationFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("id")))
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
				WithReason("The registration flow has expired. Redirect the user to the registration flow init endpoint to initialize a new registration flow.").
				WithDetail("redirect_to", redirectURL.String()).
				WithDetail("return_to", ar.ReturnTo)))
			return
		}
		h.d.Writer().WriteError(w, r, errors.WithStack(x.ErrGone.WithID(text.ErrIDSelfServiceFlowExpired).
			WithReason("The registration flow has expired. Call the registration flow init API endpoint to initialize a new registration flow.").
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

// Update Registration Flow Parameters
//
// swagger:parameters updateRegistrationFlow
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type updateRegistrationFlow struct {
	// The Registration Flow ID
	//
	// The value for this parameter comes from `flow` URL Query parameter sent to your
	// application (e.g. `/registration?flow=abcde`).
	//
	// required: true
	// in: query
	Flow string `json:"flow"`

	// in: body
	// required: true
	Body updateRegistrationFlowBody

	// HTTP Cookies
	//
	// When using the SDK in a browser app, on the server side you must include the HTTP Cookie Header
	// sent by the client to your server here. This ensures that CSRF and session cookies are respected.
	//
	// in: header
	// name: Cookie
	Cookies string `json:"Cookie"`
}

// Update Registration Request Body
//
// swagger:model updateRegistrationFlowBody
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type updateRegistrationFlowBody struct{}

// swagger:route POST /self-service/registration frontend updateRegistrationFlow
//
// # Update Registration Flow
//
// Use this endpoint to complete a registration flow by sending an identity's traits and password. This endpoint
// behaves differently for API and browser flows.
//
// API flows expect `application/json` to be sent in the body and respond with
//   - HTTP 200 and a application/json body with the created identity success - if the session hook is configured the
//     `session` and `session_token` will also be included;
//   - HTTP 410 if the original flow expired with the appropriate error messages set and optionally a `use_flow_id` parameter in the body;
//   - HTTP 400 on form validation errors.
//
// Browser flows expect a Content-Type of `application/x-www-form-urlencoded` or `application/json` to be sent in the body and respond with
//   - a HTTP 303 redirect to the post/after registration URL or the `return_to` value if it was set and if the registration succeeded;
//   - a HTTP 303 redirect to the registration UI URL with the flow ID containing the validation errors otherwise.
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
//	Responses:
//	  200: successfulNativeRegistration
//	  303: emptyResponse
//	  400: registrationFlow
//	  410: errorGeneric
//	  422: errorBrowserLocationChangeRequired
//	  default: errorGeneric
func (h *Handler) updateRegistrationFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	rid, err := flow.GetFlowID(r)
	if err != nil {
		h.d.RegistrationFlowErrorHandler().WriteFlowError(w, r, nil, node.DefaultGroup, err)
		return
	}

	f, err := h.d.RegistrationFlowPersister().GetRegistrationFlow(r.Context(), rid)
	if err != nil {
		h.d.RegistrationFlowErrorHandler().WriteFlowError(w, r, nil, node.DefaultGroup, err)
		return
	}

	if _, err := h.d.SessionManager().FetchFromRequest(r.Context(), r); err == nil {
		if f.Type == flow.TypeBrowser {
			http.Redirect(w, r, h.d.Config().SelfServiceBrowserDefaultReturnTo(r.Context()).String(), http.StatusSeeOther)
			return
		}

		h.d.Writer().WriteError(w, r, errors.WithStack(ErrAlreadyLoggedIn))
		return
	}

	if err := f.Valid(); err != nil {
		h.d.RegistrationFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, err)
		return
	}

	i := identity.NewIdentity(h.d.Config().DefaultIdentityTraitsSchemaID(r.Context()))
	var s Strategy
	for _, ss := range h.d.AllRegistrationStrategies() {
		if err := ss.Register(w, r, f, i); errors.Is(err, flow.ErrStrategyNotResponsible) {
			continue
		} else if errors.Is(err, flow.ErrCompletedByStrategy) {
			return
		} else if err != nil {
			h.d.RegistrationFlowErrorHandler().WriteFlowError(w, r, f, ss.NodeGroup(), err)
			return
		}

		s = ss
		break
	}

	if s == nil {
		h.d.RegistrationFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, errors.WithStack(schema.NewNoRegistrationStrategyResponsible()))
		return
	}

	if err := h.d.RegistrationExecutor().PostRegistrationHook(w, r, s.ID(), "", "", f, i); err != nil {
		h.d.RegistrationFlowErrorHandler().WriteFlowError(w, r, f, s.NodeGroup(), err)
		return
	}
}
