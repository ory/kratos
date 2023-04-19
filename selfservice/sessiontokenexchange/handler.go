// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sessiontokenexchange

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/ory/herodot"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

const (
	RouteExchangeCodeForSessionToken = "/self-service/exchange-code-for-session-token" // #nosec G101
)

type (
	handlerDependencies interface {
		PersistenceProvider
		config.Provider
		x.WriterProvider
		session.PersistenceProvider
	}

	HandlerProvider interface {
		SessionTokenExchangeHandler() *Handler
	}
	Handler struct {
		d handlerDependencies
	}
)

func NewHandler(d handlerDependencies) *Handler {
	return &Handler{d: d}
}

func (h *Handler) RegisterPublicRoutes(public *x.RouterPublic) {
	public.GET(RouteExchangeCodeForSessionToken, h.exchangeCode)
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(RouteExchangeCodeForSessionToken, x.RedirectToPublicRoute(h.d))
}

// Exchange Session Token Parameters
//
// swagger:parameters exchangeSessionToken
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type exchangeSessionToken struct {
	// The Session Token Exchange Code
	//
	// required: true
	// in: query
	SessionTokenExchangeCode string `json:"code"`
}

// The Response for Registration Flows via API
//
// swagger:model successfulCodeExchangeResponse
type Response struct {
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
	Session *session.Session `json:"session"`
}

// swagger:route GET /self-service/exchange-code-for-session-token frontend exchangeSessionToken
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
//	  default: errorGeneric
func (h *Handler) exchangeCode(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	code := r.URL.Query().Get("code")
	ctx := r.Context()

	if code == "" {
		h.d.Writer().WriteError(w, r, herodot.ErrBadRequest.WithReason(`"code" query param must be set`))
		return
	}

	e, err := h.d.SessionTokenExchangePersister().GetExchangerFromCode(ctx, code)
	if err != nil {
		h.d.Writer().WriteError(w, r, herodot.ErrNotFound.WithReason(`no session yet for this "code"`))
		return
	}

	sess, err := h.d.SessionPersister().GetSession(ctx, e.SessionID.UUID, session.ExpandDefault)
	if err != nil {
		h.d.Writer().WriteError(w, r, err)
		return
	}

	h.d.Writer().Write(w, r, &Response{
		Token:   sess.Token,
		Session: sess,
	})
}
