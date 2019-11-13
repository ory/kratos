package login

import (
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	BrowserLoginPath         = "/auth/browser/login"
	BrowserLoginRequestsPath = "/auth/browser/requests/login"
)

type (
	handlerDependencies interface {
		HookExecutorProvider
		RequestPersistenceProvider
		errorx.ManagementProvider
		StrategyProvider
		session.HandlerProvider
		x.WriterProvider
	}
	HandlerProvider interface {
		LoginHandler() *Handler
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
	public.GET(BrowserLoginPath, h.d.SessionHandler().IsNotAuthenticated(h.initLoginRequest, session.RedirectOnAuthenticated(h.c)))
	public.GET(BrowserLoginRequestsPath, h.fetchLoginRequest)
}

func (h *Handler) NewLoginRequest(w http.ResponseWriter, r *http.Request, redir func(request *Request) string) error {
	a := NewLoginRequest(h.c.SelfServiceLoginRequestLifespan(), r)
	for _, s := range h.d.LoginStrategies() {
		if err := s.PopulateLoginMethod(r, a); err != nil {
			return err
		}
	}

	if err := h.d.LoginHookExecutor().PreLoginHook(w, r, a); err != nil {
		if errors.Cause(err) == ErrHookAbortRequest {
			return nil
		}
		return err
	}

	if err := h.d.LoginRequestPersister().CreateLoginRequest(r.Context(), a); err != nil {
		return err
	}

	http.Redirect(w,
		r,
		redir(a),
		http.StatusFound,
	)

	return nil
}

// swagger:route GET /auth/browser/login public initializeLoginFlow
//
// Initialize a Login Flow
//
// This endpoint initializes a login flow. This endpoint **should not be called from a programatic API**
// but instead for the, for example, browser. It will redirect the user agent (e.g. browser) to the
// configured login UI, appending the login challenge.
//
// If the user-agent already has a valid authentication session, the server will respond with a 302
// code redirecting to the config value of `urls.default_return_to`.
//
// For an in-depth look at ORY Krato's login flow, head over to: https://www.ory.sh/docs/kratos/selfservice/login
//
//     Schemes: http, https
//
//     Responses:
//       302: emptyResponse
//       500: genericError
func (h *Handler) initLoginRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.NewLoginRequest(w, r, func(a *Request) string {
		return urlx.CopyWithQuery(h.c.LoginURL(), url.Values{"request": {a.ID}}).String()
	}); err != nil {
		h.d.ErrorManager().ForwardError(r.Context(), w, r, err)
		return
	}
}

// swagger:route GET /auth/browser/requests/login public getLoginRequest
//
// Get Login Request
//
// This endpoint returns a login request's context with, for example, error details and
// other information.
//
// For an in-depth look at ORY Krato's login flow, head over to: https://www.ory.sh/docs/kratos/selfservice/login
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: loginRequest
//       302: emptyResponse
//       500: genericError
func (h *Handler) fetchLoginRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ar, err := h.d.LoginRequestPersister().GetLoginRequest(r.Context(), r.URL.Query().Get("request"))
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	h.d.Writer().Write(w, r, ar.Declassify())
}
