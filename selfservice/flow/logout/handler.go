package logout

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	BrowserLogoutPath = "/auth/browser/logout"
)

type (
	handlerDependencies interface {
		x.CSRFProvider
		session.ManagementProvider
		errorx.ManagementProvider
	}
	HandlerProvider interface {
		LogoutHandler() *Handler
	}
	Handler struct {
		c configuration.Provider
		d handlerDependencies
	}
)

func NewHandler(d handlerDependencies, c configuration.Provider) *Handler {
	return &Handler{d: d, c: c}
}

func (h *Handler) RegisterPublicRoutes(router *x.RouterPublic) {
	router.GET(BrowserLogoutPath, h.logout)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	_ = h.d.CSRFHandler().RegenerateToken(w, r)

	if err := h.d.SessionManager().PurgeFromRequest(r.Context(), w, r); err != nil {
		h.d.SelfServiceErrorManager().ForwardError(r.Context(), w, r, err)
		return
	}

	http.Redirect(w, r, h.c.SelfServiceLogoutRedirectURL().String(), http.StatusFound)
}
