package schema

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/julienschmidt/httprouter"

	"github.com/ory/herodot"
	"github.com/ory/kratos/x"
)

type (
	handlerDependencies interface {
		x.WriterProvider
		IdentityTraitsSchemas() Schemas
		Logger() logrus.FieldLogger
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
		h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrNotFound.WithDebugf("%+v", err)))
		return
	}
	var src io.ReadCloser

	if s.URL.Scheme == "file" {
		src, err = os.Open(s.URL.Host + s.URL.Path)
		if err != nil {
			h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("The file for this JSON Schema ID could not be found or opened. This is a configuration issue.").WithDebugf("%+v", err)))
			return
		}
		defer src.Close()
	} else {
		resp, err := http.Get(s.URL.String())
		if err != nil {
			h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("The file for this JSON Schema ID could not be found or opened. This is a configuration issue.").WithDebugf("%+v", err)))
			return
		}
		defer resp.Body.Close()
		src = resp.Body
	}

	if _, err := io.Copy(w, src); err != nil {
		h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("The file for this JSON Schema ID could not be found or opened. This is a configuration issue.").WithDebugf("%+v", err)))
		return
	}
}
