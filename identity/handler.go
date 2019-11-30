package identity

import (
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/x/jsonx"
	"github.com/ory/x/urlx"

	"github.com/ory/x/pagination"

	"github.com/ory/kratos/x"
)

type (
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

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET("/identities", h.list)
	admin.GET("/identities/:id", h.get)
	admin.DELETE("/identities/:id", h.delete)

	admin.POST("/identities", h.create)
	admin.PUT("/identities/:id", h.update)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	limit, offset := pagination.Parse(r, 100, 0, 500)
	is, err := h.r.IdentityPool().ListIdentities(r.Context(), limit, offset)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().Write(w, r, is)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	i, err := h.r.IdentityPool().GetIdentity(r.Context(), x.ParseUUID(ps.ByName("id")))
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().Write(w, r, i)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var i Identity
	if err := errors.WithStack(jsonx.NewStrictDecoder(r.Body).Decode(&i)); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	// We do not allow setting credentials using this method
	i.Credentials = nil
	// We do not allow setting the ID using this method
	i.ID = uuid.Nil

	err := h.r.IdentityPool().CreateIdentity(r.Context(), &i)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().WriteCreated(w, r,
		urlx.AppendPaths(
			h.c.SelfAdminURL(),
			"identities",
			i.ID.String(),
		).String(),
		&i,
	)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var i Identity
	if err := errors.WithStack(jsonx.NewStrictDecoder(r.Body).Decode(&i)); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	i.ID = x.ParseUUID(ps.ByName("id"))
	if err := h.r.IdentityPool().UpdateIdentity(r.Context(), &i); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().Write(w, r, i)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.r.IdentityPool().DeleteIdentity(r.Context(), x.ParseUUID(ps.ByName("id"))); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
