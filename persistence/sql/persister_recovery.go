// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"crypto/subtle"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/persistence/sql/update"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/kratos/selfservice/strategy/link"
	"github.com/ory/x/sqlcon"
)

var (
	_ recovery.FlowPersister      = new(Persister)
	_ link.RecoveryTokenPersister = new(Persister)
)

func (p *Persister) CreateRecoveryFlow(ctx context.Context, r *recovery.Flow) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateRecoveryFlow")
	defer span.End()

	r.NID = p.NetworkID(ctx)
	return p.GetConnection(ctx).Create(r)
}

func (p *Persister) GetRecoveryFlow(ctx context.Context, id uuid.UUID) (*recovery.Flow, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetRecoveryFlow")
	defer span.End()

	var r recovery.Flow
	if err := p.GetConnection(ctx).Where("id = ? AND nid = ?", id, p.NetworkID(ctx)).First(&r); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &r, nil
}

func (p *Persister) UpdateRecoveryFlow(ctx context.Context, r *recovery.Flow) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UpdateRecoveryFlow")
	defer span.End()

	cp := *r
	cp.NID = p.NetworkID(ctx)
	return update.Generic(ctx, p.GetConnection(ctx), p.r.Tracer(ctx).Tracer(), cp)
}

func (p *Persister) CreateRecoveryToken(ctx context.Context, token *link.RecoveryToken) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateRecoveryToken")
	defer span.End()

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

func (p *Persister) UseRecoveryToken(ctx context.Context, fID uuid.UUID, token string) (*link.RecoveryToken, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UseRecoveryToken")
	defer span.End()

	var rt link.RecoveryToken

	nid := p.NetworkID(ctx)
	if err := sqlcon.HandleError(p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) (err error) {
		for _, secret := range p.r.Config().SecretsSession(ctx) {
			if err = tx.Where("token = ? AND nid = ? AND NOT used AND selfservice_recovery_flow_id = ?", p.hmacValueWithSecret(ctx, token, secret), nid, fID).First(&rt); err != nil {
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

func (p *Persister) DeleteRecoveryToken(ctx context.Context, token string) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteRecoveryToken")
	defer span.End()

	//#nosec G201 -- TableName is static
	return p.GetConnection(ctx).RawQuery(fmt.Sprintf("DELETE FROM %s WHERE token=? AND nid = ?", new(link.RecoveryToken).TableName(ctx)), token, p.NetworkID(ctx)).Exec()
}

func (p *Persister) DeleteExpiredRecoveryFlows(ctx context.Context, expiresAt time.Time, limit int) error {
	//#nosec G201 -- TableName is static
	err := p.GetConnection(ctx).RawQuery(fmt.Sprintf(
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

func (p *Persister) CreateRecoveryCode(ctx context.Context, dto *code.CreateRecoveryCodeParams) (*code.RecoveryCode, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateRecoveryCode")
	defer span.End()

	now := time.Now()

	recoveryCode := &code.RecoveryCode{
		ID:         uuid.Nil,
		CodeHMAC:   p.hmacValue(ctx, dto.RawCode),
		ExpiresAt:  now.UTC().Add(dto.ExpiresIn),
		IssuedAt:   now,
		CodeType:   dto.CodeType,
		FlowID:     dto.FlowID,
		NID:        p.NetworkID(ctx),
		IdentityID: dto.IdentityID,
	}

	if dto.RecoveryAddress != nil {
		recoveryCode.RecoveryAddress = dto.RecoveryAddress
		recoveryCode.RecoveryAddressID = uuid.NullUUID{
			UUID:  dto.RecoveryAddress.ID,
			Valid: true,
		}
	}

	// This should not create the request eagerly because otherwise we might accidentally create an address that isn't
	// supposed to be in the database.
	if err := p.GetConnection(ctx).Create(recoveryCode); err != nil {
		return nil, err
	}

	return recoveryCode, nil
}

// UseRecoveryCode attempts to "use" the supplied code in the flow
//
// If the supplied code matched a code from the flow, no error is returned
// If an invalid code was submitted with this flow more than 5 times, an error is returned
// TODO: Extract the business logic to a new service/manager (https://github.com/ory/kratos/issues/2785)
func (p *Persister) UseRecoveryCode(ctx context.Context, fID uuid.UUID, codeVal string) (*code.RecoveryCode, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UseRecoveryCode")
	defer span.End()

	var recoveryCode *code.RecoveryCode

	nid := p.NetworkID(ctx)

	flowTableName := new(recovery.Flow).TableName(ctx)

	if err := sqlcon.HandleError(p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) (err error) {
		//#nosec G201 -- TableName is static
		if err := sqlcon.HandleError(tx.RawQuery(fmt.Sprintf("UPDATE %s SET submit_count = submit_count + 1 WHERE id = ? AND nid = ?", flowTableName), fID, nid).Exec()); err != nil {
			return err
		}

		var submitCount int
		// Because MySQL does not support "RETURNING" clauses, but we need the updated `submit_count` later on.
		//#nosec G201 -- TableName is static
		if err := sqlcon.HandleError(tx.RawQuery(fmt.Sprintf("SELECT submit_count FROM %s WHERE id = ? AND nid = ?", flowTableName), fID, nid).First(&submitCount)); err != nil {
			if errors.Is(err, sqlcon.ErrNoRows) {
				// Return no error, as that would roll back the transaction
				return nil
			}

			return err
		}

		// This check prevents parallel brute force attacks to generate the recovery code
		// by checking the submit count inside this database transaction.
		// If the flow has been submitted more than 5 times, the transaction is aborted (regardless of whether the code was correct or not)
		// and we thus give no indication whether the supplied code was correct or not. See also https://github.com/ory/kratos/pull/2645#discussion_r984732899
		if submitCount > 5 {
			return errors.WithStack(code.ErrCodeSubmittedTooOften)
		}

		var recoveryCodes []code.RecoveryCode
		if err = sqlcon.HandleError(tx.Where("nid = ? AND selfservice_recovery_flow_id = ?", nid, fID).All(&recoveryCodes)); err != nil {
			if errors.Is(err, sqlcon.ErrNoRows) {
				// Return no error, as that would roll back the transaction
				return nil
			}

			return err
		}

	secrets:
		for _, secret := range p.r.Config().SecretsSession(ctx) {
			suppliedCode := []byte(p.hmacValueWithSecret(ctx, codeVal, secret))
			for i := range recoveryCodes {
				code := recoveryCodes[i]
				if subtle.ConstantTimeCompare([]byte(code.CodeHMAC), suppliedCode) == 0 {
					// Not the supplied code
					continue
				}
				recoveryCode = &code
				break secrets
			}
		}

		if recoveryCode == nil || !recoveryCode.IsValid() {
			// Return no error, as that would roll back the transaction
			return nil
		}

		var ra identity.RecoveryAddress
		if err := tx.Where("id = ? AND nid = ?", recoveryCode.RecoveryAddressID, nid).First(&ra); err != nil {
			if err = sqlcon.HandleError(err); !errors.Is(err, sqlcon.ErrNoRows) {
				return err
			}
		}
		recoveryCode.RecoveryAddress = &ra

		//#nosec G201 -- TableName is static
		return sqlcon.HandleError(tx.RawQuery(fmt.Sprintf("UPDATE %s SET used_at = ? WHERE id = ? AND nid = ?", recoveryCode.TableName(ctx)), time.Now().UTC(), recoveryCode.ID, nid).Exec())
	})); err != nil {
		return nil, err
	}

	if recoveryCode == nil {
		return nil, errors.WithStack(code.ErrCodeNotFound)
	}

	if recoveryCode.IsExpired() {
		return nil, errors.WithStack(flow.NewFlowExpiredError(recoveryCode.ExpiresAt))
	}

	if recoveryCode.WasUsed() {
		return nil, errors.WithStack(code.ErrCodeAlreadyUsed)
	}

	return recoveryCode, nil
}

func (p *Persister) DeleteRecoveryCodesOfFlow(ctx context.Context, fID uuid.UUID) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteRecoveryCodesOfFlow")
	defer span.End()

	//#nosec G201 -- TableName is static
	return p.GetConnection(ctx).RawQuery(fmt.Sprintf("DELETE FROM %s WHERE selfservice_recovery_flow_id = ? AND nid = ?", new(code.RecoveryCode).TableName(ctx)), fID, p.NetworkID(ctx)).Exec()
}
