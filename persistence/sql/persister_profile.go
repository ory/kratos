package sql

import (
	"context"

	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/selfservice/flow/profile"
)

var _ profile.RequestPersister = new(Persister)

func (p *Persister) CreateProfileRequest(ctx context.Context, r *profile.Request) error {
	r.IdentityID = r.Identity.ID
	return sqlcon.HandleError(p.GetConnection(ctx).Eager("MethodsRaw").Create(r))
}

func (p *Persister) GetProfileRequest(ctx context.Context, id uuid.UUID) (*profile.Request, error) {
	var r profile.Request
	if err := p.GetConnection(ctx).Eager().Find(&r, id); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	if err := (&r).AfterFind(p.GetConnection(ctx)); err != nil {
		return nil, err
	}

	return &r, nil
}

func (p *Persister) UpdateProfileRequest(ctx context.Context, r *profile.Request) error {
	return p.Transaction(ctx, func(tx *pop.Connection) error {
		ctx := WithTransaction(ctx, tx)
		rr, err := p.GetProfileRequest(ctx, r.ID)
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
