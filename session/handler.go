package session

import (
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/x/errorsx"

	"github.com/ory/herodot"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/x"
)

type (
	handlerDependencies interface {
		ManagementProvider
		x.WriterProvider
	}
	HandlerProvider interface {
		SessionHandler() *Handler
	}
	Handler struct {
		r handlerDependencies
	}
)

func NewHandler(
	r handlerDependencies,
) *Handler {
	return &Handler{
		r: r,
	}
}

const (
	SessionsWhoamiPath = "/sessions/whoami"
	// SessionsWhoisPath  = "/sessions/whois"
)

func (h *Handler) RegisterPublicRoutes(public *x.RouterPublic) {
	for _, m := range []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace} {
		public.Handle(m, SessionsWhoamiPath, h.whoami)
	}
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	// admin.GET(SessionsWhoisPath, h.fromPath)
}

// swagger:route GET /sessions/whoami public whoami
//
// Check who the current HTTP session belongs to
//
// Uses the HTTP Headers in the GET request to determine (e.g. by using checking the cookies) who is authenticated.
// Returns a session object or 401 if the credentials are invalid or no credentials were sent.
//
// This endpoint is useful for reverse proxies and API Gateways.
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: session
//       403: genericError
//       500: genericError
func (h *Handler) whoami(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	s, err := h.r.SessionManager().FetchFromRequest(r.Context(), w, r)
	if err != nil {
		h.r.Writer().WriteError(w, r,
			errors.WithStack(herodot.ErrUnauthorized.WithReasonf("No valid session cookie found.").WithDebugf("%+v", err)),
		)
		return
	}

	// s.Devices = nil
	s.Identity = s.Identity.CopyWithoutCredentials()

	h.r.Writer().Write(w, r, s)
}

// func (h *Handler) fromPath(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	w.WriteHeader(505)
// }

func (h *Handler) IsAuthenticated(wrap httprouter.Handle, onUnauthenticated httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if _, err := h.r.SessionManager().FetchFromRequest(r.Context(), w, r); err != nil {
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
		if _, err := h.r.SessionManager().FetchFromRequest(r.Context(), w, r); err != nil {
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
		returnTo, err := x.DetermineReturnToURL(r.URL, c.DefaultReturnToURL(), []url.URL{*c.SelfPublicURL()})
		if err != nil {
			http.Redirect(w, r, c.DefaultReturnToURL().String(), http.StatusFound)
		}

		http.Redirect(w, r, returnTo, http.StatusFound)
	}
}

func RedirectOnUnauthenticated(to string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		http.Redirect(w, r, to, http.StatusFound)
	}
}
