package profile

import "context"

type (
	RequestPersister interface {
		CreateProfileRequest(context.Context, *Request) error
		GetProfileRequest(ctx context.Context, id string) (*Request, error)
		UpdateProfileRequest(context.Context, string, *Request) error
	}
	RequestPersistenceProvider interface {
		ProfileRequestPersister() RequestPersister
	}
)
