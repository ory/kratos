// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/x/otelx"
	"github.com/ory/x/sqlcon"
)

func (p *Persister) CreateVerificationCode(ctx context.Context, params *code.CreateVerificationCodeParams) (_ *code.VerificationCode, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateVerificationCode")
	defer otelx.End(span, &err)

	now := time.Now().UTC()
	verificationCode := &code.VerificationCode{
		ID:        uuid.Nil,
		CodeHMAC:  p.hmacValue(ctx, params.RawCode),
		ExpiresAt: now.Add(params.ExpiresIn),
		IssuedAt:  now,
		FlowID:    params.FlowID,
		NID:       p.NetworkID(ctx),
	}

	if params.VerifiableAddress == nil {
		return nil, errors.WithStack(herodot.ErrNotFound().WithReason("can't create a verification code without a verifiable address"))
	}

	// Can be nil when the code is for a pending change, in which case the address doesn't exist yet. See VerifyNewAddress hook.
	if va, ok := params.VerifiableAddress.ToPersistable(); ok {
		verificationCode.VerifiableAddress = va
		verificationCode.VerifiableAddressID = uuid.NullUUID{UUID: va.ID, Valid: true}
	} else {
		verificationCode.VerifiableAddressID = uuid.NullUUID{Valid: false}
	}

	// This should not create the request eagerly because otherwise we might accidentally create an address that isn't
	// supposed to be in the database.
	if err := p.GetConnection(ctx).Create(verificationCode); err != nil {
		return nil, err
	}

	return verificationCode, nil
}

func (p *Persister) UseVerificationCode(ctx context.Context, flowID uuid.UUID, userProvidedCode string) (_ *code.VerificationCode, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UseVerificationCode")
	defer otelx.End(span, &err)

	codeRow, err := useOneTimeCode[code.VerificationCode](ctx, p, flowID, userProvidedCode, verification.Flow{}.TableName(), "selfservice_verification_flow_id")
	if err != nil {
		return nil, err
	}

	if codeRow.VerifiableAddressID.Valid {
		var va identity.VerifiableAddress
		if err := p.Connection(ctx).Where("id = ? AND nid = ?", codeRow.VerifiableAddressID, p.NetworkID(ctx)).First(&va); err != nil {
			return nil, sqlcon.HandleError(err)
		}
		codeRow.VerifiableAddress = &va
	}

	return codeRow, nil
}

func (p *Persister) DeleteVerificationCodesOfFlow(ctx context.Context, fID uuid.UUID) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteVerificationCodesOfFlow")
	defer otelx.End(span, &err)

	return p.GetConnection(ctx).Where("selfservice_verification_flow_id = ? AND nid = ?", fID, p.NetworkID(ctx)).Delete(&code.VerificationCode{})
}
