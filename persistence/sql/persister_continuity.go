package sql

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/ory/kratos/corp"

	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/continuity"
)

var _ continuity.Persister = new(Persister)

func (p *Persister) SaveContinuitySession(ctx context.Context, c *continuity.Container) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.SaveContinuitySession")
	defer span.End()

	c.NID = corp.ContextualizeNID(ctx, p.nid)
	return sqlcon.HandleError(p.GetConnection(ctx).Create(c))
}

func (p *Persister) GetContinuitySession(ctx context.Context, id uuid.UUID) (*continuity.Container, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetContinuitySession")
	defer span.End()

	var c continuity.Container
	if err := p.GetConnection(ctx).Where("id = ? AND nid = ?", id, corp.ContextualizeNID(ctx, p.nid)).First(&c); err != nil {
		return nil, sqlcon.HandleError(err)
	}
	return &c, nil
}

func (p *Persister) DeleteContinuitySession(ctx context.Context, id uuid.UUID) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteContinuitySession")
	defer span.End()

	if count, err := p.GetConnection(ctx).RawQuery(
		// #nosec
		fmt.Sprintf("DELETE FROM %s WHERE id=? AND nid=?",
			new(continuity.Container).TableName(ctx)), id, corp.ContextualizeNID(ctx, p.nid)).ExecWithCount(); err != nil {
		return sqlcon.HandleError(err)
	} else if count == 0 {
		return errors.WithStack(sqlcon.ErrNoRows)
	}
	return nil
}
