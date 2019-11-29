package sql

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gobuffalo/pop"
	"github.com/gofrs/uuid"
	"github.com/ory/herodot"
	"github.com/ory/x/errorsx"
	"github.com/ory/x/sqlcon"
	"github.com/pkg/errors"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
)

var _ identity.Pool = new(Persister)

func (p *Persister) FindByCredentialsIdentifier(ctx context.Context, ct identity.CredentialsType, match string) (*identity.Identity, *identity.Credentials, error) {
	panic("implement me")
}

func createIdentityCredentials(_ context.Context, tx *pop.Connection, i *identity.Identity) error {
	for k, cred := range i.Credentials {
		cred.IdentityID = i.ID
		if len(cred.Config) == 0 {
			cred.Config = json.RawMessage("{}")
		}

		if err := tx.Create(&cred); err != nil {
			return sqlcon.HandleError(err)
		}

		for _, ids := range cred.Identifiers {
			if strings.Contains(ids, "@") && cred.Type == identity.CredentialsTypePassword {
				ids = strings.ToLower(ids)
			}

			ci := &identity.CredentialIdentifier{
				Identifier:            ids,
				IdentityCredentialsID: cred.ID,
			}
			if err := tx.Create(ci); err != nil {
				return sqlcon.HandleError(err)
			}
		}

		i.Credentials[k] = cred
	}

	return nil
}

func (p *Persister) Create(ctx context.Context, i *identity.Identity) error {
	if i.TraitsSchemaURL == "" {
		i.TraitsSchemaURL = p.cf.DefaultIdentityTraitsSchemaURL().String()
	}

	if err := p.validateIdentity(i); err != nil {
		return err
	}

	return sqlcon.HandleError(p.c.Transaction(func(tx *pop.Connection) error {
		if err := tx.Create(i); err != nil {
			return sqlcon.HandleError(err)
		}

		return createIdentityCredentials(ctx, tx, i)
	}))
}

func (p *Persister) List(ctx context.Context, limit, offset int) ([]identity.Identity, error) {
	var is []identity.Identity
	return is, sqlcon.HandleError(p.c.RawQuery("SELECT * FROM identities LIMIT ? OFFSET ?", limit, offset).All(&is))
}

func (p *Persister) UpdateConfidential(ctx context.Context, i *identity.Identity) error {
	if err := p.validateIdentity(i); err != nil {
		return err
	}

	if c, ok := i.GetCredentials(identity.CredentialsTypePassword); ok {
		fmt.Printf("\n\npw config: %s %+v\n\n", c.Config, c.Identifiers)
	}

	return sqlcon.HandleError(p.c.Transaction(func(tx *pop.Connection) error {
		if err := tx.RawQuery(`DELETE FROM "identity_credentials" WHERE "identity_id" = ?`, i.ID).Exec(); err != nil {
			return sqlcon.HandleError(err)
		}

		if err := tx.Update(i); err != nil {
			return sqlcon.HandleError(err)
		}

		return createIdentityCredentials(ctx, tx, i)
	}))
}

func (p *Persister) Update(_ context.Context, i *identity.Identity) error {
	if err := p.validateIdentity(i); err != nil {
		return err
	}

	return sqlcon.HandleError(p.c.Update(i))
}

func (p *Persister) Delete(_ context.Context, id uuid.UUID) error {
	return sqlcon.HandleError(p.c.Destroy(&identity.Identity{ID: id}))
}

func (p *Persister) Get(_ context.Context, id uuid.UUID) (*identity.Identity, error) {
	var i identity.Identity
	if err := p.c.Find(&i, id); err != nil {
		return nil, sqlcon.HandleError(err)
	}
	i.Credentials = nil
	return &i, nil
}

func (p *Persister) GetClassified(_ context.Context, id uuid.UUID) (*identity.Identity, error) {
	var i identity.Identity
	if err := p.c.Eager().Find(&i, id); err != nil {
		return nil, sqlcon.HandleError(err)
	}

	i.Credentials = map[identity.CredentialsType]identity.Credentials{}
	for _, creds := range i.CredentialsCollection {
		var cs identity.CredentialIdentifierCollection
		if err := p.c.Where("identity_credential_id = ?", creds.ID).All(&cs); err != nil {
			return nil, sqlcon.HandleError(err)
		}

		creds.CredentialIdentifierCollection = nil
		creds.Identifiers = make([]string, len(cs))
		for k := range cs {
			creds.Identifiers[k] = cs[k].Identifier
		}
		i.Credentials[creds.Type] = creds
	}
	i.CredentialsCollection = nil

	return &i, nil
}

func (p *Persister) declassify(i identity.Identity) *identity.Identity {
	return i.CopyWithoutCredentials()
}

func (p *Persister) declassifyAll(i []identity.Identity) []identity.Identity {
	declassified := make([]identity.Identity, len(i))
	for k, ii := range i {
		declassified[k] = *ii.CopyWithoutCredentials()
	}
	return declassified
}

func (p *Persister) validateIdentity(i *identity.Identity) error {
	if err := p.r.IdentityValidator().Validate(i); err != nil {
		if _, ok := errorsx.Cause(err).(schema.ResultErrors); ok {
			return errors.WithStack(herodot.ErrBadRequest.WithReasonf("%s", err))
		}
		return err
	}

	return nil
}
