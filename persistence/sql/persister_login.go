// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"fmt"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"

	"github.com/ory/x/otelx"
	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/persistence/sql/update"
	"github.com/ory/kratos/selfservice/flow/login"
)

var _ login.FlowPersister = new(Persister)

func (p *Persister) CreateLoginFlow(ctx context.Context, r *login.Flow) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateLoginFlow")
	defer otelx.End(span, &err)

	r.NID = p.NetworkID(ctx)
	r.EnsureInternalContext()
	return p.GetConnection(ctx).Create(r)
}

func (p *Persister) UpdateLoginFlow(ctx context.Context, r *login.Flow) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UpdateLoginFlow")
	defer otelx.End(span, &err)

	r.EnsureInternalContext()
	cp := *r
	cp.NID = p.NetworkID(ctx)
	return update.Generic(ctx, p.GetConnection(ctx), p.r.Tracer(ctx).Tracer(), cp)
}

func (p *Persister) GetLoginFlow(ctx context.Context, id uuid.UUID) (_ *login.Flow, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetLoginFlow")
	defer otelx.End(span, &err)

	conn := p.GetConnection(ctx)

	var r login.Flow
	if err := conn.Where("id = ? AND nid = ?", id, p.NetworkID(ctx)).First(&r); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &r, nil
}

func (p *Persister) ForceLoginFlow(ctx context.Context, id uuid.UUID) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.ForceLoginFlow")
	defer otelx.End(span, &err)

	return p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {
		lr, err := p.GetLoginFlow(ctx, id)
		if err != nil {
			return err
		}

		lr.Refresh = true
		return tx.Save(lr, "nid")
	})
}

func (p *Persister) DeleteExpiredLoginFlows(ctx context.Context, expiresAt time.Time, limit int) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteExpiredLoginFlows")
	defer otelx.End(span, &err)
	//#nosec G201 -- TableName is static
	err = p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"DELETE FROM %s WHERE id in (SELECT id FROM (SELECT id FROM %s c WHERE expires_at <= ? and nid = ? ORDER BY expires_at ASC LIMIT %d ) AS s )",
		new(login.Flow).TableName(ctx),
		new(login.Flow).TableName(ctx),
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
