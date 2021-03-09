package sql

import (
	"context"

	"github.com/gobuffalo/pop/v5"

	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/selfservice/flow/login"
)

var _ login.FlowPersister = new(Persister)

func (p *Persister) CreateLoginFlow(ctx context.Context, r *login.Flow) error {
	return p.GetConnection(ctx).Eager().Create(r)
}

func (p *Persister) UpdateLoginFlow(ctx context.Context, r *login.Flow) error {
	return p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {
		return tx.Save(r)
	})
}

func (p *Persister) GetLoginFlow(ctx context.Context, id uuid.UUID) (*login.Flow, error) {
	conn := p.GetConnection(ctx)
	var r login.Flow
	if err := conn.Eager().Find(&r, id); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &r, nil
}

func (p *Persister) ForceLoginFlow(ctx context.Context, id uuid.UUID) error {
	return p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {
		lr, err := p.GetLoginFlow(ctx, id)
		if err != nil {
			return err
		}

		lr.Forced = true
		return tx.Save(lr)
	})
}
