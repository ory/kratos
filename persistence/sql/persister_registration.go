// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"crypto/subtle"
	"fmt"
	"time"

	"github.com/bxcodec/faker/v3/support/slice"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/persistence/sql/update"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/strategy/code"
)

func (p *Persister) CreateRegistrationFlow(ctx context.Context, r *registration.Flow) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateRegistrationFlow")
	defer span.End()

	r.NID = p.NetworkID(ctx)
	r.EnsureInternalContext()
	return p.GetConnection(ctx).Create(r)
}

func (p *Persister) UpdateRegistrationFlow(ctx context.Context, r *registration.Flow) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UpdateRegistrationFlow")
	defer span.End()

	r.EnsureInternalContext()
	cp := *r
	cp.NID = p.NetworkID(ctx)
	return update.Generic(ctx, p.GetConnection(ctx), p.r.Tracer(ctx).Tracer(), cp)
}

func (p *Persister) GetRegistrationFlow(ctx context.Context, id uuid.UUID) (*registration.Flow, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetRegistrationFlow")
	defer span.End()

	var r registration.Flow
	if err := p.GetConnection(ctx).Where("id = ? AND nid = ?",
		id, p.NetworkID(ctx)).First(&r); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &r, nil
}

func (p *Persister) DeleteExpiredRegistrationFlows(ctx context.Context, expiresAt time.Time, limit int) error {
	//#nosec G201 -- TableName is static
	err := p.GetConnection(ctx).RawQuery(fmt.Sprintf(
		"DELETE FROM %s WHERE id in (SELECT id FROM (SELECT id FROM %s c WHERE expires_at <= ? and nid = ? ORDER BY expires_at ASC LIMIT %d ) AS s )",
		new(registration.Flow).TableName(ctx),
		new(registration.Flow).TableName(ctx),
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

func (p *Persister) CreateRegistrationCode(ctx context.Context, codeParams *code.CreateRegistrationCodeParams) (*code.RegistrationCode, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateRegistrationCode")
	defer span.End()

	now := time.Now()

	registrationCode := &code.RegistrationCode{
		Address:     codeParams.Address,
		AddressType: codeParams.AddressType,
		CodeHMAC:    p.hmacValue(ctx, codeParams.RawCode),
		IssuedAt:    now,
		ExpiresAt:   now.UTC().Add(p.r.Config().SelfServiceCodeMethodLifespan(ctx)),
		FlowID:      codeParams.FlowID,
		NID:         p.NetworkID(ctx),
		ID:          uuid.Nil,
	}

	if err := p.GetConnection(ctx).Create(registrationCode); err != nil {
		return nil, sqlcon.HandleError(err)
	}
	return registrationCode, nil
}

func (p *Persister) UseRegistrationCode(ctx context.Context, flowID uuid.UUID, rawCode string, addresses ...string) (*code.RegistrationCode, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UseRegistrationCode")
	defer span.End()

	nid := p.NetworkID(ctx)

	flowTableName := new(registration.Flow).TableName(ctx)

	var registrationCode *code.RegistrationCode
	if err := sqlcon.HandleError(p.GetConnection(ctx).Transaction(func(tx *pop.Connection) error {
		if err := tx.RawQuery(fmt.Sprintf("UPDATE %s SET submit_count = submit_count + 1 WHERE id = ? AND nid = ?", flowTableName), flowID, nid).Exec(); err != nil {
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

		// This check prevents parallel brute force attacks to generate the recovery code
		// by checking the submit count inside this database transaction.
		// If the flow has been submitted more than 5 times, the transaction is aborted (regardless of whether the code was correct or not)
		// and we thus give no indication whether the supplied code was correct or not. See also https://github.com/ory/kratos/pull/2645#discussion_r984732899
		if submitCount > 5 {
			return errors.WithStack(code.ErrCodeSubmittedTooOften)
		}

		var registrationCodes []code.RegistrationCode
		if err := sqlcon.HandleError(tx.Where("nid = ? AND selfservice_registration_flow_id = ?", nid, flowID).All(&registrationCodes)); err != nil {
			if errors.Is(err, sqlcon.ErrNoRows) {
				// Return no error, as that would roll back the transaction
				return nil
			}

			return err
		}

	secrets:
		for _, secret := range p.r.Config().SecretsSession(ctx) {
			suppliedCode := []byte(p.hmacValueWithSecret(ctx, rawCode, secret))
			for i := range registrationCodes {
				code := registrationCodes[i]
				if subtle.ConstantTimeCompare([]byte(code.CodeHMAC), suppliedCode) == 0 {
					// Not the supplied code
					continue
				}
				registrationCode = &code
				break secrets
			}
		}

		if registrationCode == nil || !registrationCode.IsValid() {
			// Return no error, as that would roll back the transaction
			return nil
		}

		//#nosec G201 -- TableName is static
		return sqlcon.HandleError(tx.RawQuery(fmt.Sprintf("UPDATE %s SET used_at = ? WHERE id = ? AND nid = ?", registrationCode.TableName(ctx)), time.Now().UTC(), registrationCode.ID, nid).Exec())
	})); err != nil {
		return nil, err
	}

	if registrationCode == nil {
		return nil, errors.WithStack(code.ErrCodeNotFound)
	}

	if registrationCode.IsExpired() {
		return nil, errors.WithStack(flow.NewFlowExpiredError(registrationCode.ExpiresAt))
	}

	if registrationCode.WasUsed() {
		return nil, errors.WithStack(code.ErrCodeAlreadyUsed)
	}

	// ensure that the identifiers extracted from the traits are contained in the registration code
	if !slice.Contains(addresses, registrationCode.Address) {
		return nil, errors.WithStack(code.ErrCodeNotFound)
	}

	return registrationCode, nil
}

func (p *Persister) DeleteRegistrationCodesOfFlow(ctx context.Context, flowID uuid.UUID) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteRegistrationCodesOfFlow")
	defer span.End()

	//#nosec G201 -- TableName is static
	return p.GetConnection(ctx).RawQuery(fmt.Sprintf("DELETE FROM %s WHERE selfservice_registration_flow_id = ? AND nid = ?", new(code.RegistrationCode).TableName(ctx)), flowID, p.NetworkID(ctx)).Exec()
}

func (p *Persister) GetUsedRegistrationCode(ctx context.Context, flowID uuid.UUID) (*code.RegistrationCode, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetUsedRegistrationCode")
	defer span.End()

	var registrationCode code.RegistrationCode
	if err := p.Connection(ctx).RawQuery(fmt.Sprintf("SELECT * FROM %s WHERE selfservice_registration_flow_id = ? AND used_at IS NOT NULL AND nid = ?", new(code.RegistrationCode).TableName(ctx)), flowID, p.NetworkID(ctx)).First(&registrationCode); err != nil {
		return nil, sqlcon.HandleError(err)
	}
	return &registrationCode, nil
}
