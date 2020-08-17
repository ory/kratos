package session

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/x/decoderx"

	"github.com/ory/x/errorsx"

	"github.com/ory/herodot"

	"github.com/ory/kratos/driver/configuration"
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
type RevokeSessionParams struct {
	// The Session Token
	//
	// Invalidate this session token.
	//
	// required: true
	// in: body
	SessionToken string `json:"session_token"`
}

// swagger:route DELETE /sessions public revokeSession
//
// Revoke and Invalidate a Session
//
// Use this endpoint to revoke a session using its token. This endpoint is particularly useful for API clients
// such as mobile apps to log the user out of the system and invalidate the session.
//
// This endpoint does not remove any HTTP Cookies - use the Self-Service Logout Flow instead.
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
func (h *Handler) revoke(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var p RevokeSessionParams
	if err := h.dx.Decode(r, &p, decoderx.HTTPJSONDecoder()); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	if err := h.r.SessionPersister().RevokeSessionByToken(r.Context(), p.SessionToken); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// swagger:route GET /sessions/whoami public whoami
//
// Check who the current HTTP session belongs to
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
//     - sessionToken
//
//     Responses:
//       200: session
//       403: genericError
//       500: genericError
func (h *Handler) whoami(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	s, err := h.r.SessionManager().FetchFromRequest(r.Context(), r)
	if err != nil {
		h.r.Audit().WithRequest(r).WithError(err).Info("No valid session cookie found.")
		h.r.Writer().WriteError(w, r,
			errors.WithStack(herodot.ErrUnauthorized.WithReasonf("No valid session cookie found.")))
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

func RedirectOnAuthenticated(c configuration.Provider) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		returnTo, err := x.SecureRedirectTo(r, c.SelfServiceBrowserDefaultReturnTo(), x.SecureRedirectAllowSelfServiceURLs(c.SelfPublicURL()))
		if err != nil {
			http.Redirect(w, r, c.SelfServiceBrowserDefaultReturnTo().String(), http.StatusFound)
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
