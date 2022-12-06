// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/selfservice/flow/settings"
)

var _ settings.FlowPersister = new(Persister)

func (p *Persister) CreateSettingsFlow(ctx context.Context, r *settings.Flow) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateSettingsFlow")
	defer span.End()

	r.NID = p.NetworkID(ctx)
	r.EnsureInternalContext()
	return sqlcon.HandleError(p.GetConnection(ctx).Create(r))
}

func (p *Persister) GetSettingsFlow(ctx context.Context, id uuid.UUID) (*settings.Flow, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetSettingsFlow")
	defer span.End()

	var r settings.Flow

	err := p.GetConnection(ctx).Where("id = ? AND nid = ?", id, p.NetworkID(ctx)).First(&r)
	if err != nil {
		return nil, sqlcon.HandleError(err)
	}

	r.Identity, err = p.GetIdentity(ctx, r.IdentityID)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (p *Persister) UpdateSettingsFlow(ctx context.Context, r *settings.Flow) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UpdateSettingsFlow")
	defer span.End()

	r.EnsureInternalContext()
	cp := *r
	cp.NID = p.NetworkID(ctx)
	return p.update(ctx, cp)
}

func (p *Persister) DeleteExpiredSettingsFlows(ctx context.Context, expiresAt time.Time, limit int) error {
	// #nosec G201
	err := p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"DELETE FROM %s WHERE id in (SELECT id FROM (SELECT id FROM %s c WHERE expires_at <= ? and nid = ? ORDER BY expires_at ASC LIMIT %d ) AS s )",
		new(settings.Flow).TableName(ctx),
		new(settings.Flow).TableName(ctx),
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
