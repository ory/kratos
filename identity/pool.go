// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"

	"github.com/ory/kratos/x"
	"github.com/ory/pop/v6"
	"github.com/ory/x/crdbx"
	"github.com/ory/x/pagination/keysetpagination"
	"github.com/ory/x/sqlxx"

	"github.com/gofrs/uuid"
)

func NewUpdateIdentityOptions(opts []UpdateIdentityModifier) UpdateIdentityOptions {
	var o UpdateIdentityOptions
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// NewCreateIdentitiesOptions parses CreateIdentities modifiers.
func NewCreateIdentitiesOptions(opts []CreateIdentitiesModifier) *CreateIdentitiesOptions {
	o := &CreateIdentitiesOptions{}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// WithExtraColumns appends fixed (key, value) columns to a batch insert.
func WithExtraColumns(cols []ExtraColumn) CreateIdentitiesModifier {
	return func(o *CreateIdentitiesOptions) {
		o.ExtraColumns = append(o.ExtraColumns, cols...)
	}
}

// NewUpdateCredentialsConfigOptions parses UpdateCredentialsConfig modifiers.
func NewUpdateCredentialsConfigOptions(opts []UpdateCredentialsConfigModifier) *UpdateCredentialsConfigOptions {
	o := &UpdateCredentialsConfigOptions{}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// WithDerivedIdentifiers syncs the credential's identifier rows to the set
// derived from the post-mutation config, inside the same locked transaction.
// Like mutate, derive must be pure: it may run more than once on database
// retries.
func WithDerivedIdentifiers(derive func(newConfig []byte) ([]string, error)) UpdateCredentialsConfigModifier {
	return func(o *UpdateCredentialsConfigOptions) {
		o.DeriveIdentifiers = derive
	}
}

// DiffAgainst instructs UpdateIdentity to attempt a minimal update of the
// identity's data in the database by computing a diff against `existing` and
// only updating what is necessary, rather than bulk-replacing everything. Use
// with caution. If `existing` is different from what is stored in the database
// at the time of the update, the results are undefined. An error is returned if
// `existing` has a mismatching IdentityID or NID.
func DiffAgainst(existing *Identity) UpdateIdentityModifier {
	return func(o *UpdateIdentityOptions) {
		o.fromDatabase = existing
	}
}

func (o UpdateIdentityOptions) FromDatabase() *Identity {
	return o.fromDatabase
}

// WithoutCredentialTypes excludes the given credential types from the
// credential-association diff: their rows are neither deleted, recreated,
// nor updated by UpdateIdentity, and the returned identity carries their
// database state instead of the in-memory one. Use it for credential types
// whose rows are persisted through UpdateCredentialsConfig.
func WithoutCredentialTypes(cts ...CredentialsType) UpdateIdentityModifier {
	return func(o *UpdateIdentityOptions) {
		o.excludedCredentialTypes = append(o.excludedCredentialTypes, cts...)
	}
}

func (o UpdateIdentityOptions) ExcludedCredentialTypes() []CredentialsType {
	return o.excludedCredentialTypes
}

// WithUpdateExtraColumns appends fixed (key, value) columns to the inserts
// that UpdateIdentity performs for associated rows (addresses, credentials).
func WithUpdateExtraColumns(cols []ExtraColumn) UpdateIdentityModifier {
	return func(o *UpdateIdentityOptions) {
		o.extraColumns = append(o.extraColumns, cols...)
	}
}

func (o UpdateIdentityOptions) ExtraColumns() []ExtraColumn {
	return o.extraColumns
}

type (
	ListIdentityParameters struct {
		Expand                       Expandables
		IdsFilter                    []uuid.UUID
		CredentialsIdentifier        string
		CredentialsIdentifierSimilar string
		DeclassifyCredentials        []CredentialsType
		KeySetPagination             []keysetpagination.Option
		OrganizationID               uuid.UUID
		ConsistencyLevel             crdbx.ConsistencyLevel
		StatementTransformer         func(string) string

		// ColumnsTransformer rewrites the SELECT column list to add extra
		// columns the persister scans. Must be set together with RowScanner;
		// the persister rejects the call if only one is provided.
		ColumnsTransformer func(string) string

		// RowScanner replaces the default scan into []Identity to consume
		// the extra columns added by ColumnsTransformer. Must be set together
		// with ColumnsTransformer.
		RowScanner func(con *pop.Connection, query string, args []any) ([]Identity, error)

		// DEPRECATED
		PagePagination *x.Page
	}

	UpdateIdentityModifier func(*UpdateIdentityOptions)
	UpdateIdentityOptions  struct {
		fromDatabase            *Identity
		excludedCredentialTypes []CredentialsType
		extraColumns            []ExtraColumn
	}

	// ExtraColumn carries a (key, value) pair for an extra SQL column on a
	// batch insert (e.g. crdb_region) without coupling Identity to it.
	ExtraColumn struct {
		K string
		V any
	}

	CreateIdentitiesOptions struct {
		ExtraColumns []ExtraColumn
	}

	CreateIdentitiesModifier func(*CreateIdentitiesOptions)

	UpdateCredentialsConfigOptions struct {
		ExtraColumns      []ExtraColumn
		DeriveIdentifiers func(newConfig []byte) ([]string, error)
	}

	UpdateCredentialsConfigModifier func(*UpdateCredentialsConfigOptions)

	Pool interface {
		// ListIdentities lists all identities in the store given the page and itemsPerPage.
		ListIdentities(ctx context.Context, params ListIdentityParameters) ([]Identity, *keysetpagination.Paginator, error)

		// CountIdentities counts the number of identities in the store.
		CountIdentities(ctx context.Context) (int64, error)

		// GetIdentity returns an identity by its id. Will return an error if the identity does not exist or backend
		// connectivity is broken.
		GetIdentity(context.Context, uuid.UUID, sqlxx.Expandables) (*Identity, error)

		// FindVerifiableAddressByValue returns a matching address or sql.ErrNoRows if no address could be found.
		FindVerifiableAddressByValue(ctx context.Context, via, address string) (*VerifiableAddress, error)

		// FindRecoveryAddressByValue returns a matching address or sql.ErrNoRows if no address could be found.
		FindRecoveryAddressByValue(ctx context.Context, via, address string) (*RecoveryAddress, error)

		// FindAllRecoveryAddressValuesForIdentityByRecoveryAddressValue finds the values of all recovery addresses for an identity if at least one of its recovery addresses matches the provided value.
		FindAllRecoveryAddressValuesForIdentityByRecoveryAddressValue(ctx context.Context, anyRecoveryAddress string) ([]string, error)
	}

	PoolProvider interface {
		IdentityPool() Pool
	}

	PrivilegedPoolProvider interface {
		PrivilegedIdentityPool() PrivilegedPool
	}

	PrivilegedPool interface {
		Pool

		// FindByCredentialsIdentifier returns an identity by querying for it's credential identifiers.
		FindByCredentialsIdentifier(ctx context.Context, ct CredentialsType, match string) (*Identity, *Credentials, error)

		// DeleteIdentity removes an identity by its id. Will return an error
		// if identity does not exists, or backend connectivity is broken.
		DeleteIdentity(context.Context, uuid.UUID) error

		// DeleteIdentities removes identities by its id. Will return an error
		// if any identity does not exists, or backend connectivity is broken.
		DeleteIdentities(context.Context, []uuid.UUID) error

		// UpdateVerifiableAddress updates an identity's verifiable address.
		UpdateVerifiableAddress(ctx context.Context, address *VerifiableAddress, updateColumns ...string) error

		// CreateIdentity creates an identity. It is capable of setting credentials without encoding. Will return an error
		// if identity exists, backend connectivity is broken, or trait validation fails.
		CreateIdentity(context.Context, *Identity) error

		// CreateIdentities creates multiple identities. It is capable of setting credentials without encoding. Will return an error
		// if identity exists, backend connectivity is broken, or trait validation fails.
		CreateIdentities(ctx context.Context, identities []*Identity, opts ...CreateIdentitiesModifier) error

		// UpdateIdentity updates an identity including its confidential / privileged / protected data.
		UpdateIdentity(context.Context, *Identity, ...UpdateIdentityModifier) error

		// UpdateIdentityColumns updates targeted columns of an identity.
		UpdateIdentityColumns(ctx context.Context, i *Identity, columns ...string) error

		// UpdateCredentialsConfig atomically read-modify-writes a single
		// identity_credentials row's config under an exclusive row lock:
		// concurrent updates serialize and mutate always observes the latest
		// committed config. Use it for credential-content decisions that must
		// be mutually exclusive (PIN lockout counter, webauthn clone counter);
		// wrap mutate with UpdateConfig for typed configs.
		// mutate must be pure (it may run more than once on database retries);
		// a structurally unchanged result skips the write; the version column
		// (the config's schema version) is untouched. Calls inside a
		// surrounding transaction are rejected.
		// opts may narrow the row set with ExtraColumns (the cloud
		// multi-region persister pins crdb_region) and sync the credential's
		// identifier rows to the post-mutation config with
		// WithDerivedIdentifiers.
		UpdateCredentialsConfig(ctx context.Context, identityID uuid.UUID, ct CredentialsType, mutate func(config []byte) ([]byte, error), opts ...UpdateCredentialsConfigModifier) error

		// GetIdentityConfidential returns the identity including it's raw credentials.
		//
		// This should only be used internally. Please be aware that this method uses HydrateIdentityAssociations
		// internally, which must not be executed as part of a transaction.
		GetIdentityConfidential(context.Context, uuid.UUID) (*Identity, error)

		// ListVerifiableAddresses lists all tracked verifiable addresses, regardless of whether they are already verified
		// or not.
		ListVerifiableAddresses(ctx context.Context, page, itemsPerPage int) ([]VerifiableAddress, error)

		// ListRecoveryAddresses lists all tracked recovery addresses.
		ListRecoveryAddresses(ctx context.Context, page, itemsPerPage int) ([]RecoveryAddress, error)

		// HydrateIdentityAssociations hydrates the associations of an identity.
		//
		// Please be aware that this method must not be called within a transaction if more than one element is expanded.
		// It may error with "conn busy" otherwise.
		HydrateIdentityAssociations(ctx context.Context, i *Identity, expandables Expandables) error

		// InjectTraitsSchemaURL sets the identity's traits JSON schema URL from the schema's ID.
		InjectTraitsSchemaURL(ctx context.Context, i *Identity) error

		// FindIdentityByCredentialIdentifier returns an identity by matching the identifier to any of the identity's credentials.
		FindIdentityByCredentialIdentifier(ctx context.Context, identifier string, caseSensitive bool, expandables Expandables) (*Identity, error)

		// FindIdentityByWebauthnUserHandle returns an identity matching a webauthn user handle.
		FindIdentityByWebauthnUserHandle(ctx context.Context, userHandle []byte) (*Identity, error)

		// FindIdentityByCredentialsIdentifier returns an identity by its external ID.
		FindIdentityByExternalID(ctx context.Context, externalID string, expand sqlxx.Expandables) (*Identity, error)
	}
)

func (p ListIdentityParameters) TransformStatement(statement string) string {
	if p.StatementTransformer != nil {
		return p.StatementTransformer(statement)
	}
	return statement
}
