package sql

import (
	"context"
	"fmt"
	"time"

	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"

	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/selfservice/flow/verify"
)

var _ verify.Persister = new(Persister)

func (p Persister) CreateVerifyRequest(ctx context.Context, r *verify.Request) error {
	// This should not create the request eagerly because otherwise we might accidentally create an address
	// that isn't supposed to be in the database.
	return p.c.Create(r)
}

func (p Persister) GetVerifyRequest(ctx context.Context, id uuid.UUID) (*verify.Request, error) {
	var r verify.Request
	if err := p.c.Find(&r, id); err != nil {
		return nil, sqlcon.HandleError(err)
	}
	return &r, nil
}

func (p Persister) UpdateVerifyRequest(ctx context.Context, r *verify.Request) error {
	return sqlcon.HandleError(p.c.Update(r))
}

func (p Persister) TrackAddresses(ctx context.Context, addresses []verify.Address) error {
	return sqlcon.HandleError(p.c.Transaction(func(tx *pop.Connection) error {
		for k := range addresses {
			address := addresses[k]
			if err := p.c.Create(&address); err != nil {
				return err
			}
			addresses[k] = address
		}
		return nil
	}))
}

func (p *Persister) FindAddressByCode(ctx context.Context, code string) (*verify.Address, error) {
	var address verify.Address
	if err := p.c.Where("code = ?", code).First(&address); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &address, nil
}

func (p *Persister) FindAddressByValue(ctx context.Context, via verify.Via, value string) (*verify.Address, error) {
	var address verify.Address
	if err := p.c.Where("via = ? AND value = ?", via, value).First(&address); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	return &address, nil
}

func (p *Persister) VerifyAddress(ctx context.Context, code string) error {
	newCode, err := verify.NewVerifyCode()
	if err != nil {
		return err
	}

	return sqlcon.HandleError(p.c.RawQuery(
		fmt.Sprintf(
			"UPDATE %s SET status = ?, verified = true, verified_at = ?, code = ? WHERE code = ?",
			new(verify.Address).TableName(),
		),
		verify.StatusCompleted,
		time.Now().UTC().Round(time.Second),
		newCode,
		code,
	).Exec())
}

func (p *Persister) UpdateAddress(ctx context.Context, address *verify.Address) error {
	return sqlcon.HandleError(p.c.Update(address))
}
