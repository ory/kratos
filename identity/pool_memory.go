package identity

import (
	"context"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/go-convenience/stringslice"
	"github.com/ory/herodot"
	"github.com/ory/x/pagination"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/schema"
)

var _ Pool = new(PoolMemory)

type PoolMemory struct {
	*abstractPool
	sync.RWMutex

	is []Identity
}

func NewPoolMemory(c configuration.Provider, d ValidationProvider) *PoolMemory {
	return &PoolMemory{
		abstractPool: newAbstractPool(c, d),
		is:           make([]Identity, 0),
	}
}

func (p *PoolMemory) hasConflictingID(i *Identity) bool {
	p.RLock()
	defer p.RUnlock()

	for _, fromPool := range p.is {
		if fromPool.ID == i.ID {
			return true
		}
	}
	return false
}

func (p *PoolMemory) hasConflictingCredentials(i *Identity) bool {
	p.RLock()
	defer p.RUnlock()

	for _, fromPool := range p.is {
		if fromPool.ID == i.ID {
			continue
		}

		for fromPoolID, fromPoolCredentials := range fromPool.Credentials {
			for credentialsID, cc := range i.Credentials {
				if fromPoolID == credentialsID {
					for _, identifier := range cc.Identifiers {
						if stringslice.Has(fromPoolCredentials.Identifiers, identifier) {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

// FindByCredentialsIdentifier returns an identity by querying for it's credential identifiers.
func (p *PoolMemory) FindByCredentialsIdentifier(_ context.Context, ct CredentialsType, match string) (*Identity, *Credentials, error) {
	p.RLock()
	defer p.RUnlock()

	for _, i := range p.is {
		for ctid, c := range i.Credentials {
			if ct == ctid {
				if stringslice.Has(c.Identifiers, match) {
					return p.declassify(i), &c, nil
				}
			}
		}
	}
	return nil, nil, errors.WithStack(herodot.ErrNotFound.WithReasonf("No identity matching the credentials identifiers"))
}

func (p *PoolMemory) CreateIdentity(_ context.Context, i *Identity) error {
	insert := p.augment(*i)
	if err := p.Validate(insert); err != nil {
		return err
	}

	if p.hasConflictingID(insert) {
		return errors.WithStack(herodot.ErrConflict.WithReasonf("An identity with the given ID exists already."))
	}

	if p.hasConflictingCredentials(insert) {
		return errors.WithStack(schema.NewDuplicateCredentialsError())
	}

	p.Lock()
	p.is = append(p.is, *insert)
	p.Unlock()

	return nil
}

func (p *PoolMemory) ListIdentities(_ context.Context, limit, offset int) ([]Identity, error) {
	p.RLock()
	defer p.RUnlock()

	start, end := pagination.Index(limit, offset, len(p.is))
	identities := make([]Identity, limit)
	for k, i := range p.is[start:end] {
		identities[k] = *p.declassify(i)
	}

	return p.abstractPool.declassifyAll(p.is[start:end]), nil
}

func (p *PoolMemory) UpdateIdentityConfidential(ctx context.Context, i *Identity) error {
	return p.update(ctx, i, i.Credentials, true)
}

func (p *PoolMemory) UpdateIdentity(ctx context.Context, i *Identity) error {
	return p.update(ctx, i, nil, false)
}

func (p *PoolMemory) update(ctx context.Context, i *Identity, ct map[CredentialsType]Credentials, updateCredentials bool) error {
	insert := p.augment(*i)
	insert.Credentials = ct
	if err := p.Validate(insert); err != nil {
		return err
	}

	if updateCredentials && p.hasConflictingCredentials(insert) {
		return errors.WithStack(schema.NewDuplicateCredentialsError())
	}

	p.RLock()
	for k, ii := range p.is {
		if ii.ID == insert.ID {
			p.RUnlock()

			p.Lock()
			if !updateCredentials {
				insert.Credentials = ii.Credentials
			}
			p.is[k] = *insert
			p.Unlock()

			return nil
		}
	}
	p.RUnlock()
	return errors.WithStack(herodot.ErrNotFound.WithReasonf("Identity with identifier %s does not exist.", i.ID))
}

func (p *PoolMemory) GetIdentity(ctx context.Context, id uuid.UUID) (*Identity, error) {
	i, err := p.GetIdentityConfidential(ctx, id)
	if err != nil {
		return nil, err
	}

	return p.declassify(*i), nil
}

func (p *PoolMemory) GetIdentityConfidential(_ context.Context, id uuid.UUID) (*Identity, error) {
	p.RLock()
	defer p.RUnlock()

	for _, ii := range p.is {
		if ii.ID == id {
			return &ii, nil
		}
	}

	return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf("Identity with identifier %s does not exist.", id))
}

func (p *PoolMemory) DeleteIdentity(_ context.Context, id uuid.UUID) error {
	p.Lock()
	defer p.Unlock()

	offset := -1
	for k, ii := range p.is {
		if ii.ID == id {
			offset = k
			break
		}
	}

	if offset == -1 {
		return errors.WithStack(herodot.ErrNotFound.WithReasonf("Identity with identifier %s does not exist.", id))
	}

	p.is = append(p.is[:offset], p.is[offset+1:]...)

	return nil
}
