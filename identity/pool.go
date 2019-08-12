package identity

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"github.com/ory/hive/driver/configuration"
	"github.com/ory/hive/schema"
)

type (
	Pool interface {
		// FindByCredentialsIdentifier returns an identity by querying for it's credential identifiers.
		FindByCredentialsIdentifier(ctx context.Context, ct CredentialsType, match string) (*Identity, *Credentials, error)

		Create(context.Context, *Identity) (*Identity, error)

		List(ctx context.Context, limit, offset int) ([]Identity, error)

		Update(context.Context, *Identity) (*Identity, error)

		Delete(context.Context, string) error

		Get(context.Context, string) (*Identity, error)

		GetClassified(_ context.Context, id string) (*Identity, error)
	}

	PoolProvider interface {
		IdentityPool() Pool
	}

	abstractPool struct {
		c configuration.Provider
		d ValidationProvider
	}
)

func newAbstractPool(c configuration.Provider, d ValidationProvider) *abstractPool {
	return &abstractPool{c: c, d: d}
}

func (p *abstractPool) augment(i Identity) *Identity {
	if i.ID == "" {
		i.ID = uuid.New().String()
	}

	if i.TraitsSchemaURL == "" {
		i.TraitsSchemaURL = p.c.DefaultIdentityTraitsSchemaURL().String()
	}

	return &i
}

func (p *abstractPool) declassify(i Identity) *Identity {
	return i.CopyWithoutCredentials()
}

func (p *abstractPool) declassifyAll(i []Identity) []Identity {
	declassified := make([]Identity, len(i))
	for k, ii := range i {
		declassified[k] = *ii.CopyWithoutCredentials()
	}
	return declassified
}

func (p *abstractPool) Validate(i *Identity) error {
	if err := p.d.IdentityValidator().Validate(i); err != nil {
		if _, ok := errors.Cause(err).(schema.ResultErrors); ok {
			return errors.WithStack(herodot.ErrBadRequest.WithReasonf("%s", err))
		}
		return err
	}

	return nil
}
