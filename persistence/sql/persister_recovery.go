package sql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ory/kratos/corp"
	"github.com/ory/kratos/identity"

	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/strategy/link"
)

var _ recovery.FlowPersister = new(Persister)
var _ link.RecoveryTokenPersister = new(Persister)

func (p Persister) CreateRecoveryFlow(ctx context.Context, r *recovery.Flow) error {
	r.NID = corp.ContextualizeNID(ctx, p.nid)
	return p.GetConnection(ctx).Create(r)
}

func (p Persister) GetRecoveryFlow(ctx context.Context, id uuid.UUID) (*recovery.Flow, error) {
	var r recovery.Flow
	if err := p.GetConnection(ctx).Where("id = ? AND nid = ?", id, corp.ContextualizeNID(ctx, p.nid)).First(&r); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &r, nil
}

func (p Persister) UpdateRecoveryFlow(ctx context.Context, r *recovery.Flow) error {
	cp := *r
	cp.NID = corp.ContextualizeNID(ctx, p.nid)
	return p.update(ctx, cp)
}

func (p *Persister) CreateRecoveryToken(ctx context.Context, token *link.RecoveryToken) error {
	t := token.Token
	token.Token = p.hmacValue(ctx, t)
	token.NID = corp.ContextualizeNID(ctx, p.nid)

	// This should not create the request eagerly because otherwise we might accidentally create an address that isn't
	// supposed to be in the database.
	if err := p.GetConnection(ctx).Create(token); err != nil {
		return err
	}

	token.Token = t
	return nil
}

func (p *Persister) UseRecoveryToken(ctx context.Context, token string) (*link.RecoveryToken, error) {
	var rt link.RecoveryToken

	nid := corp.ContextualizeNID(ctx, p.nid)
	if err := sqlcon.HandleError(p.Transaction(ctx, func(ctx context.Context, tx *pop.Connection) (err error) {
		for _, secret := range p.r.Config(ctx).SecretsSession() {
			if err = tx.Where("token = ? AND nid = ? AND NOT used", p.hmacValueWithSecret(token, secret), nid).First(&rt); err != nil {
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

		var ra identity.RecoveryAddress
		if err := tx.Where("id = ? AND nid = ?", rt.RecoveryAddressID, nid).First(&ra); err != nil {
			return sqlcon.HandleError(err)
		}

		rt.RecoveryAddress = &ra

		/* #nosec G201 TableName is static */
		return tx.RawQuery(fmt.Sprintf("UPDATE %s SET used=true, used_at=? WHERE id=? AND nid = ?", rt.TableName(ctx)), time.Now().UTC(), rt.ID, nid).Exec()
	})); err != nil {
		return nil, err
	}

	return &rt, nil
}

func (p *Persister) DeleteRecoveryToken(ctx context.Context, token string) error {
	/* #nosec G201 TableName is static */
	return p.GetConnection(ctx).RawQuery(fmt.Sprintf("DELETE FROM %s WHERE token=? AND nid = ?", new(link.RecoveryToken).TableName(ctx)), token, corp.ContextualizeNID(ctx, p.nid)).Exec()
}
