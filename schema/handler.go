package schema

import (
	"github.com/julienschmidt/httprouter"
	"github.com/ory/kratos/x"
	"net/http"
	"os"
	"strings"
)

type (
	Configuration       interface{}
	handlerDependencies interface {
		PersistenceProvider
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
	s, err := h.r.SchemaPersister().GetSchema(x.ParseUUID(ps.ByName("id")))
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	var content []byte

	if strings.HasPrefix(s.URL, "file://") {
		f, err := os.Open(s.URL)
		if err != nil {
			h.r.Writer().WriteError(w, r, err)
			return
		}

		if _, err = f.Read(content); err != nil {
			h.r.Writer().WriteError(w, r, err)
			return
		}

		if err = f.Close(); err != nil {
			h.r.Writer().WriteError(w, r, err)
			return
		}
	} else {
		resp, err := http.Get(s.URL)
		if err != nil {
			h.r.Writer().WriteError(w, r, err)
			return
		}
		defer resp.Body.Close()

		if _, err = resp.Body.Read(content); err != nil {
			h.r.Writer().WriteError(w, r, err)
			return
		}
	}

	h.r.Writer().Write(w, r, content)
}
