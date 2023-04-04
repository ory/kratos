// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"fmt"
	"time"

	"github.com/gobuffalo/pop/v6"

	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/persistence/sql/update"
	"github.com/ory/kratos/selfservice/flow/login"
)

var _ login.FlowPersister = new(Persister)

func (p *Persister) CreateLoginFlow(ctx context.Context, r *login.Flow) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateLoginFlow")
	defer span.End()

	r.NID = p.NetworkID(ctx)
	r.EnsureInternalContext()
	return p.GetConnection(ctx).Create(r)
}

func (p *Persister) UpdateLoginFlow(ctx context.Context, r *login.Flow) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UpdateLoginFlow")
	defer span.End()

	r.EnsureInternalContext()
	cp := *r
	cp.NID = p.NetworkID(ctx)
	return update.Generic(ctx, p.GetConnection(ctx), p.r.Tracer(ctx).Tracer(), cp)
}

func (p *Persister) GetLoginFlow(ctx context.Context, id uuid.UUID) (*login.Flow, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetLoginFlow")
	defer span.End()

	conn := p.GetConnection(ctx)

	var r login.Flow
	if err := conn.Where("id = ? AND nid = ?", id, p.NetworkID(ctx)).First(&r); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &r, nil
}

func (p *Persister) ForceLoginFlow(ctx context.Context, id uuid.UUID) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.ForceLoginFlow")
	defer span.End()

	return p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {
		lr, err := p.GetLoginFlow(ctx, id)
		if err != nil {
			return err
		}

		lr.Refresh = true
		return tx.Save(lr, "nid")
	})
}

func (p *Persister) DeleteExpiredLoginFlows(ctx context.Context, expiresAt time.Time, limit int) error {
	//#nosec G201 -- TableName is static
	err := p.GetConnection(ctx).RawQuery(fmt.Sprintf(
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
