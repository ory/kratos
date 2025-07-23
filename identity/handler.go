// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ory/kratos/x/nosurfx"
	"github.com/ory/kratos/x/redir"

	"github.com/gofrs/uuid"

	"github.com/ory/x/crdbx"
	"github.com/ory/x/pagination/keysetpagination"

	"github.com/ory/x/pagination/migrationpagination"
	"github.com/ory/x/pagination/pagepagination"
	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/hash"
	"github.com/ory/kratos/x"

	"github.com/ory/kratos/cipher"

	"github.com/ory/herodot"

	"github.com/pkg/errors"

	"github.com/ory/x/decoderx"
	"github.com/ory/x/jsonx"
	"github.com/ory/x/openapix"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/config"
)

const (
	RouteCollection     = "/identities"
	RouteItem           = RouteCollection + "/{id}"
	RouteCredentialItem = RouteItem + "/credentials/{type}"

	BatchPatchIdentitiesLimit             = 1000
	BatchPatchIdentitiesWithPasswordLimit = 200
)

type (
	handlerDependencies interface {
		PoolProvider
		PrivilegedPoolProvider
		ManagementProvider
		x.WriterProvider
		config.Provider
		nosurfx.CSRFProvider
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
		RouteCollection+"/*/credentials/*",
		x.AdminPrefix+RouteCollection, x.AdminPrefix+RouteCollection+"/*",
		x.AdminPrefix+RouteCollection+"/*/credentials/*",
	)

	public.GET(RouteCollection, redir.RedirectToAdminRoute(h.r))
	public.GET(RouteCollection+"/by/external/{externalID}", redir.RedirectToAdminRoute(h.r))
	public.GET(RouteItem, redir.RedirectToAdminRoute(h.r))
	public.DELETE(RouteItem, redir.RedirectToAdminRoute(h.r))
	public.POST(RouteCollection, redir.RedirectToAdminRoute(h.r))
	public.PUT(RouteItem, redir.RedirectToAdminRoute(h.r))
	public.PATCH(RouteItem, redir.RedirectToAdminRoute(h.r))
	public.DELETE(RouteCredentialItem, redir.RedirectToAdminRoute(h.r))

	public.GET(x.AdminPrefix+RouteCollection, redir.RedirectToAdminRoute(h.r))
	public.GET(x.AdminPrefix+RouteCollection+"/by/external/{externalID}", redir.RedirectToAdminRoute(h.r))
	public.GET(x.AdminPrefix+RouteItem, redir.RedirectToAdminRoute(h.r))
	public.DELETE(x.AdminPrefix+RouteItem, redir.RedirectToAdminRoute(h.r))
	public.POST(x.AdminPrefix+RouteCollection, redir.RedirectToAdminRoute(h.r))
	public.PUT(x.AdminPrefix+RouteItem, redir.RedirectToAdminRoute(h.r))
	public.PATCH(x.AdminPrefix+RouteItem, redir.RedirectToAdminRoute(h.r))
	public.DELETE(x.AdminPrefix+RouteCredentialItem, redir.RedirectToAdminRoute(h.r))
}

func (h *Handler) RegisterAdminRoutes(admin *x.RouterAdmin) {
	admin.GET(RouteCollection, h.list)
	admin.GET(RouteItem, h.get)
	admin.GET(RouteCollection+"/by/external/{externalID}", h.getByExternalID)
	admin.DELETE(RouteItem, h.delete)
	admin.PATCH(RouteItem, h.patch)

	admin.POST(RouteCollection, h.create)
	admin.PATCH(RouteCollection, h.batchPatchIdentities)
	admin.PUT(RouteItem, h.update)

	admin.DELETE(RouteCredentialItem, h.deleteIdentityCredentials)
}

// Paginated Identity List Response
//
// swagger:response listIdentities
type _ struct {
	migrationpagination.ResponseHeaderAnnotation

	// List of identities
	//
	// in:body
	Body []Identity
}

// Paginated List Identity Parameters
//
// Note: Filters cannot be combined.
//
// swagger:parameters listIdentities
type _ struct {
	migrationpagination.RequestParameters

	// Retrieve multiple identities by their IDs.
	//
	// This parameter has the following limitations:
	//
	// - Duplicate or non-existent IDs are ignored.
	// - The order of returned IDs may be different from the request.
	// - This filter does not support pagination. You must implement your own pagination as the maximum number of items returned by this endpoint may not exceed a certain threshold (currently 500).
	//
	// required: false
	// in: query
	IdsFilter []string `json:"ids"`

	// CredentialsIdentifier is the identifier (username, email) of the credentials to look up using exact match.
	// Only one of CredentialsIdentifier and CredentialsIdentifierSimilar can be used.
	//
	// required: false
	// in: query
	CredentialsIdentifier string `json:"credentials_identifier"`

	// This is an EXPERIMENTAL parameter that WILL CHANGE. Do NOT rely on consistent, deterministic behavior.
	// THIS PARAMETER WILL BE REMOVED IN AN UPCOMING RELEASE WITHOUT ANY MIGRATION PATH.
	//
	// CredentialsIdentifierSimilar is the (partial) identifier (username, email) of the credentials to look up using similarity search.
	// Only one of CredentialsIdentifier and CredentialsIdentifierSimilar can be used.
	//
	// required: false
	// in: query
	CredentialsIdentifierSimilar string `json:"preview_credentials_identifier_similar"`

	// Include Credentials in Response
	//
	// Include any credential, for example `password` or `oidc`, in the response. When set to `oidc`, This will return
	// the initial OAuth 2.0 Access Token, OAuth 2.0 Refresh Token and the OpenID Connect ID Token if available.
	//
	// required: false
	// in: query
	DeclassifyCredentials []string `json:"include_credential"`

	// List identities that belong to a specific organization.
	//
	// required: false
	// in: query
	OrganizationID string `json:"organization_id"`

	crdbx.ConsistencyRequestParameters
}

func parseListIdentitiesParameters(r *http.Request) (params ListIdentityParameters, err error) {
	query := r.URL.Query()
	var requestedFilters int

	params.Expand = ExpandDefault

	if ids := query["ids"]; len(ids) > 0 {
		requestedFilters++
		for _, v := range ids {
			id, err := uuid.FromString(v)
			if err != nil {
				return params, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Invalid UUID value `%s` for parameter `ids`.", v))
			}
			params.IdsFilter = append(params.IdsFilter, id)
		}
	}
	if len(params.IdsFilter) > 500 {
		return params, errors.WithStack(herodot.ErrBadRequest.WithReason("The number of ids to filter must not exceed 500."))
	}

	if orgID := query.Get("organization_id"); orgID != "" {
		requestedFilters++
		params.OrganizationID, err = uuid.FromString(orgID)
		if err != nil {
			return params, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Invalid UUID value `%s` for parameter `organization_id`.", orgID))
		}
	}

	if identifier := query.Get("credentials_identifier"); identifier != "" {
		requestedFilters++
		params.Expand = ExpandEverything
		params.CredentialsIdentifier = identifier
	}

	if identifier := query.Get("preview_credentials_identifier_similar"); identifier != "" {
		requestedFilters++
		params.Expand = ExpandEverything
		params.CredentialsIdentifierSimilar = identifier
	}

	for _, v := range query["include_credential"] {
		params.Expand = ExpandEverything
		tc, ok := ParseCredentialsType(v)
		if !ok {
			return params, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Invalid value `%s` for parameter `include_credential`.", v))
		}
		params.DeclassifyCredentials = append(params.DeclassifyCredentials, tc)
	}

	if requestedFilters > 1 {
		return params, errors.WithStack(herodot.ErrBadRequest.WithReason("You cannot combine multiple filters in this API"))
	}

	params.KeySetPagination, params.PagePagination, err = x.ParseKeysetOrPagePagination(r)
	if err != nil {
		return params, err
	}
	params.ConsistencyLevel = crdbx.ConsistencyLevelFromRequest(r)

	return params, nil
}

// swagger:route GET /admin/identities identity listIdentities
//
// # List Identities
//
// Lists all [identities](https://www.ory.sh/docs/kratos/concepts/identity-user-model) in the system. Note: filters cannot be combined.
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
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	params, err := parseListIdentitiesParameters(r)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	is, nextPage, err := h.r.IdentityPool().ListIdentities(r.Context(), params)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	if params.PagePagination != nil {
		total := int64(len(is))
		if params.CredentialsIdentifier == "" {
			total, err = h.r.IdentityPool().CountIdentities(r.Context())
			if err != nil {
				h.r.Writer().WriteError(w, r, err)
				return
			}
		}
		u := *r.URL
		pagepagination.PaginationHeader(w, &u, total, params.PagePagination.Page, params.PagePagination.ItemsPerPage)
	} else if nextPage != nil {
		u := *r.URL
		keysetpagination.Header(w, &u, nextPage)
	}

	// Identities using the marshaler for including metadata_admin
	isam := make([]WithCredentialsAndAdminMetadataInJSON, len(is))
	for i, identity := range is {
		emit, err := identity.WithDeclassifiedCredentials(r.Context(), h.r, params.DeclassifyCredentials)
		if err != nil {
			h.r.Writer().WriteError(w, r, err)
			return
		}

		isam[i] = WithCredentialsAndAdminMetadataInJSON(*emit)
	}

	h.r.Writer().Write(w, r, isam)
}

// Get Identity Parameters
//
// swagger:parameters getIdentity
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type getIdentity struct {
	// ID must be set to the ID of identity you want to get
	//
	// required: true
	// in: path
	ID string `json:"id"`

	// Include Credentials in Response
	//
	// Include any credential, for example `password` or `oidc`, in the response. When set to `oidc`, This will return
	// the initial OAuth 2.0 Access Token, OAuth 2.0 Refresh Token and the OpenID Connect ID Token if available.
	//
	// required: false
	// in: query
	DeclassifyCredentials []CredentialsType `json:"include_credential"`
}

// Get Identity By External ID Parameters
//
// swagger:parameters getIdentityByExternalID
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
type getIdentityByExternalID struct {
	// ExternalID must be set to the ID of identity you want to get
	//
	// required: true
	// in: path
	ExternalID string `json:"externalID"`

	// Include Credentials in Response
	//
	// Include any credential, for example `password` or `oidc`, in the response. When set to `oidc`, This will return
	// the initial OAuth 2.0 Access Token, OAuth 2.0 Refresh Token and the OpenID Connect ID Token if available.
	//
	// required: false
	// in: query
	DeclassifyCredentials []CredentialsType `json:"include_credential"`
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
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	i, err := h.r.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), x.ParseUUID(r.PathValue("id")))
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	includeCredentials := r.URL.Query()["include_credential"]
	var declassify []CredentialsType
	for _, v := range includeCredentials {
		tc, ok := ParseCredentialsType(v)
		if ok {
			declassify = append(declassify, tc)
		} else {
			h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Invalid value `%s` for parameter `include_credential`.", declassify)))
			return
		}
	}

	emit, err := i.WithDeclassifiedCredentials(r.Context(), h.r, declassify)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}
	h.r.Writer().Write(w, r, WithCredentialsAndAdminMetadataInJSON(*emit))
}

// swagger:route GET /admin/identities/by/external/{externalID} identity getIdentityByExternalID
//
// # Get an Identity by its External ID
//
// Return an [identity](https://www.ory.sh/docs/kratos/concepts/identity-user-model) by its external ID. You can optionally
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
func (h *Handler) getByExternalID(w http.ResponseWriter, r *http.Request) {
	externalID := r.PathValue("externalID")
	if externalID == "" {
		h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReason("The external ID must not be empty.")))
		return
	}
	i, err := h.r.PrivilegedIdentityPool().FindIdentityByExternalID(r.Context(), externalID, ExpandEverything)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	includeCredentials := r.URL.Query()["include_credential"]
	var declassify []CredentialsType
	for _, v := range includeCredentials {
		tc, ok := ParseCredentialsType(v)
		if ok {
			declassify = append(declassify, tc)
		} else {
			h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Invalid value `%s` for parameter `include_credential`.", declassify)))
			return
		}
	}

	emit, err := i.WithDeclassifiedCredentials(r.Context(), h.r, declassify)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}
	h.r.Writer().Write(w, r, WithCredentialsAndAdminMetadataInJSON(*emit))
}

// Create Identity Parameters
//
// swagger:parameters createIdentity
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
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

	// OrganizationID is the ID of the organization to which the identity belongs.
	//
	// required: false
	OrganizationID uuid.NullUUID `json:"organization_id"`

	// ExternalID is an optional external ID of the identity. This is used to link
	// the identity to an external system. If set, the external ID must be unique
	// across all identities.
	//
	// required: false
	ExternalID string `json:"external_id,omitempty"`
}

// Create Identity and Import Credentials
//
// swagger:model identityWithCredentials
type IdentityWithCredentials struct {
	// Password if set will import a password credential.
	Password *AdminIdentityImportCredentialsPassword `json:"password"`

	// OIDC if set will import an OIDC credential.
	OIDC *AdminIdentityImportCredentialsOIDC `json:"oidc"`

	// OIDC if set will import an OIDC credential.
	SAML *AdminIdentityImportCredentialsSAML `json:"saml"`
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
	// The hashed password in [PHC format](https://www.ory.sh/docs/kratos/manage-identities/import-user-accounts-identities#hashed-passwords)
	HashedPassword string `json:"hashed_password"`

	// The password in plain text if no hash is available.
	Password string `json:"password"`

	// If set to true, the password will be migrated using the password migration hook.
	UsePasswordMigrationHook bool `json:"use_password_migration_hook,omitempty"`
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
	// A list of OpenID Connect Providers
	Providers []AdminCreateIdentityImportCredentialsOIDCProvider `json:"providers"`
}

// Create Identity and Import Social Sign In Credentials Configuration
//
// swagger:model identityWithCredentialsOidcConfigProvider
type AdminCreateIdentityImportCredentialsOIDCProvider struct {
	// The subject (`sub`) of the OpenID Connect connection. Usually the `sub` field of the ID Token.
	//
	// required: true
	Subject string `json:"subject"`

	// The OpenID Connect provider to link the subject to. Usually something like `google` or `github`.
	//
	// required: true
	Provider string `json:"provider"`

	// If set, this credential allows the user to sign in using the OpenID Connect provider without setting the subject first.
	//
	// required: false
	UseAutoLink bool `json:"use_auto_link,omitempty"`

	// The organization to assign for the provider.
	Organization uuid.NullUUID `json:"organization,omitempty"`
}

// Payload to import SAML credentials
//
// swagger:model identityWithCredentialsSaml
type AdminIdentityImportCredentialsSAML struct {
	// Configuration options for the import.
	Config AdminIdentityImportCredentialsSAMLConfig `json:"config"`
}

// Payload of SAML providers
//
// swagger:model identityWithCredentialsSamlConfig
type AdminIdentityImportCredentialsSAMLConfig struct {
	// A list of SAML Providers
	Providers []AdminCreateIdentityImportCredentialsSAMLProvider `json:"providers"`
}

// Payload of specific SAML provider
//
// swagger:model identityWithCredentialsSamlConfigProvider
type AdminCreateIdentityImportCredentialsSAMLProvider struct {
	// The unique subject of the SAML connection. This value must be immutable at the source.
	//
	// required: true
	Subject string `json:"subject"`

	// The SAML provider to link the subject to.
	//
	// required: true
	Provider string `json:"provider"`

	// The organization to assign for the provider.
	Organization uuid.NullUUID `json:"organization"`
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
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var cr CreateIdentityBody
	if err := jsonx.NewStrictDecoder(r.Body).Decode(&cr); err != nil {
		h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithError(err.Error())))
		return
	}

	i, err := h.identityFromCreateIdentityBody(r.Context(), &cr)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	if err := h.r.IdentityManager().Create(r.Context(), i); err != nil {
		if errors.Is(err, sqlcon.ErrUniqueViolation) {
			h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrConflict.WithReason("This identity conflicts with another identity that already exists.")))
		} else {
			h.r.Writer().WriteError(w, r, err)
		}
		return
	}

	h.r.Writer().WriteCreated(w, r,
		urlx.AppendPaths(
			h.r.Config().SelfAdminURL(r.Context()),
			"identities",
			i.ID.String(),
		).String(),
		WithCredentialsNoConfigAndAdminMetadataInJSON(*i),
	)
}

func (h *Handler) identityFromCreateIdentityBody(ctx context.Context, cr *CreateIdentityBody) (*Identity, error) {
	stateChangedAt := sqlxx.NullTime(time.Now())
	state := StateActive
	if cr.State != "" {
		if err := cr.State.IsValid(); err != nil {
			return nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("%s", err).WithWrap(err))
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
		OrganizationID:      cr.OrganizationID,
		ExternalID:          sqlxx.NullString(cr.ExternalID),
	}
	// Lowercase all emails, because the schema extension will otherwise not find them.
	for k := range i.VerifiableAddresses {
		i.VerifiableAddresses[k].Value = strings.ToLower(i.VerifiableAddresses[k].Value)
	}
	for k := range i.RecoveryAddresses {
		i.RecoveryAddresses[k].Value = strings.ToLower(i.RecoveryAddresses[k].Value)
	}

	if err := h.importCredentials(ctx, i, cr.Credentials); err != nil {
		return nil, err
	}

	return i, nil
}

// swagger:route PATCH /admin/identities identity batchPatchIdentities
//
// # Create multiple identities
//
// Creates multiple [identities](https://www.ory.sh/docs/kratos/concepts/identity-user-model).
//
// You can also use this endpoint to [import credentials](https://www.ory.sh/docs/kratos/manage-identities/import-user-accounts-identities),
// including passwords, social sign-in settings, and multi-factor authentication methods.
//
// You can import:
// - Up to 1,000 identities per request
// - Up to 200 identities per request if including plaintext passwords
//
// Avoid importing large batches with plaintext passwords. They can cause timeouts as the passwords need to be hashed before they are stored.
//
// If at least one identity is imported successfully, the response status is 200 OK.
// If all imports fail, the response is one of the following 4xx errors:
// - 400 Bad Request: The request payload is invalid or improperly formatted.
// - 409 Conflict: Duplicate identities or conflicting data were detected.
//
// If you get a 504 Gateway Timeout:
// - Reduce the batch size
// - Avoid duplicate identities
// - Pre-hash passwords with BCrypt
//
// If the issue persists, contact support.
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
//	  200: batchPatchIdentitiesResponse
//	  400: errorGeneric
//	  409: errorGeneric
//	  default: errorGeneric
func (h *Handler) batchPatchIdentities(w http.ResponseWriter, r *http.Request) {
	var (
		req BatchPatchIdentitiesBody
		res batchPatchIdentitiesResponse
	)
	if err := jsonx.NewStrictDecoder(r.Body).Decode(&req); err != nil {
		h.r.Writer().WriteErrorCode(w, r, http.StatusBadRequest, errors.WithStack(err))
		return
	}

	if len(req.Identities) > BatchPatchIdentitiesLimit {
		h.r.Writer().WriteErrorCode(w, r, http.StatusBadRequest,
			errors.WithStack(herodot.ErrBadRequest.WithReasonf(
				"The maximum number of identities per request that can be created or deleted at once is %d.",
				BatchPatchIdentitiesLimit)))
		return
	}

	res.Identities = make([]*BatchIdentityPatchResponse, len(req.Identities))
	// Array to look up the index of the identity in the identities array.
	indexInIdentities := make([]*int, len(req.Identities))
	identities := make([]*Identity, 0, len(req.Identities))

	var withUnHashedPasswordCount int
	for i, patch := range req.Identities {
		if patch.Create != nil {
			res.Identities[i] = &BatchIdentityPatchResponse{
				Action:  ActionCreate,
				PatchID: patch.ID,
			}
			identity, err := h.identityFromCreateIdentityBody(r.Context(), patch.Create)
			if err != nil {
				h.r.Writer().WriteError(w, r, err)
				return
			}
			identities = append(identities, identity)
			idx := len(identities) - 1
			indexInIdentities[i] = &idx

			if patch.Create.Credentials != nil && patch.Create.Credentials.Password != nil &&
				patch.Create.Credentials.Password.Config.Password != "" {
				withUnHashedPasswordCount++
			}
		}
	}

	if withUnHashedPasswordCount > BatchPatchIdentitiesWithPasswordLimit {
		h.r.Writer().WriteErrorCode(w, r, http.StatusBadRequest,
			errors.WithStack(herodot.ErrBadRequest.WithReasonf(
				"The maximum number of identities per request that can be created with a plaintext password is %d.",
				BatchPatchIdentitiesWithPasswordLimit)))
		return
	}

	err := h.r.IdentityManager().CreateIdentities(r.Context(), identities)
	partialErr := new(CreateIdentitiesError)
	if err != nil && !errors.As(err, &partialErr) {
		h.r.Writer().WriteError(w, r, err)
		return
	}
	for resIdx, identitiesIdx := range indexInIdentities {
		if identitiesIdx != nil {
			ident := identities[*identitiesIdx]
			// Check if the identity was created successfully.
			if failed := partialErr.Find(ident); failed != nil {
				res.Identities[resIdx].Action = ActionError
				res.Identities[resIdx].Error = failed.Error
			} else {
				res.Identities[resIdx].IdentityID = &ident.ID
			}
		}
	}

	h.r.Writer().Write(w, r, &res)
}

// Update Identity Parameters
//
// swagger:parameters updateIdentity
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
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

	// ExternalID is an optional external ID of the identity. This is used to link
	// the identity to an external system. If set, the external ID must be unique
	// across all identities.
	//
	// required: false
	ExternalID string `json:"external_id,omitempty"`
}

// swagger:route PUT /admin/identities/{id} identity updateIdentity
//
// # Update an Identity
//
// This endpoint updates an [identity](https://www.ory.sh/docs/kratos/concepts/identity-user-model). The full identity
// payload, except credentials, is expected. For partial updates, use the [patchIdentity](https://www.ory.sh/docs/reference/api#tag/identity/operation/patchIdentity) operation.
//
// A credential can be provided via the `credentials` field in the request body.
// If provided, the credentials will be imported and added to the existing credentials of the identity.
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
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var ur UpdateIdentityBody
	if err := h.dx.Decode(r, &ur,
		decoderx.HTTPJSONDecoder()); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	id := x.ParseUUID(r.PathValue("id"))
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
	identity.ExternalID = sqlxx.NullString(ur.ExternalID)

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

	h.r.Writer().Write(w, r, WithCredentialsNoConfigAndAdminMetadataInJSON(*identity))
}

// Delete Identity Parameters
//
// swagger:parameters deleteIdentity
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
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
// This endpoint returns 204 when the identity was deleted or 404 if the identity was not found.
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
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.r.PrivilegedIdentityPool().DeleteIdentity(r.Context(), x.ParseUUID(r.PathValue("id"))); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Patch Identity Parameters
//
// swagger:parameters patchIdentity
//
//nolint:deadcode,unused
//lint:ignore U1000 Used to generate Swagger and OpenAPI definitions
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
func (h *Handler) patch(w http.ResponseWriter, r *http.Request) {
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	id := x.ParseUUID(r.PathValue("id"))
	identity, err := h.r.PrivilegedIdentityPool().GetIdentityConfidential(r.Context(), id)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	oldState := identity.State

	patchedIdentity, err := jsonx.ApplyJSONPatch(requestBody, WithCredentialsAndAdminMetadataInJSON(*identity), "/id", "/stateChangedAt", "/credentials", "/credentials/oidc/**")
	if err != nil {
		h.r.Writer().WriteError(w, r, errors.WithStack(
			herodot.
				ErrBadRequest.
				WithReasonf("An error occured when applying the JSON patch").
				WithErrorf("%v", err).
				WithWrap(err),
		))
		return
	}

	if oldState != patchedIdentity.State {
		// Check if the changed state was actually valid
		if err := patchedIdentity.State.IsValid(); err != nil {
			h.r.Writer().WriteError(w, r, errors.WithStack(
				herodot.
					ErrBadRequest.
					WithReasonf("The supplied state ('%s') was not valid. Valid states are ('%s', '%s').", string(patchedIdentity.State), StateActive, StateInactive).
					WithErrorf("%v", err).
					WithWrap(err),
			))
			return
		}

		// If the state changed, we need to update the timestamp of it
		stateChangedAt := sqlxx.NullTime(time.Now())
		patchedIdentity.StateChangedAt = &stateChangedAt
	}

	updatedIdentity := Identity(patchedIdentity)

	if err := h.r.IdentityManager().Update(
		r.Context(),
		&updatedIdentity,
		ManagerAllowWriteProtectedTraits,
	); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().Write(w, r, WithCredentialsNoConfigAndAdminMetadataInJSON(updatedIdentity))
}

// Delete Credential Parameters
//
// swagger:parameters deleteIdentityCredentials
type _ struct {
	// ID is the identity's ID.
	//
	// required: true
	// in: path
	ID string `json:"id"`

	// Type is the type of credentials to delete.
	//
	// required: true
	// in: path
	Type CredentialsType `json:"type"`

	// Identifier is the identifier of the OIDC/SAML credential to delete.
	// Find the identifier by calling the `GET /admin/identities/{id}?include_credential={oidc,saml}` endpoint.
	//
	// required: false
	// in: query
	Identifier string `json:"identifier"`
}

// swagger:route DELETE /admin/identities/{id}/credentials/{type} identity deleteIdentityCredentials
//
// # Delete a credential for a specific identity
//
// Delete an [identity](https://www.ory.sh/docs/kratos/concepts/identity-user-model) credential by its type.
// You cannot delete passkeys or code auth credentials through this API.
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
//	  204: emptyResponse
//	  404: errorGeneric
//	  default: errorGeneric
func (h *Handler) deleteIdentityCredentials(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	identity, err := h.r.PrivilegedIdentityPool().GetIdentityConfidential(ctx, x.ParseUUID(r.PathValue("id")))
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	cred, ok := identity.GetCredentials(CredentialsType(r.PathValue("type")))
	if !ok {
		h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrNotFound.WithReasonf("You tried to remove a %s but this user have no %s set up.", r.PathValue("type"), r.PathValue("type"))))
		return
	}

	switch cred.Type {
	case CredentialsTypeLookup, CredentialsTypeTOTP:
		identity.DeleteCredentialsType(cred.Type)
	case CredentialsTypeWebAuthn:
		if err = identity.deleteCredentialWebAuthFromIdentity(); err != nil {
			h.r.Writer().WriteError(w, r, err)
			return
		}
	case CredentialsTypePassword, CredentialsTypeOIDC, CredentialsTypeSAML:
		firstFactor, err := h.r.IdentityManager().CountActiveFirstFactorCredentials(ctx, identity)
		if err != nil {
			h.r.Writer().WriteError(w, r, err)
			return
		}
		if firstFactor < 2 {
			h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReason("You cannot remove the last first factor credential.")))
			return
		}
		switch cred.Type {
		case CredentialsTypePassword:
			if err := identity.deleteCredentialPassword(); err != nil {
				h.r.Writer().WriteError(w, r, err)
				return
			}
		case CredentialsTypeOIDC, CredentialsTypeSAML:
			if err := identity.deleteCredentialOIDCSAMLFromIdentity(cred.Type, r.URL.Query().Get("identifier")); err != nil {
				h.r.Writer().WriteError(w, r, err)
				return
			}
		}
	default:
		// A bunch of credential type deletions are not yet implemented, e.g. passkeys, etc.
		h.r.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Credentials type %s cannot be deleted.", cred.Type)))
		return
	}

	if err := h.r.IdentityManager().Update(
		ctx,
		identity,
		ManagerAllowWriteProtectedTraits,
	); err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
