package sql

import (
	"context"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/registration"
)

func (p Persister) CreateRegistrationRequest(_ context.Context, r *registration.Request) error {
	return p.c.Eager().Create(r)
}

func (p Persister) GetRegistrationRequest(_ context.Context, id uuid.UUID) (*registration.Request, error) {
	var r registration.Request
	if err := p.c.Eager().Find(&r, id); err != nil {
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

	return p.c.Eager().Update(r)
}
