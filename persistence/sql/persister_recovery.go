package sql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/strategy/link"
)

var _ recovery.RequestPersister = new(Persister)
var _ link.Persister = new(Persister)

func (p Persister) CreateRecoveryRequest(ctx context.Context, r *recovery.Request) error {
	return p.GetConnection(ctx).Eager("MethodsRaw").Create(r)
}

func (p Persister) GetRecoveryRequest(ctx context.Context, id uuid.UUID) (*recovery.Request, error) {
	var r recovery.Request
	if err := p.GetConnection(ctx).Eager().Find(&r, id); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	if err := (&r).AfterFind(p.GetConnection(ctx)); err != nil {
		return nil, err
	}

	return &r, nil
}

func (p Persister) UpdateRecoveryRequest(ctx context.Context, r *recovery.Request) error {
	return p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {

		rr, err := p.GetRecoveryRequest(ctx, r.ID)
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

func (p *Persister) CreateRecoveryToken(ctx context.Context, token *link.Token) error {
	t := token.Token
	token.Token = p.hmacValue(t)

	// This should not create the request eagerly because otherwise we might accidentally create an address that isn't
	// supposed to be in the database.
	if err := p.GetConnection(ctx).Create(token); err != nil {
		return err
	}
	token.Token = t
	return nil
}

func (p *Persister) UseRecoveryToken(ctx context.Context, token string) (*link.Token, error) {
	rt := new(link.Token)
	if err := sqlcon.HandleError(p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) (err error) {
		for _, secret := range p.cf.SecretsSession() {
			if err = tx.Eager().Where("token = ? AND NOT used", p.hmacValueWithSecret(token, secret)).First(rt); err != nil {
				if !errors.Is(sqlcon.HandleError(err), sqlcon.ErrNoRows) {
					return err
				}
			} else {
				break
			}
		}
		if err != nil {
			return err
		}
		/* #nosec G201 TableName is static */
		return tx.RawQuery(fmt.Sprintf("UPDATE %s SET used=true, used_at=? WHERE id=?", rt.TableName()), time.Now().UTC(), rt.ID).Exec()
	})); err != nil {
		return nil, err
	}
	return rt, nil
}

func (p *Persister) DeleteRecoveryToken(ctx context.Context, token string) error {
	/* #nosec G201 TableName is static */
	return p.GetConnection(ctx).RawQuery(fmt.Sprintf("DELETE FROM %s WHERE token=?", new(link.Token).TableName()), token).Exec()
}
