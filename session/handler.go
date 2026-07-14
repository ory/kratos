// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/trace"

	"github.com/ory/x/httpx"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/otelx"

	"github.com/ory/kratos/x/nosurfx"
	"github.com/ory/kratos/x/redir"
	"github.com/ory/x/httprouterx"

	"github.com/ory/kratos/selfservice/sessiontokenexchange"
	"github.com/ory/x/otelx/semconv"
	"github.com/ory/x/pagination/migrationpagination"

	"github.com/ory/x/pagination/keysetpagination"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"
)

type (
	// FlowForTokenExchange looks up a login or registration flow by ID.
	// It is used by the token exchange handler to surface flow errors when no
	// session was created (e.g. because a before-registration webhook rejected
	// the request).
	FlowForTokenExchange interface {
		GetFlowForTokenExchange(ctx context.Context, flowID uuid.UUID) (any, error)
	}
	FlowForTokenExchangeProvider interface {
		FlowForTokenExchange() FlowForTokenExchange
	}

	handlerDependencies interface {
		ManagementProvider
		PersistenceProvider
		httpx.WriterProvider
		otelx.Provider
		logrusx.Provider
		nosurfx.CSRFProvider
		config.Provider
		sessiontokenexchange.PersistenceProvider
		FlowForTokenExchangeProvider
		TokenizerProvider
	}
	HandlerProvider interface {
		SessionHandler() *Handler
	}
	Handler struct{ r handlerDependencies }
)

func NewHandler(r handlerDependencies) *Handler { return &Handler{r: r} }

const (
	RouteCollection                  = "/sessions"
	RouteExchangeCodeForSessionToken = RouteCollection + "/token-exchange" // #nosec G101
	RouteWhoami                      = RouteCollection + "/whoami"
	RouteSession                     = RouteCollection + "/{id}"
)

const (
	AdminRouteIdentity           = "/identities"
	AdminRouteIdentitiesSessions = AdminRouteIdentity + "/{id}/sessions"
	AdminRouteSessionExtendId    = RouteSession + "/extend"

	// ManageSessionsMaxIDs caps the number of explicit IDs accepted per call.
	// Picked defensively — not validated against a production dataset.
	// Tied to MaxManageSessionsBodySize in cloudlib/ratelimit/inflight: 500 UUIDs
	// serialise to ~20 KiB, so 32 KiB has ~40% headroom. Raise both together.
	ManageSessionsMaxIDs = 500

	// manageSessionsWildcardBatchSize is the per-statement chunk size for the
	// wildcard ("*") variants. One HTTP call processes at most this many rows;
	// the caller drains larger networks by re-issuing the request while the
	// response reports `more: true`.
	manageSessionsWildcardBatchSize = 5000
)

func (h *Handler) RegisterAdminRoutes(admin *httprouterx.RouterAdmin) {
	admin.GET(RouteCollection, h.adminListSessions)
	admin.GET(RouteSession, h.getSession)
	admin.DELETE(RouteSession, h.disableSession)

	admin.GET(AdminRouteIdentitiesSessions, h.listIdentitySessions)
	admin.DELETE(AdminRouteIdentitiesSessions, h.deleteIdentitySessions)
	admin.PATCH(AdminRouteSessionExtendId, h.adminSessionExtend)
	admin.POST(RouteCollection, h.manageSessions)

	admin.DELETE(RouteCollection, redir.RedirectToPublicRoute(h.r))
}

func (h *Handler) RegisterPublicRoutes(public *httprouterx.RouterPublic) {
	// We need to completely ignore the whoami/logout path so that we do not accidentally set
	// some cookie.
	h.r.CSRFHandler().IgnorePath(RouteWhoami)
	h.r.CSRFHandler().IgnorePath(RouteCollection)
	h.r.CSRFHandler().IgnoreGlob(RouteCollection + "/*")
	h.r.CSRFHandler().IgnoreGlob(RouteCollection + "/*/extend")
	h.r.CSRFHandler().IgnoreGlob(AdminRouteIdentity + "/*/sessions")

	for _, m := range []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodConnect, http.MethodOptions, http.MethodTrace} {
		public.Handle(m+" "+RouteWhoami, http.HandlerFunc(h.whoami))
	}

	public.DELETE(RouteCollection, h.deleteMySessions)
	public.DELETE(RouteSession, h.deleteMySession)
	public.GET(RouteCollection, h.listMySessions)

	public.GET(RouteExchangeCodeForSessionToken, h.exchangeCode)

	public.DELETE(AdminRouteIdentitiesSessions, redir.RedirectToAdminRoute(h.r))
}

// Check Session Request Parameters
//
// swagger:parameters toSession
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
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

	// Returns the session additionally as a token (such as a JWT)
	//
	// The value of this parameter has to be a valid, configured Ory Session token template. For more information head over to [the documentation](http://ory.sh/docs/identities/session-to-jwt-cors).
	//
	// in: query
	TokenizeAs string `json:"tokenize_as"`
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
// When using a token template, the token is included in the `tokenized` field of the session.
//
//	```js
//	// pseudo-code example
//	// ...
//	const session = await client.toSession("the-session-token", { tokenize_as: "example-jwt-template" })
//
//	console.log(session.tokenized) // The JWT
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
// If none of these headers are set or the cookie or token are invalid, the endpoint returns a HTTP 401 status code.
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
//
//	Extensions:
//	  x-ory-ratelimit-bucket: kratos-public-high
func (h *Handler) whoami(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.r.Tracer(r.Context()).Tracer().Start(r.Context(), "sessions.Handler.whoami")
	defer span.End()

	// Whoami strips the credentials from the response below (CopyWithoutCredentials), so there is
	// no point in loading them from the database: it is by far the most expensive part of fetching
	// a session.
	s, err := h.r.SessionManager().FetchFromRequest(ctx, r, ExpandEverything, identity.ExpandEverythingButCredentials)
	c := h.r.Config()
	if err != nil {
		// We cache errors (and set cache header only when configured) where no session was found.
		if noSess := new(ErrNoActiveSessionFound); c.SessionWhoAmICaching(ctx) && errors.As(err, &noSess) && noSess.CredentialsMissing {
			w.Header().Set("Ory-Session-Cache-For", fmt.Sprintf("%d", int64(time.Minute.Seconds())))
		}

		h.r.Logger().WithRequest(r).WithError(err).Info("No valid session found.")
		h.r.Writer().WriteError(w, r, ErrNoSessionFound().WithWrap(err))
		return
	}

	var aalErr *ErrAALNotSatisfied
	if err := h.r.SessionManager().DoesSessionSatisfy(ctx, s, c.SessionWhoAmIAAL(ctx),
		// For the time being we want to update the AAL in the database if it is unset.
		UpsertAAL,
	); errors.As(err, &aalErr) {
		h.r.Logger().WithRequest(r).WithError(err).Info("Session was found but AAL is not satisfied for calling this endpoint.")
		h.r.Writer().WriteError(w, r, err)
		return
	} else if err != nil {
		h.r.Logger().WithRequest(r).WithError(err).Info("No valid session cookie found.")
		h.r.Writer().WriteError(w, r, herodot.ErrUnauthorized().WithWrap(err).WithReasonf("Unable to determine AAL."))
		return
	}

	// s.Devices = nil
	s.Identity = s.Identity.CopyWithoutCredentials()

	tokenizeTemplate := r.URL.Query().Get("tokenize_as")
	if tokenizeTemplate != "" {
		if err := h.r.SessionTokenizer().TokenizeSession(ctx, tokenizeTemplate, s); err != nil {
			h.r.Writer().WriteError(w, r, err)
			return
		}
	}

	// Set userId as the X-Kratos-Authenticated-Identity-Id header.
	w.Header().Set("X-Kratos-Authenticated-Identity-Id", s.Identity.ID.String())

	// Set Cache header only when configured, and when no tokenization is requested.
	if c.SessionWhoAmICaching(ctx) && len(tokenizeTemplate) == 0 {
		expiry := time.Until(s.ExpiresAt)
		if c.SessionWhoAmICachingMaxAge(ctx) > 0 && expiry > c.SessionWhoAmICachingMaxAge(ctx) {
			expiry = c.SessionWhoAmICachingMaxAge(ctx)
		}

		w.Header().Set("Ory-Session-Cache-For", fmt.Sprintf("%0.f", expiry.Seconds()))
	}

	if err := h.r.SessionManager().RefreshCookie(ctx, w, r, s); err != nil {
		h.r.Logger().WithRequest(r).WithError(err).Info("Could not re-issue cookie.")
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().Write(w, r, s)
}

// Delete Identity Session Parameters
//
// swagger:parameters deleteIdentitySessions
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
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
//
//	Extensions:
//	  x-ory-ratelimit-bucket: kratos-admin-low
func (h *Handler) deleteIdentitySessions(w http.ResponseWriter, r *http.Request) {
	iID, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		h.r.Writer().WriteError(w, r, herodot.ErrBadRequest().WithError(err.Error()).WithDebug("could not parse UUID"))
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
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type listSessionsRequest struct {
	keysetpagination.RequestParameters

	// Active is a boolean flag that filters out sessions based on the state. If no value is provided, all sessions are returned.
	//
	// required: false
	// in: query
	Active bool `json:"active"`

	// ExpandOptions is a query parameter encoded list of all properties that must be expanded in the Session.
	// If no value is provided, the expandable properties are skipped.
	//
	// required: false
	// in: query
	ExpandOptions []SessionExpandable `json:"expand"`
}

// Expandable properties of a session
// swagger:enum SessionExpandable
type SessionExpandable string

const (
	SessionExpandableIdentity SessionExpandable = "identity"
	SessionExpandableDevices  SessionExpandable = "devices"
)

// Session List Response
//
// The response given when listing sessions in an administrative context.
//
// swagger:response listSessions
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
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
//
//	Extensions:
//	  x-ory-ratelimit-bucket: kratos-admin-medium
func (h *Handler) adminListSessions(w http.ResponseWriter, r *http.Request) {
	activeRaw := r.URL.Query().Get("active")
	activeBool, err := strconv.ParseBool(activeRaw)
	if activeRaw != "" && err != nil {
		h.r.Writer().WriteError(w, r, herodot.ErrBadRequest().WithError("could not parse parameter active"))
		return
	}

	var active *bool
	if activeRaw != "" {
		active = &activeBool
	}

	// Parse request pagination parameters
	opts, err := keysetpagination.Parse(r.URL.Query(), keysetpagination.NewMapPageToken)
	if err != nil {
		h.r.Writer().WriteError(w, r, herodot.ErrBadRequest().WithError("could not parse parameter page_size"))
		return
	}

	var expandables Expandables
	if es, ok := r.URL.Query()["expand"]; ok {
		for _, e := range es {
			expand, ok := ParseExpandable(e)
			if !ok {
				h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest().WithReasonf("Could not parse expand option: %s", e)))
				return
			}
			expandables = append(expandables, expand)
		}
	}

	sess, nextPage, err := h.r.SessionPersister().ListSessions(r.Context(), active, opts, expandables)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	u := *r.URL
	keysetpagination.Header(w, &u, nextPage)
	h.r.Writer().Write(w, r, AdminSessions(sess))
}

// Session Get Request
//
// The request object for getting a session in an administrative context.
//
// swagger:parameters getSession
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type getSession struct {
	// ExpandOptions is a query parameter encoded list of all properties that must be expanded in the Session.
	// Example - ?expand=Identity&expand=Devices
	// If no value is provided, the expandable properties are skipped.
	//
	// required: false
	// in: query
	ExpandOptions []SessionExpandable `json:"expand"`

	// ID is the session's ID.
	//
	// required: true
	// in: path
	ID string `json:"id"`
}

// swagger:route GET /admin/sessions/{id} identity getSession
//
// # Get Session
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
//
//	Extensions:
//	  x-ory-ratelimit-bucket: kratos-admin-high
func (h *Handler) getSession(w http.ResponseWriter, r *http.Request) {
	if r.PathValue("id") == "whoami" {
		// for /admin/sessions/whoami redirect to the public route
		redir.RedirectToPublicRoute(h.r)(w, r)
		return
	}

	sID, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		h.r.Writer().WriteError(w, r, herodot.ErrBadRequest().WithError(err.Error()).WithDebug("could not parse UUID"))
		return
	}

	var expandables Expandables

	urlValues := r.URL.Query()
	if es, ok := urlValues["expand"]; ok {
		for _, e := range es {
			expand, ok := ParseExpandable(e)
			if !ok {
				h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest().WithReasonf("Could not parse expand option: %s", e)))
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

	h.r.Writer().Write(w, r, AdminSession(*sess))
}

// List Identity Sessions Parameters
//
// swagger:parameters listIdentitySessions
// Deactivate Session Parameters
//
// swagger:parameters disableSession
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
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
//	Security:
//		oryAccessToken:
//
//	Responses:
//		204: emptyResponse
//		400: errorGeneric
//		401: errorGeneric
//		default: errorGeneric
//
//	Extensions:
//	  x-ory-ratelimit-bucket: kratos-admin-low
func (h *Handler) disableSession(w http.ResponseWriter, r *http.Request) {
	sID, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		h.r.Writer().WriteError(w, r, herodot.ErrBadRequest().WithError(err.Error()).WithDebug("could not parse UUID"))
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
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
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
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
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
//
//	Extensions:
//	  x-ory-ratelimit-bucket: kratos-admin-medium
func (h *Handler) listIdentitySessions(w http.ResponseWriter, r *http.Request) {
	iID, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest().WithError(err.Error()).WithDebug("could not parse UUID")))
		return
	}

	activeRaw := r.URL.Query().Get("active")
	activeBool, err := strconv.ParseBool(activeRaw)
	if activeRaw != "" && err != nil {
		h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest().WithError("could not parse parameter active")))
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

	x.PaginationHeader(w, *r.URL, total, page, perPage)
	h.r.Writer().Write(w, r, AdminSessions(sess))
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
// swagger:parameters disableMyOtherSessions
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
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
//
//	Extensions:
//	  x-ory-ratelimit-bucket: kratos-public-low
func (h *Handler) deleteMySessions(w http.ResponseWriter, r *http.Request) {
	// Only session columns are read below, so skip loading the devices and the identity.
	s, err := h.r.SessionManager().FetchFromRequest(r.Context(), r, ExpandNothing, identity.ExpandNothing)
	if err != nil {
		h.r.Logger().WithRequest(r).WithError(err).Info("No valid session cookie found.")
		h.r.Writer().WriteError(w, r, herodot.ErrUnauthorized().WithWrap(err).WithReasonf("No valid session cookie found."))
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
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
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
//
//	Extensions:
//	  x-ory-ratelimit-bucket: kratos-public-low
func (h *Handler) deleteMySession(w http.ResponseWriter, r *http.Request) {
	sid := r.PathValue("id")
	if sid == "whoami" {
		// Special case where we actually want to handle the whoami endpoint.
		h.whoami(w, r)
		return
	}

	// Only session columns are read below, so skip loading the devices and the identity.
	s, err := h.r.SessionManager().FetchFromRequest(r.Context(), r, ExpandNothing, identity.ExpandNothing)
	if err != nil {
		h.r.Logger().WithRequest(r).WithError(err).Info("No valid session cookie found.")
		h.r.Writer().WriteError(w, r, herodot.ErrUnauthorized().WithWrap(err).WithReasonf("No valid session cookie found."))
		return
	}

	sessionID, err := uuid.FromString(sid)
	if err != nil {
		h.r.Writer().WriteError(w, r, herodot.ErrBadRequest().WithError(err.Error()).WithDebug("could not parse UUID"))
		return
	}
	if sessionID == s.ID {
		h.r.Writer().WriteError(w, r, herodot.ErrBadRequest().WithError("cannot revoke current session").WithDebug("use the logout flow instead"))
		return
	}

	if err := h.r.SessionPersister().RevokeSession(r.Context(), s.IdentityID, sessionID); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().WriteCode(w, r, http.StatusNoContent, nil)
}

// List My Session Parameters
//
// swagger:parameters listMySessions
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type listMySessionsParameters struct {
	migrationpagination.RequestParameters

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
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
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
//
//	Extensions:
//	  x-ory-ratelimit-bucket: kratos-public-medium
func (h *Handler) listMySessions(w http.ResponseWriter, r *http.Request) {
	// Only session columns are read below; DoesSessionSatisfy loads the identity lazily when the
	// AAL check needs it.
	s, err := h.r.SessionManager().FetchFromRequest(r.Context(), r, ExpandNothing, identity.ExpandNothing)
	if err != nil {
		h.r.Logger().WithRequest(r).WithError(err).Info("No valid session cookie found.")
		h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrUnauthorized().WithWrap(err).WithReasonf("No valid session cookie found.")))
		return
	}

	c := h.r.Config()

	var aalErr *ErrAALNotSatisfied
	if err := h.r.SessionManager().DoesSessionSatisfy(r.Context(), s, c.SessionWhoAmIAAL(r.Context())); errors.As(err, &aalErr) {
		h.r.Logger().WithRequest(r).WithError(err).Info("Session was found but AAL is not satisfied for calling this endpoint.")
		h.r.Writer().WriteError(w, r, err)
		return
	} else if err != nil {
		h.r.Logger().WithRequest(r).WithError(err).Info("No valid session cookie found.")
		h.r.Writer().WriteError(w, r, herodot.ErrUnauthorized().WithWrap(err).WithReasonf("Unable to determine AAL."))
		return
	}

	page, perPage := x.ParsePagination(r)
	sess, total, err := h.r.SessionPersister().ListSessionsByIdentity(r.Context(), s.IdentityID, new(true), page, perPage, s.ID, ExpandEverything)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	x.PaginationHeader(w, *r.URL, total, page, perPage)
	h.r.Writer().Write(w, r, sess)
}

type sessionInContext int

const (
	sessionInContextKey sessionInContext = iota
)

func (h *Handler) IsAuthenticated(wrap http.HandlerFunc, onUnauthenticated http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// The session is stored in the request context for downstream handlers, which may read
		// any of its associations - keep the full expansion.
		sess, err := h.r.SessionManager().FetchFromRequest(ctx, r, ExpandEverything, identity.ExpandEverything)
		if err != nil {
			if onUnauthenticated != nil {
				onUnauthenticated(w, r)
				return
			}

			h.r.Writer().WriteError(w, r, errors.WithStack(NewErrNoActiveSessionFound().WithReason("This endpoint can only be accessed with a valid session. Please log in and try again.")))
			return
		}

		wrap(w, r.WithContext(context.WithValue(ctx, sessionInContextKey, sess)))
	}
}

// swagger:parameters extendSession
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
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
// This endpoint returns per default a 204 No Content response on success. Older Ory Network projects may
// return a 200 OK response with the session in the body. Returning the session as part of the response
// will be deprecated in the future and should not be relied upon.
//
// This endpoint ignores consecutive requests to extend the same session and returns a 404 error in those
// scenarios. This endpoint also returns 404 errors if the session does not exist.
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
//	  204: emptyResponse
//	  400: errorGeneric
//	  404: errorGeneric
//	  default: errorGeneric
//
//	Extensions:
//	  x-ory-ratelimit-bucket: kratos-admin-low
func (h *Handler) adminSessionExtend(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest().WithError(err.Error()).WithDebug("could not parse UUID")))
		return
	}

	c := h.r.Config()
	if err := h.r.SessionPersister().ExtendSession(r.Context(), id); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	// Default behavior going forward.
	if c.FeatureFlagFasterSessionExtend(r.Context()) {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	trace.SpanFromContext(r.Context()).AddEvent(semconv.NewDeprecatedFeatureUsedEvent(r.Context(), "legacy_slower_session_extend"))

	// WARNING - this will be deprecated at some point!
	s, err := h.r.SessionPersister().GetSession(r.Context(), id, ExpandDefault)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}
	h.r.Writer().Write(w, r, s)
}

func (h *Handler) IsNotAuthenticated(wrap http.HandlerFunc, onAuthenticated http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h.r.SessionManager().SessionActiveForRequest(r.Context(), r); err != nil {
			if e := new(ErrNoActiveSessionFound); errors.As(err, &e) {
				wrap(w, r)
				return
			}
			h.r.Writer().WriteError(w, r, err)
			return
		}

		if onAuthenticated != nil {
			onAuthenticated(w, r)
			return
		}

		h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrForbidden().WithReason("This endpoint can only be accessed without a login session. Please log out and try again.")))
	}
}

func RedirectOnAuthenticated(d interface{ config.Provider }) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		returnTo, err := redir.SecureRedirectTo(r, d.Config().SelfServiceBrowserDefaultReturnTo(ctx), redir.SecureRedirectAllowSelfServiceURLs(d.Config().SelfPublicURL(ctx)))
		if err != nil {
			http.Redirect(w, r, d.Config().SelfServiceBrowserDefaultReturnTo(ctx).String(), http.StatusFound)
			return
		}

		http.Redirect(w, r, returnTo.String(), http.StatusFound)
	}
}

func RedirectOnUnauthenticated(to string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, to, http.StatusFound)
	}
}

func RespondWitherrorGenericOnAuthenticated(h herodot.Writer, err error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.WriteError(w, r, err)
	}
}

// Exchange Session Token Parameters
//
// swagger:parameters exchangeSessionToken
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type exchangeSessionToken struct {
	// The part of the code return when initializing the flow.
	//
	// required: true
	// in: query
	InitCode string `json:"init_code"`

	// The part of the code returned by the return_to URL.
	//
	// required: true
	// in: query
	ReturnToCode string `json:"return_to_code"`
}

// The Response for Registration Flows via API
//
// swagger:model successfulCodeExchangeResponse
type CodeExchangeResponse struct {
	// The Session Token
	//
	// A session token is equivalent to a session cookie, but it can be sent in the HTTP Authorization
	// Header:
	//
	// 		Authorization: bearer ${session-token}
	//
	// The session token is only issued for API flows, not for Browser flows!
	Token string `json:"session_token,omitempty"`

	// The Session
	//
	// The session contains information about the user, the session device, and so on.
	// This is only available for API flows, not for Browser flows!
	//
	// required: true
	Session *Session `json:"session"`
}

// swagger:route GET /sessions/token-exchange frontend exchangeSessionToken
//
// # Exchange Session Token
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Responses:
//	  200: successfulNativeLogin
//	  403: errorGeneric
//	  404: errorGeneric
//	  410: errorGeneric
//	  422: errorGeneric
//	  default: errorGeneric
//
//	Extensions:
//	  x-ory-ratelimit-bucket: kratos-public-medium
func (h *Handler) exchangeCode(w http.ResponseWriter, r *http.Request) {
	var (
		ctx          = r.Context()
		initCode     = r.URL.Query().Get("init_code")
		returnToCode = r.URL.Query().Get("return_to_code")
	)

	if initCode == "" || returnToCode == "" {
		h.r.Writer().WriteError(w, r, herodot.ErrBadRequest().WithReason(`"init_code" and "return_to_code" query params must be set`))
		return
	}

	e, err := h.r.SessionTokenExchangePersister().GetExchangerFromCode(ctx, initCode, returnToCode)
	if err != nil {
		// The session might not be set because the flow encountered an error (e.g. a
		// before-registration webhook rejected the request). Check whether the exchanger
		// exists without requiring a session and, if so, return the flow with its error
		// messages so that the client can act on them.
		pending, pendingErr := h.r.SessionTokenExchangePersister().GetExchangerFromCodeAllowPending(ctx, initCode, returnToCode)
		if pendingErr != nil {
			h.r.Logger().WithRequest(r).WithError(pendingErr).Info("Could not look up pending session token exchanger.")
			h.r.Writer().WriteError(w, r, herodot.ErrNotFound().WithReason(`no session yet for this "code"`))
			return
		}

		f, fErr := h.r.FlowForTokenExchange().GetFlowForTokenExchange(ctx, pending.FlowID)
		if fErr != nil {
			h.r.Logger().WithRequest(r).WithError(fErr).Info("Could not look up flow for pending session token exchange.")
			h.r.Writer().WriteError(w, r, herodot.ErrNotFound().WithReason(`no session yet for this "code"`))
			return
		}

		h.r.Writer().WriteCode(w, r, http.StatusUnprocessableEntity, f)
		return
	}

	sess, err := h.r.SessionPersister().GetSession(ctx, e.SessionID.UUID, ExpandDefault)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().Write(w, r, &CodeExchangeResponse{
		Token:   sess.Token,
		Session: sess,
	})
}

// ManageSessionsAction enumerates the supported actions for the manage-sessions
// endpoint.
//
// swagger:enum ManageSessionsAction
type ManageSessionsAction string

const (
	ManageSessionsActionDisable ManageSessionsAction = "disable"
	ManageSessionsActionDelete  ManageSessionsAction = "delete"
)

// ManageSessionsAllToken is the explicit-consent value that callers must pass
// in `identities` or `sessions` to operate on every row in the network. It
// must be the only element of the array and may not be mixed with explicit
// IDs.
const ManageSessionsAllToken = "*"

// Manage Sessions Body
//
// Body for the bulk session management endpoint. Exactly one of `identities`
// or `sessions` must be provided. To operate on every session in the network,
// pass `identities: ["*"]` — the wildcard must appear alone, never mixed with
// explicit IDs.
//
// swagger:model manageSessionsBody
type ManageSessionsBody struct {
	// Action to perform on the matching sessions.
	//
	// required: true
	// enum: disable,delete
	Action ManageSessionsAction `json:"action"`

	// Identity IDs whose sessions should be disabled or deleted, or `["*"]`
	// to operate on every session in the network. Mutually exclusive with
	// `sessions`.
	Identities []string `json:"identities"`

	// Session IDs to disable or delete. Mutually exclusive with `identities`.
	// The wildcard `["*"]` is not accepted in this field — pass
	// `identities: ["*"]` to scope the operation to every session in the
	// network.
	Sessions []string `json:"sessions"`
}

// swagger:parameters manageSessions
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type manageSessionsParameters struct {
	// in: body
	// required: true
	Body ManageSessionsBody
}

// Manage Sessions Response
//
// Response body for the bulk session management endpoint. Reports how many
// rows the call processed and, for the wildcard variant, whether the network
// still has matching rows left. Explicit-IDs requests always return
// `more: false`. Wildcard callers drain the network by re-issuing the same
// request while `more` is `true`.
//
// swagger:model manageSessionsResponse
type ManageSessionsResponse struct {
	// Number of sessions processed in this call. For `disable`, counts only
	// sessions that were active before the call (already-inactive sessions
	// are skipped). For `delete`, counts every matching row removed.
	Processed int `json:"processed"`

	// True when the call reached the per-call batch limit and additional
	// matching rows may remain. Always false for explicit-IDs requests.
	More bool `json:"more"`
}

// swagger:route POST /admin/sessions identity manageSessions
//
// # Manage sessions in bulk
//
// Disable or delete sessions for a list of identities or a list of sessions in
// a single call. The `action` field selects the operation:
//
//   - `disable` — deactivate matching sessions (sets `active = false`, preserves
//     audit data).
//   - `delete` — permanently delete matching sessions.
//
// Exactly one of `identities` or `sessions` must be provided. To scope the
// operation to every session in the network, pass `identities: ["*"]`; the
// wildcard is not accepted in the `sessions` field. Up to 500 explicit IDs
// are accepted per call.
//
// All requests return `200 OK` with `{processed, more}`. `processed` reports
// how many rows the call affected; for `disable` it counts only sessions
// that were active before the call. `more` is `true` only when a wildcard
// request reached the per-call batch limit and additional rows may remain;
// callers drain the network by re-issuing the same request while `more` is
// `true`. Explicit-IDs requests always return `more: false`.
//
//	Consumes:
//	- application/json
//
//	Schemes: http, https
//
//	Security:
//	  oryAccessToken:
//
//	Responses:
//	  200: manageSessionsResponse
//	  400: errorGeneric
//	  401: errorGeneric
//	  default: errorGeneric
//
//	Extensions:
//	  x-ory-ratelimit-bucket: kratos-admin-low
func (h *Handler) manageSessions(w http.ResponseWriter, r *http.Request) {
	var body ManageSessionsBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.r.Writer().WriteError(w, r, herodot.ErrBadRequest().WithError(err.Error()).WithReason("Invalid JSON body."))
		return
	}

	switch body.Action {
	case ManageSessionsActionDisable, ManageSessionsActionDelete:
	default:
		h.r.Writer().WriteError(w, r, herodot.ErrBadRequest().WithReasonf(
			"The 'action' field must be one of '%s' or '%s'.",
			ManageSessionsActionDisable, ManageSessionsActionDelete))
		return
	}

	identitiesSet := len(body.Identities) > 0
	sessionsSet := len(body.Sessions) > 0
	if identitiesSet == sessionsSet {
		h.r.Writer().WriteError(w, r, herodot.ErrBadRequest().WithReason(
			"Exactly one of 'identities' or 'sessions' must be a non-empty array."))
		return
	}

	var resp *ManageSessionsResponse
	var err error
	switch {
	case identitiesSet:
		resp, err = h.manageByIdentities(r.Context(), body.Action, body.Identities)
	case sessionsSet:
		resp, err = h.manageBySessions(r.Context(), body.Action, body.Sessions)
	default:
		err = errors.WithStack(herodot.ErrInternalServerError().WithReason(
			"neither 'identities' nor 'sessions' set after validation"))
	}
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}
	h.r.Writer().Write(w, r, resp)
}

// manageByIdentities applies the requested action to all sessions belonging
// to the given identity IDs, or to a single bounded batch of every session in
// the network when raw is ["*"]. `more` is set only on the wildcard variant
// when the call hit the per-call batch limit; explicit-IDs always return
// `more: false`.
func (h *Handler) manageByIdentities(ctx context.Context, action ManageSessionsAction, raw []string) (*ManageSessionsResponse, error) {
	ids, wildcard, err := parseManageSessionsIDsOrWildcard(raw)
	if err != nil {
		return nil, herodot.ErrBadRequest().WithReasonf("Could not parse 'identities' field: %s", err)
	}

	p := h.r.SessionPersister()
	switch action {
	case ManageSessionsActionDisable:
		if wildcard {
			return wildcardBatch(ctx, p.RevokeAllSessions)
		}
		n, err := p.RevokeSessionsByIdentities(ctx, ids)
		return explicitBatch(n), err
	case ManageSessionsActionDelete:
		if wildcard {
			return wildcardBatch(ctx, p.DeleteAllSessions)
		}
		n, err := p.DeleteSessionsByIdentities(ctx, ids)
		return explicitBatch(n), err
	default:
		return nil, errors.WithStack(herodot.ErrInternalServerError().WithReasonf(
			"unhandled session action %q", action))
	}
}

// manageBySessions applies the requested action to the given explicit session
// IDs. The wildcard ["*"] is not supported in this field; callers that need
// network-wide scope must use `identities: ["*"]`. See manageByIdentities for
// the response-shape contract.
func (h *Handler) manageBySessions(ctx context.Context, action ManageSessionsAction, raw []string) (*ManageSessionsResponse, error) {
	ids, err := parseManageSessionsIDs(raw)
	if err != nil {
		return nil, herodot.ErrBadRequest().WithReasonf("Could not parse 'sessions' field: %s", err)
	}
	p := h.r.SessionPersister()
	switch action {
	case ManageSessionsActionDisable:
		n, err := p.RevokeSessionsByIDs(ctx, ids)
		return explicitBatch(n), err
	case ManageSessionsActionDelete:
		n, err := p.DeleteSessionsByIDs(ctx, ids)
		return explicitBatch(n), err
	default:
		return nil, errors.WithStack(herodot.ErrInternalServerError().WithReasonf(
			"unhandled session action %q", action))
	}
}

// parseManageSessionsIDsOrWildcard recognizes the network-wide wildcard form
// ["*"] and otherwise delegates to parseManageSessionsIDs. Use it in fields
// that accept the wildcard (currently `identities`); fields that do not
// accept wildcard should call parseManageSessionsIDs directly so the token is
// rejected.
func parseManageSessionsIDsOrWildcard(raw []string) (ids []uuid.UUID, wildcard bool, err error) {
	if len(raw) == 1 && raw[0] == ManageSessionsAllToken {
		return nil, true, nil
	}
	ids, err = parseManageSessionsIDs(raw)
	return ids, false, err
}

// parseManageSessionsIDs interprets a manage-sessions filter array as a list
// of explicit UUIDs and rejects any input containing the wildcard token.
// Callers that accept the wildcard must use parseManageSessionsIDsOrWildcard.
func parseManageSessionsIDs(raw []string) ([]uuid.UUID, error) {
	if len(raw) == 0 {
		return nil, errors.New("array must not be empty")
	}
	if len(raw) > ManageSessionsMaxIDs {
		return nil, fmt.Errorf("at most %d IDs may be provided per call", ManageSessionsMaxIDs)
	}
	ids := make([]uuid.UUID, 0, len(raw))
	for _, s := range raw {
		if s == ManageSessionsAllToken {
			return nil, errors.New("the wildcard '*' is not accepted here")
		}
		id, err := uuid.FromString(s)
		if err != nil {
			return nil, fmt.Errorf("could not parse %q as UUID: %w", s, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// wildcardBatch runs a single chunked bulk-session operation and packages the
// row count plus a `more` flag for the response. `more` is true when the call
// reached the per-call batch limit, signaling that the caller should re-issue
// the request to drain the rest.
//
// When the row count is an exact multiple of the batch size, `more` is set
// even though no rows are left; the caller will issue one extra request that
// returns `{processed: 0, more: false}`. This is intentional — distinguishing
// "exactly batch-size" from "batch-size and more" would cost an extra DB
// query on every call to save one round-trip in a rare edge case.
func wildcardBatch(ctx context.Context, op func(context.Context, int) (int, error)) (*ManageSessionsResponse, error) {
	n, err := op(ctx, manageSessionsWildcardBatchSize)
	if err != nil {
		return nil, err
	}
	return &ManageSessionsResponse{
		Processed: n,
		More:      n == manageSessionsWildcardBatchSize,
	}, nil
}

// explicitBatch packages the row count from an explicit-IDs bulk call.
// `more` is always false because explicit-IDs requests are bounded by
// ManageSessionsMaxIDs and complete in a single statement.
func explicitBatch(n int) *ManageSessionsResponse {
	return &ManageSessionsResponse{Processed: n, More: false}
}
