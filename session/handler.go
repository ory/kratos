package session

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/x/decoderx"

	"github.com/ory/x/errorsx"

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
	RouteWhoami = "/sessions/whoami"
	RouteRevoke = "/sessions"
	// SessionsWhoisPath  = "/sessions/whois"
)

func (h *Handler) RegisterPublicRoutes(public *x.RouterPublic) {
	h.r.CSRFHandler().ExemptPath(RouteWhoami)
	h.r.CSRFHandler().ExemptPath(RouteRevoke)

	for _, m := range []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace} {
		public.Handle(m, RouteWhoami, h.whoami)
	}

	public.DELETE(RouteRevoke, h.revoke)
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	// admin.GET(SessionsWhoisPath, h.fromPath)
}

// swagger:parameters revokeSession
// nolint:deadcode,unused
type revokeSessionParameters struct {
	// in: body
	// required: true
	Body revokeSession
}

type revokeSession struct {
	// The Session Token
	//
	// Invalidate this session token.
	//
	// required: true
	SessionToken string `json:"session_token"`
}

// swagger:route DELETE /sessions public revokeSession
//
// Initialize Logout Flow for API Clients - Revoke a Session
//
// Use this endpoint to revoke a session using its token. This endpoint is particularly useful for API clients
// such as mobile apps to log the user out of the system and invalidate the session.
//
// This endpoint does not remove any HTTP Cookies - use the Browser-Based Self-Service Logout Flow instead.
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
//       400: genericError
//       500: genericError
func (h *Handler) revoke(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var p revokeSession
	if err := h.dx.Decode(r, &p,
		decoderx.HTTPJSONDecoder(),
		decoderx.HTTPDecoderAllowedMethods("DELETE")); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	if err := h.r.SessionPersister().RevokeSessionByToken(r.Context(), p.SessionToken); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// nolint:deadcode,unused
// swagger:parameters whoami
type whoamiParameters struct {
	// in: header
	Cookie string `json:"Cookie"`

	// in: authorization
	Authorization string `json:"Authorization"`
}

// swagger:route GET /sessions/whoami public whoami
//
// Check Who the Current HTTP Session Belongs To
//
// Uses the HTTP Headers in the GET request to determine (e.g. by using checking the cookies) who is authenticated.
// Returns a session object in the body or 401 if the credentials are invalid or no credentials were sent.
// Additionally when the request it successful it adds the user ID to the 'X-Kratos-Authenticated-Identity-Id' header in the response.
//
// This endpoint is useful for reverse proxies and API Gateways.
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Security:
//       sessionToken:
//
//     Responses:
//       200: session
//       401: genericError
//       500: genericError
func (h *Handler) whoami(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	s, err := h.r.SessionManager().FetchFromRequest(r.Context(), r)
	if err != nil {
		h.r.Audit().WithRequest(r).WithError(err).Info("No valid session cookie found.")
		h.r.Writer().WriteError(w, r, herodot.ErrUnauthorized.WithWrap(err).WithReasonf("No valid session cookie found."))
		return
	}

	// s.Devices = nil
	s.Identity = s.Identity.CopyWithoutCredentials()

	// Set userId as the X-Kratos-Authenticated-Identity-Id header.
	w.Header().Set("X-Kratos-Authenticated-Identity-Id", s.Identity.ID.String())

	h.r.Writer().Write(w, r, s)
}

func (h *Handler) IsAuthenticated(wrap httprouter.Handle, onUnauthenticated httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if _, err := h.r.SessionManager().FetchFromRequest(r.Context(), r); err != nil {
			if onUnauthenticated != nil {
				onUnauthenticated(w, r, ps)
				return
			}

			h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrForbidden.WithReason("This endpoint can only be accessed with a valid session. Please log in and try again.").WithDebugf("%+v", err)))
			return
		}

		wrap(w, r, ps)
	}
}

func (h *Handler) IsNotAuthenticated(wrap httprouter.Handle, onAuthenticated httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if _, err := h.r.SessionManager().FetchFromRequest(r.Context(), r); err != nil {
			if errorsx.Cause(err).Error() == ErrNoActiveSessionFound.Error() {
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
