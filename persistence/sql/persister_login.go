package sql

import (
	"context"
	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
)

var _ login.RequestPersister = new(Persister)

func (p *Persister) CreateLoginRequest(ctx context.Context, r *login.Request) error {
	return p.c.Eager().Create(r)
}

func (p *Persister) GetLoginRequest(ctx context.Context, id uuid.UUID) (*login.Request, error) {
	var r login.Request
	if err := p.c.Eager().Find(&r, id); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	if err := (&r).AfterFind(p.c); err != nil {
		return nil, err
	}

	return &r, nil
}

func (p *Persister) UpdateLoginRequestReauth(ctx context.Context, id uuid.UUID, reauth bool) error {
	lr, err := p.GetLoginRequest(ctx, id)
	if err != nil {
		return err
	}

	lr.IsReauthentication = reauth
	return p.c.Save(lr)
}

func (p *Persister) UpdateLoginRequestMethod(ctx context.Context, id uuid.UUID, ct identity.CredentialsType, rm *login.RequestMethod) error {
	rr, err := p.GetLoginRequest(ctx, id)
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
