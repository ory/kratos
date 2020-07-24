package sql

import (
	"context"

	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/selfservice/flow/verification"
)

var _ verification.Persister = new(Persister)

func (p Persister) CreateVerificationRequest(ctx context.Context, r *verification.Request) error {
	// This should not create the request eagerly because otherwise we might accidentally create an address
	// that isn't supposed to be in the database.
	return p.GetConnection(ctx).Create(r)
}

func (p Persister) GetVerificationRequest(ctx context.Context, id uuid.UUID) (*verification.Request, error) {
	var r verification.Request
	if err := p.GetConnection(ctx).Find(&r, id); err != nil {
		return nil, sqlcon.HandleError(err)
	}
	return &r, nil
}

func (p Persister) UpdateVerificationRequest(ctx context.Context, r *verification.Request) error {
	return sqlcon.HandleError(p.GetConnection(ctx).Update(r))
}
