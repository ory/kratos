package session

import (
	"net/http"

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
	public.GET(SessionsWhoamiPath, h.fromCookie)
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	// admin.GET(SessionsWhoisPath, h.fromPath)
}

func (h *Handler) fromCookie(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	s, err := h.r.SessionManager().FetchFromRequest(r.Context(), w, r)
	if err != nil {
		h.r.Writer().WriteError(w, r, herodot.ErrUnauthorized.WithReasonf("No valid session cookie found.").WithDebugf("%+v", err))
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
		http.Redirect(w, r, c.DefaultReturnToURL().String(), http.StatusFound)
	}
}

func RedirectOnUnauthenticated(to string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		http.Redirect(w, r, to, http.StatusFound)
	}
}
