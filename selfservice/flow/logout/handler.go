package logout

import (
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/decoderx"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/urlx"

	"github.com/julienschmidt/httprouter"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	RouteInitBrowserFlow = "/self-service/logout/browser"
	RouteAPIFlow         = "/self-service/logout/api"
	RouteSubmitFlow      = "/self-service/logout"
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
		d  handlerDependencies
		dx *decoderx.HTTP
	}
)

func NewHandler(d handlerDependencies) *Handler {
	return &Handler{
		d:  d,
		dx: decoderx.NewHTTP(),
	}
}

func (h *Handler) RegisterPublicRoutes(router *x.RouterPublic) {
	h.d.CSRFHandler().ExemptPath(RouteAPIFlow)

	router.GET(RouteInitBrowserFlow, h.initializeSelfServiceLogoutForBrowsers)
	router.DELETE(RouteAPIFlow, h.submitSelfServiceLogoutFlowWithoutBrowser)
	router.GET(RouteSubmitFlow, h.submitLogout)
}

// swagger:model logoutUrl
type logoutURL struct {
	// LogoutURL can be opened in a browser to
	//
	// format: uri
	LogoutURL string `json:"logout_url"`
}

// swagger:parameters initializeSelfServiceLogoutForBrowsers
// nolint:deadcode,unused
type initializeSelfServiceLogoutForBrowsers struct {
	// in: header
	SessionCookie string `json:"X-Session-Cookie"`
}

// swagger:route GET /self-service/logout/browser public initializeSelfServiceLogoutForBrowsers
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
// call the `/self-service/logout/api` URL directly with the Ory Session Token.
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
func (h *Handler) initializeSelfServiceLogoutForBrowsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	sess, err := h.d.SessionManager().FetchFromRequest(r.Context(), r)
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	h.d.Writer().Write(w, r, &logoutURL{
		LogoutURL: urlx.CopyWithQuery(urlx.AppendPaths(h.d.Config(r.Context()).SelfPublicURL(r), RouteSubmitFlow),
			url.Values{"session_token": {sess.Token}}).String(),
	})
}

// swagger:parameters submitSelfServiceLogoutFlowWithoutBrowser
// nolint:deadcode,unused
type submitSelfServiceLogoutWithoutBrowser struct {
	// in: body
	// required: true
	Body submitSelfServiceLogoutFlowWithoutBrowserBody
}

// nolint:deadcode,unused
// swagger:model submitSelfServiceLogoutFlowWithoutBrowserBody
type submitSelfServiceLogoutFlowWithoutBrowserBody struct {
	// The Session Token
	//
	// Invalidate this session token.
	//
	// required: true
	SessionToken string `json:"session_token"`
}

// swagger:route DELETE /self-service/logout/api public submitSelfServiceLogoutFlowWithoutBrowser
//
// Perform Logout for APIs, Services, Apps, ...
//
// :::info
//
// This endpoint is EXPERIMENTAL and subject to potential breaking changes in the future.
//
// :::
//
// Use this endpoint to log out an identity using an Ory Session Token. If the Ory Session Token was successfully
// revoked, the server returns a 204 No Content response. A 204 No Content response is also sent when
// the Ory Session Token has been revoked already before.
//
// If the Ory Session Token is malformed or does not exist a 403 Forbidden response will be returned.
//
// This endpoint does not remove any HTTP
// Cookies - use the Browser-Based Self-Service Logout Flow instead.
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       204: emptyResponse
//       400: jsonError
//       500: jsonError
func (h *Handler) submitSelfServiceLogoutFlowWithoutBrowser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var p submitSelfServiceLogoutFlowWithoutBrowserBody
	if err := h.dx.Decode(r, &p,
		decoderx.HTTPJSONDecoder(),
		decoderx.HTTPDecoderAllowedMethods("DELETE")); err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	if err := h.d.SessionPersister().RevokeSessionByToken(r.Context(), p.SessionToken); err != nil {
		if errors.Is(err, sqlcon.ErrNoRows) {
			h.d.Writer().WriteError(w, r, errors.WithStack(herodot.ErrForbidden.WithReason("The provided Ory Session Token could not be found, is invalid, or otherwise malformed.")))
			return
		}

		h.d.Writer().WriteError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

// swagger:route POST /self-service/logout public submitSelfServiceLogoutFlow
//
// Complete Self-Service Logout
//
// :::info
//
// This endpoint is EXPERIMENTAL and subject to potential breaking changes in the future.
//
// :::
//
// This endpoint logs out an identity in a self-service manner.
//
// If the `Accept` HTTP header is not set to `application/json`, the browser will be redirected (HTTP 302 Found)
// to the `return_to` parameter of the initial request or fall back to `urls.default_return_to`.
//
// If the `Accept` HTTP header is set to `application/json`, a 204 No Content response
// will be sent on successful logout instead.
//
// :::note
//
// This endpoint is NOT INTENDED for API clients and only works
// with browsers (Chrome, Firefox, ...). For API clients you can
// call the `/self-service/logout/api` URL directly with the Ory Session Token.
//
// :::
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
func (h *Handler) submitLogout(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	token := r.URL.Query().Get("session_token")
	if len(token) == 0 {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, errors.WithStack(herodot.ErrBadRequest.WithReason("Please include a session_token in the URL query.")))
		return
	}

	sess, err := h.d.SessionManager().FetchFromRequest(r.Context(), r)
	if err != nil {
		// We could handle `session.ErrNoActiveSessionFound` gracefully with `h.completeLogout()` here but that would
		// actually be an issue as it incorrectly indicates to clients that the session has been removed even if
		// `RevokeSessionByToken` has not actually been called.
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	if sess.Token != token {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, errors.WithStack(herodot.ErrForbidden.WithReason("Unable to log out because the session token in the URL query does not match the session cookie.")))
		return
	}

	if err := h.d.SessionPersister().RevokeSessionByToken(r.Context(), token); err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	h.completeLogout(w, r)
}

func (h *Handler) completeLogout(w http.ResponseWriter, r *http.Request) {
	_ = h.d.CSRFHandler().RegenerateToken(w, r)

	ret, err := x.SecureRedirectTo(r, h.d.Config(r.Context()).SelfServiceFlowLogoutRedirectURL(),
		x.SecureRedirectUseSourceURL(r.RequestURI),
		x.SecureRedirectAllowURLs(h.d.Config(r.Context()).SelfServiceBrowserWhitelistedReturnToDomains()),
		x.SecureRedirectAllowSelfServiceURLs(h.d.Config(r.Context()).SelfPublicURL(r)),
	)
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
