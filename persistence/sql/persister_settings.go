package sql

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/corp"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/selfservice/flow/settings"
)

var _ settings.FlowPersister = new(Persister)

func (p *Persister) CreateSettingsFlow(ctx context.Context, r *settings.Flow) error {
	r.NID = corp.ContextualizeNID(ctx, p.nid)
	r.EnsureInternalContext()
	return sqlcon.HandleError(p.GetConnection(ctx).Create(r))
}

func (p *Persister) GetSettingsFlow(ctx context.Context, id uuid.UUID) (*settings.Flow, error) {
	var r settings.Flow

	err := p.GetConnection(ctx).Where("id = ? AND nid = ?", id, corp.ContextualizeNID(ctx, p.nid)).First(&r)
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
	r.EnsureInternalContext()
	cp := *r
	cp.NID = corp.ContextualizeNID(ctx, p.nid)
	return p.update(ctx, cp)
}

func (p *Persister) DeleteExpiredSettingsFlows(ctx context.Context, expiresAt time.Time, limit int) error {
	// #nosec G201
	err := p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"DELETE FROM %s WHERE expires_at <= ? LIMIT ?",
		new(settings.Flow).TableName(ctx),
	),
		expiresAt,
		limit,
	).Exec()
	if err != nil {
		return sqlcon.HandleError(err)
	}
	return nil
}
