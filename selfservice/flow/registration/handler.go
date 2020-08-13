package registration

import (
	"net/http"
	"net/url"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/nosurf"
	"github.com/pkg/errors"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	RouteInitBrowserFlow = "/self-service/browser/flows/registration"
	RouteGetFlow         = "/self-service/browser/flows/requests/registration"
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
	public.GET(RouteGetFlow, h.publicFetchRegistrationRequest)
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(RouteGetFlow, h.adminFetchRegistrationRequest)
}

func (h *Handler) NewRegistrationRequest(w http.ResponseWriter, r *http.Request) (*Flow, error) {
	a := NewFlow(h.c.SelfServiceFlowRegistrationRequestLifespan(), h.d.GenerateCSRFToken(r), r, flow.TypeBrowser)
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

// swagger:route GET /self-service/browser/flows/registration public initializeSelfServiceBrowserRegistrationFlow
//
// Initialize browser-based registration user flow
//
// This endpoint initializes a browser-based user registration flow. Once initialized, the browser will be redirected to
// `selfservice.flows.registration.ui_url` with the request ID set as a query parameter. If a valid user session exists already, the browser will be
// redirected to `urls.default_redirect_url`.
//
// > This endpoint is NOT INTENDED for API clients and only works
// with browsers (Chrome, Firefox, ...).
//
// More information can be found at [ORY Kratos User Login and User Registration Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-login-user-registration).
//
//     Schemes: http, https
//
//     Responses:
//       302: emptyResponse
//       500: genericError
func (h *Handler) initBrowserFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	a, err := h.NewRegistrationRequest(w, r)
	if err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	redirTo := urlx.CopyWithQuery(h.c.SelfServiceFlowRegisterUI(), url.Values{"request": {a.ID.String()}}).String()
	if _, err := h.d.SessionManager().FetchFromRequest(r.Context(), r); err == nil {
		redirTo = h.c.SelfServiceBrowserDefaultReturnTo().String()
	}
	http.Redirect(w, r, redirTo, http.StatusFound)
}

// nolint:deadcode,unused
// swagger:parameters getSelfServiceBrowserRegistrationRequest
type getSelfServiceBrowserRegistrationRequestParameters struct {
	// Request is the Registration Request ID
	//
	// The value for this parameter comes from `request` URL Query parameter sent to your
	// application (e.g. `/registration?request=abcde`).
	//
	// required: true
	// in: query
	Request string `json:"request"`
}

// swagger:route GET /self-service/browser/flows/requests/registration common public admin getSelfServiceBrowserRegistrationRequest
//
// Get the request context of browser-based registration user flows
//
// This endpoint returns a registration request's context with, for example, error details and
// other information.
//
// When accessing this endpoint through ORY Kratos' Public API, ensure that cookies are set as they are required for CSRF to work. To prevent
// token scanning attacks, the public endpoint does not return 404 status codes.
//
// More information can be found at [ORY Kratos User Login and User Registration Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-login-user-registration).
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: registrationRequest
//       403: genericError
//       404: genericError
//       410: genericError
//       500: genericError
func (h *Handler) publicFetchRegistrationRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.fetchRegistrationRequest(w, r, true); err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

}

func (h *Handler) adminFetchRegistrationRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.fetchRegistrationRequest(w, r, false); err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}
}

func (h *Handler) fetchRegistrationRequest(w http.ResponseWriter, r *http.Request, isPublic bool) error {
	ar, err := h.d.RegistrationFlowPersister().GetRegistrationFlow(r.Context(), x.ParseUUID(r.URL.Query().Get("request")))
	if err != nil {
		if isPublic {
			return errors.WithStack(x.ErrInvalidCSRFToken.WithTrace(err).WithDebugf("%s", err))
		}
		return err
	}

	if isPublic && !nosurf.VerifyToken(h.d.GenerateCSRFToken(r), ar.CSRFToken) {
		return errors.WithStack(x.ErrInvalidCSRFToken)
	}

	if ar.ExpiresAt.Before(time.Now()) {
		return errors.WithStack(x.ErrGone.
			WithReason("The registration request has expired. Redirect the user to the login endpoint to initialize a new session.").
			WithDetail("redirect_to", urlx.AppendPaths(h.c.SelfPublicURL(), RouteInitBrowserFlow).String()))
	}

	h.d.Writer().Write(w, r, ar)
	return nil
}
