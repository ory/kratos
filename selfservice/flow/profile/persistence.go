package profile

import (
	"context"

	"github.com/gofrs/uuid"
)

type (
	RequestPersister interface {
		CreateProfileRequest(context.Context, *Request) error
		GetProfileRequest(ctx context.Context, id uuid.UUID) (*Request, error)
		UpdateProfileRequest(context.Context, *Request) error
	}
	RequestPersistenceProvider interface {
		ProfileRequestPersister() RequestPersister
	}
)
