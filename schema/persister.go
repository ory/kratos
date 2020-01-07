package schema

import "github.com/gofrs/uuid"

type (
	Schema struct {
		ID  uuid.UUID `json:"-" faker:"uuid" db:"id" rw:"r"`
		URL string    `json:"url" faker:"-" db:"url"`
	}

	Persister interface {
		GetSchema(uuid uuid.UUID) (*Schema, error)
		GetSchemaByUrl(url string) (*Schema, error)
		CreateSchema(s Schema) error
	}

	PersistenceProvider interface {
		SchemaPersister() Persister
	}
)
