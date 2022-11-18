// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/ory/x/pagination/migrationpagination"

	"github.com/ory/kratos/hash"
	"github.com/ory/kratos/x"

	"github.com/ory/kratos/cipher"

	"github.com/ory/herodot"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/x/decoderx"
	"github.com/ory/x/jsonx"
	"github.com/ory/x/openapix"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/config"
)

const RouteCollection = "/identities"
const RouteItem = RouteCollection + "/:id"

type (
	handlerDependencies interface {
		PoolProvider
		PrivilegedPoolProvider
		ManagementProvider
		x.WriterProvider
		config.Provider
		x.CSRFProvider
		cipher.Provider
		hash.HashProvider
	}
	HandlerProvider interface {
		IdentityHandler() *Handler
	}
	Handler struct {
		r  handlerDependencies
		dx *decoderx.HTTP
	}
)

func (h *Handler) Config(ctx context.Context) *config.Config {
	return h.r.Config()
}

func NewHandler(r handlerDependencies) *Handler {
	return &Handler{
		r:  r,
		dx: decoderx.NewHTTP(),
	}
}

func (h *Handler) RegisterPublicRoutes(public *x.RouterPublic) {
	h.r.CSRFHandler().IgnoreGlobs(
		RouteCollection, RouteCollection+"/*",
		x.AdminPrefix+RouteCollection, x.AdminPrefix+RouteCollection+"/*",
	)

	public.GET(RouteCollection, x.RedirectToAdminRoute(h.r))
	public.GET(RouteItem, x.RedirectToAdminRoute(h.r))
	public.DELETE(RouteItem, x.RedirectToAdminRoute(h.r))
	public.POST(RouteCollection, x.RedirectToAdminRoute(h.r))
	public.PUT(RouteItem, x.RedirectToAdminRoute(h.r))
	public.PATCH(RouteItem, x.RedirectToAdminRoute(h.r))

	public.GET(x.AdminPrefix+RouteCollection, x.RedirectToAdminRoute(h.r))
	public.GET(x.AdminPrefix+RouteItem, x.RedirectToAdminRoute(h.r))
	public.DELETE(x.AdminPrefix+RouteItem, x.RedirectToAdminRoute(h.r))
	public.POST(x.AdminPrefix+RouteCollection, x.RedirectToAdminRoute(h.r))
	public.PUT(x.AdminPrefix+RouteItem, x.RedirectToAdminRoute(h.r))
	public.PATCH(x.AdminPrefix+RouteItem, x.RedirectToAdminRoute(h.r))
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(RouteCollection, h.list)
	admin.GET(RouteItem, h.get)
	admin.DELETE(RouteItem, h.delete)
	admin.PATCH(RouteItem, h.patch)

	admin.POST(RouteCollection, h.create)
	admin.PUT(RouteItem, h.update)
}

// Paginated Identity List Response
//
// swagger:response listIdentities
// nolint:deadcode,unused
type listIdentitiesResponse struct {
	migrationpagination.ResponseHeaderAnnotation

	// List of identities
	//
	// in:body
	Body []Identity
}

// Paginated List Identity Parameters
//
// swagger:parameters listIdentities
// nolint:deadcode,unused
type listIdentitiesParameters struct {
	migrationpagination.RequestParameters
}

// swagger:route GET /admin/identities identity listIdentities
//
// # List Identities
//
// Lists all [identities](https://www.ory.sh/docs/kratos/concepts/identity-user-model) in the system.
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Security:
//	  oryAccessToken:
//
//	Responses:
//	  200: listIdentities
//	  default: errorGeneric
func (h *Handler) list(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	page, itemsPerPage := x.ParsePagination(r)
	is, err := h.r.IdentityPool().ListIdentities(r.Context(), page, itemsPerPage)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	total, err := h.r.IdentityPool().CountIdentities(r.Context())
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	// Identities using the marshaler for including metadata_admin
	isam := make([]WithAdminMetadataInJSON, len(is))
	for i, identity := range is {
		isam[i] = WithAdminMetadataInJSON(identity)
	}

	migrationpagination.PaginationHeader(w, urlx.AppendPaths(h.r.Config().SelfAdminURL(r.Context()), RouteCollection), total, page, itemsPerPage)
	h.r.Writer().Write(w, r, isam)
}

// Get Identity Parameters
//
// swagger:parameters getIdentity
// nolint:deadcode,unused
type getIdentity struct {
	// ID must be set to the ID of identity you want to get
	//
	// required: true
	// in: path
	ID string `json:"id"`

	// Include Credentials in Response
	//
	// Currently, only `oidc` is supported. This will return the initial OAuth 2.0 Access,
	// Refresh and (optionally) OpenID Connect ID Token.
	//
	// required: false
	// in: query
	DeclassifyCredentials []string `json:"include_credential"`
}

// swagger:route GET /admin/identities/{id} identity getIdentity
//
// # Get an Identity
//
// Return an [identity](https://www.ory.sh/docs/kratos/concepts/identity-user-model) by its ID. You can optionally
// include credentials (e.g. social sign in connections) in the response by using the `include_credential` query parameter.
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Security:
//	  oryAccessToken:
//
//	Responses:
//	  200: identity
//	  404: errorGeneric
//	  default: errorGeneric
func (h *Handler) get(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	i, err := h.r.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), x.ParseUUID(ps.ByName("id")))
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	if declassify := r.URL.Query().Get("include_credential"); declassify == "oidc" {
		emit, err := i.WithDeclassifiedCredentialsOIDC(r.Context(), h.r)
		if err != nil {
			h.r.Writer().WriteError(w, r, err)
			return
		}
		h.r.Writer().Write(w, r, WithCredentialsAndAdminMetadataInJSON(*emit))
		return
	} else if len(declassify) > 0 {
		h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Invalid value `%s` for parameter `include_credential`.", declassify)))
		return

	}

	h.r.Writer().Write(w, r, WithCredentialsMetadataAndAdminMetadataInJSON(*i))
}

// Create Identity Parameters
//
// swagger:parameters createIdentity
// nolint:deadcode,unused
type createIdentity struct {
	// in: body
	Body CreateIdentityBody
}

// Create Identity Body
//
// swagger:model createIdentityBody
type CreateIdentityBody struct {
	// SchemaID is the ID of the JSON Schema to be used for validating the identity's traits.
	//
	// required: true
	SchemaID string `json:"schema_id"`

	// Traits represent an identity's traits. The identity is able to create, modify, and delete traits
	// in a self-service manner. The input will always be validated against the JSON Schema defined
	// in `schema_url`.
	//
	// required: true
	Traits json.RawMessage `json:"traits"`

	// Credentials represents all credentials that can be used for authenticating this identity.
	//
	// Use this structure to import credentials for a user.
	Credentials *IdentityWithCredentials `json:"credentials"`

	// VerifiableAddresses contains all the addresses that can be verified by the user.
	//
	// Use this structure to import verified addresses for an identity. Please keep in mind
	// that the address needs to be represented in the Identity Schema or this field will be overwritten
	// on the next identity update.
	VerifiableAddresses []VerifiableAddress `json:"verifiable_addresses"`

	// RecoveryAddresses contains all the addresses that can be used to recover an identity.
	//
	// Use this structure to import recovery addresses for an identity. Please keep in mind
	// that the address needs to be represented in the Identity Schema or this field will be overwritten
	// on the next identity update.
	RecoveryAddresses []RecoveryAddress `json:"recovery_addresses"`

	// Store metadata about the identity which the identity itself can see when calling for example the
	// session endpoint. Do not store sensitive information (e.g. credit score) about the identity in this field.
	MetadataPublic json.RawMessage `json:"metadata_public"`

	// Store metadata about the user which is only accessible through admin APIs such as `GET /admin/identities/<id>`.
	MetadataAdmin json.RawMessage `json:"metadata_admin,omitempty"`

	// State is the identity's state.
	//
	// required: false
	State State `json:"state"`
}

// Create Identity and Import Credentials
//
// swagger:model identityWithCredentials
type IdentityWithCredentials struct {
	// Password if set will import a password credential.
	Password *AdminIdentityImportCredentialsPassword `json:"password"`

	// OIDC if set will import an OIDC credential.
	OIDC *AdminIdentityImportCredentialsOIDC `json:"oidc"`
}

// Create Identity and Import Password Credentials
//
// swagger:model identityWithCredentialsPassword
type AdminIdentityImportCredentialsPassword struct {
	// Configuration options for the import.
	Config AdminIdentityImportCredentialsPasswordConfig `json:"config"`
}

// Create Identity and Import Password Credentials Configuration
//
// swagger:model identityWithCredentialsPasswordConfig
type AdminIdentityImportCredentialsPasswordConfig struct {
	// The hashed password in [PHC format]( https://www.ory.sh/docs/kratos/concepts/credentials/username-email-password#hashed-password-format)
	HashedPassword string `json:"hashed_password"`

	// The password in plain text if no hash is available.
	Password string `json:"password"`
}

// Create Identity and Import Social Sign In Credentials
//
// swagger:model identityWithCredentialsOidc
type AdminIdentityImportCredentialsOIDC struct {
	// Configuration options for the import.
	Config AdminIdentityImportCredentialsOIDCConfig `json:"config"`
}

// swagger:model identityWithCredentialsOidcConfig
type AdminIdentityImportCredentialsOIDCConfig struct {
	// Configuration options for the import.
	Config AdminIdentityImportCredentialsPasswordConfig `json:"config"`
	// A list of OpenID Connect Providers
	Providers []AdminCreateIdentityImportCredentialsOidcProvider `json:"providers"`
}

// Create Identity and Import Social Sign In Credentials Configuration
//
// swagger:model identityWithCredentialsOidcConfigProvider
type AdminCreateIdentityImportCredentialsOidcProvider struct {
	// The subject (`sub`) of the OpenID Connect connection. Usually the `sub` field of the ID Token.
	//
	// required: true
	Subject string `json:"subject"`

	// The OpenID Connect provider to link the subject to. Usually something like `google` or `github`.
	//
	// required: true
	Provider string `json:"provider"`
}

// swagger:route POST /admin/identities identity createIdentity
//
// # Create an Identity
//
// Create an [identity](https://www.ory.sh/docs/kratos/concepts/identity-user-model).  This endpoint can also be used to
// [import credentials](https://www.ory.sh/docs/kratos/manage-identities/import-user-accounts-identities)
// for instance passwords, social sign in configurations or multifactor methods.
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Security:
//	  oryAccessToken:
//
//	Responses:
//	  201: identity
//	  400: errorGeneric
//	  409: errorGeneric
//	  default: errorGeneric
func (h *Handler) create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var cr CreateIdentityBody
	if err := jsonx.NewStrictDecoder(r.Body).Decode(&cr); err != nil {
		h.r.Writer().WriteErrorCode(w, r, http.StatusBadRequest, errors.WithStack(err))
		return
	}

	stateChangedAt := sqlxx.NullTime(time.Now())
	state := StateActive
	if cr.State != "" {
		if err := cr.State.IsValid(); err != nil {
			h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("%s", err).WithWrap(err)))
			return
		}
		state = cr.State
	}

	i := &Identity{
		SchemaID:            cr.SchemaID,
		Traits:              []byte(cr.Traits),
		State:               state,
		StateChangedAt:      &stateChangedAt,
		VerifiableAddresses: cr.VerifiableAddresses,
		RecoveryAddresses:   cr.RecoveryAddresses,
		MetadataAdmin:       []byte(cr.MetadataAdmin),
		MetadataPublic:      []byte(cr.MetadataPublic),
	}

	if err := h.importCredentials(r.Context(), i, cr.Credentials); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	if err := h.r.IdentityManager().Create(r.Context(), i); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().WriteCreated(w, r,
		urlx.AppendPaths(
			h.r.Config().SelfAdminURL(r.Context()),
			"identities",
			i.ID.String(),
		).String(),
		WithCredentialsMetadataAndAdminMetadataInJSON(*i),
	)
}

// Update Identity Parameters
//
// swagger:parameters updateIdentity
// nolint:deadcode,unused
type updateIdentity struct {
	// ID must be set to the ID of identity you want to update
	//
	// required: true
	// in: path
	ID string `json:"id"`

	// in: body
	Body UpdateIdentityBody
}

// Update Identity Body
//
// swagger:model updateIdentityBody
type UpdateIdentityBody struct {
	// SchemaID is the ID of the JSON Schema to be used for validating the identity's traits. If set
	// will update the Identity's SchemaID.
	//
	// required: true
	SchemaID string `json:"schema_id"`

	// Traits represent an identity's traits. The identity is able to create, modify, and delete traits
	// in a self-service manner. The input will always be validated against the JSON Schema defined
	// in `schema_id`.
	//
	// required: true
	Traits json.RawMessage `json:"traits"`

	// Credentials represents all credentials that can be used for authenticating this identity.
	//
	// Use this structure to import credentials for a user.
	// Note: this wil override completely identity's credentials. If used incorrectly, this can cause a user to lose
	// access to their account!
	Credentials *IdentityWithCredentials `json:"credentials"`

	// Store metadata about the identity which the identity itself can see when calling for example the
	// session endpoint. Do not store sensitive information (e.g. credit score) about the identity in this field.
	MetadataPublic json.RawMessage `json:"metadata_public"`

	// Store metadata about the user which is only accessible through admin APIs such as `GET /admin/identities/<id>`.
	MetadataAdmin json.RawMessage `json:"metadata_admin,omitempty"`

	// State is the identity's state.
	//
	// required: true
	State State `json:"state"`
}

// swagger:route PUT /admin/identities/{id} identity updateIdentity
//
// # Update an Identity
//
// This endpoint updates an [identity](https://www.ory.sh/docs/kratos/concepts/identity-user-model). The full identity
// payload (except credentials) is expected. It is possible to update the identity's credentials as well.
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Security:
//	  oryAccessToken:
//
//	Responses:
//	  200: identity
//	  400: errorGeneric
//	  404: errorGeneric
//	  409: errorGeneric
//	  default: errorGeneric
func (h *Handler) update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var ur UpdateIdentityBody
	if err := h.dx.Decode(r, &ur,
		decoderx.HTTPJSONDecoder()); err != nil {
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

	if ur.State != "" && identity.State != ur.State {
		if err := ur.State.IsValid(); err != nil {
			h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("%s", err).WithWrap(err)))
			return
		}

		stateChangedAt := sqlxx.NullTime(time.Now())

		identity.State = ur.State
		identity.StateChangedAt = &stateChangedAt
	}

	identity.Traits = []byte(ur.Traits)
	identity.MetadataPublic = []byte(ur.MetadataPublic)
	identity.MetadataAdmin = []byte(ur.MetadataAdmin)

	// Although this is PUT and not PATCH, if the Credentials are not supplied keep the old one
	if ur.Credentials != nil {
		if err := h.importCredentials(r.Context(), identity, ur.Credentials); err != nil {
			h.r.Writer().WriteError(w, r, err)
			return
		}
	}

	if err := h.r.IdentityManager().Update(
		r.Context(),
		identity,
		ManagerAllowWriteProtectedTraits,
	); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().Write(w, r, WithCredentialsMetadataAndAdminMetadataInJSON(*identity))
}

// Delete Identity Parameters
//
// swagger:parameters deleteIdentity
// nolint:deadcode,unused
type deleteIdentity struct {
	// ID is the identity's ID.
	//
	// required: true
	// in: path
	ID string `json:"id"`
}

// swagger:route DELETE /admin/identities/{id} identity deleteIdentity
//
// # Delete an Identity
//
// Calling this endpoint irrecoverably and permanently deletes the [identity](https://www.ory.sh/docs/kratos/concepts/identity-user-model) given its ID. This action can not be undone.
// This endpoint returns 204 when the identity was deleted or when the identity was not found, in which case it is
// assumed that is has been deleted already.
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Security:
//	  oryAccessToken:
//
//	Responses:
//	  204: emptyResponse
//	  404: errorGeneric
//	  default: errorGeneric
func (h *Handler) delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.r.IdentityPool().(PrivilegedPool).DeleteIdentity(r.Context(), x.ParseUUID(ps.ByName("id"))); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Patch Identity Parameters
//
// swagger:parameters patchIdentity
// nolint:deadcode,unused
type patchIdentity struct {
	// ID must be set to the ID of identity you want to update
	//
	// required: true
	// in: path
	ID string `json:"id"`

	// in: body
	Body openapix.JSONPatchDocument
}

// swagger:route PATCH /admin/identities/{id} identity patchIdentity
//
// # Patch an Identity
//
// Partially updates an [identity's](https://www.ory.sh/docs/kratos/concepts/identity-user-model) field using [JSON Patch](https://jsonpatch.com/).
// The fields `id`, `stateChangedAt` and `credentials` can not be updated using this method.
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Security:
//	  oryAccessToken:
//
//	Responses:
//	  200: identity
//	  400: errorGeneric
//	  404: errorGeneric
//	  409: errorGeneric
//	  default: errorGeneric
func (h *Handler) patch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	id := x.ParseUUID(ps.ByName("id"))
	identity, err := h.r.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), id)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	credentials := identity.Credentials
	oldState := identity.State

	if err := jsonx.ApplyJSONPatch(requestBody, identity, "/id", "/stateChangedAt", "/credentials"); err != nil {
		h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("%s", err).WithWrap(err)))
		return
	}

	// See https://github.com/ory/cloud/issues/148
	// The apply patch operation overrides the credentials with an empty map.
	identity.Credentials = credentials

	if oldState != identity.State {
		// Check if the changed state was actually valid
		if err := identity.State.IsValid(); err != nil {
			h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("%s", err).WithWrap(err)))
			return
		}

		// If the state changed, we need to update the timestamp of it
		stateChangedAt := sqlxx.NullTime(time.Now())
		identity.StateChangedAt = &stateChangedAt
	}

	if err := h.r.IdentityManager().Update(
		r.Context(),
		identity,
		ManagerAllowWriteProtectedTraits,
	); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().Write(w, r, WithCredentialsMetadataAndAdminMetadataInJSON(*identity))
}
