package registration

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
	RouteInitBrowserFlow = "/self-service/registration/browser"
	RouteInitAPIFlow     = "/self-service/registration/api"

	RouteGetFlow = "/self-service/registration/flows"
)

type (
	handlerDependencies interface {
		StrategyProvider
		errorx.ManagementProvider
		session.HandlerProvider
		session.ManagementProvider
		x.WriterProvider
		x.CSRFTokenGeneratorProvider
		HookExecutorProvider
		FlowPersistenceProvider
	}
	HandlerProvider interface {
		RegistrationHandler() *Handler
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
	public.GET(RouteInitBrowserFlow, h.d.SessionHandler().IsNotAuthenticated(h.initBrowserFlow, session.RedirectOnAuthenticated(h.c)))
	public.GET(RouteInitAPIFlow, h.d.SessionHandler().IsNotAuthenticated(h.initApiFlow,
		session.RespondWithJSONErrorOnAuthenticated(h.d.Writer(), errors.WithStack(ErrAlreadyLoggedIn))))

	public.GET(RouteGetFlow, h.fetchFlow)
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(RouteGetFlow, h.fetchFlow)
}

func (h *Handler) NewRegistrationFlow(w http.ResponseWriter, r *http.Request, ft flow.Type) (*Flow, error) {
	a := NewFlow(h.c.SelfServiceFlowRegistrationRequestLifespan(), h.d.GenerateCSRFToken(r), r, ft)
	for _, s := range h.d.RegistrationStrategies() {
		if err := s.PopulateRegistrationMethod(r, a); err != nil {
			return nil, err
		}
	}

	if err := h.d.RegistrationExecutor().PreRegistrationHook(w, r, a); err != nil {
		return nil, err
	}

	if err := h.d.RegistrationFlowPersister().CreateRegistrationFlow(r.Context(), a); err != nil {
		return nil, err
	}

	return a, nil
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
// More information can be found at [ORY Kratos User Login and User Registration Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-login-user-registration).
//
//     Schemes: http, https
//
//     Responses:
//       200: registrationFlow
//       400: genericError
//       500: genericError
func (h *Handler) initApiFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
// More information can be found at [ORY Kratos User Login and User Registration Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-login-user-registration).
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

	redirTo := a.AppendTo(h.c.SelfServiceFlowRegistrationUI()).String()
	if _, err := h.d.SessionManager().FetchFromRequest(r.Context(), r); err == nil {
		redirTo = h.c.SelfServiceBrowserDefaultReturnTo().String()
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
// More information can be found at [ORY Kratos User Login and User Registration Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-login-user-registration).
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
				WithDetail("redirect_to", urlx.AppendPaths(h.c.SelfPublicURL(), RouteInitBrowserFlow).String())))
			return
		}
		h.d.Writer().WriteError(w, r, errors.WithStack(x.ErrGone.
			WithReason("The registration flow has expired. Call the registration flow init API endpoint to initialize a new registration flow.").
			WithDetail("api", urlx.AppendPaths(h.c.SelfPublicURL(), RouteInitAPIFlow).String())))
		return
	}

	h.d.Writer().Write(w, r, ar)
}
