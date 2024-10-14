// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/x"
	"github.com/ory/x/otelx"
	"github.com/ory/x/pagination/migrationpagination"
)

type (
	handlerDependencies interface {
		x.WriterProvider
		x.LoggingProvider
		IdentitySchemaProvider
		x.CSRFProvider
		config.Provider
		x.TracingProvider
		x.HTTPClientProvider
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
	public.GET(fmt.Sprintf("/%s/:id", SchemasPath), h.getIdentitySchema)
	public.GET(fmt.Sprintf("/%s", SchemasPath), h.getAll)
	public.GET(fmt.Sprintf("%s/%s/:id", x.AdminPrefix, SchemasPath), h.getIdentitySchema)
	public.GET(fmt.Sprintf("%s/%s", x.AdminPrefix, SchemasPath), h.getAll)
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(fmt.Sprintf("/%s/:id", SchemasPath), x.RedirectToPublicRoute(h.r))
	admin.GET(fmt.Sprintf("/%s", SchemasPath), x.RedirectToPublicRoute(h.r))
}

// Raw JSON Schema
//
// swagger:model identitySchema
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type identitySchema json.RawMessage

func (m identitySchema) MarshalJSON() ([]byte, error) {
	return json.RawMessage(m).MarshalJSON()
}

func (m *identitySchema) UnmarshalJSON(data []byte) error {
	mm := json.RawMessage(*m)
	return mm.UnmarshalJSON(data)
}

// Get Identity JSON Schema Response
//
// swagger:parameters getIdentitySchema
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type getIdentitySchema struct {
	// ID must be set to the ID of schema you want to get
	//
	// required: true
	// in: path
	ID string `json:"id"`
}

// swagger:route GET /schemas/{id} identity getIdentitySchema
//
// # Get Identity JSON Schema
//
// Return a specific identity schema.
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Responses:
//	  200: identitySchema
//	  404: errorGeneric
//	  default: errorGeneric
func (h *Handler) getIdentitySchema(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx, span := h.r.Tracer(r.Context()).Tracer().Start(r.Context(), "schema.Handler.getIdentitySchema")
	defer span.End()

	ss, err := h.r.IdentityTraitsSchemas(ctx)
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

	src, err := h.ReadSchema(ctx, s)
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

// List of Identity JSON Schemas
//
// swagger:model identitySchemas
type IdentitySchemas []identitySchemaContainer

// An Identity JSON Schema Container
//
// swagger:model identitySchemaContainer
type identitySchemaContainer struct {
	// The ID of the Identity JSON Schema
	ID string `json:"id"`
	// The actual Identity JSON Schema
	Schema json.RawMessage `json:"schema"`
}

// List Identity JSON Schemas Response
//
// swagger:parameters listIdentitySchemas
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type listIdentitySchemas struct {
	migrationpagination.RequestParameters
}

// List Identity JSON Schemas Response
//
// swagger:response identitySchemas
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type identitySchemasResponse struct {
	migrationpagination.ResponseHeaderAnnotation

	// in: body
	Body IdentitySchemas
}

// swagger:route GET /schemas identity listIdentitySchemas
//
// # Get all Identity Schemas
//
// Returns a list of all identity schemas currently in use.
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Responses:
//	  200: identitySchemas
//	  default: errorGeneric
func (h *Handler) getAll(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx, span := h.r.Tracer(r.Context()).Tracer().Start(r.Context(), "schema.Handler.getAll")
	defer span.End()

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
		src, err := h.ReadSchema(ctx, &schema)
		if err != nil {
			h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("The file for this JSON Schema ID could not be found or opened. This is a configuration issue.").WithDebugf("%+v", err)))
			return
		}

		raw, err := io.ReadAll(io.LimitReader(src, 1024*1024))
		_ = src.Close()
		if err != nil {
			h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("The file for this JSON Schema ID could not be found or opened. This is a configuration issue.").WithDebugf("%+v", err)))
			return
		}

		ss = append(ss, identitySchemaContainer{
			ID:     schema.ID,
			Schema: raw,
		})
	}

	x.PaginationHeader(w, *r.URL, int64(total), page, itemsPerPage)
	h.r.Writer().Write(w, r, ss)
}

func (h *Handler) ReadSchema(ctx context.Context, schema *Schema) (src io.ReadCloser, err error) {
	ctx, span := h.r.Tracer(ctx).Tracer().Start(ctx, "schema.Handler.ReadSchema")
	defer otelx.End(span, &err)

	if schema.URL.Scheme == "file" {
		src, err = os.Open(schema.URL.Host + schema.URL.Path)
		if err != nil {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithWrap(err).WithReason("Unable to fetch identity schema."))
		}
	} else if schema.URL.Scheme == "base64" {
		data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(schema.RawURL, "base64://"))
		if err != nil {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithWrap(err).WithReason("Unable to fetch identity schema."))
		}
		src = io.NopCloser(strings.NewReader(string(data)))
	} else {
		resp, err := h.r.HTTPClient(ctx).Get(schema.URL.String())
		if err != nil {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithWrap(err).WithReason("Unable to fetch identity schema."))
		}
		src = resp.Body
	}
	return src, nil
}
