package logout

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	RouteBrowser = "/self-service/browser/flows/logout"
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
		c *config.Config
		d handlerDependencies
	}
)

func NewHandler(d handlerDependencies, c *config.Config) *Handler {
	return &Handler{d: d, c: c}
}

func (h *Handler) RegisterPublicRoutes(router *x.RouterPublic) {
	router.GET(RouteBrowser, h.logout)
}

// swagger:route GET /self-service/browser/flows/logout public initializeSelfServiceBrowserLogoutFlow
//
// Initialize Browser-Based Logout User Flow
//
// This endpoint initializes a logout flow.
//
// > This endpoint is NOT INTENDED for API clients and only works
// with browsers (Chrome, Firefox, ...).
//
// On successful logout, the browser will be redirected (HTTP 302 Found) to the `return_to` parameter of the initial request
// or fall back to `urls.default_return_to`.
//
// More information can be found at [ORY Kratos User Logout Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-logout).
//
//     Schemes: http, https
//
//     Responses:
//       302: emptyResponse
//       500: genericError
func (h *Handler) logout(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	_ = h.d.CSRFHandler().RegenerateToken(w, r)

	if err := h.d.SessionManager().PurgeFromRequest(r.Context(), w, r); err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	ret, err := x.SecureRedirectTo(r, h.c.SelfServiceFlowLogoutRedirectURL(),
		x.SecureRedirectUseSourceURL(r.RequestURI),
		x.SecureRedirectAllowURLs(h.c.SelfServiceBrowserWhitelistedReturnToDomains()),
		x.SecureRedirectAllowSelfServiceURLs(h.c.SelfPublicURL()),
	)
	if err != nil {
		fmt.Printf("\n%s\n\n", err.Error())
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	http.Redirect(w, r, ret.String(), http.StatusFound)
}
