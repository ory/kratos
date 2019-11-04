package errorx

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/ory/hive/x"
)

type Handler struct {
	r Registry
}

func NewHandler(
	r Registry,
) *Handler {
	return &Handler{
		r: r,
	}
}

func (h *Handler) RegisterPublicRoutes(public *x.RouterPublic) {
	public.GET("/errors", h.get)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	es, err := h.r.ErrorManager().Read(r.Context(), r.URL.Query().Get("error"))
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().Write(w, r, es)
}
