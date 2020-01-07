package schema

import (
	"github.com/julienschmidt/httprouter"
	"github.com/ory/kratos/x"
	"net/http"
)

type (
	Configuration       interface{}
	handlerDependencies interface {
		PoolProvider
		x.WriterProvider
	}
	HandlerProvider interface {
		IdentityHandler() *Handler
	}
	Handler struct {
		c Configuration
		r handlerDependencies
	}
)

func NewHandler(
	c Configuration,
	r handlerDependencies,
) *Handler {
	return &Handler{
		c: c,
		r: r,
	}
}

func (h *Handler) RegisterPublicRoutes(public *x.RouterPublic) {
	public.GET("/schemas/:id", h.get)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	s, err := h.r.SchemaPool().GetSchema(x.ParseUUID(ps.ByName("id")))
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().Write(w, r, s)
}
