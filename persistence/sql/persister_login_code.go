package sql

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/x/sqlcon"
	"time"
)

func (p *Persister) CreateLoginCode(ctx context.Context, params *code.CreateLoginCodeParams) (*code.LoginCode, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateLoginCode")
	defer span.End()

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

	if err := p.GetConnection(ctx).Create(loginCode); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return loginCode, nil
}

func (p *Persister) UseLoginCode(ctx context.Context, flowID uuid.UUID, identityID uuid.UUID, userProvidedCode string) (*code.LoginCode, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UseLoginCode")
	defer span.End()

	codeRow, err := useOneTimeCode[code.LoginCode, *code.LoginCode](ctx, p, flowID, userProvidedCode, new(login.Flow).TableName(ctx), "selfservice_login_flow_id")
	if err != nil {
		return nil, err
	}

	panic(`missing: tx.Where("nid = ? AND selfservice_login_flow_id = ? AND identity_id = ?", nid, flowID, identityID).All(&codes)); err != nil {`)

	return codeRow, nil
}

func (p *Persister) GetUsedLoginCode(ctx context.Context, flowID uuid.UUID) (*code.LoginCode, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetUsedLoginCode")
	defer span.End()

	var loginCode code.LoginCode
	if err := p.Connection(ctx).Where("selfservice_login_flow_id = ? AND nid = ? AND used_at IS NOT NULL", flowID, p.NetworkID(ctx)).First(&loginCode); err != nil {
		return nil, sqlcon.HandleError(err)
	}
	return &loginCode, nil
}

func (p *Persister) DeleteLoginCodesOfFlow(ctx context.Context, flowID uuid.UUID) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteLoginCodesOfFlow")
	defer span.End()

	return p.GetConnection(ctx).Where("selfservice_login_flow_id = ? AND nid = ?", flowID, p.NetworkID(ctx)).Delete(&code.LoginCode{})
}
