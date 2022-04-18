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

	"github.com/ory/x/urlx"

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
	h.r.CSRFHandler().IgnoreGlobs(
		"/"+SchemasPath+"/*",
		x.AdminPrefix+"/"+SchemasPath+"/*",
	)
	public.GET(fmt.Sprintf("/%s/:id", SchemasPath), h.getByID)
	public.GET(fmt.Sprintf("/%s", SchemasPath), h.getAll)
	public.GET(fmt.Sprintf("%s/%s/:id", x.AdminPrefix, SchemasPath), h.getByID)
	public.GET(fmt.Sprintf("%s/%s", x.AdminPrefix, SchemasPath), h.getAll)
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
	ss, err := h.r.IdentityTraitsSchemas(r.Context())
	if err != nil {
		h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithWrap(err)))
		return
	}

	id := ps.ByName("id")
	s, err := ss.GetByID(id)
	if err != nil {
		// Maybe it is a base64 encoded ID?
		if dec, err := base64.RawURLEncoding.DecodeString(id); err == nil {
			id = string(dec)
		}

		s, err = ss.GetByID(id)
		if err != nil {
			h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrNotFound.WithReasonf("Identity schema `%s` could not be found.", id)))
			return
		}
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

	allSchemas, err := h.r.IdentityTraitsSchemas(r.Context())
	if err != nil {
		h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to load identity schemas").WithWrap(err)))
		return
	}
	total := allSchemas.Total()
	schemas := allSchemas.List(page, itemsPerPage)

	var ss IdentitySchemas
	for k := range schemas {
		schema := schemas[k]
		src, err := ReadSchema(&schema)
		if err != nil {
			h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("The file for this JSON Schema ID could not be found or opened. This is a configuration issue.").WithDebugf("%+v", err)))
			return
		}

		raw, err := ioutil.ReadAll(src)
		_ = src.Close()
		if err != nil {
			h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("The file for this JSON Schema ID could not be found or opened. This is a configuration issue.").WithDebugf("%+v", err)))
			return
		}

		ss = append(ss, identitySchema{
			ID:     schema.ID,
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
