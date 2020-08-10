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

func (p *Persister) CreateLoginRequest(ctx context.Context, r *login.Flow) error {
	return p.GetConnection(ctx).Eager().Create(r)
}

func (p *Persister) UpdateLoginRequest(ctx context.Context, r *login.Flow) error {
	return p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {

		rr, err := p.GetLoginRequest(ctx, r.ID)
		if err != nil {
			return err
		}

		for _, dbc := range rr.Methods {
			if err := tx.Destroy(dbc); err != nil {
				return sqlcon.HandleError(err)
			}
		}

		for _, of := range r.Methods {
			of.FlowID = r.ID
			if err := tx.Save(of); err != nil {
				return sqlcon.HandleError(err)
			}
		}

		return tx.Save(r)
	})
}

func (p *Persister) GetLoginRequest(ctx context.Context, id uuid.UUID) (*login.Flow, error) {
	conn := p.GetConnection(ctx)
	var r login.Flow
	if err := conn.Eager().Find(&r, id); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	if err := (&r).AfterFind(conn); err != nil {
		return nil, err
	}

	return &r, nil
}

func (p *Persister) MarkRequestForced(ctx context.Context, id uuid.UUID) error {
	return p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {

		lr, err := p.GetLoginRequest(ctx, id)
		if err != nil {
			return err
		}

		lr.Forced = true
		return tx.Save(lr)
	})
}

func (p *Persister) UpdateLoginRequestMethod(ctx context.Context, id uuid.UUID, ct identity.CredentialsType, rm *login.FlowMethod) error {
	return p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {

		rr, err := p.GetLoginRequest(ctx, id)
		if err != nil {
			return err
		}

		method, ok := rr.Methods[ct]
		if !ok {
			rm.FlowID = rr.ID
			rm.Method = ct
			return tx.Save(rm)
		}

		method.Config = rm.Config
		if err := tx.Save(method); err != nil {
			return err
		}

		rr.Active = ct
		return tx.Save(rr)
	})
}
