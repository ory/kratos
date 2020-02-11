package sql

import (
	"context"

	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/registration"
)

func (p *Persister) CreateRegistrationRequest(ctx context.Context, r *registration.Request) error {
	return p.c.Eager().Create(r)
}

func (p *Persister) GetRegistrationRequest(ctx context.Context, id uuid.UUID) (*registration.Request, error) {
	var r registration.Request
	if err := p.c.Eager().Find(&r, id); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	if err := (&r).AfterFind(p.c); err != nil {
		return nil, err
	}

	return &r, nil
}

func (p *Persister) UpdateRegistrationRequest(ctx context.Context, id uuid.UUID, ct identity.CredentialsType, rm *registration.RequestMethod) error {
	rr, err := p.GetRegistrationRequest(ctx, id)
	if err != nil {
		return err
	}

	method, ok := rr.Methods[ct]
	if !ok {
		rm.RequestID = rr.ID
		rm.Method = ct
		return p.c.Save(rm)
	}

	method.Config = rm.Config
	return p.c.Save(method)
}
