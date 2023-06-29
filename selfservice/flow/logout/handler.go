// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package logout

import (
	"net/http"
	"net/url"

	"go.opentelemetry.io/otel/trace"

	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x/events"

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
	h.d.CSRFHandler().IgnorePath(RouteAPIFlow)

	router.GET(RouteInitBrowserFlow, h.createBrowserLogoutFlow)
	router.DELETE(RouteAPIFlow, h.performNativeLogout)
	router.GET(RouteSubmitFlow, h.updateLogoutFlow)
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(RouteInitBrowserFlow, x.RedirectToPublicRoute(h.d))
	admin.DELETE(RouteAPIFlow, x.RedirectToPublicRoute(h.d))
	admin.GET(RouteSubmitFlow, x.RedirectToPublicRoute(h.d))
}

// Logout Flow
//
// swagger:model logoutFlow
type logoutFlow struct {
	// LogoutURL can be opened in a browser to sign the user out.
	//
	// format: uri
	// required: true
	LogoutURL string `json:"logout_url"`

	// LogoutToken can be used to perform logout using AJAX.
	//
	// required: true
	LogoutToken string `json:"logout_token"`
}

// Create Browser Logout Flow Parameters
//
// swagger:parameters createBrowserLogoutFlow
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type createBrowserLogoutFlow struct {
	// HTTP Cookies
	//
	// If you call this endpoint from a backend, please include the
	// original Cookie header in the request.
	//
	// in: header
	// name: cookie
	Cookie string `json:"cookie"`

	// Return to URL
	//
	// The URL to which the browser should be redirected to after the logout
	// has been performed.
	//
	// in: query
	ReturnTo string `json:"return_to"`
}

// swagger:route GET /self-service/logout/browser frontend createBrowserLogoutFlow
//
// # Create a Logout URL for Browsers
//
// This endpoint initializes a browser-based user logout flow and a URL which can be used to log out the user.
//
// This endpoint is NOT INTENDED for API clients and only works
// with browsers (Chrome, Firefox, ...). For API clients you can
// call the `/self-service/logout/api` URL directly with the Ory Session Token.
//
// The URL is only valid for the currently signed in user. If no user is signed in, this endpoint returns
// a 401 error.
//
// When calling this endpoint from a backend, please ensure to properly forward the HTTP cookies.
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Responses:
//	  200: logoutFlow
//	  400: errorGeneric
//	  401: errorGeneric
//	  500: errorGeneric
func (h *Handler) createBrowserLogoutFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	sess, err := h.d.SessionManager().FetchFromRequest(r.Context(), r)
	if err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	conf := h.d.Config()

	requestURL := x.RequestURL(r)

	var returnTo *url.URL

	if requestURL.Query().Get("return_to") != "" {
		// Pre-validate the return to URL which is contained in the HTTP request.
		returnTo, err = x.SecureRedirectTo(r,
			h.d.Config().SelfServiceFlowLogoutRedirectURL(r.Context()),
			x.SecureRedirectUseSourceURL(requestURL.String()),
			x.SecureRedirectAllowURLs(conf.SelfServiceBrowserAllowedReturnToDomains(r.Context())),
			x.SecureRedirectAllowSelfServiceURLs(conf.SelfPublicURL(r.Context())),
		)
		if err != nil {
			h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
			return
		}
	}

	params := url.Values{"token": {sess.LogoutToken}}

	if returnTo != nil {
		params.Set("return_to", returnTo.String())
	}

	h.d.Writer().Write(w, r, &logoutFlow{
		LogoutToken: sess.LogoutToken,
		LogoutURL:   urlx.CopyWithQuery(urlx.AppendPaths(h.d.Config().SelfPublicURL(r.Context()), RouteSubmitFlow), params).String(),
	})
}

// Perform Native Logout Parameters
//
// swagger:parameters performNativeLogout
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type performNativeLogout struct {
	// in: body
	// required: true
	Body performNativeLogoutBody
}

// Perform Native Logout Request Body
//
// swagger:model performNativeLogoutBody
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type performNativeLogoutBody struct {
	// The Session Token
	//
	// Invalidate this session token.
	//
	// required: true
	SessionToken string `json:"session_token"`
}

// swagger:route DELETE /self-service/logout/api frontend performNativeLogout
//
// # Perform Logout for Native Apps
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
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Responses:
//	  204: emptyResponse
//	  400: errorGeneric
//	  default: errorGeneric
func (h *Handler) performNativeLogout(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var p performNativeLogoutBody
	if err := h.dx.Decode(r, &p,
		decoderx.HTTPJSONDecoder(),
		decoderx.HTTPDecoderAllowedMethods("DELETE")); err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}
	sess, err := h.d.SessionPersister().GetSessionByToken(r.Context(), p.SessionToken, session.ExpandNothing, identity.ExpandNothing)
	if err != nil {
		if errors.Is(err, sqlcon.ErrNoRows) {
			h.d.Writer().WriteError(w, r, errors.WithStack(herodot.ErrForbidden.WithReason("The provided Ory Session Token could not be found, is invalid, or otherwise malformed.")))
			return
		}

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

	trace.SpanFromContext(r.Context()).AddEvent(events.NewSessionRevoked(r.Context(), sess.ID, sess.IdentityID))

	w.WriteHeader(http.StatusNoContent)
}

// Update Logout Flow Parameters
//
// swagger:parameters updateLogoutFlow
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type updateLogoutFlow struct {
	// A Valid Logout Token
	//
	// If you do not have a logout token because you only have a session cookie,
	// call `/self-service/logout/browser` to generate a URL for this endpoint.
	//
	// in: query
	Token string `json:"token"`

	// The URL to return to after the logout was completed.
	//
	// in: query
	ReturnTo string `json:"return_to"`

	// HTTP Cookies
	//
	// When using the SDK in a browser app, on the server side you must include the HTTP Cookie Header
	// sent by the client to your server here. This ensures that CSRF and session cookies are respected.
	//
	// in: header
	// name: Cookie
	Cookies string `json:"Cookie"`
}

// swagger:route GET /self-service/logout frontend updateLogoutFlow
//
// # Update Logout Flow
//
// This endpoint logs out an identity in a self-service manner.
//
// If the `Accept` HTTP header is not set to `application/json`, the browser will be redirected (HTTP 303 See Other)
// to the `return_to` parameter of the initial request or fall back to `urls.default_return_to`.
//
// If the `Accept` HTTP header is set to `application/json`, a 204 No Content response
// will be sent on successful logout instead.
//
// This endpoint is NOT INTENDED for API clients and only works
// with browsers (Chrome, Firefox, ...). For API clients you can
// call the `/self-service/logout/api` URL directly with the Ory Session Token.
//
// More information can be found at [Ory Kratos User Logout Documentation](https://www.ory.sh/docs/next/kratos/self-service/flows/user-logout).
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Responses:
//	  303: emptyResponse
//	  204: emptyResponse
//	  default: errorGeneric
func (h *Handler) updateLogoutFlow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	expected := r.URL.Query().Get("token")
	if len(expected) == 0 {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, errors.WithStack(herodot.ErrBadRequest.WithReason("Please include a token in the URL query.")))
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

	if sess.LogoutToken != expected {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, errors.WithStack(herodot.ErrForbidden.WithReason("Unable to log out because the logout token in the URL query does not match the session cookie.")))
		return
	}

	if err := h.d.SessionManager().PurgeFromRequest(r.Context(), w, r); err != nil {
		h.d.SelfServiceErrorManager().Forward(r.Context(), w, r, err)
		return
	}

	trace.SpanFromContext(r.Context()).AddEvent(events.NewSessionRevoked(r.Context(), sess.ID, sess.IdentityID))

	h.completeLogout(w, r)
}

func (h *Handler) completeLogout(w http.ResponseWriter, r *http.Request) {
	_ = h.d.CSRFHandler().RegenerateToken(w, r)

	ret, err := x.SecureRedirectTo(r, h.d.Config().SelfServiceFlowLogoutRedirectURL(r.Context()),
		x.SecureRedirectUseSourceURL(r.RequestURI),
		x.SecureRedirectAllowURLs(h.d.Config().SelfServiceBrowserAllowedReturnToDomains(r.Context())),
		x.SecureRedirectAllowSelfServiceURLs(h.d.Config().SelfPublicURL(r.Context())),
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
}
