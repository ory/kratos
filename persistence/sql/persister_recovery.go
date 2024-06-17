// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/persistence/sql/update"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/strategy/link"
	"github.com/ory/x/otelx"
	"github.com/ory/x/sqlcon"
)

var (
	_ recovery.FlowPersister      = new(Persister)
	_ link.RecoveryTokenPersister = new(Persister)
)

func (p *Persister) CreateRecoveryFlow(ctx context.Context, r *recovery.Flow) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateRecoveryFlow")
	defer otelx.End(span, &err)

	r.NID = p.NetworkID(ctx)
	return p.GetConnection(ctx).Create(r)
}

func (p *Persister) GetRecoveryFlow(ctx context.Context, id uuid.UUID) (_ *recovery.Flow, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetRecoveryFlow")
	defer otelx.End(span, &err)

	var r recovery.Flow
	if err := p.GetConnection(ctx).Where("id = ? AND nid = ?", id, p.NetworkID(ctx)).First(&r); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &r, nil
}

func (p *Persister) UpdateRecoveryFlow(ctx context.Context, r *recovery.Flow) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UpdateRecoveryFlow")
	defer otelx.End(span, &err)

	cp := *r
	cp.NID = p.NetworkID(ctx)
	return update.Generic(ctx, p.GetConnection(ctx), p.r.Tracer(ctx).Tracer(), cp)
}

func (p *Persister) CreateRecoveryToken(ctx context.Context, token *link.RecoveryToken) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateRecoveryToken")
	defer otelx.End(span, &err)

	t := token.Token
	token.Token = p.hmacValue(ctx, t)
	token.NID = p.NetworkID(ctx)

	// This should not create the request eagerly because otherwise we might accidentally create an address that isn't
	// supposed to be in the database.
	if err := p.GetConnection(ctx).Create(token); err != nil {
		return err
	}

	token.Token = t
	return nil
}

func (p *Persister) UseRecoveryToken(ctx context.Context, fID uuid.UUID, token string) (_ *link.RecoveryToken, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UseRecoveryToken")
	defer otelx.End(span, &err)

	var rt link.RecoveryToken

	nid := p.NetworkID(ctx)
	if err := sqlcon.HandleError(p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) (err error) {
		for _, secret := range p.r.Config().SecretsSession(ctx) {
			if err = tx.Where("token = ? AND nid = ? AND NOT used AND selfservice_recovery_flow_id = ?", hmacValueWithSecret(ctx, token, secret), nid, fID).First(&rt); err != nil {
				if !errors.Is(sqlcon.HandleError(err), sqlcon.ErrNoRows) {
					return err
				}
			} else {
				break
			}
		}
		if err != nil {
			return err
		}

		var ra identity.RecoveryAddress
		if err := tx.Where("id = ? AND nid = ?", rt.RecoveryAddressID, nid).First(&ra); err != nil {
			if !errors.Is(sqlcon.HandleError(err), sqlcon.ErrNoRows) {
				return err
			}
		}
		rt.RecoveryAddress = &ra

		//#nosec G201 -- TableName is static
		return tx.RawQuery(fmt.Sprintf("UPDATE %s SET used=true, used_at=? WHERE id=? AND nid = ?", rt.TableName(ctx)), time.Now().UTC(), rt.ID, nid).Exec()
	})); err != nil {
		return nil, err
	}

	return &rt, nil
}

func (p *Persister) DeleteRecoveryToken(ctx context.Context, token string) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteRecoveryToken")
	defer otelx.End(span, &err)

	//#nosec G201 -- TableName is static
	return p.GetConnection(ctx).RawQuery(fmt.Sprintf("DELETE FROM %s WHERE token=? AND nid = ?", new(link.RecoveryToken).TableName(ctx)), token, p.NetworkID(ctx)).Exec()
}

func (p *Persister) DeleteExpiredRecoveryFlows(ctx context.Context, expiresAt time.Time, limit int) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteExpiredRecoveryFlows")
	defer otelx.End(span, &err)
	//#nosec G201 -- TableName is static
	err = p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"DELETE FROM %s WHERE id in (SELECT id FROM (SELECT id FROM %s c WHERE expires_at <= ? and nid = ? ORDER BY expires_at ASC LIMIT %d ) AS s )",
		new(recovery.Flow).TableName(ctx),
		new(recovery.Flow).TableName(ctx),
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
