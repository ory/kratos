// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"crypto/subtle"
	"fmt"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/pkg/errors"

	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/persistence/sql/update"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/strategy/code"
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

func (p *Persister) CreateLoginCode(ctx context.Context, codeParams *code.CreateLoginCodeParams) (*code.LoginCode, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateLoginCode")
	defer span.End()

	now := time.Now().UTC()
	loginCode := &code.LoginCode{
		IdentityID:  codeParams.IdentityID,
		Address:     codeParams.Address,
		AddressType: codeParams.AddressType,
		CodeHMAC:    p.hmacValue(ctx, codeParams.RawCode),
		IssuedAt:    now,
		ExpiresAt:   now.Add(p.r.Config().SelfServiceCodeMethodLifespan(ctx)),
		FlowID:      codeParams.FlowID,
		NID:         p.NetworkID(ctx),
		ID:          uuid.Nil,
	}

	if err := p.GetConnection(ctx).Create(loginCode); err != nil {
		return nil, sqlcon.HandleError(err)
	}
	return loginCode, nil
}

func (p *Persister) UseLoginCode(ctx context.Context, flowID uuid.UUID, identityID uuid.UUID, codeVal string) (*code.LoginCode, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UseLoginCode")
	defer span.End()

	var loginCode *code.LoginCode

	nid := p.NetworkID(ctx)
	flowTableName := new(login.Flow).TableName(ctx)

	if err := sqlcon.HandleError(p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) (err error) {
		//#nosec G201 -- TableName is static
		if err := sqlcon.HandleError(tx.RawQuery(fmt.Sprintf("UPDATE %s SET submit_count = submit_count + 1 WHERE id = ? AND nid = ?", flowTableName), flowID, nid).Exec()); err != nil {
			return err
		}

		var submitCount int
		// Because MySQL does not support "RETURNING" clauses, but we need the updated `submit_count` later on.
		//#nosec G201 -- TableName is static
		if err := sqlcon.HandleError(tx.RawQuery(fmt.Sprintf("SELECT submit_count FROM %s WHERE id = ? AND nid = ?", flowTableName), flowID, nid).First(&submitCount)); err != nil {
			if errors.Is(err, sqlcon.ErrNoRows) {
				// Return no error, as that would roll back the transaction
				return nil
			}
			return err
		}

		if submitCount > 5 {
			return errors.WithStack(code.ErrCodeSubmittedTooOften)
		}

		var loginCodes []code.LoginCode
		if err = sqlcon.HandleError(tx.Where("nid = ? AND selfservice_login_flow_id = ? AND identity_id = ?", nid, flowID, identityID).All(&loginCodes)); err != nil {
			if errors.Is(err, sqlcon.ErrNoRows) {
				return err
			}
			return nil
		}

	secrets:
		for _, secret := range p.r.Config().SecretsSession(ctx) {
			suppliedCode := []byte(p.hmacValueWithSecret(ctx, codeVal, secret))
			for i := range loginCodes {
				code := loginCodes[i]
				if subtle.ConstantTimeCompare([]byte(code.CodeHMAC), suppliedCode) == 0 {
					// Not the supplied code
					continue
				}
				loginCode = &code
				break secrets
			}
		}

		if loginCode == nil || !loginCode.IsValid() {
			// Return no error, as that would roll back the transaction
			return nil
		}

		//#nosec G201 -- TableName is static
		return sqlcon.HandleError(tx.RawQuery(fmt.Sprintf("UPDATE %s SET used_at = ? WHERE id = ? AND nid = ?", loginCode.TableName(ctx)), time.Now().UTC(), loginCode.ID, nid).Exec())
	})); err != nil {
		return nil, err
	}

	if loginCode == nil {
		return nil, errors.WithStack(code.ErrCodeNotFound)
	}

	if loginCode.IsExpired() {
		return nil, errors.WithStack(flow.NewFlowExpiredError(loginCode.ExpiresAt))
	}

	if loginCode.WasUsed() {
		return nil, errors.WithStack(code.ErrCodeAlreadyUsed)
	}

	return loginCode, nil
}

func (p *Persister) DeleteLoginCodesOfFlow(ctx context.Context, flowID uuid.UUID) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteLoginCodesOfFlow")
	defer span.End()

	//#nosec G201 -- TableName is static
	return p.GetConnection(ctx).RawQuery(fmt.Sprintf("DELETE FROM %s WHERE selfservice_login_flow_id = ? AND nid = ?", new(code.LoginCode).TableName(ctx)), flowID, p.NetworkID(ctx)).Exec()
}

func (p *Persister) GetUsedLoginCode(ctx context.Context, flowID uuid.UUID) (*code.LoginCode, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetUsedLoginCode")
	defer span.End()

	var loginCode code.LoginCode
	if err := p.Connection(ctx).RawQuery(fmt.Sprintf("SELECT * FROM %s WHERE selfservice_login_flow_id = ? AND nid = ? AND used_at IS NOT NULL", new(code.LoginCode).TableName(ctx)), flowID, p.NetworkID(ctx)).First(&loginCode); err != nil {
		return nil, sqlcon.HandleError(err)
	}
	return &loginCode, nil
}
