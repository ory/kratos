package registration

import (
	"net/http"
	"time"

	"github.com/ory/kratos/schema"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/ui/node"

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
	RouteInitBrowserFlow = "/self-service/registration/browser"
	RouteInitAPIFlow     = "/self-service/registration/api"

	RouteGetFlow = "/self-service/registration/flows"

	RouteSubmitFlow = "/self-service/registration"
)

type (
	handlerDependencies interface {
		config.Provider
		errorx.ManagementProvider
		session.HandlerProvider
		session.ManagementProvider
		x.WriterProvider
		x.CSRFTokenGeneratorProvider
		x.CSRFProvider
		StrategyProvider
		HookExecutorProvider
		FlowPersistenceProvider
		ErrorHandlerProvider
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

	public.GET(RouteInitBrowserFlow, h.d.SessionHandler().IsNotAuthenticated(h.initBrowserFlow, session.RedirectOnAuthenticated(h.d)))
	public.GET(RouteInitAPIFlow, h.d.SessionHandler().IsNotAuthenticated(h.initApiFlow,
		session.RespondWithJSONErrorOnAuthenticated(h.d.Writer(), errors.WithStack(ErrAlreadyLoggedIn))))

	public.GET(RouteGetFlow, h.fetchFlow)

	public.POST(RouteSubmitFlow, h.d.SessionHandler().IsNotAuthenticated(h.submitFlow, h.onAuthenticated))
	public.GET(RouteSubmitFlow, h.d.SessionHandler().IsNotAuthenticated(h.submitFlow, h.onAuthenticated))
}

func (h *Handler) onAuthenticated(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	handler := session.RedirectOnAuthenticated(h.d)
	if x.IsJSONRequest(r) {
		handler = session.RespondWithJSONErrorOnAuthenticated(h.d.Writer(), ErrAlreadyLoggedIn)
	}

	handler(w, r, ps)
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(RouteGetFlow, h.fetchFlow)
}

func (h *Handler) NewRegistrationFlow(w http.ResponseWriter, r *http.Request, ft flow.Type) (*Flow, error) {
	f := NewFlow(h.d.Config(r.Context()), h.d.Config(r.Context()).SelfServiceFlowRegistrationRequestLifespan(), h.d.GenerateCSRFToken(r), r, ft)
	for _, s := range h.d.RegistrationStrategies(r.Context()) {
		if err := s.PopulateRegistrationMethod(r, f); err != nil {
			return nil, err
		}
	}

	if err := SortNodes(f.UI.Nodes, h.d.Config(r.Context()).DefaultIdentityTraitsSchemaURL().String()); err != nil {
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

// swagger:route GET /self-service/registration/api public initializeSelfServiceRegistrationViaAPIFlow
//
// Initialize Registration Flow for API clients
//
// This endpoint initiates a registration flow for API clients such as mobile devices, smart TVs, and so on.
//
// If a valid provided session cookie or session token is provided, a 400 Bad Request error
// will be returned unless the URL query parameter `?refresh=true` is set.
//
// To fetch an existing registration flow call `/self-service/registration/flows?flow=<flow_id>`.
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
// More information can be found at [Ory Kratos User Login and User Registration Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-login-user-registration).
//
//     Schemes: http, https
//
//     Responses:
//       200: registrationFlow
//       400: genericError
//       500: genericError
func (h *Handler) initApiFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	a, err := h.NewRegistrationFlow(w, r, flow.TypeAPI)
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	h.d.Writer().Write(w, r, a)
}

// swagger:route GET /self-service/registration/browser public initializeSelfServiceRegistrationViaBrowserFlow
//
// Initialize Registration Flow for browsers
//
// This endpoint initializes a browser-based user registration flow. Once initialized, the browser will be redirected to
// `selfservice.flows.registration.ui_url` with the flow ID set as the query parameter `?flow=`. If a valid user session
// exists already, the browser will be redirected to `urls.default_redirect_url` unless the query parameter
// `?refresh=true` was set.
//
// :::note
//
// This endpoint is NOT INTENDED for API clients and only works with browsers (Chrome, Firefox, ...).
//
// :::
//
// More information can be found at [Ory Kratos User Login and User Registration Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-login-user-registration).
//
//     Schemes: http, https
//
//     Responses:
//       302: emptyResponse
//       500: genericError
func (h *Handler) initBrowserFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	a, err := h.NewRegistrationFlow(w, r, flow.TypeBrowser)
	if err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	redirTo := a.AppendTo(h.d.Config(r.Context()).SelfServiceFlowRegistrationUI()).String()
	if _, err := h.d.SessionManager().FetchFromRequest(r.Context(), r); err == nil {
		redirTo = h.d.Config(r.Context()).SelfServiceBrowserDefaultReturnTo().String()
	}
	http.Redirect(w, r, redirTo, http.StatusFound)
}

// nolint:deadcode,unused
// swagger:parameters getSelfServiceRegistrationFlow
type getSelfServiceRegistrationFlowParameters struct {
	// The Registration Flow ID
	//
	// The value for this parameter comes from `flow` URL Query parameter sent to your
	// application (e.g. `/registration?flow=abcde`).
	//
	// required: true
	// in: query
	ID string `json:"id"`
}

// swagger:route GET /self-service/registration/flows public admin getSelfServiceRegistrationFlow
//
// Get Registration Flow
//
// This endpoint returns a registration flow's context with, for example, error details and other information.
//
// More information can be found at [Ory Kratos User Login and User Registration Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-login-user-registration).
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: registrationFlow
//       403: genericError
//       404: genericError
//       410: genericError
//       500: genericError
func (h *Handler) fetchFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ar, err := h.d.RegistrationFlowPersister().GetRegistrationFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("id")))
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	if ar.ExpiresAt.Before(time.Now()) {
		if ar.Type == flow.TypeBrowser {
			h.d.Writer().WriteError(w, r, errors.WithStack(x.ErrGone.
				WithReason("The registration flow has expired. Redirect the user to the registration flow init endpoint to initialize a new registration flow.").
				WithDetail("redirect_to", urlx.AppendPaths(h.d.Config(r.Context()).SelfPublicURL(r), RouteInitBrowserFlow).String())))
			return
		}
		h.d.Writer().WriteError(w, r, errors.WithStack(x.ErrGone.
			WithReason("The registration flow has expired. Call the registration flow init API endpoint to initialize a new registration flow.").
			WithDetail("api", urlx.AppendPaths(h.d.Config(r.Context()).SelfPublicURL(r), RouteInitAPIFlow).String())))
		return
	}

	h.d.Writer().Write(w, r, ar)
}

// nolint:deadcode,unused
// swagger:parameters submitSelfServiceRegistrationFlow
type submitSelfServiceRegistrationFlow struct {
	// The Registration Flow ID
	//
	// The value for this parameter comes from `flow` URL Query parameter sent to your
	// application (e.g. `/registration?flow=abcde`).
	//
	// required: true
	// in: query
	Flow string `json:"flow"`

	// in: body
	Body submitSelfServiceRegistrationFlowBody
}

// swagger:model submitSelfServiceRegistrationFlow
// nolint:deadcode,unused
type submitSelfServiceRegistrationFlowBody struct{}

// swagger:route POST /self-service/registration public submitSelfServiceRegistrationFlow
//
// Submit a Registration Flow
//
// Use this endpoint to complete a registration flow by sending an identity's traits and password. This endpoint
// behaves differently for API and browser flows.
//
// API flows expect `application/json` to be sent in the body and respond with
//   - HTTP 200 and a application/json body with the created identity success - if the session hook is configured the
//     `session` and `session_token` will also be included;
//   - HTTP 302 redirect to a fresh registration flow if the original flow expired with the appropriate error messages set;
//   - HTTP 400 on form validation errors.
//
// Browser flows expect `application/x-www-form-urlencoded` to be sent in the body and responds with
//   - a HTTP 302 redirect to the post/after registration URL or the `return_to` value if it was set and if the registration succeeded;
//   - a HTTP 302 redirect to the registration UI URL with the flow ID containing the validation errors otherwise.
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
//     Responses:
//       200: registrationViaApiResponse
//       302: emptyResponse
//       400: registrationFlow
//       500: genericError
func (h *Handler) submitFlow(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
			http.Redirect(w, r, h.d.Config(r.Context()).SelfServiceBrowserDefaultReturnTo().String(), http.StatusFound)
			return
		}

		h.d.Writer().WriteError(w, r, errors.WithStack(ErrAlreadyLoggedIn))
		return
	}

	if err := f.Valid(); err != nil {
		h.d.RegistrationFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, err)
		return
	}

	i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
	var found bool
	var s identity.CredentialsType
	for _, ss := range h.d.AllRegistrationStrategies() {
		if err := ss.Register(w, r, f, i); errors.Is(err, flow.ErrStrategyNotResponsible) {
			continue
		} else if errors.Is(err, flow.ErrCompletedByStrategy) {
			return
		} else if err != nil {
			h.d.RegistrationFlowErrorHandler().WriteFlowError(w, r, f, ss.NodeGroup(), err)
			return
		}

		s = ss.ID()
		found = true
		break
	}

	if !found {
		h.d.RegistrationFlowErrorHandler().WriteFlowError(w, r, f, node.DefaultGroup, errors.WithStack(schema.NewNoRegistrationStrategyResponsible()))
		return
	}

	if err := h.d.RegistrationExecutor().PostRegistrationHook(w, r, s, f, i); err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}
}
