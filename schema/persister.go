package schema

import (
	"time"

	"github.com/gofrs/uuid"
)

type (
	Schema struct {
		ID  uuid.UUID `json:"id" faker:"uuid" db:"id" rw:"r"`
		URL string    `json:"url" faker:"-" db:"url"`
		// CreatedAt is a helper struct field for gobuffalo.pop.
		CreatedAt time.Time `json:"-" db:"created_at"`
		// UpdatedAt is a helper struct field for gobuffalo.pop.
		UpdatedAt time.Time `json:"-" db:"updated_at"`
	}

	Persister interface {
		GetSchema(uuid uuid.UUID) (*Schema, error)
		GetSchemaByUrl(url string) (*Schema, error)
		GetDefaultSchema() (*Schema, error)
		RegisterSchema(s *Schema) error
		RegisterDefaultSchema(url string) (*Schema, error)
	}

	PersistenceProvider interface {
		SchemaPersister() Persister
	}
)

func (s Schema) TableName() string {
	return "json_schemas"
}
