// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"fmt"
	"time"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/persistence/sql/update"

	"github.com/gofrs/uuid"

	"github.com/ory/x/otelx"
	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/selfservice/flow/settings"
)

var _ settings.FlowPersister = new(Persister)

func (p *Persister) CreateSettingsFlow(ctx context.Context, r *settings.Flow) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateSettingsFlow")
	defer otelx.End(span, &err)

	r.NID = p.NetworkID(ctx)
	r.EnsureInternalContext()
	return sqlcon.HandleError(p.GetConnection(ctx).Create(r))
}

func (p *Persister) GetSettingsFlow(ctx context.Context, id uuid.UUID) (_ *settings.Flow, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetSettingsFlow")
	defer otelx.End(span, &err)

	var r settings.Flow

	err = p.GetConnection(ctx).Where("id = ? AND nid = ?", id, p.NetworkID(ctx)).First(&r)
	if err != nil {
		return nil, sqlcon.HandleError(err)
	}

	r.Identity, err = p.GetIdentity(ctx, r.IdentityID, identity.ExpandDefault)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (p *Persister) UpdateSettingsFlow(ctx context.Context, r *settings.Flow) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UpdateSettingsFlow")
	defer otelx.End(span, &err)

	r.EnsureInternalContext()
	cp := *r
	cp.NID = p.NetworkID(ctx)
	return update.Generic(ctx, p.GetConnection(ctx), p.r.Tracer(ctx).Tracer(), cp)
}

func (p *Persister) DeleteExpiredSettingsFlows(ctx context.Context, expiresAt time.Time, limit int) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteExpiredSettingsFlows")
	defer otelx.End(span, &err)
	//#nosec G201 -- TableName is static
	err = p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"DELETE FROM %[1]s WHERE id in (SELECT id FROM (SELECT id FROM %[1]s c WHERE expires_at <= ? and nid = ? ORDER BY expires_at ASC LIMIT ?) AS s)",
		settings.Flow{}.TableName(),
	),
		expiresAt,
		p.NetworkID(ctx),
		limit,
	).Exec()

	return sqlcon.HandleError(err)
}
