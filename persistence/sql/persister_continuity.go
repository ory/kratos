// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/gofrs/uuid"

	"github.com/ory/x/otelx"
	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/continuity"
)

var _ continuity.Persister = new(Persister)

func (p *Persister) SaveContinuitySession(ctx context.Context, c *continuity.Container) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.SaveContinuitySession")
	defer otelx.End(span, &err)

	c.NID = p.NetworkID(ctx)
	return sqlcon.HandleError(p.GetConnection(ctx).Create(c))
}

func (p *Persister) SetContinuitySessionExpiry(ctx context.Context, id uuid.UUID, expiresAt time.Time) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.SetContinuitySessionExpiry")
	defer otelx.End(span, &err)

	if rows, err := p.GetConnection(ctx).
		Where("id = ? AND nid = ?", id, p.NetworkID(ctx)).
		UpdateQuery(&continuity.Container{
			ExpiresAt: expiresAt,
		}, "expires_at"); err != nil {
		return sqlcon.HandleError(err)
	} else if rows == 0 {
		return errors.WithStack(sqlcon.ErrNoRows)
	}

	return nil
}

func (p *Persister) GetContinuitySession(ctx context.Context, id uuid.UUID) (_ *continuity.Container, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetContinuitySession")
	defer otelx.End(span, &err)

	var c continuity.Container
	if err := p.GetConnection(ctx).Where("id = ? AND nid = ?", id, p.NetworkID(ctx)).First(&c); err != nil {
		return nil, sqlcon.HandleError(err)
	}
	return &c, nil
}

func (p *Persister) DeleteContinuitySession(ctx context.Context, id uuid.UUID) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteContinuitySession")
	defer otelx.End(span, &err)

	if count, err := p.GetConnection(ctx).RawQuery(
		//#nosec G201 -- TableName is static
		fmt.Sprintf("DELETE FROM %s WHERE id=? AND nid=?",
			new(continuity.Container).TableName(ctx)), id, p.NetworkID(ctx)).ExecWithCount(); err != nil {
		return sqlcon.HandleError(err)
	} else if count == 0 {
		return errors.WithStack(sqlcon.ErrNoRows)
	}
	return nil
}

func (p *Persister) DeleteExpiredContinuitySessions(ctx context.Context, expiresAt time.Time, limit int) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteExpiredContinuitySessions")
	defer otelx.End(span, &err)
	//#nosec G201 -- TableName is static
	err = p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"DELETE FROM %s WHERE id in (SELECT id FROM (SELECT id FROM %s c WHERE expires_at <= ? and nid = ? ORDER BY expires_at ASC LIMIT %d ) AS s )",
		new(continuity.Container).TableName(ctx),
		new(continuity.Container).TableName(ctx),
		limit,
	),
		expiresAt,
		p.NetworkID(ctx),
	).Exec()
	if err != nil {
		return sqlcon.HandleError(err)
	}
	return nil
}
