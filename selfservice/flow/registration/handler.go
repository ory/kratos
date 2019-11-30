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
	BrowserRegistrationPath         = "/auth/browser/registration"
	BrowserRegistrationRequestsPath = "/auth/browser/requests/registration"
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

// swagger:route GET /auth/browser/registration public initializeRegistrationFlow
//
// Initialize a Registration Flow
//
// This endpoint initializes a registration flow. This endpoint **should not be called from a programatic API**
// but instead for the, for example, browser. It will redirect the user agent (e.g. browser) to the
// configured registration UI, appending the registration challenge.
//
// For an in-depth look at ORY Krato's registration flow, head over to: https://www.ory.sh/docs/kratos/selfservice/registration
//
//     Schemes: http, https
//
//     Responses:
//       302: emptyResponse
//       404: genericError
//       500: genericError
func (h *Handler) initRegistrationRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.NewRegistrationRequest(w, r, func(a *Request) string {
		return urlx.CopyWithQuery(h.c.RegisterURL(), url.Values{"request": {a.ID.String()}}).String()
	}); err != nil {
		h.d.ErrorManager().ForwardError(r.Context(), w, r, err)
		return
	}
}

// swagger:route GET /auth/browser/requests/registration public getRegistrationRequest
//
// Get Registration Request
//
// This endpoint returns a registration request's context with, for example, error details and
// other information.
//
// For an in-depth look at ORY Krato's registration flow, head over to: https://www.ory.sh/docs/kratos/selfservice/registration
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: registrationRequest
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
