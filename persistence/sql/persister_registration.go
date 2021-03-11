package sql

import (
	"context"

	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/selfservice/flow/registration"
)

func (p *Persister) CreateRegistrationFlow(ctx context.Context, r *registration.Flow) error {
	return p.GetConnection(ctx).Eager().Create(r)
}

func (p *Persister) UpdateRegistrationFlow(ctx context.Context, r *registration.Flow) error {
	return p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {
		return tx.Save(r)
	})
}

func (p *Persister) GetRegistrationFlow(ctx context.Context, id uuid.UUID) (*registration.Flow, error) {
	var r registration.Flow
	if err := p.GetConnection(ctx).Eager().Find(&r, id); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &r, nil
}
