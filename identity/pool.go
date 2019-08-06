package identity

import (
	"context"
)

type Pool interface {
	// RequestID returns the pool's RequestID.
	RequestID() string

	// FindByCredentialsIdentifier returns an identity by querying for it's credential identifiers.
	FindByCredentialsIdentifier(ctx context.Context, ct CredentialsType, match string) (*Identity, *Credentials, error)

	Create(context.Context, *Identity) (*Identity, error)

	List(ctx context.Context, limit, offset int) ([]Identity, error)

	Update(context.Context, *Identity) (*Identity, error)

	Delete(context.Context, string) error

	Get(context.Context, string) (*Identity, error)

	// HasCredentialsConflict(i *Identity) bool

	// Find returns an identity using a unique identifier (phone number, email, username, urn, ...) or an error.
	// Find(ctx context.Context, search string) (*Identity, error)

	// FindByDiscriminator returns an identity using a unique identifier (phone number, email, username, urn, ...) or an error.
	// FindByDiscriminator(Discriminator string, id string) (*Identity, error)
}

type PoolProvider interface {
	IdentityPool() Pool
}
