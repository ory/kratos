// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"time"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/persistence/sql/batch"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/pop/v6"
	"github.com/ory/x/otelx"
	"github.com/ory/x/sqlcon"
)

func (p *Persister) CreateLoginCode(ctx context.Context, params *code.CreateLoginCodeParams) (_ *code.LoginCode, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateLoginCode")
	defer otelx.End(span, &err)

	now := time.Now().UTC()
	loginCode := &code.LoginCode{
		IdentityID:  params.IdentityID,
		Address:     params.Address,
		AddressType: params.AddressType,
		CodeHMAC:    p.hmacValue(ctx, params.RawCode),
		IssuedAt:    now,
		ExpiresAt:   now.UTC().Add(p.r.Config().SelfServiceCodeMethodLifespan(ctx)),
		FlowID:      params.FlowID,
		NID:         p.NetworkID(ctx),
		ID:          uuid.Nil,
	}

	// Extra columns are values the caller already knows for columns that are
	// not part of the OSS model (e.g. crdb_region on CockroachDB
	// multi-region). Writing them with the insert avoids the server-side
	// fallback (a column default or a lookup derived from a foreign key),
	// which for region columns costs a cross-region round trip per statement.
	if len(params.InsertExtraColumns) > 0 {
		if err := p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {
			return batch.Create(ctx,
				&batch.TracerConnection{Tracer: p.r.Tracer(ctx), Connection: tx},
				[]*code.LoginCode{loginCode},
				batch.WithExtraColumns(params.InsertExtraColumns))
		}); err != nil {
			return nil, sqlcon.HandleError(err)
		}
		return loginCode, nil
	}

	if err := p.GetConnection(ctx).Create(loginCode); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return loginCode, nil
}

func (p *Persister) UseLoginCode(ctx context.Context, flowID uuid.UUID, identityID uuid.UUID, userProvidedCode string) (_ *code.LoginCode, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UseLoginCode")
	defer otelx.End(span, &err)

	codeRow, err := useOneTimeCode[code.LoginCode](ctx, p, flowID, userProvidedCode, login.Flow{}.TableName(), "selfservice_login_flow_id", withCheckIdentityID(identityID))
	if err != nil {
		return nil, err
	}

	return codeRow, nil
}

func (p *Persister) GetUsedLoginCode(ctx context.Context, flowID uuid.UUID) (_ *code.LoginCode, err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetUsedLoginCode")
	defer otelx.End(span, &err)

	var loginCode code.LoginCode
	if err := p.Connection(ctx).Where("selfservice_login_flow_id = ? AND nid = ? AND used_at IS NOT NULL", flowID, p.NetworkID(ctx)).First(&loginCode); err != nil {
		return nil, sqlcon.HandleError(err)
	}
	return &loginCode, nil
}

func (p *Persister) DeleteLoginCodesOfFlow(ctx context.Context, flowID uuid.UUID) (err error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteLoginCodesOfFlow")
	defer otelx.End(span, &err)

	return p.GetConnection(ctx).Where("selfservice_login_flow_id = ? AND nid = ?", flowID, p.NetworkID(ctx)).Delete(&code.LoginCode{})
}
