package sql

import (
	"context"

	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/selfservice/flow/recovery"
)

var _ recovery.RequestPersister = new(Persister)

func (p Persister) CreateRecoveryRequest(ctx context.Context, r *recovery.Request) error {
	// This should not create the request eagerly because otherwise we might accidentally create an address
	// that isn't supposed to be in the database.
	return p.GetConnection(ctx).Create(r)
}

func (p Persister) GetRecoveryRequest(ctx context.Context, id uuid.UUID) (*recovery.Request, error) {
	var r recovery.Request
	if err := p.GetConnection(ctx).Find(&r, id); err != nil {
		return nil, sqlcon.HandleError(err)
	}
	return &r, nil
}

func (p Persister) UpdateRecoveryRequest(ctx context.Context, r *recovery.Request) error {
	return sqlcon.HandleError(p.GetConnection(ctx).Update(r))
}
