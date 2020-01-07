package sql

import (
	"errors"
	"github.com/gobuffalo/pop"
	"github.com/gofrs/uuid"
	"github.com/ory/kratos/schema"
	"github.com/ory/x/sqlcon"
)

func (p *Persister) GetSchema(id uuid.UUID) (*schema.JsonSchema, error) {
	var s schema.JsonSchema
	if err := p.c.Find(&s, id); err != nil {
		return nil, sqlcon.HandleError(err)
	}
	return &s, nil
}

func (p *Persister) CreateSchema(s schema.JsonSchema) error {
	if s.Url == "" {
		return errors.New("URL must be provided")
	}

	return sqlcon.HandleError(p.c.Transaction(func(tx *pop.Connection) error {
		return tx.Create(s)
	}))
}
