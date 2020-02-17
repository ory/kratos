package sql

import (
	"context"

	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/selfservice/flow/verify"
)

var _ verify.Persister = new(Persister)

func (p Persister) CreateVerifyRequest(ctx context.Context, r *verify.Request) error {
	// This should not create the request eagerly because otherwise we might accidentally create an address
	// that isn't supposed to be in the database.
	return p.GetConnection(ctx).Create(r)
}

func (p Persister) GetVerifyRequest(ctx context.Context, id uuid.UUID) (*verify.Request, error) {
	var r verify.Request
	if err := p.GetConnection(ctx).Find(&r, id); err != nil {
		return nil, sqlcon.HandleError(err)
	}
	return &r, nil
}

func (p Persister) UpdateVerifyRequest(ctx context.Context, r *verify.Request) error {
	return sqlcon.HandleError(p.GetConnection(ctx).Update(r))
}
