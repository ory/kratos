package sql

import (
	"context"

	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/continuity"
)

var _ continuity.Persister = new(Persister)

func (p *Persister) SaveContinuitySession(ctx context.Context, c *continuity.Container) error {
	return sqlcon.HandleError(p.GetConnection(ctx).Create(c))
}

func (p *Persister) GetContinuitySession(ctx context.Context, id uuid.UUID) (*continuity.Container, error) {
	var c continuity.Container
	if err := p.GetConnection(ctx).Find(&c, id); err != nil {
		return nil, sqlcon.HandleError(err)
	}
	return &c, nil
}

func (p *Persister) DeleteContinuitySession(ctx context.Context, id uuid.UUID) error {
	return sqlcon.HandleError(p.GetConnection(ctx).Destroy(&continuity.Container{ID: id}))
}
