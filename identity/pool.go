// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"

	"github.com/ory/kratos/x"
	"github.com/ory/x/crdbx"
	"github.com/ory/x/pagination/keysetpagination"
	"github.com/ory/x/sqlxx"

	"github.com/gofrs/uuid"
)

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

		// DEPRECATED
		PagePagination *x.Page
	}

	Pool interface {
		// ListIdentities lists all identities in the store given the page and itemsPerPage.
		ListIdentities(ctx context.Context, params ListIdentityParameters) ([]Identity, *keysetpagination.Paginator, error)

		// CountIdentities counts the number of identities in the store.
		CountIdentities(ctx context.Context) (int64, error)

		// GetIdentity returns an identity by its id. Will return an error if the identity does not exist or backend
		// connectivity is broken.
		GetIdentity(context.Context, uuid.UUID, sqlxx.Expandables) (*Identity, error)

		// FindVerifiableAddressByValue returns a matching address or sql.ErrNoRows if no address could be found.
		FindVerifiableAddressByValue(ctx context.Context, via string, address string) (*VerifiableAddress, error)

		// FindRecoveryAddressByValue returns a matching address or sql.ErrNoRows if no address could be found.
		FindRecoveryAddressByValue(ctx context.Context, via RecoveryAddressType, address string) (*RecoveryAddress, error)

		// FindAllRecoveryAddressesForIdentityByRecoveryAddressValue finds all recovery addresses for an identity if at least one of its recovery addresses matches the provided value.
		FindAllRecoveryAddressesForIdentityByRecoveryAddressValue(ctx context.Context, anyRecoveryAddress string) ([]RecoveryAddress, error)
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
		CreateIdentities(context.Context, ...*Identity) error

		// UpdateIdentity updates an identity including its confidential / privileged / protected data.
		UpdateIdentity(context.Context, *Identity) error

		// UpdateIdentityColumns updates targeted columns of an identity.
		UpdateIdentityColumns(ctx context.Context, i *Identity, columns ...string) error

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
		FindIdentityByCredentialIdentifier(ctx context.Context, identifier string, caseSensitive bool) (*Identity, error)

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
