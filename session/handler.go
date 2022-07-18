package session

import (
	"net/http"
	"strconv"

	"github.com/ory/x/pointerx"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/x/decoderx"

	"github.com/ory/herodot"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/x"
)

type (
	handlerDependencies interface {
		ManagementProvider
		PersistenceProvider
		x.WriterProvider
		x.LoggingProvider
		x.CSRFProvider
		config.Provider
	}
	HandlerProvider interface {
		SessionHandler() *Handler
	}
	Handler struct {
		r  handlerDependencies
		dx *decoderx.HTTP
	}
)

func NewHandler(
	r handlerDependencies,
) *Handler {
	return &Handler{
		r:  r,
		dx: decoderx.NewHTTP(),
	}
}

const (
	RouteCollection = "/sessions"
	RouteWhoami     = RouteCollection + "/whoami"
	RouteSession    = RouteCollection + "/:id"
)

const (
	AdminRouteIdentity           = "/identities"
	AdminRouteIdentitiesSessions = AdminRouteIdentity + "/:id/sessions"
	AdminRouteSessionExtendId    = RouteSession + "/extend"
)

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(AdminRouteIdentitiesSessions, h.adminListIdentitySessions)
	admin.DELETE(AdminRouteIdentitiesSessions, h.adminDeleteIdentitySessions)
	admin.PATCH(AdminRouteSessionExtendId, h.adminSessionExtend)

	admin.DELETE(RouteCollection, x.RedirectToPublicRoute(h.r))
	admin.DELETE(RouteSession, x.RedirectToPublicRoute(h.r))
	admin.GET(RouteCollection, x.RedirectToPublicRoute(h.r))

	for _, m := range []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut} {
		// Redirect to public endpoint
		admin.Handle(m, RouteWhoami, x.RedirectToPublicRoute(h.r))
	}
}

func (h *Handler) RegisterPublicRoutes(public *x.RouterPublic) {
	// We need to completely ignore the whoami/logout path so that we do not accidentally set
	// some cookie.
	h.r.CSRFHandler().IgnorePath(RouteWhoami)
	h.r.CSRFHandler().IgnorePath(RouteCollection)
	h.r.CSRFHandler().IgnoreGlob(RouteCollection + "/*")
	h.r.CSRFHandler().IgnoreGlob(AdminRouteIdentity + "/*/sessions")

	for _, m := range []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodConnect, http.MethodOptions, http.MethodTrace} {
		public.Handle(m, RouteWhoami, h.whoami)
	}

	public.DELETE(RouteCollection, h.revokeSessions)
	public.DELETE(RouteSession, h.revokeSession)
	public.GET(RouteCollection, h.listSessions)

	public.DELETE(AdminRouteIdentitiesSessions, x.RedirectToAdminRoute(h.r))
}

// nolint:deadcode,unused
// swagger:parameters toSession revokeSessions listSessions
type toSession struct {
	// Set the Session Token when calling from non-browser clients. A session token has a format of `MP2YWEMeM8MxjkGKpH4dqOQ4Q4DlSPaj`.
	//
	// in: header
	SessionToken string `json:"X-Session-Token"`

	// Set the Cookie Header. This is especially useful when calling this endpoint from a server-side application. In that
	// scenario you must include the HTTP Cookie Header which originally was included in the request to your server.
	// An example of a session in the HTTP Cookie Header is: `ory_kratos_session=a19iOVAbdzdgl70Rq1QZmrKmcjDtdsviCTZx7m9a9yHIUS8Wa9T7hvqyGTsLHi6Qifn2WUfpAKx9DWp0SJGleIn9vh2YF4A16id93kXFTgIgmwIOvbVAScyrx7yVl6bPZnCx27ec4WQDtaTewC1CpgudeDV2jQQnSaCP6ny3xa8qLH-QUgYqdQuoA_LF1phxgRCUfIrCLQOkolX5nv3ze_f==`.
	//
	// It is ok if more than one cookie are included here as all other cookies will be ignored.
	//
	// in: header
	Cookie string `json:"Cookie"`
}

// swagger:route GET /sessions/whoami v0alpha2 toSession
//
// Check Who the Current HTTP Session Belongs To
//
// Uses the HTTP Headers in the GET request to determine (e.g. by using checking the cookies) who is authenticated.
// Returns a session object in the body or 401 if the credentials are invalid or no credentials were sent.
// Additionally when the request it successful it adds the user ID to the 'X-Kratos-Authenticated-Identity-Id' header
// in the response.
//
// If you call this endpoint from a server-side application, you must forward the HTTP Cookie Header to this endpoint:
//
//	```js
//	// pseudo-code example
//	router.get('/protected-endpoint', async function (req, res) {
//	  const session = await client.toSession(undefined, req.header('cookie'))
//
//    // console.log(session)
//	})
//	```
//
// When calling this endpoint from a non-browser application (e.g. mobile app) you must include the session token:
//
//	```js
//	// pseudo-code example
//	// ...
//	const session = await client.toSession("the-session-token")
//
//  // console.log(session)
//	```
//
// Depending on your configuration this endpoint might return a 403 status code if the session has a lower Authenticator
// Assurance Level (AAL) than is possible for the identity. This can happen if the identity has password + webauthn
// credentials (which would result in AAL2) but the session has only AAL1. If this error occurs, ask the user
// to sign in with the second factor or change the configuration.
//
// This endpoint is useful for:
//
// - AJAX calls. Remember to send credentials and set up CORS correctly!
// - Reverse proxies and API Gateways
// - Server-side calls - use the `X-Session-Token` header!
//
// This endpoint authenticates users by checking
//
// - if the `Cookie` HTTP header was set containing an Ory Kratos Session Cookie;
// - if the `Authorization: bearer <ory-session-token>` HTTP header was set with a valid Ory Kratos Session Token;
// - if the `X-Session-Token` HTTP header was set with a valid Ory Kratos Session Token.
//
// If none of these headers are set or the cooke or token are invalid, the endpoint returns a HTTP 401 status code.
//
// As explained above, this request may fail due to several reasons. The `error.id` can be one of:
//
// - `session_inactive`: No active session was found in the request (e.g. no Ory Session Cookie / Ory Session Token).
// - `session_aal2_required`: An active session was found but it does not fulfil the Authenticator Assurance Level, implying that the session must (e.g.) authenticate the second factor.
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: session
//       401: jsonError
//       403: jsonError
//       500: jsonError
func (h *Handler) whoami(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	s, err := h.r.SessionManager().FetchFromRequest(r.Context(), r)
	if err != nil {
		h.r.Audit().WithRequest(r).WithError(err).Info("No valid session cookie found.")
		h.r.Writer().WriteError(w, r, herodot.ErrUnauthorized.WithWrap(err).WithReasonf("No valid session cookie found."))
		return
	}

	var aalErr *ErrAALNotSatisfied
	c := h.r.Config(r.Context())
	if err := h.r.SessionManager().DoesSessionSatisfy(r, s, c.SessionWhoAmIAAL()); errors.As(err, &aalErr) {
		h.r.Audit().WithRequest(r).WithError(err).Info("Session was found but AAL is not satisfied for calling this endpoint.")
		h.r.Writer().WriteError(w, r, err)
		return
	} else if err != nil {
		h.r.Audit().WithRequest(r).WithError(err).Info("No valid session cookie found.")
		h.r.Writer().WriteError(w, r, herodot.ErrUnauthorized.WithWrap(err).WithReasonf("Unable to determine AAL."))
		return
	}

	// s.Devices = nil
	s.Identity = s.Identity.CopyWithoutCredentials()

	// Set userId as the X-Kratos-Authenticated-Identity-Id header.
	w.Header().Set("X-Kratos-Authenticated-Identity-Id", s.Identity.ID.String())

	if err := h.r.SessionManager().ReIssueRefreshedCookie(r.Context(), w, r, s); err != nil {
		h.r.Audit().WithRequest(r).WithError(err).Info("Could not re-issue cookie.")
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().Write(w, r, s)
}

// swagger:parameters adminDeleteIdentitySessions
// nolint:deadcode,unused
type adminDeleteIdentitySessions struct {
	// ID is the identity's ID.
	//
	// required: true
	// in: path
	ID string `json:"id"`
}

// swagger:route DELETE /admin/identities/{id}/sessions v0alpha2 adminDeleteIdentitySessions
//
// Calling this endpoint irrecoverably and permanently deletes and invalidates all sessions that belong to the given Identity.
//
// This endpoint is useful for:
//
// - To forcefully logout Identity from all devices and sessions
//
//     Schemes: http, https
//
//     Security:
//       oryAccessToken:
//
//     Responses:
//       204: emptyResponse
//       400: jsonError
//       401: jsonError
//       404: jsonError
//       500: jsonError
func (h *Handler) adminDeleteIdentitySessions(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	iID, err := uuid.FromString(ps.ByName("id"))
	if err != nil {
		h.r.Writer().WriteError(w, r, herodot.ErrBadRequest.WithError(err.Error()).WithDebug("could not parse UUID"))
		return
	}
	if err := h.r.SessionPersister().DeleteSessionsByIdentity(r.Context(), iID); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// swagger:parameters adminListIdentitySessions
// nolint:deadcode,unused
type adminListIdentitySessions struct {
	// Active is a boolean flag that filters out sessions based on the state. If no value is provided, all sessions are returned.
	//
	// required: false
	// in: query
	Active bool `json:"active"`

	adminDeleteIdentitySessions
	x.PaginationParams
}

// swagger:route GET /admin/identities/{id}/sessions v0alpha2 adminListIdentitySessions
//
// This endpoint returns all sessions that belong to the given Identity.
//
// This endpoint is useful for:
//
// - Listing all sessions that belong to an Identity in an administrative context.
//
//     Schemes: http, https
//
//     Security:
//       oryAccessToken:
//
//     Responses:
//       200: sessionList
//       400: jsonError
//       401: jsonError
//       404: jsonError
//       500: jsonError
func (h *Handler) adminListIdentitySessions(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	iID, err := uuid.FromString(ps.ByName("id"))
	if err != nil {
		h.r.Writer().WriteError(w, r, herodot.ErrBadRequest.WithError(err.Error()).WithDebug("could not parse UUID"))
		return
	}

	activeRaw := r.URL.Query().Get("active")
	activeBool, err := strconv.ParseBool(activeRaw)
	if activeRaw != "" && err != nil {
		h.r.Writer().WriteError(w, r, herodot.ErrBadRequest.WithError("could not parse parameter active"))
		return
	}

	var active *bool
	if activeRaw != "" {
		active = &activeBool
	}

	page, perPage := x.ParsePagination(r)
	sess, err := h.r.SessionPersister().ListSessionsByIdentity(r.Context(), iID, active, page, perPage, uuid.Nil)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().Write(w, r, sess)
}

// swagger:model revokedSessions
type revokeSessions struct {
	// The number of sessions that were revoked.
	Count int `json:"count"`
}

// swagger:route DELETE /sessions v0alpha2 revokeSessions
//
// Calling this endpoint invalidates all except the current session that belong to the logged-in user.
// Session data are not deleted.
//
// This endpoint is useful for:
//
// - To forcefully logout the current user from all other devices and sessions
//
//     Schemes: http, https
//
//     Responses:
//       200: revokedSessions
//       400: jsonError
//       401: jsonError
//       404: jsonError
//       500: jsonError
func (h *Handler) revokeSessions(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	s, err := h.r.SessionManager().FetchFromRequest(r.Context(), r)
	if err != nil {
		h.r.Audit().WithRequest(r).WithError(err).Info("No valid session cookie found.")
		h.r.Writer().WriteError(w, r, herodot.ErrUnauthorized.WithWrap(err).WithReasonf("No valid session cookie found."))
		return
	}

	n, err := h.r.SessionPersister().RevokeSessionsIdentityExcept(r.Context(), s.IdentityID, s.ID)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().WriteCode(w, r, http.StatusOK, &revokeSessions{Count: n})
}

// swagger:parameters revokeSession
// nolint:deadcode,unused
type revokeSession struct {
	// ID is the session's ID.
	//
	// required: true
	// in: path
	ID string `json:"id"`
}

// swagger:route DELETE /sessions/{id} v0alpha2 revokeSession
//
// Calling this endpoint invalidates the specified session. The current session cannot be revoked.
// Session data are not deleted.
//
// This endpoint is useful for:
//
// - To forcefully logout the current user from another device or session
//
//     Schemes: http, https
//
//     Responses:
//       204: emptyResponse
//       400: jsonError
//       401: jsonError
//       500: jsonError
func (h *Handler) revokeSession(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	sid := ps.ByName("id")
	if sid == "whoami" {
		// Special case where we actually want to handle the whomai endpoint.
		h.whoami(w, r, ps)
		return
	}

	s, err := h.r.SessionManager().FetchFromRequest(r.Context(), r)
	if err != nil {
		h.r.Audit().WithRequest(r).WithError(err).Info("No valid session cookie found.")
		h.r.Writer().WriteError(w, r, herodot.ErrUnauthorized.WithWrap(err).WithReasonf("No valid session cookie found."))
		return
	}

	sessionID, err := uuid.FromString(sid)
	if err != nil {
		h.r.Writer().WriteError(w, r, herodot.ErrBadRequest.WithError(err.Error()).WithDebug("could not parse UUID"))
		return
	}
	if sessionID == s.ID {
		h.r.Writer().WriteError(w, r, herodot.ErrBadRequest.WithError("cannot revoke current session").WithDebug("use the logout flow instead"))
		return
	}

	if err := h.r.SessionPersister().RevokeSession(r.Context(), s.Identity.ID, sessionID); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().WriteCode(w, r, http.StatusNoContent, nil)
}

// swagger:parameters listSessions
// nolint:deadcode,unused
type listSessions struct {
	x.PaginationParams
}

// swagger:model sessionList
// nolint:deadcode,unused
type sessionList []*Session

// swagger:route GET /sessions v0alpha2 listSessions
//
// This endpoints returns all other active sessions that belong to the logged-in user.
// The current session can be retrieved by calling the `/sessions/whoami` endpoint.
//
// This endpoint is useful for:
//
// - Displaying all other sessions that belong to the logged-in user
//
//     Schemes: http, https
//
//     Responses:
//       200: sessionList
//       400: jsonError
//       401: jsonError
//       404: jsonError
//       500: jsonError
func (h *Handler) listSessions(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	s, err := h.r.SessionManager().FetchFromRequest(r.Context(), r)
	if err != nil {
		h.r.Audit().WithRequest(r).WithError(err).Info("No valid session cookie found.")
		h.r.Writer().WriteError(w, r, herodot.ErrUnauthorized.WithWrap(err).WithReasonf("No valid session cookie found."))
		return
	}

	page, perPage := x.ParsePagination(r)
	sess, err := h.r.SessionPersister().ListSessionsByIdentity(r.Context(), s.IdentityID, pointerx.Bool(true), page, perPage, s.ID)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().Write(w, r, sess)
}

func (h *Handler) IsAuthenticated(wrap httprouter.Handle, onUnauthenticated httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if _, err := h.r.SessionManager().FetchFromRequest(r.Context(), r); err != nil {
			if onUnauthenticated != nil {
				onUnauthenticated(w, r, ps)
				return
			}

			h.r.Writer().WriteError(w, r, errors.WithStack(NewErrNoActiveSessionFound().WithReason("This endpoint can only be accessed with a valid session. Please log in and try again.")))
			return
		}

		wrap(w, r, ps)
	}
}

// swagger:parameters adminExtendSession
// nolint:deadcode,unused
type adminExtendSession struct {
	// ID is the session's ID.
	//
	// required: true
	// in: path
	ID string `json:"id"`
}

// swagger:route PATCH /admin/sessions/{id}/extend v0alpha2 adminExtendSession
//
// Calling this endpoint extends the given session ID. If `session.earliest_possible_extend` is set it
// will only extend the session after the specified time has passed.
//
// Retrieve the session ID from the `/sessions/whoami` endpoint / `toSession` SDK method.
//
//     Schemes: http, https
//
//     Security:
//       oryAccessToken:
//
//     Responses:
//       200: session
//       400: jsonError
//       404: jsonError
//       500: jsonError
func (h *Handler) adminSessionExtend(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	iID, err := uuid.FromString(ps.ByName("id"))
	if err != nil {
		h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithError(err.Error()).WithDebug("could not parse UUID")))
		return
	}

	s, err := h.r.SessionPersister().GetSession(r.Context(), iID)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	c := h.r.Config(r.Context())
	if s.CanBeRefreshed(c) {
		if err := h.r.SessionPersister().UpsertSession(r.Context(), s.Refresh(c)); err != nil {
			h.r.Writer().WriteError(w, r, err)
			return
		}
	}

	h.r.Writer().Write(w, r, s)
}

func (h *Handler) IsNotAuthenticated(wrap httprouter.Handle, onAuthenticated httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if _, err := h.r.SessionManager().FetchFromRequest(r.Context(), r); err != nil {
			if e := new(ErrNoActiveSessionFound); errors.As(err, &e) {
				wrap(w, r, ps)
				return
			}
			h.r.Writer().WriteError(w, r, err)
			return
		}

		if onAuthenticated != nil {
			onAuthenticated(w, r, ps)
			return
		}

		h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrForbidden.WithReason("This endpoint can only be accessed without a login session. Please log out and try again.")))
	}
}

func RedirectOnAuthenticated(d interface{ config.Provider }) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		returnTo, err := x.SecureRedirectTo(r, d.Config(r.Context()).SelfServiceBrowserDefaultReturnTo(), x.SecureRedirectAllowSelfServiceURLs(d.Config(r.Context()).SelfPublicURL()))
		if err != nil {
			http.Redirect(w, r, d.Config(r.Context()).SelfServiceBrowserDefaultReturnTo().String(), http.StatusFound)
			return
		}

		http.Redirect(w, r, returnTo.String(), http.StatusFound)
	}
}

func RedirectOnUnauthenticated(to string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		http.Redirect(w, r, to, http.StatusFound)
	}
}

func RespondWithJSONErrorOnAuthenticated(h herodot.Writer, err error) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		h.WriteError(w, r, err)
	}
}
