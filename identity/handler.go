package identity

import (
	"net/http"

	"github.com/ory/herodot"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/x/jsonx"
	"github.com/ory/x/urlx"

	"github.com/ory/x/pagination"

	"github.com/ory/kratos/x"
)

const IdentitiesPath = "/identities"

type (
	handlerDependencies interface {
		PrivilegedPoolProvider
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
	admin.GET(IdentitiesPath, h.list)
	admin.GET(IdentitiesPath+"/:id", h.get)
	admin.DELETE(IdentitiesPath+"/:id", h.delete)

	admin.POST(IdentitiesPath, h.create)
	admin.PUT(IdentitiesPath+"/:id", h.update)
}

// A single identity.
//
// nolint:deadcode,unused
// swagger:response identityResponse
type identityResponse struct {
	// required: true
	// in: body
	Body *Identity
}

// A list of identities.
//
// nolint:deadcode,unused
// swagger:response identityList
type identitiesListResponse struct {
	// in: body
	// required: true
	// type: array
	Body []Identity
}

// swagger:route GET /identities admin listIdentities
//
// List all identities in the system
//
// This endpoint returns a login request's context with, for example, error details and
// other information.
//
// Learn how identities work in [ORY Kratos' User And Identity Model Documentation](https://www.ory.sh/docs/next/kratos/concepts/identity-user-model).
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: identityList
//       500: genericError
func (h *Handler) list(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	limit, offset := pagination.Parse(r, 100, 0, 500)
	is, err := h.r.PrivilegedIdentityPool().ListIdentities(r.Context(), limit, offset)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().Write(w, r, is)
}

// nolint:deadcode,unused
// swagger:parameters getIdentity
type getIdentityParameters struct {
	// ID must be set to the ID of identity you want to get
	//
	// required: true
	// in: path
	ID string `json:"id"`
}

// swagger:route GET /identities/{id} admin getIdentity
//
// Get an identity
//
// Learn how identities work in [ORY Kratos' User And Identity Model Documentation](https://www.ory.sh/docs/next/kratos/concepts/identity-user-model).
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
//       200: identityResponse
//       400: genericError
//       500: genericError
func (h *Handler) get(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	i, err := h.r.PrivilegedIdentityPool().GetIdentity(r.Context(), x.ParseUUID(ps.ByName("id")))
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().Write(w, r, i)
}

// nolint:deadcode,unused
// swagger:parameters createIdentity
type createIdentityParameters struct {
	// required: true
	// in: body
	Body *Identity
}

// swagger:route POST /identities admin createIdentity
//
// Create an identity
//
// This endpoint creates an identity. It is NOT possible to set an identity's credentials (password, ...)
// using this method! A way to achieve that will be introduced in the future.
//
// Learn how identities work in [ORY Kratos' User And Identity Model Documentation](https://www.ory.sh/docs/next/kratos/concepts/identity-user-model).
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
//       201: identityResponse
//       400: genericError
//       500: genericError
func (h *Handler) create(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var i Identity
	if err := errors.WithStack(jsonx.NewStrictDecoder(r.Body).Decode(&i)); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	// Make sure the TraitsSchemaURL is only set by kratos
	if i.TraitsSchemaURL != "" {
		h.r.Writer().WriteError(w, r, herodot.ErrBadRequest.WithReason("Use the traits_schema_id to set a traits schema."))
		return
	}
	// We do not allow setting credentials using this method
	i.Credentials = nil
	// We do not allow setting the ID using this method
	i.ID = uuid.Nil

	err := h.r.PrivilegedIdentityPool().CreateIdentity(r.Context(), &i)
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

// nolint:deadcode,unused
// swagger:parameters updateIdentity
type updateIdentityParameters struct {
	// ID must be set to the ID of identity you want to update
	//
	// required: true
	// in: path
	ID string `json:"id"`

	// required: true
	// in: body
	Body *Identity
}

// swagger:route PUT /identities/{id} admin updateIdentity
//
// Update an identity
//
// This endpoint updates an identity. It is NOT possible to set an identity's credentials (password, ...)
// using this method! A way to achieve that will be introduced in the future.
//
// The full identity payload (except credentials) is expected. This endpoint does not support patching.
//
// Learn how identities work in [ORY Kratos' User And Identity Model Documentation](https://www.ory.sh/docs/next/kratos/concepts/identity-user-model).
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
//       200: identityResponse
//       400: genericError
//       404: genericError
//       500: genericError
func (h *Handler) update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var i Identity
	if err := errors.WithStack(jsonx.NewStrictDecoder(r.Body).Decode(&i)); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	i.ID = x.ParseUUID(ps.ByName("id"))
	if err := h.r.PrivilegedIdentityPool().UpdateIdentity(r.Context(), &i); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().Write(w, r, i)
}

// nolint:deadcode,unused
// swagger:parameters deleteIdentity
type deleteIdentityParameters struct {
	// ID is the identity's ID.
	//
	// required: true
	// in: path
	ID string `json:"id"`
}

// swagger:route DELETE /identities/{id} admin deleteIdentity
//
// Delete an identity
//
// This endpoint deletes an identity. This can not be undone.
//
// Learn how identities work in [ORY Kratos' User And Identity Model Documentation](https://www.ory.sh/docs/next/kratos/concepts/identity-user-model).
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       204: emptyResponse
//       404: genericError
//       500: genericError
func (h *Handler) delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.r.PrivilegedIdentityPool().DeleteIdentity(r.Context(), x.ParseUUID(ps.ByName("id"))); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
