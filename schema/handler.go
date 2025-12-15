// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/nosurfx"
	"github.com/ory/kratos/x/redir"
	"github.com/ory/x/errorsx"
	"github.com/ory/x/otelx"
	"github.com/ory/x/pagination/migrationpagination"
)

type (
	handlerDependencies interface {
		x.WriterProvider
		x.LoggingProvider
		IdentitySchemaProvider
		nosurfx.CSRFProvider
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

const (
	SchemasPath   string = "schemas"
	maxSchemaSize        = 1024 * 1024 // 1 MB
)

func (h *Handler) RegisterPublicRoutes(public *x.RouterPublic) {
	h.r.CSRFHandler().IgnoreGlobs(
		"/"+SchemasPath+"/*",
		x.AdminPrefix+"/"+SchemasPath+"/*",
	)
	public.GET(fmt.Sprintf("/%s/{id}", SchemasPath), h.getIdentitySchema)
	public.GET(fmt.Sprintf("/%s", SchemasPath), h.getAll)
	public.GET(fmt.Sprintf("%s/%s/{id}", x.AdminPrefix, SchemasPath), h.getIdentitySchema)
	public.GET(fmt.Sprintf("%s/%s", x.AdminPrefix, SchemasPath), h.getAll)
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(fmt.Sprintf("/%s/{id}", SchemasPath), redir.RedirectToPublicRoute(h.r))
	admin.GET(fmt.Sprintf("/%s", SchemasPath), redir.RedirectToPublicRoute(h.r))
}

// Raw JSON Schema
//
// swagger:model identitySchema
type _ json.RawMessage

// Get Identity JSON Schema Response
//
// swagger:parameters getIdentitySchema
type _ struct {
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
func (h *Handler) getIdentitySchema(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.r.Tracer(r.Context()).Tracer().Start(r.Context(), "schema.Handler.getIdentitySchema")
	defer span.End()

	ss, err := h.r.IdentityTraitsSchemas(ctx)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	id := r.PathValue("id")
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

	raw, err := h.ReadSchema(ctx, s.URL)
	if err != nil {
		code, ok := errorsx.GetCodeFromHerodotError(err)

		if errors.Is(err, fs.ErrNotExist) || (ok && code == http.StatusNotFound) {
			h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrMisconfiguration.WithReason("The file for this JSON Schema ID could not be found/fetched. This is a configuration issue.").WithDebugf("%+v", err)))
		} else if ok && code == http.StatusBadGateway {
			h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrUpstreamError.WithReason("The file for this JSON Schema ID could not be fetched. This is an upstream issue.").WithDebugf("%+v", err)))
		} else {
			h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrInternalServerError.WithReason("The file for this JSON Schema ID could not be read. This is an I/O issue.").WithDebugf("%+v", err)))
		}
		return
	}

	w.Header().Add("Content-Type", "application/json")
	h.r.Writer().Write(w, r, json.RawMessage(raw))
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
	// required: true
	ID string `json:"id"`
	// The actual Identity JSON Schema
	// required: true
	Schema json.RawMessage `json:"schema"`
}

// List Identity JSON Schemas Response
//
// swagger:parameters listIdentitySchemas
type _ struct {
	migrationpagination.RequestParameters
}

// List Identity JSON Schemas Response
//
// swagger:response identitySchemas
type _ struct {
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
func (h *Handler) getAll(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.r.Tracer(r.Context()).Tracer().Start(r.Context(), "schema.Handler.getAll")
	defer span.End()

	page, itemsPerPage := x.ParsePagination(r)

	allSchemas, err := h.r.IdentityTraitsSchemas(r.Context())
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}
	total := allSchemas.Total()
	schemas := allSchemas.List(page, itemsPerPage)

	ss := make(IdentitySchemas, len(schemas))
	for i, schema := range schemas {
		raw, err := h.ReadSchema(ctx, schema.URL)
		if err != nil {
			h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrMisconfiguration.WithReasonf("The file for a JSON Schema ID could not be found or opened. This is a configuration issue.").WithWrap(err)))
			return
		}
		ss[i] = identitySchemaContainer{
			ID:     schema.ID,
			Schema: raw,
		}
	}

	x.PaginationHeader(w, *r.URL, int64(total), page, itemsPerPage)
	h.r.Writer().Write(w, r, ss)
}

func (h *Handler) ReadSchema(ctx context.Context, uri *url.URL) (data []byte, err error) {
	ctx, span := h.r.Tracer(ctx).Tracer().Start(ctx, "schema.Handler.ReadSchema")
	defer otelx.End(span, &err)

	switch uri.Scheme {
	case "file":
		data, err = os.ReadFile(uri.Host + uri.Path)
		if err != nil {
			return nil, errors.WithStack(fmt.Errorf("could not read schema file: %w", err))
		}
	case "base64":
		data, err = base64.StdEncoding.DecodeString(strings.TrimPrefix(uri.String(), "base64://"))
		if err != nil {
			return nil, errors.WithStack(fmt.Errorf("could not decode schema file: %w", err))
		}
	default:
		req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, uri.String(), nil)
		if err != nil {
			return nil, errors.WithStack(fmt.Errorf("could not create request: %w", err))
		}
		resp, err := h.r.HTTPClient(ctx).Do(req)
		if err != nil {
			return nil, errors.WithStack(herodot.ErrUpstreamError.WithReason("could not fetch schema").WithError(err.Error()).WithDetail("uri", uri))
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusOK {
			if resp.StatusCode == http.StatusNotFound {
				return nil, herodot.ErrNotFound.WithDetail("url", uri)
			}
			return nil, errors.WithStack(herodot.ErrUpstreamError.WithError("upstream error").WithDetail("status_code", resp.StatusCode).WithDetail("uri", uri))
		}
		data, err = io.ReadAll(io.LimitReader(resp.Body, maxSchemaSize))
		if err != nil {
			return nil, errors.WithStack(herodot.ErrUpstreamError.WithReason("could not read schema response").WithError(err.Error()).WithDetail("uri", uri))
		}
	}
	return data, nil
}
