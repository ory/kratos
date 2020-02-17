package sql

import (
	"context"

	"github.com/gobuffalo/pop/v5"

	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
)

var _ login.RequestPersister = new(Persister)

func (p *Persister) CreateLoginRequest(ctx context.Context, r *login.Request) error {
	return p.GetConnection(ctx).Eager().Create(r)
}

func (p *Persister) GetLoginRequest(ctx context.Context, id uuid.UUID) (*login.Request, error) {
	conn := p.GetConnection(ctx)
	var r login.Request
	if err := conn.Eager().Find(&r, id); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	if err := (&r).AfterFind(conn); err != nil {
		return nil, err
	}

	return &r, nil
}

func (p *Persister) UpdateLoginRequest(ctx context.Context, id uuid.UUID, ct identity.CredentialsType, rm *login.RequestMethod) error {
	return p.Transaction(ctx, func(tx *pop.Connection) error {
		ctx := WithTransaction(ctx, tx)
		rr, err := p.GetLoginRequest(ctx, id)
		if err != nil {
			return err
		}

		method, ok := rr.Methods[ct]
		if !ok {
			rm.RequestID = rr.ID
			rm.Method = ct
			return tx.Save(rm)
		}

		method.Config = rm.Config
		return tx.Save(method)
	})
}
