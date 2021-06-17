package logout

import (
	"github.com/ory/x/urlx"
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	RouteInitBrowserFlow = "/self-service/logout/browser"
	RouteSubmitFlow     = "/self-service/logout"
)

type (
	handlerDependencies interface {
		x.WriterProvider
		x.CSRFProvider
		session.ManagementProvider
		session.PersistenceProvider
		errorx.ManagementProvider
		config.Provider
	}
	HandlerProvider interface {
		LogoutHandler() *Handler
	}
	Handler struct {
		d handlerDependencies
	}
)

func NewHandler(d handlerDependencies) *Handler {
	return &Handler{d: d}
}

func (h *Handler) RegisterPublicRoutes(router *x.RouterPublic) {
	router.GET(RouteInitBrowserFlow, h.generateLogoutURLs)
	router.GET(RouteSubmitFlow, h.logout)
}

// swagger:model logoutUrl
type logoutURL struct {
	// LogoutURL can be opened in a browser to
	//
	// format: uri
	LogoutURL string `json:"url"`
}

// swagger:route GET /self-service/logout/urls public initializeSelfServiceLogoutForBrowsers
//
// Initialize Logout Flow for Browsers
//
// :::info
//
// This endpoint is EXPERIMENTAL and subject to potential breaking changes in the future.
//
// :::
//
// This endpoint initializes a browser-based user logout flow and a URL which can be used to log out the user.
//
// :::note
//
// This endpoint is NOT INTENDED for API clients and only works
// with browsers (Chrome, Firefox, ...). For API clients you can
// call the `/self-service/logout` URL directly with the Ory Session Token.
//
// :::
//
// The URL is only valid for the currently signed in user. If no user is signed in, this endpoint returns
// a 401 error.
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: logoutUrl
//       401: jsonError
//       500: jsonError
func (h *Handler) generateLogoutURLs(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	sess, err := h.d.SessionManager().FetchFromRequest(r.Context(), r)
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	h.d.Writer().Write(w, r, &logoutURL{
		LogoutURL: urlx.CopyWithQuery(urlx.AppendPaths(h.d.Config(r.Context()).SelfPublicURL(r), RouteLogout),
			url.Values{"session_token": {sess.Token}}).String(),
	})
}

// nolint:deadcode,unused
// swagger:parameters submitSelfServiceLogout
type submitSelfServiceLogout struct {
	// A Valid Session Token
	//
	// If you do not have a session token because you only have a session cookie,
	// call `/self-service/logout/urls` to generate a URL for this endpoint.
	//
	// in: path
	SessionToken string `json:"session_token"`
}

// swagger:route POST /self-service/logout public submitSelfServiceLogout
//
// Complete Self-Service Logout
//
// This endpoint logs out an identity in a self-service manner.
//
// If the `Accept` HTTP header is not set to `application/json`, the browser will be redirected (HTTP 302 Found)
// to the `return_to` parameter of the initial request or fall back to `urls.default_return_to`.
//
// If the `Accept` HTTP header is set to `application/json`, a 204 No Content response
// will be sent on successful logout instead.
//
// More information can be found at [Ory Kratos User Logout Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-logout).
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       302: emptyResponse
//       204: emptyResponse
//       500: jsonError
func (h *Handler) logout(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ret, err := x.SecureRedirectTo(r, h.d.Config(r.Context()).SelfServiceFlowLogoutRedirectURL(),
		x.SecureRedirectUseSourceURL(r.RequestURI),
		x.SecureRedirectAllowURLs(h.d.Config(r.Context()).SelfServiceBrowserWhitelistedReturnToDomains()),
		x.SecureRedirectAllowSelfServiceURLs(h.d.Config(r.Context()).SelfPublicURL(r)),
	)
	if err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	_ = h.d.CSRFHandler().RegenerateToken(w, r)

	token := r.URL.Query().Get("session_token")
	if len(token) > 0 {
		err := h.d.SessionPersister().RevokeSessionByToken(r.Context(), token)
		if err != nil {
			h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
			return
		}

		if x.IsJSONRequest(r) {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		http.Redirect(w, r, ret.String(), http.StatusSeeOther)
		return
	}


	if err := h.d.SessionManager().PurgeFromRequest(r.Context(), w, r); err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	if x.IsJSONRequest(r) {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Redirect(w, r, ret.String(), http.StatusSeeOther)
	return
}
