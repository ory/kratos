package identity

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/ory/x/pagination"

	"github.com/ory/hive-cloud/hive/x"
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

	// admin.POST("/identities", h.create)
	// admin.PUT("/identities", h.upsert)
	// admin.PUT("/identities/:id", h.update)
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

// swagger:route POST /identities admin createIdentity
//
// Create an identity
//
// Use this endpoint to programmatically create an identity. If an identity with the given RequestID exist already,
// an error (409) will be returned.
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       201: emptyResponse
//       409: genericError
//       500: genericError
// func (h *Handler) create(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	i := NewIdentity()
// 	if err := errors.WithStack(jsonx.NewStrictDecoder(r.Body).Decode(&i)); err != nil {
// 		h.r.Writer().WriteError(w, r, err)
// 		return
// 	}
//
// 	panic("hooks are missing")
// 	if err := errors.WithStack(h.r.IdentityValidator().Validate(i)); err != nil {
// 		h.r.Writer().WriteError(w, r, err)
// 		return
// 	}
//
// 	_, err := h.r.IdentityPool().Create(r.Context(), &i)
// 	if err != nil {
// 		h.r.Writer().WriteError(w, r, err)
// 		return
// 	}
//
// 	h.r.Writer().WriteCreated(w, r,
// 		urlx.AppendPaths(
// 			h.c.SelfAdminURL(),
// 			"identities",
// 			i.ID,
// 		).String(),
// 		http.StatusCreated,
// 	)
// }

// swagger:route PUT /identities admin upsertIdentity
//
// Upsert an identity
//
// Use this endpoint to update or insert an identity. If an identity with the given RequestID exists, the identity
// in the datastore will be overwritten. If such an identity does not exist, the identity will be added to the datastore.
//
// Use this endpoint with caution as it may override an existing identity if the IDs are in conflict.
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: identity
//       500: genericError
// func (h *Handler) upsert(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	i := NewIdentity()
// 	d := json.NewDecoder(r.Body)
// 	d.DisallowUnknownFields()
// 	if err := errors.WithStack(d.Decode(&i)); err != nil {
// 		h.r.Writer().WriteError(w, r, err)
// 		return
// 	}
//
// 	panic("hooks are missing")
// 	if err := errors.WithStack(h.r.IdentityValidator().Validate(&i)); err != nil {
// 		h.r.Writer().WriteError(w, r, err)
// 		return
// 	}
//
// 	_, err := h.r.IdentityPool().Upsert(r.Context(), &i)
// 	if err != nil {
// 		h.r.Writer().WriteError(w, r, err)
// 		return
// 	}
//
// 	h.r.Writer().Write(w, r, &i)
// }

// func (h *Handler) update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	i := NewIdentity()
// 	d := json.NewDecoder(r.Body)
// 	d.DisallowUnknownFields()
// 	if err := errors.WithStack(d.Decode(&i)); err != nil {
// 		h.r.Writer().WriteError(w, r, err)
// 		return
// 	}
//
// 	if i.ID != ps.ByName("id") {
// 		h.r.Writer().WriteError(w, r, errors.WithStack(
// 			herodot.ErrBadRequest.WithReasonf(
// 				"Value of key subject from POST body does not match subject value from URL path: %s != %s",
// 				i.ID, ps.ByName("id"),
// 			),
// 		))
// 		return
// 	}
//
// 	panic("hooks are missing")
// 	if err := errors.WithStack(h.r.IdentityValidator().Validate(&i)); err != nil {
// 		h.r.Writer().WriteError(w, r, err)
// 		return
// 	}
//
// 	_, err := h.r.IdentityPool().Update(r.Context(), &i)
// 	if err != nil {
// 		h.r.Writer().WriteError(w, r, err)
// 		return
// 	}
//
// 	h.r.Writer().Write(w, r, &i)
// }

func (h *Handler) delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.r.IdentityPool().Delete(r.Context(), ps.ByName("id")); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
