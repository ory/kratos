package identity

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/ory/go-convenience/stringslice"
	"github.com/ory/herodot"
	"github.com/ory/x/pagination"

	"github.com/ory/hive/driver/configuration"
	"github.com/ory/hive/schema"
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

func (p *PoolMemory) Create(_ context.Context, i *Identity) (*Identity, error) {
	insert := p.augment(*i)
	if err := p.Validate(insert); err != nil {
		return nil, err
	}

	logrus.New().Printf("creating identity: %s %+v", insert.ID, insert.Credentials)

	if p.hasConflictingID(insert) {
		return nil, errors.WithStack(herodot.ErrConflict.WithReasonf("An identity with the given ID exists already."))
	}

	if p.hasConflictingCredentials(insert) {
		return nil, errors.WithStack(schema.NewDuplicateCredentialsError())
	}

	p.RLock()
	insert.PK = uint64(len(p.is) + 1)
	p.RUnlock()

	p.Lock()
	p.is = append(p.is, *insert)
	p.Unlock()

	return p.abstractPool.declassify(*insert), nil
}

func (p *PoolMemory) List(_ context.Context, limit, offset int) ([]Identity, error) {
	p.RLock()
	defer p.RUnlock()

	start, end := pagination.Index(limit, offset, len(p.is))
	identities := make([]Identity, limit)
	for k, i := range p.is[start:end] {
		identities[k] = *p.declassify(i)
	}

	return p.abstractPool.declassifyAll(p.is[start:end]), nil
}

func (p *PoolMemory) Update(_ context.Context, i *Identity) (*Identity, error) {
	insert := p.augment(*i)
	if err := p.Validate(insert); err != nil {
		return nil, err
	}

	if p.hasConflictingCredentials(insert) {
		return nil, errors.WithStack(schema.NewDuplicateCredentialsError())
	}

	p.RLock()
	for k, ii := range p.is {
		if ii.ID == insert.ID {
			p.RUnlock()

			p.Lock()
			i.PK = ii.PK
			p.is[k] = *insert
			p.Unlock()

			return p.declassify(*insert), nil
		}
	}
	p.RUnlock()
	return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf("Identity with identifier %s does not exist.", i.ID))
}

func (p *PoolMemory) Get(ctx context.Context, id string) (*Identity, error) {
	i, err := p.GetClassified(ctx, id)
	if err != nil {
		return nil, err
	}

	return p.declassify(*i), nil
}

func (p *PoolMemory) GetClassified(_ context.Context, id string) (*Identity, error) {
	p.RLock()
	defer p.RUnlock()

	for _, ii := range p.is {
		if ii.ID == id {
			return &ii, nil
		}
	}

	return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf("Identity with identifier %s does not exist.", id))

}

func (p *PoolMemory) Delete(_ context.Context, id string) error {
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
