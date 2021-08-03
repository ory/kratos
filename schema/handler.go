package schema

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/ory/kratos/driver/config"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"github.com/ory/kratos/x"
)

type (
	handlerDependencies interface {
		x.WriterProvider
		x.LoggingProvider
		IdentityTraitsProvider
		x.CSRFProvider
		config.Provider
	}
	Handler struct {
		r handlerDependencies
	}
	HandlerProvider interface {
		SchemaHandler() *Handler
	}
)

func NewHandler(r handlerDependencies) *Handler {
	return &Handler{r: r}
}

const SchemasPath string = "schemas"

func (h *Handler) RegisterPublicRoutes(public *x.RouterPublic) {
	h.r.CSRFHandler().IgnoreGlobs(fmt.Sprintf("/%s/*", SchemasPath))
	public.GET(fmt.Sprintf("/%s/:id", SchemasPath), h.get)
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(fmt.Sprintf("/%s/:id", SchemasPath), x.RedirectToPublicRoute(h.r))
}

// Raw JSON Schema
//
// swagger:model jsonSchema
// nolint:deadcode,unused
type jsonSchema json.RawMessage

// nolint:deadcode,unused
// swagger:parameters getJsonSchema
type getJsonSchema struct {
	// ID must be set to the ID of schema you want to get
	//
	// required: true
	// in: path
	ID string `json:"id"`
}

// swagger:route GET /schemas/{id} v0alpha1 getJsonSchema
//
// Get a JSON Schema
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: jsonSchema
//       404: jsonError
//       500: jsonError
func (h *Handler) get(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	s, err := h.r.IdentityTraitsSchemas(r.Context()).GetByID(ps.ByName("id"))
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

	w.Header().Add("Content-Type", "application/json")
	if _, err := io.Copy(w, src); err != nil {
		h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("The file for this JSON Schema ID could not be found or opened. This is a configuration issue.").WithDebugf("%+v", err)))
		return
	}
}
