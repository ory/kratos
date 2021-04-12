package sql

import (
	"context"
	"errors"
	"fmt"
	"github.com/ory/kratos/corp"
	"github.com/ory/kratos/identity"
	"time"

	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/strategy/link"
)

var _ verification.FlowPersister = new(Persister)

func (p Persister) CreateVerificationFlow(ctx context.Context, r *verification.Flow) error {
	r.NID = corp.ContextualizeNID(ctx, p.nid)
	// This should not create the request eagerly because otherwise we might accidentally create an address
	// that isn't supposed to be in the database.
	return p.GetConnection(ctx).Create(r)
}

func (p Persister) GetVerificationFlow(ctx context.Context, id uuid.UUID) (*verification.Flow, error) {
	var r verification.Flow
	if err := p.GetConnection(ctx).Where("id = ? AND nid = ?", id, corp.ContextualizeNID(ctx, p.nid)).First(&r); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &r, nil
}

func (p Persister) UpdateVerificationFlow(ctx context.Context, r *verification.Flow) error {
	cp := *r
	cp.NID = corp.ContextualizeNID(ctx, p.nid)
	return p.update(ctx, cp)
}

func (p *Persister) CreateVerificationToken(ctx context.Context, token *link.VerificationToken) error {
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

func (p *Persister) UseVerificationToken(ctx context.Context, token string) (*link.VerificationToken, error) {
	var rt link.VerificationToken

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

		var va identity.VerifiableAddress
		if err := tx.Where("id = ? AND nid = ?", rt.VerifiableAddressID, nid).First(&va); err != nil {
			return sqlcon.HandleError(err)
		}

		rt.VerifiableAddress = &va

		/* #nosec G201 TableName is static */
		return tx.RawQuery(fmt.Sprintf("UPDATE %s SET used=true, used_at=? WHERE id=? AND nid = ?", rt.TableName(ctx)), time.Now().UTC(), rt.ID, nid).Exec()
	})); err != nil {
		return nil, err
	}

	return &rt, nil
}

func (p *Persister) DeleteVerificationToken(ctx context.Context, token string) error {
	nid := corp.ContextualizeNID(ctx, p.nid)
	/* #nosec G201 TableName is static */
	return p.GetConnection(ctx).RawQuery(fmt.Sprintf("DELETE FROM %s WHERE token=? AND nid = ?", new(link.VerificationToken).TableName(ctx)), token, nid).Exec()
}
