package schema

import (
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/x"
	"github.com/ory/x/jsonx"
	"github.com/ory/x/urlx"
)

type (
	Configuration interface {
		SelfPublicURL() *url.URL
	}
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

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.PUT("/schemas", h.registerSchema)
}

func (h *Handler) registerSchema(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var s Schema
	if err := errors.WithStack(jsonx.NewStrictDecoder(r.Body).Decode(&s)); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	if err := h.r.SchemaPersister().RegisterSchema(&s); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().WriteCreated(w, r,
		urlx.AppendPaths(
			h.c.SelfPublicURL(),
			"schemas",
			s.ID.String(),
		).String(),
		&s)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	s, err := h.r.SchemaPersister().GetSchema(x.ParseUUID(ps.ByName("id")))
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	if strings.HasPrefix(s.URL, "file://") {
		fp := strings.TrimPrefix(s.URL, "file://")
		info, err := os.Stat(fp)
		if err != nil {
			if os.IsNotExist(err) {
				h.r.Writer().WriteError(w, r, herodot.ErrNotFound)
				return
			}
		}

		if info.IsDir() {
			h.r.Writer().WriteError(w, r, herodot.ErrNotFound)
			return
		}

		http.ServeFile(w, r, fp)
	} else {
		resp, err := http.Get(s.URL)
		if err != nil {
			h.r.Writer().WriteError(w, r, err)
			return
		}
		defer resp.Body.Close()

		var content []byte

		if _, err = resp.Body.Read(content); err != nil {
			h.r.Writer().WriteError(w, r, err)
			return
		}

		h.r.Writer().Write(w, r, content)
	}
}
