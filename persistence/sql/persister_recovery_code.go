// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/x/otelx"
	"github.com/ory/x/sqlcon"
)

func (p *Persister) CreateRecoveryCode(ctx context.Context, params *code.CreateRecoveryCodeParams) (_ *code.RecoveryCode, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateRecoveryCode")
	defer otelx.End(span, &err)

	now := time.Now()
	recoveryCode := &code.RecoveryCode{
		ID:         uuid.Nil,
		CodeHMAC:   p.hmacValue(ctx, params.RawCode),
		ExpiresAt:  now.UTC().Add(params.ExpiresIn),
		IssuedAt:   now,
		CodeType:   params.CodeType,
		FlowID:     params.FlowID,
		NID:        p.NetworkID(ctx),
		IdentityID: params.IdentityID,
	}

	if params.RecoveryAddress != nil {
		recoveryCode.RecoveryAddress = params.RecoveryAddress
		recoveryCode.RecoveryAddressID = uuid.NullUUID{
			UUID:  params.RecoveryAddress.ID,
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
func (p *Persister) UseRecoveryCode(ctx context.Context, flowID uuid.UUID, userProvidedCode string) (_ *code.RecoveryCode, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UseRecoveryCode")
	defer otelx.End(span, &err)

	codeRow, err := useOneTimeCode[code.RecoveryCode](ctx, p, flowID, userProvidedCode, new(recovery.Flow).TableName(ctx), "selfservice_recovery_flow_id")
	if err != nil {
		return nil, err
	}

	var ra identity.RecoveryAddress
	if err := sqlcon.HandleError(p.GetConnection(ctx).Where("id = ? AND nid = ?", codeRow.RecoveryAddressID, p.NetworkID(ctx)).First(&ra)); err != nil {
		if errors.Is(err, sqlcon.ErrNoRows) {
			// This is ok, it can happen when an administrator initiates account recovery. This works even if the
			// user has no recovery address!
		} else {
			return nil, err
		}
	}
	codeRow.RecoveryAddress = &ra

	return codeRow, nil
}

func (p *Persister) DeleteRecoveryCodesOfFlow(ctx context.Context, flowID uuid.UUID) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteRecoveryCodesOfFlow")
	defer otelx.End(span, &err)

	return p.GetConnection(ctx).Where("selfservice_recovery_flow_id = ? AND nid = ?", flowID, p.NetworkID(ctx)).Delete(&code.RecoveryCode{})
}
