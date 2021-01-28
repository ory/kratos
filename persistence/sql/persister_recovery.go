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

var _ recovery.FlowPersister = new(Persister)
var _ link.RecoveryTokenPersister = new(Persister)

func (p Persister) CreateRecoveryFlow(ctx context.Context, r *recovery.Flow) error {
	return p.GetConnection(ctx).Eager("MethodsRaw").Create(r)
}

func (p Persister) GetRecoveryFlow(ctx context.Context, id uuid.UUID) (*recovery.Flow, error) {
	var r recovery.Flow
	if err := p.GetConnection(ctx).Eager().Find(&r, id); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	if err := (&r).AfterFind(p.GetConnection(ctx)); err != nil {
		return nil, err
	}

	return &r, nil
}

func (p Persister) UpdateRecoveryFlow(ctx context.Context, r *recovery.Flow) error {
	return p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) error {

		rr, err := p.GetRecoveryFlow(ctx, r.ID)
		if err != nil {
			return err
		}

		for _, dbc := range rr.Methods {
			if err := tx.Destroy(dbc); err != nil {
				return sqlcon.HandleError(err)
			}
		}

		for _, of := range r.Methods {
			of.ID = uuid.UUID{}
			of.Flow = rr
			of.FlowID = rr.ID
			if err := tx.Save(of); err != nil {
				return sqlcon.HandleError(err)
			}
		}

		return tx.Save(r)
	})
}

func (p *Persister) CreateRecoveryToken(ctx context.Context, token *link.RecoveryToken) error {
	t := token.Token
	token.Token = p.hmacValue(ctx, t)

	// This should not create the request eagerly because otherwise we might accidentally create an address that isn't
	// supposed to be in the database.
	if err := p.GetConnection(ctx).Create(token); err != nil {
		return err
	}
	token.Token = t
	return nil
}

func (p *Persister) UseRecoveryToken(ctx context.Context, token string) (*link.RecoveryToken, error) {
	var err error
	rt := new(link.RecoveryToken)
	if err = sqlcon.HandleError(p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) (err error) {
		for _, secret := range p.r.Config(ctx).SecretsSession() {
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
		return tx.RawQuery(fmt.Sprintf("UPDATE %s SET used=true, used_at=? WHERE id=?", rt.TableName(ctx)), time.Now().UTC(), rt.ID).Exec()
	})); err != nil {
		return nil, err
	}

	return rt, nil
}

func (p *Persister) DeleteRecoveryToken(ctx context.Context, token string) error {
	/* #nosec G201 TableName is static */
	return p.GetConnection(ctx).RawQuery(fmt.Sprintf("DELETE FROM %s WHERE token=?", new(link.RecoveryToken).TableName(ctx)), token).Exec()
}
