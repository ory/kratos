// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"crypto/subtle"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/persistence/sql/update"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/kratos/selfservice/strategy/link"
)

var _ verification.FlowPersister = new(Persister)

func (p *Persister) CreateVerificationFlow(ctx context.Context, r *verification.Flow) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateVerificationFlow")
	defer span.End()

	r.NID = p.NetworkID(ctx)
	// This should not create the request eagerly because otherwise we might accidentally create an address
	// that isn't supposed to be in the database.
	return p.GetConnection(ctx).Create(r)
}

func (p *Persister) GetVerificationFlow(ctx context.Context, id uuid.UUID) (*verification.Flow, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetVerificationFlow")
	defer span.End()

	var r verification.Flow
	if err := p.GetConnection(ctx).Where("id = ? AND nid = ?", id, p.NetworkID(ctx)).First(&r); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &r, nil
}

func (p *Persister) UpdateVerificationFlow(ctx context.Context, r *verification.Flow) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UpdateVerificationFlow")
	defer span.End()

	cp := *r
	cp.NID = p.NetworkID(ctx)
	return update.Generic(ctx, p.GetConnection(ctx), p.r.Tracer(ctx).Tracer(), cp)
}

func (p *Persister) CreateVerificationToken(ctx context.Context, token *link.VerificationToken) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateVerificationToken")
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

func (p *Persister) UseVerificationToken(ctx context.Context, fID uuid.UUID, token string) (*link.VerificationToken, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UseVerificationToken")
	defer span.End()

	var rt link.VerificationToken

	nid := p.NetworkID(ctx)
	if err := sqlcon.HandleError(p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) (err error) {
		for _, secret := range p.r.Config().SecretsSession(ctx) {
			if err = tx.Where("token = ? AND nid = ? AND NOT used AND selfservice_verification_flow_id = ?", p.hmacValueWithSecret(ctx, token, secret), nid, fID).First(&rt); err != nil {
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

func (p *Persister) DeleteVerificationToken(ctx context.Context, token string) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteVerificationToken")
	defer span.End()

	nid := p.NetworkID(ctx)
	//#nosec G201 -- TableName is static
	return p.GetConnection(ctx).RawQuery(fmt.Sprintf("DELETE FROM %s WHERE token=? AND nid = ?", new(link.VerificationToken).TableName(ctx)), token, nid).Exec()
}

func (p *Persister) DeleteExpiredVerificationFlows(ctx context.Context, expiresAt time.Time, limit int) error {
	//#nosec G201 -- TableName is static
	err := p.GetConnection(ctx).RawQuery(fmt.Sprintf(
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
func (p *Persister) UseVerificationCode(ctx context.Context, fID uuid.UUID, codeVal string) (*code.VerificationCode, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UseVerificationCode")
	defer span.End()

	var verificationCode *code.VerificationCode

	nid := p.NetworkID(ctx)

	flowTableName := new(verification.Flow).TableName(ctx)

	if err := sqlcon.HandleError(p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) (err error) {

		if err := sqlcon.HandleError(
			tx.RawQuery(
				//#nosec G201 -- TableName is static
				fmt.Sprintf("UPDATE %s SET submit_count = submit_count + 1 WHERE id = ? AND nid = ?", flowTableName),
				fID,
				nid,
			).Exec(),
		); err != nil {
			return err
		}

		var submitCount int
		// Because MySQL does not support "RETURNING" clauses, but we need the updated `submit_count` later on.
		if err := sqlcon.HandleError(
			tx.RawQuery(
				//#nosec G201 -- TableName is static
				fmt.Sprintf("SELECT submit_count FROM %s WHERE id = ? AND nid = ?", flowTableName),
				fID,
				nid,
			).First(&submitCount),
		); err != nil {
			if errors.Is(err, sqlcon.ErrNoRows) {
				// Return no error, as that would roll back the transaction
				return nil
			}

			return err
		}
		// This check prevents parallel brute force attacks to generate the verification code
		// by checking the submit count inside this database transaction.
		// If the flow has been submitted more than 5 times, the transaction is aborted (regardless of whether the code was correct or not)
		// and we thus give no indication whether the supplied code was correct or not. See also https://github.com/ory/kratos/pull/2645#discussion_r984732899
		if submitCount > 5 {
			return errors.WithStack(code.ErrCodeSubmittedTooOften)
		}

		var verificationCodes []code.VerificationCode
		if err = sqlcon.HandleError(
			tx.Where("nid = ? AND selfservice_verification_flow_id = ?", nid, fID).
				All(&verificationCodes),
		); err != nil {
			if errors.Is(err, sqlcon.ErrNoRows) {
				// Return no error, as that would roll back the transaction
				return nil
			}

			return err
		}

	secrets:
		for _, secret := range p.r.Config().SecretsSession(ctx) {
			suppliedCode := []byte(p.hmacValueWithSecret(ctx, codeVal, secret))
			for i := range verificationCodes {
				code := verificationCodes[i]
				if subtle.ConstantTimeCompare([]byte(code.CodeHMAC), suppliedCode) == 0 {
					// Not the supplied code
					continue
				}
				verificationCode = &code
				break secrets
			}
		}

		if verificationCode == nil || verificationCode.Validate() != nil {
			// Return no error, as that would roll back the transaction
			return nil
		}

		var va identity.VerifiableAddress
		if err := tx.Where("id = ? AND nid = ?", verificationCode.VerifiableAddressID, nid).First(&va); err != nil {
			return sqlcon.HandleError(err)
		}

		verificationCode.VerifiableAddress = &va

		//#nosec G201 -- TableName is static
		return tx.
			RawQuery(
				fmt.Sprintf("UPDATE %s SET used_at = ? WHERE id = ? AND nid = ?", verificationCode.TableName(ctx)),
				time.Now().UTC(),
				verificationCode.ID,
				nid,
			).Exec()
	})); err != nil {
		return nil, err
	}

	if verificationCode == nil {
		return nil, errors.WithStack(code.ErrCodeNotFound)
	}

	return verificationCode, nil
}

func (p *Persister) CreateVerificationCode(ctx context.Context, c *code.CreateVerificationCodeParams) (*code.VerificationCode, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateVerificationCode")
	defer span.End()

	now := time.Now().UTC()

	verificationCode := &code.VerificationCode{
		ID:        uuid.Nil,
		CodeHMAC:  p.hmacValue(ctx, c.RawCode),
		ExpiresAt: now.Add(c.ExpiresIn),
		IssuedAt:  now,
		FlowID:    c.FlowID,
		NID:       p.NetworkID(ctx),
	}

	if c.VerifiableAddress == nil {
		return nil, errors.WithStack(herodot.ErrNotFound.WithReason("can't create a verification code without a verifiable address"))
	}

	verificationCode.VerifiableAddress = c.VerifiableAddress
	verificationCode.VerifiableAddressID = uuid.NullUUID{
		UUID:  c.VerifiableAddress.ID,
		Valid: true,
	}

	// This should not create the request eagerly because otherwise we might accidentally create an address that isn't
	// supposed to be in the database.
	if err := p.GetConnection(ctx).Create(verificationCode); err != nil {
		return nil, err
	}
	return verificationCode, nil
}

func (p *Persister) DeleteVerificationCodesOfFlow(ctx context.Context, fID uuid.UUID) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteVerificationCodesOfFlow")
	defer span.End()

	//#nosec G201 -- TableName is static
	return p.GetConnection(ctx).
		RawQuery(
			fmt.Sprintf("DELETE FROM %s WHERE selfservice_verification_flow_id = ? AND nid = ?", new(code.VerificationCode).TableName(ctx)),
			fID,
			p.NetworkID(ctx),
		).Exec()
}
