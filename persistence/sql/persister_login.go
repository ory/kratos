package sql

import (
	"context"

	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
)

var _ login.RequestPersister = new(Persister)

func (p *Persister) CreateLoginRequest(_ context.Context, r *login.Request) error {
	return p.c.Eager().Create(r)
}

func (p *Persister) GetLoginRequest(_ context.Context, id uuid.UUID) (*login.Request, error) {
	var r login.Request
	if err := p.c.Eager().Find(&r, id); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	if err := (&r).AfterFind(p.c); err != nil {
		return nil, err
	}

	return &r, nil
}

func (p *Persister) UpdateLoginRequest(ctx context.Context, id uuid.UUID, ct identity.CredentialsType, rm *login.RequestMethod) error {
	rm.Method = ct
	rm.RequestID = id
	return p.c.Save(rm)
}
