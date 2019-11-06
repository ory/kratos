package identity

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/schema"
)

type (
	Pool interface {
		// FindByCredentialsIdentifier returns an identity by querying for it's credential identifiers.
		FindByCredentialsIdentifier(ctx context.Context, ct CredentialsType, match string) (*Identity, *Credentials, error)

		// Create creates an identity. It is capable of setting credentials without encoding. Will return an error
		// if identity exists, backend connectivity is broken, or trait validation fails.
		Create(context.Context, *Identity) (*Identity, error)

		// Create creates an identity. It is capable of setting credentials without encoding. Will return an error
		// if identity exists, backend connectivity is broken, or trait validation fails.
		List(ctx context.Context, limit, offset int) ([]Identity, error)

		// UpdateConfidential updates an identities confidential data. It is capable of setting credentials without encoding. Will return an error
		// if identity exists, backend connectivity is broken, or trait validation fails.
		//
		// Because this will overwrite credentials you always need to update the identity using `GetClassified`.
		UpdateConfidential(context.Context, *Identity, map[CredentialsType]Credentials) (*Identity, error)

		// Update updates an identity excluding its confidential data. It is capable of setting credentials without encoding. Will return an error
		// if identity exists, backend connectivity is broken, or trait validation fails.
		//
		// This update procedure works well with `Get`.
		Update(context.Context, *Identity) (*Identity, error)

		// Delete removes an identity by its id. Will return an error
		// 		// if identity exists, backend connectivity is broken, or trait validation fails.
		Delete(context.Context, string) error

		// Get returns an identity by its id. Will return an error if the identity does not exist or backend
		// connectivity is broken.
		Get(context.Context, string) (*Identity, error)

		// GetClassified returns the identity including it's raw credentials. This should only be used internally.
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
