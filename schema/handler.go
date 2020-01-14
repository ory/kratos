package schema

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"

	"github.com/ory/herodot"
	"github.com/ory/kratos/x"
)

type (
	handlerDependencies interface {
		x.WriterProvider
		IdentityTraitsSchemas() Schemas
	}
	Handler struct {
		r handlerDependencies
	}
	HandlerProvider interface {
		SchemaHandler() *Handler
	}
)

func NewHandler(r handlerDependencies) *Handler {
	return &Handler{
		r: r,
	}
}

const SchemasPath string = "schemas"

func (h *Handler) RegisterPublicRoutes(public *x.RouterPublic) {
	public.GET(fmt.Sprintf("/%s/:id", SchemasPath), h.get)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	s, err := h.r.IdentityTraitsSchemas().GetByID(ps.ByName("id"))
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	if s.URL.Scheme == "file://" {
		f, err := os.Open(s.URL.Path)
		if err != nil {
			if os.IsNotExist(err) {
				h.r.Writer().WriteError(w, r, herodot.ErrNotFound)
				return
			}

			h.r.Writer().WriteError(w, r, err)
			return
		}
		defer f.Close()

		var w http.ResponseWriter
		if _, err := io.Copy(w, f); err != nil {
			h.r.Writer().WriteError(w, r, err)
			return
		}
	} else {
		resp, err := http.Get(s.URL.String())
		if err != nil {
			h.r.Writer().WriteError(w, r, err)
			return
		}
		defer resp.Body.Close()

		if _, err = io.Copy(w, resp.Body); err != nil {
			h.r.Writer().WriteError(w, r, err)
			return
		}
	}
}
