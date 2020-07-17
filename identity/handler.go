// nolint:deadcode,unused
package identity

import (
	"encoding/json"
	"net/http"

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
		PoolProvider
		PrivilegedPoolProvider
		ManagementProvider
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
// swagger:response identityResponse
type identityResponse struct {
	// required: true
	// in: body
	Body *Identity
}

// A list of identities.
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
	is, err := h.r.IdentityPool().ListIdentities(r.Context(), limit, offset)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().Write(w, r, is)
}

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
	i, err := h.r.IdentityPool().GetIdentity(r.Context(), x.ParseUUID(ps.ByName("id")))
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().Write(w, r, i)
}

// swagger:parameters createIdentity
type createIdentityRequest struct {
	// in: body
	Body CreateIdentityRequestPayload
}

type CreateIdentityRequestPayload struct {
	// SchemaID is the ID of the JSON Schema to be used for validating the identity's traits.
	//
	// required: true
	// in: body
	SchemaID string `json:"schema_id"`

	// Traits represent an identity's traits. The identity is able to create, modify, and delete traits
	// in a self-service manner. The input will always be validated against the JSON Schema defined
	// in `schema_url`.
	//
	// required: true
	// in: body
	Traits json.RawMessage `json:"traits"`
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
func (h *Handler) create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var cr CreateIdentityRequestPayload
	if err := jsonx.NewStrictDecoder(r.Body).Decode(&cr); err != nil {
		h.r.Writer().WriteErrorCode(w, r, http.StatusBadRequest, errors.WithStack(err))
		return
	}

	i := &Identity{SchemaID: cr.SchemaID, Traits: []byte(cr.Traits)}
	if err := h.r.IdentityManager().Create(r.Context(), i); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().WriteCreated(w, r,
		urlx.AppendPaths(
			h.c.SelfAdminURL(),
			"identities",
			i.ID.String(),
		).String(),
		i,
	)
}

// swagger:parameters updateIdentity
type updateIdentityRequest struct {
	// in: body
	Body UpdateIdentityRequestPayload
}

type UpdateIdentityRequestPayload struct {
	// SchemaID is the ID of the JSON Schema to be used for validating the identity's traits. If set
	// will update the Identity's SchemaID.
	SchemaID string `json:"schema_id"`

	// Traits represent an identity's traits. The identity is able to create, modify, and delete traits
	// in a self-service manner. The input will always be validated against the JSON Schema defined
	// in `schema_id`.
	//
	// required: true
	Traits json.RawMessage `json:"traits"`
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
	var ur UpdateIdentityRequestPayload
	if err := errors.WithStack(jsonx.NewStrictDecoder(r.Body).Decode(&ur)); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	id := x.ParseUUID(ps.ByName("id"))
	identity, err := h.r.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), id)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	if ur.SchemaID != "" {
		identity.SchemaID = ur.SchemaID
	}

	identity.Traits = []byte(ur.Traits)
	if err := h.r.IdentityManager().Update(
		r.Context(),
		identity,
		ManagerAllowWriteProtectedTraits,
	); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().Write(w, r, identity)
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
	if err := h.r.IdentityPool().(PrivilegedPool).DeleteIdentity(r.Context(), x.ParseUUID(ps.ByName("id"))); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
