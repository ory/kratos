// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/persistence/sql/update"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"

	"github.com/ory/x/otelx"
	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/strategy/link"
)

var _ verification.FlowPersister = new(Persister)

func (p *Persister) CreateVerificationFlow(ctx context.Context, r *verification.Flow) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateVerificationFlow")
	defer otelx.End(span, &err)

	r.NID = p.NetworkID(ctx)
	// This should not create the request eagerly because otherwise we might accidentally create an address
	// that isn't supposed to be in the database.
	return p.GetConnection(ctx).Create(r)
}

func (p *Persister) GetVerificationFlow(ctx context.Context, id uuid.UUID) (_ *verification.Flow, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetVerificationFlow")
	defer otelx.End(span, &err)

	var r verification.Flow
	if err := p.GetConnection(ctx).Where("id = ? AND nid = ?", id, p.NetworkID(ctx)).First(&r); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &r, nil
}

func (p *Persister) UpdateVerificationFlow(ctx context.Context, r *verification.Flow) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UpdateVerificationFlow")
	defer otelx.End(span, &err)

	cp := *r
	cp.NID = p.NetworkID(ctx)
	return update.Generic(ctx, p.GetConnection(ctx), p.r.Tracer(ctx).Tracer(), cp)
}

func (p *Persister) CreateVerificationToken(ctx context.Context, token *link.VerificationToken) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateVerificationToken")
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

func (p *Persister) UseVerificationToken(ctx context.Context, fID uuid.UUID, token string) (_ *link.VerificationToken, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UseVerificationToken")
	defer otelx.End(span, &err)

	var rt link.VerificationToken

	nid := p.NetworkID(ctx)
	if err := sqlcon.HandleError(p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) (err error) {
		for _, secret := range p.r.Config().SecretsSession(ctx) {
			if err = tx.Where("token = ? AND nid = ? AND NOT used AND selfservice_verification_flow_id = ?", hmacValueWithSecret(ctx, token, secret), nid, fID).First(&rt); err != nil {
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

		var va identity.VerifiableAddress
		if err := tx.Where("id = ? AND nid = ?", rt.VerifiableAddressID, nid).First(&va); err != nil {
			return sqlcon.HandleError(err)
		}

		rt.VerifiableAddress = &va

		//#nosec G201 -- TableName is static
		return tx.RawQuery(fmt.Sprintf("UPDATE %s SET used=true, used_at=? WHERE id=? AND nid = ?", rt.TableName(ctx)), time.Now().UTC(), rt.ID, nid).Exec()
	})); err != nil {
		return nil, err
	}

	return &rt, nil
}

func (p *Persister) DeleteVerificationToken(ctx context.Context, token string) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteVerificationToken")
	defer otelx.End(span, &err)

	nid := p.NetworkID(ctx)
	//#nosec G201 -- TableName is static
	return p.GetConnection(ctx).RawQuery(fmt.Sprintf("DELETE FROM %s WHERE token=? AND nid = ?", new(link.VerificationToken).TableName(ctx)), token, nid).Exec()
}

func (p *Persister) DeleteExpiredVerificationFlows(ctx context.Context, expiresAt time.Time, limit int) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteExpiredVerificationFlows")
	defer otelx.End(span, &err)
	//#nosec G201 -- TableName is static
	err = p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"DELETE FROM %s WHERE id in (SELECT id FROM (SELECT id FROM %s c WHERE expires_at <= ? and nid = ? ORDER BY expires_at ASC LIMIT %d ) AS s )",
		new(verification.Flow).TableName(ctx),
		new(verification.Flow).TableName(ctx),
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
