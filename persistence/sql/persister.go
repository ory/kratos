package sql

import (
	"context"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/persistence"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
)

var _ persistence.Persister = new(Persister)

type Persister struct {
	c pop.Connection
}

func (p Persister) CreateRegistrationRequest(_ context.Context, r *registration.Request) error {
	return p.c.Create(r)
}

func (p Persister) GetRegistrationRequest(_ context.Context, id uuid.UUID) (*registration.Request, error) {
	var r registration.Request
	if err := p.c.Find(&r, id); err != nil {
		return nil, err
	}
	return &r, nil
}

func (p Persister) UpdateRegistrationRequest(ctx context.Context, id uuid.UUID, ct identity.CredentialsType, rm *registration.RequestMethod) error {
	var r registration.Request
	if err := p.c.Find(&r, id); err != nil {
		return err
	}

	r.Methods[ct] = rm

	return p.c.Update(r)
}

func (p Persister) CreateLoginRequest(_ context.Context, r *login.Request) error {
	return p.c.Create(r)
}

func (p Persister) GetLoginRequest(_ context.Context, id uuid.UUID) (*login.Request, error) {
	var r login.Request
	if err := p.c.Find(&r, id); err != nil {
		return nil, err
	}
	return &r, nil
}

func (p Persister) UpdateLoginRequest(ctx context.Context, id uuid.UUID, ct identity.CredentialsType, rm *login.RequestMethod) error {
	var r login.Request
	if err := p.c.Find(&r, id); err != nil {
		return err
	}

	r.Methods[ct] = rm

	return p.c.Update(r)
}
