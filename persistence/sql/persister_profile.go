package sql

import (
	"context"

	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/selfservice/flow/profile"
)

var _ profile.RequestPersister = new(Persister)

func (p *Persister) CreateProfileRequest(ctx context.Context, r *profile.Request) error {
	r.IdentityID = r.Identity.ID
	return sqlcon.HandleError(p.GetConnection(ctx).Create(r)) // This must not be eager or identities will be created / updated
}

func (p *Persister) GetProfileRequest(ctx context.Context, id uuid.UUID) (*profile.Request, error) {
	var r profile.Request
	if err := p.GetConnection(ctx).Eager().Find(&r, id); err != nil {
		return nil, sqlcon.HandleError(err)
	}
	return &r, nil
}

func (p *Persister) UpdateProfileRequest(ctx context.Context, r *profile.Request) error {
	return sqlcon.HandleError(p.GetConnection(ctx).Update(r)) // This must not be eager or identities will be created / updated
}
