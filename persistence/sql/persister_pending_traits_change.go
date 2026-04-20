// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/persistence/sql/update"
	"github.com/ory/x/otelx"
	"github.com/ory/x/sqlcon"
)

var _ identity.PendingTraitsChangePersister = new(Persister)

func (p *Persister) CreatePendingTraitsChange(ctx context.Context, ptc *identity.PendingTraitsChange) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreatePendingTraitsChange")
	defer otelx.End(span, &err)

	ptc.ID = uuid.Must(uuid.NewV4())
	ptc.NID = p.NetworkID(ctx)
	ptc.CreatedAt = time.Now().UTC().Truncate(time.Microsecond)
	ptc.UpdatedAt = ptc.CreatedAt

	if err := sqlcon.HandleError(p.GetConnection(ctx).Create(ptc)); err != nil {
		if !errors.Is(err, sqlcon.ErrUniqueViolation()) {
			return err
		}

		// Race: another concurrent request created a pending change after our delete.
		// Delete again and retry once.
		if err := p.DeletePendingTraitsChangesByIdentity(ctx, ptc.IdentityID); err != nil {
			return err
		}
		// TODO(jonas): We could do this on the SQL level with an ON CONFLICT DO UPDATE, but that would require a more complex query and we want to keep the logic in Go for better observability and error handling.
		if err := sqlcon.HandleError(p.GetConnection(ctx).Create(ptc)); err != nil {
			return err
		}
	}
	return nil
}

func (p *Persister) GetPendingTraitsChangeByVerificationFlow(ctx context.Context, flowID uuid.UUID) (_ *identity.PendingTraitsChange, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetPendingTraitsChangeByVerificationFlow")
	defer otelx.End(span, &err)

	var ptc identity.PendingTraitsChange
	if err := p.GetConnection(ctx).
		Where("nid = ? AND verification_flow_id = ? AND status = ?",
			p.NetworkID(ctx), flowID, identity.PendingTraitsChangeStatusPending).
		First(&ptc); err != nil {
		return nil, sqlcon.HandleError(err)
	}
	return &ptc, nil
}

func (p *Persister) DeletePendingTraitsChangesByIdentity(ctx context.Context, identityID uuid.UUID) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeletePendingTraitsChangesByIdentity")
	defer otelx.End(span, &err)

	return sqlcon.HandleError(
		p.GetConnection(ctx).RawQuery(
			"DELETE FROM identity_pending_traits_changes WHERE nid = ? AND identity_id = ? AND status = 'pending'",
			p.NetworkID(ctx),
			identityID,
		).Exec(),
	)
}

func (p *Persister) UpdatePendingTraitsChange(ctx context.Context, ptc *identity.PendingTraitsChange) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UpdatePendingTraitsChange")
	defer otelx.End(span, &err)

	ptc.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
	return update.Generic(ctx, p.GetConnection(ctx), p.r.Tracer(ctx).Tracer(), ptc)
}
