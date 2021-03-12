package sql

import (
	"context"

	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/selfservice/flow/settings"
)

var _ settings.FlowPersister = new(Persister)

func (p *Persister) CreateSettingsFlow(ctx context.Context, r *settings.Flow) error {
	return sqlcon.HandleError(p.GetConnection(ctx).Create(r))
}

func (p *Persister) GetSettingsFlow(ctx context.Context, id uuid.UUID) (*settings.Flow, error) {
	var r settings.Flow
	if err := p.GetConnection(ctx).Eager().Find(&r, id); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &r, nil
}

func (p *Persister) UpdateSettingsFlow(ctx context.Context, r *settings.Flow) error {
	return p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {
		return tx.Save(r)
	})
}
