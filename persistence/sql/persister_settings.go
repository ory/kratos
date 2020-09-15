package sql

import (
	"context"

	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlxx"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/selfservice/flow/settings"
)

var _ settings.FlowPersister = new(Persister)

func (p *Persister) CreateSettingsFlow(ctx context.Context, r *settings.Flow) error {
	return sqlcon.HandleError(p.GetConnection(ctx).Eager("MethodsRaw").Create(r))
}

func (p *Persister) GetSettingsFlow(ctx context.Context, id uuid.UUID) (*settings.Flow, error) {
	var r settings.Flow
	if err := p.GetConnection(ctx).Eager().Find(&r, id); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	if err := (&r).AfterFind(p.GetConnection(ctx)); err != nil {
		return nil, err
	}

	return &r, nil
}

func (p *Persister) UpdateSettingsFlow(ctx context.Context, r *settings.Flow) error {
	return p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {

		rr, err := p.GetSettingsFlow(ctx, r.ID)
		if err != nil {
			return err
		}

		for id, form := range r.Methods {
			var found bool
			for oid := range rr.Methods {
				if oid == id {
					rr.Methods[id].Config = form.Config
					found = true
					break
				}
			}
			if !found {
				rr.Methods[id] = form
			}
		}

		for _, of := range rr.Methods {
			if err := tx.Save(of); err != nil {
				return sqlcon.HandleError(err)
			}
		}

		return tx.Save(r)
	})
}

func (p *Persister) UpdateSettingsFlowMethod(ctx context.Context, id uuid.UUID, method string, fm *settings.FlowMethod) error {
	return p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {
		rr, err := p.GetSettingsFlow(ctx, id)
		if err != nil {
			return err
		}

		m, ok := rr.Methods[method]
		if !ok {
			fm.FlowID = rr.ID
			fm.Method = method
			return tx.Save(fm)
		}

		m.Config = fm.Config
		if err := tx.Save(m); err != nil {
			return err
		}

		rr.Active = sqlxx.NullString(method)
		return tx.Save(rr)
	})
}
