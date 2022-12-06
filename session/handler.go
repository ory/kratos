// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ory/x/pagination/migrationpagination"

	"github.com/ory/x/pagination/keysetpagination"

	"github.com/ory/x/pointerx"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/x/decoderx"
	"github.com/ory/x/urlx"

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
	admin.GET(RouteCollection, h.adminListSessions)
	admin.GET(RouteSession, h.getSession)
	admin.DELETE(RouteSession, h.disableSession)

	admin.GET(AdminRouteIdentitiesSessions, h.listIdentitySessions)
	admin.DELETE(AdminRouteIdentitiesSessions, h.deleteIdentitySessions)
	admin.PATCH(AdminRouteSessionExtendId, h.adminSessionExtend)

	admin.DELETE(RouteCollection, x.RedirectToPublicRoute(h.r))
}

func (h *Handler) RegisterPublicRoutes(public *x.RouterPublic) {
	// We need to completely ignore the whoami/logout path so that we do not accidentally set
	// some cookie.
	h.r.CSRFHandler().IgnorePath(RouteWhoami)
	h.r.CSRFHandler().IgnorePath(RouteCollection)
	h.r.CSRFHandler().IgnoreGlob(RouteCollection + "/*")
	h.r.CSRFHandler().IgnoreGlob(RouteCollection + "/*/extend")
	h.r.CSRFHandler().IgnoreGlob(AdminRouteIdentity + "/*/sessions")

	for _, m := range []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodConnect, http.MethodOptions, http.MethodTrace} {
		public.Handle(m, RouteWhoami, h.whoami)
	}

	public.DELETE(RouteCollection, h.deleteMySessions)
	public.DELETE(RouteSession, h.deleteMySession)
	public.GET(RouteCollection, h.listMySessions)

	public.DELETE(AdminRouteIdentitiesSessions, x.RedirectToAdminRoute(h.r))
}

// Check Session Request Parameters
//
// nolint:deadcode,unused
// swagger:parameters toSession
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

// swagger:route GET /sessions/whoami frontend toSession
//
// # Check Who the Current HTTP Session Belongs To
//
// Uses the HTTP Headers in the GET request to determine (e.g. by using checking the cookies) who is authenticated.
// Returns a session object in the body or 401 if the credentials are invalid or no credentials were sent.
// When the request it successful it adds the user ID to the 'X-Kratos-Authenticated-Identity-Id' header
// in the response.
//
// If you call this endpoint from a server-side application, you must forward the HTTP Cookie Header to this endpoint:
//
//	```js
//	// pseudo-code example
//	router.get('/protected-endpoint', async function (req, res) {
//	  const session = await client.toSession(undefined, req.header('cookie'))
//
//	  // console.log(session)
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
//	// console.log(session)
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
// This endpoint authenticates users by checking:
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
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Responses:
//	  200: session
//	  401: errorGeneric
//	  403: errorGeneric
//	  default: errorGeneric
func (h *Handler) whoami(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	s, err := h.r.SessionManager().FetchFromRequest(r.Context(), r)
	c := h.r.Config()
	if err != nil {
		// We cache errors (and set cache header only when configured) where no session was found.
		if noSess := new(ErrNoActiveSessionFound); c.SessionWhoAmICaching(r.Context()) && errors.As(err, &noSess) && noSess.credentialsMissing {
			w.Header().Set("Ory-Session-Cache-For", fmt.Sprintf("%d", int64(time.Minute.Seconds())))
		}

		h.r.Audit().WithRequest(r).WithError(err).Info("No valid session cookie found.")
		h.r.Writer().WriteError(w, r, herodot.ErrUnauthorized.WithWrap(err).WithReasonf("No valid session cookie found."))
		return
	}

	var aalErr *ErrAALNotSatisfied
	if err := h.r.SessionManager().DoesSessionSatisfy(r, s, c.SessionWhoAmIAAL(r.Context())); errors.As(err, &aalErr) {
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

	// Set Cache header only when configured
	if c.SessionWhoAmICaching(r.Context()) {
		w.Header().Set("Ory-Session-Cache-For", fmt.Sprintf("%d", int64(time.Until(s.ExpiresAt).Seconds())))
	}

	if err := h.r.SessionManager().RefreshCookie(r.Context(), w, r, s); err != nil {
		h.r.Audit().WithRequest(r).WithError(err).Info("Could not re-issue cookie.")
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().Write(w, r, s)
}

// Delete Identity Session Parameters
//
// swagger:parameters deleteIdentitySessions
// nolint:deadcode,unused
type deleteIdentitySessions struct {
	// ID is the identity's ID.
	//
	// required: true
	// in: path
	ID string `json:"id"`
}

// swagger:route DELETE /admin/identities/{id}/sessions identity deleteIdentitySessions
//
// # Delete & Invalidate an Identity's Sessions
//
// Calling this endpoint irrecoverably and permanently deletes and invalidates all sessions that belong to the given Identity.
//
//	Schemes: http, https
//
//	Security:
//	  oryAccessToken:
//
//	Responses:
//	  204: emptyResponse
//	  400: errorGeneric
//	  401: errorGeneric
//	  404: errorGeneric
//	  default: errorGeneric
func (h *Handler) deleteIdentitySessions(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

// Session List Request
//
// The request object for listing sessions in an administrative context.
//
// swagger:parameters listSessions
// nolint:deadcode,unused
type listSessionsRequest struct {
	keysetpagination.RequestParameters

	// Active is a boolean flag that filters out sessions based on the state. If no value is provided, all sessions are returned.
	//
	// required: false
	// in: query
	Active bool `json:"active"`

	// ExpandOptions is a query parameter encoded list of all properties that must be expanded in the Session.
	// Example - ?expand=Identity&expand=Devices
	// If no value is provided, the expandable properties are skipped.
	//
	// required: false
	// in: query
	ExpandOptions []Expandable `json:"expand"`
}

// Session List Response
//
// The response given when listing sessions in an administrative context.
//
// swagger:response listSessions
// nolint:deadcode,unused
type listSessionsResponse struct {
	keysetpagination.ResponseHeaders

	// The list of sessions found
	// in: body
	Sessions []Session
}

// swagger:route GET /admin/sessions identity listSessions
//
// # List All Sessions
//
// Listing all sessions that exist.
//
//	Schemes: http, https
//
//	Security:
//	  oryAccessToken:
//
//	Responses:
//	  200: listSessions
//	  400: errorGeneric
//	  default: errorGeneric
func (h *Handler) adminListSessions(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	// Parse request pagination parameters
	opts, err := keysetpagination.Parse(r.URL.Query(), keysetpagination.NewStringPageToken)
	if err != nil {
		h.r.Writer().WriteError(w, r, herodot.ErrBadRequest.WithError("could not parse parameter page_size"))
		return
	}

	var expandables Expandables
	if es, ok := r.URL.Query()["expand"]; ok {
		for _, e := range es {
			expand, ok := ParseExpandable(e)
			if !ok {
				h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Could not parse expand option: %s", e)))
				return
			}
			expandables = append(expandables, expand)
		}
	}

	sess, total, nextPage, err := h.r.SessionPersister().ListSessions(r.Context(), active, opts, expandables)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	w.Header().Set("x-total-count", fmt.Sprint(total))
	keysetpagination.Header(w, r.URL, nextPage)
	h.r.Writer().Write(w, r, sess)
}

// Session Get Request
//
// The request object for getting a session in an administrative context.
//
// swagger:parameters getSession
// nolint:deadcode,unused
type getSession struct {
	// ExpandOptions is a query parameter encoded list of all properties that must be expanded in the Session.
	// Example - ?expand=Identity&expand=Devices
	// If no value is provided, the expandable properties are skipped.
	//
	// required: false
	// in: query
	ExpandOptions []Expandable `json:"expand"`

	// ID is the session's ID.
	//
	// required: true
	// in: path
	ID string `json:"id"`
}

// swagger:route GET /admin/sessions/{id} identity getSession
//
// This endpoint returns the session object with expandables specified.
//
// This endpoint is useful for:
//
// - Getting a session object with all specified expandables that exist in an administrative context.
//
//	Schemes: http, https
//
//	Security:
//	  oryAccessToken:
//
//	Responses:
//	  200: session
//	  400: errorGeneric
//	  default: errorGeneric
func (h *Handler) getSession(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if ps.ByName("id") == "whoami" {
		// for /admin/sessions/whoami redirect to the public route
		x.RedirectToPublicRoute(h.r)(w, r, ps)
		return
	}

	sID, err := uuid.FromString(ps.ByName("id"))
	if err != nil {
		h.r.Writer().WriteError(w, r, herodot.ErrBadRequest.WithError(err.Error()).WithDebug("could not parse UUID"))
		return
	}

	var expandables Expandables

	urlValues := r.URL.Query()
	if es, ok := urlValues["expand"]; ok {
		for _, e := range es {
			expand, ok := ParseExpandable(e)
			if !ok {
				h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Could not parse expand option: %s", e)))
				return
			}
			expandables = append(expandables, expand)
		}
	}

	sess, err := h.r.SessionPersister().GetSession(r.Context(), sID, expandables)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().Write(w, r, sess)
}

// List Identity Sessions Parameters
//
// swagger:parameters listIdentitySessions
// Deactivate Session Parameters
//
// swagger:parameters disableSession
// nolint:deadcode,unused
type disableSession struct {
	// ID is the session's ID.
	//
	// required: true
	// in: path
	ID string `json:"id"`
}

// swagger:route DELETE /admin/sessions/{id} identity disableSession
//
// # Deactivate a Session
//
// Calling this endpoint deactivates the specified session. Session data is not deleted.
//
//	Schemes: http, https
//
//	Responses:
//	  204: emptyResponse
//	  400: errorGeneric
//	  401: errorGeneric
//	  default: errorGeneric
func (h *Handler) disableSession(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	sID, err := uuid.FromString(ps.ByName("id"))
	if err != nil {
		h.r.Writer().WriteError(w, r, herodot.ErrBadRequest.WithError(err.Error()).WithDebug("could not parse UUID"))
		return
	}

	if err := h.r.SessionPersister().RevokeSessionById(r.Context(), sID); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().WriteCode(w, r, http.StatusNoContent, nil)
}

// List Identity Sessions Parameters
//
// swagger:parameters listIdentitySessions
// nolint:deadcode,unused
type listIdentitySessionsRequest struct {
	migrationpagination.RequestParameters

	// ID is the identity's ID.
	//
	// required: true
	// in: path
	ID string `json:"id"`

	// Active is a boolean flag that filters out sessions based on the state. If no value is provided, all sessions are returned.
	//
	// required: false
	// in: query
	Active bool `json:"active"`
}

// List Identity Sessions Response
//
// swagger:response listIdentitySessions
// nolint:deadcode,unused
type listIdentitySessionsResponse struct {
	migrationpagination.ResponseHeaderAnnotation

	// in: body
	Body []Session
}

// swagger:route GET /admin/identities/{id}/sessions identity listIdentitySessions
//
// # List an Identity's Sessions
//
// This endpoint returns all sessions that belong to the given Identity.
//
//	Schemes: http, https
//
//	Security:
//	  oryAccessToken:
//
//	Responses:
//	  200: listIdentitySessions
//	  400: errorGeneric
//	  404: errorGeneric
//	  default: errorGeneric
func (h *Handler) listIdentitySessions(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	iID, err := uuid.FromString(ps.ByName("id"))
	if err != nil {
		h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithError(err.Error()).WithDebug("could not parse UUID")))
		return
	}

	activeRaw := r.URL.Query().Get("active")
	activeBool, err := strconv.ParseBool(activeRaw)
	if activeRaw != "" && err != nil {
		h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithError("could not parse parameter active")))
		return
	}

	var active *bool
	if activeRaw != "" {
		active = &activeBool
	}

	page, perPage := x.ParsePagination(r)
	sess, total, err := h.r.SessionPersister().ListSessionsByIdentity(r.Context(), iID, active, page, perPage, uuid.Nil, ExpandEverything)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	x.PaginationHeader(w, urlx.AppendPaths(h.r.Config().SelfAdminURL(r.Context()), RouteCollection), total, page, perPage)
	h.r.Writer().Write(w, r, sess)
}

// Deleted Session Count
//
// swagger:model deleteMySessionsCount
type deleteMySessionsCount struct {
	// The number of sessions that were revoked.
	Count int `json:"count"`
}

// Disable My Other Session Parameters
//
// nolint:deadcode,unused
// swagger:parameters disableMyOtherSessions
type disableMyOtherSessions struct {
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

// swagger:route DELETE /sessions frontend disableMyOtherSessions
//
// # Disable my other sessions
//
// Calling this endpoint invalidates all except the current session that belong to the logged-in user.
// Session data are not deleted.
//
//	Schemes: http, https
//
//	Responses:
//	  200: deleteMySessionsCount
//	  400: errorGeneric
//	  401: errorGeneric
//	  default: errorGeneric
func (h *Handler) deleteMySessions(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

	h.r.Writer().WriteCode(w, r, http.StatusOK, &deleteMySessionsCount{Count: n})
}

// Disable My Session Parameters
//
// swagger:parameters disableMySession
// nolint:deadcode,unused
type disableMySession struct {
	// ID is the session's ID.
	//
	// required: true
	// in: path
	ID string `json:"id"`

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

// swagger:route DELETE /sessions/{id} frontend disableMySession
//
// # Disable one of my sessions
//
// Calling this endpoint invalidates the specified session. The current session cannot be revoked.
// Session data are not deleted.
//
//	Schemes: http, https
//
//	Responses:
//	  204: emptyResponse
//	  400: errorGeneric
//	  401: errorGeneric
//	  default: errorGeneric
func (h *Handler) deleteMySession(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

// List My Session Parameters
//
// swagger:parameters listMySessions
// nolint:deadcode,unused
type listMySessionsParameters struct {
	x.PaginationParams

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

// List My Session Response
//
// swagger:response listMySessions
// nolint:deadcode,unused
type listMySessionsResponse struct {
	migrationpagination.ResponseHeaderAnnotation

	// in: body
	Body []Session
}

// swagger:route GET /sessions frontend listMySessions
//
// # Get My Active Sessions
//
// This endpoints returns all other active sessions that belong to the logged-in user.
// The current session can be retrieved by calling the `/sessions/whoami` endpoint.
//
//	Schemes: http, https
//
//	Responses:
//	  200: listMySessions
//	  400: errorGeneric
//	  401: errorGeneric
//	  default: errorGeneric
func (h *Handler) listMySessions(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	s, err := h.r.SessionManager().FetchFromRequest(r.Context(), r)
	if err != nil {
		h.r.Audit().WithRequest(r).WithError(err).Info("No valid session cookie found.")
		h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrUnauthorized.WithWrap(err).WithReasonf("No valid session cookie found.")))
		return
	}

	page, perPage := x.ParsePagination(r)
	sess, total, err := h.r.SessionPersister().ListSessionsByIdentity(r.Context(), s.IdentityID, pointerx.Bool(true), page, perPage, s.ID, ExpandEverything)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	x.PaginationHeader(w, urlx.AppendPaths(h.r.Config().SelfAdminURL(r.Context()), RouteCollection), total, page, perPage)
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

// swagger:parameters extendSession
// nolint:deadcode,unused
type extendSession struct {
	// ID is the session's ID.
	//
	// required: true
	// in: path
	ID string `json:"id"`
}

// swagger:route PATCH /admin/sessions/{id}/extend identity extendSession
//
// # Extend a Session
//
// Calling this endpoint extends the given session ID. If `session.earliest_possible_extend` is set it
// will only extend the session after the specified time has passed.
//
// Retrieve the session ID from the `/sessions/whoami` endpoint / `toSession` SDK method.
//
//	Schemes: http, https
//
//	Security:
//	  oryAccessToken:
//
//	Responses:
//	  200: session
//	  400: errorGeneric
//	  404: errorGeneric
//	  default: errorGeneric
func (h *Handler) adminSessionExtend(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	iID, err := uuid.FromString(ps.ByName("id"))
	if err != nil {
		h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithError(err.Error()).WithDebug("could not parse UUID")))
		return
	}

	s, err := h.r.SessionPersister().GetSession(r.Context(), iID, ExpandNothing)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	c := h.r.Config()
	if s.CanBeRefreshed(r.Context(), c) {
		if err := h.r.SessionPersister().UpsertSession(r.Context(), s.Refresh(r.Context(), c)); err != nil {
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
		ctx := r.Context()
		returnTo, err := x.SecureRedirectTo(r, d.Config().SelfServiceBrowserDefaultReturnTo(ctx), x.SecureRedirectAllowSelfServiceURLs(d.Config().SelfPublicURL(ctx)))
		if err != nil {
			http.Redirect(w, r, d.Config().SelfServiceBrowserDefaultReturnTo(ctx).String(), http.StatusFound)
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

func RespondWitherrorGenericOnAuthenticated(h herodot.Writer, err error) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		h.WriteError(w, r, err)
	}
}
