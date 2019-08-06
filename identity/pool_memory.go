package identity

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	"github.com/ory/go-convenience/stringslice"

	"github.com/ory/herodot"
	"github.com/ory/x/pagination"
)

var _ Pool = new(PoolMemory)

type PoolMemory struct {
	sync.RWMutex
	r  Registry
	is []Identity
}

func NewPoolMemory(r Registry) *PoolMemory {
	return &PoolMemory{
		is: make([]Identity, 0),
		r:  r,
	}
}

func (p *PoolMemory) RequestID() string {
	return "memory"
}

func (p *PoolMemory) hasConflict(i *Identity) bool {
	p.RLock()
	defer p.RUnlock()

	for _, fromPool := range p.is {
		if fromPool.ID == i.ID {
			return true
		}

		for _, fromPoolCredentials := range fromPool.Credentials {
			for _, cc := range i.Credentials {
				if cc.ID == fromPoolCredentials.ID {
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
		for _, c := range i.Credentials {
			if stringslice.Has(c.Identifiers, match) {
				return &i, &c, nil
			}
		}
	}
	return nil, nil, errors.WithStack(herodot.ErrNotFound.WithReasonf("No identity matching the credentials identifiers"))
}

func (p *PoolMemory) Create(_ context.Context, i *Identity) (*Identity, error) {
	if p.hasConflict(i) {
		return nil, errors.WithStack(herodot.ErrConflict.WithReasonf("An identity with the given identifier(s) exists already."))
	}

	p.RLock()
	i.PK = uint64(len(p.is) + 1)
	p.RUnlock()

	p.Lock()
	p.is = append(p.is, *i)
	p.Unlock()

	return i, nil
}

func (p *PoolMemory) List(_ context.Context, limit, offset int) ([]Identity, error) {
	p.RLock()
	defer p.RUnlock()

	start, end := pagination.Index(limit, offset, len(p.is))
	return p.is[start:end], nil
}

func (p *PoolMemory) Update(_ context.Context, i *Identity) (*Identity, error) {
	p.RLock()
	for k, ii := range p.is {
		if ii.ID == i.ID {
			p.RUnlock()
			p.Lock()
			p.is[k] = Identity{}
			p.Unlock()

			if p.hasConflict(i) {
				p.is[k] = ii
				return nil, errors.WithStack(herodot.ErrConflict.WithReasonf("An identity with the given identifier(s) exists already."))
			}

			p.Lock()
			i.PK = ii.PK
			p.is[k] = *i
			p.Unlock()

			return i, nil
		}
	}
	p.RUnlock()
	return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf("Identity with identifier %s does not exist.", i.ID))
}

func (p *PoolMemory) Get(_ context.Context, i string) (*Identity, error) {
	p.RLock()
	defer p.RUnlock()

	for _, ii := range p.is {
		if ii.ID == i {
			return &ii, nil
		}
	}

	return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf("Identity with identifier %s does not exist.", i))
}

func (p *PoolMemory) Delete(_ context.Context, i string) error {
	p.Lock()
	defer p.Unlock()

	offset := -1
	for k, ii := range p.is {
		if ii.ID == i {
			offset = k
			break
		}
	}

	if offset == -1 {
		return errors.WithStack(herodot.ErrNotFound.WithReasonf("Identity with identifier %s does not exist.", i))
	}

	p.is = append(p.is[:offset], p.is[offset+1:]...)

	return nil
}
