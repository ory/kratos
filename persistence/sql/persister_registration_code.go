package sql

import (
	"context"
	"github.com/bxcodec/faker/v3/support/slice"
	"github.com/gofrs/uuid"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/x/sqlcon"
	"github.com/pkg/errors"
	"time"
)

func (p *Persister) CreateRegistrationCode(ctx context.Context, params *code.CreateRegistrationCodeParams) (*code.RegistrationCode, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.CreateRegistrationCode")
	defer span.End()

	now := time.Now().UTC()
	registrationCode := &code.RegistrationCode{
		Address:     params.Address,
		AddressType: params.AddressType,
		CodeHMAC:    p.hmacValue(ctx, params.RawCode),
		IssuedAt:    now,
		ExpiresAt:   now.UTC().Add(p.r.Config().SelfServiceCodeMethodLifespan(ctx)),
		FlowID:      params.FlowID,
		NID:         p.NetworkID(ctx),
		ID:          uuid.Nil,
	}

	if err := p.GetConnection(ctx).Create(registrationCode); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return registrationCode, nil
}

func (p *Persister) UseRegistrationCode(ctx context.Context, flowID uuid.UUID, userProvidedCode string, addresses ...string) (*code.RegistrationCode, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.UseRegistrationCode")
	defer span.End()

	var codeRow *code.RegistrationCode // This has to be nil per default
	if err := useOneTimeCode[*code.RegistrationCode](ctx, p, flowID, userProvidedCode, codeRow, new(registration.Flow).TableName(ctx), "selfservice_registration_flow_id"); err != nil {
		return nil, err
	}

	// ensure that the identifiers extracted from the traits are contained in the registration code
	if !slice.Contains(addresses, codeRow.Address) {
		return nil, errors.WithStack(code.ErrCodeNotFound)
	}

	return codeRow, nil
}

func (p *Persister) GetUsedRegistrationCode(ctx context.Context, flowID uuid.UUID) (*code.RegistrationCode, error) {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.GetUsedRegistrationCode")
	defer span.End()

	var registrationCode code.RegistrationCode
	if err := p.Connection(ctx).Where("selfservice_registration_flow_id = ? AND used_at IS NOT NULL AND nid = ?", flowID, p.NetworkID(ctx)).First(&registrationCode); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &registrationCode, nil
}

func (p *Persister) DeleteRegistrationCodesOfFlow(ctx context.Context, flowID uuid.UUID) error {
	ctx, span := p.r.Tracer(ctx).Tracer().Start(ctx, "persistence.sql.DeleteRegistrationCodesOfFlow")
	defer span.End()

	return p.GetConnection(ctx).Where("selfservice_registration_flow_id = ? AND nid = ?", flowID, p.NetworkID(ctx)).Delete(&code.RegistrationCode{})
}
