package registration

import (
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"

	"github.com/ory/x/errorsx"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	BrowserRegistrationPath         = "/self-service/browser/flows/registration"
	BrowserRegistrationRequestsPath = "/self-service/browser/flows/requests/registration"
)

type (
	handlerDependencies interface {
		StrategyProvider
		errorx.ManagementProvider
		session.HandlerProvider
		x.WriterProvider
		HookExecutorProvider
		RequestPersistenceProvider
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
	public.GET(BrowserRegistrationPath, h.d.SessionHandler().IsNotAuthenticated(h.initRegistrationRequest, session.RedirectOnAuthenticated(h.c)))
	public.GET(BrowserRegistrationRequestsPath, h.fetchRegistrationRequest)
}

func (h *Handler) NewRegistrationRequest(w http.ResponseWriter, r *http.Request, redir func(*Request) string) error {
	a := NewRequest(
		h.c.SelfServiceRegistrationRequestLifespan(),
		r,
	)
	for _, s := range h.d.RegistrationStrategies() {
		if err := s.PopulateRegistrationMethod(r, a); err != nil {
			return err
		}
	}

	if err := h.d.RegistrationExecutor().PreRegistrationHook(w, r, a); err != nil {
		if errorsx.Cause(err) == ErrHookAbortRequest {
			return nil
		}
		return err
	}

	if err := h.d.RegistrationRequestPersister().CreateRegistrationRequest(r.Context(), a); err != nil {
		return err
	}

	http.Redirect(w,
		r,
		redir(a),
		http.StatusFound,
	)

	return nil
}

// swagger:route GET /self-service/browser/flows/registration public initializeSelfServiceBrowserRegistrationFlow
//
// Initialize browser-based registration user flow
//
// This endpoint initializes a browser-based user registration flow. Once initialized, the browser will be redirected to
// `urls.registration_ui` with the request ID set as a query parameter. If a valid user session exists already, the browser will be
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
func (h *Handler) initRegistrationRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.NewRegistrationRequest(w, r, func(a *Request) string {
		return urlx.CopyWithQuery(h.c.RegisterURL(), url.Values{"request": {a.ID.String()}}).String()
	}); err != nil {
		h.d.SelfServiceErrorManager().ForwardError(r.Context(), w, r, err)
		return
	}
}

// nolint:deadcode,unused
// swagger:parameters: getSelfServiceBrowserRegistrationRequest
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

// swagger:route GET /self-service/browser/flows/requests/registration public getSelfServiceBrowserRegistrationRequest
//
// Get the request context of browser-based registration user flows
//
// This endpoint returns a registration request's context with, for example, error details and
// other information.
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
//       500: genericError
func (h *Handler) fetchRegistrationRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ar, err := h.d.RegistrationRequestPersister().GetRegistrationRequest(r.Context(), x.ParseUUID(r.URL.Query().Get("request")))
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	h.d.Writer().Write(w, r, ar)
}
