package sql

import (
	"context"

	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/registration"
)

func (p *Persister) CreateRegistrationRequest(ctx context.Context, r *registration.Request) error {
	return p.GetConnection(ctx).Eager().Create(r)
}

func (p *Persister) UpdateRegistrationRequest(ctx context.Context, r *registration.Request) error {
	return p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {

		rr, err := p.GetRegistrationRequest(ctx, r.ID)
		if err != nil {
			return err
		}

		for id, form := range r.Methods {
			for oid := range rr.Methods {
				if oid == id {
					rr.Methods[id].Config = form.Config
					break
				}
			}
			rr.Methods[id] = form
		}

		for _, of := range rr.Methods {
			if err := tx.Save(of); err != nil {
				return sqlcon.HandleError(err)
			}
		}

		return tx.Save(r)
	})
}

func (p *Persister) GetRegistrationRequest(ctx context.Context, id uuid.UUID) (*registration.Request, error) {
	var r registration.Request
	if err := p.GetConnection(ctx).Eager().Find(&r, id); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	if err := (&r).AfterFind(p.GetConnection(ctx)); err != nil {
		return nil, err
	}

	return &r, nil
}

func (p *Persister) UpdateRegistrationRequestMethod(ctx context.Context, id uuid.UUID, ct identity.CredentialsType, rm *registration.RequestMethod) error {
	return p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {

		rr, err := p.GetRegistrationRequest(ctx, id)
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
		if err := tx.Save(method); err != nil {
			return err
		}

		rr.Active = ct
		return tx.Save(rr)
	})
}
