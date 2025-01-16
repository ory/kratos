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
		return nil, errors.WithStack(herodot.ErrNotFound.WithReason("can't create a verification code without a verifiable address"))
	}

	verificationCode.VerifiableAddress = params.VerifiableAddress
	verificationCode.VerifiableAddressID = uuid.NullUUID{
		UUID:  params.VerifiableAddress.ID,
		Valid: true,
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

	codeRow, err := useOneTimeCode[code.VerificationCode](ctx, p, flowID, userProvidedCode, new(verification.Flow).TableName(ctx), "selfservice_verification_flow_id")
	if err != nil {
		return nil, err
	}

	var va identity.VerifiableAddress
	if err := p.Connection(ctx).Where("id = ? AND nid = ?", codeRow.VerifiableAddressID, p.NetworkID(ctx)).First(&va); err != nil {
		// This should fail on not found errors too, because the verifiable address must exist for the flow to work.
		return nil, sqlcon.HandleError(err)
	}
	codeRow.VerifiableAddress = &va

	return codeRow, nil
}

func (p *Persister) DeleteVerificationCodesOfFlow(ctx context.Context, fID uuid.UUID) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteVerificationCodesOfFlow")
	defer otelx.End(span, &err)

	return p.GetConnection(ctx).Where("selfservice_verification_flow_id = ? AND nid = ?", fID, p.NetworkID(ctx)).Delete(&code.VerificationCode{})
}
