package schema

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/x/urlx"

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
	public.GET(fmt.Sprintf("/%s/:id", SchemasPath), h.getByID)
	public.GET(fmt.Sprintf("/%s", SchemasPath), h.getAll)
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(fmt.Sprintf("/%s/:id", SchemasPath), x.RedirectToPublicRoute(h.r))
	admin.GET(fmt.Sprintf("/%s", SchemasPath), x.RedirectToPublicRoute(h.r))
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

// swagger:route GET /schemas/{id} v0alpha2 getJsonSchema
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
func (h *Handler) getByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	s, err := h.r.IdentityTraitsSchemas(r.Context()).GetByID(ps.ByName("id"))
	if err != nil {
		h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrNotFound.WithDebugf("%+v", err)))
		return
	}

	src, err := ReadSchema(s)
	if err != nil {
		h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("The file for this JSON Schema ID could not be found or opened. This is a configuration issue.").WithDebugf("%+v", err)))
		return
	}
	defer src.Close()

	w.Header().Add("Content-Type", "application/json")
	if _, err := io.Copy(w, src); err != nil {
		h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("The file for this JSON Schema ID could not be found or opened. This is a configuration issue.").WithDebugf("%+v", err)))
		return
	}
}

// Raw identity Schema list
//
// swagger:model identitySchemas
type IdentitySchemas []identitySchema

// swagger:model identitySchema
type identitySchema struct {
	// The ID of the Identity JSON Schema
	ID string `json:"id"`
	// The actual Identity JSON Schema
	Schema json.RawMessage `json:"schema"`
}

// nolint:deadcode,unused
// swagger:parameters listIdentitySchemas
type listIdentitySchemas struct {
	x.PaginationParams
}

// swagger:route GET /schemas v0alpha2 listIdentitySchemas
//
// Get all Identity Schemas
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: identitySchemas
//       500: jsonError
func (h *Handler) getAll(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	page, itemsPerPage := x.ParsePagination(r)

	schemas := h.r.IdentityTraitsSchemas(r.Context()).List(page, itemsPerPage)
	total := h.r.IdentityTraitsSchemas(r.Context()).Total()

	var ss IdentitySchemas

	for _, schema := range schemas {
		s, err := h.r.IdentityTraitsSchemas(r.Context()).GetByID(schema.ID)
		if err != nil {
			h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrNotFound.WithDebugf("%+v", err)))
			return
		}

		src, err := ReadSchema(s)
		if err != nil {
			h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("The file for this JSON Schema ID could not be found or opened. This is a configuration issue.").WithDebugf("%+v", err)))
			return
		}
		defer src.Close()

		raw, err := ioutil.ReadAll(src)
		if err != nil {
			h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("The file for this JSON Schema ID could not be found or opened. This is a configuration issue.").WithDebugf("%+v", err)))
			return
		}

		ss = append(ss, identitySchema{
			ID:     s.ID,
			Schema: raw,
		})
	}

	x.PaginationHeader(w, urlx.AppendPaths(h.r.Config(r.Context()).SelfPublicURL(), fmt.Sprintf("/%s", SchemasPath)), int64(total), page, itemsPerPage)
	h.r.Writer().Write(w, r, ss)
}

func ReadSchema(schema *Schema) (src io.ReadCloser, err error) {
	if schema.URL.Scheme == "file" {
		src, err = os.Open(schema.URL.Host + schema.URL.Path)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	} else if schema.URL.Scheme == "base64" {
		data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(schema.RawURL, "base64://"))
		if err != nil {
			return nil, errors.WithStack(err)
		}
		src = io.NopCloser(strings.NewReader(string(data)))
	} else {
		resp, err := http.Get(schema.URL.String())
		if err != nil {
			return nil, errors.WithStack(err)
		}
		src = resp.Body
	}
	return src, nil
}
