package selfservice

import (
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/x/urlx"

	"github.com/ory/hive-cloud/hive/driver/configuration"
	"github.com/ory/hive-cloud/hive/errorx"
	"github.com/ory/hive-cloud/hive/session"
	"github.com/ory/hive-cloud/hive/x"
)

const (
	BrowserLoginPath                = "/auth/browser/login"
	BrowserLoginRequestsPath        = "/auth/browser/requests/login"
	BrowserRegistrationPath         = "/auth/browser/registration"
	BrowserRegistrationRequestsPath = "/auth/browser/requests/registration"
	BrowserLogoutPath               = "/auth/browser/logout"
)

type StrategyHandlerDependencies interface {
	StrategyProvider

	LoginExecutionProvider
	RegistrationExecutionProvider

	LoginRequestManagementProvider
	RegistrationRequestManagementProvider

	session.ManagementProvider
	errorx.ManagementProvider
	x.WriterProvider
}

type StrategyHandler struct {
	c configuration.Provider
	d StrategyHandlerDependencies
}

type StrategyHandlerProvider interface {
	StrategyHandler() *StrategyHandler
}

func NewStrategyHandler(d StrategyHandlerDependencies, c configuration.Provider) *StrategyHandler {
	return &StrategyHandler{d: d, c: c}
}

func (h *StrategyHandler) RegisterPublicRoutes(public *x.RouterPublic) {
	public.GET(BrowserLoginPath, h.initLoginRequest)
	public.GET(BrowserLoginRequestsPath, h.fetchLoginRequest)

	public.GET(BrowserRegistrationPath, h.initRegistrationRequest)
	public.GET(BrowserRegistrationRequestsPath, h.fetchRegistrationRequest)

	public.GET(BrowserLogoutPath, h.logout)

	for _, s := range h.d.SelfServiceStrategies() {
		s.SetRoutes(public)
	}
}

func (h *StrategyHandler) NewLoginRequest(w http.ResponseWriter, r *http.Request, redir func(request *LoginRequest) string) error {
	a := NewLoginRequest(h.c.SelfServiceLoginRequestLifespan(), r)
	for _, s := range h.d.SelfServiceStrategies() {
		if err := s.PopulateLoginMethod(r, a); err != nil {
			return err
		}
	}

	if err := h.d.LoginExecutor().PreLoginHook(w, r, a); err != nil {
		if errors.Cause(err) == ErrBreak {
			return nil
		}
		return err
	}

	if err := h.d.LoginRequestManager().CreateLoginRequest(r.Context(), a); err != nil {
		return err
	}

	http.Redirect(w,
		r,
		redir(a),
		http.StatusFound,
	)

	return nil
}

func (h *StrategyHandler) NewRegistrationRequest(w http.ResponseWriter, r *http.Request, redir func(*RegistrationRequest) string) error {

	a := NewRegistrationRequest(h.c.SelfServiceRegistrationRequestLifespan(), r)
	for _, s := range h.d.SelfServiceStrategies() {
		if err := s.PopulateRegistrationMethod(r, a); err != nil {
			return err
		}
	}

	if err := h.d.RegistrationExecutor().PreRegistrationHook(w, r, a); err != nil {
		if errors.Cause(err) == ErrBreak {
			return nil
		}
		return err
	}

	if err := h.d.RegistrationRequestManager().CreateRegistrationRequest(r.Context(), a); err != nil {
		return err
	}

	http.Redirect(w,
		r,
		redir(a),
		http.StatusFound,
	)

	return nil
}

func (h *StrategyHandler) initLoginRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.NewLoginRequest(w, r, func(a *LoginRequest) string {
		return urlx.CopyWithQuery(h.c.LoginURL(), url.Values{"request": {a.ID}}).String()
	}); err != nil {
		h.d.ErrorManager().ForwardError(w, r, err)
		return
	}
}

func (h *StrategyHandler) fetchLoginRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ar, err := h.d.LoginRequestManager().GetLoginRequest(r.Context(), r.URL.Query().Get("request"))
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	h.d.Writer().Write(w, r, ar)
}

func (h *StrategyHandler) logout(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.d.SessionManager().PurgeFromRequest(w, r); err != nil {
		h.d.ErrorManager().ForwardError(w, r, err)
		return
	}

	http.Redirect(w, r, h.c.SelfServiceLogoutRedirectURL().String(), http.StatusFound)
}

func (h *StrategyHandler) initRegistrationRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.NewRegistrationRequest(w, r, func(a *RegistrationRequest) string {
		return urlx.CopyWithQuery(h.c.RegisterURL(), url.Values{"request": {a.ID}}).String()
	}); err != nil {
		h.d.ErrorManager().ForwardError(w, r, err)
		return
	}
}

func (h *StrategyHandler) fetchRegistrationRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ar, err := h.d.RegistrationRequestManager().GetRegistrationRequest(r.Context(), r.URL.Query().Get("request"))
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	h.d.Writer().Write(w, r, ar)
}
