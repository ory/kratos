package login

import (
	"net/http"
	"net/url"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/nosurf"
	"github.com/pkg/errors"

	"github.com/ory/x/errorsx"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	BrowserLoginPath         = "/self-service/browser/flows/login"
	BrowserLoginRequestsPath = "/self-service/browser/flows/requests/login"
)

type (
	handlerDependencies interface {
		HookExecutorProvider
		RequestPersistenceProvider
		errorx.ManagementProvider
		StrategyProvider
		session.HandlerProvider
		session.ManagementProvider
		x.WriterProvider
		x.CSRFTokenGeneratorProvider
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
	public.GET(BrowserLoginPath, h.initLoginRequest)
	public.GET(BrowserLoginRequestsPath, h.publicFetchLoginRequest)
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(BrowserLoginRequestsPath, h.adminFetchLoginRequest)
}

func (h *Handler) NewLoginRequest(w http.ResponseWriter, r *http.Request, redir func(request *Request) (string, error)) error {
	a := NewLoginRequest(h.c.SelfServiceLoginRequestLifespan(), h.d.GenerateCSRFToken(r), r)
	for _, s := range h.d.LoginStrategies() {
		if err := s.PopulateLoginMethod(r, a); err != nil {
			return err
		}
	}

	if err := h.d.LoginHookExecutor().PreLoginHook(w, r, a); err != nil {
		if errorsx.Cause(err) == ErrHookAbortRequest {
			return nil
		}
		return err
	}

	if err := h.d.LoginRequestPersister().CreateLoginRequest(r.Context(), a); err != nil {
		return err
	}

	to, err := redir(a)
	if err != nil {
		return err
	}
	http.Redirect(w,
		r,
		to,
		http.StatusFound,
	)

	return nil
}

// swagger:route GET /self-service/browser/flows/login public initializeSelfServiceBrowserLoginFlow
//
// Initialize browser-based login user flow
//
// This endpoint initializes a browser-based user login flow. Once initialized, the browser will be redirected to
// `urls.login_ui` with the request ID set as a query parameter. If a valid user session exists already, the browser will be
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
func (h *Handler) initLoginRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.NewLoginRequest(w, r, func(a *Request) (string, error) {
		// we assume an error means the user has no session
		if _, err := h.d.SessionManager().FetchFromRequest(r.Context(), w, r); err != nil {
			return urlx.CopyWithQuery(h.c.LoginURL(), url.Values{"request": {a.ID.String()}}).String(), nil
		}

		if r.URL.Query().Get("prompt") == "login" {
			if err := h.d.LoginRequestPersister().MarkRequestForced(r.Context(), a.ID); err != nil {
				return "", err
			}
			return urlx.CopyWithQuery(h.c.LoginURL(), url.Values{"request": {a.ID.String()}}).String(), nil
		}

		returnTo, err := x.DetermineReturnToURL(r.URL, h.c.DefaultReturnToURL(), []url.URL{*h.c.SelfPublicURL()})
		if err != nil {
			h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		}

		return returnTo, nil
	}); err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}
}

// nolint:deadcode,unused
// swagger:parameters getSelfServiceBrowserLoginRequest
type getSelfServiceBrowserLoginRequestParameters struct {
	// Request is the Login Request ID
	//
	// The value for this parameter comes from `request` URL Query parameter sent to your
	// application (e.g. `/login?request=abcde`).
	//
	// required: true
	// in: query
	Request string `json:"request"`
}

// swagger:route GET /self-service/browser/flows/requests/login common public admin getSelfServiceBrowserLoginRequest
//
// Get the request context of browser-based login user flows
//
// This endpoint returns a login request's context with, for example, error details and
// other information.
//
// When accessing this endpoint through ORY Kratos' Public API, ensure that cookies are set as they are required for CSRF to work. To prevent
// token scanning attacks, the public endpoint does not return 404 status codes to prevent scanning attacks.
//
// More information can be found at [ORY Kratos User Login and User Registration Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-login-user-registration).
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: loginRequest
//       403: genericError
//       404: genericError
//       410: genericError
//       500: genericError
func (h *Handler) publicFetchLoginRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.fetchLoginRequest(w, r, true); err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}
}

func (h *Handler) adminFetchLoginRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.fetchLoginRequest(w, r, false); err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}
}

func (h *Handler) fetchLoginRequest(w http.ResponseWriter, r *http.Request, isPublic bool) error {
	ar, err := h.d.LoginRequestPersister().GetLoginRequest(r.Context(), x.ParseUUID(r.URL.Query().Get("request")))
	if err != nil {
		if isPublic {
			return errors.WithStack(x.ErrInvalidCSRFToken.WithTrace(err).WithDebugf("%s", err))
		}
		return err
	}

	if isPublic {
		if !nosurf.VerifyToken(h.d.GenerateCSRFToken(r), ar.CSRFToken) {
			return errors.WithStack(x.ErrInvalidCSRFToken)
		}
	}

	if ar.ExpiresAt.Before(time.Now()) {
		return errors.WithStack(x.ErrGone.
			WithReason("The login request has expired. Redirect the user to the login endpoint to initialize a new session.").
			WithDetail("redirect_to", urlx.AppendPaths(h.c.SelfPublicURL(), BrowserLoginPath).String()))
	}

	h.d.Writer().Write(w, r, ar)
	return nil
}
