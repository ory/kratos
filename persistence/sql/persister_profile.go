package sql

import (
	"context"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/selfservice/flow/profile"
)

var _ profile.RequestPersister = new(Persister)

func (p *Persister) CreateProfileRequest(_ context.Context, r *profile.Request) error {
	return p.c.Eager().Create(r)
}

func (p *Persister) GetProfileRequest(_ context.Context, id uuid.UUID) (*profile.Request, error) {
	var r profile.Request
	if err := p.c.Eager().Find(&r, id); err != nil {
		return nil, err
	}
	return &r, nil
}

func (p *Persister) UpdateProfileRequest(ctx context.Context, r *profile.Request) error {
	return p.c.Eager().Update(&r)
}
