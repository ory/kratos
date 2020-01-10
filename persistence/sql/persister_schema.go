package sql

import (
	"github.com/gobuffalo/pop"
	"github.com/gofrs/uuid"
	"github.com/ory/herodot"
	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/schema"
	"github.com/ory/viper"
	"github.com/ory/x/sqlcon"
	"github.com/pkg/errors"
)

func (p *Persister) GetSchema(id uuid.UUID) (*schema.Schema, error) {
	var s schema.Schema

	if id == uuid.Nil {
		return p.GetDefaultSchema()
	}

	if err := p.c.Find(&s, id); err != nil {
		return nil, sqlcon.HandleError(err)
	}
	return &s, nil
}

func (p *Persister) GetDefaultSchema() (*schema.Schema, error) {
	if !viper.IsSet(configuration.ViperKeyDefaultIdentityTraitsSchemaURL) {
		return nil, errors.WithStack(herodot.ErrNotFound.WithReason("The default identity traits schema is not set."))
	}

	return p.GetSchemaByUrl(viper.GetString(configuration.ViperKeyDefaultIdentityTraitsSchemaURL))
}

func (p *Persister) GetSchemaByUrl(url string) (*schema.Schema, error) {
	var s schema.Schema
	if err := p.c.Where("url = ?", url).First(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (p *Persister) RegisterSchema(s *schema.Schema) error {
	if s.URL == "" {
		return errors.WithStack(herodot.ErrBadRequest.WithReason("The schema is missing the URL property."))
	}

	return sqlcon.HandleError(p.c.Transaction(func(tx *pop.Connection) error {
		return tx.Create(s)
	}))
}

func (p *Persister) RegisterDefaultSchema(url string) (*schema.Schema, error) {
	ds := schema.Schema{
		URL: url,
	}
	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, ds.URL)
	return &ds, p.RegisterSchema(&ds)
}
