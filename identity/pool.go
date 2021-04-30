package identity

import (
	"context"

	"github.com/gofrs/uuid"
)

type (
	Pool interface {
		// ListIdentities lists all identities in the store given the page and itemsPerPage.
		ListIdentities(ctx context.Context, page, itemsPerPage int) ([]Identity, error)

		// CountIdentities counts the number of identities in the store.
		CountIdentities(ctx context.Context) (int64, error)

		// GetIdentity returns an identity by its id. Will return an error if the identity does not exist or backend
		// connectivity is broken.
		GetIdentity(context.Context, uuid.UUID) (*Identity, error)

		// FindVerifiableAddressByValue returns a matching address or sql.ErrNoRows if no address could be found.
		FindVerifiableAddressByValue(ctx context.Context, via VerifiableAddressType, address string) (*VerifiableAddress, error)

		// FindRecoveryAddressByValue returns a matching address or sql.ErrNoRows if no address could be found.
		FindRecoveryAddressByValue(ctx context.Context, via RecoveryAddressType, address string) (*RecoveryAddress, error)
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
		// if identity exists, backend connectivity is broken, or trait validation fails.
		DeleteIdentity(context.Context, uuid.UUID) error

		// UpdateVerifiableAddress updates an identity's verifiable address.
		UpdateVerifiableAddress(ctx context.Context, address *VerifiableAddress) error

		// CreateIdentity creates an identity. It is capable of setting credentials without encoding. Will return an error
		// if identity exists, backend connectivity is broken, or trait validation fails.
		CreateIdentity(context.Context, *Identity) error

		// UpdateIdentity updates an identity including its confidential / privileged / protected data.
		UpdateIdentity(context.Context, *Identity) error

		// GetIdentityConfidential returns the identity including it's raw credentials. This should only be used internally.
		GetIdentityConfidential(context.Context, uuid.UUID) (*Identity, error)

		// ListVerifiableAddresses lists all tracked verifiable addresses, regardless of whether they are already verified
		// or not.
		ListVerifiableAddresses(ctx context.Context, page, itemsPerPage int) ([]VerifiableAddress, error)

		// ListRecoveryAddresses lists all tracked recovery addresses.
		ListRecoveryAddresses(ctx context.Context, page, itemsPerPage int) ([]RecoveryAddress, error)
	}
)
