package session

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/ory/herodot"
	"github.com/ory/hive/x"
)

func NewHandler(
	r Registry,
	h herodot.Writer,
) *Handler {
	return &Handler{
		r: r,
		h: h,
	}
}

const (
	CheckPath = "/sessions/me"
)

type Handler struct {
	r Registry
	h herodot.Writer
}

func (h *Handler) RegisterPublicRoutes(public *x.RouterPublic) {
	public.GET(CheckPath, h.fromCookie)
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET("/sessions/:sid", h.fromPath)
}

func (h *Handler) fromCookie(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	s, err := h.r.SessionManager().FetchFromRequest(r)
	if err != nil {
		h.h.WriteError(w, r, err)
		return
	}

	s.Devices = nil
	s.Identity = s.Identity.WithoutCredentials()

	h.h.Write(w, r, s)
}

func (h *Handler) fromPath(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.WriteHeader(505)
}
