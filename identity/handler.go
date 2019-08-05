package identity

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/ory/x/jsonx"
	"github.com/ory/x/urlx"
	"github.com/pkg/errors"

	"github.com/ory/x/pagination"

	"github.com/ory/hive/x"
)

type handlerDependencies interface {
	PoolProvider
	ValidationProvider
	x.WriterProvider
}

type Handler struct {
	c Configuration
	r handlerDependencies
}

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
	is, err := h.r.IdentityPool().List(r.Context(), limit, offset)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	for k, i := range is {
		is[k] = *i.WithoutCredentials()
	}

	h.r.Writer().Write(w, r, is)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	i, err := h.r.IdentityPool().Get(r.Context(), ps.ByName("id"))
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().Write(w, r, i.WithoutCredentials())
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var i Identity
	if err := errors.WithStack(jsonx.NewStrictDecoder(r.Body).Decode(&i)); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	if err := errors.WithStack(h.r.IdentityValidator().Validate(&i)); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	_, err := h.r.IdentityPool().Create(r.Context(), (&i).WithoutCredentials())
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().WriteCreated(w, r,
		urlx.AppendPaths(
			h.c.SelfAdminURL(),
			"identities",
			i.ID,
		).String(),
		http.StatusCreated,
	)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var i Identity
	if err := errors.WithStack(jsonx.NewStrictDecoder(r.Body).Decode(&i)); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	if err := errors.WithStack(h.r.IdentityValidator().Validate(&i)); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	_, err := h.r.IdentityPool().Update(r.Context(), (&i).WithoutCredentials())
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().Write(w, r, &i)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.r.IdentityPool().Delete(r.Context(), ps.ByName("id")); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
