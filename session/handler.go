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
	RouteWhoami = "/sessions/whoami"
)

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	for _, m := range []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodDelete} {
		// Redirect to public endpoint
		admin.Handle(m, RouteWhoami, x.RedirectToPublicRoute(h.r))
	}
}

func (h *Handler) RegisterPublicRoutes(public *x.RouterPublic) {
	h.r.CSRFHandler().ExemptPath(RouteWhoami)

	for _, m := range []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace} {
		public.Handle(m, RouteWhoami, h.whoami)
	}
}

// nolint:deadcode,unused
// swagger:parameters toSession
type toSession struct {
	// in: header
	SessionToken string `json:"X-Session-Token"`

	// in: header
	SessionCookie string `json:"X-Session-Cookie"`
}

// swagger:route GET /sessions/whoami public toSession
//
// Check Who the Current HTTP Session Belongs To
//
// Uses the HTTP Headers in the GET request to determine (e.g. by using checking the cookies) who is authenticated.
// Returns a session object in the body or 401 if the credentials are invalid or no credentials were sent.
// Additionally when the request it successful it adds the user ID to the 'X-Kratos-Authenticated-Identity-Id' header in the response.
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
// - if the `X-Session-Token` HTTP header was set with a valid Ory Kratos Session Token;
// - if the `X-Session-Cookie` HTTP header was set containing a valid Ory Kratos Session Cookie (only the value!).
//
// If none of these headers are set or the cooke or token are invalid, the endpoint returns a HTTP 401 status code.
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: session
//       401: jsonError
//       500: jsonError
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
		returnTo, err := x.SecureRedirectTo(r, d.Config(r.Context()).SelfServiceBrowserDefaultReturnTo(), x.SecureRedirectAllowSelfServiceURLs(d.Config(r.Context()).SelfPublicURL(r)))
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
