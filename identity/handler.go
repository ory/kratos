package identity

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/jsonx"
	"github.com/ory/x/urlx"

	"github.com/ory/x/pagination"

	"github.com/ory/hive/schema"
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

	if i.ID == "" {
		i.ID = uuid.New().String()
	}

	if i.TraitsSchemaURL == "" {
		i.TraitsSchemaURL = h.c.DefaultIdentityTraitsSchemaURL().String()
	}

	if err := h.r.IdentityValidator().Validate(&i); err != nil {
		if _, ok := errors.Cause(err).(schema.ResultErrors); ok {
			h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("%s", err)))
			return
		}
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
		&i,
	)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var i Identity
	if err := errors.WithStack(jsonx.NewStrictDecoder(r.Body).Decode(&i)); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	i.ID = ps.ByName("id")

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
